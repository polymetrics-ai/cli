# Foundation Post-Fix TDD Ledger r1

This ledger is bound to the frozen 46-finding canonical review in `POSTFIX-REVIEW.md`. A test is recorded as green only when it passes from the same commit as its production change; generated artifacts are part of the green state, not a later cleanup.

| Group | Finding set | Red contract | Green / regression evidence | State |
| --- | --- | --- | --- | --- |
| 1 | B01, B02, B03, B09, B12, W01 | `TestGitHubParityGenerationOrderIsCommutative`; `TestSourceProjectionGapOperationsCannotMasqueradeAsImplemented`; `TestGoogleAdsGeneratedPOSTReadsAcceptDeclaredNestedObjects`; v2 projection digest mutation; deleted route/parameter and semantic update/delete surface-sync cases. | `go test -timeout 20m ./cmd/connectorgen`; Node parity-order and combined-ledger checks; `connectorgen validate`; `surface-sync --check`; all six affected source IDs have installed coverage. | green; remote `d3bf5da0e6a4575628dd76dd94a7522220f9d3df` |
| 2 | B04-B08, W02 | `TestGraphQLOperationVariablesRequiresExactlyOnePaginationDirection`; `TestOperationDirectReadBackwardGraphQLPaginationUsesPreviousPageInfo`; `TestGraphQLIntUsesSigned32BitDomain`; `TestGeneratedGraphQLContractsClassifySecretInputsAndBoundedIdentitySelections`; `TestWebsiteFlagProjectionPreservesEverySafetyProperty`; `TestRenderCommandSurfaceCommandRendersSafetyConstraints`. | Focused generator, engine/schema, commandrunner preflight, website, guide/skills, generated GitHub artifacts, source-import, surface-sync, certification-candidate/sweep checks. | green; remote `0565f3fd6d152b38f2062aac5dd0df29170b6d4e` |
| 3 | B13-B14, B17, B19, B24 | Classified secrets versus ordinary IDs/headers; error-bearing direct/native result; status receipt; SQS success/error receipt. | connsdk, engine, commandrunner, native SQS, App, and CLI receipt suites. | green; remote `58c86d18bd27e55f334cea37f263dd4cdf7540ee` |
| 4 | B15-B16, B18, B21, B23, B25, W03-W04 | Hook sealed bytes/compound receipt; retry/redirect/cancel receipt; >2^53 CLI value; hostile cursor; SQS redirect; idempotency header; minLength witness. | engine, commandrunner, connsdk, native SQS, CLI, and structured-body regressions. | green; remote `b0eb22feb7f413d15f747b3f78d62c6c46e314b9` |
| 5 | B22, W05 | Existing destination collision/foreign file, error cleanup, and symlink race test each fail before publication. | binary output and `go test -race` multipart publication cohorts. | green; remote `2bddbf5387d323a0dbf074074cf43fa2d40b60b5` |
| 6 | B20, B26, B33, B36, W06-W07 | Stale/revoked authorization or stream owner reaches an effect; clone mutation leaks; indeterminate durable commit; expired park retries. | App, transport, coordination, Arrow/race, and auth fence regressions. | green; remote `51ab7639e30bb2fb5585d853beba6a2550d84def` |
| 7 | B27-B32 | Budget stop looks like EOF; self-certification; receipt-free readback; >2^53 cloning/comparison; one shared deadline. | App, synctransport, engine, and provider-readback behavior suites. | green; remote `6084b1c1275b2dbe01fc49aba25785677fd8fd52` |
| 8 | B34-B38, W08 | Persisted terminal result is hidden; ambiguous finalization invents a run; CDC accepts swapped artifact; post-checkpoint error returns failure; declared route error disappears. | App, CLI, CDC/restart, transport, and state recovery suites. | green; remote `0b061b0f3149ba9b050f6a7b7ec3cc2494c08f0c` |
| 9 | B10-B11, B01 artifact closure | Evidence/certification metadata for another or stale implementation, subject, or generated surface is accepted. | Exact implementation/evidence graph, complete subject-component mutation set, and source/CLI/docs/website/skills/ledger/matrix/candidate/sweep closure checks. | red contract frozen |

## Group 8 execution plan (2026-08-21)

- **GSD/manual fallback and skills:** `scripts/gsd doctor`, all five required
  command sources/prompts, and `go run ./cmd/agentcontractgen check` passed.
  This single-owner non-Pi lane applies the generated workflow inline because
  its runtime cannot provide a compatible isolated GSD worker. The completed
  Group-7 remote base is `6084b1c1275b2dbe01fc49aba25785677fd8fd52`.
  `golang-how-to`, `golang-cli`, `golang-testing`, `golang-error-handling`,
  `golang-security`, `golang-safety`, `golang-design-patterns`,
  `golang-structs-interfaces`, `golang-context`, `golang-concurrency`, and
  `golang-database`, and `golang-documentation` govern Group 8. CLI help/manual/website parity is not
  applicable to B37 because its private recovery evidence changes no command,
  flag, help text, output schema, or generated surface; B34/B35 will revisit
  that checklist before their CLI presentation work.
- **B37 scope and ownership:** only the CDC receipt/stage/native PostgreSQL
  recovery and App change-capture paths (`transaction_stage.go`,
  `cdc_v2.go`, `change_capture_dispatch.go`, their tests, connectors receipt
  contract, and this ledger). Group 8 B34/B35 and B38/W08 files are untouched
  until their separately ordered parcels. The manifest is private durable
  evidence: version, connection, stream, generation, transaction key, record
  count, staged content digest, and exact raw-WAL/final-table digests. It must
  be atomically durable before its receipt can permit LSN progress.
- **B37 Red:**
  `TestCDCRecovery_ReceiptBindsExactWarehouseArtifacts` exercises untouched
  restart recovery, missing/truncated/swapped raw/table artifacts, a stale
  manifest generation, and a manifest checksum mismatch. Every bad case must
  fail as typed artifact reconciliation before checkpoint/LSN acknowledgement,
  without a second transaction receive; the untouched case restores exactly
  once without replay.

### Group 8 B37 observed red and compatibility set (2026-08-21)

- The named App red command,
  `go test -timeout 20m ./internal/app -run '^TestCDCRecovery_ReceiptBindsExactWarehouseArtifacts$' -count=1 -v`,
  first failed to compile because neither the typed reconciliation error nor a
  connection/stream/generation-owned manifest path existed. After adding only
  those assertion scaffolds, the unchanged happy case passed while all five
  bad restart cases incorrectly returned completed runs with an `lsn-1`
  checkpoint: missing manifest, truncated raw WAL, swapped final table, stale
  manifest generation, and mismatched raw-WAL checksum. This is the frozen
  pre-fix proof that regular-file checks advanced source progress.
- The first focused package probe,
  `go test -timeout 20m ./internal/connectors ./internal/connectors/database ./internal/connectors/native/postgres -run 'Test(CDC|CommittedTransaction|TransactionReceipt|PGOutput)' -count=1`,
  had one complete failure after the manifest path was introduced:
  `TestPGOutputV2StreamCommitUsesDurableTransactionReceiverBeforeCheckpoint`.
  Its old synthetic durable receiver intentionally returned an unbound receipt,
  so its replay correctly reached the new missing-manifest refusal. The test
  now supplies a closed, exact synthetic manifest and asserts that the native
  stage persisted/restored it; its original receive → receipt → checkpoint →
  acknowledgement ordering assertion remains exact.

### Group 8 B37 green evidence (2026-08-21)

- The durable receipt now carries a private, strict version-1 artifact manifest
  from the App receiver through the native PostgreSQL stage, including its
  transaction-stage content digest, connection, stream, generation, record
  count, and raw-WAL/final-table SHA-256 identities.  The manifest itself is
  atomically written under the connection-owned warehouse root and directory
  synced before a receipt can permit source LSN progress.  Restart recovery
  recreates the receipt only from the persisted private manifest, checks its
  exact identity against the stage record and durable manifest file, and
  re-digests both artifacts before restoration; a missing, corrupt, stale, or
  mismatched artifact returns `ChangeCaptureArtifactReconciliationError`
  before checkpoint/acknowledgement.
- Focused happy/bad/restart proof passed:
  `go test -timeout 20m ./internal/app ./internal/connectors/database ./internal/connectors/native/postgres -run 'Test(CDCRecovery_ReceiptBindsExactWarehouseArtifacts|CommittedTransactionStagePersistsPrivateArtifactManifestAcrossRestart|PGOutputV2StreamCommitUsesDurableTransactionReceiverBeforeCheckpoint)' -count=1 -v`.
  The exact App cases retain one receive/zero restore/zero LSN checkpoint for
  every bad artifact and one receive/one restore/one checkpoint for untouched
  artifacts.  The stage test proves the private manifest survives a reopen.
- Strict private receipt decoding proof passed:
  `go test -timeout 20m ./internal/connectors -run '^TestCDCArtifactManifestRestorationIsStrictAndBounded$' -count=1 -v`.
  Missing, unknown-field, trailing-value, and over-8-KiB receipt manifests are
  rejected rather than becoming a fallback receipt.
- Completion-tracked final package gate passed from this source/test set:
  `go test -timeout 20m ./internal/app ./internal/connectors ./internal/connectors/database ./internal/connectors/native/postgres -count=1`
  (`internal/app` 253.135s; `connectors` 0.616s; `database` 8.177s;
  `native/postgres` 2.473s).  It produced no failure set.  The first focused
  package-probe compatibility failure remains dispositioned above; no broad
  gate failed after its exact manifest fixture was added.
- The final focused race gate passed with the full B37 contract:
  `go test -race -timeout 20m ./internal/app ./internal/connectors ./internal/connectors/database ./internal/connectors/native/postgres -run 'Test(CDCRecovery_ReceiptBindsExactWarehouseArtifacts|CDCArtifactManifestRestorationIsStrictAndBounded|CommittedTransactionStagePersistsPrivateArtifactManifestAcrossRestart|PGOutputV2StreamCommitUsesDurableTransactionReceiverBeforeCheckpoint)' -count=1 -v`
  (`app` 40.937s; `connectors` 2.168s; `database` 1.715s;
  `native/postgres` 3.338s), with no data race or failure.

### Group 8 B34/B35 red-contract plan (2026-08-21)

- **Scope and ownership:** the ordinary CLI ETL rendering boundary and the App
  terminal/reverse-finalization state boundary (`internal/cli/cli.go`,
  `internal/app/app.go`, and their focused tests).  This parcel must share one
  exact durable reload outcome; it does not touch the Group-8-C transport,
  route, or orchestrator files reserved for B38/W08.
- **B34 red:** `TestCLI_OrdinaryETLFailurePublishesOneTerminalRun` uses the
  real persisted CLI file-source route.  Source I/O fails after `beginRun`
  has durably recorded a failed terminal Run.  The nonzero CLI invocation must
  emit exactly one `ETLRun` JSON envelope equal to that stored terminal run,
  not an `Error` envelope or an in-memory substitute.
- **B35 red:** `TestReverseFinalization_DoesNotPublishUnpersistedRun` invokes
  both regular and direct-write finalizers with a provider partial/error.  A
  pre-rename state-lock failure must return zero run plus joined operational /
  persistence error and retain no terminal run; a post-rename directory-sync
  uncertainty must reload the exact persisted terminal reverse run and plan,
  preserving the original provider error and the typed persistence outcome.

### Group 8 B34/B35 observed red evidence (2026-08-21)

- `go test -timeout 20m ./internal/app -run
  '^TestReverseFinalization_DoesNotPublishUnpersistedRun$' -count=1 -v`
  failed all four frozen cases.  Both regular and direct-write paths returned
  an in-memory failed `ReverseRun` after the definite pre-rename error; both
  post-rename directory-sync cases returned only the provider partial-write
  error, losing the state persistence outcome even though a terminal run and
  failed plan had reached the store.
- `go test -timeout 20m ./internal/cli -run
  '^TestCLI_OrdinaryETLFailurePublishesOneTerminalRun$' -count=1 -v` reached
  the real missing-file read after a durable failed Run and printed an `Error`
  JSON envelope instead of the stored `ETLRun` envelope.  The test's initial
  non-comparable `Run` assertion was corrected to `reflect.DeepEqual` before
  this behavioral red run; no production behavior changed during that test
  correction.

### Group 8 B34/B35 package-gate failure set (2026-08-21)

The one completion-tracked package command, `go test -timeout 20m
./internal/app ./internal/cli -count=1`, exited `1` after `254.915s` for App
and `804.384s` for CLI.  This complete output is frozen before a correction;
the exact observed cases and initial dispositions are:

| Exact failing test / case | Observed cause | Required disposition |
| --- | --- | --- |
| `TestRunETLTransportAcknowledgedFailureReturnsTruthfulPersistenceOutcome/committed_unlock_failure` | The new reload branch shadowed the operational `runErr` with the terminal-state lookup error, so the returned joined error contained only the persistence outcome. | Restore the original operational error to the joined outcome and prove both committed and indeterminate paths with the existing focused App test. |
| `TestRunETLTransportAcknowledgedFailureReturnsTruthfulPersistenceOutcome/indeterminate_directory_sync_failure` | Same shadowing defect; the terminal run was durable but the post-acknowledgement source error was dropped. | Same shared App correction; no test relaxation. |
| `TestCertifyCLISingleConnectorPassExitsZero` | The two declared `*_deduped` stages are typed **pre-I/O** refusals. The focused persisted-state proof shows each has exactly one durable failed run, so a prior nonterminal/`Error` inference from the scripted fixture was false. | Require a terminal stored/reloaded run before one `ETLRun` is emitted. Assert exact stored `ETLRun` plus categorized exit 1 for both stages; the scripted certification fixture must project the same terminal contract. |
| `TestFreshBinaryProvider401IsCredentialErrorWithoutWritesOrCheckpointAdvance` | This is a GitHub declared direct-read, not ordinary ETL. It correctly returned the post-provider `ConnectorCommandDirectRead` result with the bounded 401 receipt and categorized nonzero exit; its stale test expected a stand-alone `Error` envelope and hid that provider evidence. | Strengthen the test to assert the retained 401 status/receipt, `auth/credential_error`, secret masking, and unchanged checkpoint/no writes. No B34 production path changes. |
| `TestReverseETLToGitHubCreatesPullRequestAfterApproval` | The provider test server's create-PR response was accepted as a receipt but the following response binding rejected its implicit `text/plain` media type as not JSON.  The full run therefore exposed a pre-existing declared-response parsing gap, not a B34/B35 terminal reload path. | Reproduce in the focused reverse CLI test and fix the declaration-owned response-body interpretation without changing its closed/bounded route or omitting receipt fields; record the focused proof before any package rerun. |

