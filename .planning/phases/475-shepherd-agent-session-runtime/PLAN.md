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
suite passes 36/36 and focused strict TypeScript passes; the terminal result is recorded below.

Cycle 5 terminal result: PLAN was pushed at `8087b539`; test-only RED was pushed at `333c7ad6`
with 29 passes and 7 expected failures; GREEN/refactor was pushed at
`8ff2d9631809d09db26811b4cd1335b92a9c457c`. Focused 36/36 and complete Shepherd 173/173 tests,
both strict TypeScript scopes, explicit Pi 0.80.6 offline RPC, diff, immutable-base, and
issue-owned path gates pass. Fresh independent exact-head review and integration remain
parent-owned.

## Exact-Head Correction Cycle 6 — `d918617a19749cd16d6bfcf3d2fee3e5146e7380`

Fresh `codex_independent` xhigh review found three remaining text-transformer invariants. The
admission and lifecycle implementation is clean and remains unchanged. Cycle 6 stays inside the
issue-owned tool-policy source/tests, runtime consumer tests, and phase artifacts; production is
locked at the reviewed head until a committed test-only RED exists.

1. An unquoted sensitive value containing a nested mapping can cross a newline. The value-local
   closer stack currently stops at the first line boundary, so the global flow stack consumes the
   nested close as the outer close and a later same-line `client_secret` sibling can leak.
2. The apostrophe in the unquoted scalar `rock-'n-roll` is treated as a quote opener because it
   follows `-`. That false quote state can hide a later sensitive sibling.
3. Every assignment recomputes its line end from the value start. Many assignments on one flow
   line therefore rescan the remaining suffix and make the claimed monotonic scanner quadratic.

The smallest correction preserves the typed lexer. A value-local balanced closer stack may span
line boundaries and owns only the exact nested value, leaving the outer flow stack authoritative.
Quote opening becomes token-context-aware: an apostrophe inside an unquoted word is ordinary text,
while line-start, assignment, flow-delimiter, and YAML-list quote boundaries remain supported.
Assignment decisions receive the scanner's current line end; line discovery advances monotonically
instead of rescanning each suffix. An optional typed diagnostics sink counts line-boundary byte
visits so 25/50/100 KiB single-line flow inputs can assert bounded near-linear work without wall
clock thresholds.

The seven expected RED failures cover multiline-nested and punctuation-apostrophe markers through
direct calls, serialized prompts, `workspace_read`, typed capability output, and handoff
summary/finding consumers, plus the deterministic scale guard. The harmless
`{ flavor: rock-'n-roll, safe: retained }` control must remain byte-identical, and all 36 existing
focused cases remain mandatory. Strict order is PLAN → test-only RED → smallest GREEN → REFACTOR/
verify, with pushed checkpoints.

The declared phase equivalent remains focused tests, complete Shepherd tests, pinned Pi 0.80.6
strict TypeScript, offline Pi RPC, and diff/immutable-base/owned-scope only. No Go, connector,
`make verify`, runtime-backed, live-GitHub, merge, or review-bot command is permitted. GSD adapter
health passes while its 69-command registry rejects `programming-loop`, so
`manual_gsd_fallback` remains recorded. Required skills reloaded: `gsd-programming-loop`,
`javascript-testing-patterns`, `typescript-advanced-types`, `architecture-patterns`, and
`github-issue-first-delivery`, plus required routing, issue contract, universal runtime loop,
Pi-adapter, and runtime/Pi references. Execution is `local_critical_path`: all findings overlap
the same issue-owned scanner/consumer tests, and the attempted read-only architecture sidecar was
rejected by the runtime thread cap.

Cycle 6 RED result: the focused command exits 1 with 33 passes and 7 expected failures. Prompt,
handoff, `workspace_read`, and typed-capability consumer tests retain a multiline-nested or
punctuation-apostrophe marker; two direct tests isolate those lexical failures; and the scale guard
reports the absent deterministic line-boundary metric. The safe apostrophe control and all prior
cases pass. Focused strict TypeScript passes, and production remains unchanged at the reviewed
head.

Cycle 6 GREEN result: the value-local closer stack now advances across line endings while nested,
quote opening accepts `-` only as a line-local YAML sequence marker, and assignment decisions reuse
the scanner-owned line end. A typed optional diagnostics sink reports exact line-boundary visits of
25,618 / 51,218 / 102,418 for inputs of the same sizes. The overloaded redaction entry point ignores
`Array.map`'s numeric callback index, preserving every existing callback consumer. Focused tests
pass 40/40 and focused strict TypeScript passes; the terminal result is recorded below.

Cycle 6 terminal result: PLAN was pushed at `4f9c5a96`; test-only RED was pushed at `e8422d53`
with 33 passes and 7 expected failures; GREEN/refactor was pushed at
`93314a54302e84e053ad0d6ff44371fbf1a167e0`. Focused 40/40 and complete Shepherd 177/177 tests,
both strict TypeScript scopes, explicit Pi 0.80.6 offline RPC, diff, immutable-base, and
issue-owned path gates pass. Fresh independent exact-head review and integration remain
parent-owned.

## Stable-Head Correction Cycle 7 — `a3cd85a5d0871dd1c4c99dd8b30bcd609a228c45`

The combined stable-head campaign recorded 11 actionable findings (8 P1, 3 P2) against PR #486 at
<https://github.com/polymetrics-ai/cli/pull/486#issuecomment-5037079867>. The immutable comparison
base remains `e659d6f1b666f58748e2d8c86599ceb4bbc62ff8`; parent forensics and stable-head policy are read
from parent commit `2a89142e` without merging that branch or editing its shared artifacts. Cycle 7
keeps both issue-owned production modules byte-identical until one complete synthesized behavior
RED is committed and pushed.

### Lifecycle state-and-handle matrix

The lifecycle RED must execute six independent rows rather than rely on file-load, compile, or
module-missing failures:

| Phase / trigger | Expected ownership invariant | Required accounting |
|---|---|---|
| admitted run / external-signal attach throws | the run fails but reservation and scope finalize; close settles | no referenced deadline timer; no stranded active run |
| admitted run / external-signal removal throws | listener failure cannot skip scope finish or reservation release; close settles | no referenced deadline timer; cleanup remains exactly once |
| abandoned create / close then late resolve | close cannot report success before the created session is validated and cleaned | prompt 0; abort/wait/dispose 1 each; no unhandled rejection |
| abandoned create / close then late reject | close waits for terminal creation rejection before successful completion | hooks 0; no timer, reservation, or unhandled-rejection residue |
| abandoned create / close while create hangs | close is bounded and rejects/quarantines instead of succeeding or hanging | no hooks, referenced timer, or unhandled rejection; later dispatch fails closed |
| abandoned create / malformed late fulfillment | malformed SDK output is consumed and quarantines close without a detached-chain rejection | hooks 0; exactly one terminal ownership outcome; no unhandled rejection |

The architectural correction will give every admitted run an exception-safe listener lease whose
finalizer cannot be skipped by an untrusted `AbortSignal` hook. Session creation becomes a tracked
runtime-owned resource with an explicit terminal promise covering rejection, validation, and late
cleanup. `close()` joins those ownership records within a bounded close deadline: it succeeds only
after every owned late operation is terminal, while an uncancellable pending creation produces a
bounded quarantine rejection and retains a consumed continuation for any eventual result. Late
fulfillment is validated before session ownership is constructed; all detached continuations are
total and rejection-consumed.

### Redaction syntax, consumer, preservation, and work matrix

The redaction RED must cover multiline outer flow state, indented assignments, key-only and
continued YAML plain scalars, numeric sensitive values, Basic and other non-Bearer Authorization
schemes, unmatched-quote recovery, repository vocabulary aliases, and generic PKCS#8
`BEGIN PRIVATE KEY` blocks. One harmless multiline quoted scalar containing assignment-shaped
documentation must remain byte-for-byte identical. A shared compact adversarial payload is checked
at the direct transformer, serialized prompt, `workspace_read`, typed capability output, and
handoff summary/finding boundaries; each consumer reports the complete leaked-marker set rather
than stopping at its first marker.

The correction remains one typed lexical architecture. Scanner state will retain only
structurally-originated multiline quotes and flow delimiters; ordinary leading quote prose still
recovers at a line boundary. YAML value ownership uses indentation so key-only and continued plain
scalars redact their full owned block, while the next sibling remains parseable. Authorization is
sensitive independent of its authentication scheme, sensitive keys do not treat numeric scalars as
public, repository aliases use the same normalized vocabulary as Shepherd's existing secret-path
and environment contracts, and private-key recognition accepts the empty algorithm label used by
generic PKCS#8. Unmatched sensitive quotes fail closed only for their owned line/block and resume at
the next structural sibling.

Deterministic diagnostics will count all scanner character work, including structured-key and
leading-indentation discovery, for proportionally padded 25/50/100 KiB flow inputs. The test asserts
nonzero bounded near-linear work and approximate doubling, never wall-clock time. The implementation
will cache or advance line/key metadata monotonically so repeated assignments cannot rescan the same
padding outside the metric.

Cycle 7 follows PLAN -> one test-only RED -> one architectural GREEN/refactor -> declared verify,
with each checkpoint committed and pushed. Expected RED is 40 retained passes plus 13 independent
behavior failures: two signal-listener rows, four creation/close rows, a direct secret matrix, a
safe multiline preservation control, four serialized consumer rows, and one deterministic padded-
flow work row. Focused and complete Shepherd tests, both pinned Pi 0.80.6 strict TypeScript scopes,
offline RPC, diff, immutable-base/head, and issue-owned scope are the only permitted final gates.
No Go, connector, certification, `make verify`, runtime-backed, live-GitHub, review-bot, merge, or
parent-artifact command is permitted.

GSD adapter health passes while its 69-command registry still rejects `programming-loop`, so the
recorded `manual_gsd_fallback` remains active. Required skills reloaded for this cycle are
`gsd-programming-loop`, `javascript-testing-patterns`, `typescript-advanced-types`,
`architecture-patterns`, and `github-issue-first-delivery`, plus repository routing, the issue
contract, universal runtime loop, Pi adapter, and runtime/Pi integration guidance. Execution is
`local_critical_path`: the attempted read-only lifecycle sidecar was rejected by the four-thread
runtime cap, and all findings collide in the two issue-owned runtime/redaction modules and their
consumer tests.

Cycle 7 RED result: PLAN was pushed at `f40a08f1`; the one test-only RED was pushed at
`3b7e886a`. The focused command executed all 53 tests and exited 1 with 40 retained passes and 13
intended assertion failures. Both production blobs remained byte-identical to frozen candidate
`a3cd85a5d0871dd1c4c99dd8b30bcd609a228c45`; there was no compile, module, file-load, or
unrelated failure.

Cycle 7 GREEN/refactor result: the runtime now gives admitted runs an exception-safe external-
signal lease and registers each creation owner's terminal promise until rejection, validated late
cleanup, or bounded close quarantine. Detached fulfillment is total and rejection-consumed. The
redactor carries structural multiline quote/flow state, owns YAML continuation by indentation,
recognizes the repository alias vocabulary and generic PKCS#8, treats all credential-bearing
Authorization schemes and numeric sensitive values safely, and recovers after unmatched sensitive
quotes. Scanner line/key state advances monotonically with complete deterministic metrics. Focused
tests pass 53/53 and focused strict TypeScript passes against explicit Pi 0.80.6 types. Terminal
full-suite/RPC/diff/base/head/scope evidence remains pending.

Cycle 7 terminal result: GREEN/refactor was pushed at
`5c638d7f21a3910f40e499dba5c82cb7646642ac`. Focused tests pass 53/53 and the complete Shepherd
suite passes 190/190. Focused and all-production strict TypeScript pass with TypeScript 5.9.3
against the explicit Pi 0.80.6 package/type roots; the explicit Pi 0.80.6 offline RPC registers
`pm-shepherd`. Diff, immutable-base, pushed-head equality, and issue-owned path checks pass. Fresh
stable-head review and integration remain parent-owned.

## Stable-Head Correction Cycle 8 — `f219b730c63adc9188c93093a40511433a3d0110`

Cycle 8 batches the deduplicated lifecycle and security/parser review findings against frozen PR
#486 head `f219b730c63adc9188c93093a40511433a3d0110`; the immutable comparison base remains
`e659d6f1b666f58748e2d8c86599ceb4bbc62ff8`. The security/parser source is
`/tmp/475-REVIEW-SECURITY-CYCLE7.md`; the parent-provided lifecycle disposition identifies mutable
signal rereads, ambiguous `undefined` failure state, unawaited thenable cleanup, repeated request
accessors, and cleanup timeouts above Node's timer ceiling. Production stays byte-identical to the
frozen head until one complete behavior-level test-only RED is committed and pushed.

### Lifecycle, request snapshot, and bound matrix

The RED checkpoint will exercise independent assertion-level rows for:

1. a signal target whose `addEventListener` attaches and then throws, whose `removeEventListener`
   throws, whose parent listener is removed, and whose request signal accessor/target mutates after
   admission; the exact acquired listener lease must be finalized without rereading attacker state;
2. cleanup and validation throws/rejections whose reason is literally `undefined`, using explicit
   failure presence rather than an `undefined` sentinel;
3. promise/thenable-returning session `dispose` and subscription `unsubscribe`, including rejection,
   proving cleanup and quarantine wait for the owned terminal operation;
4. one normalized, immutable request snapshot covering role/task/context/authority/workspace/
   binding/signal, hostile getters, mutate-after-reload attempts, and mutator-fence checks so cwd,
   head, prompt, tool authority, and lane identity cannot drift;
5. one-above-ceiling rejection for every configurable size, count, request timeout, cleanup timeout,
   and concurrency option, with all timer ceilings at or below Node's maximum supported delay;
6. cycle-safe, bounded event accounting that rejects oversized/deep/cyclic input before unbounded
   `JSON.stringify` or terminal-output materialization;
7. one canonical normalized prefix set shared by tool construction, prompt authority, and handoff
   validation; and C0/C1 terminal controls rejected or neutralized in every handoff string field.
8. bounded parallel mutator leases: two immutable, canonically disjoint issue/branch/workspace/write
   authorities may run up to `maxConcurrency`, an overlapping authority is denied, and completing
   one run releases only its own lease while another disjoint lease remains fenced.

### Structured redaction matrix

One compact adversarial payload and focused direct controls will cover line-context comma-bearing
secrets, Digest/Signature/AWS-style Authorization auth parameters, key-only and continued unquoted
scalars inside multiline flow collections, and bounded decoding of escaped double-quoted JSON/YAML
keys (`client\\u005fsecret`, `to\\u006ben`) plus YAML doubled single quotes. Malformed escaped keys
with a secret-looking prefix fail closed. The shared payload traverses direct redaction, serialized
prompt, `workspace_read`, typed capability/mutation output, and handoff summary/finding/verification
consumers; harmless controls and all 53 prior focused tests remain mandatory. Scanner and event work
are asserted through deterministic counters/early exits, never wall-clock thresholds.

### Architectural GREEN and checkpoints

After the pushed RED only, implement one cohesive correction: normalize and deeply freeze the
request and canonical authority once; acquire an explicit listener lease with a captured cleanup
target; represent failure as a discriminated/presence state; assimilate and await cleanup thenables;
validate central hard limit constants; keep the structured redactor monotonic while decoding keys
within a 64-character bound; estimate event size with bounded cycle-safe traversal before any
serialization; reuse the canonical prefix set everywhere; replace the singleton mutator flag with a
bounded authority/scope lease map whose canonical collision predicate admits only disjoint isolated
work; and return only terminal-safe handoff text. No dependency or authority expansion is permitted.

Strict order is PLAN -> one test-only RED -> one architectural GREEN/refactor -> declared verify,
with each checkpoint committed and pushed. RED must compile under strict pinned Pi 0.80.6 types,
execute every focused test, and fail only intended assertions while the production blobs match
`f219b730`. Final gates are focused tests, the serialized complete Shepherd suite, focused and
all-production strict TypeScript 5.9.3 against explicit Pi 0.80.6 package/type roots, explicit Pi
0.80.6 offline RPC registration, and diff/base/head/issue-scope checks. Go, connectors,
certification, `make verify`, runtime services, live GitHub/CI/review bots, merge, and shared parent
artifacts remain forbidden.

GSD adapter health passes while its 69-command registry still rejects `programming-loop`, so
`manual_gsd_fallback` remains active without weakening TDD. Skills reloaded: `gsd-programming-loop`,
`javascript-testing-patterns`, `typescript-advanced-types`, `architecture-patterns`, and
`github-issue-first-delivery`, plus required routing, issue contract, universal runtime loop, Pi
adapter, and runtime/Pi guidance. The Cycle 8 plan decision is `read_only_spawned`: a read-only
lifecycle sidecar maps the exact code/test seams while this isolated worker retains the single
mutating critical path.

### Cycle 8 execution result

PLAN `9dd71a812795b7ac74b07db06c4fae03a3004871` and its pre-RED mutator-lease amendment
`04dc72f31a3bdd461045a4ef12d92c260f8ffd3f` preceded test-only RED
`11aa221231a52fab91f41dfce9742b7dfe180c02`. All 70 focused tests loaded and executed: the 53
retained tests passed and exactly 17 Cycle 8 behavior assertions failed. Focused strict TypeScript
passed, and the three production blobs remained byte-identical to frozen head `f219b730`.

The cohesive GREEN/refactor is `c4d34c377532c903238400c986a6b488fab3646d`. All 70 focused tests
pass. Focused and all-production strict TypeScript 5.9.3 pass against the explicit Pi 0.80.6
package/type roots, and the explicit Pi 0.80.6 offline RPC registers `pm-shepherd`. Diff,
immutable-base/frozen-head ancestry, and issue-owned path checks pass.

Two external terminal gates remain blocked rather than passed. The complete 207-test Shepherd run
reports 176 passes and 31 failures because this managed sandbox denies the pre-existing Darwin
process-identity probe at `state-store.ts` with `spawn EPERM`; per-file isolation identifies only
`controller.test.ts` and `state-store.test.ts` as affected, while every other Shepherd test file
passes. The required push was attempted after GREEN and failed before remote contact with
`ssh: Could not resolve hostname github.com: -65563`. Parent orchestration must rerun the complete
suite in an environment that permits `/bin/ps` and push the local commit chain before requesting
fresh exact-head review.

