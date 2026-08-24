---
phase: cli-current-foundations-main-integration-r1
reviewed: 2026-08-21T02:52:36Z
depth: deep
review_lens: orchestration-delivery
source_sha: 8a8a866ff6d5282c28bda12acceed8a624218f01
diff_base: e62ae21d428f0d27225f9bff564dc2cd797f6b65
codegraph: unavailable
files_reviewed: 136
files_reviewed_list:
  - internal/app/app.go
  - internal/app/authorization.go
  - internal/app/change_capture_dispatch.go
  - internal/app/change_capture_dispatch_test.go
  - internal/app/declarative_typed_destination_approval.go
  - internal/app/durable_coordination.go
  - internal/app/etl_mode_dispatch.go
  - internal/app/foundations_integration_test.go
  - internal/app/issue_label_transport_approval.go
  - internal/app/issue_label_warehouse_transport.go
  - internal/app/local_warehouse.go
  - internal/app/payload_identity_test.go
  - internal/app/postgres_transport_approval_test.go
  - internal/app/rest_write_command_test.go
  - internal/app/reverse_approval_recovery_test.go
  - internal/app/reverse_confirmation_test.go
  - internal/app/stream_state.go
  - internal/app/transport_composition.go
  - internal/app/transport_composition_test.go
  - internal/app/transport_dispatch.go
  - internal/app/transport_dispatch_test.go
  - internal/app/types.go
  - internal/app/util.go
  - internal/app/write_approval.go
  - internal/cli/agentic_contract_test.go
  - internal/cli/cli.go
  - internal/cli/cli_test.go
  - internal/cli/connector_command_limits_test.go
  - internal/cli/connector_command_result_internal_test.go
  - internal/cli/docs.go
  - internal/cli/errors.go
  - internal/cli/etl_transport.go
  - internal/cli/etl_transport_test.go
  - internal/cli/github_flow_roundtrip_test.go
  - internal/cli/golden_transcript_test.go
  - internal/cli/reverse_plan_redaction_test.go
  - internal/cli/runtime_helpers.go
  - internal/cli/skills.go
  - internal/cli/structured_rest_body_help_test.go
  - internal/cli/testdata/golden_transcripts.json
  - internal/connectors/command_surface.go
  - internal/connectors/commandrunner/content_preservation_test.go
  - internal/connectors/commandrunner/github_declared_parity_test.go
  - internal/connectors/commandrunner/runner.go
  - internal/connectors/commandrunner/runner_test.go
  - internal/connectors/connectors.go
  - internal/connectors/connsdk/extract.go
  - internal/connectors/connsdk/http.go
  - internal/connectors/connsdk/http_test.go
  - internal/connectors/connsdk/multipart_bounds_test.go
  - internal/connectors/connsdk/stream.go
  - internal/connectors/database/transaction_stage.go
  - internal/connectors/defs/github/cli_surface.json
  - internal/connectors/defs/github/operations.json
  - internal/connectors/defs/github/sources/github-operation-descriptor.json
  - internal/connectors/defs/github/sync_transport.json
  - internal/connectors/defs/github/writes.json
  - internal/connectors/defs/postgres/sync_transport.json
  - internal/connectors/engine/binary_read.go
  - internal/connectors/engine/binary_read_test.go
  - internal/connectors/engine/bundle.go
  - internal/connectors/engine/bundle_test.go
  - internal/connectors/engine/connector.go
  - internal/connectors/engine/connector_test.go
  - internal/connectors/engine/direct_read.go
  - internal/connectors/engine/direct_read_paginate.go
  - internal/connectors/engine/direct_read_pagination_test.go
  - internal/connectors/engine/direct_read_test.go
  - internal/connectors/engine/direct_write.go
  - internal/connectors/engine/direct_write_multipart_test.go
  - internal/connectors/engine/direct_write_test.go
  - internal/connectors/engine/github_binary_download_test.go
  - internal/connectors/engine/graphql_operation.go
  - internal/connectors/engine/graphql_operation_test.go
  - internal/connectors/engine/interpolate.go
  - internal/connectors/engine/interpolate_test.go
  - internal/connectors/engine/operation_direct_read_bindings_test.go
  - internal/connectors/engine/operation_headers.go
  - internal/connectors/engine/operation_kind.go
  - internal/connectors/engine/operation_kind_test.go
  - internal/connectors/engine/operation_multipart_test.go
  - internal/connectors/engine/operation_parameters.go
  - internal/connectors/engine/polling_watermark.go
  - internal/connectors/engine/prepared_write.go
  - internal/connectors/engine/rate_limit_coordination_test.go
  - internal/connectors/engine/rate_limit_parking.go
  - internal/connectors/engine/rate_limit_runtime.go
  - internal/connectors/engine/rate_limit_runtime_test.go
  - internal/connectors/engine/read.go
  - internal/connectors/engine/record_schema_promotion.go
  - internal/connectors/engine/record_schema_promotion_test.go
  - internal/connectors/engine/response_receipt.go
  - internal/connectors/engine/schema.go
  - internal/connectors/engine/schema/cli_surface.schema.json
  - internal/connectors/engine/schema/operations.schema.json
  - internal/connectors/engine/schema/sync_transport.schema.json
  - internal/connectors/engine/schema/writes.schema.json
  - internal/connectors/engine/schema_test.go
  - internal/connectors/engine/sensitive_transform.go
  - internal/connectors/engine/status_check.go
  - internal/connectors/engine/status_check_test.go
  - internal/connectors/engine/structured_rest_body.go
  - internal/connectors/engine/structured_rest_body_test.go
  - internal/connectors/engine/text_export_test.go
  - internal/connectors/engine/write.go
  - internal/connectors/engine/write_prepare.go
  - internal/connectors/engine/write_query_test.go
  - internal/connectors/engine/write_record_hook_test.go
  - internal/connectors/engine/write_retry_policy_test.go
  - internal/connectors/engine/write_test.go
  - internal/connectors/engine/xero_operations_test.go
  - internal/connectors/native/ashby/connector_contract_test.go
  - internal/connectors/native/ashby/engine_delegate.go
  - internal/connectors/native/postgres/bootstrap.go
  - internal/connectors/native/postgres/cdc.go
  - internal/connectors/native/postgres/cdc_v2.go
  - internal/connectors/native/postgres/reader.go
  - internal/connectors/native/postgres/transport_source.go
  - internal/connectors/native/postgres/transport_source_test.go
  - internal/connectors/operation_headers.go
  - internal/connectors/sync_transport.go
  - internal/connectors/transportpolicy/policy.go
  - internal/coordination/auth_cohort.go
  - internal/coordination/durable_store.go
  - internal/coordination/durable_store_edges_test.go
  - internal/coordination/rate_parking.go
  - internal/coordination/rate_parking_test.go
  - internal/state/store.go
  - internal/synccontract/commit.go
  - internal/synctransport/arrow_fast_path_controller.go
  - internal/synctransport/arrow_fast_path_pipeline.go
  - internal/synctransport/definition_composition.go
  - internal/synctransport/orchestrator.go
  - internal/synctransport/registry.go
  - internal/synctransport/transport_test.go
  - internal/synctransport/types.go
