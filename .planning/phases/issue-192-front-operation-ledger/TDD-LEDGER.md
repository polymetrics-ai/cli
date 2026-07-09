# TDD Ledger: Front Operation Ledger (#192)

## Red validation planned before production edits

The current Front bundle should fail operation-ledger completeness because it has only 10 endpoint rows and no `operation_ledger_version: 1`.

```bash
python3 - <<'PY'
import json
from pathlib import Path
surface = json.loads(Path('internal/connectors/defs/front/api_surface.json').read_text())
count = len(surface.get('endpoints', []))
if surface.get('operation_ledger_version') != 1:
    raise SystemExit('front api_surface.json is not in operation_ledger_version=1 mode')
if count != 255:
    raise SystemExit(f'front api_surface endpoint count {count} != captured REST registry count 255')
PY
```

Expected initial result: fail.

## Green validation planned after production edits

```bash
jq empty internal/connectors/defs/front/api_surface.json
python3 - <<'PY'
import json
from pathlib import Path
surface = json.loads(Path('internal/connectors/defs/front/api_surface.json').read_text())
counts = {}
for ep in surface['endpoints']:
    counts[ep['method']] = counts.get(ep['method'], 0) + 1
assert surface.get('operation_ledger_version') == 1
assert len(surface['endpoints']) == 255
assert counts == {'GET': 123, 'POST': 76, 'PATCH': 26, 'DELETE': 27, 'PUT': 3}, counts
covered = [ep for ep in surface['endpoints'] if ep.get('covered_by')]
assert len(covered) == 6
blocked = [ep for ep in surface['endpoints'] if ep.get('operation')]
assert len(blocked) == 249
assert not any('excluded' in ep for ep in surface['endpoints'])
PY
go run ./cmd/connectorgen validate internal/connectors/defs
```

## Broader gates planned as applicable

- `git diff --check`
- `go vet ./...`
- `go test ./cmd/connectorgen -run APISurface`
- `go test ./internal/connectors/engine -run APISurface`
- `go build ./cmd/pm`
- `go test ./...` only if time permits; previous #189 local run timed out in `internal/connectors/certify/TestWriteStagesSkipWhenDisabled` while GitHub Verify passed.
