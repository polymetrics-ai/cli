# TDD ledger — MySQL container harness R1

| ID | Guarantee | Red assertion | Green proof |
| --- | --- | --- | --- |
| H1 | Direct endpoint scope | A named, default, or remote Podman target can reach another lane. | `TestNewRejectsUnsafeEndpoints` refuses all non-local targets; `TestEveryTargetCommandRechecksTargetIdentity` proves every daemon command is immediately preceded by target identity evidence. |
| H2 | Isolation and sequencing | Runs can collide on a resource name or multiply database peak disk/memory. | Random generated container/volume/tag names are unit-tested; `TestStartReleasesTheEngineSlotOnEveryExitPath` proves the default one-slot semaphore returns on success and failure. |
| H3 | Unconditional cleanup | Assertion failure or interrupt can leave a container, volume, or image behind. | `TestCloseDuringCreateStillRemovesTheCreatedResource` and `TestStartRefusesToCreateAfterClose` cover the create/interrupt race; `Close` continues all cleanup stages and the live post-run scoped listings were empty. |
| H4 | Image ownership | Cleanup guesses whether a shared source image belongs to this run. | A successful pull is re-tagged to a run-specific local reference before startup; `TestCleanupRemovesOnlyRunOwnedResources` and `TestStartInspectsTheSourceImageOnlyToSizeThePull` prove source retention and generated-only cleanup. |
| H7 | Interrupt coverage spans teardown | Cleanup drops the interrupt handler before its own removals, so a Ctrl-C during teardown exits with resources still on the machine. | `TestCloseKeepsInterruptCleanupArmedUntilTeardownFinishes` probes the registry at every removal command and asserts the handler is still installed, then that it is released once teardown returns. |
| H8 | Pull headroom | A pull started on a target with insufficient or unproven image-store capacity fills it partway through and leaves a partial image plus a container to clean up. | `TestStartRefusesToPullWithoutHeadroomForTheImage` proves the documented threshold; `TestStartFailsClosedWhenCachedTargetCapacityCannotBeProven` proves cached sources still need capacity proof; `TestNewRequiresTheImageFootprint` stops an engine from silently skipping the gate. |
| H5 | Non-default endpoint | A connector silently assumes 3306 on the host. | `TestStartPublishesOnlyLoopback` asserts `127.0.0.1::<port>` publishing; `TestParseMappedPortRefusesTheEngineDefaultPort` refuses a default host mapping. |
| H6 | Target-store accounting | Capacity can be measured on a different filesystem than the selected daemon's image store. | `TestTargetCapacityUsesTheProvenStorePath` checks `DiskFreeAt` receives the reported graph root; unprovable target identity or capacity fails before resource mutation. |
| H9 | Forwarded local Podman machines | Podman 5.3 can expose a direct host Unix forward while reporting VM-local Unix socket and image-store paths, which a host-only proof rejects despite daemon capacity evidence. | `TestStartUsesPodmanMachineDaemonCapacity` proves a forwarded Unix report uses `GraphRootAllocated - GraphRootUsed`, never a host path; malformed forwarded capacity remains a pre-mutation refusal. |
| M1 | Honest CDC declaration | MySQL can advertise a CDC capability before a production runtime entrypoint exists. | Metadata, definition, catalog, generated docs, and website data keep `cdc: false`; no changefeed descriptor or executor is registered, while the native row-event reader remains covered by internal evidence. |
| M2 | SQL safety | Stream/schema/cursor values are concatenated into SQL. | Identifier tests reject unsafe values; queries quote only validated identifiers and bind all values. |
| M3 | Complete ETL paging and incrementals | A read proves a single record or repeats/skips at shared cursor values. | Keyset query tests cover primary-key and `(cursor, primary_key)` boundaries; the real five-row seed with `page_size=2` proves multiple pages and exact incremental rows. |
| M4 | Change capture state | Binlog rows can advance state before acknowledgement or be projected against changed metadata. | CDC tests bind checkpoint to ordered column fingerprint and position/row ordinal; live insert/update/delete events assert actual returned rows. |
| M5 | #3902 direct-read boundary | A database ETL reader is mistaken for a pagewise HTTP direct-read command and omits page context. | `TestReadIsETLNotPagewiseDirectRead` proves MySQL exposes neither `DirectReader` nor `OperationDirectReader`; the SQL `Read` drains its bounded pages to ETL rather than returning an exploratory page. |
| T1 | TLS is enforced, not declared | A strict TLS mode can quietly downgrade or be ignored by replication. | TLS-less server tests make strict modes refuse; certificate tests prove verification; live checks use the server's own `Ssl_cipher`; the binlog syncer receives the same TLS config. |
| T2 | SQL option shape does not drift | MySQL and PostgreSQL expose incompatible TLS field names or validation. | `sqltls` centralizes the vocabulary. PostgreSQL's definition accepts the runtime vocabulary and `TestPostgresPoolConfigUsesSharedTransportSecurityOptions` proves CA/server-name application and no strict plaintext fallback. |
| T3 | Resumed verify-ca TLS remains verified | Go skips `VerifyPeerCertificate` on a resumed TLS session after built-in hostname verification is disabled, allowing a resumed session to bypass the manual chain check. | `TestVerifyCARejectsAChainOutsideTheConfiguredRoot` and `TestVerifyCAAcceptsTheConfiguredRootIgnoringHostname` require `VerifyConnection`; `TestVerifyCARevalidatesResumedSessions` proves it is called on both a full and resumed TLS 1.2 handshake. |
| G1 | Production native wiring contains only connectors | A test harness or shared helper is blank-imported solely because it sits below `native/`. | `TestGen_NativesetImportsRuntimePackagesAndExcludesSupportLibraries` fails for `dbtest`/`sqltls`; `nativeSupportPackages` and regenerated wiring exclude both while retaining real connector packages. |

