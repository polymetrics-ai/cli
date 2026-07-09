# Architecture

**Analysis Date:** 2026-07-08
**Generated via:** official GSD Core Pi adapter command path
**Commands:** `scripts/gsd prompt map-codebase --fast`, `scripts/gsd prompt docs-update .planning AGENTS.md .agents --planning-only`
**Upstream GSD Core:** `open-gsd/gsd-core@20297a8ff941378b8615a5d3e8629e52c10a0f9d`

## Pattern Overview

**Overall:** Go-only local-first CLI monolith with declarative connector runtime, optional hook/native escape hatches, optional Podman-backed runtime services, RLM/Pi-agent mode, safety-gated reverse ETL, website docs, and repo-local GSD/Pi planning automation.

**Key Characteristics:**

- Single executable `pm` built from `cmd/pm`.
- CLI layer delegates to application services and connector registry/runtime packages.
- Connector definitions are embedded data under `internal/connectors/defs` and interpreted by `internal/connectors/engine`.
- Tier 2 hooks and Tier 3 native implementations exist for cases that cannot be represented safely or clearly in declarative JSON bundles.
- Safety-gated reverse ETL uses plan, preview, approval, and execute flow.
- Conformance and certification make connector capability claims testable.
- GSD planning is official-docs-backed through `scripts/gsd` and `.pi` resources rather than copied runtime-specific command files.
- Optional runtime integration uses Podman-first local orchestration, PostgreSQL, DragonflyDB/Redis-compatible coordination, and Temporal.
- RLM agent mode is opt-in and runtime-backed through Temporal plus a Podman-managed agent image.
- Website docs use Next.js 16, React 19, and Fumadocs, with generated docs/connectors data.

## Current Structural Snapshot

| Area | Count / Signal |
|---|---:|
| Connector definition directories | 547 |
| `api_surface.json` files | 547 |
| Stream definition files | 7159 |
| Write definition files | 5699 |
| Hook directories | 78 |
| Native connector directories | 37 |
| Go source files under `cmd/` + `internal/` | 491 |
| Repo-local YAML agent specs | 14 |
| Official GSD commands exposed by adapter | 69 |

These are quick map inputs, not certification claims.

## Layers

### CLI Layer

**Purpose:** Parse commands, flags, JSON output, docs/help, and user-facing errors.

**Contains:** `cmd/pm`, `internal/cli`.

**Depends on:** app services, connector registry, runtime helpers, safety redaction.

**Key safety rule:** Use `pm help <topic>` before unfamiliar commands; prefer `--json` for machine-readable output.

**Help/docs parity rule:** CLI-visible changes must keep runtime help, bare namespace command behavior, `docs/cli/**`, website docs, generated help/manual artifacts, and tests aligned. Namespace command groups such as `pm connectors` should show contextual help/subcommand summary when invoked without an action.

### Application Layer

**Purpose:** ETL, reverse ETL, query, flow, schedule, runtime-backed execution, state, warehouse, and approval flows.

**Contains:** `internal/app`, `internal/flow`, `internal/schedule`, `internal/runtime`, `internal/state`, `internal/vault`.

**Depends on:** connector interfaces, local state/warehouse, safety utilities.

**Key safety rule:** Reverse ETL remains plan → preview → approval → execute.

### Connector Runtime Layer

**Purpose:** Connector definitions, runtime execution, hooks/native overrides, conformance, certification.

**Contains:** `internal/connectors/defs`, `engine`, `hooks`, `native`, `connsdk`, `conformance`, `certify`, `bundleregistry`.

**Depends on:** connector schemas/fixtures, HTTP helpers, native clients, hook registration.

**Runtime model:** Declarative-first; use hooks/native code only when required by protocol, auth, pagination, binary, streaming, or product behavior that does not fit JSON bundle semantics.

### Optional Runtime and RLM Layer

**Purpose:** Provide opt-in runtime-backed execution, performance comparisons, RLM agent mode, and local integration testing without weakening the dependency-free default path.

**Contains:** `deploy/compose`, `deploy/temporal`, `scripts/runtime.sh`, `scripts/setup-runtime-*`, runtime docs, `pm runtime`, `pm perf --runtime`, `pm rlm`, `pm agent image`, and `pm worker` surfaces.

**Depends on:** Podman or Docker Compose for local orchestration; PostgreSQL for durable control-plane/run ledger data; DragonflyDB/Redis-compatible coordination; Temporal for durable workflows and RLM agent mode.

**Key safety rule:** Runtime-backed checks are optional and explicitly gated. Do not store plaintext credentials, raw row payloads, approval tokens, or large batches in Temporal history or DragonflyDB. Do not make runtime services mandatory for default unit tests.

### Website Documentation Layer

**Purpose:** Public documentation, CLI reference, architecture docs, connector catalog pages, and generated website data.

**Contains:** `website/content/docs`, `website/app`, `website/components`, `website/scripts`, and `website/package.json`.

**Depends on:** Next.js 16, React 19, Fumadocs, Radix UI, Lucide icons, Tailwind CSS v4 tooling.

