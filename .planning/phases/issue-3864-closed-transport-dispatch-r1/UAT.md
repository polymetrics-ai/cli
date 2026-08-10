# UAT — #3864 closed transport dispatch

## Automated acceptance evidence

| Deliverable | Evidence | Result |
| --- | --- | --- |
| A canonical #3810 mode reaches closed registered roles rather than the legacy hard stop | `TestRunETLCanonicalFullAppendUsesRegisteredTransports` | Pass: fake source → stage → descriptor-resolved destination completed without legacy `Read`/`Write`. |
| Preflight refuses incomplete or unsafe admission before a source read | `TestPreflightRejectsClosedAdmissionFailuresBeforeSourceRead` | Pass: missing executor, family mismatch, unsupported mode, unsafe acknowledgement, absent strategy, and unavailable verifier all reject. |
| One orchestrator handles all four API/database direction pairings | `TestOrchestratorDispatchesFourClosedPairingsWithoutPairBranches` | Pass: API→API, API→database, database→API, and database→database each use source → warehouse stage → destination. |
| Checkpoint advancement waits for durable acknowledgement and respects cancellation | `TestOrchestratorCommitsOnlyAfterDurableAcknowledgement`, `TestOrchestratorStopsOnCancellationBetweenWarehouseStageAndApply`, and race coverage | Pass: invalid acknowledgement or cancellation yields zero commits; the transport race suite passes. |
| Destination planning uses a declared closed apply strategy, not `upsert` | `TestSyncTransportDescriptorResolvesDeclaredApplyStrategy` | Pass: `full_append` resolves `stage_append` / `append`. |
| JSON inspection and help explain eligibility without resolving credentials | `TestConnectorInspectProjectsUnsupportedSyncTransport`, guide test, and freshly built binary probe | Pass: absent roles are `unsupported`; the help/docs wording says declaration is not certification. |
| Both tracked correction regressions remain closed | T11/#4021 and T13/#4023 in `TDD-LEDGER.md` | Pass: empty descriptors reach preflight; `generic-http` fails closed after normalization. |

No human product-judgment step is required for these deterministic, fake-backed checks. This UAT
does not mark external conformance or certification as passed; those require separately accepted
evidence and have not run here.
