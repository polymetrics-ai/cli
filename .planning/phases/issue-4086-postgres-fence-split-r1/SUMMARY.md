# Summary — Issue #4086 PostgreSQL fence split

The shared PostgreSQL/database test monoliths are split into independent lane
files without changing their declarations or bodies.

- Source: `native/postgres/source_test.go` and `database/source_read_plan_test.go`.
- Target: `database/target_admission_test.go`.
- Mapping: `database/mapping_definition_test.go`.
- CDC: `native/postgres/cdc_capability_fence_test.go`.
- Stable capability tests: `native/postgres/capability_surface_test.go`.

Base/head `pm` output, the generated capability ledger, focused/full tests,
scoped vet, CLI regression package, and inline code review are green. No
generator was run in write mode and no generated file changed.
