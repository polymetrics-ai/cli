# GitHub ETL and reverse-ETL gap map — r1

**Snapshot:** `f96a47e801b89f25386c33951a53a93d1a4c7c8d` (2026-08-10)  
**Scope:** analysis only. No production source, bundle, test, or worktree files were changed.

## Executive answer

GitHub has two materially different capabilities today:

1. **GitHub → local warehouse ETL works.** The shipped bundle drives real declarative REST and GraphQL stream reads, and the app writes a connection-owned JSONL WAL then rebuilds one DuckDB-produced Parquet table. A focused local-server test exercises GitHub `pull_requests` through the warehouse in five legacy app modes.
2. **Warehouse → GitHub one-shot reverse writes work mechanically.** A focused CLI test seeds a Parquet table, makes a `create_pull_request` reverse plan, consumes its approval token, and observes the real GitHub engine/hook issuing `POST /pulls`, a metadata `PATCH`, and a reviewer `POST` to an `httptest` server.

That is **not** an end-to-end bidirectional/change-delivery sync as accepted in `data/cli-cdc-bidirectional-changefeed-design-r1/report.md`. The current reverse path re-reads a bounded table slice and calls `Write`; it has no immutable derived workset, receipt-backed baseline, destination receipt, delivery checkpoint, or replay identity. GitHub API actions also declare no provider idempotency-key header. It is therefore a guarded, one-shot API mutation facility, not a safe, resumable GitHub change-delivery destination.

The two requested destination answers are unambiguous:

- **GitHub source → another API destination:** **No.** `runConnectorETL` rejects every destination that does not implement `synccontract.DurableETLDestination` *before it reads GitHub*. No production engine-backed API connector implements that acknowledgement method. The generic algorithm is covered only with a test fake that does implement it.
- **GitHub source → PostgreSQL/MySQL destination:** **No.** PostgreSQL and MySQL declare `write: false` and their `Write` methods return `connectors.ErrUnsupportedOperation`. There is no database-write operation kind or `DatabaseWriteExecutor`/`WriteSession` implementation in `internal/`; the accepted database design describes those as future work.

The accepted bidirectional contract deliberately defers API destinations in its first slice, rather than permitting a generic HTTP sink: it requires `observe → immutable workset → plan/preview/approval → declared destination capability → receipt/checkpoint`, and says phase one should reject API destinations until they have delete, receipt, and replay proof ([bidirectional design:1-51](../cli-cdc-bidirectional-changefeed-design-r1/report.md)). The absence is therefore both a current product gap and an intentional phase-one boundary, not evidence that a generic API write should be opened.

## Method and evidence boundary

There is no `.codegraph/` directory in this worktree, so this map was built with page-wise `rg`, `jq`, and `nl` reads, then validated with focused tests. `git status --short` was empty before and after the work.

Commands run successfully:

```text
go test -timeout 20m ./internal/app -run 'TestGithubPullRequestsETLSupportsAllSyncModes|TestConnectorETLRefusesDestinationWithoutDurableAcknowledgementBeforeRead|TestGitHubCreateIssueReversePlanUsesDeclaredEndpoint|TestRunReverseETLRejectsDestructiveConnectorCommandWithoutConfirmation|TestLimitedReversePlanPreviewsAndRunsItsExactApprovedSlice'
go test -timeout 20m ./internal/cli -run 'TestReverseETLToGitHubCreatesPullRequestAfterApproval|TestGitHubCommandWriteUsesReversePlanApproval|TestGitHubDestructiveCommandRequiresTypedConfirmation|TestGitHubCommandSurfaceRunsStreamBackedIssueList'
go test -timeout 20m ./internal/connectors/engine -run 'TestGitHubProjectsDiscussionsCommandsMapToGraphQLStreams|TestBundleLoadEmbeddedGitHubOperations'
go test -timeout 20m ./internal/connectors/native/postgres ./internal/connectors/native/mysql -run 'Test.*Write'
go test -timeout 20m ./internal/connectors/commandrunner -run '^TestEveryImplementedCommandPassesRuntimePreflight$'
go test -timeout 20m ./internal/connectors/conformance -run '^TestConformance/github$'
go test -timeout 20m ./internal/connectors/hooks/github
go test -timeout 20m ./internal/warehouse -run 'TestEnsureOwnershipRefusesAnotherConnectionsDirectory|TestAssertOwnedTableCatchesAReintroducedSharedPath|TestLocationForRejectsUnsafeIdentityComponents'
go test -timeout 20m ./internal/synccontract -run 'Test.*Acknowledg|Test.*Commit'
go test -timeout 20m ./internal/app -run 'TestRunReverseETLRejectsApprovalTokenReplay|TestRunReverseETLRejectsPreviewDigestDriftBeforeNativeWrite|TestRunWarehouseETLSyncsNewDirectoryParentChainBeforeAcknowledgement|TestSecondConnectionDoesNotDestroyFirstConnectionRows'
```

The bundle counts below were obtained from the shipped JSON files, not inferred from help text:

| GitHub bundle fact | Evidence |
|---|---|
| 37 streams | `internal/connectors/defs/github/streams.json` contains 37 `streams` entries; the bundle advertises read/write at [metadata.json:2-23](../../../.treehouse/cli-83d592/1/cli/internal/connectors/defs/github/metadata.json:2). |
| 231 declared write actions | `writes.json` contains 231 entries; the certification inventory asserts exactly that at [stages_write_accounting_internal_test.go:5-44](../../../.treehouse/cli-83d592/1/cli/internal/connectors/certify/stages_write_accounting_internal_test.go:5). |
| 341 operation declarations | `operations.json`: 162 `rest_read`, 163 `rest_write`, 10 `binary_download`, 4 `graphql_query`, 1 `graphql_mutation`, 1 `local_git`. |
| Implemented user-facing GitHub surface | `cli_surface.json`: 14 `etl`, 164 `direct_read` (162 operation-backed), 9 `binary_download`, 196 `reverse_etl` commands mapping to 193 unique write actions, and **0** implemented `direct_write` commands. |
| Fixture coverage inventory | 37 stream fixture directories (38 page fixtures) and 67 write fixtures. The focused `TestConformance/github` passed. |

The file links in this report are anchors into the inspected worktree. Paths in the two accepted design reports are relative to their common `data/` directory.

## Code graph

### Diagram

