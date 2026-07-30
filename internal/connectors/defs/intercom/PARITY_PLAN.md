# Intercom parity GSD plan

Issue scope: parent #164 and subissues #165-#171 for `internal/connectors/defs/intercom`.

## GSD command path and fallback

- Ran `scripts/gsd doctor` successfully on 2026-07-30.
- Ran `scripts/gsd prompt plan-phase issue-164-intercom-parity --skip-research` and used the generated Pi-adapter prompt inline.
- Attempted the mandatory implementation command `scripts/gsd prompt programming-loop init --phase issue-164-intercom-parity --dry-run`; the repo-local adapter returned `unknown GSD command: programming-loop` even though `scripts/gsd doctor` passed. Manual GSD programming-loop fallback is active for this connector-local slice.

## Required skills loaded

- `.agents/agentic-delivery/references/required-skills-routing.md`
- `gsd-core`
- `golang-how-to`
- `golang-cli`
- `golang-design-patterns`
- `golang-structs-interfaces`
- `golang-error-handling`
- `golang-security`
- `golang-safety`
- `golang-testing`
- `golang-documentation`
- `.agents/agentic-delivery/references/cli-help-docs-website-parity.md`

## Orchestration decision

- Cycle `plan`: `local_critical_path`.
- Reason: all seven Intercom subissues target the same connector-owned JSON/docs/fixture files, so mutating subissue workers would collide in this single disposable worktree. This worker will execute the connector-local critical path sequentially and will not edit shared runtime files.

## Safety gates

- No live Intercom credentials, provider calls, writes, certification, VPS/Thaalam work, merges, or default-branch pushes.
- GitHub issue-body updates use `gh-axi` only.
- Destructive/delete operations are not blanket-excluded as unsafe; when represented as write actions they must declare `confirm: "destructive"`, risk text, typed schemas, and the existing reverse-ETL plan -> preview -> explicit approval -> execute path.
- If a shared runtime foundation is missing for executing a documented operation safely, record the exact dependency in connector docs/metadata and keep the operation in the ledger without claiming it implemented.

## Implementation slices

1. Official-source inventory
   - Fetch and parse Intercom OpenAPI 2.16 and the Intercom REST page-data source named by #164.
   - Reconcile 230 documented operations with lane counts from #164 without changing issue counts.
2. Operation ledger
   - Replace the legacy 10-row Intercom `api_surface.json` with operation-ledger-mode rows for all official operations.
   - Keep existing implemented stream mappings honest and classify every other official operation as typed blocked/planned or connector-local executable only when supported.
3. Connector-local implementation metadata
   - Add/update `operations.json`, `cli_surface.json`, `writes.json`, `certification.json`, stream/schema/docs metadata as supported without shared runtime edits.
   - Mark destructive operations with typed destructive confirmation wherever write actions are declared.
4. Fixtures and conformance evidence
   - Preserve existing fixture-backed read coverage.
   - Add connector-owned write/direct fixtures only where the current engine can execute them safely without live provider calls.
   - Run focused conformance and validation gates; document skipped/blocked coverage honestly.
5. GitHub captain-policy addendum
   - Append an idempotent addendum to #164-#171 with `gh-axi` preserving bodies and counts.

## TDD ledger

| Slice | Red/validation before production edit | Green target |
| --- | --- | --- |
| Baseline | `go test ./internal/connectors/conformance -run 'TestConformance/intercom' -count=1` passed; `go run ./cmd/connectorgen validate internal/connectors/defs/intercom` fails because validate treats `fixtures/` and `schemas/` as connector dirs, so root validation will be used for final evidence. | Intercom conformance remains green; root validation has no Intercom findings. |
| Ledger completeness | Generated OpenAPI inventory reports 230 operations (GET 108, POST 67, PUT 23, PATCH 1, DELETE 31). | `api_surface.json` contains 230 official rows and no duplicate method/path pairs. |
| Destructive safety | Existing Intercom bundle has no write actions and excludes DELETE as unsafe/destructive. | Any implemented DELETE/destructive write action has `confirm: "destructive"`; unimplemented destructive operations are operation-ledger blocked/planned, not excluded as unsafe. |
| Docs/commands | Existing docs say read-only and only 5 endpoint groups. | Docs/command metadata describe implemented vs planned/blocked coverage without claiming live certification. |

## Verification checklist

- [ ] `node`/script inventory confirms 230 official Intercom operations and lane classifications.
- [ ] `go test ./internal/connectors/conformance -run 'TestConformance/intercom' -count=1`.
- [ ] `go run ./cmd/connectorgen validate internal/connectors/defs` (or summarized Intercom-scoped equivalent if full-root output is too large).
- [ ] `go vet ./internal/connectors/...` if connector-local generated JSON affects Go load paths.
- [ ] `go build ./cmd/pm`.
- [ ] `make connector-boundary`.
- [ ] `git diff --check`.
