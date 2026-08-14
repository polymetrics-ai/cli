# TDD ledger — Issue #4086 mechanical fence split

This is a move-only task: no test body or assertion may be edited. The red/green
evidence therefore checks the layout contract rather than inventing a behavior
test or accepting a behavior change.

| ID | Fence | Red: base layout proof | Green: original move head proof |
| --- | --- | --- |
| R1 | Source ownership | The base has source checks in shared `postgres_test.go`; `source_test.go` is absent. | Every source check is moved unchanged to `source_test.go`. |
| R2 | Capability ownership | The base has metadata/manifest/docs/interface checks in shared `postgres_test.go`; a dedicated capability file is absent. | Those checks are byte-identical in `capability_surface_test.go`. |
| R3 | CDC ownership | The base keeps one fail-closed CDC check in the shared file. | The check is byte-identical in `cdc_capability_fence_test.go`; no assertion changes. |
| R4 | Database lane ownership | The base's mapping, source-plan, and target-admission declarations share `database_test.go`. | Each declaration appears exactly once in its lane-owned file, with helpers unchanged. |
| R5 | Runtime equivalence | Head binary does not exist yet, so base artifacts cannot be compared. | All scoped command stdout/stderr files compare byte-identically. |
| R6 | Generated parity | Head generated-capability hash does not exist yet. | The generated ledger SHA-256 is identical and `surface-sync --check` passes. |
| R7 | Registration-order inertness | Each affected package passes deterministic Go shuffled runs at `5a457970b3bc15343e5ba6b7b4acf48994b63add` with seeds `408601` and `408602`. | The same two seeded shuffled runs pass at original move head `6e31ac1abfc4a46fd1dbbef3ec54086da85b682e`. |

## Baseline capture

The baseline binary was built directly from
`5a457970b3bc15343e5ba6b7b4acf48994b63add` before edits. It captured stdout and
stderr for `pm help connectors`, `pm connectors --help`,
`pm connectors inspect postgres --json`, `pm postgres --help`, `pm postgres check`,
and `pm postgres catalog` in the transient, ignored `.tmp/issue-4086-base.*`
directory.

## Green evidence

- All six stdout/stderr pairs above compare byte-for-byte with the head binary.
- `internal/connectors/defs/operation_endpoint_ledger.json` is byte-identical:
  base and head SHA-256 are both
  `3b3bfe208a4e5600ef9cdf5a7440267692dc07dc8bd0c1e69cc43536f607ad66`.
- `go run ./cmd/connectorgen surface-sync --check` scanned 552 connectors with
  zero fills and zero corrections.
- The before/after declaration-name inventories are identical for both split
  monoliths, and the production/definition changed-path check is empty.

## Registration-order green evidence

The proof uses the deterministic numeric form of Go's `-shuffle=on` facility:
`go test -count=1 -timeout 20m -shuffle=<seed> <package>`.

| Revision | Seed | `./internal/connectors/database` | `./internal/connectors/native/postgres` |
| --- | --- | --- | --- |
| Base `5a457970b3bc15343e5ba6b7b4acf48994b63add` | `408601` | `ok 6.452s` | `ok 0.881s` |
| Base `5a457970b3bc15343e5ba6b7b4acf48994b63add` | `408602` | `ok 6.229s` | `ok 0.878s` |
| Original move head `6e31ac1abfc4a46fd1dbbef3ec54086da85b682e` | `408601` | `ok 6.416s` | `ok 0.875s` |
| Original move head `6e31ac1abfc4a46fd1dbbef3ec54086da85b682e` | `408602` | `ok 6.652s` | `ok 0.870s` |
