// Copyright 2026 the libevm authors.
//
// The libevm additions to go-ethereum are free software: you can redistribute
// them and/or modify them under the terms of the GNU Lesser General Public License
// as published by the Free Software Foundation, either version 3 of the License,
// or (at your option) any later version.
//
// The libevm additions are distributed in the hope that they will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the GNU Lesser
// General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the go-ethereum library. If not, see
// <http://www.gnu.org/licenses/>.
package core_test

import (
	"math"
	"math/big"
	"testing"

	"github.com/holiman/uint256"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core"
	"github.com/ethereum/go-ethereum/core/tracing"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/core/vm"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/libevm/ethtest"
	"github.com/ethereum/go-ethereum/libevm/hookstest"
	"github.com/ethereum/go-ethereum/params"
)

// A gasChange is a recorded call to [tracing.Hooks.OnGasChange].
type gasChange struct {
	Old, New uint64
	Reason   tracing.GasChangeReason
}

// A balanceChange is a recorded call to [tracing.Hooks.OnBalanceChange].
type balanceChange struct {
	Addr      common.Address
	Prev, New uint64
	Reason    tracing.BalanceChangeReason
}

// TestStateTransitionAdjustmentsReportedToTracer asserts that every libevm
// modification of gas or balances during a state transition is reported to the
// tracer, with an appropriate reason.
func TestStateTransitionAdjustmentsReportedToTracer(t *testing.T) {
	const (
		baseFee   = 3
		gasLimit  = 100_000
		gasUsed   = params.TxGas // simple transfer to an EOA
		remaining = gasLimit - gasUsed
		refund    = 1_000 // less than the cap of gasUsed/5
	)
	coinbase := common.Address{'c', 'o', 'i', 'n'}

	tests := []struct {
		name  string
		hooks *hookstest.Stub
		// Expected [gasChange] values for the respective reasons, in order.
		wantRefunds, wantUnspecified []gasChange
		wantLeftOver                 gasChange
		// Expected increases in the coinbase balance, in order.
		wantCoinbase []balanceChange
	}{
		{
			name:  "default_behaviour",
			hooks: &hookstest.Stub{},
			wantRefunds: []gasChange{
				{remaining, remaining + refund, tracing.GasChangeTxRefunds},
			},
			wantLeftOver: gasChange{remaining + refund, 0, tracing.GasChangeTxLeftOverReturned},
		},
		{
			name:  "refunds_disabled",
			hooks: &hookstest.Stub{DisableGasRefunds: true},
			wantRefunds: []gasChange{
				{remaining, remaining + refund, tracing.GasChangeTxRefunds}, // upstream
				{remaining + refund, remaining, tracing.GasChangeTxRefunds}, // libevm reversal
			},
			wantLeftOver: gasChange{remaining, 0, tracing.GasChangeTxLeftOverReturned},
		},
		{
			name: "minimum_gas_consumption",
			hooks: &hookstest.Stub{
				MinimumGasConsumptionFn: func(limit uint64) uint64 { return limit / 2 },
			},
			wantRefunds: []gasChange{
				{remaining, remaining + refund, tracing.GasChangeTxRefunds},
			},
			wantUnspecified: []gasChange{
				{remaining + refund, gasLimit / 2, tracing.GasChangeLibEVMMinimumConsumption},
			},
			wantLeftOver: gasChange{gasLimit / 2, 0, tracing.GasChangeTxLeftOverReturned},
		},
		{
			name:  "base_fee_credited_to_coinbase",
			hooks: &hookstest.Stub{CreditBaseFeeToCoinbase: true},
			wantRefunds: []gasChange{
				{remaining, remaining + refund, tracing.GasChangeTxRefunds},
			},
			wantLeftOver: gasChange{remaining + refund, 0, tracing.GasChangeTxLeftOverReturned},
			wantCoinbase: []balanceChange{{
				Addr:   coinbase,
				Prev:   0,
				New:    (gasUsed - refund) * baseFee,
				Reason: tracing.BalanceIncreaseRewardTransactionFee,
			}},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.hooks.Register(t)

			key, err := crypto.GenerateKey()
			require.NoError(t, err, "crypto.GenerateKey()")
			eoa := crypto.PubkeyToAddress(key.PublicKey)

			config := params.TestChainConfig
			sdb, evm := ethtest.NewZeroEVM(t, ethtest.WithChainConfig(config))
			sdb.SetBalance(eoa, new(uint256.Int).SetAllOne(), tracing.BalanceChangeUnspecified)
			sdb.AddRefund(refund)

			// There is no priority fee so the only credit to the coinbase, if
			// any, is of the base fee.
			tx := types.MustSignNewTx(key, types.LatestSigner(config), &types.DynamicFeeTx{
				GasFeeCap: big.NewInt(baseFee),
				GasTipCap: big.NewInt(0),
				Gas:       gasLimit,
				To:        &common.Address{'t', 'o'},
			})

			var (
				gotGas      []gasChange
				gotCoinbase []balanceChange
			)
			tracer := &tracing.Hooks{
				OnGasChange: func(old, new uint64, reason tracing.GasChangeReason) {
					gotGas = append(gotGas, gasChange{old, new, reason})
				},
				OnBalanceChange: func(addr common.Address, prev, new *big.Int, reason tracing.BalanceChangeReason) {
					if addr == coinbase {
						gotCoinbase = append(gotCoinbase, balanceChange{addr, prev.Uint64(), new.Uint64(), reason})
					}
				},
			}

			// Balance changes are reported by the state database itself.
			sdb.SetLogger(tracer)

			var used uint64
			gp := core.GasPool(math.MaxUint64)
			_, err = core.ApplyTransaction(
				evm.ChainConfig(), nil, &coinbase, &gp, sdb,
				&types.Header{
					BaseFee: big.NewInt(baseFee),
					// Required but irrelevant fields
					Number:     big.NewInt(0),
					Difficulty: big.NewInt(0),
				},
				tx, &used, vm.Config{Tracer: tracer},
			)
			require.NoError(t, err, "core.ApplyTransaction(...)")

			byReason := make(map[tracing.GasChangeReason][]gasChange)
			for _, g := range gotGas {
				byReason[g.Reason] = append(byReason[g.Reason], g)
			}
			assert.Equal(t, tt.wantRefunds, byReason[tracing.GasChangeTxRefunds], "gas changes reported as refunds")
			assert.Equal(t, tt.wantUnspecified, byReason[tracing.GasChangeUnspecified], "gas changes reported without a specific reason")
			require.NotEmpty(t, gotGas, "gas changes reported")
			assert.Equal(t, tt.wantLeftOver, gotGas[len(gotGas)-1], "final gas change")
			assert.Equal(t, tt.wantCoinbase, gotCoinbase, "coinbase balance changes")
		})
	}
}
