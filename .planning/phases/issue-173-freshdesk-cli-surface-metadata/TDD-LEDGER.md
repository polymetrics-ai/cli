# TDD Ledger: Freshdesk CLI Surface Metadata

Sub-issue: #173
Parent issue: #172

## Red Tests / Validation

```bash
python3 - <<'PY'
import json, pathlib, sys
p=pathlib.Path('internal/connectors/defs/freshdesk/api_surface.json')
data=json.loads(p.read_text())
count=len(data.get('endpoints', []))
if count != 170:
    print(f'freshdesk api_surface endpoints={count}, want 170')
    sys.exit(1)
print('freshdesk api_surface endpoint count matches 170')
PY
test -f internal/connectors/defs/freshdesk/cli_surface.json
```

Result:

- `freshdesk api_surface endpoints=10, want 170`; exit 1.
- `cli_surface.json` presence check exit 1.

## Green Tests

Pending implementation.

Planned commands:

```bash
python3 <json/count validation>
go test ./internal/connectors/engine -run CLISurface
go test ./cmd/connectorgen -run CLISurface
go test ./cmd/connectorgen ./internal/connectors/engine
go test ./internal/connectors/conformance -run 'TestConformance/freshdesk'
go run ./cmd/connectorgen validate internal/connectors/defs
```

## Refactor Notes

- Data-only lane unless validation reveals missing embed/schema support.
- Preserve existing implemented stream references; do not overclaim writes/direct reads.
