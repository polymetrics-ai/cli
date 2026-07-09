# TDD Ledger: Front CLI Surface Metadata

Parent issue: #188
Sub-issue: #189
Branch: `feat/189-front-cli-surface-metadata`

## 2026-07-09 — plan checkpoint

- Task type: connector CLI/API metadata.
- Production behavior changed: no.
- Red evidence: pending; must be captured before editing `internal/connectors/defs/front/`.
- GSD evidence:
  - `scripts/gsd prompt plan-phase 189 --skip-research --tdd` generated the planning workflow prompt.
  - `scripts/gsd prompt programming-loop init --phase issue-189-front-cli-surface-metadata --dry-run` failed with `unknown GSD command: programming-loop`; manual GSD fallback recorded.
- Required skills loaded: `gsd-core`, `golang-how-to`, `golang-cli`, `golang-testing`, `golang-error-handling`, `golang-security`, `golang-safety`, `golang-design-patterns`, `golang-structs-interfaces`, `golang-context`, `golang-concurrency`, `golang-documentation`, `golang-spf13-cobra`, `golang-spf13-viper`, `golang-lint`.

## Red evidence

Focused metadata-completeness check failed before production edits:

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

Initial result: failed with `front api_surface endpoint count 10 != official baseline 342`.

This red check is intentionally broader than this slice's safe green path: #189 adds CLI surface
metadata without claiming full 342-operation coverage; #192 owns the complete operation ledger.

## Green evidence

- `jq empty internal/connectors/defs/front/cli_surface.json` — passed.
- `go test ./cmd/connectorgen -run CLISurface` — passed.
- `go test ./internal/connectors/engine -run CLISurface` — passed.
- `go run ./cmd/connectorgen validate internal/connectors/defs/front` — failed because the validator expects a root containing connector directories and treats `fixtures` and `schemas` as connector dirs.
- `go run ./cmd/connectorgen validate internal/connectors/defs` — passed: 547 connectors checked, 0 findings.

## Refactor notes

- JSON-only changes do not require `gofmt`; record that exemption if no Go files are edited.
- Any validator or engine change must start with a failing Go test and then run `gofmt`.
