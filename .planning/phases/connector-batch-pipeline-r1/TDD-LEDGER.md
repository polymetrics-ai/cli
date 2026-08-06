# TDD ledger — connector batch pipeline r1

| ID | Enforcement | RED evidence | GREEN evidence | Refactor / verification |
| --- | --- | --- | --- | --- |
| B1 | A candidate cannot enter a batch without complete measured evidence | Pending: focused `TestBatchPlanRejectsMissingRetrievalDate` before implementation | Pending | Pending |
| B2 | Manifest output is deterministic and preserves ledger counts/citations | Pending: focused `TestBatchPlanWritesDeterministicEvidenceManifest` before implementation | Pending | Pending |
| B3 | A failed connector is a named drop, not a batch abort or silent omission | Pending: focused `TestBatchGateContinuesAfterConnectorFailure` before implementation | Pending | Pending |
| B4 | Executable claims pass the real runtime preflight rather than a generator copy | Pending: focused `TestBatchGateUsesRuntimePreflight` before implementation | Pending | Pending |
| B5 | Surface metadata remains derived from `operations.json` | Pending: focused `TestBatchGateDropsSurfaceSyncDrift` before implementation | Pending | Pending |

## Test-first rule

Each named test is added and run against the baseline before its production
implementation. The observed failure is retained here; tests are not weakened,
skipped, or deleted to satisfy the batch report.
