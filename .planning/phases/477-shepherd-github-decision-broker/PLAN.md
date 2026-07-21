# Plan: #477 Durable GitHub Human-Decision Broker

Issue: https://github.com/polymetrics-ai/cli/issues/477
Parent: #471
Parent PR: #472
Branch: `feat/477-shepherd-github-decision-broker`
Exact base: `e659d6f1b666f58748e2d8c86599ceb4bbc62ff8`
PR base: `feat/471-pi-agent-session-shepherd`

## Objective

Implement a durable, restart-safe GitHub issue-comment decision broker with exact repository,
target, generation, and head binding; marker-owned idempotent requests; bounded polling/backoff;
allowlisted exact human commands; atomic consume-once persistence; and a distinct exact-head
parent-merge approval gate.

## Workflow and required skills

- GSD mode: `manual_gsd_fallback` because
  `scripts/gsd prompt programming-loop init --phase 477-shepherd-github-decision-broker --dry-run`
  reports `unknown GSD command: programming-loop`; `scripts/gsd doctor` passes.
- Skills read completely: `gsd-programming-loop`, `gsd-workstreams`, `gsd-plan-phase`,
  `github-issue-first-delivery`, `architecture-patterns`, `javascript-testing-patterns`.
- Repository security routing read from
  `.agents/agentic-delivery/references/required-skills-routing.md` and
  `.agents/agentic-delivery/matrices/task-skill-matrix.yaml`.
- Go-specific skills are not applicable to the owned TypeScript implementation; root Go gates
  remain mandatory verification.
- Architecture: pure decision aggregate/repository port in `human-decision.ts`; typed GitHub
  transport and adapter in `github-decision-broker.ts`; fake transport/repository first.

## Owned slice

1. Define validated decision gates, targets, bindings, durable request/decision state, exact marker
   derivation, canonical routing, restart-safe file repository, and one-time consumption.
2. RED-test invalid gates/bindings, routing, persistence/restart, consume-once behavior, expiry, and
   secret rejection before implementing the aggregate.
3. RED-test marker idempotence, bounded polling/backoff, exact command parsing, actor/type allowlist,
   edited/bot/duplicate/ambiguous/hostile/stale responses, target separation, and silence/emoji/
   review/CI non-signals against issue-namespaced fake GitHub fixtures.
4. Implement the smallest broker and typed `gh api` adapter that use host authentication without
   accepting, persisting, or logging tokens.
5. Refactor validation and parsing while preserving typed port boundaries; do not edit controller,
   domain, runner, SDK runner, target evidence, extension wiring, index, or shared parent artifacts.
6. Run focused tests, the full Shepherd suite, strict no-emit TypeScript against installed Pi
   0.80.6 types, Pi extension discovery, diff/root Go/build/full verification, then update all
   phase evidence.
7. Commit and push coherent plan, RED, GREEN/refactor, and verification checkpoints. Open a ready
   stacked PR with a Conventional Commit title and `Refs #477` plus `Refs #471`; do not merge or
   request Claude/Copilot.

## Safety and threat model

- Reject tokens/credential-shaped text before it can enter request comments or durable state.
- The transport accepts typed repository/target inputs only and invokes `gh` directly without a
  shell or token parameter.
- Exact command grammar is `/shepherd decide <request-id> <option>`; all other comment/review/check/
  reaction content is non-authoritative.
- Requests are bound to one canonical repository, target, generation, and (for PR gates) 40-hex
  head. Current binding must match before polling, deciding, or consuming.
- Only unedited comments from allowlisted `User` actors may decide. Multiple valid responses are
  ambiguous and fail closed. A decision records only option, actor, source URL, and timestamp.
- A marker collision, duplicate marker, malformed GitHub payload, unbounded response, timeout,
  expiry, or storage conflict fails closed.
- Live-comment tests are skipped unless an explicit designated sandbox is configured; this issue
  performs no live comment mutation during local verification.

## Verification

```bash
node --test .pi/extensions/shepherd/human-decision.test.ts \
  .pi/extensions/shepherd/github-decision-broker.test.ts
node --test .pi/extensions/shepherd/*.test.ts
# strict no-emit TypeScript command resolved against installed Pi 0.80.6 types
pi --list-extensions
git diff --check
go vet ./...
go test ./...
go build ./cmd/pm
make verify
```

`pi --list-extensions` is executed exactly as required even though existing parent evidence says Pi
0.80.6 may reject that flag; any rejection is recorded rather than concealed. The offline
registration probe is an additional fallback, not a substitute for reporting the exact command.

Parent-policy update during verification: the parent orchestrator superseded child-lane full-repo
verification with the declared child equivalent of focused tests, the full Shepherd suite, strict
Pi 0.80.6 TypeScript, offline Pi RPC/extension discovery, and `git diff --check`. It intentionally
terminated the in-flight `make verify` attempt and instructed this worker not to retry. The earlier
standalone Go gates remain supplemental evidence only.

## Exact-head correction cycle: 2026-07-21

Independent `codex_independent` review of exact head
`87eb80f561d416da245e753a5dbc887a3384a05d` found seven blockers and one warning. This correction
cycle remains inside the original two production files, matching tests/fixtures, and this phase
directory.

Planned RED slices:

1. Accept and canonicalize GitHub's RFC3339 second-resolution timestamps while allowing only the
   sub-second chronology tolerance needed when GitHub truncates a local millisecond timestamp.
2. Preserve safe-integer GitHub comment IDs above signed 32-bit range, and accept digits in valid
   repository names and human logins.
3. Replace directory-before-owner locks with atomically published owner records; carry the lock
   token through acquire/reclaim/release and fail closed rather than deleting a replacement lock.
4. Make `parent_merge` a discriminated request whose only affirmative option is the literal
   `approve-merge`; reject generic `approve` at both type and runtime boundaries.
5. Reject centralized-style credential key/value forms, URLs with credentials, private keys,
   vendor tokens, control/format/bidi characters, and untrusted mentions. Render the question as
   escaped quoted Markdown and mention only validated configured humans from a dedicated field.
6. Classify GitHub transport failures as transient or permanent, redact adapter-facing errors, and
   retry only transient failures with an independently bounded exponential backoff.

Execution decision for this correction cycle: `local_critical_path`. The assignment is an exact
review-fix lane with one owned write scope; another mutating worker would collide with the same two
modules and invalidate the requested RED-before-GREEN sequence. Fresh xhigh review remains
parent-owned after a new exact head is pushed.

Correction verification is limited by coordinator policy to focused #477 tests, the full Shepherd
suite, strict no-emit TypeScript using the explicitly pinned Pi 0.80.6 installation, offline Pi RPC,
diff/scope/base checks, and PR evidence updates. Go, connector, `make verify`, live GitHub comments,
Claude/Copilot requests, and merge remain out of scope.
