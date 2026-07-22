# TDD Ledger — Issue #475

## Policy

- Mode: `manual_gsd_fallback` because `scripts/gsd prompt programming-loop ...` is absent from the
  healthy 69-command adapter registry.
- Production code is blocked until the RED checkpoint below is captured.
- Tests use a fake injected Pi SDK/session; they must not require model auth, network, secrets, a Pi
  subprocess, tmux, Git/GitHub mutation, or a real workspace.

## Cycle 1 — AgentSession Authority And Lifecycle

### RED

- Status: captured
- Test files: `agent-session-runtime.test.ts`, `tool-policy.test.ts`, optional
  `role-prompts.test.ts`
- Expected coverage: exact route, bounded context, least authority, recursion prevention,
  cancellation/deadline/close/shutdown races, join once, quarantine, and bound/redacted handoff.
- Command:

  ```bash
  node --test .pi/extensions/shepherd/agent-session-runtime.test.ts \
    .pi/extensions/shepherd/tool-policy.test.ts
  ```

- Observed failure: exit 1, 0 passed / 2 file-level failures. Node reported
  `ERR_MODULE_NOT_FOUND` for `agent-session-runtime.ts` and `tool-policy.ts`. This is the expected
  pre-production RED state: the fake-SDK authority/lifecycle test contracts exist and the owned
  production adapters do not.

### GREEN

- Status: captured.
- Minimal implementation: exact role router and trusted prompt envelope; opaque workspace and
  typed-capability tool policy; injected Pi 0.80.6 `createAgentSession` lifecycle owner; strict
  bounded/redacted handoff parser.
- Command:

  ```bash
  node --test .pi/extensions/shepherd/agent-session-runtime.test.ts \
    .pi/extensions/shepherd/tool-policy.test.ts
  ```

- Observed pass: exit 0, 19 passed / 0 failed. Coverage includes 5.5/fallback/tool/terminal drift,
  injection and read-only authority, path and secret boundaries, abort/timeout/deadline/close/
  parent-shutdown races, late creation, join-once, cleanup quarantine, schema/binding/bounds, and
  mutator concurrency.

### REFACTOR

- Status: captured.
- Refactor notes: joined resource-loader setup under cancellation; bounded and quarantined hung
  setup; rejected valid-looking evidence when close/shutdown wins during child settlement; bounded
  typed capability schemas; verified secret redaction in the actual user prompt.
- Focused result after refactor: 22 passed / 0 failed.
- Complete Shepherd result: 159 passed / 0 failed.
- Strict no-emit TypeScript: owned production/tests and all Shepherd production files passed with
  `--strict` against the explicitly pinned Pi 0.80.6 installation/types.
- Offline smoke, diff, immutable-base, and owned-scope checks passed.

## Gate History

| Checkpoint | Command | Result | Evidence |
|---|---|---|---|
| Adapter | `scripts/gsd doctor` | pass | Pi adapter and registry healthy |
| GSD command | `scripts/gsd prompt programming-loop init --phase 475-shepherd-agent-session-runtime --dry-run` | unavailable | `unknown GSD command: programming-loop`; manual fallback activated |
| RED | `node --test .pi/extensions/shepherd/agent-session-runtime.test.ts .pi/extensions/shepherd/tool-policy.test.ts` | expected fail | 0 passed; missing owned production modules |
| GREEN | `node --test .pi/extensions/shepherd/agent-session-runtime.test.ts .pi/extensions/shepherd/tool-policy.test.ts` | pass | 19 passed; 0 failed |
| REFACTOR focused | same focused command | pass | 22 passed; 0 failed |
| REFACTOR full | `node --test .pi/extensions/shepherd/*.test.ts` | pass | 159 passed; 0 failed |
| TypeScript | pinned TypeScript 5.9.3 `tsc --noEmit --strict ...` against explicit Pi 0.80.6 base/type roots | pass | owned tests/production and all Shepherd production files |
| Offline smoke | explicit Pi 0.80.6 RPC `get_commands` with `PI_OFFLINE=1` | pass | `pm-shepherd` extension command registered |
| Scope/diff | `git diff --check` plus immutable-base owned-path assertion | pass | only issue #475 files changed |

## Cycle 2 — Exact-Head Review Corrections

### PLAN

- Status: captured against `4e41c2ec1175a109c10f125203dc54d381b982bd`.
- Trigger: PR #486 independent review reported two P1 blockers: abandoned late session creation and
  quoted secret forms escaping redaction.
- Orchestration decision: `local_critical_path`; the correction overlaps the same two owned source
  and test modules, and all available agent slots are already occupied.
- Skills/policy retained: `gsd-programming-loop` via recorded manual fallback,
  `javascript-testing-patterns`, `typescript-advanced-types`, `architecture-patterns`,
  `github-issue-first-delivery`, required skills routing, issue-agent contract, universal runtime
  loop, Pi adapter, and runtime/RLM/Pi integration guidance.

### RED

- Status: captured; production source remained unchanged.
- Lifecycle contract: creation resolves only after request deadline plus cleanup bound; the run
  settles without waiting forever, the late session is never prompted, and abort/wait/dispose are
  each eventually called exactly once.
- Redaction contract: synthetic quoted JSON/YAML assignments and quoted Bearer values do not leak
  through direct redaction, prompt serialization, tool output, or handoff summary/finding fields;
  ordinary prose remains unchanged.
- Command:

  ```bash
  node --test .pi/extensions/shepherd/agent-session-runtime.test.ts \
    .pi/extensions/shepherd/tool-policy.test.ts
  ```

- Observed result: exit 1, 19 passed / 5 failed. Expected failures:
  - prompt injection quoted-secret boundary leaked the task/context marker;
  - handoff quoted-secret boundary leaked summary/finding markers;
  - post-deadline plus post-cleanup-bound session creation never reached dispose;
  - direct quoted JSON/YAML/Bearer redaction probe leaked its marker;
  - typed capability tool output leaked its quoted Bearer marker.

### GREEN

- Status: captured after RED commit `93b9eca9`.
- Minimal implementation:
  - added a claimed/abandoned creation owner whose attached continuation cleans a session that
    resolves after the bounded request teardown has stopped waiting;
  - preserved bounded run completion while coalescing late abort/wait/dispose exactly once and
    quarantining any eventual cleanup failure;
  - extended line-bounded redaction for quoted JSON/YAML secret assignments and quoted/unquoted
    Authorization Bearer values while preserving their syntax delimiters.
- Focused command result: exit 0, 24 passed / 0 failed.

### REFACTOR / VERIFY

- Status: captured.
- Declared gates remain only issue-focused tests, the complete Shepherd TypeScript suite, strict
  no-emit TypeScript against pinned Pi 0.80.6, offline Pi RPC, diff check, and changed-path scope.
- Local adversarial review found no further actionable lifecycle, redaction, or strict-TypeScript
  issue. A nested reviewer was unavailable because the four-slot runtime remained saturated, so
  the recorded execution decision remained `local_critical_path`.
- Results:
  - focused tests: 24 passed / 0 failed;
  - complete Shepherd tests: 161 passed / 0 failed;
  - strict owned production/tests plus role prompt inputs: pass;
  - strict all-Shepherd production inputs: pass;
  - explicit Pi 0.80.6 offline RPC `get_commands`: pass, `pm-shepherd` registered;
  - immutable-base diff check and issue-owned changed-path assertion: pass.

## Cycle 2 Gate History

| Checkpoint | Result | Evidence |
|---|---|---|
| PLAN | pass | exact reviewed head `4e41c2ec1175a109c10f125203dc54d381b982bd`; correction scope recorded before tests |
| RED | expected fail | exit 1; 19 passed / 5 expected failures; production unchanged |
| GREEN | pass | exit 0; 24 passed / 0 failed at implementation head `f788cf16` |
| Complete Shepherd | pass | 161 passed / 0 failed |
| Strict TypeScript | pass | owned source/tests and all Shepherd production source against explicit Pi 0.80.6 types |
| Offline RPC | pass | explicit Pi 0.80.6 binary returned successful `get_commands` with `pm-shepherd` |
| Diff / scope | pass | immutable base retained; all changed paths remain issue #475-owned |

## Cycle 6 — Multiline Value Ownership, Quote Boundaries, And Linear Work

### PLAN

- Status: captured against reviewed head
  `d918617a19749cd16d6bfcf3d2fee3e5146e7380`; production remains locked until the pushed test-only
  RED checkpoint.
- GSD: `scripts/gsd doctor` passes; the 69-command registry rejects `programming-loop`, so the
  existing `manual_gsd_fallback` remains active.
- Skills/policy reloaded: `gsd-programming-loop`, `javascript-testing-patterns`,
  `typescript-advanced-types`, `architecture-patterns`, `github-issue-first-delivery`, required
  routing, issue-agent contract, universal runtime loop, Pi adapter, and runtime/Pi reference.
- Orchestration: `local_critical_path`; a read-only architecture sidecar was attempted but rejected
  by the four-thread runtime cap, while all findings overlap this worker's owned scanner/consumer
  scope.
