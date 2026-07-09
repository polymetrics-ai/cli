# TDD Ledger: Gorgias CLI Surface Metadata

Parent issue: #196  
Sub-issue: #197  
Branch: `feat/197-gorgias-cli-surface-metadata`

## 2026-07-09 — plan checkpoint

- Task type: connector CLI/API metadata.
- Production behavior changed: no.
- GSD evidence:
  - `scripts/gsd prompt plan-phase 197 --skip-research` generated the planning workflow prompt.
  - `scripts/gsd prompt execute-phase issue-197-gorgias-cli-surface-metadata --dry-run` generated the execution prompt.
  - `scripts/gsd prompt programming-loop init --phase issue-197-gorgias-cli-surface-metadata --dry-run` failed with `unknown GSD command: programming-loop`; manual GSD fallback recorded.
- Required skills loaded: `gsd-core`, `caveman`, `golang-how-to`, `golang-cli`, `golang-testing`, `golang-error-handling`, `golang-security`, `golang-safety`, `golang-design-patterns`, `golang-structs-interfaces`, `golang-context`, `golang-concurrency`, `golang-documentation`, `golang-lint`, `golang-spf13-cobra`.

## Red evidence

Focused metadata-completeness check failed before production edits:

```bash
python3 - <<'PY'
import json
from pathlib import Path
surface = json.loads(Path('internal/connectors/defs/gorgias/api_surface.json').read_text())
count = len(surface.get('endpoints', []))
cli_surface = Path('internal/connectors/defs/gorgias/cli_surface.json')
if count != 114:
    raise SystemExit(f'gorgias api_surface endpoint count {count} != official baseline 114')
if not cli_surface.exists():
    raise SystemExit('gorgias cli_surface.json is missing')
PY
```

Initial result: failed with `gorgias api_surface endpoint count 11 != official baseline 114`.

This red check is intentionally broader than this slice's safe green path: #197 adds CLI surface metadata without claiming full 114-operation implementation; #200 owns complete operation-ledger classification.

## Green evidence

- `jq empty internal/connectors/defs/gorgias/api_surface.json internal/connectors/defs/gorgias/cli_surface.json .planning/phases/issue-197-gorgias-cli-surface-metadata/OFFICIAL-OPERATIONS.json` — passed.
- `go test ./cmd/connectorgen -run CLISurface` — passed.
- `go test ./internal/connectors/engine -run CLISurface` — passed.
- `go run ./cmd/connectorgen validate internal/connectors/defs` — passed: 547 connector(s) checked, 0 findings.
- `git diff --check` — passed.
- `go test ./internal/connectors/conformance -run 'TestConformance/gorgias'` — passed.

## Refactor notes

- JSON-only changes do not require `gofmt`; record that exemption if no Go files are edited.
- Any validator or engine change must start with a failing Go test and then run `gofmt`.