## Consolidated Stable-Head Correction Cycle 9 — `0cdcda7e049b7ecfa2fdc52027c66c5de161f2c8`

Cycle 9 treats `/tmp/475-REVIEW-CYCLE8-1.md` and `/tmp/475-REVIEW-CYCLE8-2.md` as one
deduplicated contract against the exact clean reviewed candidate above. The immutable comparison
base remains `e659d6f1b666f58748e2d8c86599ceb4bbc62ff8`. This issue worker owns only the three #475
production modules, their focused tests, and this phase's artifacts. Parent, #478, #479, Go,
connectors, certification, `make`, live GitHub, merge, and dependencies remain outside authority.

### Cycle 9 behavior matrix

The single test-only RED checkpoint will execute every retained test plus independent behavior rows
for the following deduplicated invariants:

1. A creation result whose `session`, extension result, extension/error arrays, or fallback fields
   alternate or throw is captured exactly once into one closed immutable result. Foreground and
   abandoned creation validate, prompt, abort, wait, unsubscribe, dispose, and report only through
   the same owned session; an alternate session receives zero calls.
2. The exact active-tool oracle is a private frozen copy. Pi receives distinct immutable tool-name
   and custom-tool arrays, so push/splice/reorder/replacement attempts cannot alter the oracle or
   validate `bash` or any other forbidden tool.
3. Capability parameter schemas are data-only, cycle-safe, accessor-free deep snapshots before the
   first await/SDK call. Depth, node, key, array-length, and incremental byte ceilings reject
   accessor, symbol, proxy, sparse, wide, deep, cyclic, or non-JSON graphs before serialization.
   Capability and workspace read/edit/write results are captured once into immutable DTOs before
   validation, redaction, or rendering.
4. Reload and creation settlement use the explicit union `fulfilled | rejected | pending`.
   Ordinary settled reload/create rejection is the primary retryable run error, releases every
   timer/run/mutator/creation lease, consumes late rejection, permits another dispatch, and never
   quarantines. Pending, malformed fulfillment, or actual cleanup failure remains fail-closed and
   quarantining.
5. Unsubscribe and dispose are independent exactly-once owned operations. Each has its own bounded
   cleanup phase; dispose is attempted even after unsubscribe timeout. Normal, abort, deadline,
   close, parent shutdown, and abandoned creation settle within declared bounds, consume late
   rejection, release leases/timers, and quarantine on cleanup timeout/failure.
6. Signal leases capture the exact add/remove operations at acquisition and retain native
   `EventTarget` fallback detach. Caller method mutation is irrelevant; a remover throwing before
   detach still gets a reliable fallback detach while its cleanup failure remains observable and
   reservations are released. Request and parent variants share the same invariant.
7. Every public asynchronous boundary—admission/getters, SDK lookup/setup/reload/create, request
   signal attach/release, `run`, `abort`, `close`, shutdown, and parent release—rejects with
   `AgentSessionRuntimeError` containing an own `cause`, including literal `undefined`. Primary and
   cleanup failures are deterministically aggregated rather than overwritten.
8. Listener delivery parses only known terminal event kinds into bounded immutable DTOs. It checks
   type, own-key count, array length, descriptors, scalar sizes, and content incrementally; rejects
   proxies, accessors, symbols, sparse/non-enumerable surprises, mutation-after-delivery, and wide
   graphs before spreads, descriptor maps, `filter`, `reverse`, or raw-reference retention.
9. The shared redactor closes column-zero `token/password/secret =` multiword assignments, opaque
   one-token Authorization, URL userinfo/query credentials, YAML implicit flow pairs, malformed or
   mid-escaped keys, and the full 63/64/65 decoded-key boundary including all-`\\uXXXX` encoding.
   Every direct/prompt/workspace/capability/mutation/handoff consumer is covered while harmless
   syntax-aware prose remains byte-identical.
10. Root-scoped reads deny common credential-bearing registry/package-manager, netrc, Git
    credential, Kubernetes, cloud-provider, and container/Docker authentication paths, including
    nested and case variants, before invoking the workspace callback.
11. Host capability names are tokenized. Sensitive-data nouns combined with acquisition/read/list/
    get/export verbs in either order, plural, or alias form are absent for both mutating and
    read-only roles; generic transport/orchestration denials remain intact.
12. Every handoff summary, finding, verification name, and verification summary rejects raw HT,
    LF, CR, CRLF, C0/C1, Unicode line/paragraph separators, and bidi formatting controls.
13. The Pi boundary uses exported Pi 0.80.6 `ToolDefinition`/`AgentToolResult` plus TypeBox
    `TSchema` directly, always supplies `details`, contains no `unknown` cast hiding custom tools,
    and has an offline no-model argument-validation/result exercise.
14. Cycle 8 disjoint mutator aliases/capacity/per-lease cleanup, all prior lifecycle/parser rows,
    no referenced timers, and no unhandled rejections remain green.

### Cohesive GREEN architecture

The implementation will introduce four bounded ownership primitives rather than isolated patches:

- a normalized creation-result/session owner plus discriminated async settlement;
- independent bounded cleanup operations and captured listener operations with fallback detach;
- one reusable bounded data snapshot/closed-field reader for schemas, tool results, and terminal
  events, producing deeply frozen DTOs without whole-graph serialization;
- directly typed frozen Pi tool definitions backed by one private expected-name oracle.

The structured redactor and path/capability classifiers will be extended in their existing shared
policy boundary so all consumers inherit the fix. Handoff text will remain single-line terminal-
safe data. No authority or dependency expansion is permitted.

### Checkpoints and declared verification

Strict order is artifact-only PLAN -> one behavior-level test-only RED -> one architectural
GREEN/refactor -> evidence. RED must load and run every focused test under strict TypeScript,
fail only the new behavior assertions, and prove the three production blobs exactly match
`0cdcda7e`. GREEN verification comprises focused runtime/tool-policy tests, serialized complete
Shepherd tests with the known managed-sandbox `/bin/ps` `spawn EPERM` recorded separately,
focused and all-production strict TypeScript 5.9.3 against explicit Pi 0.80.6 roots, explicit Pi
0.80.6 offline RPC plus a no-model tool exercise, `git diff --check`, immutable-base/frozen-head
ancestry, clean head, and issue-owned path scope.

`scripts/gsd doctor` passes, but the healthy 69-command registry rejects
`scripts/gsd prompt programming-loop ...`; Cycle 9 therefore records the required
`manual_gsd_fallback` without weakening TDD. Loaded skills are `gsd-programming-loop`,
`javascript-testing-patterns`, `typescript-advanced-types`, `architecture-patterns`, and
`github-issue-first-delivery`, plus required routing, Pi adapter, universal runtime loop, issue
contract, and runtime/Pi guidance. Execution decision: `read_only_spawned`; a read-only seam mapper
supports this plan while the isolated #475 worker retains the only mutating path.

The completed read-only delegation is recorded in `AGENTS.md`, `agents/cycle9-seam-map.md`, and
`traces/cycle9-seam-map-trace.md`. Its two binding design corrections are now part of this PLAN:
known terminal DTOs must avoid generic hostile-Proxy enumeration, and Pi tools must use public Pi
0.80.6 types plus supported plain-JSON-schema compatibility without a transitive TypeBox import or
new dependency.

Cycle 9 execution followed the required checkpoints: PLAN `b175cc4a`, read-only seam trace
`7047a8f4`, one test-only RED `dbf796b3`, and cohesive GREEN/refactor `94918f4e`. GREEN preserves
all 70 Cycle 8 focused regressions and passes all 16 new behavior rows. Verification is recorded in
`TDD-LEDGER.md` and `VERIFICATION.md`; the only complete-suite failures are the unchanged managed-
sandbox denial of `/bin/ps` in parent-owned state-store/controller tests.

## Consolidated Stable-Head Correction Cycle 10 — `f63957aed6fd1406eb3bd9a82adbd10b23b34c33`

Cycle 10 treats the complete `/tmp/475-REVIEW-CYCLE9-1.md` and
`/tmp/475-REVIEW-CYCLE9-2.md` reports as one binding contract, including WR-01. The immutable
comparison base remains `e659d6f1b666f58748e2d8c86599ceb4bbc62ff8`. Production is frozen at
the exact clean Cycle 9 evidence head until one comprehensive test-only RED commit exists. The
frozen production blob IDs are:

| Production path | Frozen blob |
|---|---|
| `.pi/extensions/shepherd/agent-session-runtime.ts` | `03cf916b59ef291dab309e6251a6f10ebf897eb0` |
| `.pi/extensions/shepherd/tool-policy.ts` | `1c8f701091a49c60cf41f83a6c16f2ae49a896c3` |
| `.pi/extensions/shepherd/role-prompts.ts` | `cfc2d253c323ad01f34b8c9688b3bad0acd16171` |

### Cycle 10 binding invariants

1. A genuine request or parent `AbortSignal` always acquires and releases the canonical native
   `EventTarget` listener lease. Captured add/remove hooks remain observable for typed failure
   reporting, but a silent no-op, mutation, or throw cannot defeat native cancellation or detach.
2. Returned-session ownership is staged. A minimal exact cleanup capsule captures every available
   cleanup operation before full operational/result validation. Any later missing, throwing, or
   malformed surface in foreground or abandoned fulfillment runs bounded exactly-once cleanup;
   successful forced cleanup is not quarantine, while an actual cleanup failure is.
3. Every timeout used by detached late cleanup—abort, idle, unsubscribe, and dispose—is unreferenced;
   foreground awaited bounds remain referenced and useful.
4. Close, shutdown, and coalesced close calls distinguish an uncancellable pending creation from a
   cleanup pipeline already in progress. A never-settling creation remains bounded, but once cleanup
   begins no shorter outer one-phase deadline may return before its internally bounded terminal.
5. The SDK creation result, `extensionsResult`, `extensions`, `errors`, and fallback field are
   captured once through data descriptors. Only closed, non-proxy, canonical dense empty extension
   arrays are valid in foreground and late paths; accessors, symbols, hidden fields, sparse arrays,
   extra fields, and alternating length observations fail closed while retaining cleanup ownership.
6. Pi 0.80.6 cumulative `message_update` envelopes are parsed as known closed shapes and charged by
   delta once. Terminal `message_end` and `agent_end` evidence is fully snapshotted/accounted so the
   assistant-byte, event-count, and aggregate-event-byte maxima are jointly attainable.
7. Schema and external result snapshots are prototype-safe: every captured JSON key is an own data
   property, validation uses own descriptors only, and frozen in-memory versus serialized structure
   is identical for `__proto__`, `prototype`, and `constructor` at every nesting level.
8. Schema and event breadth is rejected during bounded incremental traversal before a complete
   adversarial key array is materialized. Known event kinds use closed per-kind field allowlists;
   terminal events reject unknown fields.
9. Every external SDK, workspace, capability, listener, and cleanup failure is normalized before a
   public/runtime/model boundary into a bounded terminal-safe typed and redacted error. Raw external
   errors never survive as a public cause; primary, dual, quarantine, and close aggregation preserve
   only sanitized failure DTOs.
10. The shared redactor treats documentary `=` assignments as credentials, closes
    `Proxy-Authorization`, quoted YAML/flow keys, and OAuth URL fragments, while preserving explicitly
    harmless colon prose byte-identically.
11. Root-scoped reads reject segment-aware cloud config/token/legacy stores, `.envrc`, and common
    private-key names before invoking the workspace, including nested and case variants.
12. Host capability authority rejects sensitive nouns anywhere in the bounded `host_` namespace
    unless an explicit reviewed non-secret allowlist says otherwise. Acquisition synonyms, aliases,
    noun order, singular/plural, and every role are covered.
13. Handoff strings are validated for forbidden terminal controls on the original value before
    redaction. A credential elsewhere in the same string can never turn malformed evidence into an
    accepted sanitized handoff.

All 86 Cycle 9 focused tests, the Cycle 8 disjoint-mutator contract, timer/rejection accounting,
private tool oracle, strict Pi typing, and every prior parser/lifecycle regression remain mandatory.

### Architectural GREEN target

The correction will be organized around bounded ownership components rather than finding-local
branches:

- an authoritative native signal lease that separately records captured-hook failures;
- a staged session cleanup capsule plus operational owner, with an explicit creation state machine
  for `pending | cleaning | terminal` and an internally bounded late-cleanup terminal;
- reusable foreground/detached deadline policies so only detached timers are unreferenced;
- descriptor-safe closed SDK-result/array capture and prototype-safe JSON/result materialization;
- incremental bounded record traversal followed by per-kind event DTO parsing and delta-aware Pi
  stream accounting;
- one boundary failure sanitizer shared by runtime and tool ports, backed by the existing redaction
  grammar but never retaining the external object;
- segment/token classifiers for sensitive paths and capability names, and original-text-first
  terminal validation.

No dependency, tool authority, scheduler/controller, parent issue, Go/connector, GitHub, service, or
credential scope is added.

### Ordered checkpoints and declared gates

The mandatory order is artifact-only PLAN -> one comprehensive test-only RED -> architectural
GREEN/refactor -> terminal evidence. Before GREEN, all 86 retained focused tests must still pass,
every new matrix row must execute and fail for its intended assertion, strict focused TypeScript
must compile, and all three production blob IDs above must remain exact. No production edit may be
amended into RED.

GREEN verification comprises the focused runtime/tool-policy tests, focused and all-production
strict TypeScript 5.9.3 against explicit Pi 0.80.6 package/type roots, the explicit 0.80.6 offline
RPC registration plus no-model tool validation, serialized complete Shepherd classification,
`git diff --check`, immutable-base/frozen-head ancestry, exact issue-owned path scope, and a clean
head. The known managed-sandbox `/bin/ps` `spawn EPERM` family is evidence-classified, never called
green. Push, live GitHub, review bots, merge, Go/connectors/certification, `make verify`, runtime
services, credentials, and model calls remain outside this lane.

`scripts/gsd doctor` passes, while `scripts/gsd prompt programming-loop ...` again returns
`unknown GSD command: programming-loop`; Cycle 10 therefore records the permitted
`manual_gsd_fallback` without weakening TDD. Skills loaded completely are `gsd-programming-loop`,
`javascript-testing-patterns`, `typescript-advanced-types`, `architecture-patterns`, and
`github-issue-first-delivery`, together with required routing, Pi-adapter, universal-loop,
issue-contract, project prompt/PRD, and runtime/Pi references. Execution decision is
`local_critical_path`: all findings collide in the two issue-owned runtime/policy modules, and the
attempted read-only architecture sidecar was rejected by the runtime thread cap.

### Cycle 10 execution result

The required order is complete: artifact-only PLAN `0eb7999f`, comprehensive test-only RED
`6df77689`, and architectural GREEN/refactor `a88cbe52`. RED executed all 102 focused tests: the 86
retained Cycle 9 tests passed and exactly 16 Cycle 10 behavior tests failed their intended
assertions. Strict focused TypeScript passed, and the runtime, policy, and role-prompt production
blobs remained exactly `03cf916b`, `1c8f7010`, and `cfc2d253`.

GREEN implements the thirteen binding invariants as shared ownership, deadline, snapshot, event,
boundary-sanitizer, redaction, path, and capability-classifier mechanisms. The focused suite passes
102/102; focused and all-production strict TypeScript pass against explicit Pi 0.80.6 roots; the
pinned offline RPC registers `pm-shepherd`; and the retained no-model test validates and executes a
real custom-tool call through Pi's validator. No dependency or authority was added.

All sixteen Cycle 10 RED assertions remain intact. Two narrowly documented post-RED fixture
alignments reconcile older controls with the stricter accepted contract: legacy successful-handoff
fixtures now render the same sensitive evidence as one terminal-safe line because WR-01 requires
control rejection before redaction, and the harmless documentary control now uses the allowed
colon form while an added assertion proves its former equals form redacts under BL-04.

Serialized complete-Shepherd verification executed 239 tests: 208 passed and the same 31
controller/state-store tests were environment-blocked because the managed sandbox denies the
pre-existing `/bin/ps` child with `spawn EPERM`. Isolation excluding those two parent-owned files
passes 165/165. This result is not called green. Diff, immutable-base/frozen-head ancestry,
issue-owned scope, and clean-head checks pass; no push, GitHub, service, credential, model, Go, or
connector action was attempted. Parent orchestration owns the permitted-environment rerun, fresh
exact-head review, integration, and external mutation.

## Consolidated Stable-Head Correction Cycle 11 — `1571dc4d4f45ad4285107d04f2d7c489a7f357ab`

Cycle 11 treats the complete `/tmp/475-REVIEW-CYCLE10-1.md` and
`/tmp/475-REVIEW-CYCLE10-2.md` reports as one binding contract. Their unique union is the twelve
new behavior families below plus mandatory retention of every Cycle 10 closure. The immutable
comparison base remains `e659d6f1b666f58748e2d8c86599ceb4bbc62ff8`. Production is frozen at the
exact clean Cycle 10 evidence head until one comprehensive test-only RED commit exists:

| Production path | Frozen Git blob |
|---|---|
| `.pi/extensions/shepherd/agent-session-runtime.ts` | `134697a62252f500b3c58082bf766a5c84766a91` |
| `.pi/extensions/shepherd/tool-policy.ts` | `539d061903549a764567cd1d7fad95d7d624edfe` |
| `.pi/extensions/shepherd/role-prompts.ts` | `cfc2d253c323ad01f34b8c9688b3bad0acd16171` |

### Cycle 11 binding invariants

1. The public Pi adapter accepts the actual 0.80.6 `LoadExtensionsResult` data shape
   `{extensions, errors, runtime}` through one-read own descriptors. `runtime` is required
   compatibility evidence only: Shepherd never stores, grants, invokes, or derives authority from
   it. A pinned no-model integration must traverse the actual `createAgentSession` factory/result
   path, not merely RPC registration or tool validation; fake results match the real shape.
2. Canonical native `EventTarget.prototype.addEventListener/removeEventListener` calls are the only
   signal operations invoked. Shadowed hooks are never executed, so they cannot create alternate
   capture tuples; request, parent, attach failure, and constructor rollback leave no listener.