- Multiline RED uses a nested mapping value that crosses a newline and closes before a later
  same-line outer `client_secret`. Direct, serialized-prompt, `workspace_read`, typed capability,
  and handoff summary/finding boundaries must remove its marker.
- Apostrophe RED places `rock-'n-roll` before a later sensitive sibling at the same five consumer
  boundaries. A harmless flow value containing only that scalar must remain byte-identical.
- Complexity RED builds approximately 25/50/100 KiB single-line flow maps with many assignments.
  An optional typed scan-metrics sink must report nonzero line-boundary byte visits bounded by a
  constant per input byte and near-doubling between sizes; no wall-clock assertion is allowed.
- Implementation contract: allow the exact value-local closer stack to span lines, make quote
  opening token-context-aware, pass the scanner's current line end into assignment decisions, and
  advance all subsequent line discovery monotonically. Preserve the typed lexer and all 36 prior
  focused cases.
- Expected focused command:

  ```bash
  node --test .pi/extensions/shepherd/agent-session-runtime.test.ts \
    .pi/extensions/shepherd/tool-policy.test.ts
  ```

### RED

- Status: captured with tests and phase evidence only; production remains unchanged at
  `d918617a19749cd16d6bfcf3d2fee3e5146e7380`.
- Focused command result: exit 1, 33 passed / 7 expected failures:
  - serialized prompt retained a multiline-nested or punctuation-apostrophe marker;
  - handoff summary/finding retained a multiline-nested or punctuation-apostrophe marker;
  - direct multiline nested-value ownership hid the later sensitive sibling;
  - direct `rock-'n-roll` quote state hid the later sensitive sibling;
  - deterministic 25/50/100 KiB scale metrics were absent at the reviewed head;
  - `workspace_read` retained a multiline-nested or punctuation-apostrophe marker;
  - typed capability output retained both new marker classes.
- Safe apostrophe byte-identity and every prior focused case remain green.
- Focused strict TypeScript result: exit 0 against the explicit Pi 0.80.6 package/type roots,
  proving the regression harness compiles cleanly.

### GREEN / REFACTOR / VERIFY

- GREEN status: captured after pushed test-only RED commit `e8422d53`.
- Minimal multiline implementation: the value-local closer stack advances across CR/LF line
  endings while nested, resets value quote context per line, and returns only at the exact outer
  sibling boundary. The global flow stack therefore never consumes a nested value close.
- Minimal quote implementation: `-` opens quote state only when it is a line-local YAML sequence
  marker; the hyphen inside `rock-'n-roll` remains unquoted scalar text.
- Minimal complexity implementation: assignment decisions receive the scanner-owned line end.
  Optional typed metrics count line-boundary byte visits; 25,618 / 51,218 / 102,418-byte dense
  flows report exactly 25,618 / 51,218 / 102,418 visits.
- Compatibility refactor: overloads preserve the one-argument `Array.map(redactSensitiveText)`
  callback contract and ignore its numeric index instead of treating it as diagnostics.
- Focused command result: exit 0, 40 passed / 0 failed.
- Focused strict TypeScript result: exit 0 against explicit Pi 0.80.6 package/type roots.
- REFACTOR/VERIFY status: complete at implementation head
  `93314a54302e84e053ad0d6ff44371fbf1a167e0`.
- Results:
  - focused tests: 40 passed / 0 failed;
  - complete Shepherd tests: 177 passed / 0 failed;
  - strict owned production/tests plus role prompt inputs: pass;
  - strict all-Shepherd production inputs: pass;
  - explicit Pi 0.80.6 offline RPC `get_commands`: pass, `pm-shepherd` registered;
  - immutable-base diff check and exact issue-owned changed-path assertion: pass.
- Declared gates: focused issue tests, complete Shepherd suite, pinned Pi 0.80.6 strict TypeScript,
  pinned offline RPC, diff check, immutable base, and issue-owned paths only.
- Go, connector, certification, runtime-backed, `make verify`, live-GitHub, merge, and review-bot
  commands remain forbidden.

## Cycle 6 Gate History

| Checkpoint | Result | Evidence |
|---|---|---|
| PLAN | pass | exact reviewed head `d918617a19749cd16d6bfcf3d2fee3e5146e7380`; scope pushed at `4f9c5a96` |
| RED | expected fail | exit 1; 33 passed / 7 expected failures; production unchanged; pushed at `e8422d53` |
| GREEN / REFACTOR | pass | implementation `93314a54302e84e053ad0d6ff44371fbf1a167e0`; 40 passed / 0 failed |
| Deterministic scale | pass | 25,618 / 51,218 / 102,418 line-boundary visits for equal-sized inputs |
| Focused strict TypeScript | pass | owned source/tests plus role prompt inputs against explicit Pi 0.80.6 types |
| Complete Shepherd | pass | 177 passed / 0 failed |
| Strict TypeScript | pass | focused owned inputs and all Shepherd production source against explicit Pi 0.80.6 types |
| Offline RPC | pass | explicit Pi 0.80.6 binary returned successful `get_commands` with `pm-shepherd` |
| Diff / scope | pass | immutable base retained; exact changed-path assertion remains issue #475-owned |

## Cycle 5 — Reservation Timer Ownership And Lexical State Machine

### PLAN

- Status: captured against reviewed head
  `e41f075a9b3bfb01d410296712740b54f943ba71`; production remains locked until the pushed test-only
  RED checkpoint.
- GSD: `scripts/gsd doctor` passes; the 69-command registry still rejects `programming-loop`, so
  the existing `manual_gsd_fallback` remains active.
- Skills/policy reloaded: `gsd-programming-loop`, `javascript-testing-patterns` plus advanced timer
  guidance, `typescript-advanced-types`, `architecture-patterns`,
  `github-issue-first-delivery`, required routing, issue-agent contract, universal runtime loop,
  Pi adapter, and runtime/Pi integration reference.
- Orchestration: `local_critical_path`; a read-only architecture sidecar was attempted but rejected
  by the four-thread runtime cap, while both source findings overlap this worker's owned modules.
- Lifecycle RED instruments an immediate duplicate long-timeout rejection and requires every scope
  timer created on that rejected path to be cleared or unref'ed. The preferred implementation
  creates no scope until reservation succeeds.
- Redaction RED adds nested sensitive mapping plus later unquoted sibling and leading unmatched
  apostrophe cases at direct, serialized-prompt, typed-tool-output, and handoff summary/finding
  boundaries. Ordinary unmatched-brace and flow-shaped-comment controls must remain byte-identical.
- Refactor contract: replace the accumulated traversal exceptions with one explicit deterministic
  line/flow lexical state machine using monotonic cursors, newline YAML-quote reset, balanced nested
  delimiter consumption, and comment/prose discrimination. Preserve every earlier structured,
  multiline, block, Bearer, flow, spaced-scalar, and harmless-prose test.
- Expected focused command:

  ```bash
  node --test .pi/extensions/shepherd/agent-session-runtime.test.ts \
    .pi/extensions/shepherd/tool-policy.test.ts
  ```

### RED

- Status: captured with tests and phase evidence only; production remains unchanged at
  `e41f075a9b3bfb01d410296712740b54f943ba71`.
- Focused command result: exit 1, 29 passed / 7 expected failures:
  - serialized prompt retained a nested-flow or post-apostrophe marker;
  - handoff summary/finding retained a nested-flow or post-apostrophe marker;
  - immediate duplicate long-timeout rejection left one referenced, uncleared scope timer;
  - direct nested mapping consumption hid the later sensitive sibling;
  - direct unmatched leading apostrophe hid the next structured line;
  - ordinary unmatched brace and flow-shaped comment controls were modified;
  - typed tool output retained both nested-flow and post-apostrophe markers.
- Focused strict TypeScript result: exit 0 against the explicit Pi 0.80.6 package/type roots,
  proving the timer instrumentation and regression harness compile cleanly.

### GREEN / REFACTOR / VERIFY

- GREEN status: captured after pushed test-only RED commit `333c7ad6`.
- Minimal lifecycle implementation: `#reserve()` performs duplicate, capacity, and mutator
  admission checks before constructing the `CancellationScope`; rejected admission therefore owns
  no deadline timer to clear.
- Minimal redaction implementation: replaced the accumulated depth/quote traversal with one typed
  lexical state machine. A monotonic cursor owns line boundaries, per-line quote state, comment
  skips, validated flow openers, and the outer delimiter stack; unquoted value consumption uses its
  own balanced nested delimiter stack and returns at the true outer sibling boundary.
- Focused command result: exit 0, 36 passed / 0 failed.
- Focused strict TypeScript result: exit 0 against explicit Pi 0.80.6 package/type roots.
- REFACTOR/VERIFY status: complete at implementation head
  `8ff2d9631809d09db26811b4cd1335b92a9c457c`.
- Refactor notes: scanner state is represented by closed quote/mode/closer types; assignment-key
  recognition uses bounded character classification; every search cursor advances monotonically;
  balanced nested values are consumed once and the outer flow stack remains authoritative.
