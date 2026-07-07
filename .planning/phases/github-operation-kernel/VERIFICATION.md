# Verification: GitHub Operation Kernel

Issue: #56
Branch: `feat/56-operation-kernel`

## Targeted Commands

```bash
jq . internal/connectors/defs/github/operations.json
go test ./internal/connectors/engine -run 'TestBundleLoad.*Operation|TestBundleLoadEmbeddedGitHub'
go test ./cmd/connectorgen -run 'TestValidate_.*Operation|TestValidate_CLISurface'
go test ./internal/connectors/commandrunner
go build ./cmd/pm
```

## Broader Commands

```bash
go test ./internal/connectors/...
go test ./cmd/...
go vet ./...
go test ./...
go build ./cmd/pm
make verify
```

## Results

Focused green:

- `jq . internal/connectors/defs/github/operations.json`: pass.
- `jq . internal/connectors/defs/github/api_surface.json`: pass.
- `jq . internal/connectors/engine/schema/operations.schema.json`: pass.
- `jq . internal/connectors/engine/schema/api_surface.schema.json`: pass.
- `jq . internal/connectors/engine/schema/cli_surface.schema.json`: pass.
- `go test ./internal/connectors/engine -run 'TestBundleLoad.*Operation|TestBundleLoadEmbeddedGitHub'`: pass.
- `go test ./cmd/connectorgen -run 'TestValidate_.*Operation|TestValidate_CLISurface'`: pass.
- `go test ./internal/connectors/commandrunner`: pass.
- `go test ./internal/connectors/engine ./internal/connectors/commandrunner ./cmd/connectorgen`: pass.
- `go test ./cmd/connectorgen ./internal/connectors/engine ./internal/connectors/commandrunner ./internal/connectors/conformance`: pass.
- `go test ./internal/cli -run 'TestGitHubCommandSurfaceBlocksOperationBeforeCredentialResolution|TestGitHubCommandSurfaceBlocksReverseETLCommand|TestGitHubCommandSurfaceRunsStreamBackedIssueList|TestGitHubCommandSurfaceRunsDirectReadFile'`: pass.
- `go test ./cmd/...`: pass.
- `go build ./cmd/pm`: pass.
- `go run ./cmd/connectorgen validate internal/connectors/defs/github --json`: blocked by current CLI behavior; the command treats `schemas/` and `fixtures/` as connector directories when pointed at a single connector directory.
- `go run ./cmd/connectorgen validate internal/connectors/defs --json`: pass, 547 connectors checked, zero findings, zero warnings.

Broader green:

- `go test ./cmd/...`: pass.
- `go test ./internal/connectors/engine ./internal/connectors/commandrunner ./internal/connectors/bundleregistry ./internal/connectors/conformance`: pass.
- `go vet ./...`: pass.
- `go build ./cmd/pm`: pass.

Broader warning:

- `go test ./...`: started and reported passing packages through
  `polymetrics.ai/internal/app`, then the PTY stayed open without any matching
  `go test` process in `ps`. The session was interrupted and is recorded as
  inconclusive, not passing.
- `go test ./internal/connectors/...`: started and reported passing
  `polymetrics.ai/internal/connectors` and
  `polymetrics.ai/internal/connectors/bundleregistry`, then the PTY stayed
  open without any matching `go test` process in `ps`. The session was
  interrupted and is recorded as inconclusive, not passing. Explicit connector
  packages listed above passed.
- `make verify`: not run for this stacked foundation slice after the full-suite
  PTY issue; focused validator, build, vet, command tests, and connector
  conformance gates are green.
