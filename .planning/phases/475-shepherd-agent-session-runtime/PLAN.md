# Issue #475 Plan — Scoped In-Process AgentSession Runtime

## Contract

- Primary issue: #475 (`feat(shepherd): add scoped in-process AgentSession runtime`)
- Parent issue / PR: #471 / #472
- Immutable base: `e659d6f1b666f58748e2d8c86599ceb4bbc62ff8`
- Branch / PR base: `feat/475-shepherd-agent-session-runtime` / `feat/471-pi-agent-session-shepherd`
- Worker directory: `/Users/karthiksivadas/Development/polymetrics-cli-agents/wt-475-shepherd-agent-session-runtime`
- Execution decision: `local_critical_path` — this is one isolated issue worker with an exclusive write scope; no nested worker is required or authorized.

## GSD And Skill Route

- Requested command: `scripts/gsd prompt programming-loop init --phase 475-shepherd-agent-session-runtime --dry-run`
- Result: adapter health passed, but the command registry returned `unknown GSD command: programming-loop`.
- Mode: `manual_gsd_fallback`; the universal plan → RED → GREEN → REFACTOR → verify lifecycle remains mandatory.
- Required skills loaded completely: `gsd-programming-loop`, `gsd-workstreams`, `gsd-plan-phase`, `github-issue-first-delivery`, `architecture-patterns`, `javascript-testing-patterns`.
- Required policy read: issue-agent contract, worker handoff, universal runtime loop, Pi adapter, runtime/RLM/Pi integration reference, required-skills routing, and task/security routing matrix.
- Architecture application: ports-and-adapters boundary around a fake-injected Pi SDK, opaque workspace, and typed host capabilities.
- Testing application: deterministic fake sessions, race tests, failure-path assertions, and strict RED evidence before production code.

## Owned Scope

Production:

- `.pi/extensions/shepherd/agent-session-runtime.ts`
- `.pi/extensions/shepherd/tool-policy.ts`
- `.pi/extensions/shepherd/role-prompts.ts`
- Shepherd-namespaced role prompt assets only when needed

Tests:

- `.pi/extensions/shepherd/agent-session-runtime.test.ts`
- `.pi/extensions/shepherd/tool-policy.test.ts`
- `.pi/extensions/shepherd/role-prompts.test.ts` when prompt-contract behavior warrants it

Durable issue memory:

- `.planning/phases/475-shepherd-agent-session-runtime/**`

All controller, domain, runner, SDK-runner, extension/index wiring, target-evidence, scheduler,
workspace/Git, GitHub, and shared parent artifacts are excluded.

## Design Boundaries

1. Define an injected Pi 0.80.6-compatible session factory port. Production wiring remains a later
   lane; this issue must make no subprocess, tmux, worktree, Git, GitHub, or network mutation.
2. Route implementation/correction workers only to `openai-codex/gpt-5.6-sol` with `high`;
   planning/research/review/validation/verification/orchestration only to the same model with
   `xhigh`. Reject caller overrides, 5.5, missing/unknown models, and terminal route drift.
3. Treat role input as untrusted data. Construct the system prompt from a trusted role template and
   immutable authority envelope; never let task text override issue, branch, workspace, tool,
   model, secret, recursion, or handoff authority.
4. Keep context bounded: bounded task/context/system prompt, in-memory sessions, no context files,
   extensions, skills, prompt templates, persistence, retries, or automatic compaction; bounded
   event/output accounting.
5. Create least-authority custom tools from an opaque workspace port and allowlisted typed host
   capability ports. Read-only roles get only read operations; mutating roles get workspace-bound
   read/edit/write plus explicitly declared typed capabilities. Generic shell, HTTP/SQL write,
   credentials/secrets, and orchestration/delegation tools are structurally unavailable.
6. Prevent recursive orchestration through role policy, reserved/forbidden tool namespaces, and a
   non-delegating system prompt.
7. Own session lifecycle with first-wins cancellation, timeout/deadline, abort, close, and parent
   shutdown; abort/wait/dispose and join are coalesced exactly once. Cleanup failure quarantines the
   runtime and prevents further dispatch.
8. Parse exactly one bounded JSON handoff. Validate a closed schema and exact
   `runId/generation/laneId/candidateHead/validationNonce` binding, redact secret-like material,
   bound arrays/text, and reject unknown fields or authority/tool/model claims.

