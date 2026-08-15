Refs #3992

The parent merge from `integration/4015-mvp-flat-r1` to `main` closes #3992.

## Intent

Allow an installed schedule to execute an already-authorized, warehouse-mediated
connector flow unattended, while making every retained authorization and receipt
datum safe to store or render.

## What Changed

- Add durable opaque authorization references and non-replay fire state to schedules.
- Execute scheduled flows through the connector-backed action route, retaining only safe receipt IDs.
- Surface authorization/status safely across schedule output, backend payloads,
  manuals, website docs, golden transcripts, and certification.

## Red / Green / Refactor

- Red: schedule creation and legacy certification could not supply the new durable
  authorization reference; `traces/red-run.txt` and
  `traces/certification-red.txt` retain the failing evidence.
- Green: schedule fire succeeds without a human token for an identical scope,
  invokes validate → preview → write → read, preserves a safe receipt, and
  restores the isolated scheduler backend byte-for-byte.
- Regression coverage: changed scope stops before a provider request; revocation,
  expiry, rate limit, crash/overlap, cleanup failure, and secret-output cases park
  or reject safely and never replay a non-idempotent action.

## Verification

- `go test -timeout 20m ./internal/schedule/... ./internal/flow/... ./internal/app/...`
- `go test -timeout 20m ./internal/cli -count=1`
- `go test -timeout 20m ./internal/connectors/certify -count=1`
- `go vet ./internal/connectors/certify ./internal/schedule/... ./internal/flow/... ./internal/app/... ./internal/cli`
- `go build ./cmd/pm`
- `make tidy-check`, `make lint`, `make docs-check-no-build`, `make agent-contract-check`, `make connectorgen-validate`, `make connectorgen-surface-sync`, `make connector-boundary`, `make release-workflow-check`, `make smoke-no-build`
- `npm --prefix website run typecheck`

## Lifecycle evidence

Manual inline GSD fallback is documented in `.planning/phases/issue-3992-schedule-authorization-r1/`; the adapter cannot initialize this issue-only work as a roadmap phase. Required skills: `golang-how-to`, `golang-cli`, `golang-testing`, `golang-error-handling`, `golang-security`, `golang-safety`, `golang-design-patterns`, `golang-structs-interfaces`, `golang-context`, `golang-concurrency`, and `golang-documentation`.

The recorded lifecycle ran `scripts/gsd prompt discuss-phase`, `plan-phase --tdd`,
`execute-phase`, `verify-work`, and `code-review`; the plan, ledger, verification,
summary, and review records are included in this PR.

## CLI parity

- Verified `pm help schedule`, `pm schedule`, and `pm schedule --help`.
- Updated `docs/cli/schedule.md`, website reference and architecture pages,
  generated docs, and the golden transcript.

## Safety

- Schedules, fire receipts, renderer payloads, and CLI output retain only the
  opaque `auth_<hex>` reference and non-sensitive receipt IDs—never approval
  tokens, credentials, or secret-derived material.
- Fire state prevents overlap and crash replay. Ambiguous, cleanup, and rate-limit
  outcomes park the schedule rather than replaying a potentially non-idempotent
  action.
- Flow execution still uses plan → preview → authorization → action; the schedule
  only obtains a run-scoped grant after the stored scope is re-derived.

## Delivery record

- Checkpoints: plan `add0557b1`; implementation and verification follow in this
  PR commit.
- Automated review: not requested—the task designates green CI as the gate.
- Follow-ups: none. No work from #4125, #4136, or #4090 is included.
