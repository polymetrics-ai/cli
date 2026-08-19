# Issue 3744 — Connector engine forced-redaction removal

## GSD setup

- GSD adapter preflight: `scripts/gsd doctor` passed.
- Resolved command sources: `discuss-phase`, `plan-phase --tdd`, `execute-phase`,
  `verify-work`, and `code-review`; `go run ./cmd/agentcontractgen check` passed.
- Manual-GSD fallback: this isolated worker applied the generated lifecycle prompts inline.
  The work is a single two-file engine slice and the canonical worker contract forbids
  separate mutating roles.
- Required skills: `golang-how-to`, `golang-design-patterns`,
  `golang-structs-interfaces`, `golang-error-handling`, `golang-security`,
  `golang-safety`, and `golang-testing`.

## Goal

Let secret-sensitive operations retain complete runtime content by allowing an optional
`sensitive_policy.redact_fields` declaration to be empty. Existing declarations continue
to load; all other sensitive-policy checks remain enforced.

## Scope

- `internal/connectors/engine/bundle.go`: remove only the forced non-empty
  `redact_fields` validation and correct its related diagnostic.
- `internal/connectors/engine/bundle_test.go`: test each sensitivity signal without
  redaction fields and retain the policy's input-mode, transform, and approval checks.

## Exclusions

- No connector JSON changes.
- No `direct_read.go` changes.
- No rest-write executor or commandrunner changes; those belong to the parallel write lane.
- No new dependencies, credentials, or live provider checks.

## Verification plan

1. Add the no-redaction loader regression and observe its red failure.
2. Remove the single forcing branch and run focused engine tests green.
3. Run engine and CLI tests, full vet/build, and each non-suite `make verify` gate.
4. Push the issue branch, open a PR to `main`, and leave GitHub CI as the remaining gate.
