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

package types_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	. "github.com/ethereum/go-ethereum/core/types"
)

// TestNewBlockClonesBodyPayload asserts that [NewBlock] carries the payload of
// its [Body] argument into the new [Block], as a deep copy, in keeping with
// its documented behaviour that "the input data is copied".
func TestNewBlockClonesBodyPayload(t *testing.T) {
	TestOnlyClearRegisteredExtras()
	t.Cleanup(TestOnlyClearRegisteredExtras)

	extras := RegisterExtras[
		NOOPHeaderHooks, *NOOPHeaderHooks,
		blockPayload, *blockPayload,
		struct{},
	]()

	const initial uint64 = 42
	payload := &blockPayload{x: initial}
	var body Body
	extras.Body.Set(&body, payload)

	block := NewBlock(&Header{}, &body, nil, nil)

	payload.x++
	require.Equalf(t, initial+1, extras.Body.Get(&body).x, "%T payload after modification via pointer", &body)
	assert.Equalf(t, initial, extras.Block.Get(block).x, "%T payload after modification of the %T payload from which it was cloned", block, &body)

	t.Run("nil_body", func(t *testing.T) {
		block := NewBlock(&Header{}, nil, nil, nil)
		assert.Zerof(t, extras.Block.Get(block).x, "%T payload when created from a nil %T", block, &body)
	})
}