```text
                                     SHIPPED GITHUB BUNDLE
  metadata.json ─┬─ streams.json (REST + GraphQL streams) ─┐
                 ├─ writes.json (231 declared actions) ────┼──> engine.Load / engine.New
                 ├─ operations.json (typed operation rows) ┤        [bundle.go:1153-1302,
                 └─ cli_surface.json (commands/intents) ───┘         connector.go:46-60]
                                                                          │
 ┌────────────────────────────────────────────────────────────────────────┼─────────────────────────────────────────────────────────┐
 │ GitHub read command                                                      │ GitHub ETL to local warehouse                              │
 │                                                                            │                                                             │
 │ pm github issue list                                                       │ pm etl run --connection ... --stream pull_requests         │
 │   → cli.runMaybeConnectorCommand                                           │   → app.RunETL                                              │
 │   → commandrunner.Run (implemented `etl`)                                 │   → source Connector.Read                                   │
 │   → engine.Connector.Read                                                  │   → engine.Read → Requester.Do → emitted records            │
 │   → engine.Read                                                           │   → runWarehouseETL                                         │
 │   → paginator / Requester.Do                                               │       → LocationFor / EnsureOwnership / AssertOwnedTable    │
 │   → extract → filter/project/hook → CLI records                           │       → wal/<stream>.jsonl raw envelopes                    │
 │                                                                            │       → TableWriter → DuckDB COPY → tables/<table>.parquet │
 └────────────────────────────────────────────────────────────────────────┼─────────────────────────────────────────────────────────┘
                                                                          │
                                  Reverse from warehouse                 │
  pm reverse plan ... --destination github:<credential>                  │
    → app.PlanReverseETL: pin source table; QueryTable (Parquet); map; hash; validate write
    → app.PreviewReversePlan: re-read same bounded slice; DryRunWrite; persist preview only when required
    → app.RunReverseETL: re-read/hash; consume one-time approval; writer.Write
    → engine.Connector.Write → engine.Write → prepared-write/gate
    → GitHub WriteHook for compound action, otherwise declarative HTTP request
    → GitHub REST API

  For create_pull_request specifically:
    POST /repos/{owner}/{repo}/pulls
      → decode returned `number`
      → optional PATCH /issues/{number} (labels/assignees/milestone)
      → optional POST /pulls/{number}/requested_reviewers

  Missing after reverse HTTP success:
    immutable derived workset / baseline / target receipt / delivery checkpoint / replay identity
```

### Checkable graph edges

