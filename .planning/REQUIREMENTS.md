# Requirements: Polymetrics CLI Connector Parity

**Defined:** 2026-07-08
**Core Value:** Users and agents can trust `pm` as a connector-complete, safety-gated ETL and reverse ETL interface whose advertised connector capabilities match documented upstream product surfaces without duplicate or unsafe exposure.

## v1 Requirements

### GSD Rebootstrap

- [ ] **GSD-01**: Pre-rebootstrap `.planning/` state is archived outside active `.planning/` before replacement.
- [ ] **GSD-02**: Active `.planning/` is recreated in upstream GSD Core structure with tracked `PROJECT.md`, `REQUIREMENTS.md`, `ROADMAP.md`, `STATE.md`, config, codebase maps, and phase artifacts.
- [ ] **GSD-03**: New planning artifacts identify the active milestone as connector parity across all connector technologies, not only REST APIs.
- [ ] **GSD-04**: Verification records the commands/workflows used and confirms no Go source files changed.

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

## Out of Scope

| Feature | Reason |
|---------|--------|
| Runtime connector implementation in issue #122 | This issue is planning rebootstrap only. |
| New dependencies | Requires human gate and is unnecessary for planning. |
| Credentialed connector checks | No secrets are needed; live checks are separate certification work. |
| Destructive/admin execution | Must remain human-gated. |
| Generic shell/HTTP write/SQL write tools | Violates project safety rules. |

## Traceability

| Requirement | Phase | Status |
|-------------|-------|--------|
| GSD-01 | Phase 1 | In Progress |
| GSD-02 | Phase 1 | In Progress |
| GSD-03 | Phase 1 | In Progress |
| GSD-04 | Phase 1 | In Progress |
| INV-01 | Phase 1 | Pending |
| INV-02 | Phase 1 | Pending |
| INV-03 | Phase 1 | Pending |
| INV-04 | Phase 1 | Pending |
| INV-05 | Phase 1 | Pending |
| READ-01 | Phase 2 | Pending |
| READ-02 | Phase 3 | Pending |
| READ-03 | Phase 3 | Pending |
| READ-04 | Phase 3 | Pending |
| WRITE-01 | Phase 4 | Pending |
| WRITE-02 | Phase 4 | Pending |
| WRITE-03 | Phase 4 | Pending |
| GATE-01 | Phase 5 | Pending |
| GATE-02 | Phase 5 | Pending |
| GATE-03 | Phase 1 | Pending |

**Coverage:**
- v1 requirements: 19 total
- Mapped to phases: 19
- Unmapped: 0

---
*Requirements defined: 2026-07-08*
*Last updated: 2026-07-08 after issue #122 upstream GSD Core rebootstrap*
