# TDD ledger — #4093

## Planned RED → GREEN checkpoints

| Slice | Red evidence | Green evidence | Status |
| --- | --- | --- | --- |
| Loader/projection | `go test -timeout 20m -run 'TestBundleLoadSyncTransport|TestDestinationTransportDescriptorRefusesChangeCapture' ./internal/connectors/engine ./internal/connectors` failed: the loaded `Definition().SyncTransport` was nil and every unknown/missing/wrong-version declaration was ignored. | The same command passes: versioned loader plus clone-safe `Bundle`/`Definition` projection makes a valid declaration observable while `additionalProperties: false`, strict decoding, version checks, and descriptor validation refuse the malformed cases. | green |
| Atomic definition composition | `go test -timeout 20m -run 'TestRegisterDeclaredTransports|TestDefinitionConformanceVerifier' ./internal/synctransport` did not compile because no definition composer or factory authority existed. | The same command passes: valid declarations build and register two source/destination pairs; unknown executor, invalid destination role, and altered evidence all observe zero builders, zero registrations, and zero source reads. | green |
| Production registrations | `TestOpenRegistersDefinitionOwnedProductionTransports` was absent while App composed only a GitHub wrapper. | The app-open preflight test observes the PostgreSQL `postgres_bounded_snapshot` source, GitHub `issue_label_destination`, and the warehouse-owned `local_parquet_warehouse` destination from real declarations. | green |
| Destination role rule | `go test -count=1 -timeout 20m ./internal/connectors -run TestDestinationTransportDescriptorAllowsChangeCaptureWithClosedApplyStrategy` fails because a destination that declares the closed `change_apply` strategy is refused before dispatch. | The connector test and `TestRunETLTransportDispatchesDeclaredChangeCaptureToClosedDestination` accept a legitimate `change_capture` destination, reach the exact `change_apply` plan and one apply; a malformed `change_capture`/`append` declaration still refuses before registration or I/O. | green |
| Local warehouse destination | `go test -count=1 -timeout 20m ./internal/app -run '^TestOpenRegistersDefinitionOwnedProductionTransports$'` failed because production PostgreSQL-to-warehouse preflight reported that warehouse had no destination declaration. | Production preflight resolves the exact local-Parquet reference; executor tests observe a connection-owned Parquet write, digest-checked read-back, and change-capture tombstone removal. | green |
| Closed-pair routing | Full `./internal/app` RED observed every legacy local-warehouse ETL route diverted into transport preflight after warehouse gained its real declaration. | `TestHasDeclaredSyncTransportRequiresBothEndpoints` proves a one-sided descriptor remains legacy while a two-sided malformed pair is still routed to fail closed; `go test -count=1 -timeout 20m ./internal/app` passes. | green |
| PostgreSQL live proof | Existing live native source test exercises current Go-owned descriptor. | Pending the mandated Docker/Colima run against the now definition-owned descriptor and production factory. | planned |
| Connector boundary repair | PR CI rejected the initial direct `internal/app` import of `native/postgres` as a connector-boundary violation. | `DefinitionFactoriesFromRegistry` gathers connector-provided factories without an App connector import; the production preflight test still resolves PostgreSQL and `go run ./cmd/connectorgen boundary . --json` passes. | green |
| Embedded definition repair | Full CI RED observed GitHub and PostgreSQL definitions with no transport declaration; the binary route refused the issue-label approval and native PostgreSQL tests found no source descriptor. | `defs.FS` embeds `*/sync_transport.json`; `TestPMBinaryExecutesIssueLabelWarehouseTransportLifecycle` and the PostgreSQL definition/registration tests pass, making the roles observable in the production binary. | green |
| Inspection projection repair | Full CI RED found `connectors inspect github --json` still asserted both transport roles were unsupported after the GitHub bundle declared them. | The CLI projection test observes `source.status=declared` and `destination.status=declared` from the production definition; runtime help, CLI manual, generated connector docs, and website guide explain the visible status. | green |
| Website data repair | Website CI RED detected the generated agent-guide data was stale after the transport declaration guidance changed. | `pnpm run gen:website-data` updates only `website/lib/docs.generated.ts`; the generated page includes the GitHub declared-role explanation. | green |
| CI regression repair | `go test -count=1 -timeout 20m ./internal/app -run 'TestGithubPullRequestsETLSupportsLegacyExecutableModes|TestRunETLTransportRejectsAcknowledgedCheckpointWithIncompatibleResume'` failed: the noncanonical GitHub mode ordering bypassed the exact issue-label route guard and destination change capture was refused despite declaring `change_apply`. | The focused app command, `TestRunETLTransportDispatchesDeclaredChangeCaptureToClosedDestination`, and `go test -count=1 -timeout 20m ./internal/cli -run '^TestGoldenTranscripts$'` pass: GitHub’s declaration is canonical, two-sided routing preserves legacy compatibility, and legitimate change capture reaches `change_apply`. | green |
| Generated artifact closure | CI RED: `TestSkillsGenerateMatchesTrackedSkills` found the tracked skills tree differed after the warehouse declaration regenerated its connector manual/skill. | A full source-derived sweep runs `agentcontractgen sync`, `connectorgen gen`, `surface-sync`, `certification-matrix --all`, docs/catalog generation, skills generation, golden transcript generation, and website-data generation. Their CI checks pass and only the expected GitHub and warehouse skill documents change. | green |
| R2 loaded synthetic connector route | No test loaded a second connector from its own `sync_transport.json`, sent it through `App.composeTransportRegistry`, and ran the resulting registration through `synctransport.Orchestrator`. | The new test loads a throwaway bundle, uses only its declared family/identifier/evidence plus a named test hook, and observes one source read, staged record, destination plan/apply/read-back, and committed checkpoint with no App/orchestrator/dispatch production edit. | green |
| R2 transient-stage retirement | `TestOrchestratorRetiresOwnedStageOnlyAfterCheckpointCommit` failed before the lifecycle patch: a checkpoint-committed page reported `retired receipts = 0, want 1`; the persistence-failure sibling defines the non-retirement safety boundary. | An opt-in generic stage retires exactly one receipt after a successful commit and none after a failure. R3 fixes the production connection-owned policy: it defers bounded reconciliation until the next generic execution so durable proof remains observable after ordinary Open. | green |
| R3 certification retention and deferred cleanup | CI `certify-timing` RED: `TestCertificationDeclaredTransportPairResolvesAndExecutes` reports a committed one-write pair but `warehouse_manifests=0 warehouse_parquet=0`, because eager generic retirement removes the durable proof before it is inspected. | The real certification proof retains exactly one owned manifest/Parquet after ordinary reopen. Deferred generic pre-execution reconciliation removes the exact committed receipt while preserving an active one, before source I/O. | green |

