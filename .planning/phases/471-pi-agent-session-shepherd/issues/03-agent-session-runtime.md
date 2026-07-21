# Objective

Implement the in-process Pi `AgentSession` role runtime used by autonomous workers, with exact model
routing, bounded context, recursion prevention, least-authority tools, cancellation, deadlines, and
structured redacted handoffs.

Parent: #471
Parent PR: #472
Dependency: #473 (Wave 2)
Branch: `feat/475-shepherd-agent-session-runtime`
PR base: `feat/471-pi-agent-session-shepherd`

## Allowed write scope

- `.pi/extensions/shepherd/agent-session-runtime.ts`
- `.pi/extensions/shepherd/tool-policy.ts`
- `.pi/extensions/shepherd/role-prompts.ts`
- matching tests and Shepherd-specific role prompt assets
- this issue's GSD/TDD artifacts

Do not wire the top-level scheduler or GitHub/worktree mutations here.

## Acceptance criteria

- [ ] Workers use Pi 0.80.6 `createAgentSession`; no child `pi` process or tmux transport exists.
- [ ] Implementation routes to `openai-codex/gpt-5.6-sol`/`high`; all other roles route to the same
      model/`xhigh`; routing rejects 5.5 or unknown fallback.
- [ ] Mutating sessions receive workspace-bound read/edit/write plus typed host capabilities only;
      they do not receive unrestricted generic shell, HTTP write, SQL write, or orchestration tools.
- [ ] Read-only roles cannot mutate even when prompted to do so.
- [ ] Prompt injection cannot expand issue, branch, workspace, tool, model, or secret authority.
- [ ] Abort, timeout, close, and parent shutdown terminate and join sessions exactly once.
- [ ] Handoffs are schema-validated, bounded, redacted, and bound to run/generation/lane/head/nonce.

## TDD and verification

Use a fake injected SDK for RED lifecycle/authority tests before the adapter. Required skills:
`javascript-testing-patterns`, `architecture-patterns`, repository security routing.

```bash
node --test .pi/extensions/shepherd/agent-session-runtime.test.ts \
  .pi/extensions/shepherd/tool-policy.test.ts
git diff --check
```

Human gates: none. No live credential or production call is allowed in this slice.