## TDD Slices And Checkpoints

### PLAN checkpoint

- Commit these phase artifacts before tests or production edits.

### RED checkpoint

- Add fake-injected SDK/session tests for exact model/thinking routing and route drift rejection.
- Add tool-policy tests for workspace bounds, read-only mutation denial, typed-capability allowlists,
  forbidden generic/recursive tools, and prompt-injection authority expansion attempts.
- Add lifecycle tests for abort, timeout, explicit close, parent shutdown, late creation,
  abort/wait/dispose exactly once, teardown failure, quarantine, and concurrent join coalescing.
- Add structured handoff tests for closed schema, bounds, redaction, identity/head/nonce binding,
  recursion/authority fields, and malformed/unserializable/excess output.
- Run the focused tests and record the expected missing-module/failing behavior before production.

### GREEN checkpoint

- Implement the smallest role routing, prompt envelope, tool policy, lifecycle owner, and handoff
  validator that makes focused tests pass.
- Commit and push once the focused tests are green.

### REFACTOR checkpoint

- Remove duplication, harden invariants, and improve typed boundaries without widening authority.
- Run the entire Shepherd suite, strict TypeScript, offline smoke, diff, and scope checks.
- Commit/push final evidence updates separately.

## Verification Boundary

This worker runs only:

```bash
node --test .pi/extensions/shepherd/agent-session-runtime.test.ts \
  .pi/extensions/shepherd/tool-policy.test.ts
node --test .pi/extensions/shepherd/*.test.ts
<pinned Pi 0.80.6 TypeScript compiler> --noEmit --strict <owned and Shepherd TypeScript inputs>
printf '{"id":"commands","type":"get_commands"}\n' |
  PI_OFFLINE=1 pi --mode rpc --no-session --approve \
    --no-extensions --no-skills --no-prompt-templates --no-context-files \
    -e .pi/extensions/shepherd/index.ts
git diff --check
<owned-scope status/diff assertions>
```

The exact strict TypeScript command will use the compiler bundled with the installed pinned Pi
0.80.6 distribution or another already-installed compiler; no dependency will be added.

Do **not** run `go vet ./...`, `go test ./...`, `go build ./cmd/pm`, `make verify`, connector
certification, or any repository-wide Go gate. Those are parent-integration/GitHub-CI gates. This
boundary supersedes the generic repo verification block for this worker.

## Safety And Human Gates

- No dependency, auth-scope, secret, live credential, production, deployment, destructive, reverse
  ETL, Git/worktree, GitHub, shell, generic HTTP write, or generic SQL write action is authorized.
- No human gate is expected for the owned local implementation.
- Never merge this sub-PR or request Claude/Copilot review; independent exact-head Codex review is
  owned by the parent lane.

## Exact-Head Correction Cycle — `4e41c2ec1175a109c10f125203dc54d381b982bd`

Independent `codex_independent` review found two P1 boundary failures at the reviewed PR #486
head. This correction remains inside the original issue-owned runtime, tool-policy, tests, and
phase artifacts.

1. A `createAgentSession` promise that settles after both the request deadline and bounded cleanup
   interval can currently yield an unowned live session. Add a claimed/abandoned creation owner:
   normal execution claims the session, while a bounded waiter may abandon it and return; an
   attached continuation must then abort, wait for idle, and dispose every late session exactly
   once without extending request completion indefinitely.
2. Secret redaction currently misses quoted JSON/YAML-style assignments and quoted Bearer values.
   Extend the bounded redaction grammar while preserving unrelated prose, then prove the boundary
   through direct probes and through role prompts, tool results, and handoff summary/finding fields.

The correction uses a fresh strict RED → GREEN → REFACTOR cycle. Production edits remain blocked
until both regressions fail on the reviewed exact head. Execution remains `local_critical_path`:
all four runtime slots are occupied and both findings overlap this worker's exclusive source/test
scope, so nested delegation would add collision risk rather than an independent workstream.

Correction result: RED was captured at `93b9eca9`; the minimal GREEN implementation was committed
at `f788cf16`. Focused 24/24 and complete Shepherd 161/161 tests, both strict TypeScript scopes,
pinned offline Pi 0.80.6 RPC, diff, immutable-base, and changed-path checks pass. Repository-wide
Go, connector, and `make verify` gates remain outside this lane by explicit instruction.