Every refusal test asserts the relevant side effect count is zero. No test relies
only on `err != nil` or a lack of panic.

## Red:

The loader/projection test failed against the base because the transport field
was absent; the composition test did not compile because the factory/composer
API did not exist; and the destination rule initially refused every
`change_capture` destination, including a declaration with the closed
`change_apply` strategy.

The post-push CI run failed after the production declaration activated the
existing app route guard. The focused app reproduction observed the GitHub
descriptor but rejected it as an issue-label route solely because its declared
mode order differed from `synccontract.AllModes()`; independently, the generic
destination fixture exposed that the role rule refused a legitimate closed
`change_capture` route.

The warehouse RED preflight reported `destination connector "warehouse" has no
declared destination transport`. Once that declaration was added, the full app
suite RED showed the previous one-sided routing predicate attempting transport
preflight for legacy sources. Both failures occurred before the legacy source
read or destination write that the corresponding green tests preserve.

The next CI RED was `TestSkillsGenerateMatchesTrackedSkills`: regenerating the
warehouse connector docs had changed the source content for the generated
skills tree, but the tracked tree had not been regenerated. The mismatch is an
observable byte-for-byte generated output difference, not merely a successful
command exit.

## Green:

The same focused test commands now pass with a versioned bundle loader,
definition projection, prevalidated atomic registration, factory-held evidence
authority, production GitHub/PostgreSQL declarations, and a destination-role
refusal before any executor state changes.

The CI-regression focused app command now observes GitHub-to-warehouse using
the legacy executable path. `TestRunETLTransportDispatchesDeclaredChangeCaptureToClosedDestination`
observes one source read, warehouse stage, destination plan, and destination
apply with `change_apply`; the malformed declaration composition test observes
zero builds, registrations, reads, plans, and applies. The regenerated CLI
golden transcript also records the observable declared GitHub source and
destination roles.

`TestLocalWarehouseDestinationExecutorWritesAndReadBacksConnectionOwnedParquet`
observes the reopened workset row in the owned Parquet table and refuses a
post-acknowledgement table mutation during read-back. Its change-capture
companion observes a valid tombstone remove the keyed row. The full app suite
then proves both-sided declaration routing preserves every legacy warehouse
flow while leaving malformed two-sided pairs closed.

The generated-artifact repair reruns every applicable source generator in one
pass. `TestSkillsGenerateMatchesTrackedSkills`, `TestGoldenTranscripts`,
`TestGoldenDocsGenerateMatchesTrackedCLIManuals`, and
`TestDocsGenerateIncludesConnectorCatalog` pass together; agent-contract,
connector surface/certification, docs/catalog, GitHub parity, and website-data
drift checks also pass. The generated-path diff is empty after accepting only
the updated GitHub and warehouse skill content.

## R2 Red:

`go test -count=1 -timeout 20m ./internal/synctransport -run
'^TestOrchestratorRetiresOwnedStageOnlyAfterCheckpointCommit$'` failed before
the production lifecycle patch. Its checkpoint-committed subtest observed
`retired receipts = 0, want 1; stage artifacts must survive an uncommitted
checkpoint`. The failure identified the missing post-commit retirement path;
the failing checkpoint-persistence sibling simultaneously established that no
retirement may occur before the durable commit.

## R2 Green:

The same focused test now passes, with exactly one retirement only after the
commit callback accepts the acknowledged checkpoint and zero after the callback
returns an error. `TestDeferredReconciliationRetiresOnlyCommittedConnectionOwnedTransportStages`
then simulates a process kill after that durable checkpoint: ordinary Open
retains both durable worksets, and the next bounded generic reconciliation
removes only the exact committed receipt's manifest, WAL, and Parquet while
preserving an active receipt. The independent loaded-synthetic-connector proof
passes through `engine.Load`, App composition, the registry, and the generic
orchestrator with no production registration, App dispatch, or orchestrator
change for that connector.

## R3 Red:

CI's `make certify-timing` failure was reproduced locally with
`go test -count=1 -timeout 20m ./internal/connectors/certify -run
'^TestCertificationDeclaredTransportPairResolvesAndExecutes$'`. The actual
declared GitHub pair read twice, applied once, and committed its checkpoint, but
the report failed closed with `warehouse_manifests=0 warehouse_parquet=0`.
That evidence established that eager post-commit retirement and eager Open
reconciliation, not the certification adapter, were deleting the receipt.

## R3 Green:

`make certify-timing` now passes, including the declared-pair proof. Full
`./internal/connectors/certify` and `./internal/app` suites pass. The new app
test proves a reconciliation error stops before source or destination I/O;
the deferred-reconciliation fixture proves a fresh Open retains proof and the
next generic cleanup removes only the committed owned artifact.
