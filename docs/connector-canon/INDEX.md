# Connector Delivery Canon

**Status:** current and binding as of 2026-08-10. Read this page before
planning, implementing, describing, or certifying a connector.

This index separates current decisions from retained historical material. A
historical file is evidence, not a grant of capability. When it conflicts with
this index or a later captain ruling, the current canon wins.

## Read in this order

1. [Connector terminology and lane contract](connector-terminology.md) — the
   single owner of connector vocabulary, the seven lanes, source-to-mapping
   boundaries, completion states, and no-miss rules.
2. [Implementation procedure](IMPLEMENTATION-PROCEDURE.md) — the required
   end-to-end path and its Foundation Check.
3. [Remote reproducibility](REMOTE-REPRODUCIBILITY.md) — what a clean machine
   can prove without provider access, and what it cannot.
4. [Operation evidence](OPERATION-EVIDENCE.md) — generated source-operation
   accounting and the fixed-100 executable regression gate.
5. [Source-declaration admission](DECLARATION-ADMISSION.md) — the distinct,
   provider-I/O-free completeness certificate for cited declaration rows.
6. The mechanical authoring references:
   [migration conventions](../migration/conventions.md),
   [architecture v2 design](../architecture/connector-architecture-v2-design.md),
   and [`AGENTS.md`](../../AGENTS.md).

## Binding source material

| Source | What it decides |
| --- | --- |
| [captain rulings](../../data/captain.md) + [current corrections](../../data/CURRENT-CORRECTIONS.md) | The warehouse is always the mediator; source-to-destination hops are forbidden, including API → API and a connector to itself. Certification is all-or-nothing and cannot be granted by adding a file. The correction companion prevents preserved historical measurements from being reused as current counts. |
| [database connector framework](../../data/cli-database-connector-framework-design-r1/report.md) | Typed database framework; PostgreSQL is the reference driver; no generic SQL executor, second sync-mode enum, or separate repository. |
| [large-transaction CDC strategy](../../data/cli-cdc-large-transaction-strategy-r1/report.md) | PostgreSQL 14+ streamed `pgoutput` v2; bounded durable stage; receipt before acknowledgement; fail-closed quota; cursor fallback is rejected for CDC. |
| [bidirectional changefeed design](../../data/cli-cdc-bidirectional-changefeed-design-r1/report.md) | One delivery contract with two honestly different producers; explicit tombstones only. |
| [PostgreSQL parity issue tree r2](../../data/cli-postgres-parity-issue-tree-r2/report.md) | Parent issue #3972 and its **12** sub-issues, including #3987: the warehouse-only four-flow/seven-mode conformance gate. The source-pinned [r1 baseline](../../data/archive/cli-postgres-parity-issue-tree-r1/report.md) is archived because its 11-child dependency graph is no longer current. |
| [daily-use Top 50 targets](../../data/cli-daily-use-top50-connectors-r1/report.md) | The 50 certification targets and the current quarantine evidence. |

The imported reports are source-pinned in
[`data/CANON-MANIFEST.sha256`](../../data/CANON-MANIFEST.sha256). The
`connector-canon-check` verification target checks those pins and the required
canon entry points.

## Current, honest status

- The accepted live-certification count is **zero**. A fixture, a generated
  manual, an `api_surface.json`, or a `certification.json` contract is not live
  proof and does not make a connector certified.
- The accepted quarantine list has **15 entries**. Do not repeat the historical
  “195 blocked providers” claim.
- PostgreSQL parity is tracked by #3972 and its 12 child issues. #3987 now
  gates certification on all four warehouse-mediated routes and all seven mode
  outcomes; it does not alter active #3974. GitHub source-inventory counts are
  generated from its merged source lock; the archived gap map remains void.
- A capability is only implemented when its real runtime preflight and the
  required flow foundations execute. A declaration is not a capability.

`data/captain.md` retains earlier historical prose for auditability. Its
[correction companion](../../data/CURRENT-CORRECTIONS.md) identifies the
historical 195 and 17 measurements that must not be reused; the accepted current
values above control.

## Retired and actively wrong material

Nothing below was deleted. Archive copies preserve the reasoning that led to a
decision, but must not be used to establish current state.

| Material | Status and reason |
| --- | --- |
| [GitHub ETL/reverse-ETL gap map](../../data/archive/cli-github-etl-reverse-etl-gap-map-r1/report.md) | **SUPERSEDED — ACTIVELY WRONG for every coverage number.** It measured `main` (341 operations) rather than the GitHub parity branch (768). Its architecture observations are historical background only; do not cite its counts. |
| [blocked-source recovery brief](../../data/archive/cli-blocked-source-recovery-tiers-r1/brief.md) | **SUPERSEDED — ACTIVELY WRONG.** Its “195 genuinely blocked” claim is unsubstantiated and contradicted by the accepted 15-entry quarantine list. |
| [PostgreSQL parity tree r1](../../data/archive/cli-postgres-parity-issue-tree-r1/report.md) | **SUPERSEDED FOR EXECUTION SCOPE.** Its architecture diagnosis is retained, but its 11-child graph omitted #3987 and allowed certification before the warehouse-mediated four-flow/seven-mode proof. Use r2. |
| [superseded repository planning](archive/superseded-repository-planning/) | July 2026 orchestration, project, roadmap, and research snapshots. They rely on stale inventories, old orchestration roles, and noncanonical commands. Their original entry points now point here. |
| [migration handoff](archive/superseded-repository-planning/migration-handoff-codex-2026-07-04.md) | **SUPERSEDED.** It instructed workers to use a stale parallel rollout, obsolete PR state, and stale inventory/quarantine figures. The original handoff now points to this procedure. |
| [superseded GSD material](archive/superseded-gsd/) | The old universal-programming-loop procedure and prompts. They prescribe an absent `programming-loop` command and role spawning, both contrary to the canonical delivery contract. |
| `.planning/phases/**` other than the issue being delivered | Historical execution evidence. It can explain a past decision but cannot override the binding sources or this current procedure. |

## What this index deliberately does not do

It does not certify any connector, infer provider coverage from filenames, or
merge claims from a live parity lane. Those require the procedure's Foundation
Check, flow proof, and accepted live evidence.
