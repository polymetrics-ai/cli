# Worker prompts

All workers implement with `gpt-5.6-sol/high`, strict test-first development, and the exclusive
ownership in `PLAN.md`. They are not alone in the repository: preserve concurrent edits, never
revert another lane, and stop at their owned boundary. Run only focused Shepherd tests while
iterating. Do not run connector, Go, runtime, or broad repository gates.

The final independent reviewer uses xhigh and receives the complete 17-row contract in one pass.
It reports blocker-level correctness/security gaps only; it does not start an open-ended hardening
cycle.
