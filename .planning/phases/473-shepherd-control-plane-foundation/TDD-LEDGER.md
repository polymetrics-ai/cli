# TDD Ledger: #473

## RED: cancellation ownership

Command:

```bash
node --test --test-name-pattern='shutdown wins a concurrent cancellation race|stop wins a concurrent shutdown race' .pi/extensions/shepherd/controller.test.ts
```

Result before implementation: 1 passed, 1 failed. Shutdown-first was overwritten by a later stop,
and the later stop incorrectly fulfilled. GREEN adds an explicit `cancelling` phase and centralized
first-wins cancellation. Result after implementation: 2/2, then controller suite 31/31.

## RED: lease/root/state invariants

Command:

```bash
node --test --test-name-pattern='epoch cleanup cannot|thirteenth digit|contradictory timestamps|pins the state-root|root replacement after epoch publication' .pi/extensions/shepherd/state-store.test.ts
```

Result before implementation: 0/5. The probes reproduced orphan acquisition after epoch cleanup,
13-digit epoch publication, empty halted-state acceptance, same-path root rebinding, and suppressed
root replacement during cleanup.

GREEN introduced highest-anchor retry/revalidation, exact authoritative-owner confirmation,
device/inode pinning, explicit epoch bounds, unsuppressed root guards, and coherent assessed-state
validation. Targeted result: 5/5.

## RED: cleanup ordering and token uniqueness review

An independent read-only lease review found two additional blockers. New tests proved:

- a delayed lower-epoch writer could recreate epoch N and delete authoritative epoch N+1;
- a successfully linked orphan owner reused its global token namespace on retry; and
- malformed zero/13-digit reserved epoch files were ignored.

Before correction: 0/5 across the lower-epoch, orphan-token, and malformed-namespace cases. GREEN
verifies authority before cleanup, refuses current/newer epoch deletion, revalidates after cleanup,
regenerates owner metadata after orphan publication, and rejects malformed reserved names. The
expanded targeted gate passes 9/9.

## RED: exact-diff review blockers

The first independent exact-diff review found five additional blockers not covered by the original
122-test suite. Focused RED tests proved:

- a second missing-successor read could return a tail after its epoch was concurrently cleaned;
- an all-pending `running` crash checkpoint could not be reconciled and persisted;
- a stop accepted during failing target capture could finish without a stopped checkpoint;
- ambient `GH_REPO`/an arbitrary PR URL could supply evidence for a different repository; and
- expired AgentSession deadlines disposed/unregistered unsettled work and released mutator
  ownership before cleanup really finished.

GREEN adds a final exact-anchor recheck, conservative unfinished-lane interruption, an
initializing lifecycle phase, explicit repository-bound `gh --repo` evidence, and persistent
AgentSession settlement/quarantine with a dedicated cleanup bound. The focused fixes passed their
targeted suites, including 4/4 teardown regressions repeated five times.

The clean re-review then found two lifecycle variants: shutdown during initialization could miss
cancellation, and a rejected (rather than hanging) child settlement could release the mutator
reservation. Both focused probes were RED 0/1. GREEN permits shutdown interruption during the
initializing phase and records a persistent poisoned/quarantined runner state on settlement
rejection. Both focused probes pass 1/1 and the combined controller/SDK suites pass.

A second independent reviewer found five final invariant/routing cases. RED probes proved that
extended-year ISO timestamps defeated textual chronology, failed+halted lanes could be persisted,
the controller could generate that same forbidden mixture, interrupted checkpoints could retain
pending work, and runtime terminal events could drift provider/model after the pre-prompt check.
GREEN compares timestamp instants, narrows persisted lane combinations, normalizes aggregate halt
outcomes, rejects pending interrupted work, and checks the terminal route before parsing evidence.
The final independent review verdict is CLEAN.

## Refactor and complete focused gate

- Centralized lifecycle cancellation in `requestCancellation`.
- Split lease anchor listing, successor validation, authoritative confirmation, and guarded epoch
  finalization into explicit helpers.
- Kept model output outside state authority and retained fixed persisted summary categories.

Final focused gate:

```text
tests 137
pass 137
fail 0
duration_ms 41543.020792
```

Strict TypeScript over the nine production modules passed with Pi 0.80.6 types. Offline Pi RPC
command discovery returned `true` for `pm-shepherd`/`extension`.