## Exact-Head Correction Cycle 3 — `526dfec4282b442c4b32138ab036d4cc7e97b475`

Final `codex_independent` re-review found two remaining P1 boundaries. This cycle stays inside the
same issue-owned runtime, tool-policy, tests, and phase artifacts.

1. Replace line-limited whole-value redaction regexes with a bounded linear scanner over recognized
   structured assignments. It must cover multiline quoted YAML values, YAML literal/folded block
   scalars, common `client_secret` spellings, and multiline quoted Authorization Bearer values at
   direct, prompt, tool-result, and handoff boundaries. Unquoted ambiguous prose must remain
   byte-identical; only strong structured or single-token credential forms are redacted.
2. Bound abandoned-session abort and wait-for-idle against one cleanup deadline. Whether either
   hook never settles, force a coalesced unsubscribe/dispose exactly once, quarantine future
   dispatch, consume all rejections, and leave no ref'ed background timer or duplicate cleanup
   continuation.

The strict sequence remains PLAN → test-only RED → smallest GREEN → REFACTOR/verify, with a pushed
commit at each checkpoint. Execution decision is `local_critical_path`: both findings collide in
the issue-owned source/tests and the runtime thread cap rejected a read-only design sidecar. The
declared phase equivalent remains focused tests, complete Shepherd tests, strict pinned Pi 0.80.6
TypeScript, offline Pi RPC, and diff/changed-path checks only.

Cycle 3 result: PLAN was pushed at `896b30ae`; test-only RED was pushed at `9c4ed5fd` with 20 passes
and 7 expected failures; GREEN/refactor was pushed at
`d499e721a85abbe1a1d1be7fb0069649927c923c`. Focused 27/27 and complete Shepherd 164/164 tests,
both strict TypeScript scopes, explicit Pi 0.80.6 offline RPC, diff, immutable-base, and issue-owned
path gates pass. Fresh independent exact-head review and integration remain parent-owned.

## Exact-Head Correction Cycle 4 — `b4061d4e1a1545b0c8810b14b510cf048385a567`

Fresh `codex_independent` xhigh review found two P1 boundaries at the Cycle 3 evidence head. This
cycle remains restricted to the same issue-owned runtime, tool-policy, tests, and phase artifacts.

1. The foreground/main cleanup path still lets a timeout from `abortOnce()` or `joinOnce()` skip
   forced disposal. RED will cover a four-case matrix: session creation resolving inside the
   cleanup grace versus an ordinary session claimed before cancellation, each with either
   `abort()` or `waitForIdle()` never settling. Every case must settle within a test bound,
   quarantine and reject subsequent dispatch before another prompt, dispose exactly once, consume
   detached rejections, and produce no unhandled rejection. Waiting for idle may be skipped only
   after abort itself exceeds its bound; forced disposal is unconditional.
2. The structured redactor still misses unquoted sensitive keys in YAML flow mappings and
   line-start `client_secret` scalars whose values contain spaces. RED will prove direct,
   serialized-prompt, typed-tool-output, handoff-summary, and handoff-finding consumers. Synthetic
   markers must be absent, `[REDACTED]` present, and non-assignment prose controls byte-identical.
   The clarified contract treats a line-start `client_secret:` form as a structured assignment;
   the harmless control uses prose that is not itself valid assignment syntax.

The implementation target is one explicit cleanup pipeline whose abort and idle phases are bounded
independently and whose `finally`-style coalesced disposal always executes. The redactor will add
linear flow-mapping key discovery and a strong structured-value signal without rescanning input or
introducing broad multiline regexes. Strict order remains PLAN → test-only RED → smallest GREEN →
REFACTOR/verify, with a pushed checkpoint at each stage.

GSD adapter health still passes while the 69-command registry rejects `programming-loop`, so the
recorded `manual_gsd_fallback` remains active. Skills reloaded for Cycle 4:
`gsd-programming-loop`, `javascript-testing-patterns` plus its async/timer reference,
`typescript-advanced-types`, `architecture-patterns`, and `github-issue-first-delivery`, together
with required routing, issue-agent, universal-loop, Pi-adapter, and runtime/Pi guidance. Execution
decision is `local_critical_path`: both findings collide in the two issue-owned source/test modules,
and a read-only design sidecar was rejected by the runtime thread cap.