3. Creation ownership remains associated with its run ID. A successful `abort(runId)` is terminal:
   no later session acquisition, prompt, or cleanup is possible. Resolve, reject, and late cleanup
   are joined; a still-pending uncancellable creation returns a typed non-terminal join failure,
   remains observably owned/quarantined, and repeated abort has a deterministic result.
4. Admission and close share one linearization point. Admission starts before caller/SDK-controlled
   callbacks, close observes pending admission, and reservation atomically rechecks closing state.
   When re-entrant close from request/capability access, model lookup, auth lookup, or setup wins,
   no subsequent create or prompt work starts.
5. Each run owns bounded per-content-index Pi stream state. Outer message and inner partial agree;
   text/thinking/tool-call deltas equal the actual novel suffix; start/delta/end/done/error payloads,
   usage, and diagnostics are validated and charged exactly once. Honest cumulative streams remain
   linear; shrink, replacement, skip, replay, mismatch, and one-above inputs fail closed.
6. Terminal capture is an explicit state machine: exactly one ordered `message_end`, then exactly
   one non-retrying `agent_end`, with no duplicate, out-of-order, or post-terminal event. Both sides
   compare the complete bounded assistant DTO: required api/usage; response, diagnostic, and error
   fields; text/thinking signatures and bodies; and tool IDs, names, arguments, and signatures.
7. Fixed SDK/event/result envelopes use allowlisted own-descriptor reads into closed DTOs and
   discard source extras without any whole-key enumeration. Arbitrary schema/JSON crosses a
   byte/key/node/depth-bounded serialized or trusted-data adapter before construction. Hidden and
   symbol breadth cannot trigger attacker-sized key arrays for schemas, events, creation/extension
   records and arrays, tool arrays, or workspace/capability/mutation results.
8. Failure normalization is total. Proxy/prototype/`instanceof` traps cannot escape; sanitizer
   failure collapses to a constant typed safe error. Aggregate members are pulled manually with a
   16-plus-one ceiling, iterator close is guarded, and long/infinite/throwing shapes settle without
   raw causes across primary, dual, quarantine, and close graphs.
9. Capability admission closes concatenated and separated compounds for every forbidden family:
   generic shell/exec, recursive agents, generic HTTP writes, generic SQL, and credential/token/
   private-key stores. Every role sees only reviewed typed capability semantics.
10. Root-scoped paths reject AWS SSO and CLI cache directory families before any workspace callback,
    including root, nested, and case variants with opaque content.
11. The shared redactor classifies `Cookie` and `Set-Cookie` session/auth header values through
    direct, prompt, workspace, capability/mutation, handoff, and public-error consumers while
    preserving explicit harmless header prose.
12. The shared scanner boundedly parses qualified/dotted assignment keys and classifies the final
    or compound sensitive segment for equals and colon forms. Only the reviewed colon documentary
    prose carveout remains byte-identical.
13. All 102 Cycle 10 focused tests remain mandatory, including staged malformed cleanup,
    independent phase bounds/unreferenced timers, one-read/prototype-safe snapshots, prior path/
    redaction/capability forms, controls-before-redaction, direct Pi tool types, disjoint mutator
    reservations, and transient retry.

### Comprehensive test-only RED matrix

| ID | Finding source | Independent behavior row required to fail at frozen head |
|---|---|---|
| C11-01 | R1-1 | A real pinned Pi 0.80.6 factory result with required `runtime` passes the adapter without using runtime authority; fake fixtures expose the same shape and an actual no-model factory/cancellation path cleans the returned session. |
| C11-02 | R1-2 | Own request/parent add/remove hooks that register capture listeners, throw, mutate, or no-op are never called; native cancellation and constructor rollback leave zero listener tuples. |
| C11-03 | R2 BL-01 | Run-ID abort joins resolve-before/after-bound, reject, hung late-cleanup, and never-settle ownership; only a terminal join resolves, pending returns typed deterministic failure and remains quarantined/observable. |
| C11-04 | R2 BL-02 | Re-entrant close from request, capability, model, auth, resource/setup, and reservation seams wins linearly and prevents every later create/prompt callback. |
| C11-05 | R1-3 / R2 BL-03 | Pi 0.80.6 text/thinking/tool-call start/delta/end plus done/error streams maintain per-index cumulative state; honest growth charges linearly while shrink/replacement/skip/replay/message-partial/delta mismatch and every one-above payload reject. |
| C11-06 | R1-4 / R2 BL-04 | Exactly one ordered message_end/agent_end pair is required; duplicate/out-of-order/post-terminal rows and single-field mismatches across routing/api/usage/response/diagnostics/error/text/thinking/tool evidence reject. |
| C11-07 | R1-5 / R2 BL-06 | Hidden/symbol-heavy schemas, SDK/event envelopes, result/extension/tool arrays, and workspace/capability/mutation results stay below instrumentation ceilings without `Reflect.ownKeys`-style whole-source materialization. |
| C11-08 | R1-7 / R2 BL-05 | Proxy/prototype traps and AggregateError 17/5000/infinite/throwing iterators normalize to bounded typed redacted causes, pull at most 17, close iterators, and settle across primary/dual/quarantine/close. |
| C11-09 | R1-6 / R2 BL-07 | Concatenated/separated/plural/mixed compounds for shell/exec, recursive agent, HTTP write, SQL, credential, access/refresh token, API/client secret, private key, and token cache are absent for every role. |
| C11-10 | R1-8 | `.aws/sso/cache/**` and `.aws/cli/cache/**` root/nested/case variants reject before workspace callbacks with opaque content. |
| C11-11 | R2 BL-08 | Cookie and Set-Cookie session/auth headers redact through direct, prompt, workspace, mutation/capability, handoff, and public-error consumers; harmless controls remain exact. |
| C11-12 | R1-9 | Qualified/dotted keys such as `github.token` and `oauth.client_secret` redact for equals/colon through every consumer while only reviewed colon documentary prose remains exact. |
| C11-13 | retention | The complete 102-test Cycle 10 baseline, strict public Pi tool types, timers/rejections, and mutator-lane ownership remain green throughout RED and GREEN. |

RED acceptance is one test-only commit. The preexisting focused files must first pass 102/102. The
augmented suite then executes all retained tests plus every C11-01 through C11-12 row; retained tests
stay green and each new top-level behavior row fails its intended production assertion without
skip/cancel/todo. Focused strict TypeScript must compile, and all three production blobs above must
remain exact. No production edit may be amended into RED.

### Architectural GREEN / REFACTOR target

- Put a Pi-0.80.6 anti-corruption adapter around creation results and streamed assistant DTOs; copy
  only reviewed fields, consume no extension-runtime authority, and exercise the real factory in a
  cancellation-only no-model integration.
- Replace run admission/creation bookkeeping with an explicit admission token plus per-run creation
  terminal registry shared by `run`, `abort`, `close`, and shutdown. Close waits admissions; reserve
  rechecks close atomically; abort distinguishes terminal success from typed pending ownership.
- Make signal leasing a native-only port with one canonical tuple and rollback-safe release.
- Use discriminated stream/terminal state machines with bounded per-index projections, actual-growth
  accounting, exact transition rules, and complete immutable assistant evidence.
- Replace whole-source key enumeration with allowlisted descriptor adapters for fixed envelopes and
  bounded trusted/serialized construction for arbitrary JSON. Hidden/symbol extras are inert and
  never materialized as a complete attacker-controlled key array.
- Centralize a total boundary failure sanitizer with constant fallback and manually capped aggregate
  iteration; centralize reviewed capability/path/redaction classifiers for the remaining grammar.

No dependency, tool authority, scheduler/controller, #478/#479 file, Go/connector, GitHub, service,
credential, or model scope is added.

### Ordered checkpoints and declared gates

The mandatory order is artifact-only PLAN -> one comprehensive test-only RED -> first cohesive
runtime/policy GREEN -> refactor -> terminal evidence. GREEN requires focused runtime/tool-policy,
focused and all-production strict TypeScript 5.9.3 against explicit Pi 0.80.6 roots, pinned offline
RPC plus the actual no-model create-result/factory exercise, serialized complete Shepherd
classification, the established 165-test safe isolation, `git diff --check`, immutable-base and
frozen-head ancestry, JSON and credential-pattern scans, exact issue-owned path scope, and a clean
head. The known `/bin/ps` `spawn EPERM` family is classified, never called green. No push, network,
GitHub, Go/connectors, `make`, service, credential, or model call is authorized.

`scripts/gsd doctor` passes, but `scripts/gsd prompt programming-loop ...` returns
`unknown GSD command: programming-loop`; Cycle 11 therefore records the permitted
`manual_gsd_fallback` without weakening TDD. Skills loaded completely are `gsd-programming-loop`,
`javascript-testing-patterns`, `typescript-advanced-types`, `architecture-patterns`, and
`github-issue-first-delivery`, plus required routing, Pi/runtime, issue-contract, universal-loop,
project, PRD, prompt, and workflow references. Execution decision is `read_only_spawned`: one
read-only explorer maps the installed Pi 0.80.6 factory/result/event contract while this isolated
worker retains the only mutating path.

### Cycle 11 ordered result

- Artifact-only PLAN: `9366296dcde200bf1f21e74d3cd8dec321581155`.
- Read-only Pi contract-map trace: `a2a8b0e7da426f8c0c6fac91ead65d6a19c4534a`.
- Comprehensive test-only RED: `c58865202623805f8877a583eecf5e301b589f3d`.
  All 102 retained tests passed and exactly 12 named Cycle 11 rows failed their intended behavior
  assertions; strict focused TypeScript passed; no test was skipped, cancelled, or marked todo;
  runtime/policy/role-prompt production blobs remained exactly frozen.
- First cohesive runtime/policy GREEN: `1e605675f8e021a14ed7f709451a2d3a8111c6ad`.
- Stream-accounting refactor: `d9b4eaee71907c662f87f737c9b1a901c35146f9`.
  It separately accounts the complete assistant envelope and replacement state, including
  diagnostic and signature growth, while retaining linear cumulative-text accounting.

The implementation accepts the real Pi result shape but treats `extensionsResult.runtime` only as
descriptor-checked compatibility evidence. Native signal operations, admission/close and
run-creation ownership, stateful assistant projections, total failure normalization, bounded
descriptor/JSON adapters, AWS cache denial, forbidden capability compounds, and shared Cookie/
dotted-key redaction satisfy C11-01 through C11-12. The public `HostCapability` and
`ScopedWorkspace` ports remain unchanged: fixed envelopes are allowlisted own-descriptor DTOs;
bounded arbitrary JSON copies enumerable own data and intentionally discards hidden/symbol peers.

Post-RED test edits do not weaken any Cycle 11 assertion. They add terminal signature/diagnostic
stream-accounting subcases and align retained fixtures with accepted interfaces: canonical Pi
assistant `api`/`usage`, native-only shadow-hook expectations, and inert discarded hidden/symbol
peers for arbitrary DTO/array inputs.

### Cycle 11 terminal gate result

- Focused runtime/tool-policy: 114 passed, 0 failed, 0 skipped/cancelled/todo.
- Focused and all 12-production-file strict TypeScript pass with TypeScript 5.9.3 and explicit Pi
  0.80.6 package/type roots.
- Explicit Pi binary reports 0.80.6; offline RPC registers `pm-shepherd`; C11-01 traverses the real
  no-model factory/result path and cleans the returned session.
- Complete serialized Shepherd executes 251 tests: 220 pass and the unchanged 31 controller/
  state-store rows are environment-blocked because the managed sandbox denies `/bin/ps` with
  `spawn EPERM`. Excluding only those two parent-owned files passes 177/177; the complete run is
  not represented as green.
- Diff, immutable-base/frozen-head ancestry, JSON, issue-owned path scope, no-Go/no-connector, and
  clean-head checks pass after the evidence commit. No push, network, GitHub, model, credential,
  service, Go, connector, or `make` action was attempted. Parent orchestration owns the
  process-capable rerun, fresh exact-head review, integration, and delivery.

## Cycle 12 — Pi-Faithful Lifecycle, Admission Authority, And Boundary DTOs

### PLAN

- Immutable comparison base: `e659d6f1b666f58748e2d8c86599ceb4bbc62ff8`; frozen Cycle 12
  start: `7882cd70c25971e889ec04f63b98c936d605003e`; initial worktree clean.
- Review sources read completely: `/tmp/475-REVIEW-CYCLE11-1.md` and
  `/tmp/475-REVIEW-CYCLE11-2.md`. Their ten unique open families are accepted together as
  C12-01 through C12-10; C12-11 retains every Cycle 11 closure.
- Baseline focused run: 114 passed, 0 failed/skipped/cancelled/todo.
- GSD: `scripts/gsd doctor` passes, but
  `scripts/gsd prompt programming-loop init --phase 475-shepherd-agent-session-runtime --dry-run`
  still returns `unknown GSD command: programming-loop`; the permitted manual-GSD
  PLAN -> RED -> GREEN/refactor -> verify fallback remains active.
- Skills loaded completely: `gsd-programming-loop`, `javascript-testing-patterns` plus its
  advanced-testing reference, `typescript-advanced-types`, `architecture-patterns`, and
  `github-issue-first-delivery`, plus all required repo GSD/Pi/runtime/issue references.
- Orchestration decision: `read_only_spawned`; one no-write explorer maps the explicit installed
  Pi 0.80.6 lifecycle and public offline whole-session seam while this issue worker owns the only
  mutating PLAN, RED, GREEN, verification, commit, and handoff path.
- Production lock at `7882cd70`: runtime Git blob
  `cfb1b40b8835c7bdffe162a7b4d368bde30d54f8`, policy Git blob
  `734927712eaadc9bb8eca383621740d59c5bb7b6`, and role-prompts Git blob
  `cfc2d253c323ad01f34b8c9688b3bad0acd16171`.

### Pi 0.80.6 lifecycle and ownership map

The authoritative normal no-tool run is:

`agent_start -> turn_start -> user message_start -> user message_end -> assistant message_start ->
assistant message_update* -> assistant message_end -> turn_end -> agent_end(willRetry=false) ->
agent_settled`.

A one-tool run inserts an intermediate assistant terminal whose stop reason is `toolUse`, then
`tool_execution_start -> tool_execution_update* -> tool_execution_end -> toolResult message_start
-> toolResult message_end -> turn_end -> turn_start`, followed by the final assistant stream,
`turn_end -> agent_end(willRetry=false) -> agent_settled`. `agent_end.messages` contains the user,
every intermediate assistant, every tool result, and the final assistant in transcript order.
`AgentSession._runAgentPrompt()` emits `agent_settled` in `finally` after retry/continuation handling
and before `prompt()` settles. Thus `agent_end` is attempt-terminal, while `agent_settled` is the
session-owned authoritative completion boundary.

The real-result ownership gate will call the actual pinned `createAgentSession` with public
in-memory services and a unique programmatic provider registered through
`ModelRegistry.registerProvider`. Its inert scripted `streamSimple` returns a public
`AssistantMessageEventStream`; a non-secret offline sentinel satisfies Pi's mandatory prompt-auth
preflight but is never transmitted. The entire unwrapped factory result is returned to Shepherd;
no live auth, model, network, credential, or service path is entered. Shepherd, not the test, owns the
real session's subscribe, prompt, abort/wait, unsubscribe, and dispose lifecycle. The exact factory
result has own enumerable `session`, `extensionsResult`, and `modelFallbackMessage` fields, even
when fallback is `undefined`; `extensionsResult` has exact own enumerable `extensions`, `errors`,
and `runtime` fields. `runtime` is descriptor-validated compatibility evidence and never authority.
Pi forwards inner content events as outer `message_update`; inner `done` produces the assistant
`message_end` rather than a separate outer update. Public session listeners are synchronous/void,
so capture remains subscribed through `agent_settled` and verification occurs after `prompt()`
returns. The unique dynamic API/provider is serialized and unregistered during test cleanup.

### Comprehensive test-only RED matrix

| ID | Review source | Independent behavior row required to fail at frozen head |
|---|---|---|
| C12-01 | R1-1 / R2-1 | One shared Pi-faithful event driver proves real no-tool and one-tool/multi-turn order; user/tool-result messages are bounded but never terminal candidates; the final non-retrying assistant must match the last assistant in `agent_end.messages`, and success requires the subsequent `agent_settled`. Unknown, out-of-order, and in-capture post-settled events reject. |
| C12-02 | R1-6 / R2-6 | The actual pinned factory result and its actual session cross into runtime ownership through the public inert stream seam. Exact own-enumerable result fields and exact `{extensions, errors, runtime}` are mandatory; runtime access is ignored and success fakes use the same exact shape. |
| C12-03 | R1-2 | Descriptor capture establishes a run-ID-owned admission/abort terminal before any caller or SDK callback. Abort during request, capability, model, auth, or setup seams prevents reserve/create/prompt and joins deterministically. |
| C12-04 | R1-3 | Every assistant content index has explicit `{kind, phase, state}` ownership. Exactly one matching start/delta*/end is accepted for text, thinking, and tool-call/`partialJson`; duplicate start/end, delta before/after, kind/index replacement, and cumulative mismatch reject. |
| C12-05 | R1-4 | Capture freezes and unsubscribes at the authoritative settled boundary, then rechecks failure after prompt settlement. Delayed events during idle/unsubscribe/dispose or through a retained post-settled callback cannot turn invalid evidence into a successful handoff or mutate a completed capture. |
| C12-06 | R1-5 | A bounded SDK-aware diagnostic projector accepts installed `createAssistantMessageDiagnostic` and Codex fallback DTOs, consistently omits optional own `undefined`, rejects undefined required fields and arbitrary fields, and keeps diagnostic bytes in the aggregate budget. |
| C12-07 | R2-2 | Every request/authority array is a fresh dense plain own-data array captured through indexed descriptors with authoritative length cap 64. Caller iterator/map/some/join hooks and identity are never used; proxies, accessors, custom prototypes, sparse/extra/hidden behavior, and one-above lengths reject. Only captured arrays feed policy and prompts. |
| C12-08 | R2-3 | Native `AbortSignal.prototype.aborted` brand/state reads occur inside rollback-safe native listener acquisition. False or throwing own shadows on request and parent signals cannot bypass pre-abort, leak a listener tuple, or retain runtime state. |
| C12-09 | R2-4 | Every workspace/capability tool input is captured as a bounded non-proxy own-data DTO before any field access or serialization. Signal/input proxy, accessor, `toJSON`, cycle, and host callback failures reject only as typed bounded redacted public errors. |
| C12-10 | R2-5 | Shared redaction covers Cookie/Set-Cookie strong keys inside JSON, quoted-flow, and bounded diagnostic-prefix contexts through direct, prompt, workspace, mutation/capability, handoff, and public-error consumers while preserving harmless controls. |
| C12-11 | retention | All 114 Cycle 11 focused tests, exact public Pi tool/session types, ownership/timer/error/parser/path/capability closures, and disjoint mutator authority remain green throughout RED and GREEN. |

