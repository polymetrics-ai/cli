# Objective

Prove durable crash/restart recovery, bounded audit telemetry, secure operator behavior, and a
reversible cutover plan from legacy shell Shepherd entry points to the in-process autonomous
controller. This issue prepares but does not activate legacy-shell deprecation.

Parent: #471
Parent PR: #472
Dependency: #479 (Wave 5)
Branch: `feat/480-shepherd-recovery-cutover`
PR base: `feat/471-pi-agent-session-shepherd`

## Allowed write scope

- `.pi/extensions/shepherd/recovery.ts`
- `.pi/extensions/shepherd/audit-log.ts`
- recovery/audit/cutover tests and fixtures
- `.pi/README.md`
- Shepherd-specific `.agents/agentic-delivery/**` and legacy-script deprecation documentation
- this issue's GSD/TDD artifacts

Do not delete legacy scripts or abandoned worktrees/branches in this issue.
Do not mark the legacy shell path deprecated in this issue. #481 must first pass the canary; the
parent orchestrator then owns the exact post-canary documentation/status activation.

## Acceptance criteria

- [ ] Restart reconciles state, lease, sessions, worktrees, branches, issues, PRs, CI, reviews, and
      human decisions before scheduling any action.
- [ ] Interrupted/ambiguous mutations fail closed or recover idempotently; duplicate comments,
      commits, pushes, PRs, integrations, or decision consumption are prevented.
- [ ] Audit events are bounded, redacted, schema-validated, causally linked, and sufficient to
      explain every stage transition and external mutation without storing prompts/secrets.
- [ ] Power loss, process kill, network failure, stale head, force-updated remote, conflict, review
      change, and GitHub rate-limit scenarios have deterministic tests.
- [ ] The proposed legacy-shell deprecation is staged as a reversible, unapplied post-canary
      change. Go is never named as a supported fallback, and no historical branch/worktree is
      destroyed.
- [ ] Operator docs state the trusted-local macOS boundary and exact human-decision syntax.

Detailed recovery ownership, pre-schedule ordering, and the #480/#481 activation boundary are in
`../traces/480-preflight-recovery-audit.md`.

## TDD and verification

Use fault-injection RED scenarios before recovery implementation. Required skills:
`javascript-testing-patterns`, `architecture-patterns`, repository security and docs routing.

```bash
node --test .pi/extensions/shepherd/*.test.ts
pi --list-extensions
git diff --check
```

Human gates: any destructive cleanup remains a separate explicitly approved task.
