# Plan — Issue #3992: execute authorized scheduled flow round trips

## Task Delivery Header

- Issue: `Closes #3992 — Schedule: execute an authorized warehouse-mediated connector round trip`
- Base branch: `integration/4015-mvp-flat-r1`
- Merges into: `integration/4015-mvp-flat-r1 -> main`
- Delivery: PR open against `integration/4015-mvp-flat-r1` with green checks.
- Working branch: `fm/cli-3992-schedule-authorization-r1`
- Task: Make an installed schedule actually fire an authorized warehouse-mediated round trip.
- Verification: `go test -timeout 20m ./internal/schedule/... ./internal/flow/... ./internal/app/...` plus green CI.

## Delivery method

`scripts/gsd doctor`, lifecycle sources, and `go run ./cmd/agentcontractgen check`
passed. `scripts/gsd prompt discuss-phase cli-3992-schedule-authorization-r1 --auto`
and `scripts/gsd prompt plan-phase cli-3992-schedule-authorization-r1 --tdd`
were generated, but the adapter cannot initialize this issue as a roadmap phase.
The canonical single-worker contract and unavailable compatible isolated Pi runtime
require the documented inline/manual fallback. This phase directory is the
replacement GSD evidence; no agent roles are spawned.

Loaded skills: `golang-how-to`, `golang-cli`, `golang-testing`,
`golang-error-handling`, `golang-security`, `golang-safety`,
`golang-design-patterns`, `golang-structs-interfaces`, `golang-context`,
`golang-concurrency`, and `golang-documentation`.

## Locked design

- Reuse `app.AuthorizationScope`/`AuthorizationRecord`; never use
  `reversePlanHash` as a standing authorization and do not add a second record.
- `schedule.Manifest` holds only `authorization_reference`. Rendered backend
  payloads use a quote-safe `pm schedule fire <name> --authorization <ref>`
  invocation; the fire command requires the reference to match the persisted
  manifest before forwarding it to `flow run`.
- A schedule firing has a persisted safe terminal state and receipt IDs. It
  acquires a per-schedule exclusive lease before execution; an active or
  incomplete firing is halted, not replayed. A flow error after attempting a
  run is parked with an error category only, never raw provider text.
- Scope/revocation/expiry checks remain in the completed connector flow-action
  runner, before connector validation, write, or read-back.
- No new dependency, credentialed provider run, generic writer, raw request,
  token, credential, or secret-derived material is added.

## TDD slices

1. **Red:** schedule tests require a non-empty authorization reference,
   inspectable safe state, quote-safe rendered reference, exclusive firing,
   and a terminal receipt result. CLI tests require an installed crontab
   payload to run unattended through the connector flow path; scope drift,
   revocation, and expiry must make zero connector events.
2. **Green:** add safe manifest and fire-state persistence, locking, parked
   outcome classification, removal cleanup, and render the `schedule fire`
   command on every supported backend.
3. **Green:** wire `schedule create|inspect|status|install|remove|fire`; fire
   validates its stored reference, calls the existing flow route without a
   token, and persists the terminal flow status plus opaque flow receipt IDs.
4. **Refactor/docs:** surface safe authorization/status fields in human and
   JSON output, docs/manual/website/golden artifacts; assert raw secret/token
   absence at manifest, rendered payload, state, and CLI-output boundaries.
5. **Verify/review:** targeted suites; `gofmt`, `go vet`, build and individual
   `make verify` gates; execute the documented inline `verify-work` and
   `code-review` fallbacks. CI remains the PR gate.

## Evidence table

| Requirement | Observable proof |
| --- | --- |
| identical scope fires without human token | isolated target write count increments through a rendered installed crontab command, then target read-back and opaque app receipt are present |
| changed scope stops before outbound request | target validation/write/read counters stay unchanged |
| revocation and expiry stay typed pre-dispatch refusals | each error chain has its distinct app reason and target event count is unchanged |
| no secret/token material persists or renders | tests search manifest, state, crontab payload, human and JSON output for fixture values |
| crash/overlap/ambiguous or rate-limit outcome does not replay | schedule state enters/retains halted or parked state, and a second fire leaves target write count unchanged |
| backend restoration is exact | test-file crontab is compared byte-for-byte before install and after remove |

## CLI help/docs/website parity

- [x] `pm schedule`, `pm help schedule`, and `pm schedule --help` render the new action summary.
- [x] `docs/cli/schedule.md`, website CLI reference, embedded manual, and generated help transcripts describe safe references and `--json`.
- [x] create/list/inspect/status/install/remove/fire output is verified in human and JSON forms.