findings:
  blocker: 14
  warning: 3
  info: 0
  total: 17
status: issues_found
verdict: blockers
---

# Foundation Post-Fix Orchestration/Delivery Review

## Summary

The frozen rollup is **not shippable from the orchestration/delivery lens**. The review found 14 blockers and 3 warnings. The highest-risk defects are an explicitly claimed-but-absent pre-I/O ownership fence, silent treatment of a pagination budget as source exhaustion, declaration-self-certified idempotency, incomplete provider read-back, unsafe terminal-run publication, cross-process authentication fences that do not stop admitted work, and a binary-download publication sequence that can expose or delete the wrong file.

The current implementation does close several prior defects: output is collected before read-back/checkpoint work, a successful full-overwrite publication disarms pre-publication abort, runtime tombstone checks exist, direct read/download failures retain receipts, and declared no-input actions are reachable without reopening hollow JSON writes. Those closures do not compensate for the blockers below.

I built and froze the independent finding set from current source and tests before reading the conclusions in `REVIEW.md` and `REVIEW-FIX.md`. The final cross-check found several claimed closures that the frozen code does not provide, most notably CF-B24: `REVIEW-FIX.md` maps it to `933afaa25`, but `git show 933afaa25 -- internal/app/transport_dispatch.go internal/synctransport/orchestrator.go` changes result collection/publication state only and adds no ownership claim.

## Frozen Source and Preflight

- Review root: `/Users/karthiksivadas/.treehouse/cli-83d592/57/cli`
- Required/reviewed HEAD: `8a8a866ff6d5282c28bda12acceed8a624218f01`
- Diff base: `e62ae21d428f0d27225f9bff564dc2cd797f6b65`
- HEAD was verified before review and again before report creation.
- `git status --short --untracked-files=all`, unstaged diff, and cached diff were empty before review. No source, test, generated, branch, commit, remote, or PR state was changed.
- `.codegraph/` does not exist at the repository root, so CodeGraph was unavailable by repository policy. Review used read-only source/diff tracing.

## Scope Manifest

The base-to-HEAD diff contains 819 files. The orchestration/delivery lens selected 114 changed core files (20 App, 15 CLI, 68 connector/runtime, 4 coordination, 1 sync contract, 6 sync transport), four changed GitHub operation/write/CLI definition artifacts needed for direct/binary delivery tracing, and 18 unchanged implementations required to close call paths. The exact 136-file set is preserved in `files_reviewed_list`.

Cross-module state machines traced:

1. App run creation -> route/preflight -> source extraction -> stage/reopen -> destination apply/publish -> provider read-back -> checkpoint CAS -> stage retirement/approval marker -> terminal run publication.
2. File-backed rate-limit park -> scope leader/claim -> renewal -> resume attempt -> rearm/retry -> completion/delete -> process restart recovery.
3. Authentication runtime resolution -> durable epoch admission -> provider operation -> verified-invalid report/fence -> sibling cancellation -> repair/new epoch.
4. PostgreSQL CDC staging -> application warehouse WAL/table durability -> source-stage receipt -> receipt restoration -> checkpoint persistence -> source acknowledgement.
5. Reverse/direct write preview and approval -> provider mutation/result -> terminal plan/run persistence -> CLI terminal envelope/nonzero exit.
6. Direct read, binary download, and binary upload declaration -> command preflight -> bounded provider I/O -> receipt/output -> terminal CLI publication.

The four other review lenses and generated catalog breadth were not re-reviewed as independent subjects; they were followed only where they feed one of these delivery state machines.

## Checks Run

| Check | Result |
|---|---|
| Exact `git rev-parse HEAD` | PASS: `8a8a866ff6d5282c28bda12acceed8a624218f01` |
| Source/test/generated cleanliness | PASS before review and before report creation |
| `.codegraph/` presence | Unavailable (directory absent) |
| `git diff --check e62ae21d...8a8a866f` | PASS |
| `go vet ./internal/synctransport ./internal/coordination ./internal/synccontract ./internal/app ./internal/connectors/commandrunner ./internal/connectors/engine` | PASS |
| `go test -race -count=1 ./internal/synctransport ./internal/coordination` | PASS (`synctransport` 6.068s, `coordination` 14.044s) |
| `go test -count=1 ./internal/synctransport ./internal/coordination ./internal/synccontract ./internal/app ./internal/connectors/commandrunner ./internal/connectors/engine` | PASS (`app` 324.505s; all listed packages passed) |
| Same targeted run including `./internal/cli` | INCOMPLETE: the package-wide 10-minute timeout fired while `TestSkillsGenerateWritesAgentSkills` was still decoding large connector bundles (`internal/cli/agentic_contract_test.go:92`); this is not counted as a product finding. Other package results above were already PASS. |

Passing tests do not cover the adversarial interleavings and commit-outcome cases below; several existing tests encode the incorrect behavior explicitly.

## Blockers

### ORCH-PF-B01 — Stream ownership is still checked only after provider and warehouse side effects

**Evidence:** `internal/app/transport_dispatch.go:235-253` snapshots stream state and chooses a generation but acquires no durable claim. The only ownership CAS is the `Commit` callback at `internal/app/transport_dispatch.go:282-313`, after ordinary apply/read-back (`internal/synctransport/orchestrator.go:237-283`) or full-overwrite apply/publish/read-back (`internal/synctransport/orchestrator.go:435-497`). The same ordering exists in both Arrow controllers (`internal/synctransport/arrow_fast_path_controller.go:145-216`; `internal/synctransport/arrow_fast_path_pipeline.go:277-283,125-158`) and the CDC receiver publishes WAL/table files before its state CAS (`internal/app/change_capture_dispatch.go:109-213,216-253`). Existing tests prove the loser already mutated: `TestRunETLTransportRejectsStaleCheckpointWriter` pauses only after losing apply (`internal/app/transport_dispatch_test.go:1056-1124`), and the full-overwrite assertion requires the losing apply, publish, and read-back all to have run (`internal/app/transport_dispatch_test.go:1383-1392`).

