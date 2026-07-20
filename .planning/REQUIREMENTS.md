# Requirements: Polymetrics CLI Connector Parity

**Defined:** 2026-07-08
**Refreshed via:** official GSD Core Pi adapter commands
**Core Value:** Users and agents can trust `pm` as a connector-complete, safety-gated ETL and reverse ETL interface whose advertised connector capabilities match documented upstream product surfaces without duplicate or unsafe exposure.

## v1 Requirements

### GSD and Pi Runtime

- [ ] **GSD-01**: Pre-rebootstrap `.planning/` state is archived outside active `.planning/` before replacement.
- [ ] **GSD-02**: Active `.planning/` is recreated in official GSD Core structure with tracked `PROJECT.md`, `REQUIREMENTS.md`, `ROADMAP.md`, `STATE.md`, config, codebase maps, and phase artifacts.
- [ ] **GSD-03**: New planning artifacts identify the active milestone as connector parity across all connector technologies, not only REST APIs.
- [ ] **GSD-04**: Verification records the commands/workflows used and confirms no Go source files changed.
- [ ] **GSD-05**: Official `open-gsd/gsd-core@next` source is pinned in `.gsd/upstream.lock.json` and command registry is generated from official docs.
- [ ] **GSD-06**: Pi resources expose repo-local GSD commands through `/gsd`, generated `/gsd-*` aliases, prompt fallback, and GSD Core skill defaults.
- [ ] **GSD-07**: Agents and subagents route GSD work through `.pi` commands or `scripts/gsd prompt`, with manual fallback recorded only when the adapter is unavailable.
- [ ] **GSD-08**: Agents and subagents load required Go/design skills from `.agents/agentic-delivery/references/required-skills-routing.md` and record skill evidence in GSD plans, handoffs, or PR bodies.

### Runtime, RLM, Pi Agent, and Website Integration Knowledge

- [ ] **RUNTIME-01**: Planning and agent guidance preserves canonical knowledge for Podman-first local runtime, Docker fallback, PostgreSQL, DragonflyDB/Redis-compatible coordination, Temporal, RLM agent mode, `pm runtime`, `pm rlm`, `pm agent image`, and `pm worker`.
- [ ] **RUNTIME-02**: Runtime-backed checks remain optional and explicitly gated; default unit tests and dependency-free CLI paths must not require PostgreSQL, DragonflyDB, Temporal, or Podman.
- [ ] **RUNTIME-03**: Runtime/RLM/Pi-agent work follows `.agents/agentic-delivery/references/runtime-rlm-website-integration.md` and updates docs/website parity when user-facing behavior changes.
- [ ] **WEBSITE-01**: Website architecture knowledge records Next.js 16, React 19, Fumadocs, generated docs/data scripts, and relevant website checks without adding dependencies.

### CLI Help, Manual, Docs, and Website Parity

- [ ] **CLI-DOC-01**: Every CLI command, subcommand, flag, output, connector surface, or help-topic change updates runtime help, `docs/cli/**`, website docs under `website/**`, generated help/manual artifacts, and tests together or records explicit not-applicable notes.
- [ ] **CLI-DOC-02**: Namespace command groups with no action selected, such as `pm connectors`,
      render contextual help/subcommand summaries and exit successfully, while invalid actions
      still return usage errors. Eligible dual-TTY bare `pm query` and bare `pm reverse` are the
      documented human-first workspace exceptions and fall back to deterministic help on every
      bypass/non-TTY path.
- [ ] **CLI-DOC-03**: Implementation PRs for CLI changes include parity evidence for `pm help <topic>`, `pm <namespace>`, `pm <command> --help`, docs/website search or generator checks, and any golden/help fixture updates.
- [ ] **CLI-DOC-04**: Agents and subagents follow `.agents/agentic-delivery/references/cli-help-docs-website-parity.md` for CLI-visible changes.

### Inventory and Canonical Surface Baseline

- [ ] **INV-01**: The first execution phase reconciles connector inventory across `internal/connectors/defs/`, hooks, natives, docs, `api_surface.json`, conformance, certification, CLI metadata, direct-read, binary, and blocker/quarantine state before connector fanout.
- [ ] **INV-02**: Inventory classifies each connector as fully parity-covered, partial with typed blockers, quarantined, or requiring direct-read/binary/admin/human-gated handling.
- [ ] **INV-03**: Connector counts and capability summaries are generated from the current repository and upstream docs, not stale legacy planning artifacts.
- [ ] **INV-04**: Inventory includes non-REST/protocol surfaces: GraphQL, XML/SOAP, CSV/TSV/NDJSON/report exports, binary transfers, file/object storage, SQL/database/CDC, queues, webhooks/events/audit logs, and native protocols.
- [ ] **INV-05**: Inventory applies a de-duplication policy so one upstream operation has exactly one primary classification and aliases are recorded as references.

