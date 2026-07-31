# Verification — Asana connector parity (#380)

All completed checks below were credential-free and made no live Asana/provider calls.

## Planned credential-free gates

```bash
# Official source inventory comparison (pinned public OpenAPI source; no provider credentials)
python3 .tmp/asana_verify_inventory.py

# Definition validation (full defs root or temp single-connector root)
go run ./cmd/connectorgen validate internal/connectors/defs

# Dynamic fixture/conformance for Asana only
go test ./internal/connectors/conformance -run 'TestConformance/asana' -count=1

# CLI help/docs parity spot checks (help only; no credentials/provider calls)
go build -o .tmp/pm ./cmd/pm
.tmp/pm help asana
.tmp/pm asana
.tmp/pm asana tasks delete --help

# Issue-level local gates
go test ./internal/cli -run 'Connector|Dynamic|Golden' -count=1
go vet ./...
go build ./cmd/pm
make connector-boundary
git diff --check
```

## Results

| Gate | Result |
|---|---|
| Official pinned OpenAPI inventory vs `api_surface.json` | PASS: 249 official operations, missing=0, extra=0, duplicate=0. |
| `go run ./cmd/connectorgen validate internal/connectors/defs` | PASS: 549 connectors checked, 0 findings. |
| `go test ./internal/connectors/conformance -run "TestConformance/asana" -count=1` | PASS. |
| `go test ./internal/connectors/conformance -count=1` | PASS. |
| `go test ./internal/connectors/engine -count=1` | PASS. |
| `go build ./cmd/pm` | PASS. |
| `.tmp/pm docs validate --connectors-dir docs/connectors` | PASS. |
| Focused dynamic/golden/docs CLI suite | PASS with explicit non-certification test filter. |
| `go vet ./...` | PASS. |
| `make connector-boundary` | PASS, clean report. |
| `git diff --check` | PASS. |
| `go test ./internal/cli -run 'Connector|Dynamic|Golden' -count=1` | TIMEBOXED: certification CLI tests in the broad regex repeatedly load all connector bundles and exceeded 5m. Focused relevant CLI/docs/golden tests pass. |
| `go test ./...` | TIMEBOXED: default 10m package timeout in `internal/cli` and `internal/connectors/certify` certification sweeps; no live provider calls. |

## CLI/help/docs/website parity checklist

- [x] Runtime help: `pm help asana` renders the generated Asana connector manual.
- [x] Bare provider command: `pm asana` renders command group summary and exits 0.
- [x] Command help: `pm asana tasks delete --help` renders fixed-target reverse-ETL write help and includes `--confirm`.
- [x] Invalid actions remain usage errors through dynamic connector command routing.
- [x] `docs/connectors/asana/MANUAL.md` and `docs/connectors/asana/SKILL.md` regenerated from the Asana bundle.
- [x] CLI golden transcript fixture updated because Asana is now listed as a dynamic connector command.
- [x] Website docs not changed: this connector-local slice updates generated connector manuals/skill docs only; no website generator/catalog file was changed.

## Safety notes

- No live Asana calls, no credentials, no provider writes, no live certification, no no-mistakes shipping flow, no push, no PR.
- Destructive/delete operations are in scope only as typed, bounded, planned/blocked metadata or as named reverse-ETL actions with `confirm: "destructive"` and plan -> preview -> explicit approval -> execute.
- Generic shell, raw HTTP write, raw SQL write/read, arbitrary GraphQL, file, binary, and unrestricted passthrough tools remain disallowed.