> Historical machine/connection execution entries below remain as audit history. The direct-endpoint
> rows above are the active harness contract.

## Red / Green execution evidence

**Red — initial implementation:**

```text
go test -count=1 ./internal/connectors ./internal/connectors/native/dbtest \
  -run 'Test(ChangefeedDescriptorAcceptsBinlogReplication|NewRejectsUnscopedConnection|Cleanup)'
```

This failed as intended before the harness, MySQL connector, and `binlog_replication` vocabulary
existed.

**Green — implementation:** the focused equivalent passed after those components were added.

**Red — post-rebase TLS reconciliation:**

```text
go test -count=1 -timeout 20m ./internal/connectors/native/postgres
```

`TestDefinitionAcceptsSharedTransportSecurityVocabulary` initially failed because the embedded
PostgreSQL schema rejected `disabled` even though `resolveConfig` accepted it.

**Green — post-rebase TLS reconciliation:** the same package passed after adding the shared schema
keys, `poolConfig`, and unified `openPool` usage; MySQL, PostgreSQL, and `sqltls` focused tests then
passed together.

**Red — pipeline-custody recovery:**

```text
go test -count=1 -timeout 20m ./cmd/connectorgen \
  -run '^TestGen_NativesetImportsRuntimePackagesAndExcludesSupportLibraries$'
```

The generator initially blank-imported the test-only `dbtest` directory; after that was excluded,
the same red test exposed the shared `sqltls` library. Neither package is a connector registration
unit.

**Green — pipeline-custody recovery:** `nativeSupportPackages` excludes both support libraries,
`go run ./cmd/connectorgen gen` regenerated `nativeset_gen.go`, and the focused test passed.

**Red — review round 1 fixes (harness lifecycle, machine ownership, CDC blast radius):**

```text
go test -count=1 -timeout 5m ./internal/connectors/native/dbtest \
  -run 'Test(StartRefusesToCreateAfterClose|CleanupNeverTrimsAMachineThisRunCannotProveItOwns|SetMaxConcurrentEnginesRefusesAChangeWhileASlotIsHeld)'
go test -count=1 -timeout 5m ./internal/connectors/native/mysql \
  -run 'Test(CDCQueryEventsFailClosed|PrimaryKeyDiscoveryIsBoundedToTheStreamBeingRead|DialErrorsKeepCancellationDistinguishableWithoutLeakingConfiguration)'
```

Each new assertion was run against the pre-fix behaviour, restored one at a time, and observed red:

- `TestStartRefusesToCreateAfterClose` failed with "a Start refused after Close leaked the engine
  slot" while `Close` alone returned the token, because `closeOnce` had already fired.
