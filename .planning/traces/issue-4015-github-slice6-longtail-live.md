# Issue #4015 — GitHub slice 6 longtail live certification

## Delivery scope

- Branch: `fm/cli-cert-slice6-longtail`
- Base: `origin/integration/4015-mvp-flat-r1`
- PR: #4222
- Scope: live schema-v2 evidence only; no connector behavior, generated matrix, sweep, candidates, help, docs, or website surface changed.
- Lifecycle: continuation of issue #4015's landed GSD/TDD evidence foundation from PR #4219. The live evidence lane ran inline because the canonical parent contract forbids spawning another delivery role and the task assigned one exact disjoint slice.

## Required skills

- `golang-how-to`
- `golang-design-patterns`
- `golang-structs-interfaces`
- `golang-error-handling`
- `golang-security`
- `golang-safety`
- `golang-testing`

## TDD and live-proof ledger

- Red: the slice work list contained direct-read commands without accepted schema-v2 live evidence. Each candidate was executed against the disposable GitHub identity, and plausible incorrect produced values (for example a response count inconsistent with the returned array) were checked to ensure the agent-derived assertion rejected them.
- Green: accepted executions wrote repository-salted, credential-fingerprinted schema-v2 records; every added record passed `go run ./cmd/connectorgen certification-matrix --check` before the lane advanced.
- Provider control: the three Docker migration conflict endpoints returned HTTP 400 with `Package migration for docker is no longer supported. Please contact support for assistance.` They are classified as provider-deprecated/entitlement, not product defects.
- Surface finding: three implemented commands target a retired provider capability and should be removed from the command surface or declared unsupported.
- Product defect: `issues assignees view` cannot accept GitHub's successful HTTP 204 response and reports `operation direct read response is not JSON: EOF`; the direct provider control returned 204 with an empty body.

## Coverage and cleanup

- Branch contribution over the integration base: 55 evidence files.
- Deduplicated slice coverage: 45 commands with passing evidence, 23 `no_object`, 3 provider-deprecated/entitlement, and 88 mutations paused by captain direction; total 159 commands.
- The 55 files cover 45 distinct slice commands because ten commands were independently re-executed and retained under unique run-scoped filenames.
- No fixture was created in the resumed read-only pass. Read-back showed zero gists, issue comments/assignees, packages, private registries, enterprise teams, pull requests, and workflows for the disposable scope.

## Verification

- `go test -timeout 20m ./cmd/connectorgen` — PASS
- `go run ./cmd/connectorgen certification-matrix --check` — PASS
- `go run ./cmd/agentcontractgen check` — PASS
- `bash scripts/verify-gsd-workflow origin/integration/4015-mvp-flat-r1` — initially identified this missing branch-local trace; rerun after adding the trace is required before handoff.
- `git diff --check origin/integration/4015-mvp-flat-r1...HEAD` — PASS
