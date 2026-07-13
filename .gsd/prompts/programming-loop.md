# Polymetrics universal programming loop

Apply the repository-local programming lifecycle to the requested phase and arguments.

1. Read `AGENTS.md`, the required-skills routing reference, and the issue/orchestration contracts
   that apply to the task.
2. Create or update the phase plan, TDD ledger, and verification checklist before production edits.
3. Record a failing behavioral test or capability probe before each implementation slice.
4. Implement the smallest coherent change, refactor only while tests remain green, and keep
   handoffs bounded.
5. Run focused checks followed by the applicable nested-module and repository verification gates.
6. Route automated review and preserve the final human gate. Never merge to `main` autonomously.

Use official GSD Pi as the workflow execution runtime. This prompt is a compatibility surface for
non-interactive shell callers; it does not replace GSD Pi state, orchestration, or validation.
