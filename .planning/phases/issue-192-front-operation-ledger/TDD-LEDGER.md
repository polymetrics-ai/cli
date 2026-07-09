# TDD Ledger: Front Operation Ledger (#192)

## Red validation before production edits

The current Front bundle failed operation-ledger completeness because it had only 10 endpoint rows and no `operation_ledger_version: 1`.

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

Initial result: failed with `front api_surface.json is not in operation_ledger_version=1 mode`.

## Green validation after production edits

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

Green results:

- `jq empty internal/connectors/defs/front/api_surface.json .planning/phases/issue-192-front-operation-ledger/REST-OPERATION-SUMMARY.json` — passed.
- Count/method/classifier check — passed with 255 rows, 6 covered streams, 249 blocked operation rows.
- `go run ./cmd/connectorgen validate internal/connectors/defs` — passed; 547 connectors checked, 0 findings.

## Broader gates

- `git diff --check` — passed.
- `go vet ./...` — passed.
- `go test ./cmd/connectorgen -run APISurface` — passed.
- `go test ./internal/connectors/engine -run APISurface` — passed.
- `go build ./cmd/pm` — passed.
- `go test ./...` not rerun in this slice; previous #189 local run timed out in `internal/connectors/certify/TestWriteStagesSkipWhenDisabled` while GitHub Verify passed.