**Cross-file path:** two App processes load the same state -> each enters `runTransportETL`/CDC with the same expectation -> both perform destination/local-warehouse effects -> winner commits state -> loser finally reaches App CAS and receives `errTransportStreamStateConflict`.

**Failure mode:** a stale append can duplicate non-provider-keyed effects; a stale overwrite can publish older data after the winner and then lose only the local checkpoint CAS. Local state names the winner while provider state contains the loser. Keyed idempotency does not impose ordering between different generations.

**Six-surface impact:** ETL=blocker for ordinary/full-overwrite/serial Arrow/pipeline Arrow/CDC; reverse ETL=blocker when executed as a transport destination; direct read=none; direct write=not fenced when reused as a transport destination; binary download=none; binary upload=affected if promoted into a replayable destination route.

**Exact fix:** atomically claim a durable connection+stream execution lease/fence before source, stage, or provider I/O. Persist a monotonic fence generation and stable work identity; carry them through every stage/mutation/publication, require the provider/local warehouse to reject stale fences atomically, renew while active, and make checkpoint commit retire the exact same fence. Restart takeover must reconcile a published receipt before replay.

**Exact regression test:** `TestTransport_TwoAppsFenceBeforeAnySideEffect` must pause loser before mutation, let winner claim/commit, resume loser, and assert zero losing stage/apply/publish/read-back calls (or an atomic provider fence rejection) plus winner-consistent provider/checkpoint state. Table-test ordinary append/upsert, full overwrite, serial Arrow, pipeline Arrow, CDC local publication, lease expiry, crash, renewal loss, and restart takeover.

### ORCH-PF-B02 — A pagination budget is reported as source exhaustion and can publish a truncated overwrite

**Evidence:** the installed GitHub source declares `full_overwrite` and other replayable modes (`internal/connectors/defs/github/sync_transport.json:3-55`). The generic source defaults `transport_max_pages` to one (`internal/app/issue_label_warehouse_transport.go:1171-1234,1241-1254`). Engine `Read` simply breaks when the cap is reached (`internal/connectors/engine/read.go:307-326`) and returns nil without an incomplete/continuation signal (`internal/connectors/engine/read.go:424-433`). `TestOpenComposedGitHubCommitsHonorsTransportMaxPages` explicitly expects a three-page provider to return success with only the first page when the setting is omitted (`internal/app/transport_composition_test.go:2442-2519`). Full overwrite then publishes the gathered prefix as complete (`internal/synctransport/orchestrator.go:461-497`).

**Cross-file path:** GitHub descriptor -> App declarative source -> `engine.Connector.Read` -> request cap break -> source returns nil -> orchestrator treats last emitted candidate as terminal -> destination publishes/commits.

**Failure mode:** full overwrite replaces the destination with only the first configured page and reports success. Append/upsert with a cap cannot progress either: the next run begins from page one; replay modes re-emit the same prefix, while resume-search modes may never find a checkpoint outside the cap.

**Six-surface impact:** ETL=blocker/data loss for capped declarative sources; reverse ETL=downstream targets can receive a truncated source set; direct read=its separate pagination API exposes page state and is not itself implicated; direct write=none; binary download=none; binary upload=none.

**Exact fix:** make provider pagination return typed `exhausted` versus `budget_stopped` plus an opaque continuation. Full overwrite must abort the shadow and refuse publication unless exhaustion is proven. Incremental/append modes must persist a provider-owned continuation bound into the checkpoint and resume after the cap, or refuse capped execution before I/O.

**Exact regression test:** `TestDeclarativeTransport_PageBudgetIsNotEOF` with a three-page provider and omitted/1/2 caps: full overwrite must perform zero publish and leave the target unchanged; append must persist an incomplete continuation and the next run must start at page N+1, eventually delivering each record once. Include restart between capped runs and an unlimited run that proves exhaustion.

### ORCH-PF-B03 — Connector declarations self-certify keyed delivery and admit non-idempotent POST actions

**Evidence:** App collects conformance references directly from the same connector descriptors being admitted (`internal/app/issue_label_warehouse_transport.go:78-139`), feeds them into `NewDefinitionConformanceVerifier` (`internal/app/transport_composition.go:12-31`), and that verifier accepts exactly those factory values (`internal/synctransport/definition_composition.go:46-83`). The generic destination contract checks only that the descriptor says `keyed` and that an action/method/path exists (`internal/app/issue_label_warehouse_transport.go:418-477`). A production-shaped test declares a create-style POST with no idempotency header, invents its own conformance ID, and is admitted and executed (`internal/app/transport_composition_test.go:405-489`). Engine disables only in-request retries when no header exists (`internal/connectors/engine/write.go:645-674`); it cannot prevent a whole-run replay after a failed checkpoint.

**Cross-file path:** connector definition -> factory generated from that definition -> verifier authority -> registry delivery compatibility -> generic destination -> non-idempotent provider POST -> read-back/checkpoint failure -> rerun issues another POST.

**Failure mode:** a connector can claim keyed delivery without any provider key or independently certified intrinsic idempotency. If the provider accepts the mutation and read-back/checkpoint fails, a retry creates a second object while the framework continues to advertise keyed safety.

**Six-surface impact:** ETL=blocker for reusable declarative destinations; reverse ETL=the same write action semantics are unsafe if routed as a destination; direct read=none; direct write=direct execution avoids HTTP retry but does not establish transport replay safety; binary download=none; binary upload=must not be promoted as keyed without provider proof.

**Exact fix:** build conformance authority from an immutable, independent artifact keyed by executor reference plus definition/action digest, never from the candidate descriptor. Require a provider-documented idempotency header populated from stable preview/workset identity, or a separately certified intrinsic idempotency mechanism whose digest is bound to the action. Reject claimed-keyed actions without that proof before source I/O.

**Exact regression test:** `TestDeclarativeDestination_ClaimedKeyedWithoutIndependentProofIsRejected` must use the current synthetic POST and assert composition/preflight failure and zero requests. A positive case must supply independently registered evidence plus an idempotency header, fail after provider success, retry, assert the identical key, and prove one provider mutation. Changing action bytes or evidence must invalidate admission.

### ORCH-PF-B04 — Generic read-back ignores the mutation receipt and scans an unrelated bounded prefix

**Evidence:** App passes acknowledgement output as `Receipt` (`internal/app/issue_label_warehouse_transport.go:330-340`), but engine `ReadBackDeclarativeDestination` never reads `req.Receipt`; it validates only a stream name and calls an unfiltered `Read` (`internal/connectors/engine/connector.go:242-272`). It aborts once total returned rows exceed `MaxRecords`, irrespective of where the expected records live (`:260-264`). The acknowledgement output is already the public sanitized write result (`internal/app/issue_label_warehouse_transport.go:248-267`), so it is not a complete private locator in the first place.