- `TestSetMaxConcurrentEnginesRefusesAChangeWhileASlotIsHeld` did not fail cleanly without the
  precondition guard — it hung until the 60s package timeout, which is the exact deadlock the guard
  now refuses: `releaseSlot` blocked forever on the replaced channel after cleanup had run.
- `TestCleanupNeverTrimsAMachineThisRunCannotProveItOwns` failed on all three cases when ownership
  was assumed rather than proven, showing an `fstrim` issued against an unowned machine.
- `TestCDCQueryEventsFailClosed` failed on `unrelated_schema_change`,
  `unrelated_qualified_schema_change`, `server_wide_grant`, `unrelated_analyze`, and
  `database_name_only_inside_a_literal` while every non-benign statement ended the changefeed
  regardless of which schema it touched.
- `TestInterruptCleanupClosesEveryLiveHarnessBeforeExiting` had no target at all before the fix:
  each harness registered its own signal handler and exited the process from whichever cleanup
  finished first, so there was no process-level registry to close the siblings.
- `TestPrimaryKeyDiscoveryIsBoundedToTheStreamBeingRead` and
  `TestDialErrorsKeepCancellationDistinguishableWithoutLeakingConfiguration` did not compile against
  the pre-fix signatures (`discoverPrimaryKeys` took no table, `dialError` did not exist).

**Green — review round 1 fixes:**

```text
go vet ./internal/connectors/native/...
go test -count=1 -timeout 5m ./internal/connectors/native/dbtest ./internal/connectors/native/mysql
```

Both packages passed with the sources restored. `internal/connectors/defs/mysql/docs.md` now records
the MySQL 8.4+ CDC bound with 8.4.11 as the verified server, the scoped statement-event rejection
rule, and the proven-ownership host-disk reclaim; the website connector artifacts were regenerated
from that bundle and changed only the `mysql` entry.

**Red — review round 3 fixes (ANSI-quoted CDC identifiers, machine ownership record):**

```text
go test -count=1 ./internal/connectors/native/mysql -run TestCDCQueryEventsFailClosed
go test -count=1 ./internal/connectors/native/dbtest -run TestCleanupNeverTrimsAMachineThisRunDidNotCreate
```

Both assertions were run against the pre-fix behaviour, restored, and observed red:

- `TestCDCQueryEventsFailClosed` failed on `cross_schema_ansi_quoted_alter` and
  `cross_schema_ansi_quoted_drop` while `statementIdentifiers` sent every `"`-delimited span to
  `skipQuotedLiteral`. Under `sql_mode=ANSI_QUOTES` that span is an identifier, so
  `ALTER TABLE "pm_harness"."events" CHANGE COLUMN label title VARCHAR(64)` logged under another
  default schema spelled no identifier the scan could see and the changefeed skipped it. A rename
  or type change keeps the column count, so nothing downstream would have caught it either.
- `TestCleanupNeverTrimsAMachineThisRunDidNotCreate` failed on
  `caller_supplied_a_machine_this_run_never_created` and `ownership_was_released_before_cleanup`
  while the gate was a name comparison: it issued `machine ssh <name> sudo fstrim -av` against a
  machine the process never created, and against one it had already removed. `fstrim -av` reaches
  every filesystem on a machine, so a name match is not an ownership claim.
- `TestNewMachineCreatesAndRemovesOnlyItsOwnMachine` had no target before the fix: there was no API
  to create an owned machine and therefore no ownership record to check.

**Green — review round 3 fixes:**

```text
gofmt -l cmd internal
go vet ./internal/connectors/native/... && go vet -tags databaseintegration ./internal/connectors/native/mysql
go test -count=1 -timeout 5m ./internal/connectors/native/dbtest ./internal/connectors/native/mysql
```

Both packages passed. The reclaim gate was then proven live rather than by inference: see
VERIFICATION.md "Live MySQL proof — re-run against the ownership-gated reclaim", which recorded
`reclaimed=true` on a machine the test created (`pmdb-mysql-eddc7350dec5`) and removed again.

One defect surfaced only in that live run: Podman refuses a machine name longer than 30 characters,
and the machine runner discards command output by design, so the first attempt failed with an opaque
`create database test machine: podman machine command failed`. Machine names are now built to a
shorter budget than container names and the length is asserted in `NewMachine` and in
`TestNewMachineCreatesAndRemovesOnlyItsOwnMachine`, so the limit fails as a named error rather than
as an unattributable Podman exit.