RED acceptance is one test-only commit. The preexisting focused suite first passes 114/114. The
augmented suite then executes all retained tests plus exactly ten named C12-01 through C12-10
top-level rows: all 114 retained tests remain green and every new row fails its intended production
assertion, with no compile/load failure and zero skip/cancel/todo. Strict focused TypeScript must
pass and all three production Git blobs above must remain exact. No production edit may enter RED.

### Architectural GREEN / REFACTOR target

- Replace the assistant-only terminal pair with one bounded Pi lifecycle state machine. It accepts
  known normal session events, separately owns assistant-turn stream phases, selects only the final
  assistant named by the non-retrying `agent_end`, and completes only at `agent_settled`.
- Make settled capture an explicit freeze/unsubscribe operation whose terminal verification runs
  after prompt settlement and whose cleanup callbacks cannot mutate frozen evidence.
- Create an admission registry from descriptor-captured run identity before normalization and all
  caller/SDK seams; share its abort intent and terminal join with reserve/create/active ownership.
- Centralize fresh dense descriptor-array capture, native signal state/lease acquisition, exact Pi
  result projection, SDK-aware diagnostic projection, and typed tool-input projection. Downstream
  code uses only these captured DTOs.
- Extend the existing monotonic redactor with the remaining structured/prefixed Cookie forms while
  preserving its bounded work and harmless-control behavior.

No dependency, tool authority, scheduler/controller/#478/#479 file, Go/connector, GitHub, service,
credential, model, or network scope is added.

### Ordered checkpoints and declared gates

The mandatory order is artifact-only PLAN + Pi map + finding-to-RED matrix -> one comprehensive
test-only RED -> real-Pi/no-tool first GREEN -> one-tool/multi-turn GREEN -> remaining cohesive
GREEN/refactor -> terminal evidence/freeze. Partial GREEN checkpoints must run their named targeted
rows plus strict focused TypeScript and may not weaken any still-RED assertion. Final GREEN requires
focused runtime/tool-policy, both explicit Pi 0.80.6 strict TypeScript scopes, pinned offline RPC
and actual whole-session no-tool/one-tool exercises, serialized complete Shepherd classification,
safe isolation, `git diff --check`, immutable-base/frozen-head ancestry, JSON and credential-pattern
scans, exact issue-owned path scope, and a clean head. The known `/bin/ps` `spawn EPERM` family is
classified, never called green. No push, network, GitHub, live model/auth, Go/connectors, `make`, or
runtime-service action is authorized.

### Cycle 12 execution result

The required order completed as PLAN `3a6b9299`, test-only RED `58af21f1`, RED evidence
`bc099a76`, real no-tool GREEN `11008da1`, shared one-tool/phase GREEN `b3a99d79`, and cohesive
authority/security/lifecycle refactor `3dc4de71`. The final runtime has no assistant-only success
bypass: accepted evidence must traverse the complete known Pi lifecycle through non-retrying
`agent_end` and `agent_settled`, then freeze and unsubscribe before idle/disposal cleanup.

All ten C12 behavior rows and all 114 retained rows pass, for 124/124 focused. The actual pinned
Pi exercise owns and cleans both no-tool and one-tool sessions with an intermediate tool-use
assistant, real scoped tool execution/result, subsequent final turn, settled trace, zero network,
and exact disposal. Both strict TypeScript scopes and the explicit pinned offline RPC pass. The
complete serialized suite remains honestly classified as 230/261 pass with exactly the unchanged
31 managed-sandbox `spawn EPERM` controller/state-store failures; safe isolation passes 187/187.
All local integrity gates remain issue-owned. Parent orchestration owns only the permitted-process
rerun and fresh independent exact-head review.

## Cycle 13 — Bounded Public Authority And Tool-Lifecycle Correlation

### PLAN

- Immutable comparison base: `e659d6f1b666f58748e2d8c86599ceb4bbc62ff8`; frozen Cycle 13
  start: `5dafc5725167bb74ce88a723073b8c4ceb8314e0`; initial worktree clean.
- Review sources read completely: `/tmp/475-REVIEW-CYCLE12-1.md` and
  `/tmp/475-REVIEW-CYCLE12-2.md`. Their seven non-overlapping blockers are accepted together as
  C13-01 through C13-07; C13-08 retains all 124 Cycle 12 focused behaviors.
- Baseline focused run: 124 passed, 0 failed/skipped/cancelled/todo.
- Production lock at `5dafc572`: runtime Git blob
  `62851ca6bb4b4a7bd0b65d4d1415f992b1455603`, policy Git blob
  `fd6a0e8db7f06ade82b852141eb2a497614aea79`, and role-prompts Git blob
  `cfc2d253c323ad01f34b8c9688b3bad0acd16171`.
- GSD: `scripts/gsd doctor` passes, but
  `scripts/gsd prompt programming-loop init --phase 475-shepherd-agent-session-runtime --dry-run`
  still returns `unknown GSD command: programming-loop`; the permitted manual-GSD
  PLAN -> comprehensive behavior RED -> GREEN -> REFACTOR -> verify fallback remains active.
- Skills loaded completely: `gsd-programming-loop`, `javascript-testing-patterns`,
  `typescript-advanced-types`, `architecture-patterns`, and `github-issue-first-delivery`, plus
  all required repo GSD/Pi/runtime/issue references.
- Orchestration decision: `read_only_spawned`; one bounded no-write explorer maps exact seams and
  installed Pi tool-call/result DTOs while this issue worker owns the only artifact, test,
  production, verification, and commit path.

### Comprehensive test-only RED matrix

| ID | Review source | Independent behavior row required to fail at frozen head |
|---|---|---|
| C13-01 | R1-F1 | A short declared-length request array with thousands of enumerable, hidden, and symbol peers is captured with deterministic bounded indexed-descriptor work. No whole-key/iterator/prototype operation runs, no peer accessor executes, and no peer can influence canonical authority. |
| C13-02 | R1-F2 | Abort, close, and parent shutdown triggered from each re-entrant SDK setup seam, including both `getAgentDir` callbacks, leave zero creation, prompt, reservation, listener, timer, or session leaks and one truthful terminal outcome. |
| C13-03 | R2-F1 | Exported `createToolPolicy` and prefix normalization intrinsically snapshot every authority array before iteration; custom iterators and caller `some`/`map`/`join`/constructor/prototype hooks remain unused, alias mutation cannot expand scope, and length 65 rejects. |
| C13-04 | R2-F2 | Table-driven capability names using equivalent word orders, compounds, and synonyms for generic execution/terminal/process and protected data/export operations are absent for read-only and mutating roles, while reviewed legitimate host capabilities remain available. |
| C13-05 | R2-F3 | Dotted and qualified keys whose sensitive compound is split across segments redact through direct text, both prompts, workspace output, capability/mutation results, handoff summary/finding, and every public-error graph; no synthetic marker survives. |
| C13-06 | R2-F4 | Exported `buildRolePrompts` and authority validation intrinsically snapshot context and authority arrays; caller iterator/accessor/array methods/prototype/constructor behavior stays unused, post-call alias mutation cannot alter either immutable prompt, and one-above arrays reject. |
| C13-07 | R2-F5 | One authorized Pi tool call is correlated across assistant call ID/name/arguments, active tool authority, execution start/end, tool-result message ID/name/result/error, turn results, next turn, and final handoff. Mismatch, duplicate, replacement, result-without-active, missing result, or early handoff rejects; the real pinned no-tool and one-tool paths remain accepted. |
| C13-08 | retention | All 124 Cycle 12 focused tests, actual pinned offline whole-session exercises, strict public Pi types, lifecycle/ownership/timer/parser/path/capability closures, and disjoint mutator authority remain green throughout RED and GREEN. |

RED acceptance is one test-only commit. The preexisting focused files must first pass 124/124.
The augmented suite must then execute all retained tests plus exactly seven named C13-01 through
C13-07 top-level behavior rows: all 124 retained rows remain green and every new row fails its
intended production assertion, with no load/fixture failure and zero skip/cancel/todo. Focused
strict TypeScript must pass, all three production blobs above must remain exact, and no production
edit may enter RED.

### Architectural GREEN / REFACTOR target

- Replace `captureFreshDenseArray` whole-key collection with canonical influence capture: intrinsic
  array brand, one own-data length descriptor, exactly `length` indexed own-data descriptors, and a
  fresh frozen array. Never enumerate, read, or preserve enumerable/non-enumerable/symbol peers;
  they are inert rather than accepted as authority.
- Put an active-scope assertion immediately after every re-entrant SDK callback and immediately
  before the microtask that invokes `createAgentSession`; cancellation/close observed at any seam
  must win before child creation.
- Share one intrinsic dense-array projector at the public policy and role-prompt boundaries. Every
  downstream validator/formatter receives only fresh immutable snapshots and never invokes
  caller-owned iteration, array methods, accessors, constructors, or prototypes.
- Centralize semantic capability-name classification around normalized word components and closed
  forbidden authority families; centralize dotted/qualified sensitive-key classification around
  all bounded segments and compounds.
- Extend the Pi lifecycle machine with one per-call correlation record. Assistant tool-call
  evidence creates the only authorized identity; execution start/end and tool-result messages must
  match it exactly; `turn_end.toolResults` must match the closed result set before a subsequent turn
  or final handoff is accepted. The active-name allowlist is the existing frozen policy tool set;
  scoped argument enforcement remains in the matching real custom-tool execution path.

No dependency, public authority, scheduler/controller/#478/#479 file, Go/connector, GitHub,
service, credential, model, network, or parent-artifact scope is added.

ECMAScript exposes no bounded/streaming own-key primitive. `Reflect.ownKeys`, `Object.keys`,
`Object.getOwnPropertyNames`, `Object.getOwnPropertySymbols`, and descriptor-map APIs must first
materialize the complete key set; `for...in` does not expose hidden strings or symbols and is not a
bounded proof either. Literal proof-of-absence on an arbitrary raw array is therefore incompatible
with R1-F1's bounded-work invariant. Rejecting every ordinary unbranded array would break the public
API without adding authority safety. Cycle 13 adopts the only compatibility-preserving bounded
contract: copy only the authoritative indexed data into a private immutable DTO and make every
other source field observably non-influential.

### Ordered checkpoints and declared gates

The mandatory order is artifact-only PLAN -> one comprehensive test-only RED -> cohesive boundary
and lifecycle GREEN -> no-contract-widening REFACTOR -> terminal evidence/freeze. GREEN requires
focused runtime/tool-policy/role-prompt behavior, focused and all-production strict TypeScript
5.9.3 against explicit Pi 0.80.6 roots, actual pinned no-tool and one-tool whole-session rows,
pinned offline RPC, serialized complete Shepherd classification, safe isolation, `git diff
--check`, immutable-base/frozen-head ancestry, JSON and credential-pattern scans, exact same
20-path issue scope, and a clean head. The known 31 controller/state-store `/bin/ps` `spawn EPERM`
rows remain an honest environment classification, never a green complete suite. No push, network,
GitHub, live model/auth, Go/connectors, `make`, runtime service, or credential action is authorized.

### RED evidence

- Status: captured in test-only commit
  `974d2e795038d5531c9aca39fbdcfbe73b2caf8a`.
- Focused result: exit 1; 131 executed, all 124 retained tests passed, exactly seven intended Cycle
  13 behavior rows failed, and 0 skipped/cancelled/todo. Each failure is one named C13-01 through
  C13-07 row with its intended assertion evidence.
- Focused strict TypeScript: exit 0 with TypeScript 5.9.3 against explicit Pi 0.80.6 package/type
  roots for runtime, policy, role prompts, and both focused tests.
- Production lock remains exact: runtime `62851ca6bb4b4a7bd0b65d4d1415f992b1455603`, policy
  `fd6a0e8db7f06ade82b852141eb2a497614aea79`, role prompts
  `cfc2d253c323ad01f34b8c9688b3bad0acd16171`.
- Failure evidence is behavior-specific: one forbidden whole-key call; three second-agent-dir
  create calls across abort/close/shutdown; four split-qualified secret markers; ten prompt caller
  method calls/two accessor reads/mutable result; nine accepted corrupt tool lifecycles; six policy
  caller behavior calls; and all 30 role/name combinations admitted from the forbidden capability
  table. `git diff --check` passes.

### GREEN evidence

- First cohesive GREEN: `48f546a5`. `captureFreshDenseArray` now reads only the intrinsic array
  brand, one own-data length descriptor, and the bounded indexed own-data descriptors before
  freezing a fresh canonical array; no complete key set is materialized.
- Runtime execution now reasserts lifecycle authority after every re-entrant SDK seam and
  immediately before session creation. Abort, close, and shutdown observed by the second
  `getAgentDir` callback therefore schedule no child session or prompt.
- Exported policy and role-prompt entry points project caller arrays into private frozen dense
  snapshots before validation or formatting. The prompt result is itself frozen.
- Capability denial uses normalized structural token families for generic execution/process/
  terminal authority and protected-data/export authority. Sensitive assignment classification
  scans bounded dotted compounds rather than only the final segment.
- Pi tool lifecycle capture now owns one exact assistant call identity through execution start/end,
  tool-result message, `turn_end.toolResults`, subsequent turn, and final handoff. Orphaned,
  replaced, duplicated, missing, early, or mismatched evidence fails closed.
- Result: all seven C13 behavior rows and all 124 retained rows pass, for 131/131 focused.

### REFACTOR / terminal verification

- Refactor checkpoint: `e50b5f97`. It centralizes the capability token families, expands the
  structural synonym controls, and strengthens result/error lifecycle mismatch coverage without
  weakening any RED row or widening public authority.
- Focused tests: 131 passed, 0 failed/skipped/cancelled/todo. Safe isolation excluding only the
  environment-bound controller/state-store files: 194 passed, 0 failed/skipped/cancelled/todo.
- Complete serialized Shepherd: 268 executed, 237 passed, 31 failed, 0 skipped/cancelled/todo.
  Every failure is the unchanged controller/state-store process-identity family where the managed
  sandbox rejects child creation with `spawn EPERM`; the complete suite is environment-blocked,
  not represented as green.
- Focused and all 12 non-test Shepherd strict TypeScript scopes pass with TypeScript 5.9.3 against
  the explicit installed Pi 0.80.6 package/type roots. Explicit Pi 0.80.6 offline RPC exits 0 and
  registers `pm-shepherd`; only the known global-settings-lock sandbox warnings are emitted.
- Final production Git blobs are runtime `cd5c05411933c1a1f1b239d8ac85112e47e10b8b`, policy
  `5a7f91b863f3a3eba3b489e79944c17a6511a776`, and role prompts
  `d4365dd2e32854589a7d1bee91439e5cb0a17fe0`.
- Both Cycle 12 reports were re-read in full after GREEN/refactor. Their complete two-plus-five
  blocker union remains mapped to C13-01 through C13-07 with no omitted or deferred finding.
- `git diff --check`, immutable-base/frozen-start ancestry, RUN-STATE JSON, credential-pattern,
  dependency/Go/connector, and exact 20-path issue-scope checks pass. No push, network, GitHub,
  live model/auth, credential, service, Go/connector, `make`, parent, or #478 mutation occurred.

## Cycle 14 — Closed Authority Schemas And Post-Creation Barriers

### PLAN

- Immutable comparison base: `e659d6f1b666f58748e2d8c86599ceb4bbc62ff8`; frozen Cycle 14
  start: `67050a4a3cf62d0d40660de76938ab72ac68ee96`; initial worktree clean.
- Review sources read completely: `/tmp/475-REVIEW-CYCLE13-1.md` and
  `/tmp/475-REVIEW-CYCLE13-2.md`. Review 1's three blockers and Review 2's blocker/warning reduce
  to three architecture families, C14-01 through C14-03; all 131 Cycle 13 focused rows are retained.
- Baseline focused result: 131 passed, 0 failed/skipped/cancelled/todo. Frozen production Git blobs
  are runtime `cd5c05411933c1a1f1b239d8ac85112e47e10b8b`, policy
  `5a7f91b863f3a3eba3b489e79944c17a6511a776`, and role prompts
  `d4365dd2e32854589a7d1bee91439e5cb0a17fe0`.
- GSD adapter health passes, but
  `scripts/gsd prompt programming-loop init --phase 475-shepherd-agent-session-runtime --dry-run`
  still returns `unknown GSD command: programming-loop`; the recorded manual-GSD
  PLAN -> comprehensive RED -> one cohesive GREEN -> REFACTOR -> verify fallback remains active.
- Required skills loaded completely: `gsd-programming-loop`, `architecture-patterns`,
  `javascript-testing-patterns`, `typescript-advanced-types`, and
  `github-issue-first-delivery`, including their required workflows and all repo Pi/runtime/GSD/
  issue contracts.
- Execution decision: `read_only_spawned`. One no-write explorer maps the post-creation Pi seams,
  current/#479 capability inventory, and structured-field consumers; this issue worker owns all
  artifacts, tests, production, verification, and commits.

### Why semantic blacklists are abandoned

A finite denylist cannot close an open string namespace: every added action/resource synonym still
leaves another spelling with the same authority. Cycle 13 proved only its sampled vocabulary, and
Review 2 admitted seven independently chosen semantic variants. Cycle 14 therefore stops trying to
name every unsafe operation. It enumerates the complete safe host capability domain instead. Any
string outside that finite registry is invalid regardless of spelling, order, plurality, or
meaning; generic shell/process/terminal, transport write, database write, secret export, and agent
spawn authority are absent because no registry member can represent them.

