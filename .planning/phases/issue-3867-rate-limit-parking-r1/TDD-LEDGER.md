# TDD LEDGER — issue #3867 rate-limit parking and automatic resumption

Manual-GSD fallback: #3867 is not a numbered roadmap phase. Generated GSD
prompts execute inline under the canonical single-worker contract; this ledger
is the durable RED/GREEN record.

| ID | Requirement | RED evidence | GREEN evidence | Status |
| --- | --- | --- | --- | --- |
| R1 | A typed rate-limit error with reset evidence persists a closed `parked_rate_limit` record | Pending: focused test will fail because park types/store/engine bridge do not exist | Pending | Planned |
| R2 | Same scope has zero pre-reset sends and unrelated scope continues | Pending: focused test will fail because parking admission is absent | Pending | Planned |
| R3 | Restart resumes once from the exact committed checkpoint without replay | Pending: focused test will fail because re-arm/scheduler API is absent | Pending | Planned |
| R4 | Cancellation, duplicate observations, callback failure, and races do not create extra sends/mutations | Pending: focused test will fail because lifecycle state machine is absent | Pending | Planned |
| R5 | Park/resume events carry actual typed reason and reset timestamp | Pending: focused engine test will fail because no parking event bridge exists | Pending | Planned |

## RED command log — pending

The test-first slice must run before production coordinator/engine code. Its
failure output will be recorded here verbatim, demonstrating the tests invoke a
missing observable contract rather than passing vacuously.

## GREEN command log — pending

After the smallest implementation slice, record the focused package and
required race-command output here. Each test must assert the described state
transition or zero side effect.