- Results:
  - focused tests: 36 passed / 0 failed;
  - complete Shepherd tests: 173 passed / 0 failed;
  - strict owned production/tests plus role prompt inputs: pass;
  - strict all-Shepherd production inputs: pass;
  - explicit Pi 0.80.6 offline RPC `get_commands`: pass, `pm-shepherd` registered;
  - immutable-base diff check and issue-owned changed-path assertion: pass.
- Declared gates: focused issue tests, complete Shepherd suite, pinned Pi 0.80.6 strict TypeScript,
  pinned offline RPC, diff check, immutable base, and issue-owned paths only.
- Go, connector, certification, runtime-backed, `make verify`, live-GitHub, merge, and review-bot
  commands remain forbidden.

## Cycle 5 Gate History

| Checkpoint | Result | Evidence |
|---|---|---|
| PLAN | pass | exact reviewed head `e41f075a9b3bfb01d410296712740b54f943ba71`; scope pushed at `8087b539` |
| RED | expected fail | exit 1; 29 passed / 7 expected failures; production unchanged; pushed at `333c7ad6` |
| GREEN / REFACTOR | pass | implementation `8ff2d9631809d09db26811b4cd1335b92a9c457c`; 36 passed / 0 failed |
| Focused strict TypeScript | pass | owned source/tests plus role prompt inputs against explicit Pi 0.80.6 types |
| Complete Shepherd | pass | 173 passed / 0 failed |
| Strict TypeScript | pass | focused owned inputs and all Shepherd production source against explicit Pi 0.80.6 types |
| Offline RPC | pass | explicit Pi 0.80.6 binary returned successful `get_commands` with `pm-shepherd` |
| Diff / scope | pass | immutable base retained; all changed paths remain issue #475-owned |

## Cycle 4 — Foreground Forced Disposal And Unquoted YAML Context

### PLAN

- Status: captured against exact reviewed head
  `b4061d4e1a1545b0c8810b14b510cf048385a567`.
- Production remains unchanged until the focused test-only RED checkpoint is committed and pushed.
- Lifecycle RED matrix:
  - creation settles within the cleanup grace; `abort()` never settles;
  - creation settles within the cleanup grace; `waitForIdle()` never settles;
  - session is claimed and prompted before cancellation; `abort()` never settles;
  - session is claimed and prompted before cancellation; `waitForIdle()` never settles.
- Every lifecycle row requires bounded settlement, exactly-one forced disposal, quarantine and
  rejected subsequent dispatch without another prompt, plus zero unhandled rejections. Idle wait
  is optional only for the hung-abort rows.
- Redaction RED spans direct calls, prompt serialization, typed tool output, and handoff
  summary/finding fields for flow-map unquoted assignments and line-start spaced
  `client_secret` scalars. Non-assignment prose remains byte-identical.
- Expected focused command:

  ```bash
  node --test .pi/extensions/shepherd/agent-session-runtime.test.ts \
    .pi/extensions/shepherd/tool-policy.test.ts
  ```

### RED

- Status: captured; tests and phase evidence only, with production source still locked at reviewed
  head `b4061d4e1a1545b0c8810b14b510cf048385a567`.
- Focused result: exit 1, 23 passed / 8 expected failures:
  - prompt serialization leaked a flow-map unquoted `client_secret` marker;
  - handoff summary/finding redaction leaked a spaced/flow unquoted marker;
  - cleanup-grace plus hung abort reached quarantine but disposed 0 times;
  - cleanup-grace plus hung idle reached quarantine but disposed 0 times;
  - claimed-before-cancel plus hung abort reached quarantine but disposed 0 times;
  - claimed-before-cancel plus hung idle reached quarantine but disposed 0 times;
  - the direct structured matrix leaked its flow-map marker;
  - typed tool output leaked its flow-map marker.
- Both spaced-scalar and flow-map gaps are represented independently: the handoff summary reaches
  the line-start spaced scalar before its finding assertion, while direct/tool/prompt failures reach
  flow-map inputs. Harmless controls remained green.

### GREEN / REFACTOR / VERIFY

- GREEN status: captured after test-only RED commit `21535513`.
- Minimal lifecycle implementation:
  - separated coalesced idle waiting from coalesced disposal;
  - foreground cleanup independently bounds abort and idle phases, preserves the first failure,
    and always reaches exactly-once disposal before quarantine/return;
  - removed the duplicate cleanup-grace abort step so the claimed session follows the same pipeline.
- Minimal redaction implementation:
  - tracks bounded flow-map lexical context and discovers unquoted keys after `{` and `,` without
    rescanning line prefixes;
  - treats flow-map scalars and normalized `client_secret` assignments as strong structured values,
    while non-assignment and ambiguous non-client prose remain unchanged;
  - resumes scanning preserved prose so nested flow mappings cannot hide a later credential.
- Local adversarial REFACTOR added two test-first probes after initial GREEN:
  - an apostrophe in ordinary prose previously opened an unmatched quote state and hid the next
    assignment (targeted 0/1 RED, then fixed with structured quote-opening boundaries);
  - a preserved ambiguous `token:` line previously skipped a nested flow map (targeted 0/1 RED,
    then fixed by resuming within unredacted prose).
- Focused result after refactor: exit 0, 31 passed / 0 failed.
- Strict focused TypeScript: exit 0 against explicit Pi 0.80.6 package/type roots.
- REFACTOR/VERIFY status: complete at implementation head
  `01b42ae168176956d864ff10f40d1c981f37ac04`.
- Results:
  - focused tests: 31 passed / 0 failed;
  - complete Shepherd tests: 168 passed / 0 failed;
  - strict owned production/tests plus role prompt inputs: pass;
  - strict all-Shepherd production inputs: pass;
  - explicit Pi 0.80.6 offline RPC `get_commands`: pass, `pm-shepherd` registered;
  - immutable-base diff check and issue-owned changed-path assertion: pass.
- Declared gates remain focused issue tests, the complete Shepherd suite, strict TypeScript against
  explicit Pi 0.80.6 types, pinned offline RPC, diff check, immutable-base, and owned-scope checks.
- Go, connector, certification, runtime-backed, `make verify`, live-GitHub, merge, and review-bot
  commands remain forbidden in this lane.

## Cycle 4 Gate History

| Checkpoint | Result | Evidence |
|---|---|---|
| PLAN | pass | exact reviewed head `b4061d4e1a1545b0c8810b14b510cf048385a567`; correction scope pushed at `190b0ec7` |
| RED | expected fail | exit 1; 23 passed / 8 expected failures; production unchanged; pushed at `21535513` |
| Adversarial mini-REDs | expected fail | apostrophe boundary and nested-flow hiding each failed 0/1 before its production support |
| GREEN / REFACTOR | pass | implementation `01b42ae168176956d864ff10f40d1c981f37ac04`; 31 passed / 0 failed |
| Focused strict TypeScript | pass | owned source/tests plus role prompt inputs against explicit Pi 0.80.6 types |
| Complete Shepherd | pass | 168 passed / 0 failed |
| Strict TypeScript | pass | focused owned inputs and all Shepherd production source against explicit Pi 0.80.6 types |
| Offline RPC | pass | explicit Pi 0.80.6 binary returned successful `get_commands` with `pm-shepherd` |
| Diff / scope | pass | immutable base retained; all changed paths remain issue #475-owned |

## Cycle 3 — Multiline Redaction And Bounded Abandoned Cleanup

### PLAN

- Status: captured against reviewed head `526dfec4282b442c4b32138ab036d4cc7e97b475`.
- GSD: `scripts/gsd doctor` passes; the 69-command registry still rejects `programming-loop`, so
  the already-recorded `manual_gsd_fallback` continues without weakening TDD.
- Skills/policy: `gsd-programming-loop`, `javascript-testing-patterns` (including async/timer
  reference), `typescript-advanced-types`, `architecture-patterns`,
  `github-issue-first-delivery`, required-skills routing, issue-agent contract, universal runtime
  loop, Pi adapter, and runtime/RLM/Pi integration guidance.
- Orchestration: `local_critical_path`; a read-only design sidecar was attempted but rejected by the
  runtime thread cap, while both production findings share the same issue-owned modules.

### RED

- Status: captured; production source remained unchanged.
- Redaction tests will prove multiline quoted scalars, YAML block scalars, `client_secret`, and
  multiline quoted Bearer credentials at direct, prompt, tool-output, handoff-summary, and
  handoff-finding boundaries, plus byte-identical ambiguous prose controls.
- Lifecycle tests will separately hang late-session `abort()` and `waitForIdle()` after creation
  abandonment, then require bounded single disposal, quarantine rejection on subsequent dispatch,
  zero prompt calls, no duplicate cleanup, and no unhandled rejection.
- Command:

  ```bash
  node --test .pi/extensions/shepherd/agent-session-runtime.test.ts \
    .pi/extensions/shepherd/tool-policy.test.ts
  ```