### Read, Direct-Read, Binary, and Native Parity

- [ ] **READ-01**: Every product-safe documented stream, report, event log, feed, table, queue message, CDC event, or other durable record collection is available through `pm` catalog/read/ETL or explicitly excluded with a typed reason.
- [ ] **READ-02**: Direct-read operations that are useful but not durable ETL streams are represented as safe CLI surfaces or typed exclusions.
- [ ] **READ-03**: Binary transfer operations such as artifacts, archives, attachments, documents, exports, and media are represented as binary transfer capabilities or typed exclusions.
- [ ] **READ-04**: GraphQL, XML/SOAP, CSV/NDJSON, file/object storage, SQL/CDC, queue, webhook/event, and native-protocol reads are planned using the correct runtime tier rather than forced into REST stream shapes.

### Reverse ETL and Mutation Parity

- [ ] **WRITE-01**: Every product-safe documented mutation/write action across REST, GraphQL mutations, XML/SOAP actions, file/object writes, queue sends, database writes, and other protocol-specific operations maps to a reverse ETL action or typed exclusion.
- [ ] **WRITE-02**: Reverse ETL remains plan → preview → approval → execute; no raw generic write surfaces are introduced.
- [ ] **WRITE-03**: Destructive, admin, elevated-scope, credential-management, billing, dependency-sensitive, and irreversible write paths remain human-gated or typed exclusions.

### Conformance and Certification

- [ ] **GATE-01**: Conformance verifies schema, fixture, API/surface manifest, stream, write, pagination, cursor, binary/direct-read metadata, docs, and de-duplication contracts for parity claims.
- [ ] **GATE-02**: Certification verifies CLI-level connector behavior without requiring secrets for replay/fixture gates and treats missing live credentials as uncertified, not failed.
- [ ] **GATE-03**: Planning prevents connector fanout or rollout claims until inventory and gate outputs are reconciled.

## v2 Requirements

### Long-Term Connector Governance

- **GOV-01**: Generate durable public connector capability reports from validated manifests and certification outputs.
- **GOV-02**: Add recurring drift detection for upstream product documentation across REST, GraphQL, XML/SOAP, binary, file, database, queue, webhook/event, and other connector surfaces.
- **GOV-03**: Add richer cost/timing reports for future parallel connector migration waves.
- **GOV-04**: Keep GSD/Pi adapter compatibility checked when upstream GSD Core command docs change.

## Out of Scope

| Feature | Reason |
|---|---|
| Runtime connector implementation in issue #122 | This issue is planning and agent/runtime rebootstrap only. |
| Phase regeneration in this refresh | User explicitly requested updating everything except phases. |
| New dependencies | Requires human gate and is unnecessary for planning. |
| Credentialed connector checks | No secrets are needed; live checks are separate certification work. |
| Destructive/admin execution | Must remain human-gated. |
| Generic shell/HTTP write/SQL write tools | Violates project safety rules. |

## Traceability

| Requirement | Roadmap Workstream | Status |
|---|---|---|
| GSD-01 | 0/1 | In progress |
| GSD-02 | 0/1 | In progress |
| GSD-03 | 0/1 | In progress |
| GSD-04 | 0/1 | In progress |
| GSD-05 | 0 | In progress |
| GSD-06 | 0 | In progress |
| GSD-07 | 0 | In progress |
| GSD-08 | 0 | Pending |
| RUNTIME-01 | 0 | Pending |
| RUNTIME-02 | 0 | Pending |
| RUNTIME-03 | 0 | Pending |
| WEBSITE-01 | 0 | Pending |
| CLI-DOC-01 | 0/2/3/4 | Pending |
| CLI-DOC-02 | 0/2/3/4 | Pending |
| CLI-DOC-03 | 0/2/3/4 | Pending |
| CLI-DOC-04 | 0 | Pending |
| INV-01 | 1 | Pending |
| INV-02 | 1 | Pending |
| INV-03 | 1 | Pending |
| INV-04 | 1 | Pending |
| INV-05 | 1 | Pending |
| READ-01 | 2 | Pending |
| READ-02 | 3 | Pending |
| READ-03 | 3 | Pending |
| READ-04 | 3 | Pending |
| WRITE-01 | 4 | Pending |
| WRITE-02 | 4 | Pending |
| WRITE-03 | 4 | Pending |
| GATE-01 | 5 | Pending |
| GATE-02 | 5 | Pending |
| GATE-03 | 1 | Pending |

---
*Requirements refreshed: 2026-07-08 via repo-local official GSD Core Pi adapter; phases intentionally unchanged.*