**Supervisor semantic correction (2026-08-21):** the preceding preliminary
CLI-envelope dispositions are superseded.  A returned terminal `Run` is the
App's durable presentation proof: a successful state update returns its stored
terminal run; a may-have-committed update returns only the exact reload; and a
definite pre-rename/no-commit result returns `Run{}`.  Consequently ordinary
source/provider, parked, CDC-failed, and runtime-recorder errors that accompany
a returned terminal run must emit exactly one `ETLRun` envelope while retaining
their original categorized nonzero exit.  Only a zero/nonterminal run emits a
typed `Error` envelope. The focused real-CLI state proof shows both
certification pre-I/O stages *do* store an exact terminal failed Run, so their
one-envelope contract is `ETLRun` with exit 1; a zero/nonterminal result would
instead remain `Error`. The GitHub 401 is instead a post-provider direct-read result
(not a `Run`) and retains its bounded receipt plus typed inline error rather
than a stand-alone `Error` envelope.

### Group 8 B34/B35 green evidence and complete failure disposition (2026-08-21)

- **Shared durable outcome:** `reloadExactTerminalState` is the single
  post-rename/unlock recovery boundary. `completeRunWithAcknowledgedTransportState`,
  `failRunWithAcknowledgedTransportState`, `failRunWithResult`, and both reverse
  finalizers either return the exact stored terminal object or return zero on a
  definite no-commit. The acknowledged failure path now joins the original
  operational error with the persistence outcome instead of shadowing it with
  a terminal-lookup error.
- **Ordinary CLI result rule:** `shouldPresentETLTerminalRun` accepts only a
  complete, terminal `Run` returned by App. That return is the durable
  presentation proof: normal persisted source/provider/parked/CDC/runtime
  errors retain one exact `ETLRun`; may-have-committed results were first
  reloaded by App; zero or nonterminal values remain the normal typed `Error`
  path. `alreadyReportedExecutionError` retains the classified nonzero exit
  without adding a second JSON envelope.
- **Focused happy/bad/edge green:**
  `go test -timeout 20m ./internal/app -run
  'Test(RunETLTransportAcknowledgedFailureReturnsTruthfulPersistenceOutcome|ReverseFinalization_DoesNotPublishUnpersistedRun|OrdinaryETLTerminalPersistenceReloadsExactDurableRun)$'
  passed (`8.325s`). The cases cover definite pre-rename zero state, committed
  unlock and directory-sync uncertainty with exact reload, ordinary/direct
  reverse finalization, and preservation of the original source/provider error.
- **Installed CLI and certification state green:** the focused CLI suite
  covering ordinary source failure, runtime-recorder failure, provider 401,
  reverse GitHub approval, and terminal proof passed (`20.769s` before the
  final certification-state addition). The final production-shaped
  `TestCLI_CertificationPreIORefusalsPersistExactTerminalRun` passed for both
  deduped modes (`3.56s`), proving exit `1`, failed terminal `ETLRun`, and
  byte-for-byte equality with the persisted run. The certification projection
  `TestSourceStagesAgainstSample` then passed (`1.171s`) with its scripted
  terminal fixture mirroring that durable contract.
- **Race/package gate:** the final focused
  `go test -race -timeout 20m ./internal/app ./internal/cli
  ./internal/connectors/certify -run
  'Test(RunETLTransportAcknowledgedFailureReturnsTruthfulPersistenceOutcome|ReverseFinalization_DoesNotPublishUnpersistedRun|OrdinaryETLTerminalPersistenceReloadsExactDurableRun|CLI_OrdinaryETLFailurePublishesOneTerminalRun|ETLTerminalPresentationRequiresDurableUncertainCommit|CLI_OrdinaryETLRuntimeRecorderFailurePublishesTerminalRun|CLI_CertificationPreIORefusalsPersistExactTerminalRun|FreshBinaryProvider401IsCredentialErrorWithoutWritesOrCheckpointAdvance|ReverseETLToGitHubCreatesPullRequestAfterApproval|SourceStagesAgainstSample)$'
  passed: App `91.422s`, CLI `107.195s`, certification `3.716s`, with no race
  report. The earlier 792-second broad failure set is fully enumerated above;
  these focused package gates are the required post-disposition proof, while
  the next heavyweight all-App/all-CLI run remains reserved for the final
  exact Group-8 SHA.
- **Static/generated/help green:** `go vet ./internal/app ./internal/cli
  ./internal/connectors/certify`, `go build ./cmd/pm`, and `go run
  ./cmd/connectorgen surface-sync --check` all passed; surface-sync scanned
  552 connectors and corrected zero fields. The built binary also passed `pm
  help etl`, `pm etl`, and `pm etl run --help`. No CLI flags, help text,
  manual/website source, or generated command surface changed, so help/manual/
  website regeneration is not applicable to this output-boundary correction.
- **Five-case disposition:** the App committed/unlock and directory-sync cases
  are fixed by preserving `runErr`; the certification two-stage failure is
  fixed by terminal-proof presentation backed by persisted state; the GitHub
  401 now asserts its complete `ConnectorCommandDirectRead` 401 receipt and
  inline `auth/credential_error`; and the GitHub reverse fixture explicitly
  declares its JSON response media type, retaining strict declaration-owned
  response parsing. No route, body authority, operation, provider result, or
  secret boundary was widened.

### Group 8 B34/B35 certification-envelope reconciliation (2026-08-21)

- The durable-state assertion is `TestCLI_CertificationPreIORefusalsPersistExactTerminalRun`.
  It runs each `*_deduped` pre-I/O refusal through the installed CLI route,
  reads the persisted state file, and requires exit `1` plus exactly one
  failed terminal `ETLRun` byte-for-byte equal to that stored run.  The
  contract is therefore unambiguous: any exact durable terminal run emits an
  `ETLRun` regardless of the accompanying operational error; only a definite
  no-commit or nonterminal state emits a typed `Error` without a run.  The
  separate GitHub 401 remains `ConnectorCommandDirectRead` with its complete
  bounded 401 receipt and inline `auth/credential_error`.
- **Green recheck:** `go test -timeout 20m ./internal/cli -run
  '^TestCLI_CertificationPreIORefusalsPersistExactTerminalRun$' -count=1 -v`
  passed on the B34/B35 remote base in `4.710s` (test body `3.53s`).

### Group 8 B38/W08 red-contract plan (2026-08-21)

- **Scope and ownership:** `internal/synctransport/orchestrator.go`,
  `internal/app/transport_dispatch.go`, `internal/app/etl_mode_dispatch.go`,
  the declaration-owned approval marker stores, their focused tests, and this
  ledger.  This parcel consumes the B34/B35 durable-terminal rule without
  weakening it.  It does not change Group 9, command authority, connector
  definitions, or generated artifacts by hand.
- **B38 red:** `TestTransport_PostCheckpointBookkeepingFailureRemainsDelivered`
  will inject retiring-stage, declarative marker, and managed marker cleanup
  failures after a durable checkpoint.  It must prove one provider mutation,
  retained checkpoint and complete provider output, durable
  `delivered_reconciliation_required` state, and idempotent restart cleanup
  with zero replay.
- **W08 red:** `TestETLRouteSelection_PreservesDeclaredPreflightError` will
  exercise a real `declarative_stream_source` paired with declared destination
  variants that are unregistered, have the wrong closed marker, or omit the
  binding/action/conformance.  Each must retain the typed preflight error and
  cause zero source/stage/provider I/O; the two-sided declaration-absent
  legacy case must remain an intentional non-transport selection.

### Group 8 B38/W08 observed red and green evidence (2026-08-21)

- **Final package-gate failure set (before the W08 compatibility correction):**
  `go test -timeout 20m ./internal/app -count=1` exited `1` after `254.808s`.
  Its complete failure set is the three named subcases of
  `TestGithubPullRequestsETLSupportsLegacyExecutableModes`:
  `full_refresh_append_duplicates` and
  `full_refresh_overwrite_replaces_final` now returned
  `DeclaredDestinationRouteError` for the declared
  `local_parquet_warehouse` executor; and
  `incremental_append_filters_older_cursor_and_appends_inclusive_cursor`
  returned `source transport does not support sync mode "incremental_append"`.
  The regression was introduced by removing the preflight result's
  declaration-owned local-warehouse route from the selector. The correction
  must retain the W08 requirement that an unregistered or unmarked
  two-sided declaration returns its exact typed refusal before I/O, while
  preserving this existing bounded local-warehouse representation and its
  legacy source-mode fallback.

- **Red — checkpoint evidence was lost:** `go test -timeout 20m
  ./internal/synctransport -run
  '^TestTransport_PostCheckpointBookkeepingFailureRemainsDelivered$' -count=1
  -v` initially failed to compile because `Result` had neither the durable
  reconciliation flag nor the typed outcome. Before the correction, each
  `retireCommittedReceipts` call returned its local cleanup error before the
  committed checkpoint was placed in `Result`, making a completed provider
  effect look replayable.
- **Red — App state had no terminal representation:** the focused App restart
  proof initially failed to compile on the absent
  `ETLRunStatusDeliveredReconciliationRequired` and
  `Run.DeliveryReconciliation` fields. The route red
  `TestETLRouteSelection_DeclarativeSourcePreservesDeclaredDestinationPreflightError`
  then ran and observed `selected=false reason=declaration_absent err=<nil>`
  for a real `declarative_stream_source` and an unregistered declared
  destination. That was the B38/W08 regression: the route checker hid the
  registry's pre-I/O declared-operation refusal.
- **Green — B38 happy/bad/edge/restart:** the orchestrator now sets its exact
  committed checkpoint, page/count evidence, complete destination results,
  and `DeliveredReconciliationRequired` before any retirement attempt. A
  typed `DeliveredReconciliationRequiredError` wraps only the local cause.
  `TestTransport_PostCheckpointBookkeepingFailureRemainsDelivered` and
  `TestTransport_DeferredCheckpointRetirementFailureRemainsDelivered` passed;
  they cover ordinary and deferred-final checkpoint paths. The App test
  `TestRunETLTransportPostCheckpointBookkeepingPersistsDeliveredReconciliationAndRepairsWithoutReplay`
  passed through a fresh `Open`, retaining the exact occurrence ID, then
  completed the same stored run with zero additional source/apply/read-back
  calls. `TestDeliveredReconciliationApprovalMarkersRepairOrFailClosedWithoutReplay`
  passed managed-target, declarative typed-destination, and missing-marker
  cases: both declared markers repair exactly once from durable state, while
  unknown-marker and corrupt-state rows remain terminal
  reconciliation-required with zero I/O.
- **Green — W08 typed route preservation:** the early
  `RegisteredDestination` screen was removed. Every two-sided declaration
  other than the established structural local-warehouse legacy representation
  reaches registry `Preflight`; an unregistered destination returns the new
  typed `DestinationExecutorUnregisteredError`, and a resolved destination
  without one of the bounded managed/definition/local-warehouse representations
  returns typed `DeclaredDestinationRouteError`. The focused unregistered and
  unmarked-destination route proofs passed with zero source, stage, provider,
  or legacy I/O.
- **Green — bounded local-warehouse compatibility:** the red
  `TestETLRouteSelection_DeclarativeSourceKeepsBoundedLocalWarehouseLegacyModes`
  named exactly the three legacy modes that do not belong to the dedupe
  transport cohort: `full_append`, `full_overwrite`, and
  `incremental_append`. The focused green command
  `go test -timeout 20m ./internal/app -run
  '^(TestETLRouteSelection_DeclarativeSourceKeepsBoundedLocalWarehouseLegacyModes|TestGithubPullRequestsETLSupportsLegacyExecutableModes|TestETLRouteSelection_DeclarativeSourcePreservesDeclaredDestinationPreflightError|TestETLRouteSelection_DeclarativeSourceRejectsUnmarkedResolvedDestination)$'
  -count=1 -v` passed in `4.504s`. The structural helper requires the exact
  local-warehouse destination reference, matching conformance, durable
  acknowledgement, and local materializer marker; it neither branches on a
  connector name nor admits generic/raw write authority. Dedupe and history
  modes still use the registered transport executor, while the three named
  legacy modes retain their existing bounded ordinary representation.
- **B34/B35 envelope reconciliation retained:**
  `TestCLI_CertificationPreIORefusalsPersistExactTerminalRun` was rerun after
  the B38 status addition and passed. It remains the durable-state authority:
  any complete stored terminal run — including
  `delivered_reconciliation_required` — is emitted as exactly one `ETLRun`
  with its categorized nonzero operational error. A definite no-commit or
  nonterminal result is the only `Error`/no-run case. GitHub 401 remains its
  independent `ConnectorCommandDirectRead` 401 receipt with inline
  `auth/credential_error`.
