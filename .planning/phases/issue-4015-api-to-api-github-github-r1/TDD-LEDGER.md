# TDD ledger — API → API GitHub route proof

## Red

No new production behavior is introduced by this proof-only delivery. A fake
red code edit would be dishonest. Instead, before any credentialed I/O, run the
pre-existing executable safety regressions that were created for this route:

- `TestPreflightReturnsTypedSourceStreamIneligibleErrorBeforeExecutorAccess`
  proves the typed `*SourceStreamIneligibleError` and zero source/stage/plan/
  apply/checkpoint effects.
- `TestPreflightRejectsClosedAdmissionFailuresBeforeSourceRead` proves an
  unsupported canonical mode/action combination stops before source read.

The exact commands and passing results are recorded in `VERIFICATION.md`. Their
assertions are the Red gate for a production-path proof: a regression that
reintroduces provider I/O on rejection must turn those tests red.

## Green

The Green gate is a successful fresh-binary GitHub route with all of:

1. one `issues` source record staged to durable WAL/Parquet then reopened;
2. `full_append` → `append` → `add_issue_labels` through plan/preview/stdin
   approval/ordinary ETL execution;
3. independent GitHub read-back of the exact target label;
4. durable acknowledgement and post-read-back checkpoint; and
5. explicit typed cleanup, including missing-label success.

## Refactor

Evidence and tests must retain no approval-token, credential, provider-payload,
or workspace-path material. No implementation refactor is permitted unless a
live, GitHub-local defect is demonstrated.
