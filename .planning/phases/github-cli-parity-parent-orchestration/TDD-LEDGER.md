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

### CodeRabbit review-fix slice

- Accepted with modification: CLI command `stream`/`write` mutual exclusivity is enforced in
  `connectorgen validate` because this repo's minimal schema compiler does not implement draft-07
  `not`.
- Regression test added: `TestValidate_CLISurfaceRejectsCommandWithStreamAndWrite`.
- Loader evidence added: `TestBundleLoadParsesCLISurface` now asserts `RawCLISurface` is retained.
- Green targeted command:
  `go test ./cmd/connectorgen ./internal/connectors/engine`.