| # | From | To | Data/control carried | Evidence |
|---:|---|---|---|---|
| 1 | GitHub JSON bundle | `engine.Load` | Metadata, streams, writes, operations, schemas, CLI surface, fixtures are loaded into `Bundle`. | [engine/bundle.go:1153-1302](../../../.treehouse/cli-83d592/1/cli/internal/connectors/engine/bundle.go:1153) |
| 2 | Bundle registry | engine connector | `bundleregistry.New` loads `defs.FS`, finds optional hooks, and constructs the engine-backed connector. | [bundleregistry/registry.go:15-28](../../../.treehouse/cli-83d592/1/cli/internal/connectors/bundleregistry/registry.go:15) and [engine/connector.go:46-60](../../../.treehouse/cli-83d592/1/cli/internal/connectors/engine/connector.go:46) |
| 3 | `pm github …` | connector command preflight | CLI resolves a connector surface command and calls `commandrunner.Preflight` before opening the app/project. | [cli/cli.go:696-758](../../../.treehouse/cli-83d592/1/cli/internal/cli/cli.go:696) |
| 4 | implemented ETL command | `Connector.Read` | `commandrunner.Run` turns an `etl` surface command into `ReadRequest{Stream, Config, Query, Limit}` and emits records. | [commandrunner/runner.go:243-290](../../../.treehouse/cli-83d592/1/cli/internal/connectors/commandrunner/runner.go:243) |
| 5 | engine connector wrapper | declarative reader | `Connector.Read` delegates directly to package-level `engine.Read`. | [engine/connector.go:124-126](../../../.treehouse/cli-83d592/1/cli/internal/connectors/engine/connector.go:124) |
| 6 | stream definition | runtime/requester | `Read` finds the `StreamSpec`, materializes config defaults, builds an authenticated runtime, optionally gives a stream hook first refusal, then calls the declarative reader. | [engine/read.go:24-68](../../../.treehouse/cli-83d592/1/cli/internal/connectors/engine/read.go:24), [engine/read.go:503-544](../../../.treehouse/cli-83d592/1/cli/internal/connectors/engine/read.go:503) |
| 7 | GitHub stream | provider request | `readOneSequence` selects pagination, interpolates stream path/config, merges query/page inputs, builds GraphQL payload when declared, and invokes `Requester.Do`. | [engine/read.go:248-348](../../../.treehouse/cli-83d592/1/cli/internal/connectors/engine/read.go:248), [engine/read.go:419-430](../../../.treehouse/cli-83d592/1/cli/internal/connectors/engine/read.go:419) |
| 8 | provider response | emitted records | It detects GraphQL errors, extracts records and response fields, filters, projects to schema, applies computed fields/hooks, and emits each record; paginator navigation follows the response. | [engine/read.go:353-416](../../../.treehouse/cli-83d592/1/cli/internal/connectors/engine/read.go:353) |
| 9 | GitHub REST stream declaration | edge 7 | `issues` is a REST `/repos/{{ owner }}/{{ repo }}/issues` stream with `state=all`, a `pull_request` filter, and `updated_at` incremental request configuration. | [defs/github/streams.json:74-96](../../../.treehouse/cli-83d592/1/cli/internal/connectors/defs/github/streams.json:74) |
| 10 | GitHub GraphQL stream declaration | edge 7 | `projects` is a `POST /graphql` cursor-paginated stream with a fixed document and `records.path`. The other GraphQL streams follow the same declarative reader route, not an operation executor. | [defs/github/streams.json:537-568](../../../.treehouse/cli-83d592/1/cli/internal/connectors/defs/github/streams.json:537) |
| 11 | direct-read command | bounded direct-read executor | `commandrunner.runDirectRead` calls `DirectRead` or operation-backed `OperationDirectRead`, passes page navigation, and rejects a zero-page context for requested navigation. | [commandrunner/runner.go:411-527](../../../.treehouse/cli-83d592/1/cli/internal/connectors/commandrunner/runner.go:411) |
| 12 | `pm etl run` | destination branch | `App.RunETL` resolves source/destination and routes only a `LocalWarehouseMaterializer` through `runWarehouseETL`; all other destinations go to `runConnectorETL`. | [app/app.go:1020-1080](../../../.treehouse/cli-83d592/1/cli/internal/app/app.go:1020) |
| 13 | local warehouse ETL | owned paths | The materializer validates legacy layout/format, obtains a location, ensures owner identity, obtains WAL/table paths, and independently asserts table ownership before source records are read. | [app/local_warehouse.go:60-120](../../../.treehouse/cli-83d592/1/cli/internal/app/local_warehouse.go:60) |
| 14 | source record | raw WAL/final record | The materializer enriches the final record, constructs a raw envelope with lineage, and appends the envelope to the WAL in bounded batches. | [app/local_warehouse.go:172-225](../../../.treehouse/cli-83d592/1/cli/internal/app/local_warehouse.go:172) |
| 15 | WAL | final Parquet | It fsyncs/closes the WAL, then rebuilds the table from it. The materializer writes only `raw.Record` into the final table. | [app/local_warehouse.go:233-287](../../../.treehouse/cli-83d592/1/cli/internal/app/local_warehouse.go:233), [app/local_warehouse.go:339-390](../../../.treehouse/cli-83d592/1/cli/internal/app/local_warehouse.go:339) |
| 16 | table writer | DuckDB Parquet | `TableWriter.Commit` stages JSON rows, invokes DuckDB `COPY … FORMAT parquet`, and atomically renames a single file. Reads use `read_parquet`. | [warehouse/parquet.go:88-136](../../../.treehouse/cli-83d592/1/cli/internal/warehouse/parquet.go:88), [warehouse/parquet.go:170-220](../../../.treehouse/cli-83d592/1/cli/internal/warehouse/parquet.go:170) |
| 17 | warehouse identity | physical table location | `LocationFor` uses `<root>/<workspace>/<connector>/<connection-id>`; `EnsureOwnership` and `AssertOwnedTable` fail closed on absent/foreign/corrupt owner records. | [warehouse/layout.go:1-16](../../../.treehouse/cli-83d592/1/cli/internal/warehouse/layout.go:1), [warehouse/layout.go:282-423](../../../.treehouse/cli-83d592/1/cli/internal/warehouse/layout.go:282) |
| 18 | `reverse plan` | `PlanReverseETL` | Planning pins the owner-scoped source connection, reads the Parquet table, maps records, verifies destination write capability/validation, hashes inputs, and persists a `ReversePlan`. | [app/app.go:1333-1483](../../../.treehouse/cli-83d592/1/cli/internal/app/app.go:1333) |
| 19 | `reverse preview` | dry-run preview | Generic preview re-reads and re-hashes the exact plan count, validates the writer, calls `DryRunWrite`, and persists a destructive preview/grant only when a confirmation challenge exists. | [app/app.go:1734-1931](../../../.treehouse/cli-83d592/1/cli/internal/app/app.go:1734) |
| 20 | `reverse run` | approved writer dispatch | Run rejects invalid/consumed approval, requires a stored preview only for direct-write or destructive plans, re-reads/hash-checks, atomically consumes approval, then invokes `writer.Write`. | [app/app.go:2106-2182](../../../.treehouse/cli-83d592/1/cli/internal/app/app.go:2106), [app/app.go:2311-2383](../../../.treehouse/cli-83d592/1/cli/internal/app/app.go:2311) |
| 21 | GitHub reverse action | engine write | `Connector.Write` delegates writes.json actions to `engine.Write`; validation and dry run are separately exposed by the same connector. | [engine/connector.go:186-225](../../../.treehouse/cli-83d592/1/cli/internal/connectors/engine/connector.go:186) |
| 22 | engine write | prepared/gated write | `Write` prepares/preview-binds the action and dispatches only through `ExecutePreparedWrite`; `ValidateWrite` applies the action schema/body rules. | [engine/write.go:59-107](../../../.treehouse/cli-83d592/1/cli/internal/connectors/engine/write.go:59), [engine/write.go:279-311](../../../.treehouse/cli-83d592/1/cli/internal/connectors/engine/write.go:279), [engine/prepared_write.go:42-147](../../../.treehouse/cli-83d592/1/cli/internal/connectors/engine/prepared_write.go:42) |
| 23 | approved action | hook or HTTP request | `executeApprovedWrite` gives a `WriteHook` first refusal; otherwise `executeWriteRecord` interpolates path/query/body and sends a request. | [engine/write.go:314-451](../../../.treehouse/cli-83d592/1/cli/internal/connectors/engine/write.go:314) |
| 24 | GitHub `create_pull_request` | compound REST calls | The action declares POST `/pulls`, record schema, and the `github` hook. The hook creates the PR, decodes its number, then performs optional metadata/reviewer follow-ups. | [defs/github/writes.json:186-250](../../../.treehouse/cli-83d592/1/cli/internal/connectors/defs/github/writes.json:186), [hooks/github/hooks.go:215-238](../../../.treehouse/cli-83d592/1/cli/internal/connectors/hooks/github/hooks.go:215), [hooks/github/hooks.go:326-376](../../../.treehouse/cli-83d592/1/cli/internal/connectors/hooks/github/hooks.go:326) |

### Executor inventory: declaration vocabulary versus code that can execute

The operation schema is broader than the engine execution surface. `OperationSpec` explicitly says executors are opt-in and unsupported kinds remain metadata-only ([engine/bundle.go:715-748](../../../.treehouse/cli-83d592/1/cli/internal/connectors/engine/bundle.go:715)). The loader recognizes the following operation kinds and their required metadata block ([engine/bundle.go:2010-2032](../../../.treehouse/cli-83d592/1/cli/internal/connectors/engine/bundle.go:2010)); that is not a guarantee of a runtime executor.

