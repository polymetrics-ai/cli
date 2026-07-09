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

```bash
python3 - <<'PY'
import json, pathlib
from collections import Counter
for name in ['metadata','api_surface','cli_surface']:
    p=pathlib.Path(f'internal/connectors/defs/freshdesk/{name}.json')
    json.loads(p.read_text())
api=json.loads(pathlib.Path('internal/connectors/defs/freshdesk/api_surface.json').read_text())
print('api endpoints', len(api['endpoints']), Counter(ep['method'] for ep in api['endpoints']))
print('covered', sum('covered_by' in ep for ep in api['endpoints']), 'blocked', sum('operation' in ep for ep in api['endpoints']))
PY
go run ./cmd/connectorgen validate internal/connectors/defs --json
go test ./internal/connectors/engine -run CLISurface
go test ./cmd/connectorgen -run CLISurface
go test ./cmd/connectorgen ./internal/connectors/engine
go test ./internal/connectors/conformance -run 'TestConformance/freshdesk'
```

Results:

- Freshdesk JSON parse/count passed: 170 endpoints (`GET:117`, `POST:10`, `PUT:10`, `DELETE:33`), 5 covered streams, 165 blocked-by-default operation rows.
- `go run ./cmd/connectorgen validate internal/connectors/defs --json`: passed, 547 connectors checked, 0 findings, 0 warnings.
- `go test ./internal/connectors/engine -run CLISurface`: passed.
- `go test ./cmd/connectorgen -run CLISurface`: passed.
- `go test ./cmd/connectorgen ./internal/connectors/engine`: passed.
- `go test ./internal/connectors/conformance -run 'TestConformance/freshdesk'`: passed.

## Refactor Notes

- Data-only lane; no production Go code changed.
- Existing implemented stream references remain covered; unimplemented reads/writes are blocked metadata and not exposed as raw HTTP reads/writes.
