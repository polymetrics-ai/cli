# Local verification

The shared connector-parity contract supersedes generic full-suite suggestions because the full suite exceeds the bounded execution window. Results are recorded at completion.

| Check | Command | Status |
| --- | --- | --- |
| Module integrity | `go mod verify` | pass |
| Surface regeneration | `go run ./cmd/connectorgen surface-sync` | pass; 550 scanned and only the target surface was updated before the final check |
| Surface drift | `go run ./cmd/connectorgen surface-sync --check` | pass; 550 scanned, zero changes |
| Connector validation | `go run ./cmd/connectorgen validate internal/connectors/defs/youtube-analytics` | pass; one connector, zero findings |
| Connector conformance | `go test ./internal/connectors/conformance/...` | pass |
| Runtime preflight sweep | `go test ./internal/connectors/commandrunner -run TestEveryImplementedCommandPassesRuntimePreflight` | pass |
| CLI package tests | `go test ./internal/cli/...` | pass after regenerating the YouTube connector root-help transcripts |
| Static analysis | `go vet ./internal/cli/...` | pass |
| CLI build | `go build ./cmd/pm` | pass |
| Runtime help/execution | built `pm help youtube-analytics`, bare `pm youtube-analytics`, `reports download --help`, and `connectors inspect youtube-analytics --json`; a flag-complete download reaches `missing --credential` (exit 1), not a planned-operation block | pass; no credential or provider request was used |
| Website data | `(cd website && pnpm run gen:website-data)` | pass; target connector catalog outputs regenerated |
| Whitespace/secret-safe review | `git diff --check` and manual diff review | pass; only target bundle, target generated surfaces, test golden output, and phase evidence changed |
| Review repair: typed stream query selectors | `go test ./internal/connectors/engine -run 'TestReadRequestQuery(ResolvesStreamQueryTemplate|TemplateMissingFailsBeforeRequest)$' -count=1` | pass; request-backed templates resolve and missing selector values fail before any request |