**Cross-file path:** destination write -> public output attached to durable acknowledgement -> App read-back policy -> engine ignores receipt -> collection scan from provider start -> matcher -> checkpoint.

**Failure mode:** any destination collection larger than `max_records` can fail read-back forever even when the just-written row exists, or an expected row outside the first bounded prefix is never checked. The successful mutation is then replayed because no checkpoint advances. Conversely, there is no proof that the read was causally tied to the mutation receipt.

**Six-surface impact:** ETL=blocker for generic destination verification; reverse ETL=destination durability semantics affected; direct read=the underlying declared read is functional but is not a receipt-bound verifier; direct write=none in standalone mode; binary download=none; binary upload=none.

**Exact fix:** retain a complete internal mutation receipt/read-back locator distinct from public output. The declaration must project that receipt or the exact expected identities into a provider-owned point query (or bounded pagination that ends only after every expected identity is found). Preflight must prove the admitted batch fits the read-back proof capacity; never treat a full collection prefix as receipt verification.

**Exact regression test:** `TestDeclarativeDestination_ReadBackUsesInternalReceiptLocator` must use a collection larger than `max_records`, place the expected row beyond page one, and prove receipt-targeted success. An ignored, missing, foreign, or changed receipt must prevent checkpoint. Include eventual consistency retries against the same locator and prove the public output masks a secret while the private locator remains complete.

### ORCH-PF-B05 — Declarative source cloning silently rounds integers above 2^53

**Evidence:** `cloneTransportRecord` serializes and then calls default `json.Unmarshal` (`internal/app/issue_label_warehouse_transport.go:1574-1583`), which converts generic JSON numbers to `float64`. This clone is used for configured-issue and collection reads before checkpoint hashing and destination staging (`internal/app/issue_label_warehouse_transport.go:1147-1164,1211-1230`). The provider decoder deliberately preserves `json.Number` (`internal/connectors/connsdk/extract.go:10-18`), and native PostgreSQL emits driver-native values such as `int64` (`internal/connectors/native/postgres/reader.go:99-113`).

**Cross-file path:** provider decode/native scan -> declarative source callback -> JSON round-trip clone -> stage/mapping/provider apply -> checkpoint hash/read-back.

**Failure mode:** `9007199254740993` becomes the nearest representable float before approval, write, or checkpoint identity. The system can authorize and deliver a different identifier/value than the provider returned, and subsequent resume/read-back operates on the corrupted representation.

**Six-surface impact:** ETL=blocker for declarative sources and source-to-destination mapping; reverse ETL=incorrect staged source records can feed writes; direct read=provider receipt path preserves numbers; direct write=none; binary download=metadata only if passed through this source adapter; binary upload=none.

**Exact fix:** replace the JSON round trip with the transport typed recursive clone, or decode with `UseNumber` and define a precision-preserving normalization. Preserve `json.Number`, signed/unsigned integers, nested arrays/maps, and reject unsupported mutable types explicitly.

**Exact regression test:** `TestDeclarativeTransportClone_PreservesLargeNumbers` must drive REST `json.Number("9007199254740993")`, PostgreSQL `int64`, nested numeric arrays/maps, and negative/unsigned boundaries through source -> stage -> destination/read-back/checkpoint, asserting exact type/value and caller non-mutation.

### ORCH-PF-B06 — Provider read-back compares Go numeric representation rather than value semantics

**Evidence:** provider matching uses `reflect.DeepEqual` for expected fields (`internal/app/issue_label_warehouse_transport.go:361-395`) and hashes raw Go values for identity (`:398-415`; `internal/app/util.go:307-312`). Engine/provider JSON uses `json.Number` (`internal/connectors/engine/read.go:1463-1474`), while native/staged values can be `int64` (`internal/connectors/native/postgres/reader.go:99-113`).

**Cross-file path:** source record with native numeric type -> mapped destination write -> provider JSON read-back with `json.Number` -> raw-type identity/equality matcher -> checkpoint decision.

**Failure mode:** semantically equal values such as `int64(42)` and `json.Number("42")` compare unequal; `1` and `1.0` can hash to different identities. A successful, correct provider mutation therefore never advances its checkpoint and is replayed. Large values cannot safely be normalized through float64.

**Six-surface impact:** ETL=blocker for generic read-back with numeric identity/expected fields; reverse ETL=provider verification semantics affected; direct read=returns provider truth without cross-type matching; direct write=none standalone; binary download=none; binary upload=none.

**Exact fix:** implement schema-aware canonical JSON comparison and identity encoding that preserves arbitrary integer/decimal precision. Normalize expected and provider fields through the same routine; do not coerce strings or booleans into numbers.

**Exact regression test:** `TestDeclarativeReadBack_NumericSemanticEquality` must prove `int64(42)` equals `json.Number("42")`, values above 2^53 remain exact, and `43` is rejected. Explicitly specify/test whether schema says `42` and `42.0` are equivalent, for both identity and expected fields.

### ORCH-PF-B07 — Apply/publish and read-back share one deadline budget

**Evidence:** ordinary execution creates `applyCtx` before apply and reuses it for read-back (`internal/synctransport/orchestrator.go:237-273`). Full overwrite reuses `publishCtx` (`internal/synctransport/orchestrator.go:471-485`). Serial and pipeline Arrow do the same (`internal/synctransport/arrow_fast_path_controller.go:184-199`; `internal/synctransport/arrow_fast_path_pipeline.go:125-140`). The generic read-back's own timeout is only a child of that potentially expired parent (`internal/app/issue_label_warehouse_transport.go:330-358`). `RunRequest.UnitDeadline` is documented as an apply/read-back unit bound (`internal/synctransport/types.go:505-508`).

**Cross-file path:** destination apply/publish consumes most/all unit deadline -> successful external effect -> read-back starts with residual or already-cancelled context -> checkpoint refused -> replay/reconciliation.

**Failure mode:** a slow but successful write systematically makes verification fail. Eventual-consistency retries are especially likely to be starved, producing repeated side effects despite each phase individually fitting the configured unit bound.

**Six-surface impact:** ETL=blocker for ordinary/full-overwrite/serial Arrow/pipeline Arrow; reverse ETL=affected when provider verification is in a transport destination; direct read=separate deadline; direct write=separate deadline; binary download=separate stall/size bounds; binary upload=separate write timeout.

