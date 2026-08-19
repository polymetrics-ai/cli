# Code review — Issue 3973: transactional database write sessions

Inline `code-review` of the issue diff was completed because the canonical
single-worker contract forbids role spawning in this environment.

## Scope reviewed

- `internal/connectors/database/database_write_session.go`
- `internal/connectors/database/database_write_session_test.go`
- `internal/connectors/database/registry.go`
- Issue-specific GSD evidence only

## Findings

No actionable correctness, security, or scope findings.

The review verified that the new public boundary accepts no SQL, DSN,
credential, raw relation, arbitrary strategy, or legacy connector write path;
driver errors are converted to bounded sentinels; approvals use atomic one-shot
consumption; cancellation rolls back with the same session; and uncertain
commit does not issue rollback or retry. Exact driver/definition compatibility
is checked before preview or begin. The changed-path check confirms no
PostgreSQL connector/capability, CLI, generated surface, or unrelated tracked
file changed.

## Review route after PR opening

Primary route: `claude_auto` on the non-draft stacked PR, pending GitHub’s
trusted-author trigger. Parent PR: #4100 (`integration/4015-mvp-flat-r1` →
`main`). Copilot is fallback-only if Claude is unavailable.
