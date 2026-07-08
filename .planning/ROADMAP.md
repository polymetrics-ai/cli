# Roadmap: Polymetrics CLI Connector Parity

## Overview

This roadmap replaces the legacy custom `.planning/` tree with an upstream GSD Core brownfield plan for connector parity. The first phase is deliberately inventory reconciliation: no connector fanout starts until the repository and upstream documentation are reconciled across all connector technologies, with canonical operation de-duplication and human-gated risk classification.

## Phases

**Phase Numbering:** Integer phases are planned milestone work; inserted decimal phases are reserved for urgent gate fixes.

- [ ] **Phase 1: Inventory and Surface Reconciliation** - Reconcile connector inventory and all documented surfaces before any connector fanout.
- [ ] **Phase 2: Durable Read and ETL Parity** - Ensure product-safe documented record collections are available through `pm` catalog/read/ETL or typed exclusions.
- [ ] **Phase 3: Direct-Read, Binary, and Native Surface Parity** - Cover product-safe direct-read, binary transfer, file/object, SQL/CDC, queue/event, GraphQL, XML/SOAP, CSV/NDJSON, and native surfaces without misclassification.
- [ ] **Phase 4: Reverse ETL and Mutation Parity** - Map safe mutations/writes across connector technologies to reverse ETL plan/preview/approval/run or typed exclusions.
- [ ] **Phase 5: Conformance and Certification Enforcement** - Make conformance/certification outputs authoritative for connector parity claims and rollout readiness.

## Phase Details

### Phase 1: Inventory and Surface Reconciliation
**Goal**: Produce a current, generated, de-duplicated connector parity baseline before fanout.
**Depends on**: Nothing (first phase)
**Requirements**: [GSD-01, GSD-02, GSD-03, GSD-04, INV-01, INV-02, INV-03, INV-04, INV-05, GATE-03]
**Success Criteria** (what must be TRUE):
  1. Active `.planning/` is upstream GSD Core shaped and the pre-rebootstrap archive is recorded outside active planning.
  2. Generated inventory reconciles bundle, hook, native, docs, API/surface manifest, stream, write, binary, direct-read, native protocol, blocker, quarantine, conformance, and certification counts.
  3. The inventory classifies REST, GraphQL, XML/SOAP, CSV/NDJSON, binary, file/object, SQL/CDC, queue/event/webhook/audit-log, native, direct-read, and mutation surfaces.
  4. The inventory applies canonical operation identity to avoid duplicate work from multiple docs pages/specs.
  5. No connector fanout work is dispatched until inventory outputs are reviewed.
**Plans**: 3 plans

Plans:
- [ ] 01-01: Rebootstrap planning and preserve GSD/archive evidence.
- [ ] 01-02: Generate current connector inventory and multi-technology surface classification.
- [ ] 01-03: Review inventory, de-duplication, blockers, and approve/disallow fanout entry.

### Phase 2: Durable Read and ETL Parity
**Goal**: Bring product-safe documented durable record collections to CLI and ETL parity.
**Depends on**: Phase 1
**Requirements**: [READ-01]
**Success Criteria** (what must be TRUE):
  1. Every product-safe documented stream/report/event log/feed/table/queue message/CDC event/durable record collection is covered by a connector read surface or typed exclusion.
  2. `pm connectors inspect`, catalog, read, and ETL surfaces agree on stream names, schemas, sync modes, cursor behavior, and limitations.
  3. Incremental, pagination, cursor, schema, and projection behavior is verified through conformance fixtures or typed blockers.
**Plans**: TBD after Phase 1 inventory reconciliation

Plans:
- [ ] 02-01: Prioritize durable read parity gaps from Phase 1 inventory.
- [ ] 02-02: Execute read parity slices with conformance-backed verification.

### Phase 3: Direct-Read, Binary, and Native Surface Parity
**Goal**: Cover product-safe non-stream and non-REST read surfaces without forcing them into the wrong abstraction.
**Depends on**: Phase 1
**Requirements**: [READ-02, READ-03, READ-04]
**Success Criteria** (what must be TRUE):
  1. Direct-read operations are classified separately from durable ETL streams.
  2. Binary operations have product-safe transfer surfaces or typed exclusions.
  3. GraphQL, XML/SOAP, CSV/NDJSON/report export, file/object, SQL/CDC, queue/event/webhook/audit-log, and native protocol reads use an appropriate declarative, hook, native, direct-read, binary, or exclusion path.
  4. Admin/elevated/destructive direct-read or binary operations remain human-gated.
**Plans**: TBD after Phase 1 inventory reconciliation

Plans:
- [ ] 03-01: Classify direct-read, binary, native, and protocol-specific surfaces.
- [ ] 03-02: Plan safe CLI/native coverage and typed exclusions.

### Phase 4: Reverse ETL and Mutation Parity
**Goal**: Map safe writes/mutations to reverse ETL actions with approval gates.
**Depends on**: Phase 1
**Requirements**: [WRITE-01, WRITE-02, WRITE-03]
**Success Criteria** (what must be TRUE):
  1. Product-safe mutations across REST, GraphQL, XML/SOAP, file/object, queue, database/native, and other protocol-specific operations map to reverse ETL actions or typed exclusions.
  2. Plan, preview, approval, and execute semantics are preserved for every write path.
  3. Destructive/admin/elevated-scope writes are human-gated and never exposed as generic raw write tools.
**Plans**: TBD after Phase 1 inventory reconciliation

Plans:
- [ ] 04-01: Prioritize safe reverse ETL mutation gaps from Phase 1 inventory.
- [ ] 04-02: Execute write parity slices with preview, approval, and certification evidence.

### Phase 5: Conformance and Certification Enforcement
**Goal**: Make validated gates authoritative for connector parity status.
**Depends on**: Phases 2-4
**Requirements**: [GATE-01, GATE-02, GATE-03]
**Success Criteria** (what must be TRUE):
  1. Conformance validates schemas, fixtures, surface manifests, streams, writes, direct-read/binary metadata, cursor/pagination behavior, docs, and de-duplication contracts.
  2. Certification reports distinguish replay/fixture success, live success, missing credentials (`uncertified`), typed blockers, and failures.
  3. Public or PR-facing connector parity claims derive from generated conformance/certification artifacts.
**Plans**: TBD after Phases 2-4

Plans:
- [ ] 05-01: Reconcile conformance gate coverage.
- [ ] 05-02: Reconcile certification gate coverage and reporting.

## Progress

**Execution Order:** Phase 1 → Phase 2 → Phase 3 → Phase 4 → Phase 5. Phase 1 is a hard planning gate: no connector fanout before inventory and surface reconciliation is reviewed.

| Phase | Plans Complete | Status | Completed |
|-------|----------------|--------|-----------|
| 1. Inventory and Surface Reconciliation | 0/3 | In progress | - |
| 2. Durable Read and ETL Parity | 0/TBD | Not started | - |
| 3. Direct-Read, Binary, and Native Surface Parity | 0/TBD | Not started | - |
| 4. Reverse ETL and Mutation Parity | 0/TBD | Not started | - |
| 5. Conformance and Certification Enforcement | 0/TBD | Not started | - |

---
*Roadmap generated: 2026-07-08 via upstream GSD Core brownfield workflow shape for issue #122*
