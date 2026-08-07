# Phase issue-3752-rate-limit-admission-r1 — provider-cited declarations and requester admission

## GSD setup

- Branch: `fm/cli-found-rate-limit-admission-r1`
- GSD preflight: `scripts/gsd doctor` passed on 2026-08-06.
- Contract validation: `go run ./cmd/agentcontractgen check` passed.
- Resolved commands: `discuss-phase`, `plan-phase`, `execute-phase`, `verify-work`, and `code-review` via `scripts/gsd sources <command>`.
- Prompt fallback executed inline: `scripts/gsd prompt discuss-phase 3752` and `scripts/gsd prompt plan-phase 3752 --tdd`.
- Manual-GSD fallback: Pi's interactive command runtime is unavailable here and the canonical contract forbids spawning GSD roles. The same decision, TDD, execution, verification, and review steps are recorded in this phase directory; no role or mutating subagent is spawned.

## Required skills loaded

- `golang-how-to`
- `golang-design-patterns`, `golang-structs-interfaces`, `golang-error-handling`
- `golang-security`, `golang-safety`, `golang-testing`
- `golang-context`, `golang-concurrency`, `golang-observability`
- `no-mistakes` for the later captain-directed validation/ship stage

## Scope and ownership

This is one shared-runtime foundation PR containing parent #3752 and its dependency #3751. It is
not a connector-migration lane: the exact connector target requirement is inapplicable because the
issues explicitly own the shared loader and requester contracts. The ownership evidence is the
issue tree and the paths below.

### Slice A — #3751, declaration contract (first)

1. Add optional `internal/connectors/defs/<connector>/rate_limits.json` to the engine meta-schema
   compilation set and typed loader. There is no production declaration in this slice; Go rejects
   an unmatched optional `//go:embed */rate_limits.json` pattern, so the first migration that adds
   a real declaration must add that explicit production embed pattern in the same change.
2. Add typed declaration structures and a `Bundle.RateLimits` loader result. Keep it absent/nil for
   every legacy bundle; do not change old `metadata.json.rate_limit` or `streams.json.base.rate_limit`.
3. Add the closed `rate_limits.schema.json` plus semantic validation:
   - root `schema_version: 1` and `state: declared|unknown|not_applicable`;
   - declared has at least one uniquely named policy; unknown/not_applicable have no policy and a
     nonblank reason;
   - each declared policy has a valid HTTPS provider-artifact URL and a valid ISO retrieval date,
     with provider version optional but never a substitute for the date;
   - endpoint selectors have an explicit all-target or valid HTTP method/rooted path, and optional
     tier/auth-type selectors are nonblank and unique;
   - scope subject kind is only `account`, `installation`, `application`, `endpoint`, or `ip`;
   - budgets model fixed-window, token-bucket, leaky-bucket, and cost units with positive,
     model-specific bounds; each labels itself burst or sustained.
4. Document the new optional file and non-secret scope rule in the connector authoring recipe.

### Slice B — #3752, requester core (second)

1. Add compact, consumer-owned `RateLimitAdmission` and `RateLimitObserver` interfaces plus
   typed request/observation values in `internal/connectors/connsdk`. They carry only safe
   transport metadata and typed timing/header facts; they do not log, persist, or derive a scope key.
2. Call admission immediately before each logical requester send: an outer `Client.Do` attempt or
   a permitted redirect hop in `Requester.doWithBody` and `Requester.DoStream`. A
   rejected/cancelled admission prevents that logical send. Safe replayable reads may replay inside
   `net/http` without another admission; strict non-idempotent writes suppress replays.
3. Parse deterministic `Retry-After` once against an injectable clock. Preserve its absolute reset
   time and exact delay. On 429, notify the observer even if the requester will not retry.
4. Return `*RateLimitError` for terminal HTTP 429 responses. It must wrap `*HTTPError`, preserve
   `errors.As` access to it, and expose the provider reset timestamp/delay without copying raw
   response headers or bodies into the observation.
5. Change fallback exponential retry to bounded full jitter with an injectable jitter function for
   deterministic tests. Do not call jitter for a valid provider reset, and do not cap that reset by
   `MaxBackoff`. Preserve retry counts, no-replay mode, multipart reopening, and context cancellation.

## Explicit deferred seams

- **#3753:** choose a declaration policy per runtime/request and attach a resolver implementing
  both new requester interfaces to checks, reads, all write forms, direct reads/writes, retries,
  and binary downloads. It also removes old page-only pacing. No activation is made here.
- **#3754:** implement the local/shared registry behind `RateLimitAdmission`. The future opaque key
  is `credential binding + policy ID + declared non-secret subject`; never a raw credential or
  secret-derived value.
- **#3755:** turn observations/admission waits into human/JSON operator events, deadline messages,
  CLI help/manual/website/inspect data. This branch has no CLI behavior change.

## TDD execution plan

| Order | Slice | RED evidence first | GREEN implementation | Refactor/guard |
| --- | --- | --- | --- | --- |
| A1 | `rate_limits.json` loader | table tests show an uncited policy, invalid retrieval date, malformed selector/budget, and invalid state are accepted or unavailable today | typed optional loader + closed schema + semantic checks | no legacy bundle becomes mandatory; embedded-file test |
| A2 | declaration expressiveness | load a valid endpoint/tier/auth policy with burst+sustained, points, and leaky bucket facts | retain typed values without flattening into requests-per-minute | ensure selector/scope fields cannot contain secret sources |
| B1 | provider wait defect | `Retry-After: 90` with `MaxBackoff: 30s` records a 30s sleep on current code | exact 90s wait and reset timestamp | retain ordinary fallback cap |
| B2 | fallback jitter | fallback exponential retry remains deterministic/uncapped by injection today | injected bounded full jitter and clamp untrusted injection to `[0, cap]` | valid `Retry-After` proves jitter callback was not invoked |
| B3 | logical-send admission | a context-cancelled admission has no requester seam today | admission cancellation returns before server hit for JSON, form, multipart, and stream methods; logical retries and redirects are re-admitted | safe replayable reads retain one admission through an internal replay; strict writes suppress replay; shallow requester clones retain the interface fields |
| B4 | observation and terminal typing | 429 currently returns only `*HTTPError` and emits no typed signal | typed observation + `RateLimitError` wrapping safe `HTTPError` | `errors.As`, retry cap, disabled-retry, and no-secret output checks |

The first executable RED test is B1 because it compiles against the pre-change public shape and
demonstrates the live defect without a provider call. Later test additions may initially fail to
compile until their contract type is introduced; each remains a red checkpoint in the ledger.

## Commit checkpoints

1. Plan/checklist checkpoint (this phase evidence).
2. Red defect-test checkpoint, retained locally until green implementation exists.
3. Green #3751 declaration loader/schema/documentation checkpoint.
4. Green #3752 requester/admission/observation checkpoint.
5. Review-fix checkpoint only if review finds a bounded in-scope defect.

No push, PR, or no-mistakes run occurs before the firstmate-directed handoff in the task contract.

## Verification plan

See `VERIFICATION.md`. The full suite is intentionally left to CI because the repository guidance
states `go test ./...` / `make verify` exceed ordinary per-command timeouts. Local gates are scoped
to changed packages plus individual `make verify` gates, and no gate is weakened.
