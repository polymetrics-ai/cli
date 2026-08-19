# Verification checklist — Issue 156 Zendesk Support parity

## Required local commands

- [x] `go run ./cmd/connectorgen validate internal/connectors/defs` — passed, 550 connectors, 0 findings.
- [x] Temp single-connector validation (`cp -R internal/connectors/defs/zendesk-support /tmp/zendesk-support-defs-validate/ && go run ./cmd/connectorgen validate /tmp/zendesk-support-defs-validate`) — passed, 1 connector, 0 findings. Note: the issue-listed direct connector-dir form currently treats `fixtures/` and `schemas/` as connectors and fails; no shared tool edit was made.
- [x] `go test ./internal/connectors/conformance -run 'TestConformance/zendesk-support' -count=1` — passed.
- [x] `go test ./internal/cli -run 'Connector|Dynamic|Golden' -count=1` — passed.
- [x] `go test ./cmd/connectorgen -count=1` — passed.
- [x] `go vet ./internal/connectors/... ./internal/cli/...` — passed.
- [x] `go build ./cmd/pm` — passed.
- [x] `go run ./cmd/pm docs validate --connectors-dir docs/connectors` — passed.
- [x] `make connector-boundary` — clean.
- [x] `git diff --check` — passed.
- [x] Read-only reviewer subagent re-review — no critical/warning findings after fixes.

## Non-live safety verification

- [x] No provider credentials requested or read.
- [x] No live Zendesk provider calls made.
- [x] No live writes, certification, VPS/Thaalam, merges, or dependency changes.
- [x] No behavior-changing shared connector execution-runtime edits; shared docs/website rendering surfaces changed only to expose connector-owned operation metadata, and shared Go edits were lint-only housekeeping.
- [x] Secret-shaped literal scan remains clean through connectorgen/conformance.
- [x] Destructive/delete operations are represented as in-scope blocked/typed metadata or existing typed write actions with `confirm: "destructive"`, not blanket-excluded as unsafe.

## CLI/help/docs parity notes

- [x] `cli_surface.json` validates as connector-owned command metadata.
- [x] `docs.md` records operation-ledger provenance, safety gates, blocked dependencies, and fixture-only certification status.
- [x] Generated connector docs with operation mappings, website connector data, and `internal/cli/testdata/golden_transcripts.json` were updated for the connector command surface. Broad docs generator drift unrelated to operation mapping was reverted.

## Parity checkpoint re-audit — 2026-08-01

- [x] Official source re-fetched: `https://developer.zendesk.com/zendesk/oas.yaml` with browser-style user agent; parsed as OpenAPI 3.0.3 with `info.version` 2.0.0 after quoting the provider document's bare `=` example value for YAML parsing.
- [x] Official operation count asserted: 434 paths, 625 unique method/path operations — GET 325, POST 111, PUT 89, PATCH 14, DELETE 86.
- [x] Local API-surface parity asserted: 631 endpoint rows = 625 official rows + 6 supplemental existing-bundle rows; 0 missing official rows, 0 stale official rows, 0 duplicate endpoint keys, 0 unclassified rows, 0 multi-disposition rows.
- [x] Executable dispositions asserted: 33 declared streams covered by 33 stream rows; 27 declared writes covered by 27 write rows; 0 unknown stream/write refs.
- [x] Blocked/planned dispositions asserted: 571 blocked operation rows exactly match 571 `operations.json` entries; 0 missing/stale operation metadata entries; 0 duplicate operation IDs.
- [x] Canonical command metadata asserted: 631 `cli_surface.json` commands reference 631 unique endpoint rows; 0 missing command refs, 0 stale command refs, 0 duplicate command paths, 0 duplicate endpoint refs.
- [x] Delete/destructive safety asserted: 86 official DELETE operations = 9 covered delete writes with `confirm: "destructive"` + 77 blocked typed operation rows; restore/recover rows are non-destructive admin updates, while remaining destructive, sensitive, and file-upload rows stay blocked pending typed action or shared executor foundations.