- **Final local Group-8 gates:** the complete post-correction App package
  gate, `go test -timeout 20m ./internal/app -count=1`, passed in `255.344s`.
  The captured prior 3-case GitHub regression set is disposed above; this
  final run produced no failure set. The focused package proof for B38/W08,
  legacy GitHub execution, and the durable certification/direct-read envelope
  passed across App, synctransport, and CLI (`7.239s`, `0.539s`, and
  `21.118s`). The captured focused race gate passed with no race report:
  `go test -race -timeout 20m ./internal/app ./internal/synctransport -run
  '^(TestTransport_PostCheckpointBookkeepingFailureRemainsDelivered|TestTransport_DeferredCheckpointRetirementFailureRemainsDelivered|TestRunETLTransportPostCheckpointBookkeepingPersistsDeliveredReconciliationAndRepairsWithoutReplay|TestDeliveredReconciliationApprovalMarkersRepairOrFailClosedWithoutReplay|TestETLRouteSelection_DeclarativeSourcePreservesDeclaredDestinationPreflightError|TestETLRouteSelection_DeclarativeSourceRejectsUnmarkedResolvedDestination|TestETLRouteSelection_DeclarativeSourceKeepsBoundedLocalWarehouseLegacyModes|TestGithubPullRequestsETLSupportsLegacyExecutableModes)$'
  -count=1 -v` (App `59.937s`; synctransport `1.640s`). `go vet
  ./internal/app ./internal/synctransport ./internal/cli`, `go build ./cmd/pm`,
  `go run ./cmd/connectorgen validate` (552, zero findings), `go run
  ./cmd/connectorgen surface-sync --check` (552, zero drift), and `go run
  ./cmd/pm docs validate --connectors-dir docs/connectors` passed. Website
  generated-doc parity passed through `npm run test:scripts` (34/34); the
  installed environment has no `tsc` executable, so no dependency was added
  solely to run an untracked typecheck.

## Group 7 red-contract plan (2026-08-21)

- **GSD/manual fallback and skills:** this is the preserved single-owner
  postfix phase. The existing `POSTFIX-EXECUTION.md` delivery header and
  manual GSD fallback remain in force; `golang-how-to`, `golang-testing`,
  `golang-error-handling`, `golang-security`, `golang-safety`,
  `golang-design-patterns`, `golang-structs-interfaces`, `golang-context`,
  and `golang-concurrency` govern this transport-only group. Group 8 files
  (`internal/cli/cli.go`, `internal/app/app.go`,
  `internal/app/change_capture_dispatch.go`, and
  `internal/app/etl_mode_dispatch.go`) are explicitly out of scope.
- **B27 red:** `TestDeclarativeTransport_PageBudgetIsNotEOF` must prove a
  deterministic three-page declaration reports omitted/default, one-page,
  two-page, and unlimited outcomes distinctly; a full-overwrite prefix must
  publish zero replacement tables; and an incremental continuation must emit
  page N+1 once rather than replaying an acknowledged prefix. A continuation
  is engine-owned opaque state, never a CLI/provider-cursor input.
- **B28 red:**
  `TestDeclarativeDestination_ClaimedKeyedWithoutIndependentProofIsRejected`
  and `TestDeclarativeDestination_IndependentProof` must refuse a declaration
  that merely claims keyed delivery before source I/O, then prove immutable
  executor/action-digest evidence plus one preview/workset-stable effective
  idempotency key permits exactly one provider mutation after a post-success
  retry. Foreign or changed evidence/action must fail admission.
- **B29 red:** `TestDeclarativeDestination_ReadBackUsesInternalReceiptLocator`
  must put the confirmed record beyond a first unrelated page, retain a full
  private provider receipt locator while public acknowledgement output stays
  redacted, reject absent/foreign/changed locators before checkpoint, and
  reuse the exact locator for bounded eventual-consistency retries.
- **B30 red:** `TestDeclarativeTransportClone_PreservesLargeNumbers` must
  carry nested signed/unsigned values, `json.Number`, `json.RawMessage`, and
  a value above 2^53 through source, stage, destination, read-back, and
  checkpoint without mutating the caller record; unsupported mutable values
  fail closed.
- **B31 red:** `TestDeclarativeReadBack_NumericSemanticEquality` must prove
  arbitrary-precision `int64`/`json.Number` equality above 2^53, nested
  identity equality, the explicit exact policy for `42` and `42.0`, and
  rejection of unequal numeric, string, and boolean values.
- **B32 red:** `TestTransport_ReadBackGetsIndependentUnitDeadline` must prove
  ordinary, full-overwrite, serial Arrow, and pipeline Arrow execution use
  separately bounded apply and read-back phases: 40 ms apply plus 20 ms
  read-back fits independent 50 ms bounds, while a >50 ms apply fails and a
  parent cancellation reaches both phases.

### Group 7 observed red evidence (2026-08-21)

- **B27:** `TestDeclarativeTransport_PageBudgetIsNotEOF` initially treated a
  capped `Read` prefix as ordinary EOF, so the durable path had neither a
  typed continuation nor an exhausted/non-exhausted distinction. The
  full-overwrite compatibility test then demonstrated that a capped source
  could reach its replacement boundary without a typed terminal stop.
- **B28:** the declared `delivery.idempotency=keyed` claim alone admitted a
  destination plan; the red tests showed that no independently sealed
  executor/action digest/effective header bound existed before provider I/O.
- **B29:** `TestDeclarativeDestination_ReadBackUsesInternalReceiptLocator`
  first observed a `response_index: 1` locator pass contract construction and
  broad/receipt-free read-back remained possible. The engine optional-query
  red additionally rejected a declaration-owned omitted query before request
  assembly, rather than preserving the bounded point-query route.
- **B30:** `TestDeclarativeTransportClone_PreservesLargeNumbers` first
  received `9007199254740993` as `float64` after the App JSON round trip,
  proving loss before destination/read-back/checkpoint boundaries.
- **B31:** `TestDeclarativeReadBack_NumericSemanticEquality` first reported
  an expected identity missing when an exact `int64` value was returned as
  `json.Number`; float-normalized hashing/comparison was not a valid
  compatibility bridge.
- **B32:** the test first failed to compile because `Result` had no separate
  confirmation metric. After adding that required observable, all four happy
  routes failed as intended: a 40 ms effect followed by a 20 ms read-back on
  one 50 ms context returned `context deadline exceeded`. The strict
  over-deadline apply and parent-cancellation edge cases remained red/green
  fences throughout.

### Group 7 focused green evidence (complete)

- **B27:** engine-owned `ReadContinuation` now records only definition-bound
  pagination state, `SourceReadOutcome` preserves exhaustion separately, and
  full overwrite converts the legacy typed budget stop into zero publication,
  zero read-back, zero checkpoint, and one private shadow abort. Incremental
  runs checkpoint exactly the acknowledged page plus opaque continuation.
- **B28:** `DestinationIdempotencyProof` binds the exact transport executor,
  action definition SHA-256, and declaration-owned effective header. A
  production plan is refused without all three. The existing partial-result
  assertion intentionally now expects exactly six sends: one successful first
  record plus five retry-safe attempts for the terminal second record under
  its declared stable `Idempotency-Key`; it still asserts every provider result
  is persisted and does not weaken to a lower-bound count.
- **B29:** public acknowledgement output is sanitized while bounded private
  receipt bytes stay process-local. A locator has an exact declared query
  binding, 4 KiB-or-less scalar cap, and 10-page-or-less declaration limit;
  the generic one-action path refuses compound response index ownership before
  a write. The regression now proves two exact two-page point-query attempts:
  an unrelated first page, then absent/confirmed target, always with the same
  private locator and no locator in public output. Missing and foreign private
  receipts fail before provider I/O.
- **B30/B31:** the App calls the transport lossless clone rather than JSON
  marshal/unmarshal; arbitrary precision numeric identity is canonicalized as
  an exact rational value. `42`, `42.0`, and `1e3` are equal numerically, but
  strings/booleans and unequal values fail closed.
- **B32:** `transportUnitContext` creates a fresh parent-derived 50 ms phase
  for every apply/publication and every confirmation. The focused suite passed
  ordinary transport, ordinary full overwrite, serial Arrow, pipelined Arrow,
  strict >50 ms effect refusal, and parent cancellation. `ReadBackElapsed`
  is persisted as `read_back_elapsed_ns`, distinct from effect latency.
- **Combined focused command:**
  `go test -timeout 20m ./internal/connectors/engine ./internal/synctransport ./internal/app -run 'Test(DeclarativeTransport_PageBudgetIsNotEOF|OpenComposedGitHubCommitsHonorsTransportMaxPages|OrchestratorFullOverwriteBudgetStopNeverPublishes|DeclarativeDestination_ClaimedKeyedWithoutIndependentProofIsRejected|DeclarativeDestination_IndependentProof|DeclarativeDestination_ReadBackUsesInternalReceiptLocator|DeclarativeTypedDestination_ReadBackProviderStateBeforeCheckpoint|DeclarativeTypedDestinationPersistsPartialProviderResultsOnFailedApply|DeclarativeTransportClone_PreservesLargeNumbers|DeclarativeReadBack_NumericSemanticEquality|Transport_ReadBackGetsIndependentUnitDeadline|ReadRequestQueryTemplateMissingFailsBeforeRequest|ReadRequestOptionalQueryTemplateMissingOmitsBeforeRequest)$' -count=1 -v`
  passed before the final two-page B29 strengthening; it is rerun as part of
  the Group 7 closure gates below.

### Group 7 package-gate failure set (2026-08-21)

The sole completion-tracked `go test -timeout 20m ./internal/app -count=1`
gate reached terminal exit `1` after `243.610s`. This complete observed set is
frozen before any later full App rerun:

| Exact failing test | Initial disposition to prove with focused evidence |
| --- | --- |
| `TestFoundationRollupPreservesMultiActionReverseETLComposition/first_declared_action` | B29 made a private provider receipt mandatory, but the pre-existing synthetic multi-action write fixture returns no declared `receipt_id`. Update only its closed provider fixture/read-back response so the already-declared locator is exercised. |
| `TestFoundationRollupPreservesMultiActionReverseETLComposition/second_declared_action` | Same B29 fixture gap: the write body is not a JSON receipt. Preserve multi-action composition and give the closed synthetic operation its declared receipt. |
| `TestFoundationRollupPreservesMultiActionReverseETLComposition/one_action_in_another_connector` | Same B29 fixture gap across a second declaration-owned connector; no connector-name branch is permitted. |
| `TestFoundationRollupPreservesMultiActionReverseETLComposition/refuses_missing_action_in_multi-action_connector` | Its exact expected I/O count was derived from the three successful source/write/read-back paths above. Recalculate and retain the exact count only after their closed receipts are restored. |
| `TestFoundationRollupPreservesMultiActionReverseETLComposition/refuses_other_connector_action` | Same exact-count dependency; no assertion relaxation. |
| `TestFoundationRollupPreservesMultiActionReverseETLComposition/refuses_unlisted_action` | Same exact-count dependency; no assertion relaxation. |
| `TestFoundationRollupPreservesMultiActionReverseETLComposition` final persisted-missing-action count | Same exact-count dependency; no assertion relaxation. |
| `TestPersistedConnectionSelectsDeclarativeTypedDestinationAction` (the matching three successful and four exact-count subcases) | This wrapper shares the same multi-action fixture and is expected to close with the focused composition test; it must be rerun explicitly. |
| `TestDeclarativeTransportSourceEmitsWholeProviderPageInBoundedBatches` | B27 now surfaces a capped source as `SourceBudgetStoppedError`. Determine whether its fixture is an intentionally complete scan (then make its test request unlimited) or whether production outcome handling is incorrectly translating an exhausted page; retain full-page/batch assertions. |
| `TestParkedFullAppendRateResumePreservesCheckpointAndBatchSize` | Test in isolation after the earlier duplicate App package processes are gone; current terminal error was an active-work ownership conflict. |
| `TestParkedFullAppendRateResumeRearmsLatestCheckpoint` | Same isolated ownership/rearm determination. |
| `TestParkedFullAppendRateResumeReconcilesInterruptedRearmAttempt` | Same isolated ownership/rearm determination. |
| `TestRunETLTransportRejectsAcknowledgedCheckpointWithIncompatibleResume` (three subcases) | Same isolated state-fence determination. |
| `TestRunETLTransportFullOverwriteSourceFailureAbortsWithoutCheckpoint` | Same isolated generation determination. |
| `TestRunETLTransportFullOverwriteCancellationBeforePublishAbortsWithoutCheckpoint` | Same isolated generation determination. |
| `TestRunETLTransportDistinguishesMissingAndPresentStreamState` | Same isolated state-fence determination. |
| `TestRunETLTransportCommitsAcknowledgedPageBeforeCancellation` | Same isolated state-fence determination. |
| `TestRunETLTransportRetainsInterimCheckpointWhenFinalStateSaveFails` | Same isolated state-fence determination. |
| `TestRunETLTransportFullOverwriteCompletionRebasesUnrelatedStateAfterFinalCheckpoint` | Same isolated state-fence determination. |
| `TestRunETLTransportFullOverwriteCompletionMissingRunIsTypedConflictAfterFinalCheckpoint` | Same isolated state-fence determination. |

No later full App-package rerun closes this table. Every row must have focused
red/green evidence or a reproducible environment-only determination first.

### Group 7 package-failure focused disposition (2026-08-21)

All 20 rows above are now closed by the following focused green evidence before
the one final Group-7 App package rerun:

