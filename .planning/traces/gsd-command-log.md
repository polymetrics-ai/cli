# GSD Command / Workflow Log — Issue #122

## Environment note

This Pi harness can read files, run shell commands, and edit/write files, but it does not expose Claude slash-command execution or the Claude `Task` subagent API. Therefore the upstream GSD Core command specs were executed as workflow documents from their installed command/source files, with deterministic preflight commands run through `gsd-tools.cjs` and all workflow sources recorded here.

## Preflight commands run

```bash
git status --short --branch
node "$HOME/.claude/get-shit-done/bin/gsd-tools.cjs" init new-project
node "$HOME/.claude/get-shit-done/bin/gsd-tools.cjs" init map-codebase
find .planning/phases -maxdepth 1 -mindepth 1 -type d | wc -l
test -d .planning/codebase || echo NO_CODEBASE_MAP
git diff --name-only -- cmd internal
```

## Upstream GSD command sources used

Equivalent GSD user commands:

```text
/gsd:map-codebase connector parity, all connector technologies and surfaces
/gsd:new-project --auto @.planning/traces/issue-122-gsd-onboarding-prompt.md
/gsd:plan-phase 1 --skip-research
/gsd:programming-loop init --phase 01-inventory-reconciliation --dry-run
```

Installed workflow source files read/executed:

- `/Users/karthiksivadas/.claude/commands/gsd/map-codebase.md`
- `/Users/karthiksivadas/.claude/get-shit-done/workflows/map-codebase.md`
- `/Users/karthiksivadas/.claude/commands/gsd/new-project.md`
- `/Users/karthiksivadas/.claude/get-shit-done/workflows/new-project.md`
- `/Users/karthiksivadas/.claude/commands/gsd/plan-phase.md`
- `/Users/karthiksivadas/.claude/get-shit-done/workflows/plan-phase.md`
- `/Users/karthiksivadas/.claude/commands/gsd/programming-loop.md`

## Archive evidence

The previous active `.planning/` tree was archived outside active `.planning/` before replacement:

```text
../planning-archives/polymetrics-cli-issue-122-pre-rebootstrap-20260708173641.tar.gz
sha256 e0959e4c8eba6e8610255a0cd9a98b39267902ba19600515abfdab726bfd57f5
```

## Initial red evidence

- Existing custom `.planning/phases/` contained 26 phase directories.
- `.planning/codebase/` was absent.
- `gsd-tools init new-project` reported `project_exists: true`, `is_brownfield: true`, and `needs_codebase_map: true` before the rebootstrap.
- `git diff --name-only -- cmd internal` produced no output before planning edits.

## Scope guard

Issue #122 is planning-only. Generated/updated files are confined to `.planning/` and optional ignore/archive metadata. No `cmd/` or `internal/` source edits are allowed.