- Observed result: exit 1, 20 passed / 7 expected failures:
  - prompt multiline secret boundary leaked a marker;
  - handoff summary/finding multiline boundary leaked a marker;
  - abandoned cleanup with never-settling `abort()` never disposed;
  - abandoned cleanup with never-settling `waitForIdle()` never disposed;
  - direct multiline structured secret matrix leaked its quoted YAML marker;
  - harmless assignment-like prose was modified;
  - typed tool-output YAML block scalar leaked its marker.

### GREEN

- Status: captured after the test-only RED commit `9c4ed5fd`.
- Minimal implementation:
  - replaced line-local secret regexes with a bounded structured scanner that recognizes normalized
    credential keys, multiline quoted values, YAML block scalars, and quoted/unquoted Bearer
    credentials while leaving ambiguous multiword assignment prose unchanged;
  - bounded abandoned-session abort and join against one cleanup deadline, unref'ed only the
    detached cleanup timers, coalesced forced dispose, consumed detached rejections, and
    quarantined on either timeout or hook/dispose failure.
- Focused command result: exit 0, 27 passed / 0 failed.
- Strict focused TypeScript result: exit 0 against the explicit Pi 0.80.6 package/type roots.

### REFACTOR / VERIFY

- Status: captured at implementation head `d499e721a85abbe1a1d1be7fb0069649927c923c`.
- Refactor notes: structured-key discovery now advances linearly without repeatedly slicing line
  prefixes; private-key header recognition is bounded and unmatched recognized blocks fail closed;
  detached cleanup timers are explicitly unref'ed while foreground cleanup remains awaited.
- Results:
  - focused tests: 27 passed / 0 failed;
  - complete Shepherd tests: 164 passed / 0 failed;
  - strict owned production/tests plus role prompt inputs: pass;
  - strict all-Shepherd production inputs: pass;
  - explicit Pi 0.80.6 offline RPC `get_commands`: pass, `pm-shepherd` registered;
  - immutable-base diff check and issue-owned changed-path assertion: pass.
- No Go, connector, certification, runtime-backed, `make verify`, live-GitHub, merge, or review-bot
  command is permitted in this lane.

## Cycle 3 Gate History

| Checkpoint | Result | Evidence |
|---|---|---|
| PLAN | pass | exact reviewed head `526dfec4282b442c4b32138ab036d4cc7e97b475`; correction scope recorded before tests |
| RED | expected fail | exit 1; 20 passed / 7 expected failures; production unchanged |
| GREEN | pass | exit 0; 27 passed / 0 failed |
| Focused strict TypeScript | pass | owned source/tests plus role prompt inputs against explicit Pi 0.80.6 types |
| Complete Shepherd | pass | 164 passed / 0 failed |
| Strict TypeScript | pass | focused owned inputs and all Shepherd production source against explicit Pi 0.80.6 types |
| Offline RPC | pass | explicit Pi 0.80.6 binary returned successful `get_commands` with `pm-shepherd` |
| Diff / scope | pass | immutable base retained; all changed paths remain issue #475-owned |

## Cycle 7 — Stable-Head Lifecycle Ownership And Structured Redaction

### PLAN

- Status: captured against frozen candidate
  `a3cd85a5d0871dd1c4c99dd8b30bcd609a228c45` and immutable base
  `e659d6f1b666f58748e2d8c86599ceb4bbc62ff8`; production and tests remain unchanged.
- Trigger: the combined stable-head campaign comment
  <https://github.com/polymetrics-ai/cli/pull/486#issuecomment-5037079867> synthesizes 11 actionable
  findings (8 P1, 3 P2). Parent forensics/policy is read from `2a89142e` without merge or shared-
  artifact edits.
- GSD: `scripts/gsd doctor` passes; `scripts/gsd prompt programming-loop ...` remains absent from
  the healthy 69-command registry, so `manual_gsd_fallback` preserves the required TDD sequence.
- Skills/policy reloaded: `gsd-programming-loop`, `javascript-testing-patterns`,
  `typescript-advanced-types`, `architecture-patterns`, `github-issue-first-delivery`, required
  routing, issue contract, universal runtime loop, Pi adapter, and runtime/Pi guidance.
- Orchestration: `local_critical_path`; a read-only lifecycle sidecar was attempted but rejected by
  the four-thread runtime cap, while every finding overlaps the issue-owned runtime/redactor and
  their tests.

### Behavior RED

- Status: captured and pushed at `3b7e886a`. Exactly one test-only commit contains every row;
  all 53 tests execute and the command exits 1 with 40 retained passes / 13 intended assertion
  failures. Both production blobs are byte-identical to frozen `a3cd85a5`; no compile, module, or
  file-load failure contributes.
- Lifecycle rows (six separate tests):
  1. throwing external-signal attachment cannot strand the admitted reservation or deadline timer;
  2. throwing external-signal removal cannot skip scope finish, run release, or close settlement;
  3. close begun after abandonment waits for late valid creation and exactly-once cleanup;
  4. close begun after abandonment waits for late creation rejection and then succeeds cleanly;
  5. a forever-pending creation makes close reject/quarantine within a bound, never falsely pass;
  6. malformed late fulfillment is validated, consumed, and quarantined without
     `unhandledRejection`.
- Every lifecycle row records close outcome, referenced long timers, active reservation outcome,
  prompt/abort/wait/dispose counts as applicable, and captured unhandled rejections.
- Redaction rows (seven separate tests): one direct complete secret matrix; one byte-stable harmless
  multiline-quote control; one each for serialized prompt, `workspace_read`, typed capability, and
  handoff summary/finding; and one deterministic padded-flow work diagnostic.
- Secret forms are multiline outer flow, indented assignment, key-only and continued YAML plain
  scalar, numeric secret, Basic/non-Bearer Authorization, unmatched sensitive quote plus following
  sibling recovery, repository aliases, and generic PKCS#8. Each consumer aggregates every leaked
  marker before asserting, so the whole payload is exercised.
- Complexity inputs are approximately 25/50/100 KiB and combine proportional leading indentation
  with dense flow assignments. A typed optional diagnostics sink must report nonzero total scanner
  character work, including key-start discovery, within a constant per byte and near-doubling as
  size doubles; no timing threshold is allowed.

### GREEN / REFACTOR

- Status: captured after the committed/pushed behavior RED; focused tests pass 53/53 and focused
  strict TypeScript passes against explicit Pi 0.80.6 package/type roots.
- Runtime architecture: exception-safe listener lease; runtime-owned session-creation terminal
  records; bounded close join; quarantine on pending/malformed ownership; validation before late
  session construction; total rejection-consumed continuations.
- Redactor architecture: persistent structural flow/quote state; indentation-owned YAML
  continuation; scheme-independent Authorization sensitivity; numeric secret handling; normalized
  Shepherd alias predicate; generic PKCS#8 recognition; line-bounded unmatched-quote recovery;
  cached monotonic line/key metadata with complete deterministic work metrics.
- Preserve every prior lifecycle/redaction invariant and avoid new dependencies or widened
  authority.

### VERIFY

- Status: captured at implementation head `5c638d7f21a3910f40e499dba5c82cb7646642ac`.
- Focused tests: 53 passed / 0 failed; complete Shepherd: 190 passed / 0 failed.
- Focused and all-production strict TypeScript pass with TypeScript 5.9.3 against explicit Pi
  0.80.6 package/type roots.
- Explicit Pi 0.80.6 offline RPC `get_commands` passes with `pm-shepherd` registered.
- `git diff --check`, immutable-base, local/remote-tracking head equality, and issue-owned changed
  paths pass.
- Forbidden in this lane: Go, connector/certification, `make verify`, runtime-backed services,
  live-GitHub/CI/review-bot mutation, merge, and parent-artifact edits.

## Cycle 7 Gate History

| Checkpoint | Result | Evidence |
|---|---|---|
| PLAN | pass | frozen candidate/base and full 11-finding matrix pushed at `f40a08f1` before tests |
| RED | expected fail | pushed `3b7e886a`; 40 passed / 13 assertion failures; production identical |
| GREEN / REFACTOR | pass | 53 passed / 0 failed; one lifecycle/redactor architectural correction |
| Focused / full Shepherd | pass | focused 53/53; complete 190/190 |
| Strict TypeScript / offline RPC | pass | both strict scopes and explicit Pi 0.80.6 RPC registration |
| Diff / base / head / scope | pass | immutable base, pushed implementation head, and issue-owned paths only |

## Cycle 8 — Immutable Request Ownership, Total Cleanup, Hard Bounds, And Parser Closure

### PLAN

- Status: captured against frozen reviewed head
  `f219b730c63adc9188c93093a40511433a3d0110` and immutable base
  `e659d6f1b666f58748e2d8c86599ceb4bbc62ff8`; production and tests remain unchanged.
- Review sources: `/tmp/475-REVIEW-SECURITY-CYCLE7.md` plus the parent-provided final lifecycle
  disposition covering listener target mutation, explicit undefined failures, async cleanup hooks,
  repeated request accessors, and Node timer ceilings.
- GSD: `scripts/gsd doctor` passes; `scripts/gsd prompt programming-loop ...` is absent from the
  healthy 69-command registry, so the recorded `manual_gsd_fallback` remains active.