The same inversion applies to structured fields. An endless secret-alias list cannot be complete,
while fuzzy ancestor subsequences modify public leaves. The scanner will parse one bounded
canonical segment path, recognize exact sensitive terminal compounds/schema paths, preserve only
explicit public metadata terminal/path forms, and treat every other assignment-shaped key as
unknown-sensitive. Unknown assignments fail closed; ordinary non-assignment prose remains outside
the structured grammar.

### Closed capability registry and #479 handoff

The complete legitimate `HostCapability` inventory visible to current Shepherd and the planned
#479 integration is:

| Registry identity | Mutates | AgentSession availability | Contract |
|---|---:|---|---|
| `host_inspect` | false | read-only and mutating | inspect bounded host-owned evidence through a supplied typed adapter |
| `host_verify` | true | mutating only | execute/assess only a declared bounded verification adapter |

`workspace_read`, `workspace_edit`, and `workspace_write` remain separately implemented scoped
workspace tools, not host registry members. #479's scheduler, Git/worktree, GitHub, review,
decision, and integration adapters remain controller-owned ports outside the AgentSession tool
surface. #479 may pass only the two registered host identities above; it cannot invent an arbitrary
string or expose its controller ports as dynamic capabilities. A future legitimate host operation
requires an explicit reviewed registry/type change in this contract, not an extension escape hatch.

The registry will be exported as an immutable runtime record and a compile-time literal union. The
`HostCapability` union is discriminated by exact registry name and exact mutability. Public
`createToolPolicy`, direct prompt construction, and runtime requests accept only registered tool
identities; runtime validation rejects forged JavaScript values before SDK work. Authority and
supplied capability names must match each other and the registry exactly.

### Post-creation barrier map

After `createAgentSession` fulfills and before `prompt`, all host/Pi-controlled callbacks are one
ordered acquisition state machine: result-session acquisition and mandatory cleanup root;
independent abort/wait/dispose/prompt/subscribe/active-tool operation capture; model/provider/id,
thinking, session-file, and active-tool validation; subscription acquisition; then the prompt
side effect. The cleanup root is acquired before optional validation so cancellation at any later
seam still owns disposal. An active-scope barrier follows every re-entrant callback. A final barrier
immediately precedes each next subscription or prompt side effect. Cancellation/close/shutdown may
short-circuit validation but may not skip exactly-once unsubscribe/dispose/join cleanup.

### Comprehensive test-only RED matrix

| ID | Review source | Behavior required to fail at frozen head |
|---|---|---|
| C14-01 | R1-F1 | Table abort/close/shutdown at every post-create result/session operation getter, model/provider/id/thinking/session-file getter, active-tool callback, and subscribe callback. No later validation/subscription/prompt side effect runs; unsubscribe/dispose/join are exactly once as acquired; abort/wait ownership is truthful; request listener, reservation/lease, and long timer accounting returns to zero. |
| C14-02 | R1-F2, R2-B1 | Exported immutable registry and literal union contain exactly `host_inspect`/`host_verify` with exact mutability. Every legitimate entry behaves correctly for read-only/mutating policy, while broad unknown semantic/syntactic variants reject through direct policy, prompt construction, and runtime admission before SDK work. Capability/authority identity or mutability mismatch rejects; no arbitrary-string extension path exists. |
| C14-03 | R1-F3, R2-W1 | Exact sensitive schema paths/terminal compounds and arbitrary unknown aliases redact through direct, both prompts, workspace read/edit/write, capability result/reference, handoff summary/finding/verification, and public-error graphs. Qualified public metadata terminal/path controls remain byte-identical through every consumer. Scanner work stays bounded and no synthetic value is reflected in a public failure. |
| C14-04 | retention | All 131 Cycle 13 focused behaviors, including real pinned Pi no-tool/one-tool sessions and prior lifecycle/authority/parser/path/security closures, remain green throughout RED and GREEN. |

RED acceptance is one test-only commit adding exactly three top-level C14 rows. The augmented suite
must execute 134 rows: all 131 retained rows pass and exactly C14-01 through C14-03 fail their
intended behavior assertions, with zero skip/cancel/todo. Focused strict TypeScript must pass and
all three frozen production blobs must remain exact. No production edit may enter RED.

### RED evidence

- Test-only checkpoint: `229217f4` (`agent-session-runtime.test.ts` only), adding exactly the three
  declared top-level C14 rows and no production or artifact mutation.
- Focused result: 134 executed; all 131 retained rows pass; exactly C14-01, C14-02, and C14-03 fail;
  0 skipped/cancelled/todo. C14-01 exposes continued post-trigger validation plus close/shutdown
  subscription/prompt and synchronous invalid-event prompt. C14-02 exposes the missing registry,
  admitted arbitrary names, forged mutability, prompt authority, and runtime SDK entry. C14-03
  exposes three unknown-assignment leaks and modification of all three public metadata controls
  through the shared consumer set.
- Focused strict TypeScript 5.9.3 passes against the explicit Pi 0.80.6 roots. `git diff --check`
  passes. Frozen production blobs remain exact: runtime `cd5c05411933c1a1f1b239d8ac85112e47e10b8b`,
  policy `5a7f91b863f3a3eba3b489e79944c17a6511a776`, and role prompts
  `d4365dd2e32854589a7d1bee91439e5cb0a17fe0`.
- Production is now unlocked only for the single cohesive C14 GREEN; no family may be checkpointed
  independently.

### Cohesive GREEN / REFACTOR / verification contract

The three families freeze together, not piecemeal: one architecture-level implementation replaces
the lifecycle acquisition sequence, capability string namespace, and structured-field classifier;
only after all 134 rows pass may refactoring begin. Refactor may remove obsolete semantic token/
regex sets and centralize schema tables, but may not reintroduce open strings, synonym expansion,
or weaken any RED row.

Declared gates are focused runtime/tool-policy/role-prompt behavior; focused and all 12 production
strict TypeScript 5.9.3 against explicit Pi 0.80.6 roots; retained actual pinned no-tool/one-tool
rows; explicit Pi 0.80.6 offline RPC; serialized complete Shepherd classification; safe isolation;
`git diff --check`; immutable-base/frozen-start ancestry; JSON, credential, dependency, Go/
connector, exact same 20-path scope, and clean-head checks. The known 31 controller/state-store
`spawn EPERM` rows remain an honest environment classification only. Both Cycle 13 reports must be
re-read after GREEN before evidence freeze. No push, network, GitHub, live model/auth, credential,
service, Go/connector, `make`, main, parent, #478, or #479 mutation is authorized.

### GREEN / REFACTOR / exact verification evidence

- Cohesive GREEN `9af22e72` closes all three C14 families together. Post-create session/result,
  operation, route, metadata, active-tool, and subscription callbacks are barriered before later
  work; the mandatory cleanup root remains owned. Policy, prompt, and runtime use the exported
  exact `host_inspect`/`host_verify` registry and reject unknown or forged authority before SDK
  creation. Structured assignment classification uses canonical exact paths, narrow public
  controls, and an unknown-sensitive default.
- No-contract-widening REFACTOR `27c07eec` centralizes mandatory/optional session-operation capture
  and canonical assignment classification. It adds no capability, public field, secret synonym,
  dependency, or authority seam.
- Focused verification passes 134/134 with zero fail/skip/cancel/todo, including the retained actual
  pinned Pi no-tool/one-tool sessions. Focused production/tests and all 12 non-test Shepherd files
  pass strict TypeScript 5.9.3 against explicit Pi 0.80.6 roots. Explicit offline RPC exits 0 and
  registers `pm-shepherd`; only the expected managed-filesystem settings-lock warnings occur.
- Safe isolation excluding only `controller.test.ts` and `state-store.test.ts` passes 197/197. The
  terminal serialized complete suite executes 271 tests: 240 pass and the unchanged 31 failures
  are confined to those two files' managed-sandbox process-identity `spawn EPERM` family. The full
  suite remains environment-blocked, not green.
- Both complete Cycle 13 reports were re-read after GREEN/refactor. R1-F1, R1-F2/R2-B1, and
  R1-F3/R2-W1 each map to a passing C14 row with no deferred or weakened finding.
- Production blobs at REFACTOR are runtime `e37078faaae0f1afb9583a6a3dbcd114acd92040`, policy
  `f8dfcdd32b5f559b9d3ac409ac37f66bd42a88f0`, and role prompts
  `b762787b2a63b5b02f9591c7bf3fff46394738cc`.
- Diff, immutable-base/frozen-start ancestry, RUN-STATE JSON, credential-pattern,
  dependency/Go/connector, exact same 20-path issue scope, and clean-worktree checks pass at the
  terminal artifact boundary. No push, network, GitHub, live model/auth, credential, service,
  Go/connector, `make`, main, parent, #478, or #479 mutation was attempted.

## Cycle 15 — Default-Deny Assignment Parsing And Own-Descriptor Pi Capture

### Frozen baseline and complete inputs

- Frozen blocked start: `f41cde91e01e439a5ebbbaa4867729e0fa80b371`; immutable base:
  `e659d6f1b666f58748e2d8c86599ceb4bbc62ff8`; clean worktree and the exact same 20 issue paths.
- Complete inputs: `/tmp/475-REVIEW-CYCLE14-1.md` and `/tmp/475-REVIEW-CYCLE14-2.md`. Their unique
  union is exactly three blockers: quoted unknown-sensitive values, conservative assignment-key/
  flow closure at grammar boundaries, and post-create prototype-reflection/callback batching.
- Baseline focused result: 134 passed, 0 failed/skipped/cancelled/todo. Focused strict TypeScript
  passes. Frozen production blobs are runtime `e37078faaae0f1afb9583a6a3dbcd114acd92040`, policy
  `f8dfcdd32b5f559b9d3ac409ac37f66bd42a88f0`, and prompts
  `b762787b2a63b5b02f9591c7bf3fff46394738cc`.
- Required skills loaded: `gsd-programming-loop`, `architecture-patterns`,
  `javascript-testing-patterns`, `typescript-advanced-types`, and
  `github-issue-first-delivery`. `scripts/gsd doctor` passes, but `scripts/gsd prompt
  programming-loop ...` remains absent; Cycle 15 records the required manual-GSD fallback without
  weakening PLAN -> RED -> GREEN -> REFACTOR -> VERIFY.
- Execution decision: `read_only_spawned`; one bounded no-write explorer maps the scanner and Pi
  capture seams while this worker owns every artifact, test, production edit, commit, and gate.

### Scanner default-deny architecture

The scanner will stop treating its narrow canonical-key grammar as the admission grammar. One
monotonic bounded candidate parser will recognize line and flow assignment syntax independently of
whether the key can be canonically classified. It retains at most the reviewed canonical-key
prefix, marks spaces, unsupported characters, malformed quoting/escaping, and length overflow as
uncertain, and continues only far enough to find a delimiter or the already-bounded line/container
boundary. It never enumerates an unbounded key or abandons the remaining flow after uncertainty.

Classification is closed: an unambiguous exact reviewed public terminal/path is the only
non-redacted assignment. Exact sensitive paths/terminals override. Every other recognized or
uncertain assignment is `unknown-sensitive`. Unknown-sensitive quoted values always redact their
complete quoted content; they never enter Authorization credential splitting. Unquoted, block,
multi-component, malformed, and flow values redact their complete bounded value. If a flow field
cannot be safely separated, the remaining bounded field/container is redacted conservatively and
the scanner resumes at the next provable sibling boundary. Public controls
`api.key.version`, `private.key.algorithm`, and `database.url.scheme` remain byte-identical only
under an exact unambiguous parse. Scanner metrics must remain constant-factor monotonic.

### Post-create own-descriptor capture architecture

Post-create validation will use no `for...in` or prototype-chain enumeration on Pi values. Direct
proxies reject before reflection. Result records and extension records must have approved direct
record prototypes; extension and active-tool arrays must have the canonical array prototype.
Validation captures only expected own data descriptors and canonical indexed descriptors; inert
prototype/non-authoritative peers are never traversed or invoked.

Creation-result ownership, extension-result capture, each extension array, active-tool method
return/canonical indexes, subscription call, and returned unsubscriber acquisition become explicit
ordered seams. Every unavoidable Pi getter or method callback is followed immediately by the
closure-aware execution barrier before the next callback, validation step, subscription, or
prompt. Abort, close, or shutdown may stop optional capture but cannot erase the mandatory
abort/wait/dispose cleanup root or exactly-once unsubscribe/dispose/join ownership.

### Comprehensive test-only RED matrix

One test-only commit adds exactly three top-level rows and no production/artifact change:

1. `cycle 15 quoted unknown assignments redact whole values through every shared consumer`
   tables both quote styles, line/flow forms, multi-component values, malformed/extended quoted
   unknown keys, and exact quoted public controls. Every component must disappear through direct,
   both prompts, workspace read/edit/write, capability summary/reference, handoff fields, and
   policy/runtime public errors.
2. `cycle 15 uncertain assignment keys fail closed across cutoffs flows and later siblings`
   tables both sides of the canonical character/length cutoffs, internal spaces, unsupported
   characters, malformed forms, line/flow contexts, and later protected siblings. Unknown and
   later sensitive markers must disappear through every shared consumer; exact public controls
   remain byte-identical; 25/50/100-KiB metrics remain constant-factor monotonic.
3. `cycle 15 post-create validation rejects prototype callbacks and barriers every split seam`
   tables ordinary result/extension/array values with hostile proxy prototypes plus synchronous
   abort/close/shutdown at every unavoidable getter/method/subscription seam. Prototype traps and
   later callbacks/prompt must stay zero; acquired abort/wait/unsubscribe/dispose/join ownership,
   listeners, leases, and referenced timers must settle exactly.

RED acceptance is 137 executed: all 134 retained rows pass and exactly these three new rows fail
their behavior assertions, with zero skip/cancel/todo and focused strict TypeScript green. All
three production blobs must remain exact. No production family may freeze separately; one cohesive
GREEN must close all three rows before refactoring.

### GREEN / REFACTOR / verification contract

GREEN is one architectural change across the shared scanner and post-create capture pipeline. It
may not add key aliases, public exceptions, prototype allowlist escape hatches, or dynamic
authority. REFACTOR may centralize candidate parsing, prototype predicates, and callback barriers
only after 137/137 passes and may not weaken a RED row.

Declared gates are focused runtime/tool-policy/role-prompt behavior; focused and all 12 production
strict TypeScript 5.9.3 against explicit Pi 0.80.6 roots; retained actual pinned no-tool/one-tool
rows; explicit Pi 0.80.6 offline RPC; safe isolation; serialized complete Shepherd environment
classification; `git diff --check`; immutable-base/frozen-start ancestry; RUN-STATE JSON;
credential/dependency/Go/connector scans; exact same 20 paths; and clean head. Both Cycle 14 reports
must be re-read after GREEN. No self-review/integration, push, network, GitHub, live model/auth,
credential, service, Go/connector, `make`, main, parent, #478, or #479 mutation is authorized.

### Cycle 15 execution result

- PLAN `b8b68412`; comprehensive test-only RED `5d83d519`; cohesive GREEN `38e95460`; REFACTOR
  `ee4943f4`. RED is exactly 137 executed / 134 retained pass / 3 intended fail / zero
  skip-cancel-todo with all production blobs frozen. GREEN/refactor is 137/137.
- C15-01 closes quoted unknown whole-value routing through all 13 consumers. C15-02 closes both
  sides of the key cutoff plus spaced, extended, unsupported, overlong, unclosed, line/flow, later
  sibling, exact-public, and 25/50/100-KiB bounded-work cases. C15-03 closes creation, extension,
  array, active-tool, and synchronous abort/close/shutdown prototype/barrier seams with exact
  cleanup and zero post-transition prototype callbacks.
- REFACTOR removes all runtime `for...in` and the superseded scanner decoder. Focused 137/137,
  safe isolation 200/200, both strict TypeScript scopes, and explicit pinned offline RPC pass.
  Complete Shepherd remains environment-blocked at 243/274 only by the unchanged 31
  controller/state-store `spawn EPERM` failures.
- Both Cycle 14 reports were re-read in full after GREEN. Diff/ancestry/JSON, exact same 20 paths,
  credential/dependency/Go/connector, runtime prototype-enumeration, and clean-worktree gates pass.
  Parent orchestration owns fresh exact-head review, integration, and delivery.

## Cycle 16 — Monotonic Redaction, Closed DTO Capture, And Owned Prompt Settlement

### Frozen baseline and complete review input

- Frozen blocked start: `df930d62adbf85af88cebbdf66e5eefaab587f4b`; immutable base:
  `e659d6f1b666f58748e2d8c86599ceb4bbc62ff8`; clean worktree and the exact same 20 issue-owned
  paths as Cycle 15.
- Complete inputs: `/tmp/475-REVIEW-CYCLE15-1.md` and `/tmp/475-REVIEW-CYCLE15-2.md`. Both were
  read in full. Their duplicated observations consolidate to exactly six architecture families:
  punctuation/default-deny assignment admission, malformed quote bounds, monotonic scanner work,
  finite public-scalar fidelity, zero whole-key capture, and synchronous prompt-settlement
  ownership.
- Baseline focused result: 137 passed, 0 failed/skipped/cancelled/todo. Focused strict TypeScript
  passes. Frozen blobs are runtime `0cee613ac4aabae2c7eb0fbaa58ae590f3bf0cb0`, policy
  `5750d989c25557f802d44a83a31cb349888bc948`, prompts
  `b762787b2a63b5b02f9591c7bf3fff46394738cc`, runtime test
  `f3fbdc5899820c9c59dff5bbe8f806163ac0a3be`, and tool-policy test
  `548f6681a0d579b413dae11a4cc17b32f9814c30`.
- Required skills loaded: `gsd-programming-loop`, `architecture-patterns`,
  `javascript-testing-patterns`, `typescript-advanced-types`, and
  `github-issue-first-delivery`. `scripts/gsd doctor` passes for the 69-command adapter, but
  `scripts/gsd prompt programming-loop ...` is unavailable; Cycle 16 therefore records the
  permitted manual-GSD fallback and keeps PLAN -> RED -> GREEN -> REFACTOR -> VERIFY exact.
- Execution decision: `read_only_spawned`. One no-write mapper deduplicated review families and
  mapped scanner, capture, prompt, and test seams. This issue worker owns all artifact, test,
  production, gate, and commit mutations.

