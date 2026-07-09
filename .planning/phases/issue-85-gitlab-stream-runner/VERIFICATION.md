# Verification: GitLab Stream Runner (#85)

## Commands

- `go test ./internal/cli -run 'TestGitLabCommandSurfaceRunsStreamBackedIssueList' -count=1` ✅
- `go test ./cmd/connectorgen ./internal/connectors/engine ./internal/connectors/commandrunner ./internal/cli -count=1` ✅

## Shared focused gate

- `go test ./cmd/connectorgen ./internal/connectors/engine ./internal/connectors/commandrunner ./internal/cli -count=1` ✅
- `go run ./cmd/connectorgen validate internal/connectors/defs --json` ✅ (`connectors_checked=547`, `findings=[]`)
