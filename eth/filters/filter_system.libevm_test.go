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

package filters

import (
	"math/big"
	"testing"

	"github.com/ethereum/go-ethereum/core/rawdb"
	"github.com/ethereum/go-ethereum/core/types"
)

type bloomOverriderBackend struct {
	*testBackend
	overridden chan struct{}
}

var _ BloomOverrider = (*bloomOverriderBackend)(nil)

func (b *bloomOverriderBackend) OverrideHeaderBloom(header *types.Header) types.Bloom {
	b.overridden <- struct{}{}
	return header.Bloom
}

func TestBloomOverride(t *testing.T) {
	db := rawdb.NewMemoryDatabase()
	backend, sys := newTestFilterSystem(t, db, Config{})
	sut := &bloomOverriderBackend{
		testBackend: backend,
		overridden:  make(chan struct{}),
	}
	sys.backend = sut

	t.Run("blockLogs", func(t *testing.T) {
		hdr := &types.Header{Number: big.NewInt(0)}
		h := hdr.Hash()
		rawdb.WriteHeader(db, hdr)
		rawdb.WriteCanonicalHash(db, h, 0)
		rawdb.WriteHeaderNumber(db, h, 0)

		go sys.NewBlockFilter(h, nil, nil).Logs(t.Context()) //nolint:errcheck // Known but irrelevant error
		<-sut.overridden
	})
}