**Exact fix:** cancel apply/publish immediately on return and create a fresh read-back context with its own explicitly bounded phase deadline (using the stricter declared/provider policy where applicable). Record apply/publish and read-back timing separately.

**Exact regression test:** `TestTransport_ReadBackGetsIndependentUnitDeadline` must use a 50ms unit deadline, 40ms successful apply/publish, then a 20ms read-back (including a second eventual-consistency attempt), and succeed for ordinary, full-overwrite, serial Arrow, and pipeline Arrow. An apply exceeding 50ms must still fail.

### ORCH-PF-B08 — Runtime/catalog copies still share nested mutable state, including across Arrow pipeline goroutines

**Evidence:** `cloneRuntimeConfig` copies only the catalog struct and outer `Streams` slice (`internal/synctransport/types.go:668-678`). Each `connectors.Stream` contains mutable `Fields`, `PrimaryKey`, `CursorFields`, and `Schema`, while `Catalog.Discovery` contains a pointer and mutable `Failures` (`internal/connectors/connectors.go:183-225`). Arrow begin requests clone only this shallow runtime; per-segment apply passes the original runtimes and binding directly (`internal/synctransport/arrow_fast_path_controller.go:57-64,145-151`; `internal/synctransport/arrow_fast_path_pipeline.go:51-58,277-283`). In pipeline mode the source producer and destination consumer can observe the same nested objects concurrently.

**Cross-file path:** App runtime/catalog -> orchestrator clone wrapper -> source/destination/Arrow session -> executor mutates nested catalog/binding -> caller, later segment, producer, or read-back observes mutation.

**Failure mode:** a buggy or adversarial executor can change primary keys, schemas, discovery status, or per-segment metadata through an allegedly defensive request. In the pipeline this is also a data race; later units can be authorized/mapped against mutated state.

**Six-surface impact:** ETL=blocker across ordinary and Arrow transports; reverse ETL=source mapping/destination runtime can be altered; direct read/direct write/binary download/binary upload=not passed through these clone wrappers in standalone execution.

**Exact fix:** deep-clone every catalog stream slice/raw schema and discovery/failure value. Construct a freshly cloned `ArrowBulkApplyRequest` for every segment, including runtime/source runtime/binding. Define ownership for reference-counted Arrow records and prevent the destination from releasing or retaining producer-owned references outside the call.

**Exact regression test:** `TestTransportRequests_DefensiveCopyAllNestedState` must have the first executor mutate nested catalog fields/schema/failures, config/secrets, binding primary keys, and Arrow request metadata. Assert original caller state and every later source/read-back/segment are unchanged for ordinary, full-overwrite, serial Arrow, and pipeline Arrow; run the pipeline case under `-race`.

### ORCH-PF-B09 — The ordinary ETL CLI still drops persisted terminal runs and receipts

**Evidence:** the ordinary `etl run` path returns immediately on any `RunETL` error (`internal/cli/cli.go:750-758`), unlike the approved path which emits a non-empty run then preserves the categorized nonzero exit (`internal/cli/etl_transport.go:495-519`). It also returns a runtime-ledger error before emitting a successfully persisted App run (`internal/cli/cli.go:759-767`; `internal/cli/runtime_helpers.go:31-48`). App creates the run before later failures (`internal/app/app.go:1381-1400`) and `failRunWithResult` normally returns the persisted failed run plus the execution error (`internal/app/app.go:3482-3535`). On generic may-have-committed finalization errors it can still return a zero run (`:3520-3528`); analogous completion logic returns zero unless it has an acknowledged-transport special case (`internal/app/app.go:1692-1700`).

**Cross-file path:** CLI ordinary ETL -> App persists running run -> execution fails/parks or terminal save is ambiguous -> App returns run+error (or zero on generic ambiguity) -> CLI returns error -> top-level writer emits a second generic Error instead of the terminal run.

**Failure mode:** persisted failed/parked runs and destination receipts disappear from stdout, violating the one-terminal-envelope contract. A runtime sidecar outage hides an otherwise completed App run. Directory-sync ambiguity can leave a terminal run on disk while CLI claims no run exists.

**Six-surface impact:** ETL=blocker for legacy/ordinary/CDC JSON and human flows; reverse ETL=its CLI path already handles a non-empty run but shares App ambiguity in ORCH-PF-B10; direct read=none; direct write/binary upload=reverse terminal path; binary download=none.

**Exact fix:** reconcile every App terminal save with `state.CommitOutcome`: on may-have-committed, reload the exact run and return it only if it is durable/visible. Make ordinary CLI mirror `runApprovedTransportETL`: emit exactly one `ETLRun` for every non-empty persisted terminal run, then return `alreadyReportedExecutionError`. A runtime-ledger failure must not replace the App terminal run; annotate `runtime_recorded:false` plus a safe diagnostic/nonzero classification.

**Exact regression test:** `TestCLI_OrdinaryETLFailurePublishesOneTerminalRun` must cover failed, parked, CDC-failed, pre-rename no-commit, post-rename directory-sync error, and runtime-recorder failure. Assert one valid JSON object, persisted ID/status/results, no second Error envelope, secret-free stderr, and the original nonzero exit category.

### ORCH-PF-B10 — Reverse/direct-write finalization publishes a terminal run even when that run was not persisted

**Evidence:** `finishReverseWriteWithErrorText` constructs a terminal run and calls `updateState` (`internal/app/app.go:3149-3181`), but checks `writeErr` before `persistErr` (`:3182-3191`). Therefore a provider error plus a definite pre-rename save failure returns the fabricated nonzero run and only the provider error. Both reverse CLI publication paths treat any non-empty run as persisted and suppress the following Error envelope (`internal/cli/cli.go:1831-1847,2179-2195`).

**Cross-file path:** approved reverse/direct/binary write -> provider returns partial/error result -> App terminal state save fails -> App returns in-memory run -> CLI emits `ReverseRun` and marks the error already reported -> plan remains executing/uncertain and no run is inspectable on disk.

**Failure mode:** CLI claims an inspectable terminal run that does not exist and hides the persistence failure. With an indeterminate post-rename outcome, App likewise does not reload to determine which plan/run state is real.

**Six-surface impact:** ETL=none; reverse ETL=blocker; direct read=none; direct write=blocker because generated writes use reverse finalization; binary download=none; binary upload=blocker through the same write/plan path.

**Exact fix:** handle `persistErr` first. Definite no-commit must return a zero run and `errors.Join(writeErr, persistErr)` while preserving the plan's uncertain/executing recovery state. May-have-committed must reload the exact plan/run and return a non-empty run only when the terminal transition is present; preserve both errors and the original exit class.

