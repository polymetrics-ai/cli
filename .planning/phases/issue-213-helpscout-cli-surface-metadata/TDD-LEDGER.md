# TDD Ledger: Help Scout CLI Surface Metadata

Sub-issue: #213
Parent issue: #212

## Manual GSD Fallback

`programming-loop` is unavailable in `scripts/gsd`; manual GSD loop is active. Plan, red validation, green metadata, verification, and checkpoint commits are recorded here.

## Red Evidence Captured Before Production Edits

- Existing `internal/connectors/defs/help-scout/api_surface.json` was not full-surface: it listed 8 rows, while the official Inbox API docs navigation exposed 146 endpoint pages / 145 unique method-path pairs.
- Existing bundle had no `cli_surface.json`, so provider-style command/help metadata was absent.
- Existing docs claimed 4 stream-backed endpoint groups and broad out-of-scope exclusions; this conflicted with the full-surface safety target that sensitive/admin/destructive operations should become typed, blocked/gated operations rather than blanket exclusions.

Red command:

```bash
python3 - <<'PY'
import json, sys
api=json.load(open('internal/connectors/defs/help-scout/api_surface.json'))
current=len({(ep.get('method','').upper(), ep.get('path','')) for ep in api['endpoints']})
official=len({(line.split('\t')[0], line.split('\t')[1]) for line in open('/tmp/helpscout-endpoints.tsv')})
print(f'official_unique_endpoints={official}')
print(f'api_surface_endpoints={current}')
if current != official:
    print('RED: help-scout api_surface endpoint count does not match official Inbox API crawl')
    sys.exit(1)
PY
```

Result: failed as expected (`official_unique_endpoints=145`, `api_surface_endpoints=8`).

## Green Evidence

- Added `internal/connectors/defs/help-scout/cli_surface.json` with implemented stream-backed command metadata and planned/blocked direct-read, binary, reverse-ETL, admin, and destructive command families.
- Refreshed `internal/connectors/defs/help-scout/api_surface.json` to `operation_ledger_version: 1` with 145 unique normalized method/path rows.
- Preserved executable coverage only for current streams: `conversations`, `customers`, `mailboxes`, `users`.
- Replaced broad `out_of_scope` rows with blocked-by-default operation classifications.
- Updated connector docs and generated connector/website data.

Green commands:

```bash
python3 - <<'PY'
import json
api=json.load(open('internal/connectors/defs/help-scout/api_surface.json'))
current=len({(ep.get('method','').upper(), ep.get('path','')) for ep in api['endpoints']})
official=len({(line.split('\t')[0], line.split('\t')[1]) for line in open('/tmp/helpscout-endpoints.tsv')})
print(f'official_unique_endpoints={official}')
print(f'api_surface_endpoints={current}')
if current != official:
    raise SystemExit(1)
PY
jq empty internal/connectors/defs/help-scout/*.json internal/connectors/defs/help-scout/schemas/*.json
go run ./cmd/connectorgen validate internal/connectors/defs
go test ./cmd/connectorgen -run CLISurface
go test ./internal/connectors/engine -run CLISurface
go test ./cmd/connectorgen ./internal/connectors/engine
go test ./internal/connectors/conformance -run 'TestConformance/help-scout'
go build ./cmd/pm
./pm docs validate --connectors-dir docs/connectors
gofmt -w cmd internal
go vet ./...
go test ./...
make verify
go run ./cmd/connectorgen validate internal/connectors/defs
```

Results: all passed. Note: the first `go test ./...` attempt timed out at 600s after partial success; rerun with a 1200s timeout passed.

## CLI Help / Docs / Website Parity Evidence

```bash
./pm help docs
./pm docs generate --dir docs/cli
./pm docs validate --connectors-dir docs/connectors
cd website && pnpm run gen:website-data
cd website && pnpm run typecheck
./pm help connectors
./pm connectors
./pm connectors inspect help-scout --help
./pm connectors inspect help-scout --json
```

Results:

- `./pm help docs`: passed; confirmed generation/validation command shape.
- `./pm docs generate --dir docs/cli`: generated docs, but it rewrote many unrelated connector manuals in this checkout. Broad generated rewrites are outside #213 scope, so unrelated generated files were reverted and only Help Scout generated docs/catalog outputs were retained.
- `./pm docs validate --connectors-dir docs/connectors`: passed.
- `cd website && pnpm run gen:website-data`: passed; generated website connector data only changed Help Scout-related generated data files.
- `cd website && pnpm run typecheck`: blocked (`tsc: command not found`, `node_modules` missing). No dependency install was performed because new dependencies/install steps are human-gated.
- Runtime help checks passed and wrote line counts: `pm help connectors` 116 lines, `pm connectors` 116 lines, `pm connectors inspect help-scout --help` 112 lines.
- `pm connectors inspect help-scout --json`: passed and returned connector name `help-scout`; no credentials read.

## Refactor Notes

- Kept the canonical connector slug `help-scout`; did not create a duplicate `helpscout` bundle.
- Kept operation rows blocked by default. Later lanes own executable direct reads, binary policies, and reverse-ETL writes.
- Avoided adding `api_surface` refs to planned `cli_surface.json` commands because current validator only allows refs to executable covered rows.
- `api_surface.json` source URLs are per official docs page; the duplicate thread original source JSON/RFC 822 docs pages are recorded as a single method/path row with notes.
