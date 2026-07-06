## Objective

Create the generic agentic delivery system that turns GitHub issues into implementation-ready
prompts for coding agents, with reusable YAML agent definitions, task-to-skill mapping, PR
guardrails, and hard-stop rules.

## Background

Connector CLI migration will require many incremental PRs across schema, renderer, executor,
GraphQL, sensitive/admin policy, and cross-connector rollout work. The project needs an
agent-neutral system so issues can launch implementation work without relying on chat history or a
specific agent vendor.

## Scope

- Add a generic issue-to-PR agent contract.
- Add task-to-skill mapping for implementation, architecture, security, docs, and review tasks.
- Add YAML agent best practices and an agent spec contract.
- Add reusable YAML agents for implementation, review, operation ledger, CLI surface, help
  renderer, direct-read, GraphQL, sensitive/admin, and rollout tasks.
- Add a CodeRabbit review-disposition workflow so automated review findings are answered with
  accepted, accepted_with_modification, declined, deferred, or needs_human before resolution.
- Make CodeRabbit follow-up review conditional: wait for automatic incremental review when active,
  and use manual `@coderabbitai review` only for new unreviewed commits when automatic review is
  paused, disabled, skipped, rate-limit retry is due, or auto-paused.
- Add parent issue, sub-issue, and stacked parent-branch PR workflow rules for large implementation
  efforts such as GitHub CLI feature parity.
- Add a GitHub issue form for agent implementation tasks.
- Add a PR guard that requires Conventional Commit titles and explicit issue references.
- Keep agent contracts and role specs under `.agents/`, grouped by functional area and agent type.
- Convert pre-existing `.codex/agents` connector-migration TOML specs into `.agents/` YAML specs.
- Update root agent instructions so Codex, Claude Code, and other agents share the same issue-first
  and CodeRabbit follow-up review rules.

## Non-goals

- Do not implement the full GitHub connector CLI migration in this issue.
- Do not add runtime connector command dispatch beyond the PR guard.
- Do not create or refresh GitHub Project auth scopes.
- Do not request, print, store, or invent secrets.

## Acceptance criteria

- Agent-neutral delivery contracts exist under `.agents/agentic-delivery/`.
- YAML agent definitions are grouped by functional area and type under `.agents/`.
- Pre-existing `.codex/agents` connector-migration TOML files are converted to
  `.agents/connector-migration/` YAML specs or removed if obsolete.
- YAML agent definitions parse successfully.
- PR template and CI guard require issue linkage.
- Guard tests pass and reject ambiguous references such as `Related to #123`.
- CodeRabbit review loop is documented as a post-implementation gate.
- CodeRabbit follow-up review rules avoid redundant manual incremental-review comments.
- Every actionable CodeRabbit finding must receive a reasoned disposition reply before resolve.
- Parent issue and stacked sub-PR workflow is documented and wired into reusable agents.
- GitHub CLI feature parity parent issue #44 has a saved roadmap and sub-issue hierarchy.
- Root `AGENTS.md` and `CLAUDE.md` point agents at the shared issue-first and CodeRabbit review
  contracts.
- `make verify` passes.

## TDD plan

1. Add failing tests for PR body/title validation.
2. Implement minimal Go guard and CLI wrapper.
3. Add GitHub Actions workflow.
4. Add generic docs/YAML agent definitions.
5. Validate YAML, guard tests, and repo verification.

## Verification

```bash
go test ./internal/coordination/issueguard
go test ./cmd/prissueguard ./internal/coordination/issueguard
find .agents .github/ISSUE_TEMPLATE .github/workflows -type f \( -name '*.yaml' -o -name '*.yml' \) -print0 | xargs -0 ruby -e 'require "yaml"; ARGV.each { |f| YAML.load_file(f) }'
git diff --check
make verify
```

## Safety notes

- This issue must not create GitHub Projects or refresh auth scopes.
- This issue must not add dependencies or weaken quality gates.
- This issue must not expose generic shell, unrestricted HTTP write, unrestricted SQL write, or raw
  credential tools.

## Agent execution contract

- Contract: `.agents/agentic-delivery/contracts/issue-agent-contract.md`
- CodeRabbit loop: `.agents/agentic-delivery/workflows/coderabbit-review-loop.md`
- Stacked PR loop: `.agents/agentic-delivery/workflows/stacked-parent-subissue-workflow.md`
- Task type: `pr-guardrail`
- Primary agent: `.agents/agentic-delivery/agents/implementation/issue-first-implementation-agent.agent.yaml`
- Required skill groups: `github_planning`, `go_tdd`, `security_review`, `review_disposition`