**Red — review round 4 fixes (machine-init leak, CDC quoted-span desync):**

```text
go test -count=1 ./internal/connectors/native/dbtest -run TestNewMachineRemovesWhatAFailedInitLeftBehind
go test -count=1 ./internal/connectors/native/mysql -run TestCDCQueryEventsFailClosed
```

Both were run against the pre-fix behaviour, restored, and observed red:

- `TestNewMachineRemovesWhatAFailedInitLeftBehind` failed all three cases with "NewMachine()
  returned no handle for a machine init may have created" while the ownership record and the
  interrupt registration were taken only after init reported success. `podman machine init` writes
  the VM config and a multi-GiB disk image before it can fail, and `exec.CommandContext` SIGKILLs it
  when the caller's context expires, so a machine created by a killed init had no handle, no
  ownership record, and no interrupt entry — and the integration test only defers `Remove` on a
  non-nil handle, so the disk image leaked.
- `TestCDCQueryEventsFailClosed` failed on `escaped_double_quote_before_a_target_reference`,
  `escaped_single_quote_before_a_target_reference` and all three unterminated-span cases. A `\"`
  inside a quoted span ended it one quote early, shifting every later boundary until
  `, RENAME TO pm_harness.t2` was swallowed into a single uncollected blob — no identifier matched
  the target, the default schema differed, and the statement was skipped. A RENAME keeps the column
  count, so `rows.ColumnCount != uint64(len(columns))` does not catch it either.

**Green — review round 4 fixes:**

```text
gofmt -l cmd internal
go vet ./internal/connectors/native/... && go vet -tags databaseintegration ./internal/connectors/native/mysql
go test -race -count=1 -timeout 5m ./internal/connectors/native/dbtest ./internal/connectors/native/mysql
```

Both packages passed. The quoting scan now applies one boundary rule — doubled quotes everywhere,
backslash escapes everywhere except a backquoted identifier, which is the single case MySQL settles
independently of `sql_mode` — and reports an unterminated span as unattributable so
`statementReachesDatabase` fails closed instead of concluding the statement is unrelated. The
existing "database name only inside a literal" case stays skipped.

**Red — review round 5 (harness ownership and headroom):**

```text
go test -count=1 -timeout 5m ./internal/connectors/native/dbtest \
  -run 'Test(CloseKeepsInterruptCleanupArmedUntilTeardownFinishes|StartRefusesToPullWithoutHeadroomForTheImage|CleanupRemovesTheSourceImageOnlyWithOwnershipOrAnExplicitOptIn|NewRequiresTheImageFootprint)'
```

All four failed as intended. `Close` unregistered the interrupt handler as its first statement, so
the probe recorded `registered=false armed=false` at every removal; `Start` sent `podman pull` with
no free-space measurement, so the ~854 MiB-free / ~830 MiB-image case pulled anyway; cleanup removed
`docker.io/library/mysql:8.4.11` from a machine the run never created, because the gate was the
opt-out `KeepImage` rather than machine ownership; and `New` accepted a config with no declared
image footprint, which silently skipped the gate.

**Green — review round 5:**

```text
gofmt -l cmd internal
go vet ./internal/connectors/native/... && go vet -tags databaseintegration ./internal/connectors/native/mysql
go test -race -count=1 -timeout 5m ./internal/connectors/native/dbtest ./internal/connectors/native/mysql
```

Passed. `Close` now defers the unregister so the handler covers every removal, proves machine
ownership once before both the source-image removal and the trim, retains a shared source image with
a named reason, and reports the host free-space delta as `HostDiskReleasedBytesEstimate` rather than
as per-image reclaimability. `Config.ExpectedImageBytes` is mandatory and `Start` refuses a pull
below three times that footprint.

**Red — review round 6 acceptance reconciliation:**

Inspection found MySQL declared `cdc: true` with a public `changefeed.json` and executor despite no
operator entrypoint; owned-machine initialization did not preserve the global default connection;
the pull gate used local `statfs` for a remotely targeted image store; the live TLS matrix omitted
`verify-ca`; and public MySQL documentation contained the harness maintenance procedure.

