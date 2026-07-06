# Verification: Issue-first agentic delivery foundation

## Passed

```bash
go test ./internal/coordination/issueguard
go run ./cmd/prissueguard --title 'feat(github): add cli surface metadata' --body 'Closes #123'
go run ./cmd/prissueguard --title 'add cli surface metadata' --body 'no issue'
go test ./cmd/prissueguard ./internal/coordination/issueguard
ruby -e 'require "yaml"; ARGV.each { |f| YAML.load_file(f); puts f }' .agents/connector-cli-parity/*.yaml .codex/agents/connector-cli-parity/*.agent.yaml .github/ISSUE_TEMPLATE/agent_task.yml .github/workflows/pr-issue-guard.yml
git diff --check
make verify
```

Also ran the standard secret-looking literal scan over the new `.agents`, `.codex/agents`, phase,
workflow, template, and guard files.

## Expected failure checks

The invalid PR smoke check failed with:

```text
issueguard: blocked
- PR title must use Conventional Commits, for example feat(github): add cli surface metadata
- PR body must reference an issue with Closes #123 for completed work or Refs #123 for stacked/incremental work
```

## Human gates

- GitHub Project creation remains outside this PR.
- No auth refresh should be attempted for this PR.

## `make verify` details

`make verify` passed and included:

- `gofmt -w cmd internal`
- `go mod tidy`
- `go vet ./...`
- `go test ./...`
- `go build ./cmd/pm`
- connector docs validation
- smoke test
- connector-package `golangci-lint`
- `go run ./cmd/connectorgen validate internal/connectors/defs` with 547 connectors and 0 findings
