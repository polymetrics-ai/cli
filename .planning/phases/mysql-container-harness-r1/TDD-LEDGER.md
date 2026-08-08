# TDD ledger — MySQL container harness R1

| ID | Guarantee | Red assertion | Green proof |
| --- | --- | --- | --- |
| H1 | Scoped execution | A missing or unsafe Podman connection can fall through to the global default. | `TestNewRejectsUnscopedConnection` rejects empty, whitespace, and argument-injecting names; every runner invocation receives the supplied connection. |
| H2 | Isolation and sequencing | Runs can collide on a resource name or multiply database peak disk/memory. | Random generated container/volume/tag names are unit-tested; `TestStartReleasesTheEngineSlotOnEveryExitPath` proves the default one-slot semaphore returns on success and failure. |
| H3 | Unconditional cleanup | Assertion failure or interrupt can leave a container, volume, or image behind. | `TestCloseDuringCreateStillRemovesTheCreatedResource` and `TestStartRefusesToCreateAfterClose` cover the create/interrupt race; `Close` continues all cleanup stages and the live post-run scoped listings were empty. |
| H4 | Image ownership | Cleanup guesses whether a shared image existed and deletes another run's reference. | A successful pull is immediately re-tagged to a run-specific local reference before startup; `TestStartUsesGeneratedImageReferenceWithoutInspectingSource` and `TestCleanupKeepsSourceImageOnlyWhenOptedIn` cover ownership and retention. |
| H5 | Non-default endpoint | A connector silently assumes 3306 on the host. | `TestStartPublishesOnlyLoopback` asserts `127.0.0.1::<port>` publishing; `TestParseMappedPortRefusesTheEngineDefaultPort` refuses a default host mapping. |
| H6 | Host-disk reclamation | Removing an image returns space only inside the VM. | `TestCleanupReclaimsHostDiskOnlyAfterContainerCleanup` asserts two explicit `fstrim` passes after cleanup; the live test records before/after byte counts and fails if the reclaimed run exceeds ordinary build noise. |
| M1 | Honest CDC declaration | MySQL can advertise generic/logical CDC without an executable binlog reader. | Closed-schema validation accepts only `binlog_replication`; descriptor/executor tests and the live row-event proof earn `cdc: true`. |
| M2 | SQL safety | Stream/schema/cursor values are concatenated into SQL. | Identifier tests reject unsafe values; queries quote only validated identifiers and bind all values. |
| M3 | Complete ETL paging and incrementals | A read proves a single record or repeats/skips at shared cursor values. | Keyset query tests cover primary-key and `(cursor, primary_key)` boundaries; the real five-row seed with `page_size=2` proves multiple pages and exact incremental rows. |
| M4 | Change capture state | Binlog rows can advance state before acknowledgement or be projected against changed metadata. | CDC tests bind checkpoint to ordered column fingerprint and position/row ordinal; live insert/update/delete events assert actual returned rows. |
| M5 | #3902 direct-read boundary | A database ETL reader is mistaken for a pagewise HTTP direct-read command and omits page context. | `TestReadIsETLNotPagewiseDirectRead` proves MySQL exposes neither `DirectReader` nor `OperationDirectReader`; the SQL `Read` drains its bounded pages to ETL rather than returning an exploratory page. |
| T1 | TLS is enforced, not declared | A strict TLS mode can quietly downgrade or be ignored by replication. | TLS-less server tests make strict modes refuse; certificate tests prove verification; live checks use the server's own `Ssl_cipher`; the binlog syncer receives the same TLS config. |
| T2 | SQL option shape does not drift | MySQL and PostgreSQL expose incompatible TLS field names or validation. | `sqltls` centralizes the vocabulary. PostgreSQL's definition accepts the runtime vocabulary and `TestPostgresPoolConfigUsesSharedTransportSecurityOptions` proves CA/server-name application and no strict plaintext fallback. |
| G1 | Production native wiring contains only connectors | A test harness or shared helper is blank-imported solely because it sits below `native/`. | `TestGen_NativesetImportsRuntimePackagesAndExcludesSupportLibraries` fails for `dbtest`/`sqltls`; `nativeSupportPackages` and regenerated wiring exclude both while retaining real connector packages. |

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
