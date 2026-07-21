# Objective

Implement a durable GitHub comment decision broker so Shepherd can ask the right human on the right
issue/PR, wait safely, validate one exact response, consume it once, and resume after restart.

Parent: #471
Parent PR: #472
Dependency: #473 (Wave 2)
Branch: `feat/477-shepherd-github-decision-broker`
PR base: `feat/471-pi-agent-session-shepherd`

## Allowed write scope

- `.pi/extensions/shepherd/human-decision.ts`
- `.pi/extensions/shepherd/github-decision-broker.ts`
- matching tests and GitHub API fixtures
- this issue's GSD/TDD artifacts

Do not implement general issue/PR orchestration or controller wiring here.

## Acceptance criteria

- [ ] Each request has a durable ID, idempotency marker, allowed options, target, generation, actor
      allowlist, expiry, and exact head SHA when head-specific.
- [ ] Requirements/scope gates route to the parent issue; review/merge/head gates route to the PR.
- [ ] Exactly one marker-owned comment is created across retry/restart; polling is bounded/backed off.
- [ ] Only `/shepherd decide <request-id> <option>` from an allowlisted human on the bound target is
      accepted; bots, edits, duplicates, stale heads/generations, and ambiguous replies fail closed.
- [ ] A valid decision is persisted and consumed once with actor, source URL, timestamp, and option.
- [ ] No token or secret reaches prompts/logs/state; host auth is used only by the typed adapter.
- [ ] Parent merge approval is represented as a distinct exact-head decision and cannot be inferred.

## TDD and verification

Start with fake GitHub transport RED cases; optional live tests use a designated sandbox only.
Required skills: `javascript-testing-patterns`, `architecture-patterns`, repository security routing.

```bash
node --test .pi/extensions/shepherd/human-decision.test.ts \
  .pi/extensions/shepherd/github-decision-broker.test.ts
git diff --check
```

Human gates: live comment creation requires the explicit parent target configured for testing; no
production/external action beyond issue/PR comments is authorized in this slice.
