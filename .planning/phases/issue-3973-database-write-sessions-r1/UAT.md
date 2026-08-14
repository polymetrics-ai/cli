# UAT — Issue 3973: transactional database write sessions

Inline `verify-work` classified the `SUMMARY.md` coverage entries as automated
proof. No human visual or product judgment is needed for this driver-neutral
contract slice.

| Deliverable | Evidence | Verdict |
| --- | --- | --- |
| D1 — sealed preview/approval binding | Mismatch cases observe zero begin/batch/commit/rollback/ledger calls; accepted execution observes approval consumed before begin. | passed |
| D2 — one session and bounded batches | Six records at batch size two observe one begin, three `[2,2,2]` batches, no legacy write, plus rollback/no-ack on failure and cancellation. | passed |
| D3 — durable receipt and unknown commit | Confirmed commit records one ledger receipt before acknowledgement; unknown commit observes no rollback/retry/ledger/checkpoint authority. | passed |
| D4 — atomic and canonical modes | Non-atomic overwrite observes zero calls; atomic publish occurs after batches; append/upsert/dedupe record canonical strategy in one session. | passed |

The real PostgreSQL target proof is deliberately deferred: this issue has no
native driver, DDL, SQL, credential, or capability change. #3982 owns native
driver proof and #3978 owns real-binary certification.