- **B29 multi-action rows (rows 1–8):**
  `go test -timeout 20m ./internal/app -run '^(TestFoundationRollupPreservesMultiActionReverseETLComposition|TestPersistedConnectionSelectsDeclarativeTypedDestinationAction|TestDeclarativeTransportSourceEmitsWholeProviderPageInBoundedBatches)$' -count=1 -v`
  passed. The closed synthetic response now supplies the already-declared
  `receipt_id`; all three selectable actions, both cross-connector rejection
  cases, and their exact I/O-dependent counts passed unchanged. The unrelated
  `TestDeclarativeTransportSourceEmitsWholeProviderPageInBoundedBatches` now
  explicitly asks for its intended unlimited complete scan and proves the
  two-provider-page exhaustion read, rather than converting a B27 budget stop
  into EOF.
- **Group-6 fence integration rows (rows 9–20):**
  `go test -timeout 20m ./internal/app -run '^(TestParkedFullAppendRateResumePreservesCheckpointAndBatchSize|TestParkedFullAppendRateResumeRearmsLatestCheckpoint|TestParkedFullAppendRateResumeReconcilesInterruptedRearmAttempt|TestRunETLTransportRejectsAcknowledgedCheckpointWithIncompatibleResume|TestRunETLTransportFullOverwriteSourceFailureAbortsWithoutCheckpoint|TestRunETLTransportFullOverwriteCancellationBeforePublishAbortsWithoutCheckpoint|TestRunETLTransportDistinguishesMissingAndPresentStreamState|TestRunETLTransportCommitsAcknowledgedPageBeforeCancellation|TestRunETLTransportRetainsInterimCheckpointWhenFinalStateSaveFails|TestRunETLTransportFullOverwriteCompletionRebasesUnrelatedStateAfterFinalCheckpoint|TestRunETLTransportFullOverwriteCompletionMissingRunIsTypedConflictAfterFinalCheckpoint)$' -count=1 -v`
  passed. The correction keeps parent cancellation on provider phases but gives
  an already acknowledged receipt one bounded local checkpoint attempt. An
  uncommitted exact lease is atomically released/restored; a rate-limit handoff
  releases only its exact parked owner with its committed checkpoint. A lost
  fence remains both the typed stream-state conflict and its original fence
  cause. The final-save and full-overwrite pause fixtures retain their old
  assertions but count the new exact claim/source/begin/destination/publish
  fence writes so they still inject at the intended post-checkpoint boundary.
- **Full Group-7 red/green cohort:**
  `go test -timeout 20m ./internal/connectors/engine ./internal/synctransport ./internal/app -run 'Test(DeclarativeTransport_PageBudgetIsNotEOF|OpenComposedGitHubCommitsHonorsTransportMaxPages|OrchestratorFullOverwriteBudgetStopNeverPublishes|DeclarativeDestination_ClaimedKeyedWithoutIndependentProofIsRejected|DeclarativeDestination_IndependentProof|DeclarativeDestination_ReadBackUsesInternalReceiptLocator|DeclarativeTypedDestination_ReadBackProviderStateBeforeCheckpoint|DeclarativeTypedDestinationPersistsPartialProviderResultsOnFailedApply|DeclarativeTransportClone_PreservesLargeNumbers|DeclarativeReadBack_NumericSemanticEquality|Transport_ReadBackGetsIndependentUnitDeadline|ReadRequestQueryTemplateMissingFailsBeforeRequest|ReadRequestOptionalQueryTemplateMissingOmitsBeforeRequest)$' -count=1 -v`
  passed after the final two-page private-locator strengthening.

### Group 7 final App-gate supplemental red set (2026-08-21)

The one permitted post-disposition `go test -timeout 20m ./internal/app -count=1`
run reached terminal exit `1` after `212.611s`. Its complete additional failure
set is frozen here before another package rerun:

| Exact failing test | Reproduced cause to correct with focused red/green evidence |
| --- | --- |
| `TestProductionParkingCompositionSurvivesProcessKill` | Its process fixture inherits `TestParkRateLimitedRunPersistsOnlyDeclarativeTypedDestinationPlanID`; the new parking lease release treated a direct, no-live-lease persistence call as a foreign owner instead of preserving its existing closed parking behavior. |
| `TestParkRateLimitedRunPersistsOnlyDeclarativeTypedDestinationPlanID` | A parked run with an acknowledged checkpoint but no active lease is a valid direct parking persistence seam; it must retain that checkpoint/plan proof. Only a nonempty foreign active owner is a conflict; the current exact owner is the one that must be released. |
| `TestRunETLTransportTreatsIndeterminateCheckpointPersistenceAsFailure` | An indeterminate checkpoint store outcome may already have durably committed the acknowledgement. The new uncommitted-lease rollback deleted that state even though a subsequent retry could duplicate an acknowledged provider effect. |

No further full App-package rerun is permitted until all three tests have
focused green dispositions.

### Group 7 supplemental focused green disposition (2026-08-21)

`go test -timeout 20m ./internal/app -run '^(TestProductionParkingCompositionSurvivesProcessKill|TestParkRateLimitedRunPersistsOnlyDeclarativeTypedDestinationPlanID|TestRunETLTransportTreatsIndeterminateCheckpointPersistenceAsFailure)$' -count=1 -v`
passed. A direct valid parking persistence has no lease to release, while an
exact live owner releases only itself and a foreign nonempty owner still fails
closed. A checkpoint error whose store outcome may have committed now retains
the acknowledged durable state; only a definite no-commit unacknowledged path
is eligible for lease rollback. The process-kill path proves the same behavior
through three isolated processes.

### Group 7 closure gates (2026-08-21)

- The definitive package gate, run only after dispositioning every failure in
  both frozen App failure sets above, passed:
  `go test -timeout 20m ./internal/app -count=1` (`244.005s`).
- The non-App package gate passed:
  `go test -timeout 20m ./internal/connectors/engine ./internal/synctransport -count=1`
  (`9.692s` and `1.079s`).
- The complete named Group-7/co-composed regression cohort, including B27
  continuation, B28 plan proof, B29 private two-page receipt read-back,
  B30/B31 lossless comparison, B32 ordinary/full-overwrite/serial-Arrow/
  pipelined-Arrow deadlines, and the Group-6 fence compatibility cases,
  passed under `go test -race -timeout 20m` over engine, synctransport, and
  App (`2.145s`, `2.990s`, and `603.875s` respectively).
- `go vet ./internal/connectors/engine ./internal/synctransport ./internal/app`,
  `go build ./cmd/pm`, `go run ./cmd/connectorgen validate` (552 connectors,
  zero findings), `go run ./cmd/connectorgen surface-sync --check` (zero
  drift), and `git diff --check` passed. The final diff inventory contains
  only the Group-7 engine/transport/App/sync-contract files plus this ledger;
  it leaves every Group-8 file untouched.

## Group 1 preserved test provenance

The recovery worktree contains uncommitted Group 1 production-shaped tests and generated fixtures. They are transferred as the reviewed recovery set without modifying their bytes, then run as the red/green evidence recorded above. The supplemental Node transport-order test is explicitly rerun before Group 1 commit because its ordering defect is the B01 manifestation.

## Group 1 full-CLI failure-set disposition (2026-08-21)

The first full `go test -timeout 20m ./internal/cli -count=1` run reached a
terminal **failure** after `792.193s`. The terminal capture was truncated by
the transport, so the four failures extracted from that output are frozen
below but are not represented as the entire Group 1 full-suite set. The later
exact-session terminal capture expands the frozen set immediately after this
table. A later green full run is evidence only after every row has its focused
red/green disposition. No row is treated as an environment-only exception.

| Exact failing test | Reproduced cause | Red evidence | Correction and focused green evidence |
| --- | --- | --- | --- |
| `TestSkillsGenerateMatchesTrackedSkills` | The source-derived GitHub/Google Ads and related connector command surfaces had changed while tracked `docs/skills/**` and matching connector manuals were stale. | The 792.193s full CLI output named this exact test; the tracked generated files differed from `pm skills generate`. | Regenerated only through `go run ./cmd/pm skills generate --dir docs/skills` and `go run ./cmd/pm docs generate --dir docs/cli`; the exact match test is included in the current session-tracked full rerun and must pass before this group closes. |
| `TestGoldenTranscripts/dynamic_connector_bare_json` | The generated GitHub command/help ordering changed, leaving the exact bare dynamic-connector JSON transcript stale. | The 792.193s full CLI output named this exact subtest. | Regenerated only with `POLYMETRICS_UPDATE_GOLDEN_TRANSCRIPTS=1 go test -timeout 20m ./internal/cli -run '^TestGoldenTranscripts$' -count=1`; the non-update exact transcript test is included in the current session-tracked full rerun and must pass before closure. |
| `TestReverseETLToGitHubCreatesPullRequestAfterApproval` | Projection replaced the declared GitHub compound-hook follow-up fields `labels` and `reviewers` with only the primary provider body fields, so planning rejected `record.labels` before any provider I/O. | The 792.193s full CLI output failed at `//labels: additional property not allowed`; `TestSourceProjectionPreservesDeclaredHookFollowupFieldsOutsideProviderBody` then failed before the projector repair. | Added generic declaration-owned `hook_fields`, requiring a closed schema, a registered hook, no duplicates, and no primary path/body overlap. `sourceProjectionRetainDeclaredHookFields` preserves only those named bounded fields outside `body_fields`; `TestSourceProjectionPreservesDeclaredHookFollowupFieldsOutsideProviderBody`, `TestValidateWriteHookFieldsRequireClosedSupplementalDeclarations`, and the exact reverse-ETL test now pass. |
| `TestYouTubeAnalyticsReportsDownloadRunsThroughBoundedBinaryExecutor` | The fixed binary operation exposed `{path}` but did not declare the provider resource identity; source sync reduced the command to no flags, so `--resource-name` was unknown. | The 792.193s full CLI output named this exact test and its unknown `--resource-name` failure; `TestDeriveCommandParameterFlagsUsesDeclaredCLIAliasForSafePathPlaceholder` then failed with generated `--path`. | Added a validated path-only declaration alias `cli_name`, mapping `--resource-name` to the closed `path.path` parameter with an explicit 4,096-byte cap. The alias unit test and the installed binary executor test now pass; the latter also proves a 4,097-byte value fails before another provider request. |

## Group 1 later exact-session full-CLI failure-set expansion (2026-08-21)

The non-overlapping session-tracked rerun returned terminal exit `1` after
`1201.193s`; it is red evidence, not a Group 1 closure. Its complete observed
failure set expands the frozen ledger as follows. No further full CLI run may
start until these rows, together with the preceding four, have focused
dispositions.

| Exact failing test | Reproduced cause | Red evidence | Correction and focused green evidence |
| --- | --- | --- | --- |
| `TestGitHubCommandSurfacePlansReverseETLCommand` | The immutable source declares the issue `title` through a scalar `oneOf` (`string` or `integer`). Projection retained that provider schema but generated a named `json` flag requiring JSON syntax, so the ordinary declared `--title Ship connector command plans` route failed before plan creation. | Exact-session full CLI output: `error: invalid JSON for --title: invalid character 'S' looking for beginning of value`; the focused installed test reproduced the same red before repair. | Added `allow_bare_string` only for an already-bounded, named reverse-ETL `json` flag whose concrete closed record schema has a declared string arm. `TestSourceProjectionStringUnionKeepsTextCLIAndProviderArms`, `TestValidateStructuredJSONRecordStringArmRequiresNamedDeclaredStringUnion`, and `TestRecordOverridesBareStringUnionRemainsBoundedAndRejectsMalformedContainers` pass; they retain all source union arms, reject malformed object/array JSON and 8-byte overflow, and add no raw body authority. Regenerated GitHub projection reports `cli=14`; exact installed `TestGitHubCommandSurfacePlansReverseETLCommand` passes. |
| `TestBahmniBareCommandGroupInvalidMultiPartPathIsNotHelp` | The full process exhausted its 20-minute suite budget while this test was active. Its source only creates a temporary project and asserts the invalid nested path yields usage output; it performs no provider/network action. The timeout stack showed unrelated generated-surface loading, not a Bahmni assertion failure. | Exact-session terminal output: `panic: test timed out after 20m0s`, active test `TestBahmniBareCommandGroupInvalidMultiPartPathIsNotHelp`. | Dispositioned as the prior full-suite environment/scheduling timeout, not a product failure: isolated exact runs passed three times (`2.571s`, `2.510s`, `2.490s` test elapsed; outer wall `6.15s` and `5.94s` for the repeated pair). The test still asserts the exact unknown-command and no-help/no-credential-resolution boundaries. No waiver is applied to the later combined full-suite gate. |

### Focused closure reruns

- `TestSkillsGenerateMatchesTrackedSkills` passed from the regenerated surface in `150.025s` (tracked terminal exit `0`).
- The remaining original failure tests passed serially with terminal exit `0`: `TestGoldenTranscripts/dynamic_connector_bare_json` (`2.467s`), `TestReverseETLToGitHubCreatesPullRequestAfterApproval` (`3.246s`), and `TestYouTubeAnalyticsReportsDownloadRunsThroughBoundedBinaryExecutor` (`4.168s`). The serial command reached each test only after its predecessor passed.
- `TestGitHubCommandSurfacePlansReverseETLCommand` passed after the named string-arm repair (`3.108s`); no full `internal/cli` rerun has been used to close Group 1.

## Group 1 atomic closure gates

All below commands ran from this exact pre-commit tree after the final help
renderer and generated artifact refresh. Group 1 intentionally uses focused
engine, generator, and installed-command evidence; the heavyweight combined
App/CLI/all-connectors suite is reserved for the final exact SHA.

