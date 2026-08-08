# PersistIQ artifact materialization pilot - Context

**Gathered:** 2026-08-08
**Status:** In execution under captain ruling

<domain>
## Phase Boundary

Run the PersistIQ end-to-end pilot, then validate the generator capability with
staged, evidence-only generalization shapes. The later captain extension adds
Watchmode, DocuSeal, Float, and an optional Copper Postman fallback proof; no
generated pilot bundle enters `internal/connectors/defs`. Do not start the
eligible-392 production pool. The 28 connectors already live in main and the
nine elsewhere-owned connectors remain outside this phase.

</domain>

<decisions>
## Implementation Decisions

### Pilot scope and order

- **D-01:** PersistIQ is the only connector in this phase. The exact five-step
  order is fixed: identify link; map every operation; fetch/parse and record
  bytes plus SHA-256; materialize and run static/reachability gates; report
  timings and totals.
- **D-02:** No credentials, provider data, or credentialed/live connector
  execution may be used. Network access is limited to fetching the public
  specification, with a single polite request.
- **D-03:** Use the existing `connectorgen batch plan`, `batch materialize`,
  `batch gate`, `connectorgen validate`, and `surface-sync --check` tooling.
  Do not add a generator or weaken a gate. Stage source and generated bundles
  inside this phase directory before installing any generated output into the
  connector definition tree.

### Operation-model mapping

- **D-04:** The pre-fetch mapping for the ledger's 21 operations is 11 ETL
  collection reads, one blocked direct read for `GET /v1/leads/{id}`, seven
  reverse-ETL mutations, and two direct writes for the wrapper-shaped
  `POST /v1/leads` and webhook-setting `PUT /v1/webhook_plugin`. Binary
  download and unclassified are both zero. Reconcile this map against the
  fetched artifact and report any drift instead of silently changing the
  count.

### Certification and reporting

- **D-05:** Certification is withheld. The final report must say implemented
  (if the static gates succeed), not certified, and never exercised against the
  provider. If a gate fails, report the exact failed stage and do not claim
  materialized or reachable success.
- **D-06:** Every command's reachability check is a real built `pm` binary
  invocation using help/no-network arguments; exit status alone is not enough.
  Record unknown-command and other failures by command.
- **D-07:** The captain's rerun policy supersedes the earlier fail-closed
  coverage decision: every operation documented by the fetched artifact is
  represented in the generated bundle, even when no executor exists yet.
  Unsupported operations become `not_implemented` commands with a
  machine-checkable `named_dependency=<slug>` note; they are never silently
  omitted or marked implemented.
- **D-08:** Existing source-surface operations absent from the fetched artifact
  remain in the materialized surface and carry the exact discrepancy marker
  `present-in-surface-absent-from-artifact`. PersistIQ's `/v1/mailboxes`,
  `/v1/activities`, and `/v1/accounts` are the required pilot cases.
- **D-09:** The rerun reports mapped, implemented, named-dependency,
  flagged-discrepancy, and reachable counts separately, with wall-clock times
  for each locked step. The 392-connector pool remains deferred until the
  captain reviews this pilot.
- **D-10:** Batch materialization is authoring-only and never runs the
  repository-wide gate per candidate. Fetch is bounded/concurrent, mapping and
  materialization run across staged batches, and one final gate scans the
  complete staged result. Batch boundaries remain commit/review boundaries;
  only a failed final gate triggers narrowed diagnostic gating.
- **D-11:** The multi-source contract is authoritative for generator
  validation: try OpenAPI/Swagger and bounded local/remote refs/webhooks,
  provider machine exports such as Postman, then bounded official
  HTML/Markdown/source traversal. Normalize every operation with source URL,
  kind, version, retrieval date, SHA-256, exact coordinate, and preserved
  alternatives. A narrower artifact cannot delete an existing operation, and
  ambiguous extraction remains visible as an unknown/disposition gap.