- Skills/policy reloaded: `gsd-programming-loop`, `javascript-testing-patterns`,
  `typescript-advanced-types`, `architecture-patterns`, `github-issue-first-delivery`, required
  routing, issue contract, universal runtime loop, Pi adapter, and runtime/Pi guidance.
- Orchestration: `read_only_spawned`; one read-only lifecycle sidecar maps test and implementation
  seams while this isolated issue worker owns the only mutating path.

### Behavior RED contract

- Exactly one test-only commit will cover all twelve deduplicated behavior families: listener lease
  attach/remove/parent/mutation, reasonless failures, awaited thenable cleanup, immutable normalized
  request and mutator fence, bounded disjoint mutator leases with collision-safe per-lease release,
  all hard ceilings, comma/Auth-param redaction, multiline-flow YAML continuation, escaped quoted
  keys, bounded cycle-safe event estimation, canonical normalized prefixes, and terminal-safe
  handoff strings.
- Every test must load and execute under strict pinned Pi 0.80.6 TypeScript. Only intended behavior
  assertions may fail; compile, module, fixture, and file-load failures are inadmissible.
- Production source hashes must remain identical to frozen `f219b730` through the RED commit.
- Focused command:

  ```bash
  node --test .pi/extensions/shepherd/agent-session-runtime.test.ts \
    .pi/extensions/shepherd/tool-policy.test.ts
  ```

### GREEN / REFACTOR contract

- Normalize/freeze the request and authority once, capture explicit listener leases, use explicit
  failure-presence state, await cleanup thenables, validate central hard maxima, extend the monotonic
  structured redactor with bounded key decoding, add a bounded cycle-safe event estimator, share one
  canonical prefix set, use a bounded canonical authority/scope lease map for mutator admission, and
  enforce safe handoff text.
- Preserve every prior 53 focused regression and add no dependency.
- Test-only RED is committed at `11aa221231a52fab91f41dfce9742b7dfe180c02`: all 70 tests
  executed, 53 retained tests passed, and exactly 17 Cycle 8 assertions failed; strict focused
  TypeScript passed and production remained byte-identical to `f219b730`.
- GREEN/refactor is committed at `c4d34c377532c903238400c986a6b488fab3646d`: immutable request
  ownership, listener/cleanup ownership, explicit failure presence, bounded disjoint mutation
  leases, hard maxima, bounded event accounting, canonical scopes, terminal-safe handoffs, and the
  parser closure make all 70 focused tests pass.

### VERIFY

- Pass: focused 70/70; focused and all-production strict TypeScript 5.9.3 against explicit Pi
  0.80.6 roots; explicit Pi 0.80.6 offline RPC registration; diff, ancestry, and owned-scope gates.
- Environment-blocked: the complete suite executes 207 tests but reports 176 passes / 31 failures
  because the managed sandbox rejects the existing `state-store.ts` `/bin/ps` child process with
  `spawn EPERM`. Per-file isolation limits the impact to controller/state-store tests.
- Environment-blocked: push cannot reach the remote because DNS returns
  `ssh: Could not resolve hostname github.com: -65563`.
- Disposition: keep the local ordered commits intact; parent reruns the complete suite in a process-
  capable environment, pushes, verifies remote-head equality, and requests fresh exact-head review.

## Cycle 9 — Closed SDK Ownership, Bounded DTOs, And Total Public Settlement

### PLAN

- Frozen candidate: `0cdcda7e049b7ecfa2fdc52027c66c5de161f2c8`; immutable base:
  `e659d6f1b666f58748e2d8c86599ceb4bbc62ff8`.
- Review sources: complete `/tmp/475-REVIEW-CYCLE8-1.md` and
  `/tmp/475-REVIEW-CYCLE8-2.md`, deduplicated into the fourteen invariants in `PLAN.md` and
  `REVIEW.md`.
- GSD: `scripts/gsd doctor` passes; the healthy registry still has no `programming-loop` command,
  so the manual-GSD fallback remains explicit.
- Skills: `gsd-programming-loop`, `javascript-testing-patterns`,
  `typescript-advanced-types`, `architecture-patterns`, `github-issue-first-delivery`, and the
  repository-required routing/runtime/issue/GSD references.
- Orchestration: `read_only_spawned`; one read-only explorer maps overlapping seams while this
  isolated lane owns all writes.
- Delegation trace: `AGENTS.md`, `agents/cycle9-seam-map.md`, and
  `traces/cycle9-seam-map-trace.md`; the explorer made no edits and returned the closed-field event
  parser and public-Pi-type/plain-schema constraints before RED.

### One behavior-level RED contract

One and only one test-only commit will add the consolidated matrix. It will cover:

- one-read foreground/late SDK result and same-session ownership;
- private frozen expected tools versus SDK array mutation/reorder/replacement;
- bounded accessor-free deep schemas and one-read workspace/capability/mutation results;
- fulfilled/rejected/pending reload/create settlement and retry without quarantine;
- bounded independent unsubscribe/dispose across normal and cancellation/late paths;
- captured request/parent signal operations plus fallback detach;
- typed own-cause errors and deterministic primary+cleanup aggregation at every public boundary;
- closed immutable terminal event DTOs with wide/proxy/sparse/accessor/mutation probes;
- the complete redaction, sensitive-path, capability-name, and terminal-control matrices;
- direct Pi 0.80.6 tool types, required `details`, and offline no-model argument/result behavior;
- retained Cycle 8 mutator alias/capacity/per-lease cleanup, timer, and rejection accounting.

RED acceptance is assertion-level failure only: every test must load and execute, strict focused
TypeScript must pass, and SHA-256 for `agent-session-runtime.ts`, `tool-policy.ts`, and
`role-prompts.ts` must equal the frozen candidate blobs. Production edits are forbidden before the
RED commit.

### GREEN / REFACTOR contract

Implement one cohesive correction around normalized ownership, discriminated settlement,
independent bounded cleanup, captured listener operations, bounded immutable data snapshots/event
DTOs, a private tool oracle, and directly typed Pi tools. Extend only the shared redaction/path/
capability classifiers and terminal-text validator needed by the tests. Do not weaken old tests,
add dependencies, change scheduler/parent files, or widen tools.

### VERIFY contract

- focused runtime/tool-policy tests;
- complete serialized Shepherd suite, with only the already-known `/bin/ps` sandbox denial recorded
  as environment-blocked when reproduced;
- focused and all-production strict TypeScript against explicit Pi 0.80.6 roots;
- pinned offline RPC registration and offline no-model custom-tool exercise;
- diff check, immutable-base/frozen-head ancestry, issue-owned paths, exact clean head;
- no Go/connectors/certification/`make`/runtime services/live GitHub/review bot/merge.

### GREEN / REFACTOR result

- Cohesive implementation commit: `94918f4e` (`fix(shepherd): harden agent session ownership`).
- The correction uses a one-read SDK creation claim, captured session operations, a private frozen
  active-tool oracle, explicit `fulfilled | rejected | pending` settlement, independent bounded
  teardown phases, captured signal operations with native fallback detach, typed own-cause public
  errors with deterministic aggregation, and immutable known-terminal DTOs.
- Tool policy now snapshots capability schemas deeply with node/key/depth/array/incremental-byte
  ceilings, captures results once, exposes actual Pi 0.80.6 tool/result types with required
  `details`, and closes the shared redaction/path/capability-name grammar.
- Focused GREEN: 86/86 passed (70 retained plus 16 Cycle 9 rows); strict focused TypeScript passed.

### VERIFY result

- Focused runtime/tool-policy suite: 86 passed, 0 failed.
- Strict no-emit TypeScript: focused production/tests and every Shepherd production `.ts` passed
  against the explicit Pi 0.80.6 package/type roots.
- Complete Shepherd serialization: 223 tests executed, 192 passed, and the same 31 parent-owned
  controller/state-store tests failed because the managed sandbox denies `/bin/ps` with
  `spawn EPERM`; focused issue-owned files remain fully green.
- Pinned offline Pi: explicit binary reports `0.80.6`; with `PI_CODING_AGENT_DIR` redirected to a
  temporary writable directory, RPC `get_commands` succeeded and registered `pm-shepherd`.
  The focused tool test separately invokes Pi's real `validateToolArguments` and executes a tool
  without a model or credentials.
- `git diff --check`, immutable-base/frozen-candidate ancestry, and exact issue-owned path checks
  pass. No dependency, Go/connector/certification, `make`, service, live GitHub, review-bot, merge,
  credential, or model call was made. Push remains parent/DNS-deferred.

## Cycle 10 — Consolidated Finding-To-RED Matrix

### PLAN

- Frozen candidate: `f63957aed6fd1406eb3bd9a82adbd10b23b34c33`; immutable base:
  `e659d6f1b666f58748e2d8c86599ceb4bbc62ff8`; worktree clean before planning.
- Review sources read completely: `/tmp/475-REVIEW-CYCLE9-1.md` and
  `/tmp/475-REVIEW-CYCLE9-2.md`; their union includes seven overlapping blockers, six additional
  behavior families, and WR-01.
