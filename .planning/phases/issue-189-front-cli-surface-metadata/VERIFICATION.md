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

## Red validation before production connector edits

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

Result: failed as expected with `front api_surface endpoint count 10 != official baseline 342`.

## Focused green gates completed

```bash
jq empty internal/connectors/defs/front/cli_surface.json
```

Result: pass.

```bash
go test ./cmd/connectorgen -run CLISurface
```

Result: pass.

```bash
go test ./internal/connectors/engine -run CLISurface
```

Result: pass.

```bash
go run ./cmd/connectorgen validate internal/connectors/defs/front
```

Result: expected command-shape failure; the validator expects a root containing connector directories and treated `fixtures` and `schemas` as connector directories.

```bash
go run ./cmd/connectorgen validate internal/connectors/defs
```

Result: pass; 547 connectors checked, 0 findings.

```bash
git diff --check
```

Result: pass.

## Broader handoff gates

```bash
go vet ./...
```

Result: pass.

```bash
go build ./cmd/pm
```

Result: pass.

```bash
go test ./...
```

Result: fail/blocker unrelated to this JSON/docs slice: `internal/connectors/certify` timed out after 10m while running `TestWriteStagesSkipWhenDisabled`. The prior 600s run also timed out after partial package output. Focused CLISurface and connectorgen gates passed.

```bash
gofmt -w cmd internal
```

Result: not applicable; no Go files changed.

```bash
make verify
```

Result: not run after `go test ./...` timeout; would be blocked by the same full-test failure.

```bash
go run ./cmd/connectorgen validate internal/connectors/defs
```

Result: pass; 547 connectors checked, 0 findings.

## Sub-PR

```bash
gh pr create --draft --base feat/188-front-cli-parity --head feat/189-front-cli-surface-metadata --title "feat(front): add CLI surface metadata" --body-file /tmp/front_189_pr_body.md
```

Result: pass; draft sub-PR opened at https://github.com/polymetrics-ai/cli/pull/231.

```bash
gh pr ready 231
```

Result: pass; PR #231 marked ready after checks completed. CodeRabbit had skipped the draft revision, so automatic review coverage remains pending on the ready/update commit.

## CLI help/docs/website parity

Runtime renderer/docs/website changes are #190. For this slice, verify metadata only and record that
help/docs parity is intentionally deferred to #190.
