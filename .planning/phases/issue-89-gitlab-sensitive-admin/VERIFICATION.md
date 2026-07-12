# Verification: GitLab Sensitive/Admin Policy (#89)

## Commands

- `go test ./internal/connectors/engine -run TestBundleLoadGitLabOperationLedgerFromDisk -count=1` ✅
- `go run ./cmd/connectorgen validate internal/connectors/defs --json` ✅

## Shared focused gate

- `go test ./cmd/connectorgen ./internal/connectors/engine ./internal/connectors/commandrunner ./internal/cli -count=1` ✅
- `go run ./cmd/connectorgen validate internal/connectors/defs --json` ✅ (`connectors_checked=547`, `findings=[]`)
