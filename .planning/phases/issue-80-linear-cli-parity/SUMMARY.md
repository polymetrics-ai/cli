# Summary — Issue #80 Linear CLI parity parent

Status: in progress.

Completed so far:

- Read required repo, GSD, parent-orchestration, review-routing, CLI parity, connector migration, and Go skill references.
- Validated GSD/Pi adapter with `scripts/gsd doctor`, `scripts/gsd verify-pi`, and `scripts/gsd list --json`.
- Generated the parent planning prompt with `scripts/gsd prompt plan-phase issue-80-linear-cli-parity --skip-research`.
- Recorded manual-GSD fallback because `scripts/gsd prompt programming-loop ...` is not available in this checkout.
- Created parent issue #80 plan, TDD ledger, verification checklist, run state, and orchestration state.
- Selected issue #97 (Linear CLI surface metadata) as the first local critical-path lane.
- Completed the #97 red/green slice with focused tests, full `make verify`, and generated website data.
- Committed and pushed `c9d1fa75 feat(linear): add CLI surface metadata` to `origin/feat/80-linear-cli-parity`.
- Opened draft parent PR #131: https://github.com/polymetrics-ai/cli/pull/131.

Next:

1. Continue with #98 help renderer/docs or #100 operation ledger, keeping evidence separate.
2. Keep parent PR draft while #98-#103 are incomplete.
3. Route automated review per CodeRabbit/Copilot policy after the parent PR is ready for review or when fallback coverage is required.
