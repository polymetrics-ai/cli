Closes #4352

## Intent

Provide a shared, closed source-bound read execution foundation. A mapped source GET can become a normal credential-gated fixed operation only when its exact source identity, method, path, typed inputs, and executor semantics are declared.

## What Changed

- Added `source_operation` bindings (source ID + GET method + connector-relative path) to declarative operations and command metadata.
- Extended source projection to consider non-mutating GETs without changing mutation/write projection. It promotes only existing, exact declarations; it does not create a route or generic HTTP surface.
- Added no-network engine preflight for source-bound direct reads and stream ETL commands. Stream proof requires one declared composite, the exact source route, records/schema, and pagination.
- Added a stable `missing_foundation=source-bound-read-execution-r1:` disposition for incomplete typed read contracts before provider dispatch.
- Added a source-cited partial-coverage disposition for an already implemented
  mutation whose locked request contract still needs a named shared foundation.
  It requires exact source ID/method/path and a matching incomplete action; it
  cannot downgrade, invent, or conceal an operation.
- Captain correction `021.msg` supersedes the historical 9/100 partition:
  source import materializes all source-complete retained Asana GET contracts
  in their true executor lane—106 bounded direct reads and 12 exact
  record/schema/pagination-backed ETL streams. Only
  `asana.rest.getMembership` remains deferred, explicitly naming
  `cli-openapi30-reference-sibling-foundation-r1`.
- The ETL fan-out proof uses exact locked identities and locations:
  - `asana.rest.getProjectStatusesForProject` — `paths["/projects/{project_gid}/project_statuses"].get`
  - `asana.rest.getSectionsForProject` — `paths["/projects/{project_gid}/sections"].get`
  - `asana.rest.getStoriesForTask` — `paths["/tasks/{task_gid}/stories"].get`

The source locations above are from `data/connector-operation-mapping-reports/100-connectors/batch-1/asana.json`.

## TDD / GSD Evidence

- Lifecycle: inline/manual `discuss-phase` → `plan-phase --tdd` → `execute-phase` → `verify-work` → `code-review`, recorded in `.planning/phases/issue-4352-source-bound-read-execution-r1/`. Inline fallback was required because compatible isolated workers were unavailable and the canonical contract forbids role spawning.
- Red: source projection test failed before implementation because `get_access_requests` had no source binding.
- Green: source projection, route-substitution/no-dispatch, missing-foundation/no-dispatch, direct-vs-stream, and Asana preflight controls pass. Full evidence is in `TDD-LEDGER.md` and `VERIFICATION.md`.
- Skills: `golang-how-to`, `golang-cli`, `golang-testing`, `golang-error-handling`, `golang-security`, `golang-safety`, `golang-design-patterns`, `golang-structs-interfaces`.

## Testing

- `go test -timeout 20m -count=1 ./cmd/connectorgen` — passed in 150.671s
- `go test -timeout 20m -count=1 ./internal/connectors/engine`
- `go test -timeout 20m -count=1 ./internal/connectors/commandrunner`
- Focused Asana source-bound/reverse-ETL/delete and root-help/skill tests
- `go run ./cmd/connectorgen source-import asana --read-projection-only --check`
  — 249 operations verified
- `go run ./cmd/connectorgen validate internal/connectors/defs` — 553
  connectors, 0 findings
- `go run ./cmd/connectorgen surface-sync internal/connectors/defs --check` —
  553 connectors, zero drift
- `go vet ./...`; `go build ./cmd/pm`; `make tidy-check`; `make docs-check-no-build`; `make lint`; `make agent-contract-check`; `make connectorgen-validate`; `make connectorgen-surface-sync`; `make connector-runtime-preflight`; `make smoke-no-build`; `make connector-canon-check`; `npm --prefix website run typecheck`; `git diff --check`
- Built-binary preflight in a fresh project: `pm asana access-requests get-access-requests --json` reached `missing --credential` (exit 1), with no configured credential and no provider request.

`connector-boundary` and `release-installed-github-certification.sh` exceeded the local per-command runner limit. They are intentionally not claimed as passing; CI/PR remains the gate.

## CLI / Docs / Website

- Verified `pm help docs`, bare `pm asana`, `pm asana access-requests get-access-requests --help`, and `pm docs --help`.
- Refreshed generated Asana manual/skill, connector catalog, and website connector data. No top-level command-tree change was required.

## Safety / Scope

- No credentials were read, logged, or stored; no provider calls were made.
- No arbitrary URL, method, path, header, body, shell, HTTP write, SQL write, or curl escape hatch was added.
- Legacy ETL, direct-write, reverse-ETL, binary, and delete behavior is covered by the runner/Asana regressions.
- Every eligible historical Batch-1 planned GET is now generated from its
  concrete locked contract, not merely status-promoted. A future lock refresh
  must rerun source import/check and leave any non-scalar/header/body or absent
  typed contract deferred with a concrete named foundation; it must never use
  historical `planned` metadata to hide an otherwise capable source operation.
- The retained Asana mutation inventory is 21 absent non-executable actions,
  65 implemented reverse-ETL request-schema gaps, and 4 implemented delete
  path-parameter alias gaps. The 69 existing commands remain implemented;
  no provider behavior is claimed or invoked.

## Follow-up / Integration

- Do not merge from this task. Before integrating after certification/mapping foundations land, run fresh exact-head testing, an independent audit, and real connector-shaped evidence for Outreach and every other connector using this edge.
- Automated review route: Claude automatic review is expected on PR open. Confirm it covers the final commit range, disposition any actionable findings, and use Copilot only if Claude is unavailable.

## Checkpoints

- `80f36f8c1` — planning/TDD checkpoint
- `6f41f52ae` — implementation, generated artifacts, and verification evidence