Cycle 4 result: PLAN was pushed at `190b0ec7`; test-only RED was pushed at `21535513` with 23
passes and 8 expected failures; GREEN/refactor was pushed at
`01b42ae168176956d864ff10f40d1c981f37ac04`. Focused 31/31 and complete Shepherd 168/168 tests,
both strict TypeScript scopes, explicit Pi 0.80.6 offline RPC, diff, immutable-base, and
issue-owned path gates pass. Two adversarial refactor probes each captured their own targeted RED
before production support. Fresh independent exact-head review and integration remain
parent-owned.

## Exact-Head Correction Cycle 5 — `e41f075a9b3bfb01d410296712740b54f943ba71`

Fresh `codex_independent` xhigh review found one lifecycle leak and three related lexical-state
failures. Cycle 5 remains inside the same issue-owned runtime, tool-policy, tests, and phase
artifacts; production is locked at the reviewed head until a committed test-only RED exists.

1. `run()` constructs `CancellationScope` before `#reserve()`. Duplicate-key, mutating-concurrency,
   and capacity rejection can therefore return without `finish()`, leaving a referenced deadline
   timer. Reservation must reject before scope construction, or every rejected path must prove the
   timer cleared/unref'ed.
2. A sensitive flow value that is itself a mapping can skip its opening delimiter, consume the
   nested close as the outer close, and hide a later unquoted `client_secret` sibling.
3. A leading unmatched apostrophe in ordinary prose can carry quote state across a newline and hide
   the next structured `client_secret` assignment.
4. Ordinary unmatched braces and flow-shaped comments can enter flow state and change later
   harmless assignment-like prose that must remain byte-identical.

The redactor correction is an architectural replacement of the assignment traversal, not another
global-regex or quote exception. Implement one explicit deterministic line/flow lexical state
machine with monotonic cursors, per-line YAML quote reset, comment/prose discrimination, and a
balanced delimiter stack when consuming nested flow values. Existing quoted, multiline, block,
Bearer, flow, spaced-scalar, and harmless-prose tests remain mandatory. RED covers nested-flow and
leading-apostrophe markers through direct, serialized-prompt, typed-tool-output, and handoff
summary/finding consumers, plus byte-identical brace/comment controls. A deterministic large-input
guard may be added only if it avoids timing-sensitive assertions.

Strict sequence remains PLAN → test-only RED → smallest GREEN → REFACTOR/verify, with a pushed
checkpoint at each stage. The declared phase equivalent remains focused tests, complete Shepherd
tests, pinned Pi 0.80.6 strict TypeScript, offline Pi RPC, and diff/immutable-base/owned-scope only.
No Go, connector, `make verify`, runtime-backed, live-GitHub, merge, or review-bot command is
permitted. GSD adapter health passes while its 69-command registry still rejects
`programming-loop`, so `manual_gsd_fallback` remains recorded. Required skills reloaded:
`gsd-programming-loop`, `javascript-testing-patterns` and its advanced timer guidance,
`typescript-advanced-types`, `architecture-patterns`, and `github-issue-first-delivery`, plus the
repo routing, issue contract, universal loop, Pi-adapter, and runtime/Pi references. Execution is
`local_critical_path`: both source findings overlap this worker's exclusive modules, and the
attempted read-only architecture sidecar was rejected by the runtime thread cap.

Cycle 5 RED result: the focused command exits 1 with 29 passes and 7 expected failures. The seven
failing tests isolate prompt serialization, handoff redaction, the referenced duplicate-run timer,
nested-flow direct redaction, leading-apostrophe direct redaction, byte-identical brace/comment
controls, and typed tool output. Focused strict TypeScript passes, and production remains unchanged
at the reviewed head.

Cycle 5 GREEN result: admission checks now precede `CancellationScope` construction, so every
duplicate/capacity/mutator rejection creates no deadline timer. The assignment transformer is one
typed lexical state machine with a monotonic cursor, explicit per-line quote reset, comment skips,
validated flow openers, a delimiter stack, and balanced nested-value consumption. The focused
suite passes 36/36 and focused strict TypeScript passes; complete verification remains pending.
