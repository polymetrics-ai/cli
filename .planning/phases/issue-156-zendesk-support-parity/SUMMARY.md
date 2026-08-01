# Summary — Issue 156 Zendesk Support parity

## What changed

- Appended the captain-policy addendum to #156-#163 with `gh-axi`, preserving existing issue bodies and count tables.
- Converted `internal/connectors/defs/zendesk-support/api_surface.json` to operation-ledger v1.
  - Official Zendesk Support OAS 2.0.0 operations inventoried: 625.
  - Total api-surface rows: 631, including 6 supplemental `covered_by` rows for existing fixture-backed bundle surfaces not present in the Support OAS.
  - Existing covered executable surfaces preserved: 33 streams and 27 write actions.
  - Blocked official operation rows: 571, including direct-read/search, binary, admin, sensitive, and destructive/delete operations.
- Added connector-owned `operations.json` with 571 fixed typed operation contracts, all blocked by default.
- Added connector-owned `cli_surface.json` with 631 provider-style command metadata entries: 33 implemented ETL stream commands, 27 partial reverse-ETL write-plan commands, and 571 planned blocked operation commands.
- Updated connector docs/metadata, regenerated affected connector manual/skill operation mappings, refreshed website connector data, updated root golden transcripts needed by the CLI/help/docs surface, and applied lint-only shared Go cleanups.

## Safety notes

- No live Zendesk credentials, provider calls, writes, certification, VPS/Thaalam work, behavior-changing connector execution-runtime edits, new dependencies, merges, or pushes.
- Project memory hook attempted with `fm-ensure-agents-md.sh .`; it reported both `AGENTS.md` and `CLAUDE.md` are real files and require manual reconciliation. No AGENTS/CLAUDE edits were made because this task produced no broad project rule change beyond connector-local evidence.
- Destructive/delete operations are no longer blanket-excluded as unsafe. They are either existing typed delete write actions with `confirm: "destructive"` or blocked operation metadata pending connector-local typed schema/action work and the plan -> preview -> explicit approval -> execute path.
- Shared foundation gap recorded: `scripts/gsd prompt programming-loop ...` is unavailable in this adapter (`unknown GSD command: programming-loop`), so this used the manual GSD fallback plus `plan-phase` and `quick` prompts.
- Shared executor dependencies remain blocked, not claimed complete: provider direct/query/search (#2985), CDC truthfulness/state (#2986/#2988), and bounded binary/file executor/output policy. File-upload request contracts are recorded in connector-local operation descriptions while structured file-operation schema validation remains a shared dependency outside this worker scope.

## Parity checkpoint re-audit (2026-08-01)

- Re-fetched the official Zendesk Support OAS from `https://developer.zendesk.com/zendesk/oas.yaml` with a browser-style user agent and parsed OpenAPI 3.0.3 / `info.version` 2.0.0.
- Official operation totals remain 625 unique method/path rows across 434 paths: GET 325, POST 111, PUT 89, PATCH 14, DELETE 86.
- Local ledger totals after the checkpoint: 631 `api_surface.json` rows = 625 official + 6 supplemental existing-bundle rows; 571 blocked operation rows; 33 stream-covered rows; 27 write-covered rows; 0 missing official, 0 stale official, 0 duplicate endpoint keys, 0 unclassified, 0 multi-disposition.
- Command/operation parity remains exact: 571 `operations.json` entries match blocked rows, and 631 `cli_surface.json` commands reference 631 unique endpoint rows with 0 missing/stale refs.
- Delete/destructive safety remains explicit: 86 official DELETE operations = 9 covered delete writes with `confirm: "destructive"` + 77 blocked typed operation rows; restore/recover rows are classified as non-destructive admin updates; remaining sensitive/file-upload rows stay blocked pending typed sensitive handling or file-transfer foundations.

## Verification run

- `go run ./cmd/connectorgen validate internal/connectors/defs` — passed, 550 connectors, 0 findings.
- Temp single-connector validate (`/tmp/zendesk-support-defs-validate` parent dir) — passed, 1 connector, 0 findings. The issue-listed direct connector-dir form currently treats `fixtures/` and `schemas/` as connector dirs and fails; no shared tool edit was made.
- `go test ./internal/connectors/conformance -run 'TestConformance/zendesk-support' -count=1` — passed.
- `go test ./internal/cli -run 'Connector|Dynamic|Golden' -count=1` — passed.
- `go test ./cmd/connectorgen -count=1` — passed.
- `go vet ./internal/connectors/... ./internal/cli/...` — passed.
- `go build ./cmd/pm` — passed.
- `go run ./cmd/pm docs validate --connectors-dir docs/connectors` — passed.
- `make connector-boundary` — clean.
- `git diff --check` — passed.
- Read-only reviewer subagent — final focused re-review found no critical or warning findings after fixes for mutation classification, bulk spam destructive classification, and non-implemented command examples.
