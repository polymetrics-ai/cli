# TDD ledger — MySQL container harness R1

| ID | Guarantee | Red assertion | Green proof |
| --- | --- | --- | --- |
| H1 | Scoped execution | Missing integration opt-in or `DOCKER_CONTEXT` cannot create a live pass. | Tagged test visibly skips before startup with its reason; the passing live run passed `colima` to every Docker child. |
| H2 | Resource isolation | A run can use a predictable/shared container or volume name. | Generated names are unique and every command targets only those names. |
| H3 | Cleanup | An assertion failure or interrupt can strand a container, volume, or pulled image. | Idempotent cleanup order and engine-argument ordering are unit-tested; the live test's deferred cleanup removed resources, reset Colima, and reported free disk before/after. |
| H4 | Image ownership | Cleanup deletes an image a different lane already had. | The harness removes only an image absent before its own pull, unless keep-image is explicitly enabled. |
| H5 | Non-default endpoint | The connector silently assumes the engine default host port. | Parsed Docker mapping is loopback-only and rejects the default port before config is built. |
| H6 | Colima host-disk reclamation | Docker resource cleanup leaves the Colima VM disk file growing indefinitely on the host. | The documented run opted into `colima delete` then `colima start` only after container/volume/image cleanup. Its live report was before=84802707456 and after=84784943104 bytes (17.8 MB, inside the 128 MB ordinary-build allowance), with `colima_reset=true`. |
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
focused green equivalent passed after implementation. The final tagged Docker/Colima proof passed
in 53.48 seconds against MySQL 8.4.11 and asserted real check, catalog, paged full read,
incremental read, and insert/update/delete binlog events.
