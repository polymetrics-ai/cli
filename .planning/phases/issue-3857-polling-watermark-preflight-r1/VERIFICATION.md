# Verification — #3857 polling-watermark preflight

## Required checks

| Area | Command/evidence | Status |
| --- | --- | --- |
| TDD red/green | focused polling-preflight test before/after implementation | pending |
| Immutable corpus | focused no-skip conformance runner | pending |
| Database definition/schema | focused `internal/connectors/database` tests | pending |
| Engine runtime | focused `internal/connectors/engine` tests | pending |
| App regression | `go test -timeout 20m ./internal/app` | pending |
| Formatter | `gofmt -l` on changed Go files | pending |
| Individual repository gates | tidy, lint, docs, smoke, agent-contract, connectorgen validation/surface sync, connector boundary, release workflow | pending |
| CLI parity | `pm help <topic>`, bare namespace, command help, inspect; expected no public surface change | pending/not applicable until confirmed |
| GSD verification/review | generated prompts executed inline and review record written | pending |

## Safety and scope checks

- No credentials, live database/provider calls, reverse ETL, target DML, or
  generic SQL/HTTP/shell pathway.
- `commandrunner` REST preflight is unchanged except only if a truthful native
  sync route is demonstrably required; no `api_surface` entry is added.
- No changes to #2986/#3745–#3749 changefeed capability derivation or #3762
  bounded-query taxonomy.