| Declared kind(s) | Concrete engine executor present? | GitHub declarations | GitHub user-reachability today |
|---|---|---:|---|
| Stream definition (not an `operations.json` kind): REST or GraphQL | **Yes:** `engine.Read` / `readOneSequence`. GraphQL is handled when the stream has a GraphQL block. | 37 streams; four are GraphQL streams. | **Yes.** 14 implemented `etl` commands; app ETL uses the same `Read` call. |
| `rest_read`, `provider_search` | **Yes:** `OperationDirectRead`; it rejects every other kind. | 162 `rest_read`; no GitHub `provider_search`. | **Yes:** 162 implemented operation-backed direct reads, plus two direct reads without an operation. |
| `rest_write` | **Yes:** `OperationDirectWrite`, but only a declared REST mutation and only after preview/gate. | 163 `rest_write`. | **Not through the shipped GitHub CLI:** zero implemented `direct_write` surface commands. It is an engine capability, not a GitHub user route in this bundle. |
| `binary_download` | **Yes:** `OperationBinaryDownload`, with an owned local destination root and size limit. | 10 declarations. | **Yes for 9** implemented binary-download commands. |
| Declarative `writes.json` action | **Yes:** `ValidateWrite`, `DryRunWrite`, `Write`; supports hooks and REST/GraphQL action bodies. | 231 actions. | **Yes for 196 `reverse_etl` commands / 193 unique actions**, through plan/run; the remaining actions are not all exposed in the CLI surface. |
| `graphql_query`, `graphql_mutation` operation rows | **No generic operation executor found.** `commandrunner` blocks an operation-backed command unless it is an implemented direct read/direct write/binary route. | 4 query rows, 1 delete mutation row. | The four *streams* work via `engine.Read`; their operation rows are metadata. `github.issue.delete` is explicitly `unsafe_or_disallowed`, and is blocked before credentials are resolved. |
| `xml_export`, `xml_import`, `file_upload`, `local_git`, `local_file`, `browser_open`, `stream_etl`, `composite` | **No generic operation executor found** in `internal/connectors/engine`; only the schema/block mapping exists. | 1 `local_git`; none of the other listed kinds. | **No.** The local-git row remains metadata/blocked. |
| Database write | **No declared operation kind and no database executor/session in `internal/`.** | 0. | **No.** PostgreSQL/MySQL native `Write` methods are stubs. |

Evidence for the concrete operation gates is [engine/direct_read.go:171-212](../../../.treehouse/cli-83d592/1/cli/internal/connectors/engine/direct_read.go:171), [engine/direct_write.go:59-157](../../../.treehouse/cli-83d592/1/cli/internal/connectors/engine/direct_write.go:59), [engine/direct_write.go:362-390](../../../.treehouse/cli-83d592/1/cli/internal/connectors/engine/direct_write.go:362), and [engine/binary_read.go:81-181](../../../.treehouse/cli-83d592/1/cli/internal/connectors/engine/binary_read.go:81). The command runner’s explicit block for operation-backed commands with no executor is [commandrunner/runner.go:315-383](../../../.treehouse/cli-83d592/1/cli/internal/connectors/commandrunner/runner.go:315).

## Where GitHub records land

### Physical layout and Parquet/DuckDB materialization

For a connection whose source connector is GitHub, the app derives this shape:

```text
<warehouse-root>/<workspace-id>/github/<connection-id>/
  owner.json
  wal/<stream>.jsonl
  tables/<table>.parquet
```

The source connector and opaque connection ID—not the connection display name and never a raw credential—are passed into `LocationFor` by [app/local_warehouse.go:549-554](../../../.treehouse/cli-83d592/1/cli/internal/app/local_warehouse.go:549). `Owner.SameIdentity` compares workspace, connector, and connection ID only ([warehouse/layout.go:124-144](../../../.treehouse/cli-83d592/1/cli/internal/warehouse/layout.go:124)). The writer does both `EnsureOwnership` and a separate path-derived `AssertOwnedTable` before writing ([app/local_warehouse.go:95-120](../../../.treehouse/cli-83d592/1/cli/internal/app/local_warehouse.go:95)). The focused warehouse tests passed for unsafe components, foreign ownership, and a reintroduced shared table path.

The WAL is the durable appendable raw record log. Every final table is rebuilt as one Parquet file from the WAL; the writer stages JSON, has DuckDB perform `COPY … FORMAT parquet`, fsyncs, and atomically renames ([warehouse/parquet.go:73-136](../../../.treehouse/cli-83d592/1/cli/internal/warehouse/parquet.go:73)). Both `QueryTable` and reverse planning read the final Parquet via warehouse/DuckDB paths; the general SQL engine also creates read-only views using `read_parquet` ([app/query_engine_duckdb.go:19-28](../../../.treehouse/cli-83d592/1/cli/internal/app/query_engine_duckdb.go:19), [app/query_engine_duckdb.go:125-180](../../../.treehouse/cli-83d592/1/cli/internal/app/query_engine_duckdb.go:125)).

### `_polymetrics_*` column facts

This distinction matters for reverse change delivery:

| Storage surface | Fields actually retained | Evidence | Consequence |
|---|---|---|---|
| Raw WAL envelope | `_polymetrics_raw_id`, `_polymetrics_run_id`, `_polymetrics_sync_id`, `_polymetrics_generation_id`, `_polymetrics_extracted_at`, `_polymetrics_loaded_at`, optional `_polymetrics_cursor` and cursor state, `_polymetrics_primary_key`, `_polymetrics_deleted`, plus `record`. | [app/local_warehouse.go:38-50](../../../.treehouse/cli-83d592/1/cli/internal/app/local_warehouse.go:38), [app/local_warehouse.go:196-217](../../../.treehouse/cli-83d592/1/cli/internal/app/local_warehouse.go:196) | It has useful extraction lineage, but it is not a reverse delivery checkpoint or a published change projection. |
| `raw.Record` written into final Parquet | Original source fields plus `_polymetrics_run_id`, `_polymetrics_synced_at`, `_polymetrics_deleted`, and `_polymetrics_cursor` when configured. | Enrichment: [app/local_warehouse.go:187-216](../../../.treehouse/cli-83d592/1/cli/internal/app/local_warehouse.go:187). Final write uses only `raw.Record`: [app/local_warehouse.go:339-390](../../../.treehouse/cli-83d592/1/cli/internal/app/local_warehouse.go:339). | Final Parquet does **not** retain raw ID, sync ID, generation, extracted/loaded timestamps, primary-key tuple, or opaque cursor state as standalone columns. |
| Deduped final Parquet | The newest raw record per key, excluding raw records marked deleted. | [app/local_warehouse.go:358-381](../../../.treehouse/cli-83d592/1/cli/internal/app/local_warehouse.go:358) | A missing final key is ambiguous: it might be deleted, filtered, never present, or outside a bounded source read. It is not safe delete evidence for reverse delivery. |

This exactly matches the accepted bidirectional design’s warning: current metadata is useful input but insufficient for a delivery baseline, volatile run/sync fields must not manufacture changes, and the final deduped table omits deleted rows ([bidirectional design:113-141](../cli-cdc-bidirectional-changefeed-design-r1/report.md)).

