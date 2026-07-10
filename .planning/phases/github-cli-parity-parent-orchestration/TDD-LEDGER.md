# TDD Ledger: GitHub CLI Parity Parent Orchestration

## 2026-07-07

### Rebase and model-policy setup

- Task type: orchestration/configuration.
- Red evidence: not applicable; no production behavior was changed in this slice.
- Validation target:
  - `jq empty .planning/config.json`
  - `git diff --check`
  - YAML parse check for `.agents/`
  - targeted Go/website checks before push

### Manual GSD fallback

- `scripts/programming-loop.mjs` was not present.
- `scripts/tdd-gate.mjs` was not present.
- Manual loop is recorded in `PLAN.md`; behavior-adding subissues must still start with red tests.

### Claude review-fix slice

- Accepted with modification: CLI command `stream`/`write` mutual exclusivity is enforced in
  `connectorgen validate` because this repo's minimal schema compiler does not implement draft-07
  `not`.
- Regression test added: `TestValidate_CLISurfaceRejectsCommandWithStreamAndWrite`.
- Loader evidence added: `TestBundleLoadParsesCLISurface` now asserts `RawCLISurface` is retained.
- Green targeted command:
  `go test ./cmd/connectorgen ./internal/connectors/engine`.

### Active orchestration runtime slice

- Task type: orchestration/configuration.
- Red/validation target: existing workflow allowed passive planning because the parent orchestrator
  contract had no activation field and Codex/OpenCode adapters did not require worker spawn
  decisions.
- Green evidence:
  - `activation.mode: active_owner` added to the parent orchestrator YAML.
  - Codex/OpenCode adapters added for GSD loop and parent issue orchestration.
  - `caveman` repo-local skill added under `.agents/skills/`.
  - Orchestration state schema now requires `orchestrator`, `ready_queue`, and `spawn_decisions`.
