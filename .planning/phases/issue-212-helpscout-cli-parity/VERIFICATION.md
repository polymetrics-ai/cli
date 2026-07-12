# Verification: Help Scout CLI Parity Parent

Date: 2026-07-09

## Adapter / Setup Checks

```bash
scripts/gsd doctor
scripts/gsd verify-pi
scripts/gsd list --json
scripts/gsd sources plan-phase
scripts/gsd prompt plan-phase 212 --skip-research --tdd
scripts/gsd prompt programming-loop init --phase issue-212-helpscout-cli-parity --dry-run
```

## Results

- `scripts/gsd doctor`: passed.
- `scripts/gsd verify-pi`: passed.
- `scripts/gsd list --json`: passed; 69 commands found.
- `scripts/gsd sources plan-phase`: passed; sources are `.gsd/commands.json`, `.gsd/upstream.lock.json`, `.gsd/official-docs/COMMANDS.md`.
- `scripts/gsd prompt plan-phase 212 --skip-research --tdd`: prompt generated and captured in `PROMPTS.md`.
- `scripts/gsd prompt programming-loop ...`: blocked because the registry has no `programming-loop` command; manual GSD fallback recorded in `PLAN.md` and `TDD-LEDGER.md`.

## Parent Planning Checkpoint

```bash
git status --short --branch
gh pr list --head feat/212-helpscout-cli-parity --json number,title,url,state,isDraft,baseRefName,headRefName --repo polymetrics-ai/cli
```

Result:

- Branch exists locally at `feat/212-helpscout-cli-parity`.
- Draft parent PR opened: https://github.com/polymetrics-ai/cli/pull/230.

## Required Local Verification Before Handoff

```bash
gofmt -w cmd internal
go vet ./...
go test ./...
go build ./cmd/pm
make verify
go run ./cmd/connectorgen validate internal/connectors/defs
```

Status: pending implementation.

## CLI Help / Docs / Website Parity

Applies once runtime command/help/docs surfaces are changed. For #213 metadata-only work, runtime command dispatch is not yet changed; docs/website parity will be checked or explicitly marked not applicable in the #213 artifact.
