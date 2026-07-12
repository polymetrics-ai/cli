# Prompt Snapshots

Phase: 325-agentloop-characterization

## 2026-07-12T07:47:58.813Z - universal-kickoff

- Agent role: coordinator
- Loop type: run
- Input refs: docs/plans/universal-programming-loop-prd.md, docs/prompts/universal-programming-loop-prompts.md, docs/architecture/repo-profile.json
- Downstream artifact: `SPEC.md`, `PLAN.md`, `TEST-PLAN.md`, and supporting contracts
- Verification result: not run; planning cycle only

```text
Run the GSD Universal Programming Loop using the repo PRD, prompt library, strict TDD gate, local verification, and committed phase traces.
```

## Parent worker dispatch

- Objective: implement issue #325 only, test-first, on `fix/325-agentloop-characterization`.
- Output: stacked PR and exact worker handoff template with red/green/full-gate evidence.
- Tool guidance: isolated worktree, standard library, synthetic fixtures, fake local adapters,
  `apply_patch`, Node 24 for GSD, and repository verification commands.
- Boundaries: issue write scope only; no prompt/validator/controller-phase work, connector changes,
  secrets, live provider/GitHub effects beyond this branch/PR, dependencies, or merges.
- Downstream artifact: this phase directory and eventual worker handoff.
- Verification result: pending until all declared commands actually exit 0.

## Manual-GSD fallback

- Adapter command attempted:
  `scripts/gsd prompt programming-loop init --phase 325-agentloop-characterization --dry-run`.
- Result: adapter reports unknown GSD command `programming-loop`.
- Fallback: installed `gsd-programming-loop` helper preflight plus
  `.agents/agentic-delivery/workflows/gsd-universal-runtime-loop.md` executed by this live worker.
- Helper-generated generic visual-design files were removed immediately because they were outside
  issue scope; phase-local artifacts were retained and corrected.
