# PostgreSQL certification profile plan

## Task Delivery Header

- Primary issue: Fixes #4192 — execute database transport and fail unexecutable declared stages
- Parent issue: Refs #4015 — Production MVP — certification
- Base branch: `integration/4015-mvp-flat-r1`
- Merges into: `integration/4015-mvp-flat-r1 → main`
- Delivery: One verified PR against `integration/4015-mvp-flat-r1`, combining the authorized connector-neutral certification foundation with PostgreSQL's definition-owned profile and live proof.
- Working branch: `fm/cli-postgres-certification-profile-r1`
- Task: Certify PostgreSQL's genuine database-native catalog/schema discovery, typed reads, polling-watermark source, managed-target destination/apply strategies, and admitted sync modes. The approval authorizes the generic transport-certification/evidence foundation in this same PR; `allStagesPassed` is exclusively owned by this delivery.
- Verification: Targeted unit and integration tests; a current-SHA binary certification against the shared PostgreSQL container; a deliberate post-compilation sabotage that turns the certificate red then restores green; certification-matrix byte-stability; connector validation/boundary, docs generation, lint, and repository gates.

## Evidence Table

| Acceptance criterion | Evidence | Observable assertion or fake reason |
| --- | --- | --- |
| PostgreSQL has a definition-owned certification profile | pass | The loaded postgres bundle exposes bounded `read_limit=100` and container `sslmode=disabled`; removing its profile made the focused loader test fail. |
| `pm certify` runs PostgreSQL stages against a live database | pass | A built current-SHA binary creates an ephemeral certification project, connects to a test-owned container, catalogues the seeded table, and reads all five typed rows. |
| Managed-target apply strategies and all admitted modes are certified | pass | The built `pm` binary executes PostgreSQL's exact `postgres_polling_watermark → postgres_managed_target` pair for all six modes, plans/previews/approves every run, and independently reads every managed target relation. |
| Results cannot pass without execution | pass | A declared pair with no adapter records `status=unexecutable` and fails the terminal roll-up; only explicit environmental/safety `status=skipped` stages remain benign. The schema-valid scratch `sslmode=bananas` run went red and restoration went green. |
| Matrix flags are truthful and stable | pass | The live run emitted 12 salted, proof-bearing database records: read and managed-target write for each of six modes. Two regenerations were byte-identical and only those cells are `live_tested=true`. |
| API-shaped irreducible stages are honest | pass | No REST direct-read candidate, binary candidate, or `writes.json` pairing was invented. PostgreSQL's unselected generic direct-write stages remain explicit `skipped` results with a concrete database-transport reason. |

## Scope and ownership guard

- Owned paths: `internal/connectors/defs/postgres/**`, PostgreSQL-native integration tests, the roll-up/report/evidence foundation, and the exact PostgreSQL polling-to-managed-target adapter. PostgreSQL identifiers are permitted only where binding the declared `postgres_polling_watermark → postgres_managed_target` pair or its dynamic-watermark certification semantics; matrix promotion otherwise remains generic.
- Not owned: GitHub read coverage, other connector definitions, broker/MCP/UI work, dependencies, generated hook/native imports, and PostgreSQL schema migrations.
- Every shared change requires an explicit review in `VERIFICATION.md` that it reads definition-owned data and does not hard-code PostgreSQL.

## TDD slices

1. **Profile model and validation (Red → Green, complete).** The absent-profile loader test failed, then PostgreSQL-owned bounded defaults made it green without any provider-specific shared Go condition.
2. **Live PostgreSQL source certificate (Red → Green, complete).** An opt-in `databaseintegration` test builds the current `pm` binary, runs the PostgreSQL profile against an owned container, independently verifies five seeded source rows, and checks the report's catalog/read results.
3. **Failure control (Red → Green, complete).** After `connectorgen validate` accepted the scratch profile, `sslmode=bananas` caused the built binary to exit 2. The committed default was restored and the identical run became green.
4. **Roll-up classification (Red → Green, complete).** Make an unexecutable declared transport stage a precise non-pass report result and terminal certification failure, while preserving only genuine environmental/safety skips. Exercise the allowlisted GitHub and PostgreSQL radius.
5. **Database transport foundation (Red → Green, complete).** Execute PostgreSQL's declared source/destination pair and all six modes against an isolated live database, carrying the caller-selected dynamic-table watermark and independently inspecting target/checkpoint state. This is an exact PostgreSQL adapter, not a generic database framework.
6. **Matrix evidence foundation (Red → Green, complete).** Accept only a passing PostgreSQL report with all six target/read/checkpoint proofs, write redacted database evidence, and mark only those matching cells. Regeneration is byte-stable.
7. **Refactor/review.** Remove scaffolding not needed by the profile, run review, and keep documentation/help parity explicitly marked as unchanged or updated.

## GSD execution record

Resolved command path: `scripts/gsd doctor`; `scripts/gsd sources {discuss-phase,plan-phase,execute-phase,verify-work,code-review}`; then the generated `scripts/gsd prompt … 4015 --auto` prompts. The runner has no compatible isolated GSD-role runtime and the task prohibits waiting for human input, so discuss, planning, execution, verification, and review are performed inline with this artifact set.

Required skills loaded: `golang-how-to`, `golang-cli`, `golang-design-patterns`, `golang-structs-interfaces`, `golang-error-handling`, `golang-security`, `golang-safety`, `golang-testing`, `golang-database`, and `golang-documentation`.

## CLI help/manual/website parity

The existing `pm connectors certify` surface is unchanged, but its PostgreSQL boundary is user-visible. The manual and website now state that PostgreSQL full certification requires `--write`, an explicit `--stream`, and `cursor_field`, certifies one dynamic relation plus all six declared modes, and does not claim the whole database or direct writes.
