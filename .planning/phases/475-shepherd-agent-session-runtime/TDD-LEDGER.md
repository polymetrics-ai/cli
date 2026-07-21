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
