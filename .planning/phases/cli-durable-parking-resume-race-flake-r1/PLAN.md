# Plan — durable parking resume claim race

## Scope and causal hypothesis

Two CLI processes reopen durable rate-parking state when a checkpoint becomes
due. Their `app.Open`-constructed coordinators must make exactly one durable
resume claim, and the losing process must participate safely instead of failing
the command. First reproduce and trace the state transition, lock lifetime, and
claim/release path before selecting the smallest production correction.

## TDD slices

1. **Red — production process reproduction.** Run the existing `resume-race`
   case repeatedly and retain its exit-code/output diagnostics. Add focused
   named assertions only where current evidence cannot distinguish the claiming
   outcome from a process-open/lock failure.
2. **Red — durable claim contract tests.** Before production edits, add
   separately named tests for each changed behavior:
   - **Happy path:** a due durable record is resumed through the real
     `cli.Run` → `app.Open` path; both concurrent processes complete and the
     provider sees exactly one resumed send.
   - **Bad path:** a competing live claim is rejected with the existing typed
     claim/admission refusal before it performs a resume I/O side effect.
   - **Edge — interleaved cross-process resumption:** simultaneous reopeners
     contend for the same persisted due checkpoint without data loss, duplicate
     provider send, or an unsuccessful loser process. This is the reported
     race's direct boundary.
3. **Green — production correction.** Change only the owning durable store or
   app coordinator construction/state transition. Preserve atomicity across
   processes and error identity. Do not retry, increase a timeout, reduce
   concurrency, skip, or quarantine anything.
4. **Refactor and evidence.** Run focused normal/race tests and 20 consecutive
   process-test runs with a recorded load snapshot; run relevant package/static
   gates, build, GSD verification, and inline code review. Record the observed
   causal race and why the loser previously failed.

## Acceptance evidence

| Requirement | Proof |
| --- | --- |
| Fix production, not the test timing | Diff alters durable coordination production path and the process test continues to run at full concurrency with its existing diagnostics. |
| Real construction path | `internal/cli` test executes the child CLI dispatcher and `app.Open`, rather than hand-constructing a coordinator. |
| Claim safety | A focused bad-path test asserts the typed refusal before resume provider I/O; an interleaving test proves only one provider send. |
| Reliability | At least 20 consecutive targeted `resume-race` passes and a machine-load reading appear in TDD/PR evidence. |

## GSD lifecycle

- Resolved prompts: `discuss-phase`, `plan-phase --tdd --skip-research`,
  `execute-phase`, `verify-work`, and `code-review` via `scripts/gsd prompt`.
- Inline/manual fallback: this worker has no compatible Pi isolated-agent
  runtime, and the repository's canonical single-worker contract forbids role
  spawning. This records full lifecycle evidence without weakening TDD,
  verification, or review.
