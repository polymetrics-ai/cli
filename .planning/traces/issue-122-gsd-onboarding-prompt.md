# Issue #122 GSD Core Onboarding Prompt

**Purpose:** Prompt used to run the upstream GSD Core brownfield onboarding workflow for connector parity.

**Upstream command/workflow sources:**

- `/Users/karthiksivadas/.claude/commands/gsd/map-codebase.md`
- `/Users/karthiksivadas/.claude/get-shit-done/workflows/map-codebase.md`
- `/Users/karthiksivadas/.claude/commands/gsd/new-project.md`
- `/Users/karthiksivadas/.claude/get-shit-done/workflows/new-project.md`
- `/Users/karthiksivadas/.claude/commands/gsd/plan-phase.md`
- `/Users/karthiksivadas/.claude/get-shit-done/workflows/plan-phase.md`
- `/Users/karthiksivadas/.claude/commands/gsd/programming-loop.md`

## Prompt

Run upstream GSD Core brownfield onboarding for the Polymetrics CLI repository after archiving and replacing the legacy/custom `.planning/` tree.

### Required source context

Read and preserve the constraints from:

- `AGENTS.md`
- `.agents/agentic-delivery/contracts/issue-agent-contract.md`
- `docs/migration/HANDOFF-CODEX.md`
- `docs/migration/conventions.md`
- `docs/architecture/connector-architecture-v2-design.md`
- `docs/architecture/connector-certification-design.md`
- `docs/plans/universal-programming-loop-prd.md`
- GitHub issue #122: `chore(gsd): rebootstrap upstream GSD Core planning for connector parity`

### Brownfield codebase mapping scope

Map the current codebase before requirements/roadmap creation. The map must cover:

- Go CLI entry points and command surfaces (`cmd/pm`, `internal/cli`)
- ETL, reverse ETL, scheduling, query, runtime, vault/credentials, safety, and connector packages
- Connector Architecture v2: defs bundles, engine, conformance, certify, hooks, natives, codegen
- GitHub Actions, Makefile gates, and docs generation
- Current connector inventory signals generated from the working tree, not stale prior `.planning/`

### Connector parity north star

Seed the project around connector parity, not only REST API parity. The GSD artifacts must treat every connector as a system surface that can include multiple technologies and protocols:

- REST/HTTP JSON APIs
- GraphQL queries/mutations/subscriptions where the upstream product documents them
- XML/SOAP APIs and XML feeds
- CSV, TSV, NDJSON, and report-export endpoints
- File/object storage sources and sinks (local files, S3-like APIs, signed downloads, uploads)
- SQL/database connectors and CDC surfaces
- Queues and event streams (SQS-like, webhook/event resources, audit logs)
- Binary transfer surfaces (artifacts, archives, attachments, documents, media)
- Direct-read command surfaces that are useful but not durable ETL streams
- Admin, destructive, elevated-scope, credential-management, and dependency-gated operations

### De-duplication policy

Do not double-count one upstream capability in multiple surfaces. For every documented operation, choose exactly one primary classification:

1. ETL stream
2. Reverse ETL write action
3. Direct-read CLI command
4. Binary transfer command
5. Native/database/CDC/queue/file capability
6. Webhook/event/audit-log surface
7. Typed exclusion with a closed-vocabulary reason

If an operation appears in generated API docs and a product-specific guide, reconcile it by canonical upstream operation identity: method/protocol + normalized path/resource + operation id/name + product scope. Record aliases as references, not duplicate work.

### Required roadmap ordering

The roadmap must start with inventory reconciliation before fanout. Inventory must reconcile bundle counts, hook/native counts, `api_surface.json`, schema/stream/write declarations, non-REST protocols, docs, conformance, certification, direct-read/binary metadata, blockers, and quarantine state.

### Safety and gates

Preserve Polymetrics safety overlays:

- No secrets are requested, printed, stored, or summarized.
- No credentialed connector checks in this issue.
- Reverse ETL remains plan → preview → approval → execute.
- Do not expose generic shell, generic HTTP write, or generic SQL write tools.
- Destructive/admin/elevated-scope/dependency changes are human-gated.
- Do not change `cmd/` or `internal/` in this planning-only issue.
- Do not merge to `main` without human approval.

### Expected active planning outputs

Create upstream GSD Core style active planning artifacts:

- `.planning/config.json`
- `.planning/PROJECT.md`
- `.planning/REQUIREMENTS.md`
- `.planning/ROADMAP.md`
- `.planning/STATE.md`
- `.planning/codebase/STACK.md`
- `.planning/codebase/INTEGRATIONS.md`
- `.planning/codebase/ARCHITECTURE.md`
- `.planning/codebase/STRUCTURE.md`
- `.planning/codebase/CONVENTIONS.md`
- `.planning/codebase/TESTING.md`
- `.planning/codebase/CONCERNS.md`
- `.planning/phases/01-inventory-reconciliation/*`
- Verification artifacts recording command/workflow sources and no Go source changes.