**Key parity rule:** Runtime/RLM/CLI-visible changes update website docs or record explicit not-applicable notes. Website/design work loads the required design skills from `.agents/agentic-delivery/references/required-skills-routing.md`.

### Planning and Agent Layer

**Purpose:** Official GSD command execution, issue-first delivery contracts, reusable agent specs, migration inventories, parity status, and PR review routing.

**Contains:** `.gsd/`, `.pi/`, `.agents/`, `.planning/`, `docs/migration`, `docs/architecture`.

**Depends on:** official `open-gsd/gsd-core@next` docs snapshot, repo-local Pi extension, issue contracts, CodeRabbit/Copilot routing docs.

**Runtime model:**

- Pi interactive: `/gsd <command>` and generated `/gsd-*` aliases from `.pi/extensions/gsd/index.ts`.
- Shell/non-interactive: `scripts/gsd prompt <command> [args...]`.
- Agent/subagent default: load `.pi/skills/gsd-core/SKILL.md`, then follow agent contract.
- Fallback: only record manual-GSD fallback when the adapter is unavailable.

## Data Flow

### ETL Read Flow

1. User or agent invokes `pm etl ...` / connector read surface.
2. CLI loads project config and credential references.
3. App resolves connector, stream, sync mode, state, and destination.
4. Connector runtime reads from engine/hook/native implementation.
5. Records flow to local warehouse/output; cursor/state advances.

### Direct-Read / Binary Flow

1. Inventory classifies an operation as non-durable direct-read or binary transfer.
2. Product-safe operations get typed CLI/native surfaces; risky operations get typed exclusions or human gates.
3. Execution uses bounded output, explicit paths, no traversal, and no broad raw HTTP/SQL/write tools.
4. Conformance and docs record behavior separately from ETL streams.

### Reverse ETL Write Flow

1. User creates a reverse plan from warehouse table to connector/action mapping.
2. Preview validates mapping and records.
3. Approval token gates execution.
4. Writer executes product-specific write action.
5. Result/ledger/status is stored; destructive/admin paths require extra human gates.

### Connector Certification Flow

1. `pm connectors certify` stages inspect metadata, credentials, catalog, source reads, write pairings, replay, flow, schedule, redaction, and report outputs.
2. Replay/fixture gates run without secrets.
3. Live checks require explicit credentials and treat missing credentials as uncertified.
4. Parity claims derive from generated certification/conformance outputs, not manual assertions.

### Runtime / RLM / Pi Agent Flow

1. User or agent starts local runtime only when needed: `scripts/runtime.sh up`.
2. `pm runtime doctor --json` checks PostgreSQL, DragonflyDB, and Temporal endpoints.
3. Dependency-free RLM modes (`deterministic`, `fixture`) run locally without runtime services.
4. RLM `agent` mode uses `pm agent image ensure`, Temporal workflows, and `pm worker serve/status`.
5. Runtime-backed verification remains optional and is gated by `POLYMETRICS_INTEGRATION=1`.

### GSD/Pi Agent Flow

1. Agent reads `AGENTS.md` and issue contract.
2. Agent runs `scripts/gsd doctor` or `/gsd doctor` when needed.
3. Agent generates the relevant official command prompt, e.g. `scripts/gsd prompt plan-phase 1 --skip-research` or `/gsd plan-phase 1 --skip-research`.
4. Agent reads required references for the task, including `runtime-rlm-website-integration.md` for runtime/RLM/Pi-agent/website architecture work.
5. Agent updates required planning/TDD/verification artifacts before production edits.
6. Agent verifies and records GSD evidence in planning traces or PR body.

## Connector Surface Model

Planning must model a connector as a collection of surfaces, not as a REST-only API:

- ETL streams for durable record collections.
- Reverse ETL write actions for product-safe mutations.
- Direct-read commands for useful reads that are not durable streams.
- Binary transfer commands for archives, artifacts, documents, files, exports, and media.
- Native protocol capabilities for SQL/database/CDC/queue/file systems.
- GraphQL, XML/SOAP, CSV/NDJSON, webhook/event/audit-log and other protocol-specific surfaces.
- Typed exclusions for destructive, elevated-scope, deprecated, duplicate, non-data, or out-of-scope operations.

## De-duplication Rule

A documented upstream operation must have exactly one primary classification. Duplicate docs pages, aliases, generated API references, and product guide variants should point to the same canonical operation identity instead of creating duplicate work.

## Architectural Guardrails

- No secrets in prompts, logs, artifacts, PR bodies, or generated docs.
- No new dependencies without human approval.
- No credentialed connector checks during planning-only work.
- No reverse ETL execution without plan, preview, approval, execute.
- No generic shell, generic HTTP write, or generic SQL write tools.
- No `cmd/` or `internal/` edits for issue #122 planning-only work.
- No `.planning/phases/**` regeneration in this refresh by explicit user request.

---
*Architecture analysis refreshed: 2026-07-08 via repo-local official GSD Core Pi adapter.*