- Red/green: `TestGitHubIssueCreateHelpDescribesDeclaredBareStringArm` failed while help printed `--title (json)` and passed after the renderer published `--title (json or string)`. The paired normal plan test also passed.
- Package gates: `go test -timeout 20m ./cmd/connectorgen -count=1` (`105.213s`), `./internal/connectors/engine` (`6.988s`), and `./internal/connectors/commandrunner` (`21.707s`) passed.
- Required idempotence: `TestSourceProjectionGapCreatesCommandFromExistingClosedActionVariant` passed twice with the exact `stats.CLI == 1` assertion intact; `TestSyncBundleDerivesRequiredPathFlagFromRESTParameter` and `TestSyncBundleProviderParameterProjectionIsIdempotent` passed. The following immutable source check then verified `1525` operations with no drift.
- Generated boundaries: source-import `--cache-dir .source-import-qualification-cache.OHUVcH --check`, `surface-sync --check`, GitHub certification candidates `--check`, sweep `--check` (`1616` rows; `1612` CLI commands), and `connectorgen validate` (`552` connectors, `0` findings) passed.
- Surface parity: `make github-parity-artifacts-check` passed all 16 Node tests and both generator checks; regenerated `docs/skills`, `docs/cli`, connector manuals, and `TestGoldenTranscripts` update completed. `TestSkillsGenerateMatchesTrackedSkills` then passed (`169.312s`), and the final focused installed CLI failure-set suite passed (`9.119s`).
- Build hygiene: `make docs-check`, `go build ./cmd/pm`, `go vet ./...`, and `git diff --check` passed.

## Group 1 source-lock cache qualification

- Red: `TestSourceImportCommandUsesExplicitVerifiedCacheRoot` initially failed with `unknown flag "--cache-dir"`; this demonstrated that an empty, qualification-owned cache could not be selected and that the earlier XDG-based observations did not establish a cold cache.
- Green: the command now accepts only an existing non-symlink `--cache-dir`, wraps the normal source fetcher at that root, and never changes the connector-owned URL, digest, byte count, or request authority. The focused cache suite passed: `TestSourceImportArtifactCacheColdSlowFetchWritesOnlyVerifiedBytes`, `TestSourceImportArtifactCacheHitVerifiesWithoutNetwork`, `TestSourceImportArtifactCacheRejectsCorruptionAndOnlyRecoversFromVerifiedFetch`, `TestSourceImportCommandUsesExplicitVerifiedCacheRoot`, and `TestSourceImportCommandContractAndMigrationDocumentation`.
- Live locked GitHub REST source (2026-08-21): a newly empty explicit root at `/Users/karthiksivadas/.treehouse/cli-83d592/57/cli/.source-import-qualification-cache.OHUVcH` completed `go run ./cmd/connectorgen source-import github --cache-dir <root> --check` in `real 9.01` seconds. It created exactly `/Users/karthiksivadas/.treehouse/cli-83d592/57/cli/.source-import-qualification-cache.OHUVcH/80850db290cde4eb487e0efb587cf27f305e77b6bef96933ed8a09b5169d5b1d.artifact`, with `12,920,264` bytes and SHA-256 `80850db290cde4eb487e0efb587cf27f305e77b6bef96933ed8a09b5169d5b1d`.
- Warm live check: the same exact root and immutable file completed the same command in `real 5.50` seconds; byte count and SHA-256 were rechecked and the artifact modification time remained `2026-08-21T10:55:15Z`. This temporary qualification root is inventoried here and must be removed recoverably before the Group 1 commit.

Correction: the cache root and recovery sentinel were deliberately retained outside
the index until Group 1 remote SHA `d3bf5da0e6a4575628dd76dd94a7522220f9d3df`
was verified, then moved recoverably to Trash. Neither was committed.

## Group 2 red-contract plan (2026-08-21)

- **GSD/manual fallback:** `scripts/gsd doctor`, the five required command
  prompts, and `go run ./cmd/agentcontractgen check` were resolved in this
  single-owner Herdr worktree. The frozen POSTFIX review fixes proceed inline;
  no specialist owns files or commits in this lane.
- **Skills:** `golang-how-to`, `golang-cli`, `golang-testing`,
  `golang-error-handling`, `golang-security`, `golang-safety`,
  `golang-design-patterns`, `golang-structs-interfaces`, `golang-lint`, and
  `golang-documentation` were loaded. CLI help/manual/website parity is part
  of this group because the generated GraphQL flags and skills change.
- **B04/B05 red:** a fixed source-backed connection must reject neither/both
  `first`/`last` and cursor-without-direction before I/O; a `last` page must
  use `hasPreviousPage`/`startCursor`, reject malformed backward state, and
  never substitute an unrelated forward cursor.
- **B06/B07 red:** source-derived query scalar `invitationToken` must be
  env-only; classified `tempCloneToken` and `verificationToken` must not be
  selected; bounded `createIssue`/`addComment` selections retain source-owned
  IDs/URLs plus `clientMutationId`, while ordinary token-count and occurrence
  IDs remain unchanged. No caller document, selection, or raw-body channel is
  added.
- **B08 red:** GraphQL `Int` exact signed-32-bit boundaries must pass and
  adjacent values must fail in root, nested object, and list variables before
  any HTTP request.
- **W02 red:** every safety-relevant flag property (`env_only`, byte/item
  bounds, repeatability, allow-empty, format, required, values, maps-to, and
  numeric minimum) must survive website projection and generated skills/help
  must state env-only and active limits.

### Group 2 observed red evidence

- `go test -timeout 20m ./internal/connectors/engine -run 'Test(GraphQLOperationVariablesRequiresExactlyOnePaginationDirection|OperationDirectReadBackwardGraphQLPaginationUsesPreviousPageInfo|GraphQLIntUsesSigned32BitDomain)$' -count=1` failed as intended: neither direction and cursor-without-direction reached variable validation, mixed direction reported the older non-exact message, backward page state completed from `hasNextPage`, and schema compilation rejected the undeclared `minimum` keyword.
- `node --test scripts/tests/gen-github-graphql-parity.test.mjs` failed as intended with the complete frozen B06/B07 set: invitation token stayed inline; `tempCloneToken`/`verificationToken` remained selected in the five source-owned fixed documents; and `createIssue`/`addComment` omitted nested provider identity.
- `node --test website/scripts/cli-surface.test.mjs` failed as intended: website projection dropped `env_only`, `max_bytes`, `min_items`, and `max_items`.
- `go test -timeout 20m ./internal/connectors -run '^TestRenderCommandSurfaceCommandRendersSafetyConstraints$' -count=1` failed as intended: generated guide text rendered only `--input (required)` and omitted env-only, non-empty, item, and byte constraints.

### Group 2 green evidence

- **B04/B05:** `TestGraphQLOperationVariablesRequiresExactlyOnePaginationDirection`,
  `TestGraphQLOperationVariablesRejectsMixedPaginationDirections`, and
  `TestOperationDirectReadBackwardGraphQLPaginationUsesPreviousPageInfo` pass
  through the real engine. The fixed connection accepts exactly one of
  `first`/`last`, rejects neither, both, and cursor-without-direction before
  transport, and derives backward continuation solely from
  `hasPreviousPage`/`startCursor`.
- **B06/B07:** `TestGeneratedGraphQLContractsClassifySecretInputsAndBoundedIdentitySelections`
  passes. Source-declared query invitation tokens are `env_only` and map only
  to their exact root variable under a source-declared env/redaction policy;
  no raw document/body or undeclared variable form was added. Classified
  result tokens are absent from fixed selections. Mutation selections are
  source-derived, cycle/depth bounded (`3`), field-budget bounded (`64`), and
  retain `clientMutationId` plus provider IDs/numbers/URLs for `createIssue`
  and `addComment`.
- **B08:** `TestGraphQLIntUsesSigned32BitDomain` passes root, nested object,
  and list values at `[-2147483648, 2147483647]`, and adjacent values reject
  before I/O. `TestValidateFlagMaximumIsOptIn` proves the generated CLI flag
  maximum is independently enforced. The schema compiler now keeps exact
  numeric bounds rather than float-rounded substitutions.
- **W02:** `TestWebsiteFlagProjectionPreservesEverySafetyProperty` and
  `TestRenderCommandSurfaceCommandRendersSafetyConstraints` pass. Snake/camel
  projections retain env-only, byte/item/numeric bounds, repeatability,
  allow-empty/bare-string, type/values/mapping/format/requiredness; generated
  guide command lists render active safety qualifiers without hiding the
  command surface.
- **Static closure:** the red `TestValidate_CLISurfaceEnvOnlyFlagRequiresDeclaredSecretGraphQLContract`
  now proves the validator allows only a source-schema-declared scalar
  `graphql_query` variable with `input_mode=env`, `transform=none`, and exact
  `redact_fields`, while still refusing an omitted policy. The established
  typed-secret GraphQL mutation form remains unchanged.
- **Generation:** `connectorgen validate` reports `552 connector(s)` and zero
  findings; `surface-sync --check` reports zero drift; GitHub parity artifacts,
  source-drift, combined-ledger, certification-matrix, certification-candidates,
  and certification-sweep checks passed. The sweep remains exactly `1616`
  rows / `1612` CLI commands. `pm docs generate --dir docs/cli`,
  `pm skills generate --dir docs/skills`, and `pm docs validate --connectors-dir
  docs/connectors` regenerated/validated the tracked manual and skill surfaces.

## Group 3 red-contract plan (2026-08-21)

- **B13/B14:** `TestPublicReceiptProjectionMasksOnlyExactConfiguredAndDeclaredScalars` will drive quote, backslash, `<`, `>`, `&`, non-ASCII, short values (`id`, `token`, `0`, and one character), and encoded configured values through raw and decoded REST receipts. It must prove canonical public JSON has no concrete/proven encoding, while keys (`occurrence_id`, `trained_tokens`), header names (`WWW-Authenticate`), repeatable ordinary values, and the internal receipt remain unchanged. `TestGraphQLErrorMetadataDoesNotKeywordRedactOrdinaryProviderWords` will preserve the ordinary provider message `Unknown token type` and leave exact-value masking to the public result boundary.
- **B17:** `TestCommandRunnerPreservesLegacyPostProviderResultWithoutReceipt` and `TestAshbyOperationDirectReadPreservesEngineResultOnEnvelopeFailure` will require result-plus-error propagation for legacy, operation, navigation, and Ashby logical-envelope failures.
- **B19:** `TestOperationStatusCheckPreservesPostResponseFailureResult` will make a declared bounded-header validation error after a received HEAD response retain its operation/status/path/result; receipt and CLI tests then require a bounded complete 204/404/error envelope without body decoding.
- **B24:** `TestOperationDirectReadPreservesCompleteSQSReceiptOnSuccessAndProviderError` will table-drive ordinary 200 XML and terminal 4xx XML with repeated headers, raw byte count, decoded success body, and a received result on failure. Malformed XML, cap+1, and body-read errors share that same received-response path.
- **GSD/manual fallback:** the exact same single-owner, no-specialist fallback recorded for Group 2 applies. The red package set is connector output, engine, commandrunner, native Amazon SQS/Ashby, and CLI; no generated surface changes are expected for this behavioral group.

### Group 3 observed red evidence

- TestPublicReceiptProjectionMasksOnlyExactConfiguredScalars initially failed
  because public sanitization rewrote the provider-owned occurrence_id key to
  occurrence_[masked]. TestGraphQLErrorMetadataDoesNotKeywordRedactOrdinaryProviderWords
  initially replaced Unknown token type merely because it contained token.
- TestCommandRunnerPreservesLegacyPostProviderResultWithoutReceipt,
  TestOperationStatusCheckPreservesPostResponseFailureResult,
  TestOperationDirectReadPreservesCompleteSQSReceiptOnSuccessAndProviderError,
  and TestAshbyOperationDirectReadPreservesEngineResultOnEnvelopeFailure
  each initially observed a zero result after a received provider response or
  logical provider-envelope failure.
- The review-added TestPublicReceiptProjectionPreservesRawJSONBytesWhenNoMaskApplies
  failed with canonicalized/reordered JSON although the configured credential
  did not occur. Its opaque-byte companion proves a short configured value
  does not erase provider bytes that merely contain that substring.
- The review-added TestRunOmitsResultEnvelopeBeforeProviderResponse initially
  failed for legacy direct-read, status, and binary transport errors: each
  returned an empty result envelope even though no provider response existed.

### Group 3 green evidence

- B13/B14: engine receipts retain immutable raw headers/body; the public
  projection masks only exact configured or declared scalar values, keeps
  provider map/header identities intact, preserves unmodified JSON/base64
  bytes byte-for-byte, masks a real longer opaque credential without making a
  fabricated partial binary, and keeps ordinary GraphQL wording intact.
- B17/B19/B24: runner results retain legacy/direct/binary/status evidence
  with a non-nil error, the CLI emits the matching bounded result envelope
  even without a receipt, status HEAD retains raw metadata without decoding a
  body, and SQS/Ashby preserve complete provider receipts on ordinary,
  terminal, malformed, and logical-envelope paths.
  The runner now omits all three result forms when an executor error occurs
  before provider response evidence, preserving the fail-before-I/O boundary.
- Exact focused package gates from this final Group-3 tree passed: connectors
  (0.615s), engine (10.355s), commandrunner (21.832s), native Amazon SQS
  (1.245s), native Ashby (1.050s), and the focused App direct-write receipt
  suite (3.666s). The focused CLI receipt/envelope suite, including the
  no-receipt result-envelope regression, passed in 1.156s.
- go build ./cmd/pm and git diff --check passed. This behavioral group changes
  no connector declaration or generated surface; generator checks remain
  reserved for their owning source/surface groups and the final exact-SHA
  combined gate.

## Group 4 red-contract plan (2026-08-21)

