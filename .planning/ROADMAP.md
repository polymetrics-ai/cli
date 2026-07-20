# Roadmap: Polymetrics CLI Connector Parity

**Generated via:** official GSD Core Pi adapter command path
**Commands:** `scripts/gsd prompt onboard --fast --skip-phases`, `scripts/gsd prompt new-project --from-existing --non-interactive`, `scripts/gsd prompt milestone-summary --planning-only`
**Upstream GSD Core:** `open-gsd/gsd-core@20297a8ff941378b8615a5d3e8629e52c10a0f9d`
**Phase regeneration:** skipped by request

## Overview

This roadmap replaces the legacy custom `.planning/` tree with an official GSD Core brownfield plan for connector parity. The repo-local Pi adapter is the command path for future GSD work: agents use `/gsd <command>` or generated `/gsd-*` aliases in Pi, and non-interactive automation uses `scripts/gsd prompt <command>`.

The first delivery gate remains inventory reconciliation. No connector fanout starts until the repository and upstream documentation are reconciled across all connector technologies, canonical operation de-duplication is applied, and human-gated risk classification is reviewed.

## North-Star Milestone

Deliver connector parity that is:

1. **Surface-complete** — covers REST, GraphQL, XML/SOAP, CSV/NDJSON/report export, binary, file/object, SQL/CDC, queue/event/webhook/audit-log, native protocol, direct-read, and reverse ETL write surfaces.
2. **Safety-gated** — keeps secrets, auth scopes, destructive/admin operations, live credential checks, reverse ETL execution, new dependencies, quality-gate reductions, and `main` merges under human control.
3. **De-duplicated** — assigns each documented upstream operation exactly one primary classification.
4. **Conformance-backed** — derives parity claims from generated inventory, fixture/replay gates, conformance checks, and certification status.
5. **Agent-runnable** — uses official GSD Core commands through `.pi` and `scripts/gsd`, not runtime-specific copied command files.
6. **Skill-routed** — agents and subagents load required Go/design skills such as `golang-how-to`, `golang-cli`, `golang-testing`, `golang-security`, `golang-documentation`, `frontend-design`, `web-design-guidelines`, and `vercel-react-best-practices` as applicable.
7. **Help/docs/website-complete** — every CLI-visible feature keeps runtime help, bare namespace command behavior, `docs/cli/**`, website docs, generated help/manual artifacts, and tests in parity.

## Workstreams

### Active Sibling Milestone: CLI Architecture v2