**Exact regression test:** `TestReverseFinalization_DoesNotPublishUnpersistedRun` must inject a provider partial/error and a pre-rename store failure, assert zero returned run/no stored run/plan uncertain, and assert CLI emits one Error rather than `ReverseRun`. A post-rename sync failure must reload and emit exactly the durable terminal run if present.

### ORCH-PF-B11 — Durable authentication fencing does not stop already-admitted work across processes or ambiguous CAS commits

**Evidence:** `AuthCohortCoordinator` explicitly owns members only within one process (`internal/coordination/auth_cohort.go:121-129`). `AuthCohortRuntime.Execute` checks durable health once, before the whole operation (`:153-171`), although `AuthCohortAdmission.Check` is documented as the send boundary and reloads durable health (`:294-325`). Engine and PostgreSQL wrap whole multi-request operations with `Execute`, not each provider send (`internal/connectors/engine/connector.go:166-190,222-239,289-296,344-445`; `internal/connectors/native/postgres/connection.go:457`). Separately, file CAS sets `swapped=true` inside the update callback (`internal/coordination/durable_store.go:97-117`); a post-rename sync/unlock error can mean the fence committed (`internal/state/store.go:276-315`), but `Report`, `Repair`, and `Fence` return before local cancellation whenever CAS reports any error (`internal/coordination/auth_cohort.go:364-377,410-423,454-467`).

**Cross-file path:** process A admits a paginated/write operation -> process B persists a verified-invalid fence or repair epoch -> A's context is not in B's member map and A never rechecks durable epoch before subsequent sends. In the ambiguous-CAS variant, even same-process member cancellation is skipped although disk already contains the new fence.

**Failure mode:** after verified credential invalidation/revocation, an already admitted operation can keep issuing provider requests with the invalid credential. That defeats the durable cohort fence specifically during the window it is meant to close and can continue mutations/downloads/uploads after operator repair or revocation.

**Six-surface impact:** ETL=blocker for multi-page/source and destination requests; reverse ETL=blocker; direct read=blocker for paginated/multi-send operations; direct write=blocker; binary download=blocker for any subsequent request/retry boundary; binary upload=blocker.

**Exact fix:** install an admission token/checker at the actual requester/database send boundary and check durable epoch/fence before every HTTP request, page, retry, upload, download, query, and statement. A successful repair/fence must invalidate generations across processes. Change CAS APIs to return a typed commit outcome; on may-have-committed reload exact health and always cancel local old-epoch members when persisted state advanced.

**Exact regression test:** `TestAuthFence_StopsNextSendAcrossCoordinators` must use two coordinators sharing one file store: A completes send/page one, B fences or repairs, and A must make zero later sends. Repeat with injected post-rename directory-sync error and two local admissions, asserting both contexts cancel and future admission sees the persisted state. Cover read, write, download, upload, PostgreSQL query, and repair epoch rollover.

### ORCH-PF-B12 — Binary no-overwrite publication exposes a placeholder and can overwrite or delete a foreign file

**Evidence:** `streamBinaryDownloadToRoot` creates the final path as a visible zero-byte `O_EXCL` reservation (`internal/connectors/engine/binary_read.go:461-485`), then streams to a temp file. Error cleanup blindly removes the final path (`:487-500,510-535`), and `root.Rename(tempName, fileName)` replaces whatever occupies that path by publication time (`:530-537`). There is no containing-directory fsync after rename.

**Cross-file path:** binary command -> provider response -> visible final placeholder -> temp stream/fsync -> ordinary rename over final -> return receipt.

**Failure mode:** process death leaves a zero-byte final file that looks complete and blocks retry. If another process/user removes the placeholder and installs a real file, success overwrites it despite `allow_overwrite=false`; any later error can delete it. A crash after rename can also lose the directory entry because it was never synced.

**Six-surface impact:** ETL=none; reverse ETL=none; direct read=none; direct write=none; binary download=blocker/data loss; binary upload=none.

**Exact fix:** write and fsync only a hidden temp, then publish with an atomic no-replace primitive (`renameat2(RENAME_NOREPLACE)`, `linkat`-based abstraction, or platform-equivalent under `os.Root`). Never expose a reservation placeholder and never clean up the final path. Remove only the owned temp and fsync the containing directory after publication.

**Exact regression test:** `TestBinaryDownload_NoOverwritePublicationIsCrashAndRaceSafe` must kill a subprocess while streaming and assert no final file, install a foreign final file just before publish and assert it survives byte-for-byte while the download fails, cover error cleanup, and inject/observe directory-sync completion after a successful publish.

### ORCH-PF-B13 — CDC receipt recovery accepts any regular WAL/table files as proof of the staged transaction

**Evidence:** PostgreSQL's durable stage retains transaction key, record count, and content digest (`internal/connectors/database/transaction_stage.go:221-270`), but recovery reduces that evidence to receipt ID/sink/time before calling App (`internal/connectors/native/postgres/cdc_v2.go:271-285`). App's receipt ID is derived only from connection ID and transaction ID (`internal/app/change_capture_dispatch.go:377-379`). `RestoreCDCTransactionReceipt` checks that the ID/sink match and that the shared WAL and table paths are regular files; it does not bind or hash their content, generation, record count, or transaction (`internal/app/change_capture_dispatch.go:256-287`). It then allows checkpoint commit.

**Cross-file path:** PostgreSQL source stage has receipt -> process restarts -> App opens shared stream paths -> any two regular files pass restore -> source commits checkpoint/acknowledges LSN -> stage retires.

**Failure mode:** a stale concurrent writer, accidental replacement, or corruption can leave unrelated regular files at both paths. Recovery accepts them as the original durable transaction and advances the source checkpoint, permanently losing the transaction whose receipt was restored.

**Six-surface impact:** ETL=blocker/data loss for change capture; reverse ETL=none; direct read=none; direct write=none; binary download=none; binary upload=none.

**Exact fix:** extend the CDC restoration contract with immutable transaction key, record count, and content digest. Atomically persist a connection/stream/generation-owned warehouse receipt manifest containing those values plus WAL/final artifact digests. Recovery must verify that exact manifest and artifacts before manufacturing acknowledgement; mismatch must enter typed reconciliation/rebootstrap without advancing LSN.

**Exact regression test:** `TestCDCRecovery_ReceiptBindsExactWarehouseArtifacts` must crash after the source stage records the receipt, replace WAL/final with valid unrelated regular files, restart, and assert no checkpoint/LSN acknowledgement plus a typed recovery error. Untouched exact artifacts must restore without receiving/writing the transaction twice; include stale-writer generation mismatch and truncated-file cases.

