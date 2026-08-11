# #4067 execution record

**Status:** Planned — no production mutation has occurred.

## TDD gate

The first execution action is a behavioral RED in `internal/app` against the real persisted JSON state path. It must demonstrate the Sol F1 durable `running` leak and zero/non-terminal returned run after an acknowledged checkpoint and an unrelated post-checkpoint writer. No `internal/` production file may change until that RED command and the test-only commit are recorded below.

## Planned checkpoints

| Checkpoint | Required evidence | Status |
|---|---|---|
| Planning | #4067 issue/readback, manual GSD context/plan/TDD ledger, clean rejected baseline custody | Complete before this commit |
| RED | Focused test, non-zero exit due to durable symptom, test-only commit | Pending |
| GREEN | Minimal completion-boundary implementation and same test passing | Pending |
| Focused expansion | all modes, reopen, cancellation, fail-closed eligibility, race, #4046/R7/R8 | Pending |
| Generated remediation | canonical generator commands and candidate-owned diff only | Pending |
| Heavy validation | only after required user-facing window notification | Pending |
| Review/no-mistakes/CI | all findings dispositioned; fresh 0/5 no-mistakes; exact-head CI | Pending |

