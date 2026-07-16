# Verification — Issue 398 CLI Architecture v2 Bootstrap

## Required Commands

```bash
scripts/gsd doctor
scripts/gsd prompt new-milestone "CLI Architecture v2" >/tmp/gsd-new-milestone-cli-arch-v2.prompt
test -s /tmp/gsd-new-milestone-cli-arch-v2.prompt
scripts/gsd prompt plan-phase 398 --skip-research >/tmp/gsd-plan-phase-398.prompt
test -s /tmp/gsd-plan-phase-398.prompt
scripts/gsd prompt programming-loop init --phase 398 --dry-run >/tmp/gsd-programming-loop-398.prompt
```

Expected programming-loop result: command currently exits non-zero with `scripts/gsd: unknown GSD command: programming-loop`; record Pi-local `/pm-gsd-loop` fallback instead of skipping TDD/planning.

```bash
rg -n "CLI Architecture v2|feat/cli-architecture-v2|#397|#398" .planning/PROJECT.md .planning/ROADMAP.md .planning/phases/398-cli-architecture-v2-bootstrap .planning/traces
rg -n "Golden transcript|Cobra router|typed Viper|events bus|slog foundation|OpenTelemetry metrics|Architecture v2 cleanup" .planning/ROADMAP.md
rg -n "Polymetrics CLI Connector Parity|Inventory and Surface Reconciliation|Durable Read and ETL Parity|Conformance and Certification Enforcement" .planning/PROJECT.md .planning/ROADMAP.md
git diff --check
git diff --name-only -- cmd internal
```

After push/PR creation:

```bash
gh pr view --json number,url,state,isDraft,baseRefName,headRefName,title
```

## Results

- `scripts/gsd doctor`: pass.
- New-milestone and plan-phase generated prompts: non-empty.
- Programming-loop shell prompt: unavailable with exact error `scripts/gsd: unknown GSD command: programming-loop`; Pi-local `/pm-gsd-loop` fallback recorded.
- Planning/roadmap preservation grep checks: pass.
- `git diff --check`: pass.
- `git diff --name-only -- cmd internal`: no output.
- Seed commit: `2f012400632ad64b1c0c3e2ba98d8bd98999b25d`.
- Draft parent PR: [#438](https://github.com/polymetrics-ai/cli/pull/438), `feat/cli-architecture-v2` → `main`.

## Expected Results

- GSD doctor exits 0.
- GSD new-milestone and plan-phase prompts are non-empty.
- Programming-loop shell prompt is unavailable and explicitly recorded as fallback.
- Active planning includes CLI Architecture v2 milestone, parent branch, parent issue, and Stage 0 issue.
- 22 source-plan phases are visible in `.planning/ROADMAP.md`.
- Existing connector-parity roadmap/workstreams remain present.
- `git diff --check` exits 0.
- `git diff --name-only -- cmd internal` prints no output.
- Draft parent PR targets `main` from `feat/cli-architecture-v2` and includes `Refs #397`.

## Not Run

- `go test ./...` / `make verify`: not required for this planning-only bootstrap unless production Go changes appear.
- Credentialed connector checks: out of scope.
- Reverse ETL execution: out of scope.
- Runtime-backed checks: out of scope.
