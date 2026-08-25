Closes #4352

## Intent

Provide a shared, closed source-bound read execution foundation. A mapped source GET can become a normal credential-gated fixed operation only when its exact source identity, method, path, typed inputs, and executor semantics are declared.

## What Changed

- Added `source_operation` bindings (source ID + GET method + connector-relative path) to declarative operations and command metadata.
- Extended source projection to consider non-mutating GETs without changing mutation/write projection. It promotes only existing, exact declarations; it does not create a route or generic HTTP surface.
- Added no-network engine preflight for source-bound direct reads and stream ETL commands. Stream proof requires one declared composite, the exact source route, records/schema, and pagination.
- Added a stable `missing_foundation=source-bound-read-execution-r1:` disposition for incomplete typed read contracts before provider dispatch.
- Materialized three bounded Asana direct-read controls plus an exact-source-bound existing workspace ETL control:
  - `asana.rest.getAccessRequests` — `paths["/access_requests"].get`
  - `asana.rest.getAgentsForWorkspace` — `paths["/workspaces/{workspace_gid}/agents"].get`
  - `asana.rest.getAgent` — `paths["/agents/{agent_gid}"].get`
  - `asana.rest.getWorkspaces` — `paths["/workspaces"].get`

The source locations above are from `data/connector-operation-mapping-reports/100-connectors/batch-1/asana.json`.

## TDD / GSD Evidence

- Lifecycle: inline/manual `discuss-phase` → `plan-phase --tdd` → `execute-phase` → `verify-work` → `code-review`, recorded in `.planning/phases/issue-4352-source-bound-read-execution-r1/`. Inline fallback was required because compatible isolated workers were unavailable and the canonical contract forbids role spawning.
- Red: source projection test failed before implementation because `get_access_requests` had no source binding.
- Green: source projection, route-substitution/no-dispatch, missing-foundation/no-dispatch, direct-vs-stream, and Asana preflight controls pass. Full evidence is in `TDD-LEDGER.md` and `VERIFICATION.md`.
- Skills: `golang-how-to`, `golang-cli`, `golang-testing`, `golang-error-handling`, `golang-security`, `golang-safety`, `golang-design-patterns`, `golang-structs-interfaces`.

## Testing

- `go test -timeout 20m ./internal/connectors/engine -count=1`
- `go test -timeout 20m ./internal/connectors/commandrunner -count=1`
- `go test -timeout 20m ./internal/connectors/defs/asana -count=1`
- Focused `cmd/connectorgen` source-projection tests
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
- This does **not** claim all Batch-1 reads: three of the stated baseline 100 planned Asana GETs are materialized here; 97 still need exact source/typed-contract materialization. Unsupported rows stay declaration-pending or named `missing_foundation`.

## Follow-up / Integration

- Do not merge from this task. Before integrating after certification/mapping foundations land, run fresh exact-head testing, an independent audit, and real connector-shaped evidence for Outreach and every other connector using this edge.
- Automated review route: Claude automatic review is expected on PR open. Confirm it covers the final commit range, disposition any actionable findings, and use Copilot only if Claude is unavailable.

## Checkpoints

- `80f36f8c1` — planning/TDD checkpoint
- `6f41f52ae` — implementation, generated artifacts, and verification evidence
