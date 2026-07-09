# Plan: Issue #106 Jira Stream Runner

Parent: #81 Jira CLI feature parity parent roadmap
Issue: #106 Jira: stream runner (CLI parity)
Branch: `feat/81-jira-cli-parity`
Worker mode: local critical path (no Pi subagent tool exposed in this harness)

## Required skills loaded

- gsd-core
- golang-how-to
- golang-cli
- golang-testing
- golang-error-handling
- golang-security
- golang-safety
- golang-spf13-cobra (CLI-command-surface context)
- golang-documentation

## GSD status

- `scripts/gsd prompt plan-phase issue-106-jira-stream-runner --skip-research` generated a repo-local Pi adapter prompt.
- `scripts/gsd prompt programming-loop init --phase issue-106-jira-stream-runner --dry-run` failed with `unknown GSD command: programming-loop`.
- Manual GSD/TDD fallback remains active and recorded.

## Scope

Implement and prove safe stream-backed execution for Jira connector CLI commands through the generic connector runner:

- `pm jira issue list` -> Jira `issues` stream (`GET /rest/api/3/search`)
- `pm jira project list` -> Jira `projects` stream (`GET /rest/api/3/project/search`)
- `pm jira user list` -> Jira `users` stream (`GET /rest/api/3/users/search`)

Add only bounded read/query flags that map to existing Jira REST query parameters. Do not add write actions, raw API escape hatches, credentialed live checks, or destructive/admin execution.

## TDD slices

1. Red: CLI stream command test for `pm jira issue list --jql ... --limit 1 --json` against `httptest`, expecting query forwarding and a `ConnectorCommandRead` envelope.
2. Red: CLI stream command tests for `pm jira project list --query ...` and `pm jira user list --query ...`, expecting Jira endpoints and stream envelopes.
3. Green: add Jira `cli_surface.json` stream command flags mapping to `query.*` targets.
4. Refactor: keep generic runner code unchanged unless tests expose a shared runner defect; prefer metadata-only updates.
5. Verify: targeted CLI/commandrunner/connectorgen checks, docs/website parity if generated output changes, then full gates.

## Safety gates

- No secrets in prompts, logs, docs, or JSON output.
- Test credentials use local `httptest` and env vars only; no live Jira calls.
- Reads are bounded by `--limit`; runner already clamps oversized limits.
- Unknown stream command flags must be rejected.
- Planned/direct-write/admin/sensitive commands remain blocked.
- Parent PR merge to `main` remains human-gated.
