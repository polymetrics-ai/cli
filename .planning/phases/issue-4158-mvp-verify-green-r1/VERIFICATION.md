# Verification — issue #4158 / Production MVP verify green

| Requirement | Evidence | Status |
| --- | --- | --- |
| Fresh externally built binary completes GitHub → warehouse flow | `TestFreshBinaryDeclarativeGitHubWarehouseFlowRoundTrip`: fresh 152,132,226-byte binary passes in 37.53s; 1 sync, 1 action, 1 warehouse row, durable checkpoint and receipt. | Pass |
| Valid managed-target control path reaches durable acknowledgement | The tagged PostgreSQL control skipped without its explicit container opt-in; #4158 is independently routed and unmodified by this PR. | Out of scope, evidenced |
| No approved job reference is typed and pre-I/O | `TestFlowActionWithoutApprovedJobReferenceRefusesBeforeIO`: malformed `*flow.JobReferenceError`, `validation/flow_job_reference_refused`, no saved flow / target event. | Pass |
| Revoked and stale references are typed and pre-I/O | `TestFlowActionRevokedJobReferenceRefusesBeforeIO` and `TestFlowActionStaleJobReferenceRefusesBeforeIO`: exact unapproved/missing typed reasons and zero target events. | Pass |
| Causal explanation distinguishes trigger, mask, symptom | `SUMMARY.md` §1–§5 and `TDD-LEDGER.md` T1–T5 | Complete |
| Smallest counterfactual and falsifier checks are retained | `TDD-LEDGER.md` T3–T4 | Complete |
| CLI help/docs/website parity | `docs/cli/flow.md` already describes job-backed flows and pre-write refusal. This fixture-only migration changes no command or documentation surface. | Not applicable / confirmed |
