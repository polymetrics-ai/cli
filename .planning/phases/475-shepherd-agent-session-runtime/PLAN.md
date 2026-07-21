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
