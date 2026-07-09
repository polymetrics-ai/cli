# Verification: GitLab Direct Read (#87)

## Commands

- `go test ./internal/connectors/engine -run TestDirectReadJSONRedactedPolicyRemovesSensitiveFields -count=1` ✅
- `go test ./internal/cli -run 'TestGitLabCommandSurfaceRunsBoundedDirectReadProjectView' -count=1` ✅

## Shared focused gate

- `go test ./cmd/connectorgen ./internal/connectors/engine ./internal/connectors/commandrunner ./internal/cli -count=1` ✅
- `go run ./cmd/connectorgen validate internal/connectors/defs --json` ✅ (`connectors_checked=547`, `findings=[]`)