## Direct answers

### 1. API to API: can GitHub source data end to end into another API destination?

**No—not through `pm etl run` today.** The failure occurs before the GitHub source is read.

`App.RunETL` sends every non-local-warehouse destination to `runConnectorETL` ([app/app.go:1070-1075](../../../.treehouse/cli-83d592/1/cli/internal/app/app.go:1070)). That routine first requires the destination to implement `synccontract.DurableETLDestination` and immediately returns `DestinationDurabilityAdmissionError` if it does not ([app/app.go:1082-1089](../../../.treehouse/cli-83d592/1/cli/internal/app/app.go:1082)). Only after that admission does it call `source.Read` ([app/app.go:1152-1180](../../../.treehouse/cli-83d592/1/cli/internal/app/app.go:1152)); after complete batches it expects an unforgeable acknowledgement before forming a checkpoint ([app/app.go:1190-1205](../../../.treehouse/cli-83d592/1/cli/internal/app/app.go:1190)).

The interface exists and is meaningful: `DownstreamAcknowledgement` has an unexported durable marker and `CommitAfterDownstreamAcknowledgement` validates it before checkpoint commit ([synccontract/commit.go:11-96](../../../.treehouse/cli-83d592/1/cli/internal/synccontract/commit.go:11)). But a production-only search for `AcknowledgeETLDurability` found only the interface and the app caller—no production destination implementation. The engine connector wrapper exposes read/direct-read/direct-write/binary/write/validate/dry-run methods but no acknowledgement method ([engine/connector.go:124-225](../../../.treehouse/cli-83d592/1/cli/internal/connectors/engine/connector.go:124)). Thus a GitHub destination or another API connector with `Write` is still not an admitted ETL sink.

What is tested:

- **Failure behavior is tested.** `TestConnectorETLRefusesDestinationWithoutDurableAcknowledgementBeforeRead` asserts the typed admission error and verifies zero source requests ([app/sync_state_test.go:1147-1184](../../../.treehouse/cli-83d592/1/cli/internal/app/sync_state_test.go:1147)); it was run for this report.
- **The generic algorithm is tested only with a fake durable connector.** `batchDestination` implements both `Write` and `AcknowledgeETLDurability`, and the bounded-batch test uses that test double ([app/streaming_etl_test.go:97-128](../../../.treehouse/cli-83d592/1/cli/internal/app/streaming_etl_test.go:97)). That proves the app loop, not any API destination.
- **GitHub-as-source behavior is tested separately**, but only into local warehouse: see Question 5.

There is no test of GitHub source → a production API destination because no such destination path exists. The accepted bidirectional contract intentionally excludes generic API delivery in phase one and requires a declared destination capability/receipt/replay proof before later API admission ([bidirectional design:254-271](../cli-cdc-bidirectional-changefeed-design-r1/report.md), [bidirectional design:348-391](../cli-cdc-bidirectional-changefeed-design-r1/report.md)).

### 2. API to database: where precisely does GitHub → PostgreSQL/MySQL break?

Firstmate’s reading is correct.

| Breakpoint | PostgreSQL | MySQL | Evidence |
|---|---|---|---|
| Advertised capability | `write: false` | `write: false` | [defs/postgres/metadata.json:7-18](../../../.treehouse/cli-83d592/1/cli/internal/connectors/defs/postgres/metadata.json:7), [defs/mysql/metadata.json:8-22](../../../.treehouse/cli-83d592/1/cli/internal/connectors/defs/mysql/metadata.json:8) |
| Native `Write` call | Stub returns `connectors.ErrUnsupportedOperation`. | Same. | [native/postgres/connector.go:97-101](../../../.treehouse/cli-83d592/1/cli/internal/connectors/native/postgres/connector.go:97), [native/mysql/connector.go:43-46](../../../.treehouse/cli-83d592/1/cli/internal/connectors/native/mysql/connector.go:43) |
| App ETL admission | A database destination cannot supply a durable acknowledgement, so `runConnectorETL` fails before source read. | Same. | [app/app.go:1082-1089](../../../.treehouse/cli-83d592/1/cli/internal/app/app.go:1082), [synccontract/commit.go:27-49](../../../.treehouse/cli-83d592/1/cli/internal/synccontract/commit.go:27) |
| Reverse plan admission | `PlanReverseETL` rejects a destination whose metadata says `Write=false`. | Same. | [app/app.go:1395-1413](../../../.treehouse/cli-83d592/1/cli/internal/app/app.go:1395) |
| Engine executor | No database write kind is listed in the operation-kind map, and no `DatabaseWriteExecutor`, `WriteSession`, or `DurabilityReceipt` exists anywhere under `internal/`. | Same shared gap. | [engine/bundle.go:2010-2032](../../../.treehouse/cli-83d592/1/cli/internal/connectors/engine/bundle.go:2010); repository search recorded in Method. |

The `synccontract.Mode` names seven database modes, but its own documentation says parsing/declaring one does not make it executable ([synccontract/mode.go:8-62](../../../.treehouse/cli-83d592/1/cli/internal/synccontract/mode.go:8)). `RunETL` separately refuses contract modes without a matching native executor/conformance corpus ([app/app.go:1038-1050](../../../.treehouse/cli-83d592/1/cli/internal/app/app.go:1038)).

The two native stub tests passed during this investigation. They prove that the stubs still fail closed; they do not prove a database write. The accepted database framework reaches the same conclusion: the missing unit is a typed `DatabaseWriteExecutor`/`WriteSession` that returns a commit-coupled `DurabilityReceipt`, not another mode string or generic SQL escape hatch ([database framework:68-101](../cli-database-connector-framework-design-r1/report.md), [database framework:250-290](../cli-database-connector-framework-design-r1/report.md)).

### 3. Reverse ETL for GitHub specifically: can warehouse rows go back to GitHub?

**Mechanically yes, as a one-shot, approved mutation; no, not as accepted bidirectional delivery.**

The concrete positive path is not speculative:

