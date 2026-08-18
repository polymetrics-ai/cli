Refs #4015

## Summary

- certifies 29 GitHub mutation commands from writes-b slice 2 with live schema-v2 evidence
- records one honest, disjoint outcome for all 146 assigned commands: 29 certified, 1 no-object, 22 entitlement, 1 not-implemented, 88 product defects, and 5 captain escapes
- completes independent produced-value proof and provider-terminal cleanup for every certified mutation
- preserves the fleet-level `integer_id_scientific_notation` defect classification without redundant controls
- does not regenerate shared certification artifacts

## GSD / TDD evidence

- lifecycle artifacts: `.planning/phases/github-mut-slice2-writes-b/`
- inline/manual GSD fallback used because the task forbids role spawning
- Red: every mutation assertion rejected a missing, unchanged, or plausibly wrong provider value
- Green: independent GitHub read-back matched the planned values, then direct provider cleanup and an independent terminal read-back proved containment
- review: inline evidence review found no actionable findings; automatic Claude review is left to the repository PR trigger
- required skills recorded in `PLAN.md`: `golang-how-to`, `golang-cli`, `golang-testing`, `golang-error-handling`, `golang-security`, `golang-safety`

## Testing

- `go test -timeout 20m ./cmd/connectorgen`
- `go vet ./cmd/connectorgen`
- `go run ./cmd/connectorgen certification-matrix --check`
- `go run ./cmd/connectorgen surface-sync --check`
- `go run ./cmd/agentcontractgen check`
- `scripts/verify-gsd-workflow`
- `git diff --check`

All passed. The full 550+ connector suite and `make verify` are deferred to CI as directed by `AGENTS.md` for per-command-timeout agents.

## Safety

- all live effects stayed within `Polymetrics-Cert`, its private disposable repositories, and the certification user
- credentials came from Keychain and appear in evidence only as repository-salted fingerprints
- commands 120–123 issued no request because they would create public organization visibility
- command 142 issued no request because autonomous hosted-agent cost was genuinely unknowable
- no PR is merged by this branch

## CLI/docs parity

Not applicable: this PR changes certification evidence and planning records only; it does not change CLI behavior, command surfaces, help, docs, or website content.
