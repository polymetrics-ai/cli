# Local Verification

CI is active on PR #3737. Repository-wide `go test ./...` / `make verify` were
intentionally not run locally because their duration exceeds the per-command window;
the scoped packages and individual `make verify` gates below were run instead.

| Check | Status | Command / evidence |
| --- | --- | --- |
| Formatting | pass | `gofmt` on every changed Go file; `git diff --check` |
| Targeted Go tests | pass | `go test ./internal/connectors/engine ./cmd/connectorgen ./internal/connectors/commandrunner ./internal/connectors/defs/zendesk-support -count=1` |
| Hollow-command runtime sweep | pass | `go test ./internal/connectors/commandrunner -run '^TestEveryImplementedCommandPassesRuntimePreflight$' -count=1` |
| CLI package | pass | `go test ./internal/cli -count=1` |
| Static analysis | pass | `go vet ./cmd/connectorgen ./internal/connectors/engine ./internal/connectors/commandrunner ./internal/connectors/defs/zendesk-support ./internal/cli` |
| Build | pass | `go build ./cmd/pm` |
| Dependency / lint | pass | `make tidy-check`; `make lint` (0 issues) |
| Connector/docs gates | pass | `make docs-check-no-build`; `make connectorgen-validate`; `make connectorgen-surface-sync`; `make connector-boundary`; `make release-workflow-check` |
| Runtime smoke | pass | `make smoke-no-build` |
| Zendesk-specific surface | pass | `go run ./cmd/connectorgen validate internal/connectors/defs/zendesk-support --json`; `go run ./cmd/connectorgen surface-sync --check` |
| CLI help parity | pass | `pm help zendesk-support`; bare `pm zendesk-support`; `pm zendesk-support operations update_permission_policy --help` |
| Generated docs / website data | pass | `pm docs generate --dir docs/cli`; `website: npm run gen:website-data`; generated Zendesk manual/skill and catalog checked by grep |
| Website TypeScript check | not run | Local `website/node_modules` lacks `tsc`; generated-data script requires only Node built-ins and passed. CI owns the dependency-installed typecheck. |
| CodeQL remediation | pass locally; CI rerun pending | `go test ./internal/connectors/engine -run '^Test(MergeRecordSchemaRequired|InspectRecordSchema|ValidatePromotableRecordSchema)' -count=1`; commandrunner promotion tests; `go vet ./internal/connectors/engine ./internal/connectors/commandrunner`; the PR's next CodeQL result is authoritative for the eliminated capacity-overflow finding. |

Sweep note: connectorgen's initial whole-defs pass revealed Amazon SQS native
empty-record queue actions, which are intentionally handled by its Tier-3 native
executor against configured `queue_url`. The final runtime sweep passed; no other
declarative implemented command has the Zendesk hollow-schema defect.
