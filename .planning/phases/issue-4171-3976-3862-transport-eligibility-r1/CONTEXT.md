# Context — transport source eligibility club (#4171, #3976, #3862)

## Task Delivery Header

- Issue: Refs #4171 — GitHub transport stream eligibility; Refs #3976 — PostgreSQL shared polling source adapter; Refs #3862 — transport spine conformance.
- Base branch: `integration/4015-mvp-flat-r1`
- Merges into: `integration/4015-mvp-flat-r1 → main`
- Delivery: A committed and pushed pull request open from `fm/cli-transport-stream-eligibility-club-r1` against `integration/4015-mvp-flat-r1`, with the API-reported base verified and focused/local gates green.
- Working branch: `fm/cli-transport-stream-eligibility-club-r1`
- Task: Close three source-side admission gaps without weakening the closed registry: explicitly admit every executable GitHub stream, bind PostgreSQL's native keyset reader to the shared polling executor and production transport path, and prove the resulting cross-family spine plus typed ineligibility refusals from the shipped entry point.
- Verification: Production-composition red/green tests, focused Go package/race tests, built-`pm` live PostgreSQL and authenticated GitHub→warehouse→PostgreSQL evidence, individual repository gates, inline GSD verify/review, generated-artifact drift checks, and GitHub API base read-back.

## Evidence table

| Acceptance criterion | Evidence | Observable assertion or fake reason |
| --- | --- | --- |
| `github/commits` crosses the production transport into PostgreSQL | live | A built `pm` run uses the real GitHub credential and real PostgreSQL, then independent warehouse/target reads compare delivered rows with the 99,345-row extracted baseline. If the container cannot start, the exact unavailable proof is reported and never called green. |
| The GitHub allowlist remains positive and fail-closed | live | A production `app.Open` registry accepts declared `commits`, while a stream absent from the declaration returns a typed ineligible-stream error and records zero source requests, staged pages, target rows/sends, or checkpoint change. |
| PostgreSQL resumable polling is routed through the shared executor | live | A built `pm` run against real PostgreSQL exercises definition selection, native PostgreSQL page fetch, shared polling page/checkpoint sequencing, warehouse stage/reopen, managed PostgreSQL apply/read-back, and persisted resume; exact IDs and checkpoint tuples are asserted across restart. |
| Cancellation, replay, ordering, schema/auth, concurrency, and interruption boundaries remain safe | live/fake | Real PostgreSQL tests cover observable database and checkpoint effects where deterministic; in-process timing/failure injection is used only for process-death windows that a live provider cannot deterministically pause, and each such case separately asserts typed error plus zero/unchanged effects. |
| The closed transport spine has declared conformance coverage | live | Shipped-entry production composition proves API→database and database→database routes; definition/evidence mismatch and undeclared stream cases remain typed pre-I/O refusals. |

Every live check must assert a state change that would be absent for a no-op. Exit status alone is not evidence.

## Locked decisions

- The task brief explicitly clubs #4171, #3976, and #3862 into one direct PR. No fourth issue or opportunistic correction enters this branch.
- Preserve `synctransport.Registry` closure exactly: declarations select exact executor references, stream/action allowlists, modes, and externally admitted evidence. No capability-bit inference, wildcard relaxation for GitHub, or fallback to `Connector.Read` is allowed.
- GitHub source eligibility is an explicit list of executable bundle streams. The production declarative source adapter validates the selected stream again and emits bounded batches through the existing warehouse mediator. GitHub's mutating issue-label destination remains limited to its existing declared actions and singleton policy.
- `max_pages` on the GitHub transport source is a bounded navigation control: omitted means one provider page; a positive integer tightens the read; `0`, `all`, or `unlimited` means follow the declared paginator until exhaustion. This task does not add a generic unbounded transport control to other connectors.
- A PostgreSQL polling adapter uses the shared `engine.PollingSourceExecutor`; native PostgreSQL owns catalog binding, lossless cursor encoding, composite-primary-key tie encoding, lexicographic query rendering, and traversal validation. The shared executor continues to own page sequencing and advances only after warehouse/destination acknowledgement and App checkpoint persistence.
- Polling remains distinct from CDC. PostgreSQL's existing explicit bootstrap/CDC path is preserved; polling is selected only by the implemented polling declaration and its supported modes, never as fallback for `change_capture`.
- Dynamic PostgreSQL relation identity, credential/account scope, cursor field, schema fingerprint, and primary-key tuple are bound from the already-resolved connection/catalog request by the named native adapter. No raw SQL, table expression, DSN, credential, generic query, or caller-authored checkpoint reaches the shared executor.
- Out-of-order or non-advancing native pages, null/lossy cursors, missing stable primary keys, schema/key drift, auth refusal, stale source generation, and ineligible streams fail before downstream mutation. Refusal tests assert the typed error and zero/unchanged rows, sends/stages, and checkpoint.
- #4125, #4158, and #4169 are explicit non-goals. Any other defect is recorded as `needs-decision:` and left unchanged.

## Existing foundation reused

- Definition-owned loading/composition: `internal/connectors/{sync_transport.go,engine/bundle.go}` and `internal/synctransport/definition_composition.go`.
- Production entry and warehouse path: `cmd/pm` → `internal/cli` → `internal/app` → `internal/synctransport.Orchestrator` → connection warehouse stage → PostgreSQL managed target.
- Shared polling admission/execution: `internal/connectors/engine/{polling_preflight.go,polling_source.go}`.
- PostgreSQL typed catalog/read policy: `internal/connectors/native/postgres/{typed_catalog.go,transport_source.go}` and `internal/connectors/database`.
- Managed PostgreSQL destination and durable evidence: `internal/connectors/native/postgres/transport_destination.go` plus the existing workset/receipt path.

## GSD and skill execution mode

`scripts/gsd doctor`, all lifecycle source resolutions, and `go run ./cmd/agentcontractgen check` passed. `discuss-phase --auto` is executed inline because the user required autonomous work and the canonical single-worker contract forbids lifecycle-role spawning. The generated prompts retain their TDD, verification, and review gates.

Loaded skills: `golang-how-to`, `golang-cli`, `golang-design-patterns`, `golang-structs-interfaces`, `golang-error-handling`, `golang-security`, `golang-safety`, `golang-testing`, `golang-context`, `golang-concurrency`, and `golang-database`, plus `gsd-discuss-phase`, `gsd-plan-phase`, `gsd-execute-phase`, `gsd-verify-work`, and `gsd-code-review`.

