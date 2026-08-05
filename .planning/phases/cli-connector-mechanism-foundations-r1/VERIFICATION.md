# Verification: Dual-Mechanism Connector Foundations (P0)

Branch: `fm/cli-connector-mechanism-foundations-r1`

## Focused checks

```bash
go test ./internal/browserauth/...
go test ./internal/vault
go test ./internal/connectors/engine
go test ./internal/connectors/bundleregistry
go test ./internal/cli
go build ./cmd/pm
go vet ./...
```

## Surface and repository checks

```bash
go run ./cmd/pm help connectors
go run ./cmd/pm connectors
go run ./cmd/pm connectors inspect github --json
make tidy-check
make lint
make docs-check
make smoke-no-build
make connectorgen-validate
make connectorgen-surface-sync
make connector-boundary
make release-workflow-check
```

## Results

Fresh local green results:

- `go test ./internal/browserauth/...`
- `POLYMETRICS_BROWSER_INTEGRATION=1 go test ./internal/browserauth/driver -run TestFlowYieldsRealBrowserSessionCredential -count=1`
- `go test ./internal/vault ./internal/connectors ./internal/connectors/engine ./internal/connectors/bundleregistry`
- `go test ./internal/app`
- `go test ./internal/cli`
- `go build ./cmd/pm`
- `go vet ./...`
- `make tidy-check`
- `make lint`
- `make docs-check`
- `make smoke-no-build`
- `make connectorgen-validate`
- `make connector-boundary`
- `make release-workflow-check`
- `./pm help connectors`, `./pm connectors`, `./pm connectors --help`, and `./pm connectors inspect github --json`
- docs/website parity grep for `web_session`, `[UNOFFICIAL]`, and `connectors enable`

The prior no-mistakes pipeline recovered its commits but failed at the
agent-process test step, not at a recorded Go test failure. Full `go test
./...` and `make verify` are intentionally left to CI because this checkout's
per-command timeout makes them unreliable as a single local invocation.

`connectorgen surface-sync --check` is **not performed** on this stale branch:
its command is absent here and present on current `main`. After no-mistakes
rebases the branch onto current `main`, run that exact command locally and
record its actual exit result before calling this verification complete.
