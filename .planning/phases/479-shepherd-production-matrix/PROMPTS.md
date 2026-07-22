# Worker prompts

All workers implement with `gpt-5.6-sol/high`, strict test-first development, and the exclusive
ownership in `PLAN.md`. They are not alone in the repository: preserve concurrent edits, never
revert another lane, and stop at their owned boundary. Run only focused Shepherd tests while
iterating. Do not run connector, Go, runtime, or broad repository gates.

The final independent reviewer uses xhigh and receives the complete 17-row contract in one pass.
It reports blocker-level correctness/security gaps only; it does not start an open-ended hardening
cycle.

## Remote CI repair kickoff snapshot

- Objective: close PR #489's two ordinary-host CI failures without changing Shepherd behavior.
- Execution decision: `local_critical_path`; a read-only parent orchestrator audits stack topology.
- Parent action: merge current `origin/main` into `feat/471-pi-agent-session-shepherd` to inherit
  the existing `golang.org/x/text v0.39.0` security fix.
- Child action: pin Go `1.25.12` in `.github/workflows/shepherd.yml`, then merge the updated parent.
- RED: run `29959846371` cleanup `EACCES`; run `29959846280` `GO-2026-5970`.
- Downstream artifact: `PLAN.md`, `TDD-LEDGER.md`, `VERIFICATION.md`, and `RUN-STATE.json`.
- Verification result: pending implementation and rerun.
