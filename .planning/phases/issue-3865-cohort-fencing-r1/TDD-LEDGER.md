# TDD LEDGER — issue #3865 verified-auth cohort fencing

Manual-GSD fallback: #3865 is not a numbered roadmap phase. Generated GSD prompts are executed inline by the canonical single worker; this ledger is the durable RED/GREEN record.

| ID | Requirement | RED evidence | GREEN evidence | Status |
| --- | --- | --- | --- | --- |
| R1 | A closed typed outcome allows only verified invalid authentication to fence | Planned: focused test cannot compile before the outcome and coordinator contracts exist | Test proves transport/provider/unverified outcomes leave admission and the send counter available; verified invalid changes health | Planned |
| R2 | Fencing cancels siblings and prevents all later admissions/sends | Planned: member/admission API is absent | Race test waits for cancellation, verifies a post-fence admission failure, and asserts `sends == 0` | Planned |
| R3 | Cohorts isolate and repair starts a new epoch | Planned: epoch/repair APIs are absent | Test asserts other cohort sends once, repaired epoch increases, stale member is rejected without sending, and fresh member sends once | Planned |
| R4 | Restart reloads a fence deterministically and concurrent post-fence admission is fail-closed | Planned: persistence/state reload seam is absent | Under `-race`, a new coordinator from the same opaque state store accepts zero post-fence admissions/sends and reports the stale epoch error | Planned |

## RED command log

Pending Slice A. The exact failing command and output will be appended before production coordinator code changes.

## GREEN command log

Pending Slice B/C.
