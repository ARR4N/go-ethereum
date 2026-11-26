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

package state

import (
	"testing"

	"github.com/ethereum/go-ethereum/common"
)

// TestPrefetcherReadsThroughOwnTrie verifies that, unless a [WorkerPool] is in
// use, a [subfetcher] reads through its own [Trie], as unmodified geth does;
// and that a [WorkerPool] is not used at all when collecting a stateless
// witness, which is built from the nodes that said [Trie] has accessed.
func TestPrefetcherReadsThroughOwnTrie(t *testing.T) {
	tests := []struct {
		name          string
		witnessing    bool
		withPool      bool
		wantPool      bool
		wantOwnTrieIO bool
	}{
		{name: "no_witness_no_pool", wantOwnTrieIO: true},
		{name: "no_witness_with_pool", withPool: true, wantPool: true},
		{name: "witness_no_pool", witnessing: true, wantOwnTrieIO: true},
		{name: "witness_with_pool", witnessing: true, withPool: true, wantOwnTrieIO: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var (
				opts      []PrefetcherOption
				ctorCalls int
			)
			if tt.withPool {
				opts = append(opts, WithWorkerPool(func() WorkerPool {
					ctorCalls++
					return &countingWorkerPool{}
				}))
			}

			// Opening a trie accesses its root node, so there MUST be enough
			// accounts for reading one of them to access further nodes.
			db := filledStateDB()
			for i := range 256 {
				db.SetNonce(common.BytesToAddress([]byte{1, byte(i)}), 1)
			}
			root, err := db.Commit(0, false)
			if err != nil {
				t.Fatalf("%T.Commit() error %v", db, err)
			}

			p := newTriePrefetcher(db.db, root, "", !tt.witnessing /*noreads*/, opts...)
			if got, want := ctorCalls == 1, tt.wantPool; got != want {
				t.Errorf("newTriePrefetcher([noreads = %t], [%d options]) constructed %T = %t; want %t", !tt.witnessing, len(opts), p.workers, got, want)
			}

			keys := [][]byte{common.HexToAddress("0xaffeaffeaffeaffeaffeaffeaffeaffeaffeaffe").Bytes()}
			if err := p.prefetch(common.Hash{}, root, common.Address{}, keys, false /*read*/); err != nil {
				t.Fatalf("%T.prefetch() error %v", p, err)
			}
			p.terminate(false /*async*/)

			for id, sf := range p.fetchers {
				if got, want := sf.pool.workers != nil, tt.wantPool; got != want {
					t.Errorf("%T[%x] has %T = %t; want %t", sf, id, sf.pool.workers, got, want)
				}
				const rootOnly = 1
				if got, want := len(sf.trie.Witness()) > rootOnly, tt.wantOwnTrieIO; got != want {
					t.Errorf("%T[%x] own %T accessed nodes other than root = %t; want %t", sf, id, sf.trie, got, want)
				}
			}
			if len(p.fetchers) != 1 {
				t.Errorf("%T has %d subfetchers; want 1", p, len(p.fetchers))
			}
		})
	}
}