1. `create_pull_request` is a GitHub `writes.json` action with POST `/repos/{{ config.owner }}/{{ config.repo }}/pulls`, a typed record schema, and `hook: "github"` ([defs/github/writes.json:186-250](../../../.treehouse/cli-83d592/1/cli/internal/connectors/defs/github/writes.json:186)).
2. `PlanReverseETL` reads the source Parquet table, maps the selected fields, checks the GitHub writer capability and schema, and persists a hash-bound plan ([app/app.go:1368-1483](../../../.treehouse/cli-83d592/1/cli/internal/app/app.go:1368)).
3. `RunReverseETL` re-reads/re-hashes the plan’s bounded row set, consumes the approval token once, and invokes `writer.Write` ([app/app.go:2106-2182](../../../.treehouse/cli-83d592/1/cli/internal/app/app.go:2106)).
4. The engine sees the GitHub `WriteHook`; `createPullRequest` posts the core PR, decodes `number`, then conditionally patches metadata and adds reviewers ([hooks/github/hooks.go:326-376](../../../.treehouse/cli-83d592/1/cli/internal/connectors/hooks/github/hooks.go:326)).

`TestReverseETLToGitHubCreatesPullRequestAfterApproval` seeds a warehouse table, uses `reverse plan` with mappings, invokes `reverse run`, and asserts exactly these three requests plus `records_succeeded: 1` ([cli/reverse_cli_test.go:122-216](../../../.treehouse/cli-83d592/1/cli/internal/cli/reverse_cli_test.go:122)). It was run and passed. The focused app test also exercises `create_issue` through plan, preview, run, and an observed REST endpoint ([app/reverse_confirmation_test.go:210-258](../../../.treehouse/cli-83d592/1/cli/internal/app/reverse_confirmation_test.go:210)).

Important limits on that claim:

- The test uses a local HTTP server, not GitHub.com. It proves correct local request dispatch, not credential scopes, provider side effects, rate limiting, or live provider receipts.
- For a **non-destructive generic action**, a plan may mint a token without a persisted preview ([app/app.go:1469-1476](../../../.treehouse/cli-83d592/1/cli/internal/app/app.go:1469)); the PR end-to-end test demonstrates plan → run directly. The CLI surface wording says plan/preview/approval/execute, but the current generic code only mandates stored preview for destructive or operation-direct-write plans ([app/app.go:2061-2063](../../../.treehouse/cli-83d592/1/cli/internal/app/app.go:2061)).
- `ReverseRun` stores counts, status, error, and (for direct REST operation writes) an operation response; it has no destination receipt or delivery checkpoint field ([app/types.go:245-258](../../../.treehouse/cli-83d592/1/cli/internal/app/types.go:245)).
- GitHub’s `writes.json` declares no `idempotency_key_header`. The engine disables retries for a non-idempotent action when that proof is absent ([engine/write.go:455-481](../../../.treehouse/cli-83d592/1/cli/internal/connectors/engine/write.go:455)). This is safer than blind retry, but it leaves a timeout/unknown-outcome user-visible and unreconciled.
- The plan represents a finite current table slice, not a delta since a prior delivery. It neither creates an immutable candidate file nor advances a receipt-backed baseline.

So a user can create a PR/issue or perform another exposed GitHub action from warehouse rows today. They cannot truthfully claim a warehouse-to-GitHub **sync** with the accepted contract’s delete, replay, progress, or recovery guarantees. The accepted first slice explicitly rejects API destinations until those facts are available; GitHub’s existing reverse route should be described as an approved one-shot action path, not as phase-one derived change delivery.

### 4. Database-destination foundations: what already exists, what is partial, what is absent?

| Foundation requested | Current state | Evidence of implementation | Test status | Assessment against accepted design |
|---|---|---|---|---|
| Plan, source pinning, mapping, input hash | **Exists and tested.** | Reverse plans pin the source connection, query/map rows, calculate a plan hash, and persist state. | Generic Parquet reverse path and GitHub action tests exist; focused reverse tests passed. | **Partial.** Hashing a bounded current slice protects approval drift, but is not a complete immutable Change Delivery Workset. |
| Preview | **Exists and tested for dry run.** | Preview re-reads/re-hashes and calls `DryRunWrite` without network dispatch. | GitHub command preview asserts zero HTTP calls ([cli/reverse_cli_test.go:286-305](../../../.treehouse/cli-83d592/1/cli/internal/cli/reverse_cli_test.go:286)). | **Partial.** Persisted preview is not required for ordinary non-destructive generic actions; database design requires it for every database plan. |
| One-time approval / consumption uncertainty | **Exists and tested.** | Approval token/hash, authenticated destructive grant/seal, and atomic consumption-before-dispatch are present. | Replay and drift tests were run; consumption logic is [app/app.go:2311-2383](../../../.treehouse/cli-83d592/1/cli/internal/app/app.go:2311). | **Usable foundation.** Must be retained for database work, but it does not itself provide a target receipt. |
| Typed destructive confirmation | **Exists and tested.** | Closed confirmation kind plus prepared-preview-bound gate. | GitHub delete test proves no HTTP before preview/approval/typed confirmation ([cli/reverse_cli_test.go:399-470](../../../.treehouse/cli-83d592/1/cli/internal/cli/reverse_cli_test.go:399)); focused tests passed. | **Usable foundation.** Correctly protects destructive actions; it does not make a non-destructive append/idempotency-safe. |
| Typed write contract | **Exists, but only as a one-shot connector DTO.** | `WriteRequest`, `WritePreview`, `WriteResult`, validators, dry runners, and prepared request digests are typed. | Engine write/direct-write tests exist and focused engine tests passed. | **Partial.** It lacks typed approved database plan, table/owner/schema/key/mode fields, `WriteSession`, and multi-batch atomicity. |
| Inbound/local durability acknowledgement | **Exists and tested for local warehouse/fakes.** | Local materializer creates durable acknowledgement only after WAL/Parquet/directory sync; synccontract refuses forged/missing acknowledgement. | Focused app/synccontract tests passed. | **Partial.** It is an ETL checkpoint acknowledgement, not an outbound target receipt. |
| Reverse receipt and delivery checkpoint/baseline | **Absent.** | Reverse completion persists only result counts/status. | No implementation or test can prove the required receipt/baseline semantics. | **Absent.** Accepted contract requires receipt then checkpoint/baseline advance. |
| Idempotency/replay identity | **Partial fail-safe only.** | Engine disables retry without a provider idempotency header; approval cannot replay. | Generic non-idempotent no-retry tests exist. No GitHub idempotency/reconciliation test. | **Absent for accepted delivery.** No stable workset/change identity, provider status lookup, or receipt ledger; GitHub adds no header declaration. |
| Warehouse ownership assertion | **Exists and tested.** | Structural identity path, `Owner.SameIdentity`, `EnsureOwnership`, and independent `AssertOwnedTable`. | Focused warehouse isolation/foreign-owner tests passed. | **Usable local foundation.** A database managed-target owner/control registry is still absent. |
| Database write session / managed target ownership | **Absent.** | PostgreSQL/MySQL source-only stubs and no DB executor implementation. | Stub tests pass; no positive database destination test exists. | **Absent.** The accepted framework’s `BeginWrite → Stage* → Commit → DurabilityReceipt` state machine is design-only. |

