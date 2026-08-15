# Discussion log

## Inputs read

- Full issue bodies and amendments for #3994, #3992, and #4049.
- Parent #4084, reference issue #4090, and reference PR #4161.
- Existing #3994/#3992 GSD artifacts and phase-601/#4049 evidence under the #3754 phase.
- Existing live GitHub evidence from the #3990/#4091 proof club.
- `AGENTS.md`, issue-agent contract, GSD adapter, required-skill routing, CLI parity, and runtime
  integration references.

## Resolved gray areas

- The one-time operator approval is not reintroduced per schedule run. It creates the durable
  authorization once; the new per-fire grant is derived automatically only after exact scope
  revalidation.
- Prepared identity and authorization scope identity remain separate because one binds payload and
  the other deliberately does not.
- Grant replay is prevented by a durable `O_EXCL` consumption marker under the existing approval
  authority rather than schedule state alone. Schedule state remains the crash/automatic-replay
  guard.
- Typed rate refusal belongs in `connsdk`, while coordinator detail remains a wrapped internal
  cause.
- No product questions remain. The launch brief explicitly fixes scope, naming, edge cases,
  exclusions, delivery branch, and PR base.

## Workflow note

`scripts/gsd prompt discuss-phase issue-3994-3992-4049-connector-action-auth-r1` was generated and
executed inline. The task is not a numbered roadmap phase and repository policy requires one worker,
so this directory is the documented manual-GSD fallback.
