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

Next:

1. Commit the #80/#97 planning and green implementation slice.
2. Create or update the parent PR from `feat/80-linear-cli-parity` to `main` when push/PR creation is allowed.
3. Continue with #98 help renderer/docs or #100 operation ledger, keeping evidence separate.
4. Route automated review per CodeRabbit/Copilot policy after PR creation and local verification.