**Parent issue:** [#397](https://github.com/polymetrics-ai/cli/issues/397)
**Stage 0:** [#398](https://github.com/polymetrics-ai/cli/issues/398)
**Parent branch:** `feat/cli-architecture-v2`
**Parent PR:** draft PR to `main`; merge remains human-gated.

CLI Architecture v2 is a sibling program that preserves the connector-parity roadmap while improving the CLI substrate. Source-of-truth documents:

- `docs/plans/cli-architecture-v2-improvement-plan.md`
- `docs/prompts/cli-architecture-v2-gsd-execution-prompt.md`
- `docs/design/tui-ux-design.md`
- `docs/design/terminal-ui-research-and-design-system.md`
- `docs/adr/0002-cobra-viper-cli-framework.md`
- `docs/adr/0003-interactive-tui-layer.md`
- `docs/adr/0004-opentelemetry-observability.md`
- `.planning/traces/cli-architecture-v2-issue-backlog.md`
- `.planning/traces/cli-architecture-v2-pi-prompts.md`

#### CLI Architecture v2 phase roster

| Phase | Issue | Track | Summary | Dependency gate |
|---:|---|---|---|---|
| 0 | #398 | planning | Register active GSD milestone, source docs, parent branch, and draft parent PR | none |
| 1 | #399 | A | Golden transcript safety net + docs-generate-diff test | #398 |
| 2 | #400 | A | Cobra router shell strangler, byte-identical dispatcher | #399 |
| 3 | #401 | A | Typed Viper configuration with explicit precedence | #400 |
| 4 | #402 | A | Migrate scattered environment reads onto config | #401 |
| 5 | #403 | B | Dependency-free progress event bus and instrumentation | #402 |
| 6 | #404 | C | Redacted per-run `slog` foundation and Temporal logger bridge | #402 |
| 7 | #405 | B | stdin+stdout TTY gate, `--plain`/`--no-input`, and `--progress ndjson` | #403 |
| 8 | #406 | A | Nativize pilot `catalog` namespace | #402 |
| 9 | #407 | A | Nativize remaining namespaces through serialized grandchildren #421–#437 | #406, #437 |
| Design gate | #462 | B | Freeze Bubble Tea interaction, responsive layout, chart grammar, design skill, and TUI worker prompts | #405; before #408/#409/#411/#412/#414/#416/#418/#463/#469 |
| 10 | #408 | B | Flow and ETL run dashboards | #405, #462/D-TUI |
| 11 | #409 | B | Flow and schedule creation wizards | #408, #462/D-TUI |
| 12 | #410 | C | Opt-in OpenTelemetry tracing | #402 |
| 13 | #411 | B | Connector browser, `query tables`, and human-first `pm query` workspace with `pm query grid` alias | #409, #462/D-TUI |
| 13b | #463 | B | Read-only query charts and reusable terminal dashboard compositions | #411, #462/D-TUI; renderer dependency requires explicit human approval |
| 14 | #412 | B | Terminal docs viewer | #409, #462/D-TUI |
| 15 | #413 | A | Connector-aware shell completion | #407 |
| 16 | #414 | B | Certify batch table and RLM agent dashboards | #407, #408, #462/D-TUI |
| 17 | #415 | C | OpenTelemetry metrics | #410 |
| 18 | #416 | B | Human-first `pm reverse` guided session with `pm reverse guide` alias | #409, #462/D-TUI |
| 18b | #469 | B | TTY-progressive credential and connection setup | #409, #462/D-TUI; child of #416 |
| 19 | #417 | A | Help tree deepening and generated man pages | #411, #412, #413, #414, #416, #469 |
| 20 | #418 | B | Accessibility audit and `pm a11y` topic | #411, #412, #414, #416, #469, #462/D-TUI; #463 after #411 when chart slice is included |
| 21 | #419 | C | Optional OpenTelemetry log bridge, human-skippable | #404, #410 |
| 22 | #420 | A | Architecture v2 cleanup, docs updates, final verification | #415, #417, #418 |

#### CLI Architecture v2 dependency waves

1. Bootstrap serial chain: #398 → #399 → #400 → #401 → #402.
2. First parallel fan-out after #402: #403, #404, #406, and #410 may run concurrently in isolated worktrees after write-scope collision checks.
3. Phase 9 namespace grandchildren #421–#437 are serialized because they share central CLI routing/help files.
4. TUI design issue #462/D-TUI must integrate before production UI work starts. UX fan-out after
   #409: #411, #412, #416, and #469 may run in isolated worktrees after write-scope collision
   checks; #414 waits for #407, #408, and #462/D-TUI. #411 owns the human-first `pm query`
   workspace and its `pm query grid` alias; #416 owns the human-first `pm reverse` workspace and
   its `pm reverse guide` alias; #469 owns
   credential and connection setup. Chart child #463 follows #411 and #462/D-TUI, then joins the
   #418 accessibility convergence when included. Parent orchestrator must update GitHub blocked-by
   metadata; worker docs do not mutate issue metadata.
5. Convergence: #413 after #407; #417 and #418 after their UI/CLI dependencies; #420 last. Phase #419 is optional and requires an explicit human decision if skipped or if dependency budget changes.

#### CLI Architecture v2 gates

- Every behavior-changing issue must use repo-local GSD/TDD, record red → green → refactor evidence, and keep issue-specific plan/TDD/verification artifacts current.
- CLI-visible work must keep runtime help, bare namespace behavior, `docs/cli/**`, website docs, generated help/manual artifacts, completion metadata, and tests in parity.
- Dependency additions are allowed only in the phase/version lines approved by ADRs 0002–0004; any deviation is a human gate.
- TUI workers must load `bubble-tea-tui-design`, cite both design documents, and record
  stdin+stdout TTY activation, `stdin-piped+stdout-TTY` fallback, `stdout-piped`, `CI`, `--json`,
  `--plain`, `--no-input`, modal-key, responsive-layout, accessibility/plain fallback, sanitation,
  cancellation, and JSON/stdout/stderr parity evidence. `--plain`, `--json`, and `--no-input` must
  bypass Bubble Tea, Huh, and all prompts; sequential prompts are allowed only in explicit
  accessible mode after the stdin+stdout TTY gate passes and no bypass flag is set. Piped/non-TTY
  stdin must not be consumed, hang, or be bypassed through `/dev/tty`. `ntcharts/v2` remains
  unapproved until a dedicated chart child issue #463 receives an explicit human dependency
  decision.
- TTY-progressive action commands prompt only for missing fields. Fully specified invocations run
  directly and complete-but-invalid invocations return ordinary validation errors. Ordinary bare
  namespaces remain contextual help; eligible dual-TTY bare `pm query` and bare `pm reverse` are
  the narrow human-first workspace exception, with deterministic help on every bypass path. Agent
  documentation uses `--json --no-input`; long-running commands may
  add `--progress ndjson`. Do not introduce a global `--agent-mode`, because that name already has
  query-specific result-shaping semantics.
- Parent PR to `main` remains draft until all required sub-issues are integrated, final verification passes, automated review coverage is recorded, and a human is asked for final approval.

### 0. GSD Runtime and Agent Enablement

**Goal:** Make the official GSD command surface the default for humans, Pi sessions, and reusable agents.

**Status:** In progress on issue #122.

**Required outcomes:**

- `.gsd/upstream.lock.json` pins official GSD source.
- `.gsd/commands.json` lists official commands generated from official docs.
- `.pi/extensions/gsd/index.ts` exposes `/gsd` plus `/gsd-*` aliases.
- `.pi/skills/gsd-core/SKILL.md` sets default planning/implementation behavior.
- `.agents/**` instructions route agents/subagents through the Pi adapter or `scripts/gsd`.
- `.agents/agentic-delivery/references/required-skills-routing.md` defines required Go/design skill routing for agents and subagents.
- `.agents/agentic-delivery/references/runtime-rlm-website-integration.md` preserves runtime/RLM/Pi-agent/website integration knowledge for Podman, PostgreSQL, DragonflyDB/Redis-compatible coordination, Temporal, RLM agent mode, and website docs.
- `.agents/agentic-delivery/references/cli-help-docs-website-parity.md` defines the CLI help/manual/website parity gate.
- Manual-GSD fallback is only used when the adapter is unavailable and must be recorded.
- Runtime-backed checks remain optional and gated; dependency-free CLI paths must not require PostgreSQL, DragonflyDB, Temporal, or Podman.

### 1. Inventory and Surface Reconciliation

**Goal:** Produce a current, generated, de-duplicated connector parity baseline before fanout.

**Depends on:** Workstream 0 for command/runtime consistency.

**Success criteria:**

- Active `.planning/` is official GSD Core shaped and the pre-rebootstrap archive is recorded outside active planning.
- Inventory reconciles bundles, hooks, natives, docs, API/surface manifests, streams, writes, binary surfaces, direct-read surfaces, native protocol surfaces, blockers, quarantine, conformance, and certification.
- Inventory classifies REST, GraphQL, XML/SOAP, CSV/NDJSON, binary, file/object, SQL/CDC, queue/event/webhook/audit-log, native, direct-read, and mutation surfaces.
- Canonical operation identity avoids duplicate work from multiple docs pages/specs.
- Connector fanout is blocked until inventory outputs are reviewed.

### 2. Durable Read and ETL Parity

**Goal:** Bring product-safe documented durable record collections to CLI and ETL parity.

**Depends on:** Workstream 1.

**Success criteria:**

- Every product-safe documented stream/report/event log/feed/table/queue message/CDC event/durable record collection is covered by a connector read surface or typed exclusion.
- `pm connectors inspect`, catalog, read, and ETL surfaces agree on stream names, schemas, sync modes, cursor behavior, and limitations.
- Incremental, pagination, cursor, schema, and projection behavior is verified through conformance fixtures or typed blockers.
- CLI-visible read surfaces update runtime help, bare namespace summaries, `docs/cli/**`, website docs, generated help/manual artifacts, and tests.

### 3. Direct-Read, Binary, and Native Surface Parity

**Goal:** Cover product-safe non-stream and non-REST read surfaces without forcing them into the wrong abstraction.

**Depends on:** Workstream 1.

**Success criteria:**

- Direct-read operations are classified separately from durable ETL streams.
- Binary operations have product-safe transfer surfaces or typed exclusions.
- GraphQL, XML/SOAP, CSV/NDJSON/report export, file/object, SQL/CDC, queue/event/webhook/audit-log, and native protocol reads use an appropriate declarative, hook, native, direct-read, binary, or exclusion path.
- CLI help/manual/website docs explain direct-read and binary surfaces without misclassifying them as durable ETL streams.
- Admin/elevated/destructive direct-read or binary operations remain human-gated.

### 4. Reverse ETL and Mutation Parity

**Goal:** Map safe writes/mutations to reverse ETL actions with approval gates.

**Depends on:** Workstream 1.

**Success criteria:**

- Product-safe mutations across REST, GraphQL, XML/SOAP, file/object, queue, database/native, and other protocol-specific operations map to reverse ETL actions or typed exclusions.
- Plan, preview, approval, and execute semantics are preserved for every write path.
- Runtime help, manual docs, and website docs consistently describe plan → preview → approval → execute for write paths.
- Destructive/admin/elevated-scope writes are human-gated and never exposed as generic raw write tools.

### 5. Conformance and Certification Enforcement

**Goal:** Make validated gates authoritative for connector parity status.

**Depends on:** Workstreams 2–4.

**Success criteria:**

- Conformance validates schemas, fixtures, surface manifests, streams, writes, direct-read/binary metadata, cursor/pagination behavior, docs, and de-duplication contracts.
- Certification reports distinguish replay/fixture success, live success, missing credentials (`uncertified`), typed blockers, and failures.
- Public or PR-facing connector parity claims derive from generated conformance/certification artifacts.

## Phase Mapping

The existing phase files are preserved in this refresh by request. They still map to the roadmap as follows:

| Existing phase | Roadmap workstream | Status |
|---|---|---|
| Phase 1: Inventory and Surface Reconciliation | Workstreams 0–1 | In progress |
| Phase 2: Durable Read and ETL Parity | Workstream 2 | Not started |
| Phase 3: Direct-Read, Binary, and Native Surface Parity | Workstream 3 | Not started |
| Phase 4: Reverse ETL and Mutation Parity | Workstream 4 | Not started |
| Phase 5: Conformance and Certification Enforcement | Workstream 5 | Not started |

## Progress

| Workstream | Status | Notes |
|---|---|---|
| 0. GSD Runtime and Agent Enablement | In progress | Official docs pinned; Pi adapter and agent guidance refreshed. |
| 1. Inventory and Surface Reconciliation | In progress | Planning exists; generated inventory still must be reviewed before fanout. |
| 2. Durable Read and ETL Parity | Not started | Blocked on Workstream 1. |
| 3. Direct-Read, Binary, and Native Surface Parity | Not started | Blocked on Workstream 1. |
| 4. Reverse ETL and Mutation Parity | Not started | Blocked on Workstream 1. |
| 5. Conformance and Certification Enforcement | Not started | Blocked on Workstreams 2–4. |

## Command Path for Future Updates

Use official GSD commands through the repo-local adapter:

```bash
scripts/gsd doctor
scripts/gsd list
scripts/gsd prompt map-codebase --fast
scripts/gsd prompt new-project --from-existing --non-interactive
scripts/gsd prompt plan-phase 1 --skip-research
scripts/gsd prompt programming-loop init --phase <phase> --dry-run
```

In Pi after trust/reload:

```text
/gsd doctor
/gsd list
/gsd map-codebase --fast
/gsd plan-phase 1 --skip-research
/gsd-programming-loop init --phase <phase> --dry-run
```

---
*Roadmap refreshed: 2026-07-08 via repo-local official GSD Core Pi adapter; `.planning/phases/**` intentionally unchanged.*
