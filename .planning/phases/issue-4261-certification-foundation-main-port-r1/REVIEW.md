# Inline code review — issue 4261

## Scope reviewed

- The complete `origin/main` diff for the certification evidence publisher,
  proof construction, sweep projection, matrix scope checks, runner, tests,
  generator entry point, documentation, generated sweep, and live evidence.
- The new schema-v2 GitHub proof record and the worktree-only pre-push hook.

## Findings

No actionable findings.

The review specifically confirmed that the draft importer accepts only a
direct child of the run-local staging directory, validates strict accepted
evidence before publication, uses no-replace publication, leaves accepted
evidence intact after scoped matrix drift, and passes the Keychain credential
only by environment-variable reference. The live runner is limited to the
declaration-owned read-only candidate in this delivery.

## Automated review route

The planned PR route is `claude_auto`: this non-draft, main-targeted PR will be
opened by the repository owner. No manual Claude or Copilot request is made
before the automatic review trigger is evaluated.
