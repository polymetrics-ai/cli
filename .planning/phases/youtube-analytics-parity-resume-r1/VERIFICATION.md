# Local verification

The shared connector-parity contract supersedes generic full-suite suggestions because the full suite exceeds the bounded execution window. Results are recorded at completion.

| Check | Command | Status |
| --- | --- | --- |
| Module integrity | `go mod verify` | pending |
| Surface regeneration | `go run ./cmd/connectorgen surface-sync` | pending |
| Surface drift | `go run ./cmd/connectorgen surface-sync --check` | pending |
| Connector validation | `go run ./cmd/connectorgen validate youtube-analytics` | pending |
| Connector conformance | `go test ./internal/connectors/conformance/...` | pending |
| Runtime preflight sweep | `go test ./internal/connectors/commandrunner -run TestEveryImplementedCommandPassesRuntimePreflight` | pending |
| CLI package tests | `go test ./internal/cli/...` | pending |
| Static analysis | `go vet` for changed packages | pending |
| CLI build | `go build ./cmd/pm` | pending |
| Runtime help/execution | built `pm` representative commands | pending |
| Website data | `(cd website && pnpm run gen:website-data)` | pending |
| Whitespace/secret-safe review | `git diff --check` and manual diff review | pending |
