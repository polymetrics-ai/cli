# TDD LEDGER — polling-watermark changefeed executor

Manual GSD execution: each production behavior begins with a recorded failing
test. No provider, credential, database, or wall-clock sleep is used.

| ID | Behavior | Required red evidence | Green evidence |
| --- | --- | --- | --- |
| R1 | Test bundle is CDC-capable only through its matching executor | an implemented `polling_watermark` declaration alone remains non-capable | the existing real projection reports CDC only with a full matching executor |
| R2 | Complete checkpoint contract is mandatory | a declaration/executor missing `kind`, `keys`, `commit_after`, or `on_invalid` is accepted or promoted | each omission rejects/does not promote |
| R3 | Ties replay safely | page-edge records sharing a watermark are skipped or `at_least_once` is not enforced | inclusive boundary rereads edge records and delivers duplicates truthfully |
| R4 | Timestamp safety lag includes late arrivals | clock-derived lower bound omits an earlier timestamp | injected clock produces a lagged lower bound; a missing lag rejects timestamp use |
| R5 | Delete observability is honest | a hard-delete-only declaration claims a tombstone or a soft-delete marker is ignored | `not_available` remains non-tombstone; declared marker yields tombstone |
| R6 | Checkpoint follows durable destination acknowledgement | destination failure or persistence failure advances the checkpoint | no advance on failure; post-accept persistence failure reruns the same window |
| R7 | Work and cancellation are bounded | executor exceeds page/request limits or ignores cancelled context | bounded calls stop at every limit; cancellation returns without a follow-up fetch |

## Run log

Pending red-test implementation. Each entry will retain the exact focused
command and failure before its corresponding production edit.
