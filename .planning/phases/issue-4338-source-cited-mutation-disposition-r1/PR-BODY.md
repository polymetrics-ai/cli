Refs #4338

## Intent

Add a closed, source-cited disposition for a provider mutation that exists in
the locked source but lacks a complete executable action. It deliberately keeps
the mutation visible as a runtime gap; it never manufactures an action,
contract, or command.

## What Changed

- Read an optional per-connector mutation-dispositions document during source
  import and attach a provider-cited runtime missing-foundation gap to the exact
  source operation.
- Accept that gap in projection and executable-coverage validation only when
  there is neither a complete action nor a matching `implemented` command
  claim.
- Keep mutation dispositions separate from read-only coverage: a read-only
  foundation cannot cover POST, PUT, PATCH, DELETE, or GraphQL mutation
  operations.
- Add real-path behavioural fixtures for distinct Asana absent-action, Jira
  incomplete-contract, Sentry SCIM PATCH/dashboard POST, and Vercel
  bulk/nested/destructive mutation forms, including a 159-operation Vercel-sized
  batch. Existing GitHub projection output remains byte-identical.

## TDD / GSD Delivery Record

- Red: the focused behavioural source-projection test failed before production
  code existed with undefined source-cited disposition symbols.
- Green: the same real descriptor marshal → source projection →
  executable-coverage path passes for the four-provider suite and closed-boundary
  controls. See `.planning/phases/issue-4338-source-cited-mutation-disposition-r1/TDD-LEDGER.md`.
- Lifecycle: `scripts/gsd doctor`, `scripts/gsd sources`, and generated prompts
  for `discuss-phase`, `plan-phase --tdd`, `execute-phase`, `verify-work`, and
  `code-review` were resolved. The isolated Pi worker runtime was unavailable,
  so the repository-authorized inline/manual fallback is documented in the
  phase evidence.
- Skills: `golang-how-to`, `golang-testing`, `golang-error-handling`,
  `golang-security`, `golang-safety`, `golang-design-patterns`, and
  `golang-structs-interfaces`.

## Verification

- `GOCACHE=<isolated-cache> go test -count=1 -timeout 20m ./cmd/connectorgen`
- `GOCACHE=<isolated-cache> go test -count=1 -timeout 20m ./internal/cli`
- `GOCACHE=<isolated-cache> go vet ./...`
- `GOCACHE=<isolated-cache> go build ./cmd/pm`
- `make tidy-check`, `make lint`, `make agent-contract-check`
- `make connectorgen-validate`, `make connectorgen-surface-sync`,
  `make connectorgen-operation-evidence`, `make github-parity-artifacts-check`
- certification, connector-boundary/canon/runtime-preflight, release, docs,
  and smoke checks (full results in `VERIFICATION.md`)
- `scripts/verify-gsd-workflow origin/main`

The repository directs per-command-timeout agents not to invoke `go test ./...`
or `make verify` as one process; affected package tests plus `internal/cli` and
each `make verify` gate ran serially under the unchanged 20-minute budget.

## Scope / Safety

- No connector definition or generated connector artifact is modified.
- No credentialed provider calls, external writes, or new dependencies.
- Vercel's 159 source mutations are not a blanket waiver: every operation that
  batch 1 intends to execute must receive a complete action. This disposition
  applies only to individually cited operations intentionally left
  non-executable.
- `origin/main` was fetched and merged immediately before push; it remained
  `e338cd301` and was already up to date.

## CLI help/docs/website parity

Not applicable: this changes internal `connectorgen` projection behaviour, not
the `pm` CLI command/help/manual/website surface. Generated connector and docs
checks are included above.

## Review route

Primary route: `claude_auto` for this non-draft, main-targeted PR opened by a
trusted repository author. Status at creation: pending automatic review. No
manual Claude/Copilot request unless the automatic route fails or is skipped.
