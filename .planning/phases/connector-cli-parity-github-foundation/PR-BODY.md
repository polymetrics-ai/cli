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

Closes #43

## Verification

```bash
go test ./internal/coordination/issueguard
go test ./cmd/prissueguard ./internal/coordination/issueguard
find .agents .github/ISSUE_TEMPLATE .github/workflows -type f \( -name '*.yaml' -o -name '*.yml' \) -print0 | xargs -0 ruby -e 'require "yaml"; ARGV.each { |f| YAML.load_file(f) }'
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
6. Request an incremental CodeRabbit pass with `@coderabbitai review`.
7. Repeat until CodeRabbit has no actionable findings.
8. Post `@coderabbitai resolve` only after every actionable finding has been addressed.
9. Ping the human coordinator for approval before merge.

## CodeRabbit disposition status

- Initial full review requested on PR #47.
- Full review produced 7 actionable findings.
- Disposition: all 7 findings were accepted as valid, in-scope fixes.
- Fixed areas: PR guard CLI exit-code test coverage, explicit agent denied-tools, structured
  conditional denied paths, `Closes`/`Refs` prompt wording, connector-migration quality-gate hard
  stop, run-state wording, and verification-log consistency.
- Declined findings: none.
- Deferred findings: none.
- Follow-up incremental review will be requested after the fix commit is pushed. If CodeRabbit is
  rate-limited, the review request remains pending until the next available review window.

## Checklist

- [x] Tests or docs updated for behavior changes
- [x] `make verify` passes locally
- [x] CodeRabbit review requested after implementation
- [x] Every actionable CodeRabbit finding has a reasoned disposition reply or summary
- [x] Branch name follows `<type>/<description>`
- [x] PR title follows Conventional Commits
- [x] No credentials, tokens, private URLs, or customer data included
