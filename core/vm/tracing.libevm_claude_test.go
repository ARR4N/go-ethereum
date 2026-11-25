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
package vm_test

import (
	"testing"

	"github.com/holiman/uint256"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/tracing"
	"github.com/ethereum/go-ethereum/core/vm"
	"github.com/ethereum/go-ethereum/libevm"
	"github.com/ethereum/go-ethereum/libevm/ethtest"
	"github.com/ethereum/go-ethereum/libevm/hookstest"
)

// A gasChange is a recorded call to [tracing.Hooks.OnGasChange].
type gasChange struct {
	Old, New uint64
	Reason   tracing.GasChangeReason
}

// gasChangeRecorder returns a tracer that appends to the returned slice.
func gasChangeRecorder() (*tracing.Hooks, *[]gasChange) {
	var got []gasChange
	return &tracing.Hooks{
		OnGasChange: func(old, new uint64, reason tracing.GasChangeReason) {
			got = append(got, gasChange{old, new, reason})
		},
	}, &got
}

// TestPrecompileGasReportedToTracer asserts that all gas accounting performed
// on behalf of a stateful precompile is reported to the tracer, with the same
// reasons that upstream uses for equivalent charges.
func TestPrecompileGasReportedToTracer(t *testing.T) {
	rng := ethtest.NewPseudoRand(314159)
	precompile := rng.Address()
	callee := rng.Address() // no code, so it returns all of its gas

	const ownGas = 123
	var beforeCall, afterCall uint64
	hooks := &hookstest.Stub{
		PrecompileOverrides: map[common.Address]libevm.PrecompiledContract{
			precompile: vm.NewStatefulPrecompile(func(env vm.PrecompileEnvironment, input []byte) ([]byte, error) {
				if !env.UseGas(ownGas) {
					return nil, vm.ErrOutOfGas
				}
				beforeCall = env.Gas()
				_, err := env.Call(callee, nil, env.Gas(), uint256.NewInt(0))
				afterCall = env.Gas()
				return nil, err
			}),
		},
	}
	hooks.Register(t)

	_, evm := ethtest.NewZeroEVM(t, ethtest.WithAllEIPs())
	tracer, got := gasChangeRecorder()
	evm.Config.Tracer = tracer

	_, _, err := evm.Call(vm.AccountRef(rng.Address()), precompile, nil, 1e6, uint256.NewInt(0))
	require.NoError(t, err, "evm.Call([precompile that uses gas and calls an account])")

	var (
		ownUse           []uint64 // decreases
		callCost, refund uint64
	)
	for _, g := range *got {
		switch g.Reason {
		case tracing.GasChangeCallPrecompiledContract:
			ownUse = append(ownUse, g.Old-g.New)
		case tracing.GasChangeCallOpCode:
			callCost += g.Old - g.New
		case tracing.GasChangeCallLeftOverRefunded:
			refund += g.New - g.Old
		}
	}

	assert.Contains(t, ownUse, uint64(ownGas), "gas used by the precompile itself")
	assert.NotZero(t, callCost, "gas charged for the outbound call")
	assert.NotZero(t, refund, "gas returned from the outbound call")
	assert.Equal(t, beforeCall-afterCall, callCost-refund, "net reported cost of the outbound call")
}

// TestPreprocessingGasReportedToTracer asserts that the charge returned by
// [vm.Hooks.PreprocessingGasCharge] is reported to the tracer.
func TestPreprocessingGasReportedToTracer(t *testing.T) {
	const (
		gas    = 1_000_000
		charge = 1_234
	)
	txHash := common.Hash{'t', 'x'}

	vm.RegisterHooks(&preprocessingCharger{
		charge: map[common.Hash]uint64{txHash: charge},
	})
	t.Cleanup(vm.TestOnlyClearRegisteredHooks)

	sdb, evm := ethtest.NewZeroEVM(t)
	sdb.SetTxContext(txHash, 0)
	tracer, got := gasChangeRecorder()
	evm.Config.Tracer = tracer

	_, _, err := evm.Call(vm.AccountRef(common.Address{'e', 'o', 'a'}), common.Address{'t', 'o'}, nil, gas, uint256.NewInt(0))
	require.NoError(t, err, "evm.Call(...)")

	require.NotEmpty(t, *got, "gas changes reported")
	want := gasChange{gas, gas - charge, tracing.GasChangeLibEVMPreprocessing}
	assert.Equal(t, want, (*got)[0], "first gas change")
}
