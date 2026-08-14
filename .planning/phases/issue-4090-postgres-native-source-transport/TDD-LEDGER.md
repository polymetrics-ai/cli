# TDD ledger — Issue #4090

| ID | RED failure | GREEN contract |
| --- | --- | --- |
| R1 | PostgreSQL has no declared transport source or preflight rejects it. | `Definition()` declares exactly the PostgreSQL `native_database` source and preflight resolves it. |
| R2 | A wrong-family definition is able to reach source I/O. | Family mismatch rejects before the instrumented source can open/query. |
| R3 | A declared but unregistered executor is able to reach source I/O. | Exact missing registration rejects before source invocation. |
| R4 | No descriptor reaches an implicit `Read()` fallback. | Absence rejects before source invocation; no capability bit substitutes for it. |
| R5 | Full modes can emit an unbounded, unordered, loosely typed record set. | Both full modes emit pages bounded by request and definition caps, with catalog-derived projection and stable ordering. |
| R6 | Output identity/checkpoint can vary or omit schema/source binding. | A valid candidate carries the exact source identity, typed schema fingerprint/barrier, generation, and deterministic dedupe boundary. |
| R7 | A fake proves only exit status. | A real PostgreSQL 16.10 dbtest logs the emitted rows plus identity/schema/checkpoint values. |
| R8 | A driver-native UUID cursor or temporal infinity reaches page two as an unencodable value. | The typed-catalog cursor normalizer emits pgx-encodable UUID/date/timestamp values and rejects an unsupported key kind before querying. |
| R9 | The live proof only covers happy-path scalar rows. | PostgreSQL 16.10 dbtest asserts emitted zone-less timestamp, UUID, exact JSON number, JSON null, `-infinity` timestamp key, and UUID-key page-two values. |

## RED command

```sh
go test -timeout 20m -count=1 ./internal/connectors/native/postgres -run 'TestPostgres.*Transport'
```

The exact RED output is retained in `traces/source-transport-red.txt` before
the production executor is added.

**Captured RED:** `go test -timeout 20m -count=1
./internal/connectors/native/postgres -run 'TestPostgres.*Transport'` failed
with `PostgreSQL definition has no declared source transport`.

**Captured RED (registration):** the same focused package test then failed
to build with `undefined: RegisterSnapshotTransportSource`, proving that a
declaration without the PostgreSQL-owned registry adapter cannot resolve a
source executor.

**Captured RED (cursor normalizer):**
`TestPostgresSnapshotStableKeyPaginationValuesAreTypedOrRefused` failed with a
raw `[16]byte` UUID cursor and with no refusal for a JSON stable key. This
proved that the previous infinity-only conversion still delegated a page-two
failure to pgx rather than refusing at the typed pagination boundary.

## GREEN evidence

- `Connector.Definition()` now adds the exact `native_database` descriptor;
  `RegisterSnapshotTransportSource` is a PostgreSQL-local explicit adapter.
- `transport_source.go` opens a read-only repeatable-read transaction only
  after request validation, discovers the existing typed catalog in that same
  transaction, and renders a finite primary/unique-key ordered page query.
- `TestPostgresTransportRegistryPreflightRefusesBeforeSourceIO` separately
  proves missing descriptor, wrong family, and unregistered executor errors
  with a source whose every I/O method increments a counter.
- `TestPostgresSnapshotReadPlanAndCheckpointUseTypedStableIdentity` proves
  selected catalog order, finite parameterized page shape, schema fingerprint,
  source identity, full-snapshot barrier, and deterministic dedupe boundary.
- `TestPostgresSnapshotStableKeyPaginationValuesAreTypedOrRefused` proves a
  raw UUID stable key becomes `pgtype.UUID`, a negative timestamp infinity
  becomes `pgtype.Timestamp`, and a JSON key is refused with its logical kind
  before a query is constructed.
- `TestPostgresDynamicTypedCatalogUsesLiveMetadata` passes via Docker against
  PostgreSQL 16.10 for both `full_append` and `full_overwrite`, then pages a
  `-infinity` timestamp key and a UUID key at batch size one. It asserts the
  emitted civil timestamp, UUID, exact JSON number, JSON null, and edge-key
  values; the real output is retained in `traces/live-source-green.txt`.

## Review regression slice

**Red:** `TestPostgresSnapshotCheckpointDoesNotRequireJSONEncodableKeyValues`
failed because a valid PostgreSQL floating `NaN` stable-key value reached
`json: unsupported value: NaN` while constructing a checkpoint dedupe key.
`TestPostgresSnapshotRelationRefAllowsTypedCatalogDatabaseName` also exposed
that an identity-only catalog name accepted by `database.Catalog` (`analytics-db`)
was incorrectly treated as a SQL identifier.

**Green:** checkpoint dedupe now uses the repeatable-read snapshot, typed schema,
source identity, and bounded page ordinal rather than serializing provider key
values. The catalog component is compared only as identity; schema and relation
remain strictly validated before they are quoted into the PostgreSQL query.
