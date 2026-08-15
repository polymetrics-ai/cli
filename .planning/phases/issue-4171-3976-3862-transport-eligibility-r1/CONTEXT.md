# Context — transport source eligibility club (#4171, #3862; #3976 deconflicted)

## Scope correction — 2026-08-16

PR 4175 owns #3976. This branch leaves PostgreSQL polling `planned`: `app.Open` composes the
outward transport but cannot bind its dynamic catalog object, while the attempted bind was inside
`ReadTransport` after authentication and catalog I/O. No PostgreSQL polling adapter or executable
contract belongs in this PR.

## Task Delivery Header

- Issue: Refs #4171 — GitHub transport stream eligibility; Refs #3862 — transport spine
  conformance. #3976 is deconflicted to PR 4175.
- Base branch: `integration/4015-mvp-flat-r1`
- Merges into: `integration/4015-mvp-flat-r1 → main`
- Delivery: A committed and pushed pull request open from `fm/cli-transport-stream-eligibility-club-r1` against `integration/4015-mvp-flat-r1`, with the API-reported base verified and focused/local gates green.
- Working branch: `fm/cli-transport-stream-eligibility-club-r1`
- Task: Close the GitHub admission and transport-spine gaps without weakening the closed registry:
  explicitly admit every executable GitHub stream and prove typed ineligibility refusals from the
  shipped entry path; PostgreSQL polling stays planned until PR 4175 proves a production bind.
- Verification: Production-composition red/green tests, focused Go package/race tests, built-`pm` live PostgreSQL and authenticated GitHub→warehouse→PostgreSQL evidence, individual repository gates, inline GSD verify/review, generated-artifact drift checks, and GitHub API base read-back.

## Evidence table

| Acceptance criterion | Evidence | Observable assertion or fake reason |
| --- | --- | --- |
| `github/commits` crosses the production transport into PostgreSQL | live | A built `pm` run uses the real GitHub credential and real PostgreSQL, then independent warehouse/target reads compare delivered rows with the 99,345-row extracted baseline. If the container cannot start, the exact unavailable proof is reported and never called green. |
| The GitHub allowlist remains positive and fail-closed | live | A production `app.Open` registry accepts declared `commits`, while a stream absent from the declaration returns a typed ineligible-stream error and records zero source requests, staged pages, target rows/sends, or checkpoint change. |
| PostgreSQL polling remains honestly planned | local | The unchanged real CLI inspection guard proves `status=planned`, a reason, and no executable polling contract. |
| Cancellation, replay, and interruption boundaries remain safe | fake/live | GitHub and orchestrator tests make typed/zero-effect assertions; the live GitHub proof remains pending under the stated runtime and credential limits. |
| The closed transport spine has declared conformance coverage | local | Shipped-entry production composition proves the GitHub API→database route; definition/evidence mismatch and undeclared stream cases remain typed pre-I/O refusals. |

Every live check must assert a state change that would be absent for a no-op. Exit status alone is not evidence.

## Locked decisions

- The task brief initially clubbed #4171, #3976, and #3862 into one direct PR. The later
  deconfliction directs #3976 to PR 4175; no fourth issue or opportunistic correction enters this
  branch.
- Preserve `synctransport.Registry` closure exactly: declarations select exact executor references, stream/action allowlists, modes, and externally admitted evidence. No capability-bit inference, wildcard relaxation for GitHub, or fallback to `Connector.Read` is allowed.
- GitHub source eligibility is an explicit list of executable bundle streams. The production declarative source adapter validates the selected stream again and emits bounded batches through the existing warehouse mediator. GitHub's mutating issue-label destination remains limited to its existing declared actions and singleton policy.
- `max_pages` on the GitHub transport source is a bounded navigation control: omitted means one provider page; a positive integer tightens the read; `0`, `all`, or `unlimited` means follow the declared paginator until exhaustion. This task does not add a generic unbounded transport control to other connectors.
- PostgreSQL polling remains distinct from CDC and stays `planned` here. Its explicit bootstrap/CDC path
  is unchanged; no polling fallback, native adapter, source/apply executor, or dynamic catalog
  binding is claimed by this PR.
- The attempted dynamic relation/cursor/key binding is not a preflight because it appears only
  inside `ReadTransport` after authentication and typed-catalog I/O. PR 4175 owns the correction;
  this branch preserves the plan-state guard rather than exposing a generic query or raw input.
- For the GitHub route, an ineligible stream still fails before downstream mutation with the typed
  error and zero source/stage/send/row/checkpoint effects.
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
