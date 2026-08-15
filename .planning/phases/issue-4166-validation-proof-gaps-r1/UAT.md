# UAT — Issue #4166 certification proof gaps

GSD `verify-work` was executed inline. Browser or subjective UI review does not apply; all deliverables are backend validation proofs.

| ID | Deliverable | Automated evidence | Result |
| --- | --- | --- | --- |
| D1 | Broken write actions fail full certification | `TestFullWriteSweepFailsForDeliberatelyBrokenDeclaredAction`, `TestFullWriteSweepFailsForDeliberatelyBrokenPreviouslyBlockedAction`, `TestFullWriteSweepExercisesAll607DeclaredGitHubWriteActions` | pass |
| D2 | Declared transport executes; absence/unregistration fail | `TestDeclaredTransportCertificationFailsWhenDeclarationIsMissing` and the three `TestCertificationDeclaredTransportPair*` tests | pass |
| D3 | Fresh-binary composed flow completes through the faithful provider | `TestFreshBinaryDeclarativeGitHubWarehouseFlowRoundTrip`; exact counts and binary identity in `SUMMARY.md` | pass |
| D4 | Dedicated real GitHub repository round trip and zero residue | `TestLiveFreshBinaryGitHubWarehouseFlowRoundTrip`; live GitHub evidence reports flow records, provider read-back, checkpoint/receipt, refusal invariants, repository deletion, and post-delete 404 | pass |

Verdict: **all three gaps verified.** The observed 401 classification defect is tracked by #4169 and is not fixed here.
