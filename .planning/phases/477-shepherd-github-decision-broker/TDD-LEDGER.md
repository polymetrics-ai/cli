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

GREEN/refactor command:

```bash
node --test .pi/extensions/shepherd/human-decision.test.ts \
  .pi/extensions/shepherd/github-decision-broker.test.ts
```

Initial GREEN after the aggregate, file repository, broker, fake transport, and typed `gh api`
adapter: 25 passed, 0 failed, 1 skipped (the live sandbox mutation), duration 794.79625 ms.

Refactor additions remained green while adding strict persisted-schema validation, no-follow state
reads, fsync+atomic replacement, dead-process lock recovery, concurrent request serialization,
bounded ten-page GitHub reads, duplicate-marker checks during polling, future-timestamp rejection,
and a request-specific transport write method rather than a generic comment-write surface.

Strict no-emit TypeScript for both production modules passed against the TypeScript/Node types
bundled with installed Pi 0.80.6.

Final-review hardening added support for a bot-authenticated host and ordinary `name[bot]` comments
without ever admitting them to the human actor allowlist. Final focused result: 27 tests total,
26 passed, 0 failed, 1 sandbox-only skip, duration 305.387208 ms.

Final full Shepherd result: 164 tests total, 163 passed, 0 failed, 1 sandbox-only skip, duration
48877.375042 ms. Strict no-emit TypeScript then passed over all 11 production Shepherd modules.

## Exact-head review correction cycle

Reviewed head: `87eb80f561d416da245e753a5dbc887a3384a05d`.

RED cases to add before correction production edits:

- GitHub timestamps such as `2026-07-21T12:30:12Z` normalize to canonical milliseconds and remain
  valid when a local creation time is within the bounded sub-second truncation window;
- request comment ID `5034006493`, repository `owner2/repo3`, and actor `maintainer2` are valid;
- a published transaction lock is a complete atomic owner record, dead owners are reclaimed, and
  release with an obsolete token cannot delete a replacement lock;
- `parent_merge` rejects `approve` and accepts only the exact affirmative `approve-merge` contract;
- credential-like `token=...`, JSON key/value, environment, URL-userinfo/query, private-key, and
  vendor-token forms are rejected without echoing their value;
- control/bidi/format characters and untrusted mentions cannot enter the durable question or alter
  the rendered Markdown structure; configured validated humans are explicitly mentioned;
- transient transport failures receive bounded backoff and retry, while permanent failures fail
  immediately without sleeping or exposing raw transport text.

Correction RED command:

```bash
node --test .pi/extensions/shepherd/human-decision.test.ts \
  .pi/extensions/shepherd/github-decision-broker.test.ts
```

Result before any correction production edit: exit 1; 39 tests total, 23 passed, 15 failed, and
1 designated-sandbox test skipped. The expected failures independently exposed the reviewed gaps:
second-resolution expiry/comment timestamps; safe-integer comment IDs; numeric repository/login
validation; generic parent-merge approval; unescaped Markdown/missing allowlist mentions;
credential/bidi/mention acceptance; directory-before-owner lock publication; unfenced replacement
deletion; regular-file dead-lock reclaim; unclassified transient/permanent transport errors; and
raw adapter error propagation. Duration: 282.318708 ms.

Correction GREEN/refactor evidence: pending.
