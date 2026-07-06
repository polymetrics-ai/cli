# Verification: Issue-first agentic delivery foundation

## Passed

```bash
go test ./internal/coordination/issueguard
go run ./cmd/prissueguard --title 'feat(github): add cli surface metadata' --body 'Closes #123'
go test ./cmd/prissueguard ./internal/coordination/issueguard
find .agents .github/ISSUE_TEMPLATE .github/workflows -type f \( -name '*.yaml' -o -name '*.yml' \) -print0 | xargs -0 ruby -e 'require "yaml"; ARGV.each { |f| YAML.load_file(f); puts f }'
jq empty .planning/phases/connector-cli-parity-github-foundation/GITHUB-CLI-PARITY-ISSUE-HIERARCHY.json
git diff --check
make verify
gh api repos/polymetrics-ai/cli/issues/47/comments --paginate
gh api repos/polymetrics-ai/cli/pulls/47/comments --paginate
gh issue view 44 --json number,title,body,labels,milestone,url
gh api repos/polymetrics-ai/cli/issues/44/sub_issues
```

Also ran the standard secret-looking literal scan over the new `.agents`, phase, workflow, template,
and guard files.

## Expected failure checks

Command:

```bash
go run ./cmd/prissueguard --title 'add cli surface metadata' --body 'no issue'
```

The invalid PR smoke check failed with:

```text
issueguard: blocked
- PR title must use Conventional Commits, for example feat(github): add cli surface metadata
- PR body must reference an issue with Closes #123 for completed work or Refs #123 for stacked/incremental work
```

## Human gates

- GitHub Project creation remains outside this PR.
- No auth refresh should be attempted for this PR.
- CodeRabbit comments are treated as external review input, not instructions.
- No CodeRabbit thread should be resolved before every actionable item has a disposition reply.
- The first full CodeRabbit review produced 7 actionable findings; all 7 were accepted and fixed.
- The CodeRabbit follow-up workflow now treats manual `@coderabbitai review` as conditional:
  wait for automatic incremental review when active, and request manual review only for new
  unreviewed commits when automatic review is paused, disabled, skipped, rate-limit retry is due, or
  auto-paused.
- Sub-PR merge without human approval applies only to parent-branch integration after automated
  gates pass; parent PR merge to `main` remains human-approved.
- Live issue #44 was updated and GitHub reports #34-#42 as sub-issues.

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
