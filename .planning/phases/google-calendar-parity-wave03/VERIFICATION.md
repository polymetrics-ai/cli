# Google Calendar parity resume — verification checklist

## Required after implementation

- [ ] `go run ./cmd/connectorgen surface-sync`
- [ ] `go run ./cmd/connectorgen surface-sync --check`
- [ ] `go run ./cmd/connectorgen validate internal/connectors/defs/google-calendar`
- [ ] `go test ./internal/connectors/conformance/...`
- [ ] `go test ./internal/connectors/commandrunner -run TestEveryImplementedCommandPassesRuntimePreflight`
- [ ] `go test ./internal/cli/...`
- [ ] `go vet` and `go build` for changed packages
- [ ] `go build ./cmd/pm`
- [ ] `cd website && pnpm run gen:website-data`
- [ ] `./pm google-calendar --help`, `./pm google-calendar freebusy query --help`, and representative uncredentialed validation paths
- [ ] `git diff --check`

## Constraints

Run no credentialed provider checks and no write execution. Do not claim full `make verify` or `go test ./...`; the shared parity contract requires focused gates because the full suite exceeds bounded command windows.
