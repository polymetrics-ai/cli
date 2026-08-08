# TDD ledger — MySQL container harness R1

| ID | Guarantee | Red assertion | Green proof |
| --- | --- | --- | --- |
| H1 | Scoped execution | Missing integration opt-in or `POLYMETRICS_PODMAN_CONNECTION` cannot create a live pass. | Tagged test visibly skips before startup with its reason; `TestNewRejectsUnscopedConnection` refuses empty, whitespace and argument-injecting connections; the passing live run passed `--connection fm-cli-db-harness-r1` to every Podman child. |
| H2 | Resource isolation | A run can use a predictable/shared container or volume name. | Generated names carry a crypto-random suffix; every command targets only those names. Engines are sequential through a bounded slot, released on every exit path (`TestStartReleasesTheEngineSlotOnEveryExitPath`). |
| H3 | Cleanup | An assertion failure or interrupt can strand a container, volume, or pulled image. | `TestCloseDuringCreateStillRemovesTheCreatedResource` fires Close from inside a create and asserts nothing is left behind; `TestStartRefusesToCreateAfterClose` proves no create follows cleanup; `-race` clean. Two live failing runs (before the reserved-word fixes) left zero containers, volumes or images. |
| H4 | Image ownership | Cleanup deletes an image reference this run did not create. | Ownership is claimed from this run's own successful pull and its run-specific generated tag, never inferred from an earlier inspect. `TestCleanupKeepsSourceImageOnlyWhenOptedIn` covers the keep-image opt-in. |
| H5 | Non-default endpoint | The connector silently assumes the engine default host port. | `TestStartPublishesOnlyLoopback` and `TestParseMappedPortRefusesTheEngineDefaultPort`: the publish is `127.0.0.1::<port>` and a mapping back to the engine default is rejected before config is built. |
| H6 | Host-disk reclamation | Removing an image frees space inside the VM but leaves the host's sparse disk file inflated. | Measured: one `fstrim` returned 241 of ~950 MiB; a second returned effectively all of it (machine `.raw` 2609.0 -> 3563.4 -> 2609.4 MiB). `TestCleanupReclaimsHostDiskOnlyAfterContainerCleanup` asserts **two** passes, after container cleanup. A full live run moved the `.raw` by +0.2 MiB. |
| M1 | Honest declaration | MySQL can claim logical replication or public CDC without executable binlog code. | `binlog_replication` is closed-schema valid; bundle/executor descriptors match and public CDC is true only for the executable connector. |
| M2 | SQL safety | Stream/schema/cursor strings can be concatenated unvalidated. | Unit tests reject unsafe identifiers and reads use only quoted validated identifiers plus parameter values. |
| M3 | Paging/incremental | Reads only prove one record or ignore state. | Live seed/read results prove multiple request pages and exact cursor filtering. |
| M4 | Change capture | A binlog event succeeds without emitting/asserting a real row change or safely committing state. | Real insert/update/delete events are decoded, delivered, and their file/position state commits only after acknowledgement. |

## Red evidence

```text
go test -count=1 ./internal/connectors ./internal/connectors/native/dbtest \
  -run 'Test(ChangefeedDescriptorAcceptsBinlogReplication|NewRejectsUnscopedDockerContext|Cleanup)'

mysql_binlog_changefeed_test.go: binlog_replication was rejected as an unsupported mechanism
harness_test.go: New and Config were undefined because no harness existed
```

The red command failed as intended before the vocabulary and harness implementation existed. Its
focused green equivalent passed after implementation. The final tagged Podman proof passed
in 53.48 seconds against MySQL 8.4.11 and asserted real check, catalog, paged full read,
incremental read, and insert/update/delete binlog events.
| T1 | Transport security is enforced, not declared | A mode can be declared and silently ignored, or a strict mode can quietly downgrade to plaintext. | `TestStrictTransportSecurityIsNotDowngradedAgainstATLSLessServer` dials a hand-written TLS-less MySQL handshake: `required`/`verify-ca`/`verify-identity` all refuse, `preferred`/`disabled` connect. `TestVerifyCARejectsAChainOutsideTheConfiguredRoot` proves chain verification is real. Live subtests read the server's own `Ssl_cipher` for every mode. |
| T2 | No cross-connector drift | MySQL and PostgreSQL grow two spellings of one choice. | One `sqltls` vocabulary; `TestSSLModeAcceptsTheSharedCanonicalVocabulary` pins that existing libpq values keep byte-identical behaviour while canonical names are also accepted. |
| T3 | PostgreSQL configuration is executable | The shared TLS fields parse in MySQL but PostgreSQL's definition rejects them, or Catalog/Read ignore the requested verified server name. | **Red:** `TestDefinitionAcceptsSharedTransportSecurityVocabulary` rejected `disabled` at the embedded definition boundary. **Green:** the PostgreSQL schema accepts each runtime spelling; `TestPostgresPoolConfigUsesSharedTransportSecurityOptions` proves `sslrootcert` reaches pgx, `sslservername` becomes TLS SNI/verification name, strict verification has no plaintext fallback, and Check/Catalog/Read use the same `openPool` path. |
| C1 | Catalog never over-promises | One unquotable table name makes the whole database undiscoverable, or discovery advertises a stream every read would reject. | `TestCatalogSkipsUnreadableIdentifiersWithoutFailingDiscovery` keeps the readable tables and re-parses every advertised stream through the reader's own `qualifyStream`. |
| P1 | Release portability | `syscall.Statfs` is Unix-only and breaks the Windows release build. | Split by build tag; `GOOS=windows go build ./...` passes. |
