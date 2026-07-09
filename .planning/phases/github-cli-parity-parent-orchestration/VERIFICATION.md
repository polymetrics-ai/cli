# Verification: GitHub CLI Parity Parent Orchestration

## Completed

- `jq empty .planning/config.json`
- YAML parse check for `.agents/`, `.github/ISSUE_TEMPLATE`, and `.github/workflows`
- `git diff --check`
- `go test ./cmd/connectorgen ./internal/connectors/engine`
- `go run ./cmd/connectorgen validate internal/connectors/defs`
- `pnpm --filter cli-polymetrics-ai test`
- `pnpm --filter cli-polymetrics-ai build`
- `go vet ./...`
- `go test ./...`
- `go build ./cmd/pm`
- `make verify`

## Pending

- broader verification required by #44 before parent PR human gate
- active-orchestration slice checks after current edits:
  - YAML parse check for `.agents/`, `.github/ISSUE_TEMPLATE`, `.github/workflows`, and
    `.opencode/agents`
  - TOML parse check for `.codex/agents`
  - `jq empty .planning/phases/github-cli-parity-parent-orchestration/ORCHESTRATION-STATE.json`
  - `git diff --check`

## Notes

The parent branch has integrated #69 locally. Push and remote CI are pending for the active
orchestration updates.
