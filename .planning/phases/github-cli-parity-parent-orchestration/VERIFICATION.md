# Verification: GitHub CLI Parity Parent Orchestration

## Pending

- `jq empty .planning/config.json`
- YAML parse check for `.agents/`, `.github/ISSUE_TEMPLATE`, and `.github/workflows`
- `git diff --check`
- `go test ./cmd/connectorgen ./internal/connectors/engine`
- website checks for the docs/catalog changes
- broader verification required by #44 before parent PR human gate

## Notes

The parent branch has been rebased locally onto `origin/main`. Push and remote CI are still pending.
