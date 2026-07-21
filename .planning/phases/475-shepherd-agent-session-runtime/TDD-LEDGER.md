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
