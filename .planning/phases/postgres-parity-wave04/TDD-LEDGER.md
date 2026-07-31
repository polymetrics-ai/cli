# TDD Ledger — PostgreSQL parity wave 04 r1

| Slice | Test/evidence | Initial status | Final status |
| --- | --- | --- | --- |
| A1 metadata | `go test ./internal/connectors/native/postgres -run TestPostgresDefinitionIncludesBoundedWriteActions` | red: metadata still read-only/no writes | green: manifest/metadata expose five bounded writes |
| A2 validation | `go test ./internal/connectors/native/postgres -run TestPostgresWriteValidationRejectsArbitrarySQLAndUnsafeIdentifiers` | red: `ValidateWrite` missing | green: unknown SQL/unsafe identifiers/missing keys rejected |
| A3 SQL build | `go test ./internal/connectors/native/postgres -run TestPostgresWriteBuildsParameterizedSQL` | red: `buildWriteStatement` missing | green: templates use placeholders, no value inlining |
| A4 fixture write | `go test ./internal/connectors/native/postgres -run TestPostgresFixtureWriteDryRunAndExecute` | red: `DryRunWrite` missing | green: fixture dry-run/write count works with no DB |
| A5 read limit | `go test ./internal/connectors/native/postgres -run TestPostgresReadUsesRequestLimitBeforeConfiguredLimit` | red run blocked by writer compile failures; test in place | green: request limit applied before config read_limit |
| B1 bundle schemas | connectorgen/static conformance for postgres | planned | green: `connectorgen validate` and `TestConformance/postgres` pass |
| C1 CLI surface | focused CLI help/write-plan test for postgres | planned | green: postgres command help and write plan pass |
| D1 local gates | `go build ./cmd/pm`, `make connector-boundary`, `make verify`, `git diff --check` | planned | green: final `make verify` and `git diff --check` pass |
| E1 typed MERGE correction | `go test ./internal/connectors/native/postgres -run TestPostgresWriteBuildsParameterizedSQL` | red/review: `upsert_row` used `INSERT ... ON CONFLICT` while ledger claims `MERGE` | green: `upsert_row` fixture/test now use parameterized `MERGE`; focused/native/conformance/verify gates pass |

Notes:

- No tests may require a live PostgreSQL service.
- Fixture mode may validate schemas and counts but must not claim live certification.
