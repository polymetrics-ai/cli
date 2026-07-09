# Verification: Front Operation Ledger (#192)

## Plan checkpoint

```bash
scripts/gsd prompt plan-phase 192 --skip-research --tdd
```

Result: prompt generated and saved to `GSD-PROMPT.md`.

```bash
scripts/gsd prompt programming-loop init --phase issue-192-front-operation-ledger --dry-run
```

Result: failed with `scripts/gsd: unknown GSD command: programming-loop`; manual GSD fallback recorded in `PLAN.md`.

## Red validation

Pending plan checkpoint commit; see `TDD-LEDGER.md`.

## Focused green gates

Pending production edit:

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
git diff --check
```

## Broader gates

Pending:

```bash
go vet ./...
go test ./cmd/connectorgen -run APISurface
go test ./internal/connectors/engine -run APISurface
go build ./cmd/pm
```

## CLI/help/docs/website parity

This slice changes operation-ledger metadata only. Runtime help renderer and docs/website command parity remain #190 scope. No executable CLI command is added here.
