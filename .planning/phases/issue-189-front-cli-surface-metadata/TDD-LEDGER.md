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

## Planned red evidence

Run a focused check that fails on the current Front metadata before production edits. Target facts:

- `internal/connectors/defs/front/api_surface.json` has 10 endpoint rows.
- Official baseline in issue #188 is 342 operations.
- `internal/connectors/defs/front/cli_surface.json` is absent.

Candidate command:

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

## Planned green evidence

- `jq empty internal/connectors/defs/front/api_surface.json internal/connectors/defs/front/cli_surface.json`
- `go test ./cmd/connectorgen -run CLISurface`
- `go test ./internal/connectors/engine -run CLISurface`
- `go run ./cmd/connectorgen validate internal/connectors/defs`

## Refactor notes

- JSON-only changes do not require `gofmt`; record that exemption if no Go files are edited.
- Any validator or engine change must start with a failing Go test and then run `gofmt`.
