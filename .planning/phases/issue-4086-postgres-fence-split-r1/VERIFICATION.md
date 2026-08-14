# Verification checklist — Issue #4086

- [x] Declaration-level move inventory records old and new file for every moved declaration.
- [x] No production Go or definition JSON path changed.
- [x] Base/head `pm` output comparison is byte-identical on scoped PostgreSQL and database commands.
- [x] Generated capability artifact SHA-256 values are identical; `surface-sync --check` passes.
- [x] `go test -count=1 -timeout 20m ./internal/connectors/database` passes.
- [x] `go test -count=1 -timeout 20m ./internal/connectors/native/postgres` passes.
- [x] Complete package tests for every touched package pass.
- [x] Inline code review finds no behavior/surface change.

## Commands passed

```sh
go test -count=1 -timeout 20m ./internal/connectors/database -run 'Test(DatabaseDefinition|ResourcePolicy|LogicalType|StructuredCatalog|DatabaseLoad|ReadPlan|DriverAdmission|DatabaseNativeAdmission|DriverRegistry|WarehouseMediation|MySQLLayerTwo)'
go test -count=1 -timeout 20m ./internal/connectors/native/postgres -run 'Test(NameAndMetadata|ManifestProjectsBundleCredentials|GeneratedDocsDescribeAuthenticationRequirements|CertificationArchitectureUsesExecutablePostgresResumeBindings|NoInitRegistration|ConnectorSatisfiesCoreInterfaces|Check|Catalog|Read|InitialState|CDCIsFailClosedUntilStreamedStagingExists|WriteUnsupported)'
go vet ./internal/connectors/database ./internal/connectors/native/postgres
go test -count=1 -timeout 20m ./internal/connectors/database
go test -count=1 -timeout 20m ./internal/connectors/native/postgres
go test -count=1 -timeout 20m ./internal/cli
go build ./cmd/pm
go run ./cmd/connectorgen surface-sync --check
```
