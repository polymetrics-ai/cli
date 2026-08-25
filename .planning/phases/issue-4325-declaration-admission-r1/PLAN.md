# Plan — issue 4325 declaration-admission foundation

## Scope boundary

This is a shared tooling/schema PR. It does not convert a Batch-1 connector,
refresh a provider artifact, call a provider, add credentials, or perform a
write/delete. It must not weaken `commandrunner.Preflight`, `surface-sync`,
source-lock verification, runtime certification, or live certification.

## TDD execution slices

1. **Red — admission contract:** Add focused table-driven tests for a cited
   runnable read; deferred reverse-ETL write/delete; deferred binary
   download/upload; importer/descriptor gap; missing/duplicate/stale/base-path
   mismatch; false implementation; and an all-deferred zero-runnable bundle.
   The tests should fail because `declaration-admission` and its schema/type do
   not exist.
2. **Green — shared declaration checker:** Add the optional, versioned
   `sources/<connector>-declaration-admission.json` sidecar and a
   `connectorgen declaration-admission [defs-dir] [--json]` command. It checks
   only opt-in sidecars, deterministically cross-links source identity, lane,
   canonical endpoint, command, destructive/delete metadata, and runtime
   state. It never fetches provider data or requires source artifact bytes,
   hashes, request schemas, or fixtures.
3. **Green — explicit deferred command state:** Extend the command surface’s
   shared deferred/foundation metadata only as needed so an admitted deferred
   command stays discoverable and `commandrunner` returns a typed
   missing-foundation refusal before any executor. Keep implemented preflight
   rules unchanged.
4. **Refactor/document/gate:** Add the Make target and concise certification
   design/canon documentation distinguishing declaration admission from
   runtime and live certificates. Run formatting, targeted tests, relevant
   generator/check targets, review, and full feasible local verification.

## CLI docs parity

`connectorgen declaration-admission` is an internal generator command, not a
new `pm` command. The applicable docs are its `connectorgen` usage and the
connector certification/design canon. `pm help`, bare namespace behavior,
`docs/cli/**`, website pages, manual generation, and shell completion are not
applicable. Deferred connector commands retain normal `cli_surface.json`
discovery and are covered by commandrunner tests.

## Commit checkpoints

- Plan/TDD evidence checkpoint before production changes.
- Red-test checkpoint when the repository’s test convention permits it.
- Green implementation/documentation checkpoint after targeted gates.
- Review-fix checkpoint if inspection finds an actionable defect.