**Green — review round 6:**

```text
POLYMETRICS_DATABASE_INTEGRATION=0 go test -count=1 -timeout 20m -tags=databaseintegration \
  ./internal/connectors/native/dbtest \
  ./internal/connectors/native/mysql \
  ./internal/connectors/bundleregistry
```

Passed. The focused check includes the tagged MySQL integration source without opting into a live
container. It covers the non-public CDC projection, default-connection restoration and concurrent
change guard, target-capacity refusal, scoped certificate copy, and compilation of the new
`verify-ca` live case.

**Red — review round 7 endpoint ownership correction:**

Source review confirmed that the removed task-owned-machine lifecycle required an unsupported Podman
5.3 init flag and wrote the global default connection. It also found that cached source images could
skip target capacity proof, shared source-image removal was caller-selectable, and a closed harness
could register itself with interrupt cleanup after `closeOnce` had already run.

**Green — review round 7 endpoint ownership correction:**

```text
go test -count=1 -timeout 5m ./internal/connectors/native/dbtest
```

Passed. The harness now accepts only a direct local Unix endpoint, binds its socket and graph-root
identity before every daemon command, retains every shared source image, and refuses an interrupt
registration after close.

**Red — review round 8 interrupt drain and checkpoint presence:**

Source inspection found that the signal watcher stopped notification before resource removal, and
its one-time live-harness snapshot let a queued start proceed after the slot was released. It also
found that a committed empty checkpoint was indistinguishable from an observed empty cursor, while
the shared cursor adapter rejected empty and whitespace source values. Focused regression coverage
will exercise drain admission, handler lifetime, absent-versus-empty checkpoint state, and MySQL's
strict empty cursor boundary after the complete fix round.

**Green — review round 8:**

```text
go test -count=1 -timeout 20m ./internal/connectors/native/dbtest ./internal/app ./internal/connectors/native/mysql ./internal/synccontract
```

Passed. Interrupt cleanup now remains subscribed through teardown, drains later starts before any
Podman action, and keeps the watcher active until removal completes. Checkpoint envelopes record
whether a source position was observed, so an acknowledged empty run remains non-resumable while
an observed empty or whitespace MySQL cursor is preserved as an opaque strict boundary.

**Red — review round 9 source-native cursor checkpoint:**

Source inspection found that both legacy ETL destination paths converted MySQL `[]byte` cursor
values through JSON/base64 and selected a client-side maximum with Go ordering before committing
the shared checkpoint. The next MySQL read therefore received a different lower-bound value and
could replay, skip, or retain stale source rows. This review round defers execution until all
source, checkpoint, and regression-test changes are in place; the focused Green check will cover
both destination paths and the MySQL wire-boundary codec together.

**Green — review round 9:**

```text
go test -count=1 -timeout 20m -run 'Test(SourceOrderedOpaqueCursorResumesWithoutLossOrReplayAcrossDestinations|OpaqueCursorStatePreservesNativeMySQLBoundaryValues)$' ./internal/app ./internal/connectors/native/mysql
```

Passed. The shared cursor tracker now commits each source-ordered MySQL boundary as an untouched
opaque token and resumes through a typed request state, while the legacy paths retain their prior
comparison behavior. The regression covers binary-state JSON persistence, no reconstructed lower
bound, no replay, no skipped later byte-ordered row, and both warehouse and connector destinations.

**Red — review round 10 source-faithful cursor lifecycle:**

Source inspection found that both ETL destinations still constructed `ReadRequest` independently,
always carried a checkpoint into full-refresh reads, and enabled MySQL's source-ordered tracker
without proving that its stream cursor matched `Config["cursor_field"]`. The warehouse raw WAL also
stored only the report cursor and folded it with generic string ordering, so binary cursor values
could retain an older record even when the opaque resume checkpoint advanced. This review round
will centralize request construction, reject unbound source-ordered cursor fields before `Read`,
and persist opaque source cursor tokens through deduplicated materialization before running one
focused Green command.

**Green — review round 10 source-faithful cursor lifecycle:**

