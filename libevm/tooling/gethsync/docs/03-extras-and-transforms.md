# 3. Extras and transforms

How libevm changes group 2 (upstream) files, and the rules those changes obey.
Needed to resolve a conflict at a hook site by re-deriving intent instead of
merging text.

## Measures of disruption (ordered)

1. **MUST keep semantic backwards compatibility by default.** With no extras
   registered, libevm at version X behaves as geth at version X. "Identical
   behaviour" is *defined* as upstream's tests passing (the geth oracle).
   Zero changes to upstream tests has always been of extreme importance.
2. **SHOULD keep syntactic backwards compatibility by default.** Upstream code
   written against the original API, including code not yet written, still
   compiles. Hence variadic options, optional interface extensions, etc.
   Known limit: variadic options protect call sites, **not other implementers**
   of an interface whose methods gained them.
3. **SHOULD be expressible as a mechanical (AST-like) transform**: add
   arguments, copy arguments, wrap a call. Utopian as automation, but it makes
   intent obvious in a conflict. A change that is not scaffold-like (arbitrary
   code mid-function) must aim for minimal conflict likelihood and **MUST
   document the rationale for its location**.

## Transform shapes seen so far

| Shape | Example |
|---|---|
| Rename and wrap | `TransitionDb` → `transitionDb` in the upstream file; exported name reinstated in `core/state_transition.libevm.go` |
| Same-named method shadowing a function | `RunPrecompiledContract(p, …)` → `args.RunPrecompiledContract(p, …)` with `args := &evmCallArgs{…}` |
| Argument pass-through | `blockCtx, … = overrideNewEVMArgs(blockCtx, …)`; `evm.TxContext, evm.StateDB = evm.overrideEVMResetArgs(txCtx, statedb)` |
| Early-return guard | `PrecompileOverride` check ahead of the standard precompile lookup |
| Append / delegate to a hook | `return banner + c.Hooks().Description()`; `return c.Hooks().CheckConfigForkOrder()` |
| Payload field | `extra *pseudo.Type // See RegisterExtras()` on a struct |
| Variadic option on a method | `GetState(addr, key, ...stateconf.StateDBStateOption)` |
| Arbitrary placement with rationale | `canCreateContract` in `EVM.create`, with a comment that it must follow the caller's nonce increment and precede all other state changes |

This list is from sampling `core/vm/evm.go`, `core/state_transition.go` and
`params/config.go`, not from a full inventory (open item).

## The canonical extra: `params`

Created first because `ChainConfig` and `Rules` are plumbed through the whole
codebase, so they carry behaviour anywhere without further modification.

- **Payload:** one unexported `extra *pseudo.Type` on each of `ChainConfig` and
  `Rules`. Typed access only via the `ExtraPayloads[C, R]` accessor returned by
  `RegisterExtras`.
- **Registration:** `RegisterExtras` once, in `init()`. Tests use
  `TestOnlyClearRegisteredExtras`; off-chain consumers use
  `WithTempRegisteredExtras` under `libevm.WithTemporaryExtrasLock`.
- **Hooks:** `c.Hooks()` / `r.Hooks()` return the registered type, else
  `NOOPHooks`. **Each no-op is the geth default**, which is what delivers
  measure 1. Embedding `NOOPHooks` implements a subset.
- **Upstream edits are one-liners** that append or delegate; all logic lives in
  `*.libevm.go` (`config.libevm.go`, `hooks.libevm.go`, `json.libevm.go`).
- geth always preserves backwards compatibility through `params.ChainConfig`
  and `params.Rules`; this is why libevm tests can stay semantically unchanged
  across a sync.

## Marker comments are not reliable

No strong convention was ever set. In upstream-owned files on `main`: three
`//libevm:start` / `//libevm:end` pairs and about 16 looser comments
(`// libevm:`, `// libevm extra`, `// libevm addition`, ...). Most hook sites
are identifiable only by diffing against upstream. **The diff against upstream
is the source of truth for the transform inventory, not the comments.**

## Resolving a conflict at a hook site

1. Upstream's side wins; the transform is then re-seated on top.
2. Never adapt upstream code to fit libevm.
3. Re-derive the transform from its shape and its hook; do not merge text.
4. If the shape cannot be named, **stop and ask**.
5. If upstream changed something the hook depends on, absorb it in the
   `*.libevm.go` file so the edit to the upstream file keeps its shape.
6. If upstream now provides natively what a hook exists for: stop and raise it.
   Expected outcome is deletion of the local change; the maintainer doubts the
   case will arise.

## Known limits of measure 2 in practice

- Interface implementers need the variadic parameter added: upstream test
  double `dummyStatedb` in `eth/tracers/logger/logger_test.go`; in the old
  branch, upstream's new `core/tracing/hooks.go` at v1.14.0. Expect more.
- #110 (`Block.WithBody(Body)`) changed an upstream signature; it pre-empts
  upstream and is treated as a cherry-pick.
- #280 (`IntrinsicGas` bools → `params.Rules`) changed an upstream signature
  and is libevm's own. See [06](06-cherry-picks.md).
