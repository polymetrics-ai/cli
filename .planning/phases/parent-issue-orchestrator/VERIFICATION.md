# Verification

Issue: #50

## Planned Commands

```bash
find .agents .github/ISSUE_TEMPLATE .github/workflows -type f \( -name '*.yaml' -o -name '*.yml' \) -print0 | xargs -0 ruby -e 'require "yaml"; ARGV.each { |f| YAML.load_file(f) }'
git diff --check
```

`go test ./cmd/prissueguard ./internal/coordination/issueguard` is required because this slice
updates issue and PR templates that feed the issue-first workflow.

## Results

```bash
find .agents .github/ISSUE_TEMPLATE .github/workflows -type f \( -name '*.yaml' -o -name '*.yml' \) -print0 | xargs -0 ruby -e 'require "yaml"; ARGV.each { |f| YAML.load_file(f) }'
```

Result: passed.

```bash
git diff --check
```

Result: passed.

```bash
go test ./cmd/prissueguard ./internal/coordination/issueguard
```

Result: passed.

## Notes

`scripts/programming-loop.mjs` is not present in this clone, so this phase used the manual GSD
fallback described in `PLAN.md`.

CodeRabbit usage policy was corrected after the first PR comment showed a review-limit warning from
an unnecessary manual `@coderabbitai full review` command. Future non-draft PRs targeting `main`
should wait for automatic CodeRabbit review.

The reviewer routing policy now records GitHub Copilot as a backup route only when CodeRabbit is
rate-limited, skipped, disabled, paused, or unavailable and review coverage is blocking progress.
Copilot review comments are dispositioned like CodeRabbit comments, but Copilot review is not
approval and does not replace human approval for parent PRs into `main`.
