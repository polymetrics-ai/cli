# Issue #4411 — Asana seven-lane execution-proof plan

## Task Delivery Header

- Issue: Refs #4411 — Asana — direct/binary/ETL/reverse-ETL/sync proof and Foundation Atlas adoption
- Base branch: `codex/asana-track-b-projection-r1` at `c177998cda07c05a395aba677272d5e58afdb654`
- Merges into: `codex/asana-track-b-projection-r1` → `main`
- Delivery: Commit and push the scoped proof branch after focused checks are green; do not open a PR or merge.
- Working branch: `codex/asana-track-c-proof-r1`
- Task: Prove the Track B Asana definitions through embedded projection, registry/CLI credential boundaries, and local fake-provider execution. Preserve the 249-row source matrix and its dispositions without changing connector behavior.
- Verification: Deliberate invalid/missing-artifact proof, focused source/matrix projection tests, embedded-registry/CLI credential-boundary test, fake-provider DuckDB ETL test, existing focused direct/reverse/binary/sync execution tests, enabled-contract validation, JSON parsing, and `git diff --check`.

## Scope and source denominator

- Owned paths: `internal/connectors/defs/asana/*proof*_test.go` and this phase's planning/evidence files only.
- Not changed: shared runtime, foundations, generator/import code, Asana definition JSON, source inputs, or generated artifacts.
- Sole source denominator: the Track A/Track B `asana-source-lane-matrix.json` and its v3 source lock/descriptor. It preserves all 249 source IDs, not a second hand-maintained denominator.

| Lane | Source-cited dispositions to prove | Bounded claim |
| --- | --- | --- |
| Direct read | 119 implemented / 130 N/A | Valid source-bound CLI commands resolve through the embedded registry and stop at the missing-credential boundary before provider I/O; existing fake execution retains a source-locked path. |
| Direct write | 130 implemented / 119 N/A | Existing every-action fake-provider harness proves typed request plus approval/response. |
| Binary download | 249 N/A | The matrix/projection test rejects a fabricated artifact/backlink; no download capability is inferred. |
| Binary upload | 1 implemented / 248 N/A | The named attachment upload stays source-bound, approval-gated, and has existing fake execution through the typed write route. |
| ETL through DuckDB | 12 implemented / 52 mapped_unproven / 185 N/A | A local Asana fixture materializes the `tasks` stream to the local warehouse. The 52 candidate cells remain explicitly non-executable without source-backed scope/fanout plus a runtime artifact. |
| Reverse ETL from DuckDB | 130 implemented / 119 N/A | Existing bulk proof materializes a local table, previews, approves, and executes every named action against a local fixture. |
| Sync transport through DuckDB | 3 implemented / 246 N/A | Existing source-token fake execution proves the constrained Asana source transport; the descriptor remains source-only and does not claim API-to-API delivery. |

## Evidence table

| Acceptance criterion | Evidence | Observable assertion or fake reason |
| --- | --- | --- |
| Projection is the active embedded connector, not disk-only JSON | local runtime | `engine.Load(defs.FS, "asana")`, the bundle registry, and CLI dispatch expose the same source-bound command paths. |
| Direct-read/write/binary-upload commands stop before provider I/O without a credential | fake | A local transport spy is required because live Asana I/O and credentials are out of scope; valid command inputs must fail exactly at `missing --credential` with zero sends. |
| An implemented ETL stream reaches DuckDB | fake | A local `httptest` server is required because no live provider I/O is authorized. It serves one source-shaped `/tasks` record; the real app persists and reads it back from the local warehouse. |
| Direct write, reverse ETL, binary upload, and sync retain execution evidence | fake | Existing focused tests use local capture connectors/servers because live provider access is disallowed. Their fixtures assert the bounded typed request, approval, source token, checkpoint, or tombstone behavior rather than claiming certification. |
| Mapped-unproven, missing, and N/A cells cannot promote silently | local validation | Track B's copied-artifact tests deliberately remove a lane/backlink and promote a mapped-unproven ETL cell; each must fail before the real tree passes. |

## GSD and skills

- Resolved `gsd-discuss-phase`, `gsd-plan-phase`, `gsd-execute-phase`, `gsd-verify-work`, and `gsd-code-review` via `scripts/gsd sources` and generated the five prompts. This task is executed inline/manual because its assignment prohibits spawning runtime roles.
- `$connector-lane-build-order` governs the lock → matrix/projection → artifact → proof ordering.
- `$go-engineering` and the repository Go routing skills (`golang-how-to`, `golang-testing`, `golang-security`, `golang-safety`, `golang-error-handling`, `golang-design-patterns`, and `golang-structs-interfaces`) govern the focused Go test.

## Foundation Atlas disposition

- `runtime.direct-execution.v1`: reuse; the CLI boundary and existing direct-write execution are enough for this proof-only slice.
- `warehouse.stage-etl.v1`: reuse; the local DuckDB materialization witness exercises the existing implementation.
- `warehouse.reverse-etl.v1`: reuse; existing bulk plan → preview → approval → execution evidence remains applicable.
- `transport.sync-contract.v1` and `asana.event-token-source.v1`: reuse; existing source-token tests cover only the source-cited task event/hydration/snapshot scope.
- `source.projection-admission.v1`: reuse; the 52 `mapped_unproven` ETL cells remain visible as a source/projection admission gap, not executable streams.
- No actual missing runtime foundation is found. Consequently, no new demand record and no shared code are authorized.

## TDD and execution plan

1. Preserve the Track B structural red cases: remove a required artifact/backlink and promote one mapped-unproven ETL cell in a copied candidate; validation must fail without changing the real definitions.
2. Add one connector-local proof test that loads the embedded Asana bundle/registry and runs valid direct-read, direct-write, and binary-upload CLI forms to the credential boundary with a no-I/O transport spy.
3. In that test, use a local Asana `/tasks` fixture with the real app and existing definitions to materialize one ETL record into local DuckDB, then read it back under the connection owner.
4. Re-run the focused existing direct-read, direct-write, reverse-ETL, binary-upload, and source-token sync tests. Record their separate lane claims; do not relabel `mapped_unproven`, N/A, or missing cells.
5. Run inline verify/review, scope/diff checks, then commit/push only the green proof slice.
