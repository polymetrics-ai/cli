# Issue #4365 verification checklist

- [ ] `go test -timeout 20m ./internal/connectors/engine -run 'SentrySeerModels|OperationRoutes'`
- [ ] `go test -timeout 20m ./internal/connectors/commandrunner -run 'SentrySeerModels|Preflight'`
- [ ] `go test -timeout 20m ./internal/cli -run 'SentrySeerModels|Connector'`
- [ ] `go run ./cmd/connectorgen validate internal/connectors/defs`
- [ ] `go run ./cmd/connectorgen surface-sync` then `--check`
- [ ] `go run ./cmd/connectorgen declaration-admission internal/connectors/defs --json`
- [ ] `go run ./cmd/connectorgen operation-evidence . --check`
- [ ] `go run ./cmd/connectorgen surface-reconcile internal/connectors/defs --check`
- [ ] `go run ./cmd/connectorgen boundary . --json --base origin/main`
- [ ] generated connector docs/manual command and `docs-check`
- [ ] `pm help sentry`, `pm sentry`, `pm sentry seer`, and `pm sentry seer list-models --help`
- [ ] built-binary missing-credential/zero-transport proof in an isolated initialized project
- [ ] `gofmt`, `go vet ./...`, scoped package tests, `go build ./cmd/pm`, individual verify gates, and `git diff --check`
- [ ] inline GSD `execute-phase`, `verify-work`, and `code-review` records plus independent exact-head audit request
