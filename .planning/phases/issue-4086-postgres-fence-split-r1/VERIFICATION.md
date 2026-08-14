# Verification checklist — Issue #4086

- [x] Declaration-level move inventory records old and new file for every moved declaration.
- [x] No production Go or definition JSON path changed.
- [x] Base/head `pm` output comparison is byte-identical on scoped PostgreSQL and database commands.
- [x] Generated capability artifact SHA-256 values are identical; `surface-sync --check` passes.
- [x] `go test -count=1 -timeout 20m ./internal/connectors/database` passes.
- [x] `go test -count=1 -timeout 20m ./internal/connectors/native/postgres` passes.
- [x] Registration-order changes are inert under Go's `-shuffle=on` facility,
  reproduced with explicit seeds `408601` and `408602` for both packages at
  comparison base `5a457970b3bc15343e5ba6b7b4acf48994b63add` and original move
  head `6e31ac1abfc4a46fd1dbbef3ec54086da85b682e`.
- [x] Complete package tests for every touched package pass.
- [x] Inline code review finds no behavior/surface change.

## Registration-order proof

The commands use Go's deterministic numeric shuffled mode:
`go test -count=1 -timeout 20m -shuffle=<seed> <package>`.

| Revision | Seed | `./internal/connectors/database` | `./internal/connectors/native/postgres` |
| --- | --- | --- | --- |
| Base `5a457970b3bc15343e5ba6b7b4acf48994b63add` | `408601` | `ok 6.452s` | `ok 0.881s` |
| Base `5a457970b3bc15343e5ba6b7b4acf48994b63add` | `408602` | `ok 6.229s` | `ok 0.878s` |
| Original move head `6e31ac1abfc4a46fd1dbbef3ec54086da85b682e` | `408601` | `ok 6.416s` | `ok 0.875s` |
| Original move head `6e31ac1abfc4a46fd1dbbef3ec54086da85b682e` | `408602` | `ok 6.652s` | `ok 0.870s` |

## Commands passed

```sh
go test -count=1 -timeout 20m -shuffle=408601 ./internal/connectors/database
go test -count=1 -timeout 20m -shuffle=408601 ./internal/connectors/native/postgres
go test -count=1 -timeout 20m -shuffle=408602 ./internal/connectors/database
go test -count=1 -timeout 20m -shuffle=408602 ./internal/connectors/native/postgres
go test -count=1 -timeout 20m ./internal/connectors/database -run 'Test(DatabaseDefinition|ResourcePolicy|LogicalType|StructuredCatalog|DatabaseLoad|ReadPlan|DriverAdmission|DatabaseNativeAdmission|DriverRegistry|WarehouseMediation|MySQLLayerTwo)'
go test -count=1 -timeout 20m ./internal/connectors/native/postgres -run 'Test(NameAndMetadata|ManifestProjectsBundleCredentials|GeneratedDocsDescribeAuthenticationRequirements|CertificationArchitectureUsesExecutablePostgresResumeBindings|NoInitRegistration|ConnectorSatisfiesCoreInterfaces|Check|Catalog|Read|InitialState|CDCIsFailClosedUntilStreamedStagingExists|WriteUnsupported)'
go vet ./internal/connectors/database ./internal/connectors/native/postgres
go test -count=1 -timeout 20m ./internal/connectors/database
go test -count=1 -timeout 20m ./internal/connectors/native/postgres
go test -count=1 -timeout 20m ./internal/cli
go build ./cmd/pm
go run ./cmd/connectorgen surface-sync --check
```
