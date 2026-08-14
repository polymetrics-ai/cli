## Intent

Mechanically split Issue #4086's shared PostgreSQL database tests and connector
capability fence from comparison base `integration/4015-mvp-flat-r1` at
`5a457970b3bc15343e5ba6b7b4acf48994b63add`. This is file layout only.

## Linked Issue

Closes #4086

## Lane ownership after this PR

| Lane | Files |
| --- | --- |
| Source | `internal/connectors/native/postgres/source_test.go`; `internal/connectors/database/source_read_plan_test.go` |
| Target | `internal/connectors/database/target_admission_test.go` |
| Mapping | `internal/connectors/database/mapping_definition_test.go` |
| CDC | `internal/connectors/native/postgres/cdc_capability_fence_test.go` |

`internal/connectors/native/postgres/capability_surface_test.go` is deliberately
the stable capability fence, not an execution-lane file. Test-only shared
scaffolding is neutral in `internal/connectors/database/test_helpers_test.go`.

## Declaration-level move inventory

Every item in each row moved unchanged from the named old file to the named new
file; no declaration was renamed or modified.

| Old file | New file | Moved declarations |
| --- | --- | --- |
| `internal/connectors/native/postgres/postgres_test.go` | `internal/connectors/native/postgres/capability_surface_test.go` | `TestNameAndMetadata`; `TestManifestProjectsBundleCredentials`; `TestGeneratedDocsDescribeAuthenticationRequirements`; `TestCertificationArchitectureUsesExecutablePostgresResumeBindings`; `TestNoInitRegistration`; `TestConnectorSatisfiesCoreInterfaces`; `TestWriteUnsupported` |
| `internal/connectors/native/postgres/postgres_test.go` | `internal/connectors/native/postgres/source_test.go` | `fixtureConfig`; `TestCheckFixtureModeOK`; `TestCheckRejectsCtxCancelled`; `TestCatalogFixtureMode`; `TestReadFixtureEmitsRows`; `TestReadFixtureIncrementalCursor`; `TestReadUnknownFixtureStream`; `TestReadRequiresStream`; `TestInitialStateStatefulReader`; `TestCheckConfigValidationTable` |
| `internal/connectors/native/postgres/postgres_test.go` | `internal/connectors/native/postgres/cdc_capability_fence_test.go` | `TestCDCIsFailClosedUntilStreamedStagingExists` |
| `internal/connectors/database/database_test.go` | `internal/connectors/database/mapping_definition_test.go` | `TestDatabaseDefinitionStrictLoadAndDefensiveProjection`; `TestDatabaseDefinitionRejectsAmbiguousMembers`; `TestDatabaseDefinitionEnforcesSchemaNumericConstraints`; `TestResourcePolicyBoundsEveryDatabaseResource`; `TestLogicalTypeCompatibilityIsLosslessOrRejected` |
| `internal/connectors/database/database_test.go` | `internal/connectors/database/source_read_plan_test.go` | `TestStructuredCatalogIdentityAndReadPlanAreStable`; `TestDatabaseLoadAndReadPlanHonorCancellation`; `TestDatabaseLoadChecksCancellationBeforeReturningProjection`; `TestReadPlanChecksCancellationBeforeReturningProjection` |
| `internal/connectors/database/database_test.go` | `internal/connectors/database/target_admission_test.go` | `TestDriverAdmissionRequiresRegisteredCompatibleNativeAdmission`; `TestDatabaseNativeAdmissionIsBoundToOneWarehouseLeg`; `TestDriverRegistryRejectsSharedNativeContractAcrossWarehouseLegs`; `TestDriverRegistryRejectsCrossDriverNativeContractReuse`; `TestWarehouseMediationUsesSharedArtifactAndSeparateDatabaseLegs`; `TestMySQLLayerTwoReferenceCompilesAgainstSharedWarehouseArtifact` |
| `internal/connectors/database/database_test.go` | `internal/connectors/database/test_helpers_test.go` | `mysqlLayerTwo`; `mysqlLayerTwo.DatabaseDriverDescriptor`; `mysqlLayerTwo.DatabaseNativeAdmissions`; `loadTestDefinition`; `testReadPlanRequest`; `definitionWithAdmittedMode`; `testInboundCommand`; `testOutboundCommand`; `declaredDriver`; `declaredDriver.DatabaseDriverDescriptor`; `admittedDriver`; `admittedDriver.DatabaseNativeAdmissions`; `nativeAdmission`; `nativeContract`; `nativeAdmissionFor`; `testInboundNativeAdmission`; `testOutboundNativeAdmission`; `nativeAdmission.NativeSyncExecutorDescriptor`; `nativeAdmission.NativeSyncConformanceEvidence`; `cloneDatabaseNativeAdmissions`; `cancelOnErrCallContext`; `cancelOnErrCallContext.Deadline`; `cancelOnErrCallContext.Done`; `cancelOnErrCallContext.Err`; `cancelOnErrCallContext.Value`; `postgresDriverDescriptor`; `mysqlDefinitionJSON`; `validDefinitionJSON` |

## Behavior and generated-capability parity

Built `pm` binaries at the stated base and head. The stdout and stderr were
byte-identical for `pm help connectors`, `pm connectors --help`,
`pm connectors inspect postgres --json`, `pm postgres --help`, `pm postgres check`,
and `pm postgres catalog`.

Generated capability output is unchanged:

```text
3b3bfe208a4e5600ef9cdf5a7440267692dc07dc8bd0c1e69cc43536f607ad66  operation_endpoint_ledger.json (base)
3b3bfe208a4e5600ef9cdf5a7440267692dc07dc8bd0c1e69cc43536f607ad66  operation_endpoint_ledger.json (head)
```

`go run ./cmd/connectorgen surface-sync --check` scanned 552 connectors with
zero fills and corrections. No generator ran in write mode and no generated
artifact changed.

## Testing and delivery evidence

- Focused and complete `internal/connectors/database` tests passed.
- Focused and complete `internal/connectors/native/postgres` tests passed.
- `go vet ./internal/connectors/database ./internal/connectors/native/postgres` passed.
- `go test -count=1 -timeout 20m ./internal/cli` and `go build ./cmd/pm` passed.
- Required skills: `golang-how-to`, `golang-design-patterns`,
  `golang-structs-interfaces`, `golang-error-handling`, `golang-security`,
  `golang-safety`, `golang-testing`, and `golang-database`.
- GSD commands were resolved and executed with the documented inline/manual
  fallback; evidence is in `.planning/phases/issue-4086-postgres-fence-split-r1/`.

No behavior-adjacent defect was found. No PostgreSQL literal moved into shared
foundation code, and no shared declaration moved into the connector bundle.
