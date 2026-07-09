# Local Verification

- CI detected: no
- Local harness required: yes

| Check | Status | Command | Notes |
| --- | --- | --- | --- |
| Install or lockfile validation | missing | TBD | No matching command was detected in the repo profile. |
| Format check | configured | gofmt -l cmd internal |  |
| Lint | configured | go vet ./... |  |
| Typecheck or static analysis | configured | go vet ./... |  |
| Unit tests | configured | go test ./... |  |
| Integration tests | configured | make smoke |  |
| E2E or smoke tests | configured | make smoke |  |
| Build | configured | go build ./... |  |
| Dependency vulnerability scan | missing_optional_tool | TBD | Add or configure dependency scanning before production release. |
| Secret scan | configured | git diff --check |  |
| Accessibility check | not_applicable | N/A | Backend engine change; no user-facing UI in this phase. |
| Load or benchmark | not_applicable | N/A | Small request-body construction change; package tests cover behavior. |

## Phase Checks

- Passed: `go test ./internal/connectors/engine -run 'TestReadGraphQL|TestWriteGraphQL|TestBundleLoad.*GraphQL'`
- Passed: `go test ./internal/connectors/engine`
- Passed: `go test ./cmd/connectorgen ./internal/connectors/engine`
- Passed: `go run ./cmd/connectorgen validate internal/connectors/defs --json` (`547` connectors, `0` findings, `0` warnings)
- Passed: `go test ./internal/connectors/conformance -run 'TestConformance/github'`
- Passed with local crontab redirection: `PM_CRONTAB_FILE=$(mktemp) go test ./...`
- Passed: `go vet ./...`
- Passed: `go build ./cmd/pm`
- Passed: `jq empty internal/connectors/engine/schema/streams.schema.json internal/connectors/engine/schema/writes.schema.json .planning/phases/github-graphql-engine/TDD-GATE.json .planning/phases/github-graphql-engine/RUN-STATE.json .planning/phases/github-graphql-engine/AGENT-ORCHESTRATION.json`
- Passed: `git diff --check`

## Notes

An unguarded `go test ./...` on this local macOS environment timed out in
`internal/cli TestScheduleCLI_Remove` while waiting for the real `crontab` command. The same test
passes when `PM_CRONTAB_FILE` redirects crontab writes to a temp file, matching the existing schedule
test harness used by other crontab tests.
