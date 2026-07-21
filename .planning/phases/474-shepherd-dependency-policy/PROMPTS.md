# Issue #474 Prompt Snapshot

## Worker kickoff

- Objective: implement the pure Shepherd dependency policy and reconciler from issue #474.
- Exact base: `e659d6f1b666f58748e2d8c86599ceb4bbc62ff8`.
- Owned production files: `autonomy-policy.ts`, `dependency-graph.ts`, `reconciler.ts` only.
- Lifecycle: plan -> RED -> GREEN -> refactor -> full verification -> stacked PR.
- GSD route: manual fallback after the repository adapter returned
  `unknown GSD command: programming-loop`.
- Downstream artifact: `PLAN.md` created; implementation pending.
- Verification result: pending.