### One streaming monotonic assignment lexer

`redactStructuredAssignments` will use one forward cursor and an explicit finite mode/state record
for line members, flow members, quoted/unquoted keys, delimiter admission, plain/quoted/block
values, comments, and malformed recovery. Opening `{` and `[` updates the bounded flow stack when
the cursor reaches the opener; no suffix lookahead decides whether a flow should open. Line,
quoted-value, and container ends are discovered by this same cursor, not by `findLineEnd`,
`findQuotedValueEnd`, `looksLikeFlowMapping`, `looksLikeFlowSequence`, or any equivalent rescan.
The cursor never regresses and a source offset is visited by the main lexer at most once.

The exported `RedactionScanMetrics` contract will report exactly `sourceLength`,
`cursorAdvances`, `cursorRegressions`, `maxMainCursorVisits`, `keyCharacterVisits`,
`boundaryCharacterVisits`, and `totalWork`. For every input, `cursorRegressions === 0`,
`maxMainCursorVisits <= 1`, `cursorAdvances <= sourceLength`, and
`totalWork <= (8 * sourceLength) + 64`. Key/path classification is charged to the same metric and
may inspect only the bounded candidate retained by the lexer. Dense opener, malformed-flow, and
terminal-sibling fixtures at 25/50/100 KiB must satisfy the exact ceiling and linear-growth
controls; a large family may not be hidden behind a separately quadratic helper.

### Default-deny admission, malformed quote bounds, and finite fidelity

Assignment syntax admission is independent from key classification. At a proven line/member start,
unquoted, quoted, spaced, punctuation-bearing, extended-character, overlong, escaped, or malformed
key candidates remain assignment candidates once `:` or `=` is reached. Only an exact,
unambiguous reviewed public field is emitted unchanged; exact protected fields and every other
recognized/uncertain assignment are redacted. If member separation is ambiguous, the lexer may
redact the whole proven flow container or remaining line rather than exposing a value or later
sibling. It may never abandon a candidate because its key is outside a preferred alphabet.

Quoted values have one proven bound: a syntactically closed quote followed by an admitted scalar
boundary, otherwise the current line, containing flow container, or EOF. A quote-like character
inside malformed content does not become a close merely because it precedes punctuation. When no
safe sibling boundary is knowable, the larger proven line/container is redacted. Subsequent
line-level or post-container siblings must still be parsed and protected. Both quote styles,
escape variants, flow/line forms, and malformed values are one table.

The public scalar grammar is finite and explicit, not inferred from “looks harmless.” It preserves
byte-identically only exact reviewed public paths/terminals and these complete scalar forms:
scheme/locator values (including query `=` characters), time-like/RFC3339 values, booleans and
finite numerics already admitted by the exact public field, and syntactically closed single- or
double-quoted scalar lines with no delimiter outside the closing quote. The reviewed public field
set remains `safe`, `enabled`, `retained`, `flavor`, `message`, `api.version`,
`api.key.version`, `private.key.algorithm`, and `database.url.scheme`. Every consumer receives the
same classification and exact bytes; adjacency to an uncertain/protected assignment never changes
the public scalar.

### Schema-known capture with zero whole-key primitives

Untrusted Pi/session/event/request/handoff capture will call none of `Reflect.ownKeys`,
`Object.keys`, `Object.getOwnPropertyNames`, `Object.getOwnPropertySymbols`,
`Object.getOwnPropertyDescriptors`, `Object.values`, or `Object.entries`. Each known boundary gets
an explicit projector whose field names are fixed by its schema and read with direct intrinsic own
descriptor lookup. Dense arrays are captured from a bounded own `length` plus canonical indexed
descriptors only. Known lifecycle envelopes, creation/session records, request authority,
workspace/tool arguments, assistant/tool-result/turn fields, and handoff DTOs each use their own
fixed field table. Inert undeclared fields are ignored without being discovered or invoked.

Arbitrary records such as provider diagnostics, tool `details`, generic host arguments/results, or
unknown JSON objects do not become traversable public evidence. They reduce to a fixed opaque or
redacted summary. Registered workspace tools may project only their declared argument/result DTO;
runtime-owned identity records use a separate trusted canonical serializer. Tests instrument all
seven whole-key primitives over records with 4,096 enumerable, non-enumerable, symbol, and accessor
peers and require exactly zero calls, zero peer getter/proxy influence, bounded canonical reads,
and retained known-field behavior at creation, request, event, result, and handoff boundaries.

### Prompt settlement ownership before the barrier

The return or synchronous throw from `session.prompt` becomes an owned settlement record before
the first post-prompt lifecycle barrier. Native promises and foreign thenables are immediately
assimilated with both fulfillment and rejection handlers into an always-fulfilled internal record:
`{ status: "fulfilled" } | { status: "rejected"; reason: unknown }`. A synchronous throw is stored
as the same rejected settlement form. Only after this record is stored may abort, close, or shutdown
win the execution barrier. Await/race logic consumes the owned record and normalizes its rejected
reason through the existing bounded public-error path.

Cleanup retains that observation even when cancellation wins, so a late native/thenable rejection
cannot become unhandled. The RED table covers native promise, native rejection, foreign thenable,
and synchronous throw under abort/close/shutdown, with zero post-transition prompt effects,
exactly-once acquired unsubscribe/dispose/join ownership, and zero listeners, leases, referenced
timers, or `unhandledRejection` observations.

### Comprehensive test-only RED matrix

One test-only checkpoint adds exactly these six top-level rows and no artifact or production edit:

1. `cycle 16 punctuation-bearing assignments fail closed and preserve later siblings through every shared consumer`
2. `cycle 16 malformed quoted values resynchronize quoted siblings across line flow and quote styles`
3. `cycle 16 assignment lexer is one-pass monotonic across dense openers malformed flows and terminal siblings`
4. `cycle 16 finite public scalar grammar remains byte-identical while adjacent assignments fail closed`
5. `cycle 16 schema captures avoid every whole-key primitive across creation request event and handoff boundaries`
6. `cycle 16 prompt settlement is owned before synchronous abort close and shutdown barriers`

Rows 1, 2, and 4 route every protected marker through all 13 shared consumers: direct redaction;
task/context prompts; workspace read/edit/write; capability summary/reference; handoff
summary/finding/verification; policy error; and runtime error. Row 3 proves the exact exported work
metric on dense and retained controls. Row 5 includes known-schema positive controls and hidden-peer
instrumentation. Row 6 tables each termination cause against each return/rejection style and exact
cleanup/unhandled accounting.

RED acceptance is 143 executed: all 137 retained rows pass and exactly the six rows above fail only
their intended behavior assertions, with zero skip/cancel/todo. Focused strict TypeScript passes;
all five frozen production/test-control blobs remain exact except the single intended test file.
No production edit is allowed before that evidence is committed.

### Cohesive GREEN, REFACTOR, and verification contract

GREEN closes all six rows as one architecture change: a single lexer replaces rescans, fixed DTO
projectors replace whole-key capture, and prompt settlement is owned before the barrier. It may not
grow secret/public synonym lists, reintroduce speculative lookahead, enumerate unknown records, or
weaken cleanup ownership. After 143/143, REFACTOR removes superseded scanners, enumeration helpers,
and compatibility paths; any retained legacy metric field or generic record walker is a failed
refactor, not a harmless shim. If retained Cycle 10 fixtures expect rejection of inert undeclared
creation/extension peers, they will be aligned only after RED to the stricter non-discovery
contract: ignored peers are safer than enumerating them.

Declared gates are focused 143/143 behavior; focused and all 12 production strict TypeScript 5.9.3
against explicit Pi 0.80.6 roots; retained actual pinned no-tool/one-tool rows; explicit Pi 0.80.6
offline RPC; safe isolation excluding only controller/state-store; serialized complete Shepherd
classification; `git diff --check`; immutable-base/frozen-start ancestry; RUN-STATE JSON;
credential/dependency/Go/connector scans; source/instrumented zero whole-key capture; exact same 20
paths; and clean head. Both Cycle 15 reports must be re-read in full after GREEN/refactor before
evidence freeze. No self-review/integration, push, network, GitHub, live model/auth, credential,
service, Go/connector, `make`, main, parent, #478, or #479 mutation is authorized.

### Cycle 16 execution result

- Artifact PLAN `6e3b05c8`; comprehensive test-only RED `7ec93fcd`; cohesive GREEN `c30cfe88`;
  behavior-preserving exact-metric REFACTOR `af9222b1`.
- RED executed 143 rows: all 137 retained rows passed and exactly C16-01 through C16-06 failed,
  with zero skip/cancel/todo, focused strict TypeScript green, and every production/control blob
  frozen except the intended runtime-test file.
- GREEN/refactor passes 143/143 focused and 206/206 safe isolation. Both focused and all 12
  production strict TypeScript scopes pass against TypeScript 5.9.3 and explicit Pi 0.80.6 roots.
  Explicit offline RPC registers `pm-shepherd` and exits 0 with only expected settings-lock
  warnings.
- Complete Shepherd executes 280 tests: 249 pass and the unchanged 31 controller/state-store
  process-identity rows fail because the managed sandbox denies `spawn`; this gate is classified
  environment-blocked, never green.
- Both complete Cycle 15 reports were re-read after refactor. Runtime capture contains no whole-key
  primitive, the scanner contains no suffix-flow/quote lookahead or legacy metric field, the exact
  same 20 paths remain, and immutable-base/frozen-start ancestry is intact.
- Final production blobs at REFACTOR are runtime `cd754a1e9b5baddf738c163cbba4d9fd1f279527`,
  policy `b7e2296123fb6da2fb0122f9d879c8aacf9dd2d6`, and unchanged prompts
  `b762787b2a63b5b02f9591c7bf3fff46394738cc`. Parent orchestration owns fresh independent review,
  process-capable replay, integration, and delivery.

## Consolidated Exact-Head Correction Cycle 17 Plan

Frozen start is exact reviewed head `5f0bef9cae08dd9e6285dca7b95e089e2fda02ce`; immutable base
remains `e659d6f1b666f58748e2d8c86599ceb4bbc62ff8`; the allowlist remains the exact same 20 paths.
Both `/tmp/475-REVIEW-CYCLE16-1.md` and `/tmp/475-REVIEW-CYCLE16-2.md` were read in full before
this plan. Their ten reported blockers plus one evidence warning deduplicate to the nine contracts
below. None is declined, deferred, or treated as a serial patch.

### One architectural correction

Cycle 17 replaces the scanner/capture/adoption boundary as one cohesive change:

1. `redactSensitiveText` scans the immutable original source. It records original-coordinate,
   half-open redaction ranges and renders only after lexical classification; no pre-redaction may
   change the coordinate space presented by the metrics.
2. The lexer owns one monotonic main cursor and explicit typed state: line/member start, key,
   separator, scalar, quoted scalar, composite, comment, and malformed recovery, plus a typed
   `}`/`]` closer stack bounded only by source length. The 257th opener retains its type; there is
   no untyped overflow depth. Failed member candidates terminate at the boundary already reached
   by the main cursor and can never launch a suffix scan.
3. A public-field fast path is admitted only after the lexer proves a finite scalar. A public
   quote must close before its current line/container boundary. `{` and `[` values remain under
   normal descendant traversal, so protected members are classified; malformed public values
   conservatively redact the proven remainder while resynchronizing at a typed sibling boundary.
4. Private-key, URL/basic-auth, structured assignment, range merge, and render work share one
   original-source ledger. Any retained pass is named and charged; helper ranges are disjoint or
   amortized, never suffix rescans. Metrics are observations, not initialized conclusions.
5. Each accepted closed host capability schema is descriptor-captured and compiled once into a
   bounded validator/projector. Unknown, missing, wrong-type, and out-of-range input fails before
   host execution. The projected, deeply frozen DTO is the only callback input and the same
   projector/DTO definition supplies tool-call lifecycle identity; `target` is not a hard-coded
   generic identity shortcut.
6. Prompt ownership publishes one always-fulfilled internal settlement before invoking Pi. The
   entire return adoption and handler installation is inside one fallible owned boundary. Fresh
   native adoption observes supported native and foreign fulfillment/rejection and converts every
   assimilation throw into a settled rejection joined by abort, close, and shutdown.
7. The trusted Pi adapter contract is explicit: Pi must return an exact intrinsic native Promise
   with direct `Promise.prototype` and no own `constructor` or `then`; a wider adapter must normalize
   and own its result before exposing the session. Foreign thenables are assimilated through a
   fresh intrinsic promise whose getter/callback failures are owned. An already-rejected native
   promise with an unobservable poisoned own `then` cannot be made safe after return in JavaScript;
   Cycle 17 rejects unsupported native shapes as adapter-contract failures instead of claiming
   impossible hostile-promise support or manufacturing an unowned rejection in its tests.
8. `buildRolePrompts` descriptor-captures the complete public input, including `role` and every
   binding scalar, into a fresh deeply frozen DTO before route selection, validation, rendering,
   or JSON serialization. Proxies, accessors, inherited authority, and non-approved prototypes are
   rejected without invoking caller behavior; one snapshot drives route/system/user/schema data.
9. Every remaining schema/argument/result capture boundary uses fixed own descriptors and requires
   its exact approved direct prototype before bounded traversal. `for...in` is prohibited; inherited
   proxy prototypes and enumerable peers have zero influence and zero trap calls.

### Formal lexer state, coordinates, and work accounting

Let the original source be `S`, `n = S.length`, main cursor `i`, typed closer stack `D`, lexical
mode `q`, pending candidate start `k`, and ordered redaction ranges `R`. The transition relation is
`(i,q,D,k,R) -> (j,q',D',k',R')` with `j > i` for every consuming transition and no transition to
an earlier source offset. Every opener pushes its exact closer; only the matching closer pops it.
A mismatch enters conservative malformed recovery without changing unrelated stack entries.
Ranges satisfy `0 <= start < end <= n`; merge is ordered and output is assembled from `S` and the
merged ranges, so expanding replacement text never changes a measured coordinate.

The authoritative work record charges actual events:

- `cursorAdvances`: original-source main-cursor advances, never transformed-output length;
- `maxMainCursorVisits`: maximum actually observed visits to one original offset (zero for empty,
  exactly one for a completed non-empty main scan), updated by visit events rather than assigned;
- `keyCharacterVisits` and `boundaryCharacterVisits`: real bounded helper/token transitions;
- every named credential/private-block/range-merge/render pass contributes to `totalWork`.

No helper may read from a candidate to an unknown suffix end. Per-offset main visits are at most
one, each token/helper owns a disjoint or amortized source interval, and the fixed number of named
passes yields `totalWork <= 8n + 64`. The RED work row installs a fail-fast accessor on
`totalWork`, uses geometric sizes, and includes dense plain/quoted/whitespace failed members,
expanding and shrinking credential replacements, later protected siblings, and real main/helper
counts. A metric constant or omitted pass must fail the row.

### Comprehensive RED matrix

After this artifact-only checkpoint, one test-only commit adds exactly nine top-level rows:

| ID | Contract / executable matrix |
|---|---|
| C17-01 | dense plain, quoted, and whitespace failed flow-member candidates at geometric sizes remain within `8n + 64`, fail fast, and redact a later sibling |
| C17-02 | malformed public single/double quotes cannot cross line or typed container bounds through direct redaction and all 12 indirect shared consumers |
| C17-03 | registered closed host schemas reject extra/missing/type/range input before callback and project valid non-`target` fields into distinct lifecycle identities |
| C17-04 | exact-native and foreign fulfill/reject, synchronous throw, throwing then getter/callback, and double settlement join under abort/close/shutdown; unsupported fulfilled native shapes fail the explicit adapter contract without an orphan |
| C17-05 | exported prompt building rejects accessor/proxy/inherited role/binding input and renders one post-mutation-stable role/binding snapshot everywhere |
| C17-06 | public fields preserve only proven finite scalars; valid/malformed object/sequence descendants redact through every shared consumer |
| C17-07 | mixed delimiter nesting at 255/256/257 keeps delimiter identity and protects terminal siblings |
| C17-08 | expanding/shrinking credentials, dense candidates, and every claimed pass use one original coordinate system with measured visits/work |
| C17-09 | schema, capability-result, and workspace-result prototype proxies cause zero traps; exact prototypes and no `for...in` influence are required |

Expected RED is 152 executed: all 143 retained rows pass and exactly C17-01 through C17-09 fail
their behavior assertions, with zero skip/cancel/todo. Focused strict TypeScript must pass. Frozen
blobs before RED are runtime `cd754a1e9b5baddf738c163cbba4d9fd1f279527`, policy
`b7e2296123fb6da2fb0122f9d879c8aacf9dd2d6`, prompts
`b762787b2a63b5b02f9591c7bf3fff46394738cc`, runtime test
`b3df3c5f0bf6c0f02b8576200066cd683a994480`, and policy test
`b7fa0a1c1f1fbfb1ec0b10b9fbd022229e84e56f`.

### Ordered execution and gates

Manual-GSD fallback remains explicit: `scripts/gsd doctor` passes 69 commands, while
`scripts/gsd prompt programming-loop ...` reports unknown command. Required skills used are
`gsd-programming-loop`, `architecture-patterns`, `javascript-testing-patterns`,
`typescript-advanced-types`, and `github-issue-first-delivery`; the repository routing, issue
contract, universal runtime loop, GSD adapter, and runtime integration references were also read.
The order is artifact PLAN -> comprehensive RED -> one cohesive GREEN -> REFACTOR -> both-report
replay -> VERIFY. No production edit precedes committed RED evidence.

Declared terminal gates remain focused 152/152; focused and all-production strict TypeScript 5.9.3
against explicit Pi 0.80.6 roots; retained actual pinned sessions and offline RPC; safe isolation;
serialized complete-suite classification; diff, ancestry, JSON, credential/dependency/Go/connector
scans; source guards; exact 20 paths; and clean worktree. No dependency, parent, #478, network,
push, integration, live model/auth, credential, service, Go/connector, `make`, or main mutation is
authorized.

### Cycle 17 execution result