- GSD: `scripts/gsd doctor` passed. The healthy adapter still lacks `programming-loop`, so the
  required manual-GSD PLAN -> RED -> GREEN/refactor -> verify fallback is explicit.
- Orchestration: `local_critical_path`. A read-only architecture explorer was attempted but rejected
  by the runtime thread cap; one isolated worker retains the only mutating scope.
- Production lock at RED: runtime `03cf916b59ef291dab309e6251a6f10ebf897eb0`, policy
  `1c8f701091a49c60cf41f83a6c16f2ae49a896c3`, role prompts
  `cfc2d253c323ad01f34b8c9688b3bad0acd16171`.

### Comprehensive test-only RED matrix

| ID | Finding source | Independent behavior row required to fail at frozen head |
|---|---|---|
| C10-01 | R1-1 / BL-03 | Genuine request and parent signals with own silent no-op add/remove hooks still cancel and finish with zero native abort listeners; captured throw/mutation rows remain. |
| C10-02 | R1-2 / BL-01 | Each missing/invalid/throwing session method or getter, after any cleanup method becomes available, is cleaned exactly once in foreground and abandoned fulfillment; successful forced cleanup permits retry, cleanup failure alone quarantines. |
| C10-03 | R1-3 | Hung late unsubscribe and hung late dispose each leave zero referenced timeout handles while close observes the eventual cleanup terminal. |
| C10-04 | BL-02 | Late abort/idle/unsubscribe/dispose may cumulatively exceed one phase bound while each phase succeeds; close, shutdown, and coalesced close return only after dispose without quarantine, while never-fulfilling creation remains bounded. |
| C10-05 | R1-4 | Foreground and late creation reject proxy, sparse, hidden/accessor/extra-field, symbol, and alternating-length extension containers captured once as exact closed empty arrays, while disposing the exact returned session. |
| C10-06 | R1-5 | A Pi-0.80.6-shaped cumulative `message_update` stream reaches a near-bound terminal handoff without quadratic charging; one-above delta/event/assistant/terminal bounds still reject. |
| C10-07 | R1-6 | Top-level and nested `__proto__`, `prototype`, and `constructor` schema/result keys remain own frozen data properties and serialize identically; inherited authority cannot satisfy required schema/result fields. |
| C10-08 | BL-07 | Wide plain schemas/events prove bounded early enumerable visits before rejection, and every known terminal event rejects an unknown field without first allocating the adversarial complete key set. |
| C10-09 | R1-7 | Unique secret markers in SDK setup/session/listener/cleanup and workspace/capability primary, dual, quarantine, and close failures never appear in public/model errors or causes; only bounded typed redacted snapshots cross the boundary. |
| C10-10 | BL-04 | Documentary prefixes after `=` redact, `Proxy-Authorization`, quoted YAML/flow keys, and OAuth fragments redact through direct, prompt, workspace, mutation/capability, and every handoff consumer; explicit harmless colon prose stays byte-identical. |
| C10-11 | BL-05 | AWS config, Azure token caches, GCloud legacy credentials, `.envrc`, and common private-key names reject before callback for root, nested, and case variants with opaque non-assignment content. |
| C10-12 | BL-06 | show/view/download/obtain/copy/reveal/lookup/print and sensitive noun compounds in both orders, aliases, and singular/plural are absent for read-only and mutating roles. |
| C10-13 | WR-01 | Every forbidden control combined with a redacted assignment in summary, finding, verification name, and verification summary rejects the original string instead of returning sanitized success. |

RED acceptance requires one test-only commit. First run the preexisting focused files at their
Cycle 9 revision and record 86/86. Then run the augmented focused suite: every retained test must
remain green, every C10 row must load/execute and fail only its intended behavior assertion, and
focused strict TypeScript must pass. Finally compare both Git object IDs and content hashes for all
three production files against `f63957ae`; any production difference invalidates RED.

### GREEN / REFACTOR contract

- Implement the ownership/state/snapshot/event/sanitizer/classifier architecture in `PLAN.md` only
  after RED is committed.
- Do not weaken, skip, rename away, or merge independent matrix assertions merely to reduce the RED
  failure count.
- Commit the first coherent GREEN after focused tests and strict focused TypeScript pass; refactor
  only while retained and Cycle 10 behavior remains green.
- Terminal evidence must record exact counts, commands, commit IDs, head, production scope, and the
  serialized full-suite environment classification.

### RED result

- PLAN checkpoint: `0eb7999f29e538c5a15d9c10f37b167be19817de`.
- One comprehensive test-only RED checkpoint:
  `6df77689d7bcd3a25d9028af258694e84d24f238`.
- The augmented focused run executed 102 tests. All 86 retained tests passed and exactly 16 named
  Cycle 10 behavior tests failed their intended assertions; no test was skipped, cancelled, or
  marked todo.
- Strict focused TypeScript passed at RED.
- Production stayed frozen: runtime `03cf916b59ef291dab309e6251a6f10ebf897eb0`, policy
  `1c8f701091a49c60cf41f83a6c16f2ae49a896c3`, and role prompts
  `cfc2d253c323ad01f34b8c9688b3bad0acd16171`.

### GREEN / REFACTOR result

- Cohesive implementation checkpoint: `a88cbe5242f070059ea49446ffac6914716a8c5d`.
- Native signal leasing, staged cleanup ownership, detached timer policy, close joining, exact SDK
  result capture, cumulative event accounting, prototype-safe snapshots, bounded typed boundary
  failures, redaction grammar, sensitive paths, capability names, and original-text terminal
  validation now satisfy C10-01 through C10-13.
- Focused GREEN: 102 passed, 0 failed; strict focused TypeScript passed.
- Every Cycle 10 RED assertion remains unchanged. Post-RED test edits were limited to interface
  alignment for retained controls: three older successful-handoff fixtures now express identical
  sensitive records on one terminal-safe line because WR-01 rejects original controls before
  redaction; one harmless documentary fixture now uses `password: ...`, while an additive exact
  assertion proves `password = ...` redacts under BL-04.

### VERIFY result

- Focused runtime/tool-policy: 102 passed, 0 failed, 0 skipped/cancelled/todo.
- Strict TypeScript: focused production/tests and all 12 non-test Shepherd production files pass
  with TypeScript 5.9.3 and explicit Pi 0.80.6 package/type roots.
- Complete serialized Shepherd suite: 239 executed, 208 passed, 31 environment-blocked only in
  `controller.test.ts` and `state-store.test.ts` where the managed sandbox denies `/bin/ps` with
  `spawn EPERM`. The isolation run excluding those files passes 165/165; the complete suite is not
  represented as green.
- Pinned Pi: binary reports `0.80.6`; offline RPC `get_commands` registers `pm-shepherd`; the focused
  no-model row passes Pi's real argument validator and executes a custom tool.
- `git diff --check`, immutable-base and frozen-head ancestry, exact issue-owned paths, and clean
  worktree pass. No push or GitHub mutation was authorized or attempted; parent orchestration owns
  the permitted-environment rerun and fresh exact-head review.

## Cycle 11 — Real Pi Compatibility, Linearizable Ownership, And Bounded Evidence

### PLAN

- Frozen candidate: `1571dc4d4f45ad4285107d04f2d7c489a7f357ab`; immutable base:
  `e659d6f1b666f58748e2d8c86599ceb4bbc62ff8`; initial worktree clean.
- Review sources read completely: `/tmp/475-REVIEW-CYCLE10-1.md` and
  `/tmp/475-REVIEW-CYCLE10-2.md`; all unique rows are accepted into C11-01 through C11-12, with
  C11-13 retaining the full Cycle 10 baseline.
- Baseline focused run: 102 passed, 0 failed/skipped/cancelled/todo.
- GSD: doctor passed; the 69-command adapter still lacks `programming-loop`, so the explicit
  manual-GSD PLAN -> RED -> GREEN/refactor -> verify fallback remains active.
- Skills: `gsd-programming-loop`, `javascript-testing-patterns`,
  `typescript-advanced-types`, `architecture-patterns`, `github-issue-first-delivery`, and all
  repository-required GSD/Pi/runtime/issue references.
- Orchestration: `read_only_spawned`; the `cycle11-pi-contract-map` explorer has no write authority,
  while this issue worker owns all PLAN, RED, GREEN, verification, commit, and handoff actions.
- Production lock: runtime `134697a62252f500b3c58082bf766a5c84766a91`, policy
  `539d061903549a764567cd1d7fad95d7d624edfe`, role prompts
  `cfc2d253c323ad01f34b8c9688b3bad0acd16171`.

### One comprehensive test-only RED contract

One test-only commit will add twelve independent top-level behavior rows exactly matching
C11-01 through C11-12 in `PLAN.md`; C11-13 is proved by all 102 retained tests. Each new row may
contain a bounded table of subcases, but its failure must identify the named behavior family rather
than a fixture/load/type error. The RED evidence must record exact executed/pass/fail counts, zero
skip/cancel/todo, strict focused TypeScript success, and exact frozen production blobs.

