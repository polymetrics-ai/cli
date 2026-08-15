# Verification — issue #4158 / Production MVP verify green

| Requirement | Evidence | Status |
| --- | --- | --- |
| Fresh externally built binary completes GitHub → warehouse flow | `TestFreshBinaryDeclarativeGitHubWarehouseFlowRoundTrip`, repeated at final head | Pending |
| Valid managed-target control path reaches durable acknowledgement | `TestPostgresManagedTargetDriverLiveControlAssertions` plus new happy regression | Pending |
| Non-PostgreSQL history route remains typed and pre-I/O | New bad regression; fake driver zero side effects | Pending |
| Causal explanation distinguishes trigger, mask, symptom | `SUMMARY.md` and PR body cite exact runs / commits | Pending |
| Smallest counterfactual and falsifier checks are retained | `TDD-LEDGER.md` records commands and observations | Pending |
| CLI help/docs/website parity | Not applicable unless public CLI surface changes | Pending re-evaluation |
