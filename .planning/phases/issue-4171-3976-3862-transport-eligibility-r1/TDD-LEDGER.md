# TDD ledger — transport source eligibility club

Status values begin `planned`; exact commands and retained traces are filled during execute-phase.

| ID | Guarantee | Red assertion | Green proof | Status |
| --- | --- | --- | --- | --- |
| R1 | GitHub source eligibility is an explicit executable-stream allowlist | Production-composed `commits` cannot pass the declaration, and the red test cannot compile until the typed admission contract exists. | GitHub declaration names every executable bundle stream and no wildcard; `app.Open` composition accepts `commits`. | green — `TestOpenRegistersDefinitionOwnedProductionTransports` |
| R2 | The allowlist still refuses absence before effects | `SourceStreamIneligibleError` is undefined in both registry and `app.Open` tests. | `errors.As` matches `SourceStreamIneligibleError`; source requests, stages, target sends/rows, and checkpoint updates all remain zero. Case-equivalent `ISSUES` is also refused. | green — `TestPreflightReturnsTypedSourceStreamIneligibleErrorBeforeExecutorAccess` |
| R3 | Declarative source pages rather than collecting one capped issue batch | The source is hard-coded to `issues`, ten pages, and 1,000 buffered records. | A multi-page `commits` fixture emits five bounded batches containing all 103 records and binds each candidate to stream/ordinal/content. | green — `TestOpenComposedGitHubCommitsSourceEmitsEveryUnlimitedPageInBoundedBatches` |
| R4 | `max_pages` is honest | Omitted and unlimited settings cannot be distinguished by the transport adapter. | Omitted reads exactly one provider page; positive values cap; `0`/`all`/`unlimited` exhaust declared pagination. | green — `TestOpenComposedGitHubCommitsHonorsTransportMaxPages` |
| R5 | PostgreSQL polling is definition-selected in production | `app.Open` loads PostgreSQL with polling status `planned`, so its composed source cannot expose the exact shared executor reference. | Loaded PostgreSQL definition, `app.Open` factory composition, and native adapter resolve exact polling references/evidence and invoke the shared executor. | green — `TestOpenComposesPostgresImplementedSharedPollingRoute` |
| R6 | Resume uses strict `(cursor, complete primary-key tuple)` ordering | Existing scalar reader cannot represent duplicate watermark values or composite keys safely. | Native tests verify strict duplicate-cursor/composite-key planning and reject null/lossy/unsupported shapes; live row delivery remains pending. | green locally — `TestPostgresPollingPlan*` |
| R7 | Shared polling owns acknowledgement/checkpoint sequencing | A hand-built shared executor is the only proof and production PostgreSQL has no call chain. | Production route delegates to the shared executor; orchestrator commits only after durable acknowledgement and replay retains the last durable candidate. | green locally — `TestOpenComposesPostgresImplementedSharedPollingRoute`, `TestOrchestratorCommitsOnlyAfterDurableAcknowledgement` |
| R8 | Native page traversal is strict | Out-of-order or non-advancing native pages can reach the emitter in a faulty adapter. | Typed traversal refusal occurs before stage/send/checkpoint and leaves target/checkpoint unchanged. | green — `TestPostgresPollingPlanBindsCompleteCompositeKeyAndStrictResume` |
| R9 | Schema/auth/generation drift fails closed | Production adapter has no dynamic fingerprint/key binding to validate. | Schema/key/source-generation drift returns typed rebootstrap/admission outcomes before downstream effects; live auth proof remains pending. | green locally — typed planning and integration-gated assertions |
| R10 | Cancellation and concurrent-run fencing are safe | No composed-route proof exists. | Cancellation terminates bounded work without advancing the interrupted candidate; same-target concurrent runs retain state CAS fencing. | green locally — transport/app cancellation and conflict tests |
| R11 | Empty, single, and large cardinalities are honest | Current GitHub adapter's cap hides large inputs and polling has no production route. | Empty sends zero/does not advance existing state; single sends one; 103 GitHub records deliver in bounded pages. Live large certification remains pending. | green locally |
| R12 | Acknowledged replay and duplicate/out-of-order delivery remain idempotent | No cross-family production proof. | Replayed acknowledged page does not add target rows; duplicate/out-of-order inputs follow declared keyed/ordered policies or fail typed with zero effects. | green locally |
| R13 | Closed-spine conformance remains exact | API/database combinations can be hand-built without proving shipped composition. | `app.Open` loads exact references and accepted immutable evidence for API→database and database→database; mismatch/absence remains pre-I/O refusal. | green — production-composition tests |
| G1 | Certification guards track the declaration-owned source reference exactly | At merge base `ef3c71caf`, both certification guards pass; at this branch's `73280ed81` head they fail because they still expect the replaced `issue_label_source` ID. | The assertions now require `declarative_stream_source` both for successful resolution and the deliberately unregistered typed refusal. Neither guard is removed or weakened. | green — `TestCertificationDeclaredTransportPair*` |

## Initial red commands

```sh
go test -count=1 ./internal/synctransport -run 'Test.*SourceStream.*Ineligible'
go test -count=1 ./internal/app -run 'Test.*Transport.*(GitHub|Commits|Ineligible|Composition)'
go test -count=1 ./internal/connectors/native/postgres -run 'Test.*Polling.*(Definition|Transport|Resume|Refus)'
```

Initial failures and the corresponding green reruns are retained in `traces/`. A test that manually
registers the subject executor is supporting unit coverage only and cannot close R1, R5, R7, or
R13.

## Execution record

- **Red:** retained in `traces/red-stream-admission.txt` and
  `traces/red-postgres-polling-production.txt` before the production changes.
- **Green:** focused package, race, vet, build, generation, and repository drift gates are recorded
  in `VERIFICATION.md`; live Docker/PostgreSQL and authenticated GitHub certification are explicitly
  pending under the task's known external limits.
