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

package tracing

// libevm-specific reasons for gas changes, reported to tracers via
// [Hooks.OnGasChange].
//
// TODO(arr4n): give these dedicated values. Until then they are
// indistinguishable from [GasChangeUnspecified], which is the only existing
// value that doesn't misdescribe them.
const (
	// GasChangeLibEVMMinimumConsumption is the additional gas charged, after
	// refunds, to honour [params.RulesHooks.MinimumGasConsumption].
	GasChangeLibEVMMinimumConsumption = GasChangeUnspecified
	// GasChangeLibEVMPreprocessing is the gas charged, between intrinsic gas
	// and execution, as returned by [vm.Hooks.PreprocessingGasCharge].
	GasChangeLibEVMPreprocessing = GasChangeUnspecified
)
