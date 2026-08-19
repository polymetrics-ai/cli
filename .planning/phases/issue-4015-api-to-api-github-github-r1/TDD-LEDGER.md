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

Observed Green: the fresh built binary completed the private run-owned
GitHub `issues` #1 → durable WAL/Parquet receipt → `add_issue_labels` on #2
route through the ordinary plan/preview/stdin-approval/ETL carrier.
`gh-axi` independently observed exactly `pm-api-to-api-route-r1` on #2;
the receipt and acknowledged checkpoint were present locally. The same
approved run replayed without corrupting or duplicating the GitHub label.

## Refactor

Evidence and tests must retain no approval-token, credential, provider-payload,
or workspace-path material. No implementation refactor is permitted unless a
live, GitHub-local defect is demonstrated.

Observed Refactor/edge: the separately approved `remove_issue_label` inverse
returned #2 to no labels; a second missing-label cleanup succeeded by declared
missing-status behavior, while replaying its consumed approval was refused.
Focused source-eligibility, unsupported-mode, explicit-zero-result, absent
first-page mapping, receipt, and cleanup regressions pass. No source code was
changed because no GitHub-local defect was observed.
