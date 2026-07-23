# TDD Ledger: #480

## Plan-first state

- Production edits: not started.
- GSD mode: existing manual fallback because `programming-loop` is absent from the healthy
  69-command adapter; no repeated retry.
- Orchestration decision: `worker_ready` through the single durable `/pm-shepherd` run after the
  parent reconciliation checkpoint is committed and pushed.
- Required review budget: one comprehensive Codex 5.6-sol xhigh round; at most one correction pass.

## Required RED

| ID | Failing behavior required before production edit |
|---|---|
| R0 | an unprotected non-default parent branch yields an exact empty required-check policy while protected `main` retains its exact checks; current production instead aborts policy intake on GitHub 404 |
| R1 | restart schedules no work before state/lease/worktree/ref/GitHub reconciliation completes |
| R2 | every ambiguous external-effect window reconciles once or fails closed without duplicate mutation |
| R3 | audit records reject unknown/oversized/control/secret-bearing payloads and remain causally ordered |
| R4 | cancellation/stop aborts and joins recovery/audit work before lease release or terminal persistence |
| R5 | stale heads, force movement, conflicts, review change, and rate limiting preserve exact authority |
| R6 | cutover preparation is reversible and cannot activate deprecation before a bound #481 pass receipt |

The coordinator owns R0 as the local critical-path unblocker before `/pm-shepherd start`; it must
commit the failing test before the production fix. The worker records exact focused failing
command/counts for R1-R6 after they execute. A missing-module/compile failure alone is not behavior
RED; use a compiling throwing scaffold if new modules are introduced.

### R0 RED evidence

```bash
node --test --test-name-pattern='records an exact empty policy' \
  .pi/extensions/shepherd/gh-orchestration-transport.test.ts
```

Result: **expected RED — 0 pass / 1 fail**. The assertion executed against compiling production and
failed because `GhRequiredCheckPolicySource` called the absent required-status-check resource before
observing that the non-default parent branch is unprotected. No load, compile, timeout, skip, or
unrelated assertion contributed.

## GREEN / refactor / review

Pending worker handoff. GREEN must cite focused assertions and the full Shepherd suite. Refactor must
retain the same focused GREEN. Review findings require one written disposition and, when accepted,
a focused behavior RED before the single allowed correction pass.
