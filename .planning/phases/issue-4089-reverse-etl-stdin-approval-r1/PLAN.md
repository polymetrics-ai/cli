---
phase: issue-4089-reverse-etl-stdin-approval-r1
plan: 01
type: tdd
status: planned
---

# #4089 — carry reverse-ETL approval through bounded stdin

**Issue:** [#4089](https://github.com/polymetrics-ai/cli/issues/4089)
**Parent:** [#3988](https://github.com/polymetrics-ai/cli/issues/3988)
**Branch:** `fm/cli-4089-reverse-etl-stdin-approval-r1`

## Manual GSD lifecycle

The adapter doctor, all five required command sources, and `go run ./cmd/agentcontractgen check` passed. The following prompts were resolved through `scripts/gsd prompt`: `discuss-phase 4089 --auto`, `plan-phase 4089 --tdd --skip-research`, `execute-phase 4089 --interactive`, `verify-work 4089 --auto`, and `code-review 4089 --files=...`.

The issue is outside the numeric roadmap and repository policy forbids the roles those prompts would spawn, so this worker executes each step inline.

## Scope and boundaries

- Reuse `readApprovalTokenFromStdin` as the only token reader.
- Replace both reverse-ETL argv call sites in `internal/cli/cli.go`.
- Reject the retired argv channel explicitly, without echoing its value.
- Preserve app-layer approval hash validation, persisted-preview validation, destructive confirmation, and one-time consumption.
- Update runtime manual, generated CLI manual, skills, website source and generated website documentation; remove stale argv examples.
- Do not add a connector-specific literal, new dependency, approval-model redesign, generic write surface, or external mutation.

## TDD slices

### Slice 1 — carrier contract (RED)

Red: add a real-binary CLI test that requires the bare stdin marker, supplies the token only by a pipe, and fails against the current argv implementation.

Red: cover empty, over-limit, multiline/malformed, valued-marker, retired-argv, and replay input; each rejection must observe zero destination-side effects.

Red: assert six distinct secret surfaces: argv (with the live process command line recorded), environment, durable project files, captured logs, receipts, and emitted evidence.

### Slice 2 — minimal generic wiring (GREEN)

Green: add one generic helper that validates the bare marker, refuses the retired argv path, and delegates byte validation solely to `readApprovalTokenFromStdin(os.Stdin)`.

Green: wire its returned token into both existing `RunReverseETLRequest` construction paths before `RunReverseETL`; do not change the app contract.

Green: update existing CLI tests to use bounded stdin while retaining their approval and replay assertions.

### Slice 3 — parity and refactor

Green: update `internal/cli/docs.go`, regenerated `docs/cli/**`, generated skills, website source, generated website data, and transcript fixtures so the visible syntax requires the stdin marker and one bounded standard-input line.

Refactor: format the small helper, remove only now-dead argv validation, and run a repository-wide stale-syntax scan plus targeted security review.

## Required skills

`github-issue-first-delivery`, `golang-how-to`, `golang-cli`, `golang-testing`, `golang-error-handling`, `golang-security`, `golang-safety`, `golang-design-patterns`, `golang-structs-interfaces`, `golang-documentation`, `vercel-react-best-practices`, `vercel-composition-patterns`, `gsd-discuss-phase`, `gsd-plan-phase`, `gsd-execute-phase`, `gsd-verify-work`, `gsd-code-review`, and `no-mistakes`.

## Verification plan

```text
go test -count=1 -timeout 20m -run '^TestReverseETLApprovalUsesBoundedStdin$' -v ./internal/cli
go test -count=1 -timeout 20m ./internal/cli
go test -count=1 -timeout 20m ./internal/app
go vet ./internal/cli ./internal/app
go run ./cmd/pm help reverse
go run ./cmd/pm reverse
go run ./cmd/pm reverse --help
go run ./cmd/pm docs generate --dir <temporary-dir>
rg -n -- '--approve' internal/cli docs/cli docs/skills website/content website/lib/docs.generated.ts
make tidy-check
make lint
make docs-check
make agent-contract-check
make connectorgen-validate
make connectorgen-surface-sync
make connector-boundary
make release-workflow-check
```