- **GSD/manual fallback:** this remains the single-owner Herdr lane described
  for Groups 2 and 3. The frozen review is the plan authority; tests are added
  and observed red before their owning production correction. No outside
  worker, component branch, or recovery worktree is modified.
- **B15:** a handled GitHub/Ashby write must either execute a prepared, sealed
  ordered request plan or fail before I/O. Mutating the caller record while
  approval blocks must neither alter the wire bytes nor produce an unpreviewed
  effect. A compound failure must retain every attempted provider receipt in
  order and count only completed effects.
- **B16:** REST, GraphQL, form, and multipart operation writes must dispatch
  bytes/config/headers sealed at preview. Mutation of nested request/config/
  secret/digest state or a replaced multipart file after preview must result
  in the original material or a pre-I/O refusal; no live remarshal/reread is
  admitted.
- **B18:** refused 302/307 and terminal 429/503 followed by cancellation must
  retain the latest typed provider response with bounded metadata while never
  contacting a redirect target.
- **B21:** REST and GraphQL command values preserve exact numeric lexemes
  (including `9007199254740993`, `0.10000000000000001`, negative, and exponent
  forms) and compare exact declared minima before I/O.
- **B23/B25:** opaque pagination continuations and native SQS tokens reject
  unknown/duplicate fields, controls, malformed encoding, and capped-size
  expansion before authentication/request dispatch; a native SQS 302/307 must
  not contact the target or forward a session token.
- **W03/W04:** runtime-owned idempotency headers fail declaration validation
  rather than silently disappearing; minimum-witness generation produces a
  valid bounded string for a supported `minLength` schema.

### Group 4 observed red evidence

- **B15:** the frozen review established that `executeApprovedWrite` let
  `WriteHook.ExecuteWrite` choose physical requests after the preview, so the
  request set itself was absent from the approval digest even after receipt
  retention was repaired. Red-first
  `TestPreparedWriteHookSealsEveryPhysicalRequestAndRetainsTerminalReceipts`
  initially failed to build with the required plan fields/types absent
  (`PreparedRequest.Action`, `ResponseBinding`, and
  `PreparedWriteHookPlan`). The executable red requires two declaration-owned
  steps in preview order, a bounded first-receipt `id` binding for the second
  path, caller mutation after planning to remain off both wire bodies, and
  ordered 201/400 receipts to remain in the terminal result.
- **B16:** `TestPrepareOperationDirectWriteSealsNestedVariablesAndRuntimeMaps`
  initially failed with `prepared variables` carrying `mutated-name`; after
  sealing nested values it then failed with `prepared runtime config =
  map[tenant:mutated-tenant] / secret:mutated-secret`. Both failures occurred
  before I/O. The prepared JSON/form dispatch now consumes the digest-bound
  bytes, while multipart retains its immediate approved-digest revalidation.
- **B18:** `TestRequesterRetryTransportFailureRetainsEarlierProviderResponse`
  initially failed because `errors.As` found no `*HTTPError` after the first
  503 was followed by a transport failure. It now requires both the retained
  503 response/headers/body and the terminal transport cause. The companion
  stream redirect/backoff-cancellation tests cover the open-body branch.
- **B21:** `TestCoerceFlagValuePreservesExactNumericLexemes` observed
  `9007199254740993` as an `int`, `0.10000000000000001` as `float64(0.1)`, and
  `-1.25e-3` as `float64(-0.00125)`. Exact `json.Number` transport plus rational
  bound comparison replaces both conversion paths. The prior focused sweep
  exposed four stale integer/number expectations; they now assert the exact
  lexemes rather than reintroducing float coercion.
- **B23/B25:** the cursor red admitted `admin=true` on a same-origin link URL;
  the SQS red sent a CR/LF cursor to its endpoint. The redirect red delivered
  the session token to the redirected target. The added cursor/SQS tests prove
  pre-I/O refusal, no target contact, a retained redirect receipt, and the
  original `redirect refused` error identity together.
- **W03/W04:** `Idempotency-Key` and `X-Idempotency-Key` declaration validation
  initially returned nil; the minLength witness red returned `cannot prove a
  schema-valid string witness` for a closed `minLength:3` body. Both have
  focused green tests in engine.

### Group 4 focused green evidence to date

- `go test -timeout 20m ./internal/connectors/engine -run
  'Test(PrepareOperationDirectWriteSealsNestedVariablesAndRuntimeMaps|PreparedWriteHookSealsEveryPhysicalRequestAndRetainsTerminalReceipts|LegacyWriteHookClaimIsRefusedBeforeUnpreviewedTransport|OperationHeaderDeclarationsRejectRuntimeOwnedIdempotencyNames|StructuredRESTBodyMinimumWitnessHonorsMinLength|DirectReadCursorURLAdmitsOnlyBoundedDeclaredContinuationQuery)$' -count=1`
  passed.
- **B15 closure:** `PreparedWriteHook` now admits only named existing write
  declarations plus one bounded scalar JSON response field mapped to one
  earlier-step declared path field. The engine validates every selected action
  record/body, caps plans at eight physical requests per source record, seals
  the flattened action/binding/body/header set into the preview digest, and
  executes the same private ordered plan. A legacy `WriteHook` that claims an
  action fails before I/O. `TestPreparedWriteHookSealsEveryPhysicalRequestAndRetainsTerminalReceipts`
  passed against a real two-request server (create `201`, bound update `400`),
  retaining both complete provider receipts. `TestPreparedWritePlanEnumeratesEveryGitHubCompoundRequest`
  fixes all eight GitHub compound action variants and their exact declaration
  identities; `TestGitHubPreparedPlanExecutesOnlyPreviewedStepsAndRetainsTerminalReceipt`
  drives the real GitHub bundle through `engine.Write`, stops before an
  unplanned reviewers request after metadata `400`, and retains the create and
  terminal metadata receipts. Ashby's one-step native envelope route now uses
  the same plan/receipt boundary and its package suite passes.
- `go test -timeout 20m ./internal/connectors/connsdk -run
  'Test(RequesterRetryTransportFailureRetainsEarlierProviderResponse|RequesterMutationRetryCancellationRetainsLastResponse|DoStreamRetainsLastProviderEvidenceAcrossRedirectAndCancelledRetry|DoStreamDisableRetriesRejectsMutationRedirect)$' -count=1`
  passed.
- `go test -timeout 20m ./internal/connectors/commandrunner -run
  'Test(CoerceFlagValueNumber|CoerceFlagValuePreservesExactNumericLexemes|ValidateFlagNumericBoundsUseExactDeclarationLexemes|BuildWriteCommandPlansReopenAndPRSharedCommands|RecordOverridesBuildsExplicitNestedScalarFields|BuildOperationDirectWriteCommandUsesTypedInputsAndPlanLifecycle)$' -count=1`
  passed.
- `go test -timeout 20m ./internal/connectors/native/amazon-sqs -run
  'Test(ApprovedDestructiveWriteRefusesRedirectToUnapprovedTarget|OperationDirectReadRefusesSQSRedirectWithoutForwardingSessionToken|OperationDirectReadRefusesUnsafeSQSContinuationBeforeSigning)$' -count=1`
  passed.

### Group 4 atomic closure gates (2026-08-21)

- Focused behavioral gates passed from this exact uncommitted tree: engine
  (`0.993s`), GitHub prepared-plan (`1.052s`), Ashby native write (`1.047s`),
  connsdk retry/redirect (`0.335s`), commandrunner exact-number/write-plan
  (`1.035s`), SQS redirect/cursor (`0.935s`), and CLI structured-body/help
  (`1.160s`).
- Full affected package gates passed: `internal/connectors` (`0.618s`),
  `engine` (`8.931s`), `connsdk` (`3.463s`), `commandrunner` (`21.973s`),
  `hooks/github` (`5.031s`), `native/ashby` (`1.047s`), and
  `native/amazon-sqs` (`1.287s`).
- Transport/handler race gates passed: engine sealed-plan/cancellation
  (`2.132s`) and connsdk retry/stream boundary (`1.341s`) with `-race`.
- `go vet` over engine, connsdk, commandrunner, GitHub, Ashby, SQS, and CLI;
  `go build ./cmd/pm`; `connectorgen validate` (552 connectors, zero
  findings); `connectorgen surface-sync --check` (zero drift); and
  `git diff --check` all passed. No generated artifact changed in this group.

The B15 planner boundary and the frozen Group-4 package/generator gates are
closed and committed in remote-verified Group-4 SHA
`b0eb22feb7f413d15f747b3f78d62c6c46e314b9`; no recovery, credential, or
generated artifact was included.

## Group 5 red-contract plan (2026-08-21)

- **B22:** a binary download with `allow_overwrite=false` must leave no final
  name while its owned hidden temp is staging; a foreign final inserted before
  publication must survive byte- and inode-identically, and all owned temp
  entries must be removed after the link conflict. `TestBinaryDownloadNoOverwritePublicationIsCrashAndRaceSafe`
  is production-shaped: its reader blocks after writing staged bytes, letting
  the test observe the pre-publish directory and install the competing file.
- **W05:** the multipart escaping-symlink refusal must wait for handler
  completion before observing whether secret bytes reached the server. The
  handler has a request-completion channel and the test disables retries so the
  one observation has a single owner.

### Group 5 observed red evidence

- **B22:** before the state-machine correction,
  `TestBinaryDownloadNoOverwritePublicationIsCrashAndRaceSafe` failed at
  `final name exists before publication: <nil>`. The old `O_EXCL` reservation
  exposed a zero-byte final while the reader was blocked, so a crash could
  leave it and a later rename/cleanup could overwrite/delete a foreign file.
- **W05:** `go test -race -timeout 20m ./internal/connectors/connsdk -run
  '^TestRequesterDoMultipartRefusesEscapingSymlinkSwappedAfterValidation$'
  -count=1` reported a concrete data race between the test's read of
  `sawFile` and `uploadEcho`'s handler write. That result is retained as the
  red proof; the test was not treated as a reliable security assertion.

### Group 5 green evidence to date

- B22 now stages only an owned hidden temp, `Sync`s that file, uses
  `os.Root.Link` as the atomic no-replace final-name claim, removes only the
  owned temp, and syncs the containing `os.Root` directory. The focused test
  kills a helper process only after its first temp bytes are staged and proves
  no final name remains; its in-process race half preserves the pre-publish
  foreign sentinel byte-for-byte/on the same inode and removes the owned temp.
- W05's focused symlink test passes under `-race` after waiting on the handler
  completion channel before reading the observation.

### Group 5 atomic closure gates (2026-08-21)

- `go test -timeout 20m ./internal/connectors/engine -count=1` passed
  (`10.716s`); `go test -timeout 20m ./internal/connectors/connsdk -count=1`
  passed (`5.656s`).
- Race gates passed: the binary crash/no-replace test (`2.208s`) and the
  multipart escaping-symlink refusal repeated twenty times (`1.398s`) under
  `-race`.
- `go vet ./internal/connectors/engine ./internal/connectors/connsdk`,
  `go build ./cmd/pm`, and `git diff --check` passed. This group changes no
  generated declaration or CLI/manual surface.

## Group 6 red-contract evidence (2026-08-21)

- **B33:** `TestCloneRuntimeConfigDefensivelyCopiesCatalogNestedState` failed
  before the production correction. Mutating the request clone's catalog field,
  primary-key/cursor slices, raw schema, and discovery failure changed the
  caller-owned runtime; its first failure reported the source field renamed to
  `clone`. This establishes the nested aliasing boundary independently of
  ordinary record cloning. Arrow request cloning will use the same deep runtime
  clone and must create a fresh apply request for each segment.
- **B20:** `TestOrchestratorRevalidatesDestinationAuthorizationImmediatelyBeforeApply`
  failed with a nil run result after the test revoked authority in the completed
  warehouse-stage callback. The existing admission happened only before stage,
  and `ApplyDestination` still executed. This proves the live check was not at
  the external mutation boundary.
- **B36:** `TestRequesterRechecksRequestAdmissionBeforeRetry` initially failed
  to compile because the HTTP request boundary exposed no admission hook at
  all. The existing cohort wrapper checked only once around an operation, so a
  retry/page/send could not re-read a durable fence after the first request.
  `TestAuthCohortFenceIndeterminateCommitCancelsOldLocalEpoch` then failed at
  its one-second cancellation boundary: a post-rename directory-sync error had
  already persisted the fence, but `Fence` returned early without cancelling
  old local admissions.
- **B26:** `TestTransportTwoAppsFenceBeforeAnySideEffect` first failed under
  the former late-checkpoint model: the second App reached source/stage/apply
  before its state CAS could lose. The replacement holds the first owner at
  its pre-I/O source boundary and proves the contender has exactly zero
  source, stage, apply, and publication calls.
- **W06/W07:** the post-rename Create test initially left zero live timers
  despite one durable parked record, proving memory had returned on an
  indeterminate write without reloading it. The terminal authorization test
  initially failed to compile because parking had no closed
  `needs_reauthorization` outcome or cross-coordinator claim refusal.

## Group 6 focused green evidence (2026-08-21)

- **B20:** ordinary/full-overwrite and Arrow paths call the standing approval
  exactly once immediately before each apply/publication effect. The prior
  per-unit regression is now `TestOrchestratorRechecksAuthorizationImmediatelyBeforeEachApplyAndRefusesSecondEffect`: two pages stage, only the first provider apply occurs after the second final check revokes authority. The final-after-stage ordinary, full-overwrite, serial Arrow, and pipeline Arrow tests pass.
- **B26:** `TestTransportTwoAppsFenceBeforeAnySideEffect` and the ordinary,
  per-page, full-overwrite, and CDC/restart focused App cohort pass. A durable
  `ActiveWorkID`/monotonic fence is claimed before source I/O, renewed at each
  source boundary, and retired only by the matching terminal work owner;
  terminal/crashed owners are safely replaced while live owners have zero
  contender effects.
