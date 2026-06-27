# EVAL-PLAN — Action Step (Phase 1)

## Prompt-eval notes (no LLM in this phase)

This phase contains no LLM calls. Prompt-eval is N/A.

## Correctness gates (automated, via go test)

| Gate | Test ID | Pass criterion |
|------|---------|----------------|
| Idempotency | T-11 | 0 duplicate sends on re-run |
| Identity mapping | T-12 | pm_id→ext_id persisted; re-run skips |
| Dedupe | T-13 | N-unique records sent when source has duplicates |
| 429 backoff | T-14 | Succeeds after retry; server called 3× |
| DLQ quarantine | T-15 | RecordsFailed==1; DLQ file exists |
| Schema drift halt | T-16 | ErrSchemaDrift; 0 records sent |
| Additive field OK | T-16b | No error; run proceeds |
| Receipts | T-17 | Ledger has receipt entry |
| Approval gate | T-18 | ErrApprovalRequired without token |
| CLI token flag | T-19 | Engine called with token |

## Performance (not benchmarked in this phase)

Benchmark is deferred to Phase 5 (Agent Mode token efficiency).
Action step throughput is bounded by destination connector latency, not pm internals.

## Definition of done for this phase

- All T-10 through T-19 tests pass.
- `make verify` is green.
- TDD-LEDGER.md has red-evidence entries for each test before its implementation.
