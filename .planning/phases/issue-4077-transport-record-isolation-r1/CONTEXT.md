# #4077 — Transport record-isolation correction context

**Gathered:** 2026-08-13
**Status:** TDD plan ready; RED pending
**Issue:** [#4077](https://github.com/polymetrics-ai/cli/issues/4077)
**Parent chain:** #4077 → #3864 → #3862 → #4015
**Stacked parent:** [#4019](https://github.com/polymetrics-ai/cli/pull/4019) / `feat/3862-any-to-any-transport`
**Exact accepted base:** `c67f40a5ff67a131950f3123e70527027dca8493`

## Lifecycle fallback

`scripts/gsd doctor`, all five required `scripts/gsd sources` resolutions, and generated
`discuss-phase --auto` / `plan-phase --tdd` prompts completed. `gsd-sdk query init.phase-op
issue-4077-transport-record-isolation-r1` reports `phase_found: false`: the issue is a named
follow-on, not a numeric roadmap phase. This Codex workspace is not a compatible Pi runtime, and
the repository's canonical single-worker contract forbids spawning lifecycle roles. This directory
therefore records the required inline/manual GSD fallback — discussion, TDD planning, execution,
verification, review, and no-mistakes — rather than waiving any gate.

## Phase boundary

Repair only the residual mutable-value aliasing in the accepted closed Transport record-copy
boundary. A source record must stay independent when a warehouse stage or destination mutates an
explicitly supported nested value, and an unsupported mutable value must fail closed instead of
crossing that boundary by alias.

## Locked decisions

- **D-01:** The reproduction baseline is exactly `c67f40a5ff67a131950f3123e70527027dca8493`,
  fetched from the current remote parent branch. No correction starts from `main` or rewrites the
  shared parent.
- **D-02:** `json.RawMessage` and `map[string]string` are explicit supported values. They receive
  independent storage at direct record clone, source-page-to-stage, and stage-workset-to-destination
  boundaries; nesting is through the existing recursive `connectors.Record`, `map[string]any`,
  `[]any`, and `[]connectors.Record` cases.
- **D-03:** Existing `[]byte` and `map[string]any` behavior is a disconfirming control, not a
  substitute for the new regression. The exact-head temporary test showed both controls independent.
- **D-04:** Keep the contract closed: supported scalar JSON-like values and explicitly enumerated
  mutable containers are copied or passed by value; unknown values, including unknown mutable maps,
  slices, pointers, functions, and channels, are rejected with context before they cross a boundary.
  Do not use reflection-based arbitrary cloning or add provider value types.
- **D-05:** Fail-closed rejection is propagated from clone helpers to both source-page staging and
  stage-workset destination application. It must prevent downstream execution, not replace an
  unknown value with `nil`, panic, or silently preserve it.
- **D-06:** No polling, provider adapter, GitHub/PostgreSQL, rate/auth, certification, stage format,
  checkpoint/acknowledgement/CAS semantics, registry topology, CLI, docs/website, or generated
  surface changes belong here. CLI help/manual/website parity is explicitly not applicable.
- **D-07:** The behavioral RED is committed before production code. It proves direct aliasing and
  real source → stage → destination mutation paths, plus a closed-boundary unsupported-value refusal.
- **D-08:** Verification includes focused normal/race transport tests, preserved app canonical-mode
  evidence, selected split repository gates, a real `pm` build, and local no-mistakes. Live
  GitHub/PostgreSQL credentials are irrelevant to this in-memory boundary and will not be sought.
- **D-09:** Use at most five no-mistakes correction loops. The canonical child command skips
  push/PR/CI for the known unsafe stacked-draft route; if delivery requires a manual push/PR
  exception, record `needs-decision` and stop before any such action.

## Canonical references

- `AGENTS.md` — mandatory issue-first, GSD/TDD, stacked delivery, verification, and Transport
  boundary rules.
- `.agents/agentic-delivery/contracts/issue-agent-contract.md` — issue/PR evidence contract.
- `.agents/agentic-delivery/contracts/parent-orchestrator-contract.md` — single-worker stacked
  parent ownership and no-role-spawning rule.
- `.agents/agentic-delivery/canonical/delivery-contract.json` — canonical state vocabulary and
  `no-mistakes axi run --skip=push,pr,ci` child command.
- `.agents/agentic-delivery/references/gsd-pi-adapter.md` — shell prompt and inline-fallback route.
- `.agents/agentic-delivery/references/required-skills-routing.md` — required Go skills.
- `.agents/agentic-delivery/workflows/stacked-parent-subissue-workflow.md` — stacked draft topology.
- `.agents/agentic-delivery/workflows/automated-review-routing-loop.md` and
  `.agents/agentic-delivery/workflows/claude-review-loop.md` — later review routing.
- `.planning/phases/issue-3864-closed-transport-dispatch-r1/CONTEXT.md`, `TDD-LEDGER.md`, and
  `VERIFICATION.md` — accepted closed-dispatch contract and preserved regression scope.
- `.planning/phases/issue-4067-acknowledged-completion-rebase-r1/CONTEXT.md` and `TDD-LEDGER.md`
  — accepted parent evidence convention.
- `internal/synctransport/types.go` — current `cloneRecordValue` closed copy implementation.
- `internal/synctransport/orchestrator.go` — source-page and workset clone boundaries.
- `internal/synctransport/transport_test.go` — closest focused clone/stage fixtures.

## Existing code insights

- `connectors.Record` is `map[string]any`; its closed Transport boundary currently describes
  JSON-like composite values but has no generic/reflection cloning mechanism.
- `cloneSourcePage` runs immediately before `WarehouseStage.Stage`; `cloneWarehouseWorkset` runs
  immediately before `DestinationExecutor.ApplyDestination`. These are the only two record-copy
  crossings in the core orchestrator.
- Existing `TestCloneRecordCopiesBinaryValuesAtEveryNestingLevel` covers `[]byte` inside direct,
  map, and list contexts. `TestOrchestratorStagesDeepCopyOfProviderRecords` proves nested
  `map[string]any` stage isolation.

## Deferred ideas

None. Any broader record-schema/type-validation or provider-value redesign needs a separate issue.