```text
go test -count=1 -timeout 20m -run 'Test(SourceOrderedOpaqueCursorResumesWithoutLossOrReplayAcrossDestinations|SourceOrderedCursorFieldMismatchRefusesBothDestinationPaths|SourceOrderedFullRefreshStartsWithoutResumeAcrossDestinations|SourceOrderedBinaryCursorDedupeRetainsLatestAndResumes|OpaqueCursorStatePreservesNativeMySQLBoundaryValues|SourceOrderedCursorFieldMustMatchMySQLConfiguration|OpaqueCursorComparisonRetainsNativeBinaryOrder)$' ./internal/app ./internal/connectors/native/mysql
```

Passed. Both destination paths now use one mode-aware request constructor, source-ordered readers
reject a stream/config cursor mismatch before reading, and full refreshes never receive a prior
boundary. The warehouse WAL retains opaque source cursor tokens for deduplication; MySQL compares
native binary and numeric tokens directly and uses already-proven source emission order where a
server collation is not portable. The focused regressions cover mismatch refusal, repeated full
refresh overwrite, binary `0x00` to `0xff` deduplication, and resume without replay.

**Red — review round 11 cancellation-safe harness teardown:**

Source inspection found that `Close` passed its caller context directly into every Podman removal
while `sync.Once` permanently consumed the cleanup path. A caller-canceled or expired context could
therefore skip all generated-resource removal and unregister interrupt cleanup before any later
close could retry. The focused regression uses a context-aware fake runner and a canceled caller
context to prove that teardown receives an independent bounded cleanup context.

**Green — review round 11 cancellation-safe harness teardown:**

```text
go test -count=1 -timeout 20m ./internal/connectors/native/dbtest
```

Passed. `Close` now gives the generated-resource cleanup sequence its own three-minute context, so
a canceled caller cannot prevent container, volume, and run-image removal or consume the one
idempotent teardown path.

**Red — test-phase Podman-machine compatibility:**

```text
go test -count=1 -timeout 5m ./internal/connectors/native/dbtest \
  -run '^TestStartUsesPodmanMachineDaemonCapacity$'
```

This failed with `target Podman endpoint returned an invalid identity`: Podman 5.3 returned a
VM-local `unix://` remote socket and a VM image-store path over the supplied direct host Unix
forward, while the harness accepted only a host-path-equivalent socket and host `statfs` capacity.

**Green — test-phase Podman-machine compatibility:**

```text
go test -count=1 -timeout 5m ./internal/connectors/native/dbtest
POLYMETRICS_DATABASE_INTEGRATION=1 POLYMETRICS_PODMAN_ENDPOINT=<direct-local-unix-forward> \
  go test -tags=databaseintegration -count=1 -v -timeout 20m \
  -run '^TestMySQLContainerHarness$' ./internal/connectors/native/mysql
```

Both passed. The focused unit proof accepts only a reported safe Unix forward with numeric daemon
store capacity and rejects malformed forwarded capacity before any pull, tag, volume, or container
mutation. The real MySQL proof then completed check, catalog discovery, five-record full and
incremental reads, all TLS modes including `verify-ca`, and internal insert/update/delete CDC.

**Red — CI Snyk verify-ca resumption remediation:**

```text
go test -count=1 -timeout 5m ./internal/connectors/native/sqltls \
  -run 'TestVerifyCA(RejectsAChainOutsideTheConfiguredRoot|AcceptsTheConfiguredRootIgnoringHostname)$'
```

This failed before the implementation because `TLSConfig` set only
`VerifyPeerCertificate`; both tests reported `verify-ca installed no connection verifier`. Go
documents that callback as skipped on resumed connections when built-in certificate verification is
disabled, so its manual chain check did not cover every verify-ca handshake.

**Green — CI Snyk verify-ca resumption remediation:**

```text
go test -count=1 -timeout 5m ./internal/connectors/native/sqltls
go test -count=1 -timeout 20m ./internal/connectors/native/sqltls \
  ./internal/connectors/native/mysql \
  ./internal/connectors/native/postgres
go vet ./internal/connectors/native/sqltls \
  ./internal/connectors/native/mysql \
  ./internal/connectors/native/postgres
govulncheck ./internal/connectors/native/sqltls \
  ./internal/connectors/native/mysql \
  ./internal/connectors/native/postgres
```

All passed. `verify-ca` now uses `VerifyConnection`, which Go invokes for every connection,
including resumptions, while preserving its documented chain-only (no-hostname) verification.
