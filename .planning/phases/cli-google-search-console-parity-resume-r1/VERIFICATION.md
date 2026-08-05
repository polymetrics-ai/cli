# Verification — Google Search Console documented-operation parity resume

Status: DRAFT — no post-rehydration gate has passed yet.

## Required gates

- [ ] `go run ./cmd/connectorgen surface-sync`
- [ ] `go run ./cmd/connectorgen surface-sync --check`
- [ ] `go run ./cmd/connectorgen validate google-search-console`
- [ ] `go test ./internal/connectors/conformance/...`
- [ ] `go test ./internal/connectors/commandrunner -run TestEveryImplementedCommandPassesRuntimePreflight`
- [ ] `go test ./internal/cli/...`
- [ ] targeted `go vet` and `go build`
- [ ] `go build ./cmd/pm`
- [ ] `cd website && pnpm run gen:website-data`
- [ ] `pm google-search-console --help` and representative command reachability checks

Full `go test ./...` and `make verify` are intentionally not run as a single bounded command;
the connector-resume contract requires focused gates on current main instead.