No production, prompt, or role behavior changes before that commit. Post-RED test edits may only be
additive or align an older fixture with a newly accepted interface; every Cycle 11 assertion must
remain intact and every alignment must be named in the GREEN checkpoint.

### GREEN / REFACTOR contract

Implement one cohesive correction around the Pi adapter, native signal lease, admission/per-run
creation state machines, stateful stream/terminal DTOs, fixed-envelope descriptor adapters,
bounded JSON/result construction, total failure sanitizer, and shared capability/path/redaction
classifiers. Do not add dependencies, widen tools, edit scheduler/controller/#478/#479 scope, or
weaken prior lifecycle, timer, parser, Pi typing, and disjoint-mutator guarantees.

### VERIFY contract

- focused runtime/tool-policy tests;
- complete serialized Shepherd classification and the established 165-test safe isolation;
- focused and all-production strict TypeScript with explicit Pi 0.80.6 roots;
- pinned offline RPC plus an actual no-model create-result/factory exercise;
- diff, immutable-base/frozen-head ancestry, JSON, credential-pattern, exact issue-owned scope, and
  clean-head checks;
- no Go/connectors/`make`, runtime services, credentials, model calls, network, push, or GitHub.

### RED result

- PLAN checkpoint: `9366296dcde200bf1f21e74d3cd8dec321581155`.
- Read-only installed-Pi contract trace: `a2a8b0e7da426f8c0c6fac91ead65d6a19c4534a`.
- One comprehensive test-only RED checkpoint:
  `c58865202623805f8877a583eecf5e301b589f3d`.
- The augmented focused run executed 114 tests: all 102 retained assertions passed and exactly 12
  named C11-01 through C11-12 rows failed their intended production assertions. There were zero
  skipped, cancelled, or todo tests.
- Strict focused TypeScript passed at RED. Production stayed exact: runtime
  `134697a62252f500b3c58082bf766a5c84766a91`, policy
  `539d061903549a764567cd1d7fad95d7d624edfe`, role prompts
  `cfc2d253c323ad01f34b8c9688b3bad0acd16171`.

### GREEN / REFACTOR result

- First cohesive runtime/policy GREEN:
  `1e605675f8e021a14ed7f709451a2d3a8111c6ad`.
- Complete-envelope stream-accounting refactor:
  `d9b4eaee71907c662f87f737c9b1a901c35146f9`.
- C11-01 through C11-12 pass through shared mechanisms: the real Pi 0.80.6 result adapter,
  descriptor-checked/inert extension runtime, native signal lease, admission and per-run creation
  state machines, cumulative assistant projections, complete terminal identity, fixed-envelope and
  bounded JSON adapters, total aggregate/error sanitizer, capability/path classifiers, and shared
  redaction grammar.
- Focused GREEN remained 114/114 after the refactor; strict focused TypeScript passed.
- Every Cycle 11 assertion remains intact. Additive stream cases prove large text signatures and
  changed diagnostic envelopes consume the aggregate budget. Retained-fixture alignments are
  limited to canonical assistant `api`/`usage`, native-only shadow hooks, and the accepted inert
  discard of hidden/symbol peers at arbitrary JSON/array boundaries.

### VERIFY result

- Focused runtime/tool-policy: 114 passed, 0 failed, 0 skipped/cancelled/todo.
- Strict no-emit TypeScript: focused production/tests and all 12 non-test Shepherd production
  files pass against explicit Pi 0.80.6 roots with TypeScript 5.9.3.
- Complete Shepherd: 251 executed, 220 passed, and 31 environment-blocked failures remain confined
  to `controller.test.ts` and `state-store.test.ts`, all at the managed-sandbox `/bin/ps`
  `spawn EPERM` boundary. Excluding only those files passes 177/177.
- Explicit Pi reports 0.80.6; offline RPC `get_commands` registers `pm-shepherd`; the focused real
  no-model factory/result row passes and cleans its actual session.
- `git diff --check`, immutable-base/frozen-head ancestry, JSON, exact issue-owned scope,
  no-Go/no-connector, and clean-head checks pass after terminal artifacts. No external mutation or
  disallowed gate was attempted; parent owns the process-capable full rerun and exact-head review.

## Cycle 12 — Pi-Faithful Lifecycle And Authority Boundaries

### PLAN / baseline

- Frozen start: `7882cd70c25971e889ec04f63b98c936d605003e`; immutable base:
  `e659d6f1b666f58748e2d8c86599ceb4bbc62ff8`; worktree clean.
- Complete review union: `/tmp/475-REVIEW-CYCLE11-1.md` and
  `/tmp/475-REVIEW-CYCLE11-2.md`; ten independent RED rows C12-01 through C12-10, with C12-11
  retaining all prior behavior.
- Focused baseline command:

  ```bash
  node --test .pi/extensions/shepherd/agent-session-runtime.test.ts \
    .pi/extensions/shepherd/tool-policy.test.ts
  ```

  Result: 114 passed, 0 failed, 0 skipped/cancelled/todo.
- Frozen production Git blobs: runtime `cfb1b40b8835c7bdffe162a7b4d368bde30d54f8`, policy
  `734927712eaadc9bb8eca383621740d59c5bb7b6`, role prompts
  `cfc2d253c323ad01f34b8c9688b3bad0acd16171`.
- GSD adapter result: doctor passes; `programming-loop` remains absent, so
  `manual_gsd_fallback` is recorded. Required testing/type/architecture/issue-delivery skills and
  repo Pi/runtime/GSD contracts were loaded before edits.
- Execution decision: `read_only_spawned`; the lifecycle mapper is read-only, and the issue worker
  alone owns artifacts, tests, production, commits, and verification.

### RED contract

One comprehensive test-only checkpoint adds exactly ten named top-level rows. Acceptance is 124
executed: all 114 retained assertions pass and exactly C12-01 through C12-10 fail at their intended
behavior assertions, with strict focused TypeScript green, no skip/cancel/todo, and all frozen
production blobs unchanged. The rows cover full Pi lifecycle/final selection/settled ownership,
real whole-session result ownership, pre-callback run admission, per-index phases, settled freeze,
SDK diagnostics, dense descriptor arrays, native abort state, hostile tool-input DTOs, and
structured/prefixed Cookie redaction.

### Ordered GREEN checkpoints

1. Real pinned whole-session no-tool prompt and lifecycle selection.
2. Real/shared one-tool multi-turn lifecycle and per-content phase machine.
3. Remaining admission/freeze/diagnostic/array/signal/tool-input/redaction mechanisms and refactor.
4. Final evidence/freeze with the complete Cycle 12 suite and declared Shepherd-only gates.

No RED assertion may be weakened. Any retained-fixture alignment must be minimal, explicit in the
GREEN record, and preserve all Cycle 11 behavioral assertions. No push/network/GitHub/live model,
Go/connectors, `make`, credentials, services, or parent-owned integration gate is permitted.

### RED evidence

- Status: captured in the test-only commit
  `58af21f14538375a39b5e48efb8935b5a1182ff0`.
- Focused result: exit 1; 124 executed, 114 retained passes, exactly 10 intended Cycle 12 behavior
  failures, and 0 skipped/cancelled/todo. Every failure is one named top-level Cycle 12 row; there
  is no fixture, loader, timeout, or compile failure.
- Strict focused TypeScript: exit 0 against the explicit Pi 0.80.6 package/type roots.
- Production lock: runtime `cfb1b40b8835c7bdffe162a7b4d368bde30d54f8`, policy
  `734927712eaadc9bb8eca383621740d59c5bb7b6`, and role prompts
  `cfc2d253c323ad01f34b8c9688b3bad0acd16171`; all remain exact to frozen start `7882cd70`.
- `git diff --check`: pass. The RED commit contains only the two issue-owned test files.

### GREEN / REFACTOR result

- Real pinned-Pi no-tool ownership checkpoint:
  `11008da118f34c91c11fa0f3b61fb9e4e8e53ae3`.
- Shared one-tool/multi-turn and per-content phase checkpoint:
  `b3a99d79cfe1b990a81d38c133100d745c0feaf4`.
- Remaining authority, signal, freeze, diagnostic, tool-input, redaction, and strict lifecycle
  refactor checkpoint: `3dc4de7114d5ee501fdc4ecfb4364244a58a3ab9`.
- C12-01 through C12-10 pass through one complete Pi lifecycle machine, descriptor-owned run
  admission, explicit per-index assistant phases, settled freeze/unsubscribe, SDK-aware diagnostic
  projection, fresh dense request arrays, native abort-state leases, bounded tool-input DTOs, and
  the shared Cookie/Set-Cookie redaction grammar. The legacy assistant-only acceptance path was
  removed rather than retained as a compatibility bypass.
- The actual pinned Pi row owns two public `createAgentSession` results. The no-tool run performs
  one offline provider turn; the one-tool run performs an intermediate tool-use turn with three
  content updates, one real scoped `workspace_read`, the emitted tool result, and a subsequent final
  turn. The row asserts two prompts, three provider calls, one workspace callback, complete
  `agent_end`/`agent_settled` traces, zero `fetch` calls, and two disposals after settlement.
