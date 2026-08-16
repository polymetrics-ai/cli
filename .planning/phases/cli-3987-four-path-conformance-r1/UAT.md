# UAT — #3987 four-path warehouse conformance

The generated `SUMMARY.md` coverage entries are all automated and do not require human visual or product judgment.

| Deliverable | Automated proof | Result |
| --- | --- | --- |
| D1 — distinct production direction contracts | `TestWarehouseMediatedFourPathConformance` resolves each generated flow ID to its exact GitHub/PostgreSQL descriptors and persisted dispatch selection. | pass |
| D2 — sealed warehouse mediation | `TestOrchestratorDispatchesFourClosedPairingsWithoutPairBranches` records planned-before-read, connection ownership, stage/reopen, target apply/read-back, and final checkpoint ordering. | pass |
| D3 — current mode truth | `TestWarehouseMediatedModeConformance` covers all seven closed modes; six currently executable transports pass and source-only `change_capture` is a reasoned refusal excluded from the pass roll-up. | pass |

No live provider or database behavior is marked passed by this UAT. The task preserves the existing owner lanes' fresh-binary/live route proof and leaves #3978’s final live certification/publication untouched.
