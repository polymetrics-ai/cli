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
| M1 | Honest CDC declaration | MySQL can advertise a CDC capability before a production runtime entrypoint exists. | Metadata, definition, catalog, generated docs, and website data keep `cdc: false`; no changefeed descriptor or executor is registered, while the native row-event reader remains covered by internal evidence. |
| M2 | SQL safety | Stream/schema/cursor values are concatenated into SQL. | Identifier tests reject unsafe values; queries quote only validated identifiers and bind all values. |
| M3 | Complete ETL paging and incrementals | A read proves a single record or repeats/skips at shared cursor values. | Keyset query tests cover primary-key and `(cursor, primary_key)` boundaries; the real five-row seed with `page_size=2` proves multiple pages and exact incremental rows. |
| M4 | Change capture state | Binlog rows can advance state before acknowledgement or be projected against changed metadata. | CDC tests bind checkpoint to ordered column fingerprint and position/row ordinal; live insert/update/delete events assert actual returned rows. |
| M5 | #3902 direct-read boundary | A database ETL reader is mistaken for a pagewise HTTP direct-read command and omits page context. | `TestReadIsETLNotPagewiseDirectRead` proves MySQL exposes neither `DirectReader` nor `OperationDirectReader`; the SQL `Read` drains its bounded pages to ETL rather than returning an exploratory page. |
| T1 | TLS is enforced, not declared | A strict TLS mode can quietly downgrade or be ignored by replication. | TLS-less server tests make strict modes refuse; certificate tests prove verification; live checks use the server's own `Ssl_cipher`; the binlog syncer receives the same TLS config. |
| T2 | SQL option shape does not drift | MySQL and PostgreSQL expose incompatible TLS field names or validation. | `sqltls` centralizes the vocabulary. PostgreSQL's definition accepts the runtime vocabulary and `TestPostgresPoolConfigUsesSharedTransportSecurityOptions` proves CA/server-name application and no strict plaintext fallback. |
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
