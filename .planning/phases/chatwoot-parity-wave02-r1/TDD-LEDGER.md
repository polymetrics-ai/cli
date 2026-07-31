# TDD LEDGER — Chatwoot parity wave02 r1

## Red

- Current `internal/connectors/defs/chatwoot/api_surface.json` failed official parity before production edits: 71 endpoint rows vs 145 official operations, no `operation_ledger_version=1`, and 58 legacy `excluded` rows.
- Initial generator attempts also produced expected red/repair evidence:
  - Direct `connectorgen validate internal/connectors/defs/chatwoot --json` treats nested folders as connectors; validation must use a temp parent copy.
  - `operations.json` with unsupported kind `stream_etl` failed schema validation.
  - Newly implemented write CLI commands failed conformance when flags were not scalar-safe.
  - `minimum` in generated JSON Schema failed supported-schema conformance.

## Green

- Regenerated Chatwoot-owned bundle now represents all 145 official Chatwoot OpenAPI operations exactly once.
- Declared 7 fixture-backed streams, 73 reverse-ETL write actions, and 65 planned/blocked operation-ledger rows.
- All 18 DELETE operations are named write actions; DELETE/destructive/admin/elevated actions carry `confirm: destructive` and CLI approval text requiring `--confirm destructive`.
- Targeted gates passed:
  - Temp parent `go run ./cmd/connectorgen validate <tmp-with-chatwoot> --json`
  - Full `go run ./cmd/connectorgen validate internal/connectors/defs --json`
  - `go test ./internal/connectors/conformance -run 'TestConformance/chatwoot' -count=1`
  - `go test ./internal/connectors/conformance -count=1`
  - `go test ./internal/connectors/engine -count=1`
  - `go test ./internal/cli -run 'Connector|Dynamic|Golden' -count=1`
  - `go run ./cmd/pm docs validate --connectors-dir docs/connectors`

## Refactor

- Kept generation deterministic in `.planning/phases/chatwoot-parity-wave02-r1/generate_chatwoot_defs.py` so Chatwoot-owned JSON/docs/fixtures can be reproduced.
- Preserved connector-local scope; no shared runtime/foundation code changed.

## Red evidence

```bash
python3 - <<'PY'
# compared official-operations.json against current api_surface.json
PY
```

Result before production edits:

```text
api_surface rows 71 != official operations 145
operation_ledger_version=None
legacy excluded rows=58
```
