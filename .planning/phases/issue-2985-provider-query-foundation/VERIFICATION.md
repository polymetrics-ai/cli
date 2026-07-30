# Verification: Issue 2985 Provider Search/Query Foundation

## Required issue checks

```bash
go test ./cmd/connectorgen -run 'TestValidate_Provider|TestValidate_CLISurfaceProvider|TestValidate_InvalidOperationsJSONFindingNamesOperationsFile|TestValidate_CLISurfaceOperation'
go test ./internal/connectors/engine -run 'TestBundleLoad.*Provider|TestBundleLoad.*Operation'
go test ./internal/connectors/commandrunner -run 'Test.*Provider|TestRunImplementedOperationCommandIsFeatureGated'
go test ./internal/connectors -run 'Test.*Provider|TestDefinition|TestGuide'
go test ./internal/cli -run 'Test.*Provider|TestQuery|TestConnectors'
go build ./cmd/pm
make connector-boundary
```

## Broader local gates before handoff

```bash
gofmt -w cmd internal
go test ./cmd/connectorgen ./internal/connectors/engine ./internal/connectors/commandrunner ./internal/connectors ./internal/cli
go build ./cmd/pm
make connector-boundary
```

## No-mistakes prerequisites

- Implementation committed on `feat/connector-provider-query-2985`.
- Fetch `origin`; if `origin/feat/connector-wave-01` exists, rebase cleanly onto it before no-mistakes.
- If the integration branch is still absent, append the required paused status and stop.
- Use no-mistakes help/axi to prove PR base is exactly `feat/connector-wave-01`; do not open against `main`.

## Results

Completed before commit/rebase:

```bash
go test ./cmd/connectorgen ./internal/connectors/engine ./internal/connectors/commandrunner ./internal/connectors ./internal/cli
```

Result: pass.

```bash
go run ./cmd/connectorgen validate internal/connectors/defs --json
```

Result: pass; 548 connectors checked, zero findings, zero warnings.

```bash
go build ./cmd/pm
make connector-boundary
```

Result: pass; connector-boundary returned zero findings.

```bash
cd website && npm run typecheck
```

Result: not runable in this worktree because `tsc` is not installed (`exit 127`). No package install was performed.