- Retained-fixture alignments add the complete Pi prefix/suffix to formerly assistant-only success
  fixtures, preserve hidden-field and sparse-array assertions at their intended boundary, and raise
  the Cycle 11 stream test's aggregate event allowance from 4,096 to 8,192 bytes to account for the
  newly required user/turn/agent lifecycle envelopes. Every dishonest stream case still rejects.
- Focused GREEN: 124 passed, 0 failed, 0 skipped/cancelled/todo; strict focused TypeScript passed.

### VERIFY result

- Verified implementation checkpoint: `3dc4de7114d5ee501fdc4ecfb4364244a58a3ab9`; immutable base:
  `e659d6f1b666f58748e2d8c86599ceb4bbc62ff8`.
- Focused runtime/tool-policy: 124 passed, 0 failed, 0 skipped/cancelled/todo.
- Strict no-emit TypeScript: focused production/tests and all 12 non-test Shepherd production
  files pass against explicit Pi 0.80.6 roots with TypeScript 5.9.3.
- Complete serialized Shepherd classification: 261 executed, 230 passed, and 31 failures remain
  confined to the pre-existing controller/state-store process-identity paths where this managed
  sandbox rejects child process creation with `spawn EPERM`. Every reported failure has that same
  environmental cause; the result is not represented as a green suite.
- Safe isolation excluding only `controller.test.ts` and `state-store.test.ts`: 187 passed,
  0 failed, 0 skipped/cancelled/todo.
- Explicit pinned Pi 0.80.6 offline RPC `get_commands` exits 0 and registers `pm-shepherd`. Its only
  warnings are the managed filesystem denial for the global settings lock. No model, credential,
  service, or network path is used.
- `git diff --check`, immutable-base ancestry, exact issue-owned changed paths, RUN-STATE JSON, and
  credential-pattern/no-dependency scans pass. No Go, connector, `make`, service, push, GitHub,
  live-model, or credential action was attempted. Parent orchestration owns the process-capable
  full rerun and independent exact-head review.

## Cycle 13 — Bounded Public Authority And Tool-Lifecycle Correlation

### PLAN / baseline

- Frozen start: `5dafc5725167bb74ce88a723073b8c4ceb8314e0`; immutable base:
  `e659d6f1b666f58748e2d8c86599ceb4bbc62ff8`; worktree clean.
- Complete review union: `/tmp/475-REVIEW-CYCLE12-1.md` and
  `/tmp/475-REVIEW-CYCLE12-2.md`; seven independent RED rows C13-01 through C13-07, with C13-08
  retaining all prior behavior.
- Focused baseline command:

  ```bash
  node --test .pi/extensions/shepherd/agent-session-runtime.test.ts \
    .pi/extensions/shepherd/tool-policy.test.ts
  ```

  Result: 124 passed, 0 failed, 0 skipped/cancelled/todo.
- Frozen production Git blobs: runtime `62851ca6bb4b4a7bd0b65d4d1415f992b1455603`, policy
  `fd6a0e8db7f06ade82b852141eb2a497614aea79`, role prompts
  `cfc2d253c323ad01f34b8c9688b3bad0acd16171`.
- GSD adapter result: doctor passes; `programming-loop` remains absent, so
  `manual_gsd_fallback` is recorded. All required testing/type/architecture/issue-delivery skills
  and repo Pi/runtime/GSD contracts were reloaded before edits.
- Execution decision: `read_only_spawned`; one seam mapper is read-only, and the issue worker alone
  owns artifacts, tests, production, commits, verification, and handoff.

### RED contract

One comprehensive test-only checkpoint adds exactly seven named top-level behavior rows. Acceptance
is 131 executed: all 124 retained assertions pass and exactly C13-01 through C13-07 fail at their
intended behavior boundaries, with strict focused TypeScript green, no skip/cancel/todo, and all
frozen production blobs unchanged. The rows cover capped dense-array ingress, SDK-seam lifecycle
rechecks, public policy array capture, semantic capability denial, split qualified sensitive keys,
public role-prompt array capture, and complete authorized tool-call/result/turn correlation.

For C13-01, bounded canonical influence supersedes literal raw-source shape proof: ECMAScript has
no streaming own-key API, so proving absence of arbitrary hidden/symbol peers necessarily
materializes an attacker-sized key set. RED therefore requires exactly one bounded length read,
bounded indexed own-data descriptor reads, zero enumeration/iterator/peer accessor work, and zero
extra-field influence. The fresh captured array is the only accepted downstream authority.

### Ordered GREEN / REFACTOR checkpoints

1. Intrinsic dense-array canonical influence capture shared by runtime, public tool policy, and
   public role prompts, plus post-callback scope assertions.
2. Closed capability and split-qualified sensitive-key classifiers shared by all current consumers.
3. Per-tool-call Pi lifecycle correlation through assistant call, execution, result message,
   `turn_end`, next turn, and final handoff, retaining real pinned no-tool and one-tool behavior.
4. Refactor/evidence freeze with focused/all-production strict TypeScript and the declared
   Shepherd-only gates.

No RED assertion may be weakened. Production remains locked until the compiled seven-row RED is
captured. No push/network/GitHub/live model, Go/connectors, `make`, credentials, services, or
parent-owned integration gate is permitted.

### RED evidence

- Test-only checkpoint: `974d2e795038d5531c9aca39fbdcfbe73b2caf8a`.
- Focused command: exit 1; 131 executed, 124 retained passes, exactly 7 intended Cycle 13 failures,
  0 skipped/cancelled/todo.
- Exact failing rows: bounded indexed request influence; SDK-seam cancellation terminality;
  intrinsic public policy arrays; structural capability names; split-qualified sensitive keys;
  intrinsic immutable prompt arrays; complete authorized Pi tool lifecycle correlation.
- No fixture/load/timeout failure occurred. Focused strict TypeScript passes with the explicit Pi
  0.80.6 roots and TypeScript 5.9.3.
- Production Git blobs remain exactly frozen at Cycle 13 start. The test-only commit changes only
  `agent-session-runtime.test.ts` and `tool-policy.test.ts`; `git diff --check` passes.

### GREEN

- Status: captured at first cohesive GREEN `48f546a5`.
- All seven top-level C13 behavior rows pass together with all 124 retained rows: 131 executed,
  131 passed, 0 failed/skipped/cancelled/todo.
- Minimal mechanisms: bounded indexed-descriptor array projection; post-SDK-seam lifecycle
  assertions; direct public policy/prompt snapshots; structural capability denial; bounded dotted
  sensitive compounds; and one exact assistant/execution/result/turn tool-call identity.

### REFACTOR / VERIFY

- Status: captured after no-contract-widening refactor `e50b5f97`.
- The refactor centralizes capability families and adds terminal/program/cookie synonym controls,
  plus exact tool-result result/error mismatch cases. No C13 assertion was removed or weakened.
- Focused: 131/131 pass. Safe isolation: 194/194 pass.
- Complete serialized Shepherd: 268 executed, 237 passed, 31 environment-classified failures in
  only `controller.test.ts` and `state-store.test.ts`; every one reports managed-sandbox
  `spawn EPERM`. Zero skipped/cancelled/todo.
- Focused strict TypeScript and all 12 non-test Shepherd production TypeScript files pass with the
  explicit Pi 0.80.6 roots. Explicit Pi 0.80.6 offline RPC registers `pm-shepherd` and exits 0.
- Diff, immutable-base/frozen-start ancestry, JSON, credential-pattern, no-dependency/Go/connector,
  exact 20-path scope, and clean-worktree checks pass at the terminal artifact boundary.
- Both complete Cycle 12 reports were re-read after implementation; every blocker remains
  represented by one passing C13 row. External/process-capable rerun and independent review remain
  parent-owned; this lane performed no push or other external mutation.

## Cycle 13 Gate History

| Checkpoint | Result | Evidence |
|---|---|---|
| PLAN | pass | artifact `61a364e0`; exact start/base and seven-row matrix recorded before tests |
| Seam-map amendment | pass | read-only mapping artifact `5e86520c`; bounded canonical-influence contract recorded |
| RED | expected fail | test-only `974d2e79`; 124 retained pass plus exactly 7 intended failures; production frozen |
| RED evidence | pass | artifact `6d3ce03a`; behavior-specific failures and frozen blobs recorded |
| GREEN | pass | `48f546a5`; all 131 focused rows pass |
| REFACTOR | pass | `e50b5f97`; strengthened classifiers/lifecycle controls; 131/131 retained |
| Strict TypeScript | pass | focused source/tests and all 12 production files, TypeScript 5.9.3 / Pi 0.80.6 roots |
| Offline RPC | pass | explicit Pi 0.80.6 `get_commands`; `pm-shepherd` registered |
| Safe isolation | pass | 194 passed, 0 failed/skipped/cancelled/todo |
| Complete Shepherd | environment-blocked | 268 executed; 237 pass; unchanged 31 controller/state-store `spawn EPERM` failures |
| Integrity / scope | pass | diff, ancestry, JSON, credential/dependency/Go/connector scans; exact same 20 paths |
