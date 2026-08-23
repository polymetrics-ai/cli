## Intent

Refs #4333

Allow a v3 source document to state that the fixed query on its locked artifact
URL identifies the document, without widening source-import into a generic URL
or credential transport mechanism.

## What Changed

- Added v3 REST `artifact.identity_query:true`; absent/false stays no-query.
- Preserved the declaration through the cache to the HTTP fetcher.
- Reused bounded citation-query limits and credential-shaped-key rejection for
  identity queries.
- Kept batch materialization, legacy locks, GraphQL artifacts, and redirects
  on the default no-query policy.
- Documented the v3 source-lock authoring rule.

## Red / Green / Refactor

- **Red:** `TestSourceImportVersion3FetchesDeclaredIdentityQuery` failed before
  production edits because strict v3 decoding rejected `identity_query`.
- **Green:** behavioral tests prove the real lock parser/importer/cache/HTTP
  path sends locked `?version=`, provenance `?slug=` stays unfetched,
  absent/false declarations project identically, bounds and credential-shaped
  queries reject, and every URL/DNS guard remains active.
- **Refactor:** a default-deny internal URL policy and shared bounded-query
  validator keep the legacy parser/request entry points byte-identical.

## Verification

- `go test -timeout 20m ./cmd/connectorgen -count=1` — pass (167.414s), then
  pass again after merging current main (206.088s).
- `go test -timeout 20m ./internal/cli -count=1` — pass (697.549s).
- `go vet ./...`, `go build ./cmd/connectorgen`, `go build ./cmd/pm` — pass.
- `make tidy-check`, `make lint`, `make docs-check`, `make smoke-no-build`,
  `make agent-contract-check`, `make connectorgen-validate`,
  `make connectorgen-surface-sync`, `make connector-boundary`, and
  `make release-workflow-check` — pass.
- Focused identity-query regression suite and merge package vet/build — pass.

## GSD and Skills

- Generated and executed `discuss-phase`, `plan-phase --tdd`, `execute-phase`,
  `verify-work`, and `code-review` prompts inline; the task is not a numbered
  roadmap phase and the single-worker runtime has no compatible isolated-role
  execution.
- Loaded `golang-how-to`, `golang-cli`, `golang-testing`,
  `golang-error-handling`, `golang-security`, `golang-safety`,
  `golang-design-patterns`, `golang-structs-interfaces`, `golang-lint`, and
  `golang-documentation`.

## CLI / Docs Parity

No public command, flag, help topic, output, manual, website page, or generated
help surface changed. The source-lock contract is documented in
`docs/migration/conventions.md`, and the existing source-import help/docs
contract test remains green.

## Safety and Review

- No credentials, runtime-supplied request URL/query, connector-specific
  bypass, generic HTTP write path, or CodeQL suppression was added.
- Default HTTPS/userinfo/fragment/ordinary-host/public-IP/DNS/redirect/digest
  guards remain enforced and behaviorally tested.
- Inline security/code review found no actionable findings.
- Direct-PR delivery is authorized for #4130; `/no-mistakes` was not run.

## Review Route

Claude automatic review should run on PR open. No automated findings have been
received at publication time; any actionable finding will be dispositioned
before requesting human merge.