- Artifact PLAN `15a0b70f2b94ccd14f22dd0bcad410d512fb8c4f`; comprehensive test-only RED
  `ed8e79e3489e5826b1be8078c32c01d945256ea7`; cohesive GREEN
  `e7dfbb358824efb5423e284ee2c6a78ea2f3cd30`; behavior-preserving REFACTOR
  `81c34cc5b7db9bffe4bd27d17122007d47ecedb6`.
- RED executed all 152 focused rows: all 143 retained rows passed and exactly C17-01 through
  C17-09 failed their intended behavior assertions, with zero skipped/cancelled/todo, focused
  strict TypeScript green, and every production blob frozen.
- GREEN/refactor passes 152/152 focused and 215/215 safe isolation, the latter comprising the
  retained 206 isolation rows plus the nine Cycle 17 rows. Both focused and all 12 production
  strict TypeScript scopes pass against TypeScript 5.9.3 and explicit Pi 0.80.6 roots.
- Explicit pinned Pi 0.80.6 offline RPC exits 0 and registers `pm-shepherd`. Serialized complete
  Shepherd executes 289 tests: 258 pass and the unchanged 31 controller/state-store rows fail only
  because the managed sandbox denies process creation with `spawn EPERM`; zero rows are skipped,
  cancelled, or todo. That broad gate remains environment-blocked and is not called green.
- Diff/source/ancestry checks retain the exact 20 authorized paths against immutable base
  `e659d6f1b666f58748e2d8c86599ceb4bbc62ff8`. No feature surface or external mutation was added.
  Local evidence is terminal, but `verificationPassed` remains false: two independent exact-head
  reviews of the final evidence candidate are still pending and parent-owned, as are the
  process-capable replay, integration, and delivery gates.

## Consolidated Exact-Head Correction Cycle 18 Plan

Frozen start is exact independently reviewed candidate
`687d053df5f7e7d08c4cab7d2a2d8f153850e673`; immutable base remains
`e659d6f1b666f58748e2d8c86599ceb4bbc62ff8`; the allowlist remains the same exact 20 paths.
Both `/tmp/475-REVIEW-CYCLE17-1.md` and `/tmp/475-REVIEW-CYCLE17-2.md` were read in full before
this plan (SHA-256 `fb309309fe47002b6a803f8c4558e22d278f477113f074cf9607c1befcdfc4a2`
and `4d65a09fbaa8fe2e6557007545af78fb0d2ca4c4ba1545d8eb1d2558fdceb2d1`). Their ten
reported observations deduplicate to the seven blocking contracts below. None is declined,
deferred, or answered by expanding a blacklist or claiming generic JSON Schema support.

### One lexer, one finite contract, one descriptor boundary

1. **Quoted flow ownership.** While a value is quoted, escape/doubled-quote handling and the
   matching close are evaluated before any comma, semicolon, or closer. Punctuation inside an open
   quote is value data, never a member delimiter. An unclosed quote owns and conservatively redacts
   through its proven physical-line or typed-container boundary; recovery may resume only at a
   delimiter proven to belong to that same typed depth. No protected tail or later sibling may
   escape through direct redaction or any of the 12 indirect shared consumers.
2. **Candidate reset and public proof.** Locator state suppresses only its internal `://` colon.
   Every flow comma, semicolon, and matching typed closer first terminates and clears the current
   candidate, then reprocesses that boundary exactly once. No candidate bit survives a completed
   member. A public value is preserved only after the lexer proves the entire finite scalar at its
   boundary; locator-like or partially public text cannot shield a later protected assignment.
3. **Sensitive scalars are default-secret.** `token`, `password`, `passwd`, `secret`, and every
   other protected key redact one-, two-, and many-word plain scalars. Whitespace alone never makes
   a protected scalar documentary or public. The only documentary escape is the existing finite
   prefix grammar in `isBoundedDocumentaryProse`; retained prose controls prove fidelity.
4. **Finite supported host schema.** Registration compiles and canonicalizes the sole effective
   contract. Supported node vocabularies are exact own-enumerable data fields:
   object `{type, additionalProperties, properties, required}`; string
   `{type, minLength?, maxLength?, enum?}`; integer/number
   `{type, minimum?, maximum?, enum?}`; boolean `{type, enum?}`; and bounded array
   `{type, items, minItems?, maxItems?}`. Object `additionalProperties` is exactly `false`; its
   exact property-name set equals its dense authoritative `required` vector, so optional host
   fields do not exist. Enums are non-empty exact arrays whose every value matches the node type
   (safe integers for `integer`). `pattern`, `format`, `$ref`, combinators, nullable/multi-type,
   defaults, optional properties, unknown keywords, and mixed enums fail registration at any
   depth. Hidden/symbol peers remain inert and have no authority. Pi `parameters` are rebuilt from
   the compiled representation; Pi validation, `projectArguments`, direct `execute`, callback DTO,
   and every tool lifecycle identity consume that same frozen canonical contract and projected DTO.
5. **Exact descriptor DTOs at tool/event/result boundaries.** Tool policy captures reflection
   intrinsics once, rejects proxy or unapproved direct prototypes before descriptor work, and
   snapshots the public policy envelope, authority, workspace ports, capability contracts, tool
   event envelopes, and workspace/capability results through fixed own enumerable data descriptors.
   Accessors/setters/inherited fields reject without invocation. No later step directly rereads a
   caller envelope; a fresh frozen DTO is the only behavior-affecting input. The historical Cycle 9
   getter-acceptance fixture is explicitly superseded and may be aligned only after RED is frozen.
6. **Exact runtime arrays.** Every dense array entering `agent-session-runtime.ts` must have direct
   prototype equal to the captured intrinsic `Array.prototype` before length, index, SDK, tool, or
   lifecycle lookup. This includes request context/capabilities/read prefixes/write prefixes/
   capability names and all creation, active-tool, event content/messages/tool-results/diagnostics,
   and terminal handoff arrays. Proxy-backed prototypes reject with zero traps and zero later
   side effects; exact arrays retain current behavior.
7. **Transition-derived work accounting.** The main cursor and all metrics use original UTF-16
   source coordinates. `cursorAdvances` charges each original-source consume; key and lexical
   counters charge only their actual transitions; each of the five strong recognizers is charged
   once per character it examines; frame/recovery actions, every emitted and examined range, every
   range insert/coalescence, each replacement emission, and each original source unit actually
   rendered are charged at the operation. A no-range identity return has no fictional render pass.
   `totalWork` is the sum of these increments, never a bulk asserted formula; per-offset main visits
   are measured (zero for empty, one for completed non-empty input). The independent RED oracle
   covers empty/one-byte/public/protected/quoted/locator/private-key/overlap/recovery inputs, setter
   fail-fast probes, exact equality, and the derived ceiling `totalWork <= 16n + 64`.

### Comprehensive seven-row RED matrix

After this artifact-only checkpoint, one test-only commit adds exactly seven named top-level rows:

| ID | Exact row / executable boundary |
|---|---|
| C18-01 | `cycle 18 quoted flow punctuation stays inside owned quotes and cannot expose protected siblings` — both quote styles × comma/semicolon × valid/malformed flow ownership through all 13 consumers |
| C18-02 | `cycle 18 locator and public scalar candidates reset at proven flow boundaries` — mapping/sequence, quoted/unquoted locator controls, multiple later sibling positions, exact preservation plus marker absence and work bound |
| C18-03 | `cycle 18 multiword sensitive scalars redact while bounded documentary prose remains public` — one/two/many-word values for every affected key through all consumers plus retained prose controls |
| C18-04 | `cycle 18 one finite compiled host schema governs Pi execution and lifecycle identity` — root/nested optional and unsupported keywords, mixed enums, exact supported nested controls, real Pi validator/direct callback/projector/event identity equality |
| C18-05 | `cycle 18 tool event and result envelopes use exact descriptor DTOs without caller callbacks` — root/nested policy, workspace/capability, result, and tool-event accessors/proxies/inheritance/mutation with zero getter/trap/re-read counts |
| C18-06 | `cycle 18 every runtime dense array requires the exact intrinsic Array prototype` — request, SDK result, active-tool, event, and handoff arrays reject proxy direct prototypes before any index/SDK/lifecycle side effect |
| C18-07 | `cycle 18 redaction work metrics exactly charge recognizers ranges and render in original coordinates` — independent exact oracle, per-counter setters, range overlap/coalescing, identity render, max visits, and affine bound |

Expected RED is 159 executed: all 152 retained rows pass and exactly C18-01 through C18-07 fail
their behavior assertions, with zero skip/cancel/todo. Focused strict TypeScript must pass. Frozen
production blobs through RED are runtime `66c92cf368746b9fcf5ba3fdc5cd28aebc21a8e4`, policy
`00d8482d4f320fb948abcbef893e87cf0690d1a3`, and prompts
`c5b6c27fc1ba6f738fbfd36d49d38c94c7b13b73`; frozen controls are runtime test
`e9fb05b8d1dd5b438cd66da707c7549f33e754c6` and policy test
`b7fa0a1c1f1fbfb1ec0b10b9fbd022229e84e56f` before the one intended RED edit.

### Ordered execution and boundaries

Baseline at the exact start is focused 152/152 and focused strict TypeScript 5.9.3 against explicit
Pi 0.80.6 roots. `scripts/gsd doctor` passes 69 commands; the adapter again rejects
`programming-loop`, so the explicit manual-GSD PLAN -> RED -> cohesive GREEN -> REFACTOR -> VERIFY
fallback remains recorded. Skills used are `gsd-programming-loop`, `architecture-patterns`,
`javascript-testing-patterns`, `typescript-advanced-types`, and `github-issue-first-delivery`, plus
the repository routing, runtime/Pi, issue, adapter, and universal-loop references. A requested
read-only mapper could not start because the runtime agent-thread limit was saturated; execution
decision is `not_spawned_runtime_capability_missing`, and the collision-heavy seven-row RED remains
on the local critical path.

No production edit precedes the committed RED evidence. No dependency, parent, #478, network,
push, integration, live model/auth, credential, service, Go/connector, `make`, main, prompt-return
contract, or path-scope change is authorized. GREEN must close all seven contracts together; the
trusted Pi 0.80.6 prompt adoption/settlement behavior from Cycle 17 is frozen and retained.

### Cycle 18 execution result

The ordered checkpoints are artifact-only PLAN `aa92f751`, comprehensive test-only RED
`50b04a24`, cohesive GREEN `ea23780f`, and operation-accounting REFACTOR `6354b156`. RED executed
159 focused rows: all 152 retained rows passed and exactly C18-01 through C18-07 failed, with zero
skipped/cancelled/todo, focused strict TypeScript passing, and all production blobs frozen.
GREEN/refactor passes 159/159. Final blobs are runtime
`e952557d987ef6bcba3e99ac4a7820fefc0a0ce3`, policy
`efc7564ec0adc8a424c30d62cab97f1f4fca7a53`, unchanged prompts
`c5b6c27fc1ba6f738fbfd36d49d38c94c7b13b73`, runtime test
`d918930205120d6a491a6288a95ee14550a0c567`, and policy test
`39fb20e0dc4667b9743c5acc4f87223b01128788`.

The `16n + 64` ceiling is derived from complete charging rather than a relaxed linearity claim.
`totalWork` charges the main consume/boundary step; each of five strong recognizers once for each
UTF-16 source unit it examines; key and lexical transitions; frame push/pop, recovery, and
finalization; every range emission, examination, insertion, and coalescence; every replacement
emission; and each original source unit actually rendered. Identity returns with no ranges charge
no fictional render. The main cursor still visits each completed non-empty source offset at most
once and zero offsets for empty input.

Terminal local gates pass focused 159/159, safe isolation 222/222 (retained 215 plus seven), strict
TypeScript 5.9.3 over all 12 Shepherd production files with explicit Pi 0.80.6 roots, and pinned
offline Pi 0.80.6 RPC registration. The broad serialized suite is honestly environment-blocked at
265/296 by the unchanged 31 controller/state-store `spawn EPERM` failures. Exact ancestry, JSON,
diff/source guards, the same 20-path issue scope, and the 13-path Cycle 18 delta pass. Independent
exact-head reviews and process-capable replay remain pending; `verificationPassed` stays false.

## Consolidated Exact-Head Correction Cycle 19 Plan

Frozen start is the exact independently reviewed evidence candidate
`a0cd1057a0da642185f10b4ddfe72263602c7513`; candidate tree is
`d428be42e501d96a9e4d197738d5d5326f25e322`; immutable base remains
`e659d6f1b666f58748e2d8c86599ceb4bbc62ff8`; and the allowlist remains exactly 20 paths. Both
`/tmp/475-REVIEW-CYCLE18-1.md` and `/tmp/475-REVIEW-CYCLE18-2.md` were read in full before this
plan (SHA-256 `48144b90ee7ba8dca980c30d1e1acac1254e72ce71182d5291f1cd955cab971b`
and `13b3b492fbd67126bdcb39707d385dc54a1f5628a5e478457f49cdc84aafdf18`). Their nine reported
observations deduplicate to the five contracts below; none is declined, deferred, or narrowed by
claiming generic JSON Schema compatibility.

### One finite normalized capability contract

Pi 0.80.6's installed `validateToolArguments` clones and converts plain JSON-schema arguments
before validation. Cycle 19 therefore defines one finite normalization rather than advertising
exact-type semantics Pi does not provide:

- `string`: a string remains exact; `null` becomes `""`; a finite number or boolean becomes its
  JavaScript string form. The canonical schema always emits `minLength` (default 0) and a bounded
  `maxLength` no wider than the active policy write ceiling; both bounds count Unicode code points,
  matching Pi/JSON Schema rather than UTF-16 code units.
- `integer`: `null` becomes 0, booleans become 0/1, and a string containing at least one
  non-whitespace character is admitted only when `Number(value)` is integral. The normalized value
  must be a safe integer. Every canonical integer
  schema emits the author bounds intersected with
  `[-9007199254740991, 9007199254740991]`.
- `number`: `null` becomes 0, booleans become 0/1, and a string containing at least one
  non-whitespace character is admitted only when `Number(value)` is finite. The normalized value
  must remain finite.
- `boolean`: `null` becomes false; only exact strings `"true"`/`"false"` and numbers 1/0 convert;
  every other non-boolean rejects.
- Every normalized numeric `-0` becomes canonical `+0` before bounds, enum comparison, callback,
  event identity, or serialization. Numeric schema bounds and enum members use the same
  canonicalization; enums use SameValueZero identity and canonical duplicates reject.
- Arrays remain exact intrinsic dense arrays and every canonical schema emits `minItems` (default
  0) and `maxItems` (default and hard ceiling 512). Closed objects remain required-only. Optional
  fields, unknown keywords, unions/combinators, and other unsupported semantics still reject.

The compiler-built schema is Pi's only registered surface. The same compiler normalizes and
descriptor-projects direct input, Pi/event arguments, callback DTOs, and lifecycle identities into
one frozen DTO. Real pinned-Pi tests cover unsafe integers, every supported coercion family, an
omitted 512-item ceiling, signed zero, canonical enum identity, direct execution, callback, event,
and lifecycle correlation. If any primitive cannot be proven equal to Pi 0.80.6 under this finite
matrix, GREEN must reject that primitive schema at registration rather than advertise drift.

### Captured reflection and exact terminal arrays

All reflection/serialization capabilities are captured once at module initialization and used
exclusively afterward: object/array prototypes, `Object.keys`/fixed descriptor reads, create,
define, freeze, has-own/value discovery, `Array.isArray`, proxy detection, `Reflect.apply`, and
JSON parse/stringify where those operations cross schema, raw input, event, result, reference, or
error boundaries. No callback can replace an ambient primitive after import and influence a later
classification. Error capture is descriptor-based or reduced to a constant sanitized fallback;
it never invokes caller getters.

`changedPaths`, `verification`, and `findings` use the same exact array rule as request/event arrays:
captured intrinsic `Array.prototype`, non-proxy root and direct prototype, authoritative own length
descriptor, dense enumerable own data indexes, maximum 32, and one fresh frozen snapshot. Mapping
is indexed over that snapshot only; caller `.map` and inherited iteration are never consulted.

### Independent category work trace

`RedactionScanMetrics` exposes one field for every charged class:

- existing `cursorAdvances`, `boundaryCharacterVisits`, and `keyCharacterVisits`;
- `recognizerCharacterVisits`, one per character examined by each of the five strong recognizers;
- `lexicalTransitions`, `frameOperations`, and `recoveryTransitions`;
- `rangeEmissions`, `rangeExaminations`, `rangeInsertions`, and `rangeCoalescences`;
- `replacementEmissions` and `renderedSourceUnits` over original coordinates.

`totalWork` is exactly the sum of those public category counters. No uncategorized work charge is
allowed. The RED oracle uses hard-coded, test-side expected traces for empty, one-character
identity, frame, recovery, protected assignment, private-key, and overlapping/coalescing fixtures.
It then omits each category from a copied expected trace and proves the total no longer matches.
Unit-write and `16n + 64` ceiling assertions remain secondary guards; neither supplies the oracle.

### One shared incremental projection budget

Every compiled projection receives one mutable budget before root inspection. It charges before
descent or output insertion and is shared across the complete graph:

- at most 4,096 projected schema-node executions;
- at most 4,096 object-key descriptor examinations;
- at most 4,096 dense array-item descriptor examinations;
- root depth 0 through maximum depth 64;
- at most 4,096 object/array identity encounters, with cycles and any repeated DAG identity
  rejected before the second descent;
- exact incremental encoded JSON bytes no greater than
  `min(callMaximumBytes, MAX_TOOL_INPUT_BYTES)`.

Braces/brackets, commas, quoted keys, colons, scalar encodings, and escaped string bytes are charged
as they are admitted. Per-node safe-integer, string-length, and array-cardinality limits appear in
the canonical Pi schema because JSON Schema can express them. Aggregate nodes/keys/items/depth/
identity/encoded-byte limits are enforced by the same projector on direct, callback, and event
paths; they are not misrepresented as standard schema keywords. Near-limit trees retain behavior;
nested fanout, cycles, and shared DAG aliases reject before multiplicative traversal/allocation.

### Comprehensive five-row RED matrix

After this artifact-only checkpoint, one test-only commit adds exactly five top-level rows:

