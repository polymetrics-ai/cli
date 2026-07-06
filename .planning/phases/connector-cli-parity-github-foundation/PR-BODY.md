## Summary

Adds the issue-first agentic delivery foundation for connector CLI parity:

- generic issue-to-PR agent contract
- task-to-skill matrix
- YAML agent definitions for implementation, review, operation-ledger, CLI surface, help renderer,
  direct-read, GraphQL, sensitive/admin, and rollout work
- GitHub agent task issue form
- PR issue guard command and GitHub Actions workflow
- isolated `.agents/` layout for agent conventions and role specs, grouped by function and type
- migration of pre-existing `.codex/agents` TOML specs into `.agents/connector-migration/`
- CodeRabbit review-disposition workflow, reply template, source-backed guidance, and review agent
- parent issue, sub-issue, milestone roadmap, and stacked PR workflow for GitHub CLI feature parity

Closes #43

## Verification

```bash
go test ./internal/coordination/issueguard
go test ./cmd/prissueguard ./internal/coordination/issueguard
find .agents .github/ISSUE_TEMPLATE .github/workflows -type f \( -name '*.yaml' -o -name '*.yml' \) -print0 | xargs -0 ruby -e 'require "yaml"; ARGV.each { |f| YAML.load_file(f) }'
jq empty .planning/phases/connector-cli-parity-github-foundation/GITHUB-CLI-PARITY-ISSUE-HIERARCHY.json
git diff --check
make verify
```

## CodeRabbit review loop

After the PR exists:

1. Request a complete CodeRabbit pass with `@coderabbitai full review`.
2. Collect inline review comments, top-level CodeRabbit comments, generated tasks, and summaries.
3. Classify every actionable item as accepted, accepted with modification, declined, deferred, or
   needs human.
4. Reply with the disposition, reason, and evidence before resolving the item.
5. Implement accepted in-scope fixes, then rerun targeted checks and `make verify` when behavior or
   shared contracts changed.
6. For fix commits, wait for automatic CodeRabbit incremental review when it is active.
7. Use manual `@coderabbitai review` only when automatic review is paused, disabled, skipped,
   rate-limit retry is due, or the automatic pause threshold was reached and there are new
   unreviewed commits.
8. Repeat until CodeRabbit has no actionable findings.
9. Post `@coderabbitai resolve` only after every actionable finding has been addressed.
10. Ping the human coordinator for approval before merge.

## CodeRabbit disposition status

- Initial full review requested on PR #47.
- Full review produced 7 actionable findings.
- Disposition: all 7 findings were accepted as valid, in-scope fixes.
- Fixed areas: PR guard CLI exit-code test coverage, explicit agent denied-tools, structured
  conditional denied paths, `Closes`/`Refs` prompt wording, connector-migration quality-gate hard
  stop, run-state wording, and verification-log consistency.
- Declined findings: none.
- Deferred findings: none.
- Follow-up incremental review was requested after the fix commit, and CodeRabbit resolved the
  addressed comments.
- A later manual `@coderabbitai review` produced CodeRabbit's incremental-review note instead of a
  new pass. The workflow has been updated so agents now wait for automatic review when active and
  only use manual incremental review for new unreviewed commits when automatic review is paused,
  disabled, skipped, rate-limit retry is due, or the automatic pause threshold was reached.

## Parent/sub-issue roadmap status

- Issue #44 is the GitHub CLI feature parity parent issue.
- Issues #34-#42 are attached as GitHub sub-issues for the GitHub pilot.
- Parent branch: `feat/44-github-cli-parity`.
- Sub-PRs target the parent branch and use `Refs #<sub-issue>` plus `Refs #44`.
- Parent PR targets `main`, includes closing keywords, and remains human-gated.

## Checklist

- [x] Tests or docs updated for behavior changes
- [x] `make verify` passes locally
- [x] CodeRabbit review completed or manually requested only when needed
- [x] Every actionable CodeRabbit finding has a reasoned disposition reply or summary
- [x] Branch name follows `<type>/<description>`
- [x] PR title follows Conventional Commits
- [x] No credentials, tokens, private URLs, or customer data included