The accepted framework says the app safety gates should be reused but database write must add a persisted preview for every plan, target owner/object/schema checks, one bounded `WriteSession`, a confirmed receipt, and only then progress ([database framework:369-425](../cli-database-connector-framework-design-r1/report.md)). The accepted bidirectional contract adds a fuller requirement: freeze a current/baseline comparison into an immutable workset, then persist target receipt and delivery checkpoint ([bidirectional design:190-266](../cli-cdc-bidirectional-changefeed-design-r1/report.md), [bidirectional design:324-391](../cli-cdc-bidirectional-changefeed-design-r1/report.md)).

### 5. Testing gaps: what is real coverage versus reachability or assumption?

#### What is actually covered

| Behavior | Coverage | Limits |
|---|---|---|
| GitHub REST stream command | `TestGitHubCommandSurfaceRunsStreamBackedIssueList` observes path, query, projected/computed record fields, and JSON output against `httptest`. | One command/stream and local server only. [cli/cli_test.go:587-656](../../../.treehouse/cli-83d592/1/cli/internal/cli/cli_test.go:587) |
| GitHub direct reads/page contract | Direct-read tests observe request parameters, actual page size/number, and 120 returned records across two pages. | Selected REST endpoints only. [cli/cli_test.go:793-939](../../../.treehouse/cli-83d592/1/cli/internal/cli/cli_test.go:793) |
| All 37 GitHub stream definitions through engine replay | `TestConformance/github` passed. `runDynamicChecks` runs a real `engine.Read` against fixture replay for every stream with fixtures, and schema validation across every fixtured stream. | Recorded fixture behavior, not live GitHub. Pagination termination is only one first eligible stream; cursor advance is only one first eligible incremental stream. [conformance/dynamic.go:74-112](../../../.treehouse/cli-83d592/1/cli/internal/connectors/conformance/dynamic.go:74), [conformance/dynamic.go:287-323](../../../.treehouse/cli-83d592/1/cli/internal/connectors/conformance/dynamic.go:287), [conformance/dynamic.go:404-460](../../../.treehouse/cli-83d592/1/cli/internal/connectors/conformance/dynamic.go:404) |
| GitHub GraphQL streams declared/mapped | Bundle test checks all four map to `POST /graphql` and a GraphQL document; fixtures replay those four stream responses. | No live GraphQL/provider auth/scope/error/pagination-after-first-page proof. [engine/bundle_test.go:995-1043](../../../.treehouse/cli-83d592/1/cli/internal/connectors/engine/bundle_test.go:995) |
| GitHub action request construction | Fixture conformance executes real `engine.Write` against a capture server for every provided fixture (67 action fixtures), including schema validation and method/path/body expectation. | 67/231 actions; a fixture absence is explicitly skipped. It does not prove live side effects or durable provider outcome. [conformance/dynamic.go:539-592](../../../.treehouse/cli-83d592/1/cli/internal/connectors/conformance/dynamic.go:539) |
| GitHub special hooks | Unit tests cover close/reopen, PR follow-ups, labels, and fallback behavior. | Does not cover all 231 actions or a live provider. [hooks/github/hooks_test.go:211-543](../../../.treehouse/cli-83d592/1/cli/internal/connectors/hooks/github/hooks_test.go:211) |
| GitHub → local warehouse ETL | Five local app modes (`full_refresh_append`, overwrite, overwrite-deduped, incremental append, incremental append-deduped) use real GitHub bundle against a local server and query final warehouse rows. | Only `pull_requests`, a one-page test server, local warehouse, and no assertion of every final metadata field. [app/github_sync_modes_test.go:14-216](../../../.treehouse/cli-83d592/1/cli/internal/app/github_sync_modes_test.go:14) |
| Warehouse → GitHub reverse | CLI test creates a PR plus follow-ups after approval; app test creates an issue; CLI command test proves plan/preview/approval for close; destructive GitHub delete tests prove confirmation gate. | Local server; selected actions only; no receipt/recovery/delta semantics. [cli/reverse_cli_test.go:122-216](../../../.treehouse/cli-83d592/1/cli/internal/cli/reverse_cli_test.go:122), [cli/reverse_cli_test.go:218-397](../../../.treehouse/cli-83d592/1/cli/internal/cli/reverse_cli_test.go:218), [app/reverse_confirmation_test.go:22-258](../../../.treehouse/cli-83d592/1/cli/internal/app/reverse_confirmation_test.go:22) |
| Command surface admission | `TestEveryImplementedCommandPassesRuntimePreflight` walks every registered implemented command through the exact CLI preflight. | It is explicitly no-network, no-credential, no-action execution; it proves reachability/admission only. [commandrunner/runner_test.go:2547-2617](../../../.treehouse/cli-83d592/1/cli/internal/connectors/commandrunner/runner_test.go:2547) |

The “1,571 reachable commands” measurement belongs in the last row, not in a behavior claim. The test’s loop increments `checked` and calls `Preflight`, whose purpose is to prevent a surface from claiming `implemented` while the runtime would reject it; it deliberately does not resolve credentials or make a network request. It cannot establish source extraction, destination durability, provider response handling, or an end-to-end sync.

#### User-important gaps not covered by behavior tests

