# Issue 4166 Verification Checklist

**Status:** planned; execution not started.

- [ ] Gap 1 exact negative-control tests fail the sabotaged action and pass the intact definition.
- [ ] Gap 1 report records exact operation counts and non-live categories/reasons.
- [ ] Gap 2 exact tests observe declared transport source/stage/destination/read-back execution.
- [ ] Gap 2 missing/unregistered declaration fails with zero false-positive execution.
- [ ] Gap 3 one fresh binary runs real ETL and reverse action as one composed flow.
- [ ] Gap 3 independently reopens durable Parquet and observes an advanced checkpoint.
- [ ] Gap 3 independently reads the exact mutation from GitHub.
- [ ] Gap 3 replay, expired/unapproved, and authentication refusals are typed and leave provider state unchanged.
- [ ] Disposable GitHub repository and all created resources leave zero residue.
- [ ] Generic definition/flow path and every GitHub-specific hop are documented.
- [ ] Focused package tests, vet, formatting, drift, agent-contract, and GSD workflow checks pass.
- [ ] Changed artifacts contain no credential, approval token, response body, or rendered rate scope.
- [ ] Derived artifacts regenerated once and `git status` clean after commit.
- [ ] PR opened with per-gap evidence and issue #4166 receives the closing evidence comment.
- [ ] API-reported PR base equals `integration/4015-mvp-flat-r1`.