### ORCH-PF-B14 — Errors after checkpoint commit convert delivered work into a failed run

**Evidence:** ordinary checkpoint commit precedes receipt retirement, whose error is returned (`internal/synctransport/orchestrator.go:277-290,525-537`); full overwrite likewise commits then can fail retirement (`:492-502`). After the orchestrator returns with a committed pointer, App performs managed-target/declarative plan markers and returns their storage errors (`internal/app/transport_dispatch.go:357-390`; marker implementations at `internal/app/declarative_typed_destination_approval.go:346-368` and `internal/app/postgres_transport_approval.go:283-308`). `dispatchETLMode` sends any such error to `failAcknowledgedTransportRun` (`internal/app/etl_mode_dispatch.go:66-75`), which persists status `failed` even though provider state and checkpoint are durable (`internal/app/app.go:1612-1629,1710-1760`).

**Cross-file path:** provider apply/read-back -> checkpoint CAS succeeds -> cleanup or approval marker save fails -> result returns error with committed state -> terminalizer records the run as failed.

**Failure mode:** an operator sees a failed run even though the requested data and checkpoint are committed. Retry starts after the checkpoint and may produce an empty success, so neither run truthfully describes delivery; plan state may remain `approval_consumption_uncertain` after completed external effects.

**Six-surface impact:** ETL=blocker for every staged/approved transport; reverse ETL=approval plan lifecycle is misleading when used as a transport destination; direct read=none; direct write=standalone finalization is covered by ORCH-PF-B10; binary download=none; binary upload=standalone uses reverse finalization.

**Exact fix:** model checkpointed delivery as a durable phase distinct from cleanup/marker reconciliation. Atomically persist checkpoint, run delivery phase, and plan consumption where possible; otherwise return a `delivered_reconciliation_required` terminal state and never downgrade it to failed or replay provider work. Restart reconciliation must finish retirement/marker idempotently from the committed receipt.

**Exact regression test:** `TestTransport_PostCheckpointBookkeepingFailureRemainsDelivered` must inject stage-retire, declarative marker, and managed-target marker failures after checkpoint commit. Assert one provider mutation, retained checkpoint/output, a delivered/reconciliation terminal status (not failed), and restart repair with zero source/provider replay.

## Warnings

### ORCH-PF-W01 — File-backed rate-parking mutations do not reconcile indeterminate commit outcomes

**Evidence:** file store methods return callback-local flags/results plus the raw JSON store error for create, rearm, claim, phase, renew, complete, and delete (`internal/coordination/durable_store.go:194-253,269-306,317-373,401-437`). A post-rename directory-sync/unlock error means the mutation may already be on disk (`internal/state/store.go:276-315`). Coordinator `Park` returns before adding the committed record/timer on any create error (`internal/coordination/rate_parking.go:541-578`); `Cancel` retains memory after a possibly committed delete (`:650-673`); completion retries after a possibly committed removal (`:800-816`).

**Cross-file path:** coordinator transition -> file store JSON rename -> directory-sync error -> callback flags discarded -> live maps/timers diverge from disk until restart.

**Failure mode:** a durable parked scope can have no timer in the live process, blocking same-scope admission indefinitely; a durable completion/delete can leave an in-memory retry loop for a record that no longer exists.

**Six-surface impact:** ETL/reverse ETL=availability/retry robustness; direct read/direct write/binary download/binary upload=affected only when routed through shared rate parking.

**Exact fix:** make every store mutation return typed commit outcome and reload/reconcile exact durable state on may-have-committed before changing timers/maps or choosing retry.

**Exact regression test:** `TestRateParking_IndeterminateMutationReconcilesLiveState` must inject post-rename sync failures for Create, Rearm, Claim, MarkResumeCompleted, Complete, and Delete, comparing the live coordinator with a reopened store and asserting exactly one correct timer/no churn.

### ORCH-PF-W02 — Expired or revoked authorization puts a parked run into an unbounded rearm loop

**Evidence:** write plans expire after 24 hours (`internal/app/write_approval.go:21-24`), and per-unit authorization rejects expired/revoked records (`internal/app/authorization.go:248-282`; `internal/app/declarative_typed_destination_approval.go:194-214`). Exact provider `Retry-After` can outlive that (`internal/connectors/connsdk/http.go:1402-1416`). Rearm reconstructs the original plan ID (`internal/app/durable_coordination.go:178-235,338-347`); failed attempt state remains linked and retryable (`:255-279`), while the coordinator retries every resume error without terminal classification (`internal/coordination/rate_parking.go:780-790`). No App/CLI production caller exposes `RateParkingCoordinator.Cancel`.

**Cross-file path:** provider parks beyond authorization expiry -> timer claims -> new ETL attempt -> authorization fails before provider I/O -> attempt persisted failed -> coordinator rearms forever.

**Failure mode:** the parked record and provider scope never clear, failed attempt records accumulate, and an operator has no supported reapprove/takeover/cancel path.

**Six-surface impact:** ETL/reverse ETL=stuck resume; direct read/direct write/binary download/binary upload=same provider scope can remain blocked by the parked record.

**Exact fix:** classify authorization expiry/revocation as a durable terminal `needs_reauthorization` reconciliation state, stop automatic retries, and expose a safe reapproval takeover or cancel/abandon command bound to the exact scope/checkpoint.

**Exact regression test:** `TestRateParking_ExpiredAuthorizationStopsAndCanRecover` must park beyond 24h and separately revoke before reset, assert one needs-auth event, zero provider sends/unbounded attempts, then prove exact-scope reapproval or cancellation unblocks admission.

### ORCH-PF-W03 — Route selection still hides declared-route preflight errors as declaration absence

**Evidence:** after confirming both endpoints declare transports, `selectTransportRoute` returns `transportRouteDeclarationAbsent` when a declarative source's destination executor is unregistered or lacks one of three semantic marker types (`internal/app/transport_dispatch.go:52-75`). It therefore never invokes registry preflight, whose exact error would identify the missing executor/binding/conformance (`internal/synctransport/registry.go:145-225`). `dispatchETLMode` later replaces that with a generic no-matching-transport error for contract modes (`internal/app/etl_mode_dispatch.go:51-95`).

**Cross-file path:** valid source declaration + invalid/unsupported destination declaration -> early semantic filter -> `declaration_absent` -> generic contract refusal instead of exact preflight reason.

**Failure mode:** the route fails closed, but diagnostics falsely say no route was declared and conceal the actionable broken executor/adapter/conformance configuration. This is a remaining partial closure of CF-W06.