| ID | Exact executable boundary |
|---|---|
| C19-01 | `cycle 19 Pi direct callback event and lifecycle share one finite normalized schema` — real Pi 0.80.6 unsafe integers, full supported coercion table, implicit 512-item limit, `+0`/`-0`, enum canonicalization, and exact normalized DTO identities |
| C19-02 | `cycle 19 every schema raw event result reference and error path uses captured reflection` — post-import poisoning of each admitted reflection primitive before/after awaited host and workspace callbacks, success/rejection paths, zero poison calls |
| C19-03 | `cycle 19 terminal handoff arrays use exact intrinsic dense snapshots without caller map` — changed paths, verification, and findings x direct proxy/proxy prototype/accessor/sparse/post-capture mutation with zero traps and frozen DTOs |
| C19-04 | `cycle 19 redaction work trace independently accounts for every category` — hard-coded exact traces plus one-category omission checks and retained affine/unit-write guards |
| C19-05 | `cycle 19 one incremental projection budget rejects before multiplicative traversal` — nodes/keys/items/depth/bytes/DAG counters, nested shared aliases, pre-allocation rejection, expressible Pi limits, and near-limit controls |

Expected RED is 164 executed: all 159 retained rows pass and exactly C19-01 through C19-05 fail
their intended behavior assertions, with zero skipped/cancelled/todo. Focused strict TypeScript must
pass. Frozen production blobs through RED are runtime
`e952557d987ef6bcba3e99ac4a7820fefc0a0ce3`, policy
`efc7564ec0adc8a424c30d62cab97f1f4fca7a53`, and prompts
`c5b6c27fc1ba6f738fbfd36d49d38c94c7b13b73`; pre-RED test blobs are runtime test
`d918930205120d6a491a6288a95ee14550a0c567` and policy test
`39fb20e0dc4667b9743c5acc4f87223b01128788`.

Baseline at the exact start is focused 159/159 and focused strict TypeScript 5.9.3 against explicit
Pi 0.80.6 roots. `scripts/gsd doctor` passes 69 commands while the adapter rejects
`programming-loop`; explicit manual-GSD PLAN -> RED -> GREEN -> REFACTOR -> VERIFY remains active.
Skills used are `gsd-programming-loop`, `architecture-patterns`, `javascript-testing-patterns`,
`typescript-advanced-types`, and `github-issue-first-delivery`, plus repository routing, runtime/Pi,
issue, adapter, and universal-loop guidance. The four-agent runtime is saturated, so no read-only
mapper was spawned; Cycle 19 records `not_spawned_runtime_capability_missing` and stays on the local
critical path.

No production edit precedes committed RED. No dependency, prompt/lease/route contract, parent,
#478, network, push, integration, live model/auth, credential, service, Go/connector, `make`, main,
or path-scope change is authorized.

### Cycle 19 execution and correction evidence

The ordered implementation chain is:

1. artifact PLAN `337cba1731178d1b7ef51c62ec45b15159b3cca3`;
2. test-only RED `8519df27dc0617332f2273ac38ff6d82e59a813e`;
3. cohesive GREEN `6edb1d5bcdb3ccf780e2153cbef238eb4f00cf17`;
4. post-GREEN review fix `e9bdddd03e2fee4e4db791eec17a63233698e67a`.

RED changed only `agent-session-runtime.test.ts`. It executed 164 rows, retained all 159 prior
passes, and failed exactly C19-01 through C19-05, with zero skipped/cancelled/todo. Strict
TypeScript passed, and production stayed byte-exact at runtime
`e952557d987ef6bcba3e99ac4a7820fefc0a0ce3`, policy
`efc7564ec0adc8a424c30d62cab97f1f4fca7a53`, and prompts
`c5b6c27fc1ba6f738fbfd36d49d38c94c7b13b73`. The RED runtime-test blob was
`8bbe1143cae474e647866a5fe6c04b020da39cc1`.

Initial GREEN made all five rows and the complete 164-row focused suite pass. The required
post-GREEN read-only audit then found four real correction points: whitespace-only numeric strings
were accepted directly but rejected by Pi; UTF-16 length rejected a one-code-point astral string;
runtime event/error paths still reached ambient reflection and invoked an untrusted message getter;
and the category-omission matrix did not actually recompute a copied trace. The review-fix commit
requires trimmed numeric source, counts Unicode code points allocation-free, captures every direct
runtime reflection primitive, reads Error messages only from captured own data descriptors, and
zeros each positive category in a copied metric trace before recomputing the total. A temporary
ambient `Object.freeze` mutation sentinel made C19-02 fail exactly once in each runtime phase, then
was removed and the row returned green, proving the direct-caller poison probe is non-vacuous.

Terminal evidence is five Cycle 19 rows passing, exact retained replay 164/164, safe isolation
227/227 (the retained 222 plus five), focused and all-12-production strict TypeScript passing with
explicit Pi 0.80.6 roots, and explicit pinned Pi 0.80.6 offline RPC registration. The serialized
broad suite executes 301 rows: 270 pass and only the unchanged 31 controller/state-store process-
identity rows fail because the managed sandbox returns `spawn EPERM`; zero rows are
skipped/cancelled/todo, so the gate is environment-blocked and not green. `git diff --check`, JSON,
source/blob guards, exact-start/PLAN/RED/GREEN/review-fix ancestry, and the same 20 immutable-base
paths pass. Final blobs before the
artifact summary commit are runtime `e678193e5ff6022b0ac40628dfd9fc07a928bb2a`, policy
`499f5a0b6cd0730c1279b1d33728604cb06c09c9`, unchanged prompts
`c5b6c27fc1ba6f738fbfd36d49d38c94c7b13b73`, runtime test
`d25a8e1b25df3b80058e54f5ad3e7072d1362ad0`, and unchanged policy test
`c744c7d3d1ea798a30a214d94c9f44165cf25d84`. Parent-owned independent exact-head reviews and the
process-capable broad replay are still pending, so `verificationPassed` remains false.

## Consolidated Exact-Head Correction Cycle 20 Plan

Frozen start is the exact independently reviewed Cycle 19 evidence candidate
`427cad088c8cbd5c14bfcb70ff2b3fa165b55e86`; candidate tree is
`f5057e96454b93bbd9dbe997ce96935a13aec573`; immutable base remains
`e659d6f1b666f58748e2d8c86599ceb4bbc62ff8`; and the allowlist remains exactly 20 paths. Both
`/tmp/475-REVIEW-CYCLE19-1.md` and `/tmp/475-REVIEW-CYCLE19-2.md` were read completely before this
plan (211 and 244 lines; SHA-256
`04c3a1ebf695459a638bcf83b54a8cb3ac305cf8112d44f5e3df08dfab58358d` and
`8deafaf0e90056e50d3d7b126df949658fe2097f72d26b56112c5586912c51c2`). Their five independently
reproduced blockers consolidate without deferral into the four contracts below. Cycle 19's exact
handoff-array snapshots and independent redaction-category oracle remain byte-for-byte retained.

### C20-01 — one captured intrinsic realm across every awaited boundary

The runtime and policy capture or eliminate every mutable numeric, text, UTF-8, error, collection,
reflection, and classification capability they use after module initialization. The reviewed set
is `Number` conversion plus `isFinite`, `isInteger`, `isSafeInteger`, `parseInt`, and safe-integer
constants; `String` conversion plus `charCodeAt`, `codePointAt`, `trim`, and `fromCharCode`; the
used `Math` operations; JSON receiver plus parse/stringify; any `TextEncoder` constructor/method;
any `Intl.Segmenter` constructor/segment operation; `Error`/`AggregateError` constructors and
approved prototypes; the `WeakSet` constructor and used identity-set operations; used `nodeTypes`
predicates; Array constructor/prototype/static and used
prototype operations; Object constructor/prototype and used reflection methods; and used Reflect
operations. Captured methods are invoked only through the captured `Reflect.apply` with their
captured receiver. Redundant conversions and mutable prototype dispatch are removed rather than
captured when indexed loops or exact type checks suffice. Projection uses a manual UTF-8 counter,
so neither `TextEncoder` nor `Intl` is consulted after import.

The executable poison matrix replaces globals and writable properties only after the dynamic
modules load, activates them before and after awaited workspace/host/session callbacks, and proves
zero poison calls for accepted and rejected schema, raw argument, result, reference, event,
lifecycle, handoff, primary-error, cleanup-error, aggregate-error, and quarantine paths. It also
asserts invariant Pi/direct/callback/event/handoff DTOs and byte enforcement. All substitutions
are restored by captured test-side descriptors in `finally`.

### C20-02 — exact pinned Pi grapheme constraints and bounded scalar JSON bytes

String `minLength`/`maxLength` mirrors the installed TypeBox 0.80.6 guard, including its fast path,
instead of claiming generic code-point behavior:

1. The fast scan uses captured `charCodeAt` and falls back to the full cluster algorithm on a high
   surrogate (`D800-DBFF`), a `0300-036F` combining mark, or ZWJ (`200D`). It otherwise performs
   TypeBox's exact early min/max comparison over UTF-16 units.
2. The full `NextGraphemeClusterIndex` uses captured `codePointAt`; consumes one initial code point;
   consumes combining ranges `0300-036F`, `1AB0-1AFF`, `1DC0-1DFF`, and `FE20-FE2F` plus variation
   selectors `FE00-FE0F`; consumes every subsequent `ZWJ + code point + modifiers`; and consumes
   one adjacent regional-indicator partner for an initial `1F1E6-1F1FF` code point.
3. A valid surrogate pair consumes two UTF-16 units as one point. Lone high and lone low
   surrogates consume one unit each, exactly matching the pinned guard. No locale, host ICU, or
   ambient `Intl` decision is admitted.

The real-Pi matrix preserves the pinned fast-path quirks as well as the full cluster behavior:
base-plus-combining, combining-only, multi-ZWJ emoji, regional-indicator flags, lone high/low
surrogates, high-surrogate-plus-combining, and high-surrogate-plus-ZWJ-plus-ASCII pass
`maxLength: 1`; high-surrogate-plus-ASCII, low-surrogate-plus-ASCII, and dangling
high-surrogate-plus-ZWJ fail. ASCII plus VS16 and ASCII plus a `1AB0-1AFF` mark also follow the
pinned fast result (two units, rejection at max one), even though the full counter alone would
group them. This is pinned compatibility rather than a claim of full Unicode segmentation.

Projected JSON byte charging never serializes or encodes a complete untrusted scalar before the
budget decision. For strings it charges the two quotes first, performs an O(1) source-unit lower-
bound check, then walks captured UTF-16 code units once and stops at the first over-budget unit.
It charges escaped quote, escaped backslash, and the five short control escapes as two ASCII bytes,
other `U+0000-001F` and lone surrogates as six-byte `\\uXXXX`, unescaped ASCII as one byte,
`0080-07FF` as two, ordinary BMP as three, and valid surrogate pairs as four. JSON punctuation and
quoted property names use the same counter. `null` and booleans use fixed bounded encodings. Finite
canonical numbers use only a bounded captured canonical representation after raw numeric-string
work has been pre-charged; no 100,000-unit source reaches `trim`, numeric conversion, stringify,
encode, or an observer once its scalar-work or byte budget is exhausted.

One bounded traversal couples the exact pinned length state with JSON/UTF-8 charging so byte
exhaustion stops further source visits even when the length bound would otherwise continue. The
one shared projection state therefore adds a scalar-normalization work counter to the retained
node/key/item/depth/container/DAG/exact-byte counters. A string's authoritative `.length` is read
in O(1) and charged before grapheme, trim, or numeric conversion; character visits and emitted
bytes are then incremental. Exact near-bound controls cover quotes, backslashes, every control
escape family, BMP two/three-byte forms, astral pairs, and malformed surrogates.

### C20-03 — descriptor-only exact runtime failure snapshots

Failure normalization classifies against captured exact native error prototypes/constructors and
never consults a later ambient `instanceof`, constructor, `Symbol.iterator`, collection method, or
caller property read. `message`, `name`, `stack`, `cause`, and AggregateError `errors` are read only
with the captured own-property-descriptor primitive; only own data descriptors are admitted.
Accessor, inherited, proxy, malformed, or unavailable fields become bounded constant sanitized
fallbacks.

Aggregate members are accepted only when the own `errors` data descriptor contains an exact
non-proxy intrinsic dense array: exact captured Array prototype, authoritative non-enumerable own
length, at most 16 entries, and enumerable own data indexes. Members are copied by indexed
descriptor reads into a fresh dense snapshot before recursive normalization. No iterable factory,
`next`, step `done`/`value`, `return`, getter, array iterator, or caller callback is invoked. Cause
depth remains four, repeated identities remain denied, and every emitted Error/AggregateError has
bounded own data descriptors for its normalized fields and only normalized nested snapshots.

This explicitly supersedes Cycle 11's former compatibility assertion that a hostile aggregate
iterator is pulled up to 16 times and closed once. The retained safety intent is strengthened:
the same row continues to require total bounded typed failures and secret-marker absence, but now
permits only zero iterator/close callbacks. Primary, aggregate, cleanup, dual-failure, quarantine,
and close paths all exercise the descriptor rule after an awaited boundary.

### C20-04 — bounded lazy plain-record keys and pre-work rejection

Closed object projection keeps exact Pi `additionalProperties: false` semantics. After the retained
non-proxy and exact `Object.prototype`/null-prototype checks, one audited helper uses a bounded
`for...in` cursor without retaining a key vector. It charges every enumerated visit before captured
has-own and schema-membership checks, rejects immediately when the remaining shared key budget is
exhausted, rejects an enumerable own unknown key immediately, and then requires the exact compiled
required-field count. Inherited enumerable pollution is charged and bounded but never treated as
input. Symbols and non-enumerable peers remain ignored, matching pinned Pi and the previous
`Object.keys` semantics. Known fields are subsequently captured only by their compiled names and
own data descriptors.

Standard JavaScript exposes no lazy own-string-key reflection primitive: `Object.keys`,
`Object.getOwnPropertyNames`, and `Reflect.ownKeys` all allocate the complete vector. Therefore this
single bounded enumerator explicitly narrows Cycle 17's blanket source-level `for...in` ban. The
retained Cycle 17 test will allow exactly the named helper and continue to reject every other
`for...in` occurrence, proxy prototype, accessor, or inherited-authority path. Stripping unknown
keys is rejected because it would diverge from pinned Pi's closed-object validation.

The proof is intentionally observable rather than an unverifiable engine-allocation claim:
no whole-vector API is called, no key vector is retained, and descriptor/value reads stop at the
first unknown key or remaining-budget-plus-one visit. ECMAScript permits an engine to snapshot
enumeration state internally. If independent review requires proof about that hidden allocation,
ordinary extensible records cannot satisfy it without narrowing the public record contract; GREEN
must then stop for contract direction rather than claim an impossible guarantee.

Key visits, scalar source work, grapheme work, numeric normalization, punctuation, and output bytes
all charge before the corresponding descriptor descent, trim/conversion, allocation, or output
insertion. A 5,001-key object must reject by visit 4,097 without an `Object.keys`/ownKeys vector;
a 100,000-unit scalar under a tiny remaining budget must reject before whole-scalar stringify,
encode, trim, conversion, or scan; repeated DAG aliases still reject before second descent; and
near-limit closed records and escaped/astral scalars remain exact frozen DTOs.

### Comprehensive four-row RED matrix

After this artifact-only checkpoint, one test-only commit adds exactly four top-level Cycle 20
rows and narrowly updates the two superseded retained assertions described above:

| ID | Exact executable boundary |
|---|---|
| C20-01 | `cycle 20 every mutable normalization text byte error and reflection API is captured once` — post-import and across-await poison matrix with zero ambient calls and invariant DTO/error/byte outcomes |
| C20-02 | `cycle 20 pinned Pi graphemes and incremental JSON UTF8 bytes are exact` — real Pi/direct/callback/event/lifecycle combining, variation, multi-ZWJ, flag, astral, malformed-surrogate matrix plus escape/UTF-8 near bounds and zero whole-scalar observers |
| C20-03 | `cycle 20 runtime errors use own data descriptors and exact dense aggregate arrays` — cause/errors/message/name/stack accessors, hostile iterators and post-await primary/cleanup/quarantine graphs with zero caller behavior |
| C20-04 | `cycle 20 projection rejects keys scalar work and DAGs before full discovery` — 5,001-key, 100,000-unit string/numeric, tiny-byte, repeated-DAG, inherited-pollution and near-limit controls with bounded visit observers |

Expected RED is 168 executed: all 164 retained Cycle 19 rows pass under their current behavior,
including the unchanged handoff-array and category-oracle rows, and exactly C20-01 through C20-04
fail their intended assertions with zero skipped/cancelled/todo. The Cycle 11 iterator assertion is
narrowly changed to accept zero callbacks while C20-03 requires zero; the Cycle 17 source assertion
is narrowed to allow at most the named bounded enumerator while C20-04 requires it. Those changes
pass at frozen production and do not hide a RED failure. Focused strict TypeScript must remain
green.

Frozen blobs through RED are runtime `e678193e5ff6022b0ac40628dfd9fc07a928bb2a`, policy
`499f5a0b6cd0730c1279b1d33728604cb06c09c9`, and prompts
`c5b6c27fc1ba6f738fbfd36d49d38c94c7b13b73`; pre-RED test blobs are runtime test
`d25a8e1b25df3b80058e54f5ad3e7072d1362ad0` and policy test
`c744c7d3d1ea798a30a214d94c9f44165cf25d84`.

Baseline at the exact start is focused 164/164. `scripts/gsd doctor` passes 69 commands while the
adapter rejects `programming-loop`; explicit manual-GSD PLAN -> RED -> GREEN -> REFACTOR -> VERIFY
remains active. Skills used are `gsd-programming-loop`, `architecture-patterns`,
`javascript-testing-patterns`, `typescript-advanced-types`, and
`github-issue-first-delivery`, plus repository routing, runtime/Pi, issue, adapter, and universal-
loop guidance. One bounded read-only Cycle 20 source mapper was spawned with no edit, commit,
network, or gate authority; the issue worker retains every repository write and the complete
checkpoint path.

No production edit precedes committed RED. No dependency, prompt/lease/route behavior, handoff
array/category-oracle behavior, parent, #478, network, push, integration, live model/auth,
credential, service, Go/connector, `make`, main, or path-scope change is authorized.
