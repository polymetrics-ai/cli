# Verification: Front CLI Surface Metadata

## Completed

```bash
scripts/gsd prompt plan-phase 189 --skip-research --tdd
```

Result: pass; planning prompt generated.

```bash
scripts/gsd prompt programming-loop init --phase issue-189-front-cli-surface-metadata --dry-run
```

Result: blocked; `scripts/gsd` reported `unknown GSD command: programming-loop`. Manual GSD fallback recorded in `PLAN.md` and `TDD-LEDGER.md`.

## Pending before production connector edits

```bash
python3 - <<'PY'
import json
from pathlib import Path
surface = json.loads(Path('internal/connectors/defs/front/api_surface.json').read_text())
count = len(surface.get('endpoints', []))
cli_surface = Path('internal/connectors/defs/front/cli_surface.json')
if count != 342:
    raise SystemExit(f'front api_surface endpoint count {count} != official baseline 342')
if not cli_surface.exists():
    raise SystemExit('front cli_surface.json is missing')
PY
```

Expected initial result: fail.

## Focused green gates

```bash
jq empty internal/connectors/defs/front/api_surface.json internal/connectors/defs/front/cli_surface.json
go test ./cmd/connectorgen -run CLISurface
go test ./internal/connectors/engine -run CLISurface
go run ./cmd/connectorgen validate internal/connectors/defs
```

## Broader handoff gates

```bash
gofmt -w cmd internal
go vet ./...
go test ./...
go build ./cmd/pm
make verify
go run ./cmd/connectorgen validate internal/connectors/defs
```

`gofmt` is not applicable if this remains JSON/docs-only.

## CLI help/docs/website parity

Runtime renderer/docs/website changes are #190. For this slice, verify metadata only and record that
help/docs parity is intentionally deferred to #190.
