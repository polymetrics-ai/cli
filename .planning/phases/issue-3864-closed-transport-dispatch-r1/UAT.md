# UAT — #3864 closed transport dispatch

## Automated acceptance evidence

| Deliverable | Evidence | Result |
| --- | --- | --- |
| A canonical #3810 mode reaches closed registered roles rather than the legacy hard stop | `TestRunETLCanonicalFullAppendUsesRegisteredTransports` | Pass: fake source → stage → descriptor-resolved destination completed without legacy `Read`/`Write`. |
| Preflight refuses incomplete or unsafe admission before a source read | `TestPreflightRejectsClosedAdmissionFailuresBeforeSourceRead` | Pass: missing executor, family mismatch, unsupported mode, unsafe acknowledgement, absent strategy, and unavailable verifier all reject. |
| One orchestrator handles all four API/database direction pairings | `TestOrchestratorDispatchesFourClosedPairingsWithoutPairBranches` | Pass: API→API, API→database, database→API, and database→database each use source → warehouse stage → destination. |
| Checkpoint advancement waits for durable acknowledgement and persists before post-ack failure | T15, T20, T21, `TestOrchestratorCommitsOnlyAfterDurableAcknowledgement`, `TestOrchestratorStopsOnCancellationBetweenWarehouseStageAndApply`, and `TestOrchestratorCommitsAcknowledgedPageBeforeReturningCancellation` | Pass: invalid acknowledgement or pre-apply cancellation yields zero commits; an acknowledged page validates its source identity before persistence, preserves prior completion metadata, and advances only its expected target stream across all seven modes. |
| Destination planning uses a declared closed apply strategy, not `upsert` | `TestSyncTransportDescriptorResolvesDeclaredApplyStrategy` | Pass: `full_append` resolves `stage_append` / `append`. |
| JSON inspection and help explain eligibility without resolving credentials | T16 and T19, `TestConnectorInspectProjectsUnsupportedSyncTransport`, and generated docs/transcripts | Pass: absent or invalid roles are `unsupported`; valid siblings remain declared, runtime full validation stays closed, and help/docs say declaration is not certification. |
| Both tracked correction regressions remain closed | T11/#4021 and T13/#4023 in `TDD-LEDGER.md` | Pass: empty descriptors reach preflight; `generic-http` fails closed after normalization. |
| Review correction loop 4 remains locally bounded | T15–T19 in `TDD-LEDGER.md` | Pass: deterministic fakes cover state, descriptor, verifier, byte-copy, and help/manual fixes without provider, credential, daemon, or warehouse-format changes. |
| Review correction loop 5 remains locally bounded | T20–T21 in `TDD-LEDGER.md` | Pass: deterministic fake-backed tests reject identity/generation mismatches and stale stream writers while retaining unrelated state, cancellation, state-store, and seven-mode controls. |

No human product-judgment step is required for these deterministic, fake-backed checks. This UAT
does not mark external conformance or certification as passed; those require separately accepted
evidence and have not run here.