- **D-12:** The three required generalization pilots must pass after webhook
  and external-reference support. Copper is additional static fallback
  evidence; its legacy native scaffold has no generated command surface, so
  reachability is not claimed for it.

### the agent's Discretion

- Temporary staging paths and the exact parser invocation are implementation
  details, provided they stay inside this worktree and the existing batch
  parser remains authoritative for materialization.

</decisions>

<canonical_refs>
## Canonical References

**Downstream agents MUST read these before planning or implementing.**

### Task and lifecycle contract

- `/Users/karthiksivadas/karthik-agent-workspace/data/cli-mass-artifact-materialize-r1/CAPTAIN-ORDER-mass-materialize.md` — authoritative correction, pilot order, gates, and certification boundary.
- `/Users/karthiksivadas/karthik-agent-workspace/data/cli-mass-artifact-materialize-r1/CAPTAIN-ORDER-multisource-mapping.md` — authoritative source order, provenance, fallback, and traversal contract.
- `AGENTS.md` — mandatory issue-first GSD lifecycle, connector-surface rules, and verification commands.
- `.agents/agentic-delivery/references/required-skills-routing.md` — required Go and connector skills.
- `.agents/agentic-delivery/references/gsd-pi-adapter.md` — repo-local GSD command resolution and manual fallback rules.

### Connector authoring and batch implementation

- `docs/migration/HANDOFF-CODEX.md` — connector architecture v2 collision and delivery rules.
- `docs/migration/conventions.md` — declarative bundle and operation-surface conventions.
- `docs/architecture/connector-architecture-v2-design.md` — bundle/runtime architecture.
- `cmd/connectorgen/batch.go` — manifest planning and batch gate behavior.
- `cmd/connectorgen/batch_materialize.go` — artifact fetch/cache, OpenAPI/Swagger parsing, provenance, and materialization.
- `internal/connectors/defs/persistiq/` — existing PersistIQ source bundle used as the pilot input.
- `.agents/agentic-delivery/references/cli-help-docs-website-parity.md` — command-surface and binary-help parity checklist.

### Prior art

- `/Users/karthiksivadas/karthik-agent-workspace/data/cli-top50-fixed-schema-sweep-r1/tools/` — prior fixed-schema planning and red/green evidence tooling.

</canonical_refs>

<code_context>
## Existing Code Insights

### Reusable Assets

- `connectorgen batch` — existing single-candidate manifest, materialization,
  provenance, and runtime-preflight orchestration.
- `internal/connectors/commandrunner.Preflight` — authoritative implemented
  command admission check used by batch gate.
- `internal/connectors/defs/persistiq` — source bundle with existing stream,
  write, schema, and API-surface declarations.

### Established Patterns

- Batch materialization writes to a new destination and refuses collisions;
  stage source and destination separately so a failed pilot leaves the source
  intact.
- Provider artifact rows must have provenance or an explicit blocked/excluded
  disposition; an invented endpoint must fail static/reachability gates.
- Reverse ETL remains plan → preview → approval → execute; reachability tests
  must not execute provider writes.

### Integration Points

- The generated bundle is consumed by `defs.FS`, `connectorgen validate`,
  `surface-sync`, commandrunner preflight, and the real `pm` binary.

</code_context>

<specifics>
## Specific Ideas

PersistIQ was selected because its ledger identifies a direct Swagger JSON
artifact with 21 operations (12 read / 9 write), making it a bounded pilot that
exercises both read and write mapping without provider credentials.

</specifics>

<deferred>
## Deferred Ideas

- Bulk fetching/materialization of the eligible 392 connectors.
- Feasibility planning for the 99 unresolved connectors, including AI prose
  extraction and its mandatory validation/reachability failure behavior.
- Production generation of the eligible 392 and the seven-connector consolidation until PR #3957 merges.

</deferred>

---

*Phase: persistiq-artifact-materialize-pilot-r1*
*Context gathered: 2026-08-08 via autonomous manual-GSD fallback*