**Six-surface impact:** ETL=diagnostic and operability degradation; reverse ETL=destination route diagnosis; direct read/direct write/binary download/binary upload=none.

**Exact fix:** once both declarations exist, always run registry preflight and return its typed error. Apply semantic opt-in only after successful preflight and use a distinct typed `declared_but_not_selected` reason when intentional legacy routing is allowed.

**Exact regression test:** `TestETLRouteSelection_PreservesDeclaredPreflightError` must cover unregistered executor, wrong adapter marker, missing source binding/action, and rejected conformance, asserting the exact error and zero catalog/source/provider I/O; retain a separate two-sided-absent legacy case.

## Prior-Finding Closure Table (Orchestration/Delivery Lens)

| Prior ID | Prior claim in `REVIEW-FIX.md` | Frozen-HEAD assessment | Evidence/current finding |
|---|---|---|---|
| CF-B05 | Binary upload mapped to upload origin/executor | **Closed in this lens** | GitHub declaration/CLI/write preparation retain binary upload binding and preview digest; targeted engine/CLI tests passed. |
| CF-B06 | Authorization/CAS checked before I/O and checkpoint | **Partial / open** | Per-unit authorization and executed-marker idempotence improved, but post-checkpoint marker/cleanup failures still downgrade delivery: ORCH-PF-B14. |
| CF-B07 | Provider-owned bounded read-back | **Partial / open** | Provider I/O now exists, but receipt is ignored, a collection prefix is scanned, numeric equality is representation-sensitive, and deadline is shared: ORCH-PF-B04/B06/B07. |
| CF-B08 | Public receipts mask configured credentials | **Closed in this lens** | Public sanitizer is used at destination/direct output boundaries; no regression found in the traced state machines. |
| CF-B09 | Persisted failed runs rendered with nonzero status | **Partial / open** | Approved/reverse paths improved; ordinary ETL still returns before publication and reverse can publish an unpersisted run: ORCH-PF-B09/B10. |
| CF-B11 | Complete REST/GraphQL/binary failure receipts | **Closed in this lens** | CLI publishes direct-read/download receipt-bearing failures once and retains categorized nonzero exit. |
| CF-B16 | Safe REST/binary diagnostics | **Closed in this lens** | Shared bounded status-only error text and receipt output are preserved. |
| CF-B24 | Ownership CAS precedes provider side effects | **Open; claimed fix absent** | ORCH-PF-B01. `933afaa25` does not alter App claim order and current tests require losing side effects. |
| CF-B25 | Published overwrite is not aborted on read-back failure | **Closed** | `published=true` is set immediately after successful publish in ordinary and both Arrow controllers. |
| CF-B26 | Full-overwrite/Arrow outputs propagate | **Closed** | Shared `collectDestinationResult` runs before later verification/checkpoint work. |
| CF-B27 | Delivery guarantees/tombstones enforced | **Partial / open** | Compatibility/tombstone checks exist, but keyed delivery and conformance can be self-declared without provider proof: ORCH-PF-B03. |
| CF-W04 | Transient parking errors retain retries | **Partial / open** | Nominal retry timer is restored; indeterminate store outcomes and terminal authorization failures still wedge/retry: ORCH-PF-W01/W02. |
| CF-W05 | Same-scope resumes serialized | **Closed** | Durable leader/claim ordering and cross-scope concurrency are present and coordination race tests passed. |
| CF-W06 | Typed route reasons preserved | **Partial / open** | Two-sided declared routes can still be relabeled absent before preflight: ORCH-PF-W03. |
| CF-W09 | Store APIs reject unloadable state | **Closed as originally scoped** | Next-state validation/claim input validation exists; commit-outcome reconciliation is the distinct ORCH-PF-W01. |

## Six-Surface Assessment

| Surface | Traced terminal/delivery path | Assessment |
|---|---|---|
| ETL | CLI -> App run -> route/authorization -> ordinary/full/Arrow/CDC -> checkpoint -> terminal run | **Blocked** by ownership, pagination, delivery proof, read-back, cloning, deadline, terminal publication, CDC recovery, and post-checkpoint phase defects. |
| Reverse ETL | Preview/token -> provider write/result -> App terminal plan/run -> CLI envelope/nonzero | **Blocked** by unpersisted terminal publication, auth fencing, and transport-destination issues. |
| Direct read | Declaration/command preflight -> bounded REST/GraphQL -> complete receipt -> CLI | Receipt/nonzero path passes; **blocked at shared auth fencing** for already-admitted multi-send work. |
| Direct write | Declaration-owned action/approval -> provider result -> reverse terminal run | No-input reachability and hollow-JSON rejection pass; **blocked** by reverse finalization and shared auth fencing. |
| Binary download | Root-confined stream -> file publish -> complete receipt -> CLI | Receipt/size/stall/traversal checks pass; **blocked** by unsafe no-overwrite publication and shared auth fencing. |
| Binary upload | Declaration-owned binary file/digest -> approval -> provider upload -> reverse terminal run | Origin/media/digest plumbing passes; **blocked** by reverse finalization and shared auth fencing. |

## Verified Passing Behaviors

- Runtime pages validate tombstones and refuse destinations without tombstone delivery (`internal/synctransport/orchestrator.go:150-163`); full overwrite refuses tombstones (`:386-394`).
- Full-overwrite and Arrow controllers mark publication successful before read-back, so a verification failure does not call pre-publication abort.
- Destination output is defensively copied into results before read-back/checkpoint errors (`internal/synctransport/orchestrator.go:510-523`).
- Duplicate, missing, and nonnumeric mismatched provider identities are rejected by the current generic matcher.
- Direct read and binary-download receipt-bearing failures are emitted once and preserve nonzero exit (`internal/cli/cli.go:1329-1357`).
- Declaration-owned no-input/no-body actions are reachable, while hollow JSON operations and record-bound empty schemas are rejected (`internal/connectors/engine/connector.go:481-515`; `internal/connectors/engine/record_schema_promotion_test.go:94-147`).
- Nominal rate-parking retry timer ownership and same-scope leadership pass targeted/race tests; the warnings concern failure-outcome branches not exercised by those tests.

## Final Verdict

**Verdict: blockers.** The frozen rollup must not ship until all 14 blockers are fixed and their exact behavioral tests pass at the reviewed source SHA. In particular, CF-B24, CF-B07, CF-B09, and CF-B27 must not remain recorded as closed on the current evidence.

---

_Reviewed: 2026-08-21T02:52:36Z_
_Reviewer: gsd-code-reviewer (independent post-fix orchestration/delivery lens)_
_Depth: deep_