1. **No GitHub source → production API destination success test exists**, because the route is rejected before source read. The present negative test is valuable but does not substitute for a durable destination protocol.
2. **No GitHub source → PostgreSQL/MySQL destination success test exists**, because both writers are intentionally unsupported. There is no live DB target conformance or failure/recovery test for the accepted framework’s write-session state machine.
3. **No test proves a derived reverse delta.** There is no prior receipt-backed baseline, complete key reconciliation, immutable candidate workset, unchanged-payload suppression, tombstone propagation, or explicit physical-absence delete policy. The existing limited-slice test proves only that plan/preview/run re-read the same prefix.
4. **No test proves target-receipt ordering.** Current reverse tests assert request count/result count, not “confirmed destination receipt persisted before delivery checkpoint/baseline advances.” That state does not exist to test.
5. **No test proves GitHub timeout/unknown-outcome reconciliation.** The engine’s no-retry policy avoids a blind duplicate on transient non-idempotent error, but no GitHub action has provider idempotency metadata or status lookup; there is no durable replay identity test.
6. **No live GitHub integration test proves scopes, GraphQL access, rate-limit recovery, provider pagination variations, response semantics, or real side effects.** Fixture replay gives deterministic engine coverage, not an account/provider acceptance test.
7. **164 GitHub write actions lack even a request-shape fixture.** There are 67 write fixtures for 231 declared actions. The inventory test accounts for all 231 actions, but accounting/pairing is not request behavior.
8. **Operation metadata is not behavior.** GitHub has a GraphQL mutation operation (`github.issue.delete`) and GraphQL query rows, but no generic GraphQL operation executor. A command test proves the delete route is blocked; no test can prove an executor that does not exist.
9. **Final Parquet’s lineage shape is not validated as a reverse delivery contract.** Tests show warehouse/Parquet and reverse reading work, but no test asserts the accepted set of stable delivery fields, a publication projection, or the exclusion of volatile `run_id`/`synced_at` from a change fingerprint.

## Prioritised gap table

`Tested` means a behavior test for the current state exists; “negative only” means the test proves refusal rather than a working destination. It does not mean that the desired end state is covered.

| Priority / gap | Which flow it blocks | Exists / partial / absent | Tested yes/no | GitHub-specific or shared foundation |
|---|---|---|---|---|
| P0 — No database write executor/session, no managed DB target, and PostgreSQL/MySQL `Write` stubs | GitHub → PostgreSQL/MySQL; all API → database delivery | **Absent** | Yes—stub/refusal only; no positive destination test | Shared foundation |
| P0 — No immutable derived Change Delivery Workset or receipt-backed baseline/delta | Any truthful warehouse-derived reverse sync, including GitHub | **Absent** | No | Shared foundation |
| P0 — No target `DurabilityReceipt`, delivery checkpoint, or baseline advance after reverse write | Reverse recovery/progress; database target and eventual API target | **Absent** | No | Shared foundation |
| P0 — API destinations have no declared durable acknowledgement/capability/receipt protocol | GitHub source → another API via normal ETL; accepted future API delivery | **Absent** | Yes—negative admission only | Shared foundation (intentionally deferred by accepted phase one) |
| P0 — No stable delivery/replay identity and no provider reconciliation for unknown outcomes | GitHub reverse actions after timeout/crash; all future API destinations | **Partial** (no blind retry, but no proof/reconciliation) | No GitHub-specific test | Shared foundation with GitHub-specific manifestation |
| P1 — Generic non-destructive reverse plans can run without persisted preview | Guaranteed plan → preview → approval contract for ordinary GitHub creates and future DB plans | **Partial** | Yes—the PR end-to-end test observes plan → run; no test asserts it should be disallowed | Shared foundation |
| P1 — Final Parquet retains limited per-record metadata, not a stable delivery projection/tombstone contract | Delta derivation, key/revision lineage, deletes | **Partial** | No delivery-contract test | Shared foundation |
| P1 — GitHub GraphQL operation rows are metadata-only; no GraphQL operation executor | `github.issue.delete` and any direct GraphQL operation route | **Absent** | Yes—blocked command test; no positive executor test | GitHub-specific manifestation of shared engine gap |
| P1 — 164/231 GitHub actions have no write-request fixture | Broad correctness of exposed/unexposed GitHub action mapping | **Partial** | Yes for 67 only; no for remaining 164 | GitHub-specific |
| P1 — No live GitHub behavioral suite | Auth scopes, provider pagination/rate limits, real response/outcome semantics | **Absent** | No | GitHub-specific |
| P2 — Stream fixture pagination/cursor tests are not per stream | Completeness/resume confidence for each GitHub REST/GraphQL stream | **Partial** | Yes for fixture non-empty/schema across 37; no per-stream pagination/cursor proof | GitHub-specific |

## Recommendations (do not implement in this scout task)

1. **Treat the current GitHub reverse route as an approved one-shot action feature, not a bidirectional sync claim.** Its current tests justify that narrower statement. Do not advertise receipt-backed or resumable GitHub delivery.
2. **Implement the accepted database phase first, rather than a generic API or SQL writer.** The accepted database design calls for a typed approved plan, owner-scoped managed target, one `WriteSession`, bounded staging, confirmed commit, and `DurabilityReceipt` ([database framework:369-425](../cli-database-connector-framework-design-r1/report.md)). That fills the actual P0 database gap without weakening safety.
3. **Build reverse change delivery above that database writer as specified:** complete Parquet publication projection, destination-scoped prior baseline, immutable candidate workset, deterministic delta, target receipt, then delivery checkpoint. Do not derive it from a bounded reverse-plan slice or from the JSONL WAL ([bidirectional design:231-252](../cli-cdc-bidirectional-changefeed-design-r1/report.md)).
4. **Keep API destinations deferred until each has a closed capability/receipt/replay contract.** A bare successful HTTP response is not enough. For GitHub specifically, that would require an action-specific deterministic idempotency/reconciliation design, not merely exposing more `operations.json` rows.
5. **Add tests before capability promotion.** The accepted first database slice names the required live assertions: foreign/missing owner refusal, two-connection isolation, mode/keys/types, failure/rollback/unknown commit, no checkpoint before receipt, and source → warehouse → managed target data assertions ([database framework:794-837](../cli-database-connector-framework-design-r1/report.md), [database framework:1622-1637](../cli-database-connector-framework-design-r1/report.md)). For GitHub, add action fixtures/live-sandbox evidence only for a deliberately admitted, receipt-capable future API delivery action; do not interpret the 1,571 preflight sweep as that evidence.

## Final conclusion

The current implementation is strongest at **GitHub read → connection-owned local Parquet warehouse** and at **approved, one-shot warehouse → GitHub action dispatch**. It deliberately does not contain an API destination ETL protocol or a database destination writer. The reverse UI and engine provide valuable safety primitives—typed input validation, hashing, destructive confirmation, one-time approval, and warehouse owner isolation—but lack the delivery-state primitives required by the accepted bidirectional contract. The highest-value next work is therefore the shared database/change-delivery foundation, with strict test evidence, not a GitHub-specific generic write expansion.