- **B33:** `TestCloneRuntimeConfigDefensivelyCopiesCatalogNestedState` passes.
  Runtime catalogs now deep-copy stream fields, keys, cursors, raw schemas and
  discovery failures; serial and pipeline Arrow calls receive fresh cloned
  request objects. The focused Arrow clone cohort passes under `-race`.
- **B36:** retry and redirect request-admission tests pass; auth-cohort
  indeterminate fence cancellation and the PostgreSQL pre-I/O admission test
  pass. HTTP sends, redirects, refreshes, PostgreSQL connects/queries, and
  transaction statements re-check the admitted durable epoch at their physical
  boundary while cleanup rollback remains available.
- **W06:** `TestRateParkingCoordinator_ReconcilesPostCommitCreateBeforeReturningUncertainty`, `TestRateParkingCoordinator_ReconcilesEachPostCommitResumeMutation`, and `TestRateParkingCoordinator_ReconcilesPostCommitRearmAndDelete` pass. They inject an indeterminate post-rename outcome after each of Create, Rearm, Claim, BeginResume, MarkResumeCompleted, Complete, and Delete, then assert the exact reloaded record set and one-or-zero correct timer outcome.
- **W07:** expired and revoked App authorization errors become a secret-free
  durable `needs_reauthorization` parking outcome. The memory and reopened
  file-store tests prove no retry timer or second-coordinator claim can revive
  it; its opaque scope remains blocked until explicit safe cancellation.

### Group 6 atomic closure gates

- `go test -timeout 20m ./internal/coordination -count=1` (`4.276s`),
  `./internal/connectors/connsdk` (`5.983s`),
  `./internal/connectors/native/postgres` (`1.471s`), and
  `./internal/synctransport` (`0.766s`) passed.
- Focused App work-fence, full-overwrite, CDC recovery, and terminal
  authorization cohort passed (`6.773s`); the scoped Arrow clone cohort passed
  under `go test -race` (`1.821s`).
- `go vet` over every Group 6 package, `go build ./cmd/pm`, and
  `git diff --check` passed. No declaration or generated-surface source
  changed, so the final full generator sweep remains reserved for the exact
  all-group SHA.

## Group 1 frozen GitHub mutation delta crosswalk

`874 -> 906` is held provisional until this crosswalk, the runner-bound proof,
and the generator suite are green. The independent test
`TestGitHubFoundationMutationDeltaHasUniqueClosedBoundedSourceCrosswalk`
compares the base candidate artifact at
`c9824b5837f487acaa2c2a39126d29cf401d7fb5` with the generated artifact,
requires exactly these 32 unique command paths, rejects duplicate write actions
or source identities, checks the `fixture_required_mutations`/`reverse_etl`/
`reverse_plan` cohort, and rechecks the immutable source URL, SHA-256, and
12,920,264-byte count. Every root has `additionalProperties: false`.

Bound legend: `S32` is a commandrunner-enforced 32,768-byte UTF-8 cap plus
schema `maxLength: 8192`; `S1120` is 1,120 bytes plus `maxLength: 280`; `J1M`
is the one-value commandrunner JSON cap of 1,048,576 bytes; `A256` is
`maxItems: 256`; `O256` is `maxProperties: 256` on the named dynamic object;
`C` is a recursively closed, bounded named JSON schema; `I`, `B`, and `E` are
the runner's parsed integer, parsed boolean, and finite enum forms. No row has
a raw body, method, path, or header channel.

| Unique command path | Immutable source identity | Write declaration | Method / path | additionalProperties | Certified cohort | Effective runner-enforced inputs |
| --- | --- | --- | --- | --- | --- | --- |
| `api agents set-selected-repos-for-org-secret` | `agents/set-selected-repos-for-org-secret` | `agents_set_selected_repos_for_org_secret` | `PUT /orgs/{org}/agents/secrets/{secret_name}/repositories` | false | fixture_required_mutations | org S32; secret_name S32; selected_repository_ids J1M+A256 |
| `api agents set-selected-repos-for-org-variable` | `agents/set-selected-repos-for-org-variable` | `agents_set_selected_repos_for_org_variable` | `PUT /orgs/{org}/agents/variables/{name}/repositories` | false | fixture_required_mutations | name S32; org S32; selected_repository_ids J1M+A256 |
| `api agents update-org-variable` | `agents/update-org-variable` | `agents_update_org_variable` | `PATCH /orgs/{org}/agents/variables/{name}` | false | fixture_required_mutations | name S32; org S32; selected_repository_ids J1M+A256; value S32; visibility E |
| `api code-scanning update-alert` | `code-scanning/update-alert` | `update_code_scanning_alert` | `PATCH /repos/{owner}/{repo}/code-scanning/alerts/{alert_number}` | false | fixture_required_mutations | alert_number I; assignees J1M+A256; create_request B; dismissed_comment S1120; dismissed_reason E; state E |
| `api dependabot update-alert` | `dependabot/update-alert` | `update_dependabot_alert` | `PATCH /repos/{owner}/{repo}/dependabot/alerts/{alert_number}` | false | fixture_required_mutations | agent_assignment J1M+C; alert_number I; assignees J1M+A256; dismissed_comment S1120; dismissed_reason E; state E |
| `api git create-ref` | `git/create-ref` | `create_ref` | `POST /repos/{owner}/{repo}/git/refs` | false | fixture_required_mutations | ref S32; sha S32 |
| `api git update-ref` | `git/update-ref` | `update_ref` | `PATCH /repos/{owner}/{repo}/git/refs/{ref}` | false | fixture_required_mutations | force B; ref S32; sha S32 |
| `api issues add-assignees` | `issues/add-assignees` | `add_issue_assignees` | `POST /repos/{owner}/{repo}/issues/{issue_number}/assignees` | false | fixture_required_mutations | assignees J1M+A256; issue_number I |
| `api issues add-labels` | `issues/add-labels` | `add_issue_labels` | `POST /repos/{owner}/{repo}/issues/{issue_number}/labels` | false | fixture_required_mutations | issue_number I; labels J1M+A256 |
| `api issues create-milestone` | `issues/create-milestone` | `create_milestone` | `POST /repos/{owner}/{repo}/milestones` | false | fixture_required_mutations | description S32; due_on S32; state E; title S32 |
| `api issues remove-assignees` | `issues/remove-assignees` | `remove_issue_assignees` | `DELETE /repos/{owner}/{repo}/issues/{issue_number}/assignees` | false | fixture_required_mutations | assignees J1M+A256; issue_number I |
| `api issues set-labels` | `issues/set-labels` | `set_issue_labels` | `PUT /repos/{owner}/{repo}/issues/{issue_number}/labels` | false | fixture_required_mutations | issue_number I; labels J1M+A256 |
| `api issues update-comment` | `issues/update-comment` | `update_issue_comment` | `PATCH /repos/{owner}/{repo}/issues/comments/{comment_id}` | false | fixture_required_mutations | body S32; comment_id I |
| `api issues update-milestone` | `issues/update-milestone` | `update_milestone` | `PATCH /repos/{owner}/{repo}/milestones/{milestone_number}` | false | fixture_required_mutations | description S32; due_on S32; milestone_number I; state E; title S32 |
| `api pulls create-review-comment` | `pulls/create-review-comment` | `create_review_comment` | `POST /repos/{owner}/{repo}/pulls/{pull_number}/comments` | false | fixture_required_mutations | body S32; commit_id S32; in_reply_to I; line I; path S32; position I; pull_number I; side E; start_line I; start_side E; subject_type E |
| `api pulls dismiss-review` | `pulls/dismiss-review` | `dismiss_pull_request_review` | `PUT /repos/{owner}/{repo}/pulls/{pull_number}/reviews/{review_id}/dismissals` | false | fixture_required_mutations | event E; message S32; pull_number I; review_id I |
| `api pulls request-reviewers` | `pulls/request-reviewers` | `request_reviewers` | `POST /repos/{owner}/{repo}/pulls/{pull_number}/requested_reviewers` | false | fixture_required_mutations | pull_number I; reviewers J1M+A256; team_reviewers J1M+A256 |
| `api pulls submit-review` | `pulls/submit-review` | `submit_pull_request_review` | `POST /repos/{owner}/{repo}/pulls/{pull_number}/reviews/{review_id}/events` | false | fixture_required_mutations | body S32; event E; pull_number I; review_id I |
| `api pulls update-review-comment` | `pulls/update-review-comment` | `update_review_comment` | `PATCH /repos/{owner}/{repo}/pulls/comments/{comment_id}` | false | fixture_required_mutations | body S32; comment_id I |
| `api repos add-collaborator` | `repos/add-collaborator` | `add_collaborator` | `PUT /repos/{owner}/{repo}/collaborators/{username}` | false | fixture_required_mutations | permission S32; username S32 |
| `api repos create-commit-comment` | `repos/create-commit-comment` | `create_commit_comment` | `POST /repos/{owner}/{repo}/commits/{commit_sha}/comments` | false | fixture_required_mutations | body S32; commit_sha S32; line I; path S32; position I |
| `api repos create-deployment` | `repos/create-deployment` | `create_deployment` | `POST /repos/{owner}/{repo}/deployments` | false | fixture_required_mutations | auto_merge B; description S32; environment S32; payload J1M+O256+maxLength 1,048,576; production_environment B; ref S32; required_contexts J1M+A256; task S32; transient_environment B |
| `api repos create-or-update-environment` | `repos/create-or-update-environment` | `create_or_update_environment` | `PUT /repos/{owner}/{repo}/environments/{environment_name}` | false | fixture_required_mutations | deployment_branch_policy J1M+C; environment_name S32; prevent_self_review B; reviewers J1M+A256; wait_timer I |
| `api repos create-or-update-file-contents` | `repos/create-or-update-file-contents` | `create_or_update_file` | `PUT /repos/{owner}/{repo}/contents/{path}` | false | fixture_required_mutations | author J1M+C; branch S32; committer J1M+C; content S32; message S32; path S32; sha S32 |
| `api repos create-webhook` | `repos/create-webhook` | `create_webhook` | `POST /repos/{owner}/{repo}/hooks` | false | fixture_required_mutations | active B; config J1M+O256; events J1M+A256; name S32 |
| `api repos delete-file` | `repos/delete-file` | `delete_file` | `DELETE /repos/{owner}/{repo}/contents/{path}` | false | fixture_required_mutations | author J1M+C; branch S32; committer J1M+C; message S32; path S32; sha S32 |
| `api repos merge` | `repos/merge` | `merge_branch` | `POST /repos/{owner}/{repo}/merges` | false | fixture_required_mutations | base S32; commit_message S32; head S32 |
| `api repos replace-all-topics` | `repos/replace-all-topics` | `replace_repo_topics` | `PUT /repos/{owner}/{repo}/topics` | false | fixture_required_mutations | names J1M+A256 |
| `api repos update-commit-comment` | `repos/update-commit-comment` | `update_commit_comment` | `PATCH /repos/{owner}/{repo}/comments/{comment_id}` | false | fixture_required_mutations | body S32; comment_id I |
| `api repos update-release-asset` | `repos/update-release-asset` | `update_release_asset` | `PATCH /repos/{owner}/{repo}/releases/assets/{asset_id}` | false | fixture_required_mutations | asset_id I; label S32; name S32; state S32 |
| `api repos update-webhook` | `repos/update-webhook` | `update_webhook` | `PATCH /repos/{owner}/{repo}/hooks/{hook_id}` | false | fixture_required_mutations | active B; add_events J1M+A256; config J1M+O256; events J1M+A256; hook_id I; remove_events J1M+A256 |
| `api secret-scanning update-alert` | `secret-scanning/update-alert` | `update_secret_scanning_alert` | `PATCH /repos/{owner}/{repo}/secret-scanning/alerts/{alert_number}` | false | fixture_required_mutations | alert_number I; assignee S32; resolution E; resolution_comment S32; state E; validity E |

## Group 9 execution plan (2026-08-21)

- **GSD/manual fallback and skills:** this remains the single-owner inline GSD
  lane recorded in `POSTFIX-EXECUTION.md`; no compatible isolated GSD worker is
  available. `golang-how-to`, `golang-cli`, `golang-testing`,
  `golang-error-handling`, `golang-security`, `golang-safety`,
  `golang-design-patterns`, `golang-structs-interfaces`, and
  `golang-documentation` govern this evidence and generator boundary. The
  exact clean remote base is
  `0b061b0f3149ba9b050f6a7b7ec3cc2494c08f0c`.
- **B10 red contract:** a production-shaped repository graph fixture must
  reject independently mutated implementation SHA, diff-base SHA, each of
  the five component head SHAs, and each preserving merge; it must also reject
  a current-HEAD/self-reference claim. The reviewed SHA is checked only
  against the review record and can never stand in for the implementation
  identity. A valid evidence closure has a distinct evidence commit directly
  above the verified implementation commit, so the evidence file never claims
  to hash its own commit.
- **B11 red contract:** `TestCertificationEvidenceBecomesStaleWhenSubjectChanges`
  begins with a matching live proof, then independently changes the PM binary,
  build identity, declarations digest, source/projection digest, CLI command
  mapping digest, relevant configuration digest, and proof protocol. Each
  mutation must move the record to historical/stale evidence and clear
  `live_tested`; only an exact new subject can restore it.
- **B01 final-closure red contract:** the closure records an exact digest for
  every regenerated source, CLI, docs, website, skills, ledger, matrix,
  candidate, sweep, and Foundation-evidence artifact. It rejects a duplicate
  path, an omitted required category, a mutated byte, and a subject/implementation
  mismatch. The final regeneration is deliberately one ordered producer pass
  from the immutable implementation commit; `--check` proves no second pass
  would change bytes.
