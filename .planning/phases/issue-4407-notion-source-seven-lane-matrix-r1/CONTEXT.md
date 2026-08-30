# Task Delivery Header

- **Issue:** #4407 — Notion source-to-seven-lane matrix
- **Branch:** `feat/4407-notion-track-a-matrix-r1`
- **Base:** `origin/main@813f457a925f7ee3fe3bea101a43e445992c8552`
- **Source snapshot:** `fm/cli-top100-declaration-batch-r1@dc481bac1a8b78d60ac0b4a2b0dfd1a9068ce8db`
- **Scope:** exact Notion source artifacts absent from main, connector-local matrix, local reconciliation test, and issue evidence only
- **Delivery boundary:** Track A mapping only; no shared runtime, generator, certification, CLI-surface, or merge work

## Objective

Account for every frozen Notion source operation across direct read, direct write, binary download, binary upload, ETL, reverse ETL, and sync transport without promoting a source fact to runtime proof.

## Fixed facts

- The retained source lock has 49 REST rows: 20 GET, 17 POST, 8 PATCH, and 4 DELETE.
- The source-bound crosswalk has 49 exact rows and two visible surface-only identities; neither boundary record enters the 49-row denominator.
- The lock embeds the provider OpenAPI contract and records `https://developers.notion.com/openapi.json`, SHA-256 `dee5763763b0b9fbad2aa8d5adb173ca350ec26dda557e658c5dbe9d2ea2f258`, and 1,304,814 provider bytes.
- The two source files are copied byte-identically from the Batch R1 snapshot, verified by Git blob identity before mapping.

## Decisions

- A locked GET is a direct-read source candidate only; it is not executable proof.
- The four source-semantic non-GET reads are `post-database-query`, `post-search`, `query-meeting-notes`, and `introspect-token`; HTTP method alone does not classify a mutation.
- The 25 remaining POST/PATCH/DELETE rows receive independent direct-write and reverse-ETL source cells.
- Twelve source-backed collection/read rows receive ETL cells: nine cursor-query GETs, two cursor-body POST queries, and `query-meeting-notes`, whose `results` plus `has_more` lacks a retained continuation input and remains a mapping restriction.
- `upload-file` is the sole multipart binary-upload candidate. No 2xx source response declares a binary-download media contract.
- The lock contains no callback or webhook-registration operation row. Its 40 webhook-named schemas are retained non-operation evidence and do not manufacture a sync cell.

## Foundation Atlas

The Batch R1 Atlas snapshot was consulted as authoring-only evidence. `source.retention-import.v1`, `source.projection-admission.v1`, `runtime.direct-execution.v1`, `warehouse.stage-etl.v1`, `warehouse.reverse-etl.v1`, and `transport.sync-contract.v1` are available. This Track A change therefore records no new `missing_foundation` cell and makes no execution claim.

## Mapping-control restrictions

- The byte-identical v2 lock retains `source_operation` provider payloads, while the current strict source-projection parser rejects that field. The matrix preserves all 49 rows and records the exact parser-acceptance repair required; it does not change shared parser code.
- No canonical Notion source descriptor exists in the fixed Batch R1 snapshot. The current `surface-sync --check` therefore reports that missing descriptor. The matrix preserves this as an all-row mapping restriction rather than generating a derived descriptor or removing source facts.

## GSD fallback

`scripts/gsd doctor` passed and the generated discuss/plan/execute/verify/review prompts were read. The Pi runtime cannot run the official interactive subagent workflow in this isolated coding session, so discussion, planning, implementation, verification, and review are recorded manually in this phase directory.
