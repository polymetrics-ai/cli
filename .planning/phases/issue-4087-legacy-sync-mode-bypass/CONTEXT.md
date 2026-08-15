# Issue #4087 context: close legacy sync-mode bypass

## Discussion outcome

Issue #4087 is a narrow connector-neutral application-contract correction. The public compatibility spellings remain accepted; no connector definition, credential, generated connector bundle, or CLI flag changes are in scope. Runtime help and its generated/website documentation must correct the prior false claim that these aliases execute legacy dedupe behavior.

The live base confirms the reported omission in `internal/app/sync_modes.go`: the two deduped legacy families return an empty `ContractMode` from both normal parsing (lines 54-55 and 58-59) and persisted-legacy parsing (lines 94-95 and 98-99). `RunETL` checks typed admission at `internal/app/app.go:1104`; the empty mode makes `IsContractMode()` false and execution continues into the ordinary legacy ETL branch.

The current base has an additional relevant guard: `IsContractMode` also excludes `LegacyCompatibility`. The implementation must therefore make only the two affected compatibility spellings parse as typed-contract aliases while preserving the three unrelated compatibility adapters.

## Locked decisions

- `full_refresh_overwrite_deduped` maps to `synccontract.ModeFullOverwrite`.
- `incremental_append_deduped` maps to `synccontract.ModeIncrementalDedupe`.
- Their alternative accepted spellings preserve the same canonical output name and typed mapping.
- The public mode-name-to-contract and capability decision has one connector-neutral authority in `internal/synccontract/public_modes.go`; persisted and normal parsing consume it through `internal/app/sync_modes.go`.
- The aliases take the existing typed execution/refusal path. With the ordinary scripted source fixture and no matching transport, that is the existing typed pre-I/O `ModeNotExecutableError`; no source read is allowed first.
- Existing canonical synccontract mode names retain their current parsed source, destination, contract, and typed-admission behavior.

## Delivery method

The repo-local Pi GSD adapter was validated with `scripts/gsd doctor`; all required commands were resolved with `scripts/gsd sources`. Compatible isolated GSD roles are unavailable in this worker and repository rules forbid role spawning, so `discuss-phase`, `plan-phase --tdd`, execution, verification, and review are recorded as an inline/manual fallback.

Required skills used: `golang-how-to`, `golang-cli`, `golang-testing`, `golang-error-handling`, `golang-security`, `golang-safety`, `golang-design-patterns`, `golang-structs-interfaces`, and `golang-documentation`.