- **Frozen intended red command:**
  `go test -timeout 20m ./cmd/connectorgen -run '^(TestFoundationEvidenceRejectsEveryStaleGraphIdentityAndArtifact|TestCertificationEvidenceBecomesStaleWhenSubjectChanges|TestCertificationSubjectFingerprintIncludesEveryComponent)$' -count=1 -v`.
  The initial compilation failure is expected until the graph-aware evidence
  validator, deterministic subject builder, and artifact-closure validator
  exist. No production code or generated artifact is changed before that full
  named set is present and observed.
- **Full package red set (2026-08-21, 111.488s):** the first non-overlapping
  `go test -timeout 20m ./cmd/connectorgen -count=1` exited 1 with exactly
  `TestCertificationEvidencePostgresTransportPromotesOnlyCompletedModes`,
  `TestCertificationEvidencePostgresChangeCapturePromotesOnlyReceiptBackedBinaryProof`,
  `TestCertificationEvidenceReportImportsDefinitionBoundHTTPProofWithoutSecrets`,
  `TestCertificationEvidenceReportUsesSecondConnectorDefinitionWithoutSharedBranch`,
  `TestCertificationEvidenceWriterUsesRepositoryLocalSaltBeforePersistence`, and
  `TestCertificationCheckIgnoresMalformedNonAllowlistedRuntimeLedgerEntry`.
  Each disposable test root lacked
  `internal/connectors/certifications/current-subject.json`; the new
  declaration-owned producer/check gate correctly refused to publish or
  project unbound evidence. The green correction must give only those
  disposable fixtures a deterministic non-secret valid subject, and the
  malformed-ledger test must regenerate its fixture output before proving its
  unrelated-ledger assertion. Production remains fail-closed when the artifact
  is absent or does not match its exact subject.
- **Green after the complete disposition:** the named focused suite passed in
  `83.753s`, and the one non-overlapping full rerun passed in `152.752s`:
  `go test -timeout 20m ./cmd/connectorgen -count=1`. The corrected fixtures
  assert the exact subject is embedded in successful evidence; they do not
  change the absent/mismatched-subject production refusal.
- **Atomic-subcommit green gate:** after adding the closure-subject and
  trailing-manifest negatives, the focused suite passed in `80.972s` and the
  final package pass completed in `146.597s`. `go vet ./cmd/connectorgen`,
  `go build ./cmd/connectorgen`, `connectorgen validate internal/connectors/defs`
  (552 connectors, zero findings), and `connectorgen surface-sync --check`
  (552 scanned, zero changes) also passed. The certification-subject and
  final artifact/evidence producers are intentionally deferred until both
  incoming immutable component heads are ancestry-preservingly merged, so this
  subcommit contains no generated final Foundation evidence.
- **Isolation correction:** the new disposable command-workspace bootstrap
  initially exposed that its older helper attempted `os.Link`, so a generator
  write could mutate the source `postgres/certification-matrix.json`. The
  exact source file was restored before any commit. The helper now always
  byte-copies, and
  `TestCertificationCheckIgnoresMalformedNonAllowlistedRuntimeLedgerEntry`
  snapshots and proves the source matrix byte-identical around its bootstrap.
  The focused red/green check passed in `80.027s`, then the final
  package-wide `go test -timeout 20m ./cmd/connectorgen -count=1` passed in
  `138.887s`; no generated matrix is retained by this pre-merge subcommit.

### Structured-body preserving merge gate — frozen diagnostic failure set (2026-08-22)

- Context: the active non-squash merge of exact PR #4313 head
  `66b4c3cbea00bf71cfc4082d09cd7f72917e08b5` adds its declaration-bound
  structured-body contracts to the Group-9 implementation parent. Focused
  red/green corrections covered nested `body.targets.0.token` withholding,
  terminal secret-store receipt preservation, conditional API-key query
  ownership, and sparse array CLI projection before this package gate.
- Frozen complete failure set from
  `go test -timeout 20m ./internal/connectors/engine ./internal/connectors/commandrunner -count=1`:
  `TestOperationDirectWriteMultipartNeverRetriesOrReplaysRedirect` (both
  `Too_Many_Requests` and `Internal_Server_Error` cases),
  `TestOperationDirectWriteNeverRetriesNonIdempotentFailure`,
  `TestOperationDirectWriteRetainsTerminalGraphQLHTTPResponse`, and
  `TestOperationDirectWriteSecretOperationRetainsProviderResponses` (its
  `non_JSON_HTTP_error` case). `internal/connectors/commandrunner` passed in
  the same command.
- Cause: the first reconciliation made ordinary provider HTTP bodies printable
  unconditionally. That contradicted the established secret/multipart/GraphQL
  public-diagnostic boundary. The repair must distinguish safe ordinary REST
  diagnostic detail from those existing protected operation classes while
  retaining the exact typed provider receipt for durable result handling.
- Green: `operationDirectWritePrintsProviderHTTPBody` now permits printable
  provider detail only for ordinary fixed REST `json` results, after exact
  declaration/request/config/credential redaction. Multipart, GraphQL,
  `json_redacted`, secret-stored, and other redacted contracts retain their
  complete provider receipt only in the typed result/cause. The full rerun
  `go test -timeout 20m ./internal/connectors/engine ./internal/connectors/commandrunner -count=1`
  passed for engine (14.253s); the independent commandrunner package rerun and
  `go test -timeout 20m ./internal/app -count=1` also completed green. The
  focused red/green regressions cover ordinary safe diagnostics, an echoed
  declared header credential, nested structured-body redaction, secret-store
  terminal receipt preservation, conditional API-key ownership, and sparse
  array projection.

### Group 9 generated-surface reconciliation proof (2026-08-22)

- **Red:** `TestETLManualAndSkillDescribeDeliveredReconciliationTerminalRun`
  failed before source changes because the Group 8
  `delivered_reconciliation_required` contract existed only in previously
  hand-edited generated `docs/cli/etl.md` and `docs/skills/pm-etl/SKILL.md`.
  `pm etl`, the closed declarative destination help, and `baseSkillDocs` did
  not describe the durable terminal `ETLRun`, nonzero exit, durable repair
  before endpoint resolution, or no-replay boundary. The initial focused run
  exited 1 in 10.403s with the missing manual status as the complete failure.
- **Green:** the canonical `internal/cli/docs.go`,
  `internal/cli/etl_transport.go`, and `internal/cli/skills.go` generators now
  state the exact persisted terminal result and bounded no-replay repair. The
  focused test passed in 9.152s after the source correction. The final
  source/CLI/docs/website/skills/matrix/candidate/sweep generation pass starts
  from that corrected source; the earlier pre-correction generator output is
  diagnostic only and is not claimed as final evidence.

### Group 9 final package gate — frozen failure set (2026-08-22)

- The completion-tracked unchanged-`ab9cb9cca` foundation package command was
  `go test -count=1 -timeout 20m ./cmd/connectorgen ./internal/app
  ./internal/cli ./internal/connectors/engine ./internal/connectors/commandrunner
  ./internal/synctransport ./internal/connectors/connsdk
  ./internal/connectors/native/amazon-sqs ./internal/connectors/native/postgres
  ./internal/connectors/database ./internal/connectors/certify`. It reported
  exactly four failing contracts before any follow-up correction:
  `TestCertificationMatrixPromotesPostgresChangeCaptureOnlyWithReceiptBackedLiveProof`,
  `TestPostgresPublishesOnlyGenericCapabilitiesWithMatchingLiveCertification`,
  `TestGoldenTranscripts/help_etl`, and
  `TestGenerateRecordForGitHubLabelIncludesColor`. A focused attempt to select
  a hypothetical `help_etl_json` transcript was itself refused because no such
  independently named fixture exists; it is not an additional product failure.
- The complete disposition is required before another full CLI/package run:
  the first two tests retained stale PostgreSQL live-proof expectations after
  the Group 9 subject check correctly made their old binary/protocol proof
  historical; the ETL transcript was the intended source-help addition not
  yet captured in the golden fixture; and the label test incorrectly promoted
  the source-declared optional `color` input to required instead of proving
  the bounded optional field remains available and its declared override is
  preserved. All other packages in that command passed: App (312.245s), engine
  (21.914s), commandrunner (27.698s), synctransport (4.002s), connsdk
  (11.368s), native SQS (8.191s), native PostgreSQL (1.493s), database
  (13.815s), and certify otherwise (21.629s). `cmd/connectorgen` failed only
  the named stale-proof assertions (198.112s); CLI failed only the named
  golden transcript contract (871.161s).

### Group 9 frozen-failure red/green disposition (2026-08-22)

- **Red:** `TestCertificationMatrixPromotesPostgresChangeCaptureOnlyWithReceiptBackedLiveProof`
  and `TestPostgresPublishesOnlyGenericCapabilitiesWithMatchingLiveCertification`
  failed because they treated a persisted PostgreSQL receipt whose binary and
  proof protocol fingerprint no longer matches the current subject as live.
  The existing `TestCertificationEvidenceBecomesStaleWhenSubjectChanges`
  independently mutates every subject component; the focused replacement
  additionally proves that the exact stale CDC capability and change-capture
  sync-mode receipts remain retained as historical (one each), contribute no
  live evidence, leave the declared CDC capability implementation intact, and
  keep the exact change-capture route unimplemented until current matching
  proof exists. **Green:** `go test -count=1 -timeout 20m ./cmd/connectorgen
  -run '^(TestCertificationMatrixKeepsPostgresChangeCaptureEvidenceHistoricalWhenSubjectDiffers|TestPostgresPublishesOnlyGenericCapabilitiesWithCurrentLiveCertification)$'
  passed (28.759s and 47.409s independently).
- **Red:** `TestGenerateRecordForGitHubLabelIncludesColor` wrongly required
  `color`. The pinned immutable GitHub artifact at
  `80850db290cde4eb487e0efb587cf27f305e77b6bef96933ed8a09b5169d5b1d`
  declares exactly `[name]` required while `color` is an optional string.
  The accepted `writes.json` projection carries that field with
  `maxLength: 8192`; the declaration-owned certification override supplies
  `ededed`. **Green:** renamed
  `TestGenerateRecordForGitHubLabelPreservesOptionalColorOverride` asserts the
  exact required set, optional typed/bounded field, and declared override;
  `go test -count=1 -timeout 20m ./internal/connectors/certify -run
  '^TestGenerateRecordForGitHubLabelPreservesOptionalColorOverride$'` passed
  (1.366s). This preserves the bounded provider input rather than promoting it
  to a required field.
- **Red:** `TestGoldenTranscripts/help_etl` differed because the tested
  `etlManual` source had gained the Group 8 durable
  `delivered_reconciliation_required` explanation but the selected generated
  transcript was stale. **Green:** the fixture was regenerated only through
  `POLYMETRICS_UPDATE_GOLDEN_TRANSCRIPTS=1
  POLYMETRICS_GOLDEN_TRANSCRIPT_NAMES=help_etl go test -count=1 -timeout 20m
  ./internal/cli -run '^TestGoldenTranscripts$'`; a non-update rerun of the
  same selected transcript passed (1.218s). The source contract,
  `TestETLManualAndSkillDescribeDeliveredReconciliationTerminalRun`, tracked
  skill parity, and tracked manual parity then passed together in 164.581s.
- Required skills applied for this correction: `golang-how-to`,
  `golang-testing`, `golang-cli`, `golang-error-handling`, `golang-safety`,
  `golang-security`, and `golang-documentation`. The existing manual GSD
  fallback remains applicable: this exact continuation has planning, red, and
  green evidence here and no compatible isolated worker may be spawned under
  the canonical single-worker delivery contract.

### Group 9 final package gate — second frozen failure set (2026-08-22)

- The one permitted post-disposition broad rerun of
  `go test -count=1 -timeout 20m ./cmd/connectorgen ./internal/app
  ./internal/cli ./internal/connectors/engine ./internal/connectors/commandrunner
  ./internal/synctransport ./internal/connectors/connsdk
  ./internal/connectors/native/amazon-sqs ./internal/connectors/native/postgres
  ./internal/connectors/database ./internal/connectors/certify` completed at
  unchanged `f59b7cbe35aa723255feebd271e35e8a00b90577`. Its complete product
  failure set is the two derived transcript cases
  `TestGoldenTranscripts/bare_etl_manual` and
  `TestGoldenTranscripts/json_etl_manual`; both omit the exact
  `delivered_reconciliation_required` manual paragraph already proved at the
  shared source. A focused no-update reproduction selecting precisely those
  two names failed with precisely those two subtests and no others. The broad
  run's remaining packages all passed: connectorgen (173.000s), App (293.485s),
  engine (20.566s), commandrunner (29.933s), synctransport (9.618s), connsdk
  (10.145s), native SQS (8.037s), native PostgreSQL (5.105s), database
  (13.654s), and certify (15.314s); CLI otherwise ran 83 bounded certification
  command invocations. Redis connection-refused lines were expected isolated
  negative-test diagnostics, not additional failing contracts.
- **Red:** the two transcript names above. **Green:** regenerated only those
  two declaration-derived views from `etlManual` with
  `POLYMETRICS_UPDATE_GOLDEN_TRANSCRIPTS=1
  POLYMETRICS_GOLDEN_TRANSCRIPT_NAMES=bare_etl_manual,json_etl_manual go test
  -count=1 -timeout 20m ./internal/cli -run '^TestGoldenTranscripts$'` and
  reran the same selected names without update; both passed (1.188s). No second
  full CLI/package run occurs until this derived-fixture checkpoint is remote
  verified.
