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
   - Recovery re-audit on 2026-08-01 found the current official OpenAPI 2.16 shared JSON now documents 231 operations. The stale local ledger omitted `POST /fin/csat` (`submitFinCsat`). Treat this as a required fix, not a waiver.
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
| Ledger completeness | Recovery re-audit script fetches the official OpenAPI 2.16 shared JSON and reports 231 operations (GET 108, POST 68, PUT 23, PATCH 1, DELETE 31); stale local state has 230 rows and omits `POST /fin/csat`. | `operations.json`, `api_surface.json`, `cli_surface.json`, `writes.json`, fixtures, and docs contain 231 official rows / 114 typed writes / 112 blocked-planned rows with no duplicate method/path pairs and no missing official operations. |
| Destructive safety | Current branch still has `update_ip_allowlist` without `confirm: "destructive"`. | All 40 destructive/admin-dangerous write actions, including 31 DELETEs and `update_ip_allowlist`, declare `confirm: "destructive"`; destructive/delete operations are canonical typed commands behind reverse ETL plan -> preview -> explicit approval -> execute. |
| Docs/commands | Generated connector manual/skill still say read-only/write=false and only 5 endpoint groups. | Connector-owned generated/manual docs describe implemented vs planned/blocked coverage without claiming live certification and reflect write=true, 231 operations, 114 writes, and 112 blocked/planned rows. |

## Verification checklist

- [x] Recovery `node`/script inventory confirms current official Intercom OpenAPI 2.16 has 231 operations and lane classifications (`OFFICIAL_INVENTORY.md`: 55 read, 114 write, 42 direct, 7 binary, 12 CDC/changefeed-like, 1 duplicate/not-applicable); no missing/extra local `operations.json` or `api_surface.json` rows.
- [x] `go test ./internal/connectors/conformance -run 'TestConformance/intercom' -count=1`.
- [x] `go run ./cmd/connectorgen validate internal/connectors/defs --json`.
- [x] `go test ./internal/cli -run 'Connector|Dynamic|Golden' -count=1`.
- [x] `go vet ./internal/connectors/... ./internal/cli/...`.
- [x] `go build ./cmd/pm`.
- [x] `make connector-boundary`.
- [x] `git diff --check`.
- [x] CLI help parity smoke checks: `./pm help intercom`, `./pm intercom --help`, `./pm intercom contacts list --help`, `./pm intercom ai content create content import source --help`, `.tmp/pm intercom fin agent submit a CSAT rating --help`, `.tmp/pm intercom ip allowlist update ip allowlist settings --help`.
- [x] `gh-axi` captain-policy addendum applied once to #164-#171; marker `intercom-captain-policy-addendum-destructive-r1` appears exactly once per issue.

## Recovery verification evidence (2026-08-01)

- Official OpenAPI 2.16 current total: 231 operations across 164 paths; method counts GET 108, POST 68, PUT 23, DELETE 31, PATCH 1.
- Connector totals after fix: 231 `operations.json` rows, 231 `api_surface.json` rows, 231 CLI command metadata rows, 114 write actions, 119 covered rows, 112 blocked/planned operation rows, 40 destructive confirmations, and 0 missing write fixtures.
- Added `POST /fin/csat` / `submitFinCsat` as typed write `submit_fin_csat`; updated `update_ip_allowlist` to typed destructive confirmation in write, operation, CLI, and generated docs metadata.
- Validation passed with isolated `GOCACHE=$PWD/.cache/go-build`: Intercom conformance, `connectorgen validate internal/connectors/defs --json`, targeted `go test ./internal/cli -run 'Connector|Dynamic|Golden|Docs'`, `go vet ./internal/connectors/... ./internal/cli/...`, `go build -o .tmp/pm ./cmd/pm`, `.tmp/pm docs validate --connectors-dir docs/connectors`, `make connector-boundary`, and `git diff --check`.
