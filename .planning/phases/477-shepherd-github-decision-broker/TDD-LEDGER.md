# TDD Ledger: #477

Status: RED captured; no production edit had been made when the failure below was recorded.

## Slice 1: durable decision aggregate

Planned RED cases:

- gate routing: requirements/scope to the parent issue; review/head/merge/parent-merge to the PR;
- PR gates require exact 40-hex head, while issue gates reject head binding;
- parent merge is a distinct gate and cannot be reconstructed from generic merge approval;
- request ID, marker, repository, generation, allowed options, actor allowlist, expiry, target, and
  head are validated and durably round-trip after repository restart;
- changed retry specifications conflict with persisted state;
- secrets/credential-shaped question content are rejected before comments/state;
- expired/stale bindings fail closed;
- a persisted decision is consumed exactly once across concurrent repository instances/restart.

## Slice 2: GitHub broker and typed adapter

Planned RED cases against issue-namespaced fake fixtures:

- one exact marker-owned request comment across retry and restart;
- duplicate/colliding markers fail closed;
- bounded paginated listing, command timeout, maximum attempts, and capped exponential backoff;
- only an exact `/shepherd decide <request-id> <option>` command is accepted;
- bot/app, edited, disallowed actor, unknown option, malformed, multiline, duplicate, ambiguous,
  hostile, stale generation/head/target, and expired responses never authorize a decision;
- silence, emoji/reactions, review text, CI/check state, and request-comment text are not approval;
- accepted evidence persists actor, URL, timestamp, and option, with no raw response body;
- issue and PR routes use the exact bound repository/number;
- typed `gh api` adapter uses argv execution and ambient host auth only, with bounded output and
  schema validation;
- live-comment coverage is skipped absent an explicitly designated sandbox.

## RED / GREEN / refactor evidence

Initial RED command:

```bash
node --test .pi/extensions/shepherd/human-decision.test.ts \
  .pi/extensions/shepherd/github-decision-broker.test.ts
```

Result: exit 1; 0 passed, 2 failed test-file entries. Both failed with
`ERR_MODULE_NOT_FOUND` for `.pi/extensions/shepherd/human-decision.ts`, proving the contract tests
precede both production modules. Duration: 66.299917 ms.

GREEN/refactor evidence remains pending.
