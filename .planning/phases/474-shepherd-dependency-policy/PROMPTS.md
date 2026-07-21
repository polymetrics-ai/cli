# Issue #474 Prompt Snapshot

## Worker kickoff

- Objective: implement the pure Shepherd dependency policy and reconciler from issue #474.
- Exact base: `e659d6f1b666f58748e2d8c86599ceb4bbc62ff8`.
- Owned production files: `autonomy-policy.ts`, `dependency-graph.ts`, `reconciler.ts` only.
- Lifecycle: plan -> RED -> GREEN -> refactor -> full verification -> stacked PR.
- GSD route: manual fallback after the repository adapter returned
  `unknown GSD command: programming-loop`.
- Downstream artifact: pure policy/reconciler implemented with three exact-head correction loops
  and final GSD/TDD/verification artifacts.
- Verification result: parent-declared phase-equivalent child gate passed at 41 focused and 178
  full Shepherd tests; full-repo `make verify` remains excluded by superseding parent policy.
