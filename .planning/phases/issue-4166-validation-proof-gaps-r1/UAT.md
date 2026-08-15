# UAT — Issue #4166 certification proof gaps

GSD `verify-work` was executed inline. Browser or subjective UI review does not apply; all deliverables are backend validation proofs. The credential-gated live provider proof is deliberately not auto-passed.

| ID | Deliverable | Automated evidence | Result |
| --- | --- | --- | --- |
| D1 | Broken write actions fail full certification | `TestFullWriteSweepFailsForDeliberatelyBrokenDeclaredAction`, `TestFullWriteSweepFailsForDeliberatelyBrokenPreviouslyBlockedAction`, `TestFullWriteSweepExercisesAll607DeclaredGitHubWriteActions` | pass |
| D2 | Declared transport executes; absence/unregistration fail | `TestDeclaredTransportCertificationFailsWhenDeclarationIsMissing` and the three `TestCertificationDeclaredTransportPair*` tests | pass |
| D3 | Fresh-binary composed flow completes through the faithful provider | `TestFreshBinaryDeclarativeGitHubWarehouseFlowRoundTrip`; exact counts and binary identity in `SUMMARY.md` | pass |
| D4 | Dedicated real GitHub repository round trip and zero residue | `TestLiveFreshBinaryGitHubWarehouseFlowRoundTrip` | open — credential variables absent, test skipped |

Verdict: **Gaps 1 and 2 verified; Gap 3 faithful control verified; mandatory live Gap 3 remains open.** A skip is not counted as certification evidence. The observed 401 classification defect is tracked by #4169 and is not fixed here.
