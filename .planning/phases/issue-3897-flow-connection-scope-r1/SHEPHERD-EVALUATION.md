# #3897 Shepherd-Compatible Evaluation

**Result:** PASS — bounded equivalent evaluation

The delivery contract forbids role spawning and no compatible external
Shepherd runtime is available for this issue-phase fallback. This record is a
transparent equivalent evaluation, not a claim that an external Shepherd bot
ran.

| Criterion | Result | Evidence |
|---|---|---|
| Correct owner selection | PASS | Flow CLI test builds two same-named Parquet tables and proves query/action rows are `acme-1`/`globex-1` only. |
| Omitted selector safety | PASS | Typed `*warehouse.AmbiguousTableError`; remedy is manifest syntax and contains no CLI flag. |
| Root ownership | PASS | `_unattributed` query/action tests return the root-only row. |
| Approval-source binding | PASS | JSON round-trip plus action-runner capture retain `SourceConnection`; source rows are read before the runner. |
| Scope fence | PASS | Local capture runner only; no provider or reverse lifecycle code changed. |
| Public contract | PASS | Runtime manual, docs, website, generated docs, and golden transcripts agree. |

Correction 1 / 5 found and fixed a capped action-source read during review.
The 101-row local regression now passes without provider mutation.
