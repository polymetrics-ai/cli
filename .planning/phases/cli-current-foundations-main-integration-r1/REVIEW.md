---
phase: cli-current-foundations-main-integration-r1
reviewed: 2026-08-20T18:02:34Z
depth: deep
source_sha: 9e5329f34e015e39160bb8e951452bbd071a698a
source_ledgers:
  - .planning/phases/cli-current-foundations-main-integration-r1/reviews/MAPPING-REVIEW.md
  - .planning/phases/cli-current-foundations-main-integration-r1/reviews/RUNTIME-REVIEW.md
  - .planning/phases/cli-current-foundations-main-integration-r1/reviews/ORCHESTRATION-REVIEW.md
files_reviewed: 77
files_reviewed_list:
  - cmd/connectorgen/main_test.go
  - cmd/connectorgen/validate.go
  - internal/app/app.go
  - internal/app/authorization.go
  - internal/app/declarative_typed_destination_approval.go
  - internal/app/durable_coordination.go
  - internal/app/etl_mode_dispatch.go
  - internal/app/foundations_integration_test.go
  - internal/app/issue_label_transport_approval.go
  - internal/app/issue_label_warehouse_transport.go
  - internal/app/local_warehouse.go
  - internal/app/rest_write_command_test.go
  - internal/app/transport_composition_test.go
  - internal/app/transport_dispatch.go
  - internal/app/transport_dispatch_test.go
  - internal/app/types.go
  - internal/cli/cli.go
  - internal/cli/cli_test.go
  - internal/cli/docs.go
  - internal/cli/etl_transport.go
  - internal/cli/etl_transport_test.go
  - internal/cli/golden_transcript_test.go
  - internal/cli/skills.go
  - internal/cli/structured_rest_body_help_test.go
  - internal/cli/testdata/golden_transcripts.json
  - internal/connectors/commandrunner/runner.go
  - internal/connectors/commandrunner/runner_test.go
  - internal/connectors/connectors.go
  - internal/connectors/connsdk/http.go
  - internal/connectors/connsdk/http_test.go
  - internal/connectors/connsdk/stream.go
  - internal/connectors/engine/bundle.go
  - internal/connectors/engine/bundle_test.go
  - internal/connectors/engine/connector.go
  - internal/connectors/engine/connector_test.go
  - internal/connectors/engine/direct_read.go
  - internal/connectors/engine/direct_write.go
  - internal/connectors/engine/direct_write_multipart_test.go
  - internal/connectors/engine/direct_write_test.go
  - internal/connectors/engine/graphql_operation.go
  - internal/connectors/engine/graphql_operation_test.go
  - internal/connectors/engine/interpolate.go
  - internal/connectors/engine/interpolate_test.go
  - internal/connectors/engine/prepared_write.go
  - internal/connectors/engine/rate_limit_parking.go
  - internal/connectors/engine/read.go
  - internal/connectors/engine/record_schema_promotion.go
  - internal/connectors/engine/record_schema_promotion_test.go
  - internal/connectors/engine/schema.go
  - internal/connectors/engine/schema_test.go
  - internal/connectors/engine/structured_rest_body.go
  - internal/connectors/engine/structured_rest_body_test.go
  - internal/connectors/engine/write.go
  - internal/connectors/engine/write_query_test.go
  - internal/connectors/engine/write_test.go
  - .planning/phases/issue-4305-rest-structured-body-r1/RUN-STATE.json
  - .planning/phases/synctransport-4303-destination-r1/RUN-STATE.json
  - data/cli-current-foundations-main-integration-r1/evidence-manifest.json
  - data/cli-current-foundations-main-integration-r1/input-manifest.json
  - internal/connectors/sync_transport.go
  - internal/connectors/sync_transport_test.go
  - internal/connectors/transportpolicy/policy.go
  - internal/connectors/write_result_output_test.go
  - internal/coordination/durable_store.go
  - internal/coordination/rate_parking.go
  - internal/coordination/rate_parking_test.go
  - internal/synccontract/commit.go
  - internal/synctransport/arrow_fast_path_controller.go
  - internal/synctransport/arrow_fast_path_pipeline.go
  - internal/synctransport/orchestrator.go
  - internal/synctransport/registry.go
  - internal/synctransport/types.go
  - website/content/docs/cli-reference.mdx
  - website/content/docs/etl.mdx
  - website/lib/docs.generated.ts
  - website/scripts/gen-docs-data.mjs
  - website/scripts/gen-docs-data.test.mjs
findings:
  critical: 27
  warning: 9
  info: 0
  total: 36
raw_ledger_finding_ids: 37
raw_atomic_claims: 41
status: issues_found
---

# Foundation Rollup: Canonical Code Review

**Reviewed:** 2026-08-20T18:02:34Z
**Depth:** deep
**Frozen source:** `9e5329f34e015e39160bb8e951452bbd071a698a`
**Status:** issues_found — merge blocked

## Summary

The frozen source is not releasable. The three complete lens ledgers contain 37 finding IDs. Four IDs contain two independently testable atomic claims, producing the required 41-row raw crosswalk. Correlation found one true cross-lens duplicate: `MAP-BL-08` and `RT-B01` describe the same no-op public write-result sanitizers and the same credential-disclosure boundary. The count transformation is exact: **41 atomic claims - 4 compound-ID regroupings - 1 cross-lens duplicate = 36 canonical findings (27 BLOCKER, 9 WARNING)**.

No production or test source was changed during synthesis. At intake and immediately before this artifact was written, `HEAD` was the exact frozen SHA; tracked source and index were clean; only the three expected lens ledgers were untracked.

## Narrative Findings (AI reviewer)

### Source-ledger coverage

| Ledger | Reviewed scope | Raw IDs | Atomic claims | Canonical coverage | Result |
| --- | ---: | ---: | ---: | --- | --- |
| Mapping | 25 files | 12 (11 blocker, 1 warning) | 14 (13 blocker, 1 warning) | `CF-B01`–`CF-B11`, `CF-W01` | Complete; `MAP-BL-08` overlaps runtime. |
| Runtime | 30 files | 15 (13 blocker, 2 warning) | 16 (14 blocker, 2 warning) | `CF-B08`, `CF-B12`–`CF-B23`, `CF-W02`–`CF-W03` | Complete. |
| Orchestration | 22 files | 10 (4 blocker, 6 warning) | 11 (4 blocker, 7 warning) | `CF-B24`–`CF-B27`, `CF-W04`–`CF-W09` | Complete. |
| **Total** | **77 files** | **37 IDs** | **41 claims** | **36 unique findings** | **27 blockers, 9 warnings.** |

The atomic expansion is explicit, not inferred new review scope: `MAP-BL-01`, `MAP-BL-11`, `RT-B03`, and `ORCH-W09` each contain two separately testable claims. The 41 rows are 31 blocker claims plus 10 warning claims. Regrouping the three compound blocker pairs and one compound warning pair yields the ledgers' 28 blocker + 9 warning IDs (37). De-duplicating blocker IDs `MAP-BL-08`/`RT-B01` yields the canonical 27 blocker + 9 warning findings (36). The crosswalk at the end preserves both the parent source ID and the atomic suffix.

### Merge disposition

**BLOCKED.** All `CF-B*` findings must be fixed and verified in the ordered wave below before merge. Warnings do not independently block merge, but should be included in the same wave because they touch the same generator, request-boundary, coordination, and evidence surfaces.

### Six-surface rollup

| Surface | Frozen-source disposition | Primary blockers |
| --- | --- | --- |
| ETL | Blocked: destination read-back can be local-only, authorization bookkeeping can fail after durable effects, stale writers mutate before ownership CAS, and declared delivery guarantees are not enforced. | `CF-B06`, `CF-B07`, `CF-B24`–`CF-B27` |
| Reverse ETL | Blocked: provider inputs are missing, public receipts can leak credentials or disappear, idempotent writes can redirect/retry outside approval, and execution can rematerialize after approval. | `CF-B03`, `CF-B08`–`CF-B10`, `CF-B13`, `CF-B15`, `CF-B17`–`CF-B21`, `CF-B24`–`CF-B27` |
| Direct read | Blocked: GraphQL selections/arguments are narrowed, REST bindings are open, complete receipts are absent, error formatting leaks provider text, and name heuristics delete ordinary IDs. | `CF-B04`, `CF-B11`, `CF-B14`, `CF-B16`, `CF-B19`–`CF-B23` |
| Direct write | Blocked: public output can disclose credentials, terminal receipts are lost in multiple error classes, and exact typed input validation has precision/duplicate/boundary gaps. | `CF-B08`–`CF-B10`, `CF-B12`, `CF-B18`–`CF-B21` |
| Binary download | Successful file confinement/integrity passes, but failure receipts are narrowed and unsafe error reconstruction can disclose provider/query secrets. | `CF-B11`, `CF-B14`, `CF-B16`, `CF-B20`–`CF-B21` |
| Binary upload | Blocked: the installed GitHub upload is an implemented JSON action to the wrong origin with no file or required `name`; the distinct `file_upload` executor remains an explicit gap. | `CF-B01`–`CF-B05`, `CF-B20`–`CF-B21` |

### Provider output versus secret handling — canonical contract

The correct contract has two representations and one-way projection:

1. The internal engine receipt remains complete: operation identity, response presence, status, ordered repeatable headers, body presence/byte count/raw encoding, decoded value, GraphQL envelope, provider IDs/occurrence IDs, and typed error cause.
2. Persisted and printable App/CLI output is a deep clone of that receipt. It preserves every ordinary field and value, including fields named `token`, large integers, unfamiliar keys, duplicate headers, invalid UTF-8 via explicit encoding, and provider extensions.
3. The public clone masks exact selected/configured credential values and proven encodings everywhere they occur, and masks declaration-classified secret response fields/headers even when their values differ from configured credentials. The field remains present with an explicit masked marker and safe byte/presence metadata.
4. Field-name substrings are never authority to delete output. Provider bodies and raw query strings are never reconstructed into printable diagnostics. Typed causes remain inspectable internally.

This resolves the apparent ledger tension: “complete provider output” means complete ordinary provider truth, not verbatim disclosure of known credentials.

### Source lock, mapping, and generated-surface contract

The source lock must become the root of the installed surface rather than an isolated report:

`strict REST+GraphQL source lock -> canonical descriptor -> operations/actions/API surface -> CLI flags/help/manual -> website/skills -> runtime preflight/certification`.

Every source identity and request/response field must have exactly one generated projection or a typed, source-identity-bound gap with reason and owner. `source-import --check`, `validate`, `surface-sync --check`, installed-command reachability, docs parity, and skill generation must all fail on missing/stale projections. Endpoint presence alone is not coverage. Connector-name allowlists, silently ignored source-lock sections, and hand-authored generated fields are prohibited.

## Critical Issues

### CF-B01 — Authoritative source import is nonfunctional and silently ignores GraphQL

**Severity:** BLOCKER

**Raw:** `MAP-BL-01a`, `MAP-BL-01b`

**Evidence / affected:** `cmd/connectorgen/sourceimport.go:109-113` (`sourceImportLock`) models only `rest`; `parseSourceImportLock` at `:304-324` accepts the enriched lock without modeling `graphql`; `sourceReferenceIndexByteLimit` at `:1103-1121` reuses the artifact-byte ceiling and the exact checked-in GitHub lock fails with `source grammar position byte limit exceeded`. Affected lock: `internal/connectors/defs/github/sources/github-operation-source-lock.json`.

**Required change:** Define a strict versioned REST+GraphQL lock schema with unknown-field rejection; import all 1,220 REST, 31 query, and 274 mutation identities plus GraphQL arguments/types; give index accounting its own measured bound; certify the exact checked-in bytes/SHA.

**Required tests:** `TestSourceImport_CheckedInGitHubLockCoversRESTAndGraphQL` must assert 1,525 identities and exact source locations; `TestSourceImport_RejectsUnknownSectionAndIndependentIndexOverflow` must reject unknown lock members, SHA/byte drift, duplicates, and measured index overflow before output.

**Six-surface impact:** ETL=source projection untrusted; reverse ETL=blocked; direct read=blocked; direct write=blocked; binary download=projection untrusted; binary upload=blocked.

### CF-B02 — Imported descriptors are orphaned from generation and validation

**Severity:** BLOCKER

**Raw:** `MAP-BL-02`

**Evidence / affected:** `runSourceImportWithFetcher` writes arbitrary `--out` only (`cmd/connectorgen/sourceimport.go:6230-6283`); `validateDir` validates existing bundle directories only (`cmd/connectorgen/validate.go:225-301`); `syncBundle` iterates existing CLI commands/operations (`cmd/connectorgen/surfacesync.go:268-380`); `params-import` silently skips absent source path/method pairs (`cmd/connectorgen/paramsimport.go:183-217`). Exact check reported 211 drifted operations while a temporary updated bundle still validated cleanly.

**Required change:** Check in a canonical descriptor per provider lock; generate all operation/action/API/CLI/help inputs from it; add field-complete source identity validation and a typed gap ledger; make validation and drift checks fail for missing or stale projections.

**Required tests:** `TestSourceProjection_AddChangeDeletePropagatesToEverySurface`; `TestSourceProjection_MissingOperationOrFieldFailsValidateAndSurfaceCheck`; include alias/semantic-action union cases without double counting.

**Six-surface impact:** ETL=unproven source; reverse ETL=blocked; direct read=blocked; direct write=blocked; binary download=unproven; binary upload=blocked.

### CF-B03 — Reverse-ETL actions omit provider inputs and advertise invalid requests

**Severity:** BLOCKER

**Raw:** `MAP-BL-03`

**Evidence / affected:** `cmd/connectorgen/validate.go:1664-1753` checks only authored `record_schema`/`path_fields`; `internal/connectors/defs/github/writes.json:2687-2699` omits required `access_level`, `:2777-2789` omits `name`, `runner_group_id`, and `labels`; exact audit found 281 omitted body fields across 99 endpoints, including 99 required fields across 65 endpoints, plus required query omissions.

**Required change:** Generate closed typed query/body schemas and CLI flags from provider source; validate action-union coverage for placement, requiredness, type, enum, nullability, and nesting; mark incomplete commands unavailable with source-bound gaps.

**Required tests:** `TestInstalledReverseActions_CoverProviderRequestContract`; `TestInstalledReverseActions_RequiredFieldRemovalFailsBeforeIO`; include optional/nested/nullable and multi-action union cases.

**Six-surface impact:** ETL=destination actions affected; reverse ETL=blocked; direct read=none; direct write=operation-backed path unaffected; binary download=none; binary upload=contributes to broken upload mapping.

### CF-B04 — Generated GraphQL operations expose placeholder results and incomplete pagination

**Severity:** BLOCKER

**Raw:** `MAP-BL-04`

**Evidence / affected:** 303 of 305 generated GraphQL root operations select only `__typename`; five queries omit locked `before`/`last` arguments. Example: `internal/connectors/defs/github/operations.json:8877-8886`; validation at `cmd/connectorgen/validate.go:998-1080` verifies variables/kind but not result coverage.

**Required change:** Generate bounded typed selection sets from locked result types, exhaustive union/interface fragments, every root argument, and mutually exclusive forward/backward pagination; validate documents against the locked schema and require typed gaps for unreachable fields.

**Required tests:** `TestGeneratedGraphQLOperations_SelectDeclaredFieldsAndArguments`; `TestGeneratedGraphQLOperations_RejectTypenameOnlyOrMissingPagination`; cover scalar roots, unions/interfaces, deprecation/preview metadata, and partial-data envelopes.

**Six-surface impact:** ETL=GraphQL sources narrowed; reverse ETL=GraphQL mutations narrowed; direct read=blocked; direct write=blocked for returned payload completeness; binary download=none; binary upload=none.

### CF-B05 — Installed GitHub binary upload is a JSON request to the wrong origin

**Severity:** BLOCKER

**Raw:** `MAP-BL-05`

**Evidence / affected:** Provider operation `repos/upload-release-asset` requires `uploads.github.com`, query `name`, optional `label`, and `application/octet-stream`; `internal/connectors/defs/github/writes.json:5750-5770` declares JSON with no file/query/server; `cli_surface.json:10448-10469` advertises implemented `releases assets view-3` with only `release-id`.

**Required change:** Model a bounded root-confined binary upload with required `name`, optional `label`, pinned upload origin/media type, preview-bound file identity/digest, accurate command name/help, no redirects, and no retry without proven idempotency.

**Required tests:** `TestGitHubReleaseAssetUpload_InstalledCommandSendsExactBytes`; `TestGitHubReleaseAssetUpload_RejectsMissingChangedUnsafeOrOversizeFile`; cover empty files, URL encoding, no-response failures, and complete masked 4xx receipts.

**Six-surface impact:** ETL=none; reverse ETL=misclassified path; direct read=none; direct write=approval machinery reusable; binary download=none; binary upload=blocked.

### CF-B06 — Reusable typed-destination authorization can fail after write and checkpoint

**Severity:** BLOCKER

**Raw:** `MAP-BL-06`

**Evidence / affected:** `internal/app/declarative_typed_destination_approval.go:165-172` accepts durable authorization reuse; `markDeclarativeTypedDestinationPlanExecuted` at `:339-359` rejects already executed plans; `internal/app/transport_dispatch.go:351-359` calls it after committed provider/checkpoint effects.

**Required change:** Separate reusable authorization from per-run completion; make same-binding executed state idempotent; ensure no post-checkpoint bookkeeping converts successful external effects into failed commands; reconcile interrupted marker updates.

**Required tests:** `TestDeclarativeTypedDestinationAuthorization_ReusesWithoutPostEffectFailure`; `TestDeclarativeTypedDestinationAuthorization_RejectsChangedOrForeignBindingBeforeIO`; include zero-record, concurrent, interruption, and rate-limit resume cases.

**Six-surface impact:** ETL=blocked; reverse ETL=destination execution affected; direct read=none; direct write=none; binary download=none; binary upload=none.

### CF-B07 — Generic destination read-back never reads the provider

**Severity:** BLOCKER

**Raw:** `MAP-BL-07`

**Evidence / affected:** `ApplyDestination` performs the write (`internal/app/issue_label_warehouse_transport.go:195-266`), but `ReadBackDestination` at `:297-321` checks only local plan/workset/ack data; `internal/synctransport/orchestrator.go:249-283` treats it as independent proof before checkpoint. Existing composition test expects source GET + destination POST only.

**Required change:** Add declaration-owned provider read-back operation, receipt-to-identity projection, expected-state matcher, bounds, and conformance evidence; perform provider read under timeout before checkpoint; declare a weaker/unavailable guarantee when provider verification is impossible.

**Required tests:** `TestDeclarativeTypedDestination_ReadBackProviderStateBeforeCheckpoint`; `TestDeclarativeTypedDestination_MissingOrMismatchedReadBackDoesNotCommit`; include eventual consistency, rate limit, duplicate IDs, partial batches, and cancellation.

**Six-surface impact:** ETL=blocked; reverse ETL=destination durability affected; direct read=read-back operation dependency; direct write=none; binary download=none; binary upload=none.

### CF-B08 — Public provider receipts disclose configured credentials

**Severity:** BLOCKER

**Raw:** `MAP-BL-08`, `RT-B01` (cross-lens duplicate)

**Evidence / affected:** `SanitizeWriteResultForOutput` and `SanitizeOperationDirectWriteResultForOutput` ignore secrets and return inputs unchanged (`internal/connectors/connectors.go:921-931`); App persists these results (`internal/app/app.go:2924-2929`, `:3149-3155`); `internal/connectors/write_result_output_test.go:12-65`, App tests, docs, and skills explicitly require credential echoes to survive.

**Required change:** Implement the two-representation contract above: immutable complete internal receipt; deep-cloned public receipt masking concrete credentials/proven encodings and declared secret response locations while preserving all ordinary fields, presence, ordering, IDs, and byte metadata. Rewrite unsafe docs/tests.

**Required tests:** `TestWriteResultOutput_PreservesOrdinaryProviderTruth`; `TestWriteResultOutput_MasksConfiguredAndDeclaredSecretsEverywhere`; include overlapping/short/numeric/base64 secrets, repeated headers, nested values, invalid UTF-8, ordinary `token` fields, and input non-mutation.

**Six-surface impact:** ETL=blocked for provider destinations; reverse ETL=blocked; direct read=shared public-receipt policy required; direct write=blocked; binary download=shared policy required; binary upload=must inherit fix.

### CF-B09 — CLI discards persisted failed runs and their receipts

**Severity:** BLOCKER

**Raw:** `MAP-BL-09`

**Evidence / affected:** `runApprovedTransportETL` returns on App error before encoding the non-zero run (`internal/cli/etl_transport.go:495-506`); generated connector writes and `pm reverse run` do the same (`internal/cli/cli.go:1808-1814`, `:2146-2152`). App tests prove failed runs can contain complete provider receipts.

**Required change:** Carry terminal runs in a typed execution error or honor the returned non-zero run; emit exactly one masked JSON terminal envelope before the categorized nonzero error; human output must include safe run ID/status lookup without duplicating provider bodies to stderr.

**Required tests:** `TestCLI_ETLFailureEmitsTerminalRunJSON`; `TestCLI_ReverseFailureEmitsTerminalRunJSON`; cover partial batches, persist failure, no response, writer failure, valid single JSON stdout, and secret-free stderr.

**Six-surface impact:** ETL=blocked failure contract; reverse ETL=blocked; direct read=none; direct write=blocked when invoked through generated writes; binary download=none; binary upload=affected after implementation.

### CF-B10 — Direct-write no-response failures lose attempted operation identity

**Severity:** BLOCKER

**Raw:** `MAP-BL-10`

**Evidence / affected:** `internal/connectors/engine/direct_write.go:92-99`, `:155-185` populates result identity only when response is non-nil; App persists only when `ResponseReceived` is true (`internal/app/app.go:2924-2929`).

**Required change:** Initialize result identity from the sealed prepared request before I/O, set `response_received=false`, add response fields only when present, and persist any attempted-operation result independently of response presence.

**Required tests:** `TestOperationDirectWrite_NoResponseRetainsAttemptIdentity`; `TestOperationDirectWrite_PreflightFailureHasNoAttemptReceipt`; include timeout/cancellation races and normal-response regression.

**Six-surface impact:** ETL=none; reverse ETL=generated direct operations affected; direct read=none; direct write=blocked audit/retry contract; binary download=none; binary upload=must inherit.

### CF-B11 — Direct-read and binary-download receipts omit provider truth and error responses

**Severity:** BLOCKER

**Raw:** `MAP-BL-11a`, `MAP-BL-11b`

**Evidence / affected:** `DirectReadResult` lacks operation/body/presence/raw fields and admits only declared headers (`internal/connectors/connectors.go:458-477`); REST read returns zero result on provider error (`internal/connectors/engine/direct_read.go:145-174`); GraphQL drops paths/locations/extensions (`graphql_operation.go:707-753`); binary read returns zero on provider error (`binary_read.go:160-200`); CLI prints only narrowed views.

**Required change:** Introduce a shared complete response receipt for reads/writes/download metadata; retain full GraphQL envelope; return typed receipt plus nonzero provider error; keep output policy as a convenience projection; binary success reports confined file size/digest without inlining bytes.

**Required tests:** `TestDirectRead_CompleteReceiptOnSuccessAndProviderError`; `TestBinaryDownload_CompleteMetadataAndNoErrorArtifact`; include duplicate headers, empty/absent body, invalid JSON/UTF-8, GraphQL partial data, oversize/media mismatch, exact masking, and large IDs.

**Six-surface impact:** ETL=source observability dependency; reverse ETL=none; direct read=blocked; direct write=none; binary download=blocked failure contract; binary upload=shared receipt model needed.

### CF-B12 — Accepted-success-status mismatch discards terminal direct-write receipt

**Severity:** BLOCKER

**Raw:** `RT-B02`

**Evidence / affected:** `internal/connectors/connsdk/http.go:1210-1212` returns nil response before capture when a 2xx status is outside `AcceptedStatuses`; `internal/connectors/engine/operation_headers.go:425-447` installs those ranges; direct write can therefore retain only the error.

**Required change:** Capture bounded bytes, close body, clone metadata, and return the terminal response together with `UnexpectedStatusError`; keep mismatch terminal/non-retryable and body-free in printable diagnostics.

**Required tests:** `TestRequester_AcceptedStatusMismatchReturnsReceiptAndTypedError`; cover undeclared 202, empty 204, JSON `null`, binary, cap+1, and repeated headers.

**Six-surface impact:** ETL=none; reverse ETL=operation writes affected; direct read=none; direct write=blocked; binary download=none; binary upload=must inherit.

### CF-B13 — Idempotency-enabled reverse writes can redirect outside approval and lose final failure receipts

**Severity:** BLOCKER

**Raw:** `RT-B03a`, `RT-B03b`

**Evidence / affected:** `internal/connectors/engine/write.go:605-631` enables retries for idempotency headers; `internal/connectors/connsdk/http.go:1138-1143` then ceases treating the mutation as strict; `:1253-1273` can return nil after final 4xx/5xx outside destructive context.

**Required change:** Separate retry policy from redirect and receipt policy; all writes refuse redirects and retain final responses; only same-original-URL attempts may retry with one preview-bound stable provider-scoped idempotency key.

**Required tests:** `TestDeclarativeWrite_IdempotencyRetriesOriginalURLOnly`; `TestDeclarativeWrite_AllRedirectsRefusedAndFinalFailureRetained`; cover all 3xx codes, same/cross origin, multipart/bodyless, auth refresh, cancellation, and changing receipt headers.

**Six-surface impact:** ETL=destination writes blocked; reverse ETL=blocked; direct read=none; direct write=separate operation path passes; binary download=none; binary upload=retry policy dependency.

### CF-B14 — Buffered declarative HTTP follows cross-origin redirects with custom credentials

**Severity:** BLOCKER

**Raw:** `RT-B04`

**Evidence / affected:** runtime leaves `Requester.RedirectPolicy` nil (`internal/connectors/engine/read.go:565-571`); the buffered client bypasses credential stripping when nil (`internal/connectors/connsdk/stream.go:214-250`), while declarative reads call it directly (`read.go:355-368`).

**Required change:** Apply fail-closed redirect defaults to buffered requests: bounded hops, no downgrade, same-origin unless explicitly allowed, and stripping all credential/default-derived headers before an allowed cross-origin hop.

**Required tests:** `TestBufferedRequester_DefaultRedirectPolicyProtectsCustomCredentials`; include standard/custom auth, cookies, subdomains, ports, case, loops, relative locations, downgrade, and explicitly allowed credential-free downloads.

**Six-surface impact:** ETL=source and destination HTTP affected; reverse ETL=affected; direct read=declarative path blocked; direct write=operation path already explicit; binary download=allowed-host policy dependency; binary upload=must refuse redirects.

### CF-B15 — Reverse-write previews print resolved credentials and sensitive identifiers

**Severity:** BLOCKER

**Raw:** `RT-B05`

**Evidence / affected:** `internal/connectors/engine/write_prepare.go:62-69` appends resolved URL to warnings; tests at `write_test.go:895-928`, `:1000-1057` require base-URL secrets and `RedactFields` path values in preview; App returns these previews.

**Required change:** Keep exact URL/query/header/body only in private prepared material/digest; render warnings from declaration identity plus separately redacted URL and declaration-classified values, preserving ordinary IDs.

**Required tests:** `TestWritePreview_RedactsResolvedSecretsButDigestBindsThem`; include query/userinfo, nested/repeated/percent-encoded values, empty options, multi-record preview, and ordinary unclassified `token`.

**Six-surface impact:** ETL=destination preview blocked; reverse ETL=blocked; direct read=none; direct write=preview policy should share fix; binary download=none; binary upload=file/path preview must inherit.

### CF-B16 — REST and binary error formatting bypasses safe HTTP diagnostics

**Severity:** BLOCKER

**Raw:** `RT-B06`

**Evidence / affected:** `internal/connectors/engine/direct_read.go:44-58` unwraps `HTTPError` and rebuilds printable text from raw URL/query/body; used at `:145-155`, `:329-339`, and `internal/connectors/engine/binary_read.go:160-173`, bypassing `HTTPError.Error` redaction in `connsdk/http.go:55-74`.

**Required change:** Use one safe formatter containing status plus declaration-owned method/path identity and safe hint; never render typed body/raw query; retain typed cause privately for classification/parking and apply concrete secret masking at output.

**Required tests:** `TestRESTAndBinaryErrors_DoNotPrintProviderBodyOrQuerySecrets`; verify `errors.As` internally and cover 401, 429 reset evidence, truncation, invalid UTF-8, URLs, and terminal controls.

**Six-surface impact:** ETL=source diagnostics affected; reverse ETL=none; direct read=blocked disclosure; direct write=diagnostic policy dependency; binary download=blocked disclosure; binary upload=must inherit.

### CF-B17 — Declarative writes rematerialize records after approval

**Severity:** BLOCKER

**Raw:** `RT-B07`

**Evidence / affected:** preparation invokes `applyWriteRecordHook` and digest-binds material (`internal/connectors/engine/write_prepare.go:57-76`); approved execution passes original records and invokes mapper again (`internal/connectors/engine/write.go:302-381`).

**Required change:** Materialize once into an immutable execution plan consumed by the gated closure; if any rematerialization remains, canonicalize and compare every request component immediately before send and refuse divergence without I/O.

**Required tests:** `TestApprovedWrite_MapperRunsOnceAndExecutionMatchesDigest`; `TestApprovedWrite_StatefulMapperCannotChangeRequest`; cover every body format, records/order, handled=false, errors, caller mutation, file replacement, and cancellation.

**Six-surface impact:** ETL=destination mutation blocked; reverse ETL=blocked; direct read=none; direct write=prepared direct path separate; binary download=none; binary upload=file digest binding dependency.

### CF-B18 — Write receipt parsers fabricate body presence for empty JSON

**Severity:** BLOCKER

**Raw:** `RT-B08`

**Evidence / affected:** reverse parsing sets presence from JSON declaration (`internal/connectors/engine/write.go:520-551`); direct-write parsing does the same (`direct_write.go:2244-2269`); tests encode the incorrect empty-JSON behavior.

**Required change:** Derive presence from captured transport bytes (or explicit transport presence), share one receipt materializer, then parse by policy; zero-byte invalid JSON may error but must remain `body_present=false`.

**Required tests:** `TestWriteReceipts_BodyPresenceMatchesTransportBytes`; cover non-empty/null/text/binary, zero-byte 200/204, whitespace, invalid UTF-8, cap+1, bad content type, and repeated headers in both paths.

**Six-surface impact:** ETL=destination receipts affected; reverse ETL=blocked exact receipt; direct read=shared receipt semantics; direct write=blocked; binary download=shared semantics; binary upload=must inherit.

### CF-B19 — Numeric enum validation collapses integers above 2^53

**Severity:** BLOCKER

**Raw:** `RT-B09`

**Evidence / affected:** `internal/connectors/engine/schema.go:627-665` normalizes `json.Number`, `int`, and `int64` through `float64`; adjacent large integers compare equal.

**Required change:** Compare exact canonical JSON numbers using integers/rationals; document equivalence for `1`, `1.0`, and `1e0`; reject non-finite programmatic floats.

**Required tests:** `TestSchemaEnum_ExactNumbersBeyondFloatPrecision`; include adjacent positive/negative integers, decimals, exponents, mixed Go numeric types, negative zero, overflow, NaN, and infinity.

**Six-surface impact:** ETL=typed mappings affected; reverse ETL=blocked exact authorization; direct read=GraphQL/body variables affected; direct write=blocked; binary download=none; binary upload=typed metadata affected.

### CF-B20 — Singleton typed CLI flags silently use the last occurrence

**Severity:** BLOCKER

**Raw:** `RT-B10`

**Evidence / affected:** `internal/connectors/commandrunner/runner.go:2023-2037` validates all values then takes the last; literal text at `:2005-2020` does the same; occurrence checks at `:1497-1515` cover only limited direct-write query/header cases.

**Required change:** Enforce exactly one occurrence for every non-repeatable canonical target before coercion; reject aliases converging on one singleton; retain ordered repeats only for declared arrays/repeatable headers.

**Required tests:** `TestCommandRunner_RejectsDuplicateSingletonTargetsBeforeIO`; include path/query/body/form/GraphQL/raw body, aliases, identical values, booleans, empties, header casing, arrays, and structured JSON.

**Six-surface impact:** ETL=generated source flags possible; reverse ETL=affected; direct read=blocked exact mapping; direct write=blocked; binary download=path/query affected; binary upload=must inherit.

### CF-B21 — Caller-controlled path and query values have no byte bound

**Severity:** BLOCKER

**Raw:** `RT-B11`

**Evidence / affected:** `OperationParameter.MaxBytes` is header-only (`internal/connectors/engine/bundle.go:785-800`); command projection lacks a byte limit (`connector.go:959-974`); coercion and direct read/write path/query validation do not cap encoded values (`runner.go:2023-2037`, `direct_write.go:1993-2024`, `direct_read.go:909-978`).

**Required change:** Require source-owned or conservative engine path/query byte limits, project them into command metadata, enforce after exact wire encoding before auth/runtime construction, and validate fixed declarations at load.

**Required tests:** `TestOperationParameters_EnforceEncodedPathAndQueryByteCaps`; cover cap/cap+1, multibyte UTF-8, percent expansion, separators, long numerics, fixed query, and config fallback.

**Six-surface impact:** ETL=source URL inputs affected; reverse ETL=affected; direct read=blocked bounds; direct write=blocked bounds; binary download=blocked bounds; binary upload=blocked bounds.

### CF-B22 — REST direct reads remain an open query/body escape hatch

**Severity:** BLOCKER

**Raw:** `RT-B12`

**Evidence / affected:** `internal/connectors/engine/direct_read.go:61-108` merges arbitrary caller query and extra path values; `:431-474` merges arbitrary JSON body fields under potentially open schemas; read preflight (`internal/connectors/connectors.go:599-605`, `commandrunner/runner.go:748-775`) lacks exact bindings unlike direct write.

**Required change:** Add exact declared path/query/body bindings to read preflight; reject unknowns, fixed collisions, aliases, cross-bound mappings, and unused paths; require recursively closed bounded JSON schemas; preserve raw body only for exact bounded `text/plain` root strings.

**Required tests:** `TestOperationDirectRead_ClosedBindingsRejectUnknownsBeforeIO`; cover required groups, config paths, pagination controls, fixed collisions, text/plain presence, and GraphQL override rejection.

**Six-surface impact:** ETL=none; reverse ETL=none; direct read=blocked authority; direct write=closed path passes; binary download=separate typed path; binary upload=none.

### CF-B23 — Name-based JSON redaction deletes ordinary provider identifiers

**Severity:** BLOCKER

**Raw:** `RT-B13`

**Evidence / affected:** `internal/connectors/engine/direct_read.go:671-697`, `:744-755`, `:805-822` deletes fields by substrings such as `token`; real ordinary fields include Braintree primary key `token` and Nebius counter `trained_tokens`.

**Required change:** Preserve provider fields by default; redact only declaration-classified locations or concrete secret values at the public boundary; retain explicit presence markers and migrate broad policies to field lists.

**Required tests:** `TestJSONOutputRedaction_PreservesOrdinaryCredentialNamedFields`; `TestJSONOutputRedaction_MasksDeclaredOrConcreteSecrets`; include nulls, nested arrays/maps, case/punctuation, secret under ordinary key, and ordinary value under secret-looking key.

**Six-surface impact:** ETL=source output may narrow; reverse ETL=none; direct read=blocked truth preservation; direct write=public sanitizer policy related; binary download=metadata policy related; binary upload=receipt policy related.

### CF-B24 — Provider side effects occur before stale-writer ownership CAS

**Severity:** BLOCKER

**Raw:** `ORCH-B01`

**Evidence / affected:** ordinary apply/readback precedes commit (`internal/synctransport/orchestrator.go:233-285`); full overwrite and Arrow publish/readback also precede commit (`orchestrator.go:472-495`, `arrow_fast_path_controller.go:184-216`, `arrow_fast_path_pipeline.go:125-157`); App's only stream-state CAS is later (`internal/app/transport_dispatch.go:252-278`). Tests confirm losing writers already mutated.

**Required change:** Acquire durable stream-scoped execution lease/fencing generation atomically before I/O; carry fence and stable receipt/idempotency identity through apply/publish; renew/reconcile; checkpoint CAS must validate/retire same fence; refuse routes without fencing/keyed idempotency where replay is possible.

**Required tests:** `TestTransport_TwoAppsFenceBeforeProviderMutation`; `TestTransport_CrashAndLeaseTakeoverReconcileWithoutReplay`; cover late overwrite publisher, renewal cancellation, restart, and provider/checkpoint generation consistency.

**Six-surface impact:** ETL=blocked; reverse ETL=blocked for destination transports; direct read=none; direct write=none unless orchestrated as destination; binary download=none; binary upload=affected as destination.

### CF-B25 — Post-publication readback failure invokes pre-publication abort

**Severity:** BLOCKER

**Raw:** `ORCH-B02`

**Evidence / affected:** `AbortFullOverwrite` is pre-publication cleanup (`internal/synctransport/types.go:184-195`), but published flags are set only after readback in standard and both Arrow controllers (`orchestrator.go:352-361`, `:474-485`; Arrow files cited above).

**Required change:** Track attempted, publish-succeeded, and readback-succeeded states separately; disarm pre-publication abort immediately after successful publish; persist ambiguous publication receipt and reconcile/read back without republishing.

**Required tests:** `TestFullOverwrite_PublishSuccessReadbackFailureDoesNotAbort`; run standard/Arrow serial/pipeline; cover pre-publish error abort-once, timeout between publish/readback, and restart reconciliation.

**Six-surface impact:** ETL=blocked full overwrite; reverse ETL=blocked destination publication; direct read=none; direct write=none; binary download=none; binary upload=none.

### CF-B26 — Full-overwrite and Arrow publication outputs never reach results

**Severity:** BLOCKER

**Raw:** `ORCH-B03`

**Evidence / affected:** ordinary path appends output (`internal/synctransport/orchestrator.go:237-248`), but full overwrite and Arrow publication acknowledgements are dropped (`orchestrator.go:436-504`, `arrow_fast_path_controller.go:186-223`, `arrow_fast_path_pipeline.go:127-164`); App can persist only collected results.

**Required change:** Use one defensive-copy result collector for ordinary/full-overwrite/Arrow paths; extend session APIs or typed errors to carry provider output; append acknowledgement before readback/checkpoint while applying `CF-B08` public masking.

**Required tests:** `TestFullOverwriteAndArrow_PersistPublicationOutput`; `TestPublicationFailure_RetainsMaskedProviderReceipt`; cover later readback/checkpoint failure, duplicate headers, empty body, unknown fields, large IDs, base64, and credential echoes.

**Six-surface impact:** ETL=blocked full overwrite; reverse ETL=blocked output; direct read=none; direct write=none; binary download=none; binary upload=publication output contract dependency.

### CF-B27 — Declared delivery guarantees are never enforced

**Severity:** BLOCKER

**Raw:** `ORCH-B04`

**Evidence / affected:** `DeliveryGuarantees` defines idempotency/ordering/deletes (`internal/connectors/sync_transport.go:41-70`), but validation checks enum spelling only (`:296-312`); `Registry.Preflight` (`internal/synctransport/registry.go:131-218`) does not compare endpoints/mode; tombstones are passed without guarantee checks.

**Required change:** Add mode/strategy compatibility policy to preflight; reject replayable non-idempotent, ordering-incompatible, and tombstone-incompatible pairs; assert runtime page behavior before stage/provider I/O.

**Required tests:** `TestTransportPreflight_EnforcesDeliveryGuaranteeCompatibility`; `TestTransportRuntime_RejectsUndeclaredTombstoneBeforeIO`; cover all canonical modes/strategies, restart, and deferred checkpoints.

**Six-surface impact:** ETL=blocked; reverse ETL=blocked for sync destinations; direct read=none; direct write=none; binary download=none; binary upload=destination mode dependency.


## Warnings

### CF-W01 — Generated connector skills are hard-coded to five connector names

**Severity:** WARNING

**Raw:** `MAP-WR-01`

**Evidence / affected:** `internal/cli/skills.go:137-145` gates generic `connectorSkill` generation by five names; goldens do not cover a non-allowlisted connector.

**Required change:** Generate from manifest capability or an explicit validated publication property; skip only with typed reason, never name literals.

**Required tests:** `TestSkillGeneration_NonAllowlistedConnectorWithGuideIsPublished`; cover missing guide, opt-out, local-only, and sanitization collisions.

**Six-surface impact:** ETL=discoverability; reverse ETL=discoverability; direct read/write=discoverability; binary download/upload=discoverability.

### CF-W02 — Path interpolation validates before filters, not after final path context

**Severity:** WARNING

**Raw:** `RT-W14`

**Evidence / affected:** `internal/connectors/engine/interpolate.go:154-226` checks the original value only; `last_path_segment`, `join`, and `const` can emit raw output (`:492-518`, `:562-581`); final path check only rejects decoded `..`.

**Required change:** Make filter admission context-aware; encode final dynamic path output as one segment or reject multi-segment filters; revalidate separators, traversal, controls, bidi/invisible, query, and fragment syntax; validate filter arguments at load.

**Required tests:** `TestInterpolatePath_RevalidatesFilteredOutput`; cover `?`, `#`, `%2f..%2f`, controls/bidi, percent signs, empty/trailing segments, join separators, and chains.

**Six-surface impact:** ETL/reverse/direct read/direct write/binary download/upload=latent wherever declaration path interpolation is used.

### CF-W03 — Accepted GitHub secret transform has no executor

**Severity:** WARNING

**Raw:** `RT-W15`

**Evidence / affected:** `internal/connectors/engine/bundle.go:874-897`, `:2949-2983` models and accepts `github_secret_encryption`; no runtime reads `SensitivePolicySpec.Transform`; tests validate spelling only.

**Required change:** Use connector-neutral registered transforms bound into prepared digest and fail closed for executable declarations without implementation; reject this transform until implemented.

**Required tests:** `TestSensitiveTransform_RegisteredExecutionBindsDigest`; `TestSensitiveTransform_UnimplementedFailsBeforePreview`; cover errors, key/input changes, zeroization, retry determinism, and no connector-name dispatch.

**Six-surface impact:** reverse ETL/direct write=bundle-dependent latent risk; other surfaces=none.

### CF-W04 — Transient rate-parking failures remove the only retry timer

**Severity:** WARNING

**Raw:** `ORCH-W05`

**Evidence / affected:** `internal/coordination/rate_parking.go:600-694` deletes timers and returns on claim/release/resume/complete failures except one rearm branch.

**Required change:** Classify pre/post-provider phases; bounded-retry safe pre-provider failures; persist post-provider reconciliation state without replay; surface release errors and secret-safe events; schedule against backoff/reset/claim deadlines.

**Required tests:** `TestRateParking_TransientFailuresRemainScheduledOrReconcile`; cover claim retry, complete-after-success without second provider call, release error, renewal loss, close, restart, and retry ceiling.

**Six-surface impact:** ETL/reverse ETL=rate-limited orchestration robustness; direct/binary surfaces=none unless routed through parking.

### CF-W05 — Same provider-scope parked runs resume concurrently

**Severity:** WARNING

**Raw:** `ORCH-W06`

**Evidence / affected:** parking records/timers/claims are keyed by run ID (`internal/coordination/rate_parking.go:110-126`, `:386-474`; `durable_store.go:194-287`) while admission is scope-aware.

**Required change:** Persist a per-scope queue/leader and atomic scope claim; only exact leader bypasses admission; retain concurrency across independent scopes.

**Required tests:** `TestRateParking_SerializesSameScopeAcrossCoordinators`; cover different scopes, leader cancellation/expiry/takeover, restart, reset ordering, and rearm.

**Six-surface impact:** ETL/reverse ETL=provider rate-limit coordination; direct/binary surfaces=none unless parked.

### CF-W06 — Route selection hides actionable transport preflight errors

**Severity:** WARNING

**Raw:** `ORCH-W07`

**Evidence / affected:** `internal/app/transport_dispatch.go:40-91` converts preflight/connection/stream/action errors to false; `etl_mode_dispatch.go:51-106` then emits generic error or falls into legacy I/O.

**Required change:** Return typed `(selected, reason, error)`; allow legacy only for explicit declaration absence; once selected, propagate all preflight errors before catalog/source/provider calls.

**Required tests:** `TestETLRouteSelection_PropagatesDeclaredRoutePreflightErrors`; cover valid legacy absence, missing executor/mapping/action/evidence, invalid metadata, contract modes, and zero-I/O failure.

**Six-surface impact:** ETL=diagnostic/fail-closed; reverse ETL=destination route; other surfaces=none.

### CF-W07 — Generated website data is deterministic but semantically stale

**Severity:** WARNING

**Raw:** `ORCH-W08`

**Evidence / affected:** `website/content/docs/cli-reference.mdx:83-93` omits `postgres-managed-target`; `website/content/docs/etl.mdx:376-388` omits exit code 7 and `:45-58` uses wrong `<credential>:<credential-name>` grammar; generator test compares only generated bytes to the same prose.

**Required change:** Correct MDX, regenerate `website/lib/docs.generated.ts`, and derive command map/exit codes/endpoint grammar from structured CLI metadata with semantic parity tests.

**Required tests:** `TestGeneratedWebsiteData_RuntimeCLIParity`; cover hidden/aliases, code 7, endpoint grammar, URL uniqueness, separators, and two-run determinism.

**Six-surface impact:** ETL/reverse/direct read/direct write/binary download/upload=documentation/discovery accuracy.

### CF-W08 — Evidence status is inconsistent and overstates coverage

**Severity:** WARNING

**Raw:** `ORCH-W09a`, `ORCH-W09b`

**Evidence / affected:** `data/.../evidence-manifest.json` claims complete provider retention and full App success while phase `TDD-LEDGER.md`, `PLAN.md`, and `VERIFICATION.md` remain pending; #4303 run-state overstates output despite `CF-B26`; manifest records composite code SHA but no reviewed-evidence SHA.

**Required change:** Generate gate status from one command ledger; update planning/evidence atomically; store separate `code_sha` and `reviewed_sha`; bind each claim to named tests/modes; retract broad output coverage until every path is tested.

**Required tests:** `TestFoundationEvidence_AllPassedClaimsHaveMatchingGateAndSHA`; cover pending-vs-passed, reviewed SHA mismatch, cached/deferred/provisional/mode-limited evidence.

**Six-surface impact:** all six=release evidence cannot currently prove claimed coverage.

### CF-W09 — Rate-parking store APIs can persist unloadable state

**Severity:** WARNING

**Raw:** `ORCH-W10`

**Evidence / affected:** memory `Create` skips domain validation; memory/file `Claim` and `RenewClaim` accept blank owner/non-forward deadlines (`internal/coordination/rate_parking.go:110-202`, `durable_store.go:194-305`); persisted validation later rejects incomplete claims; JSON updates lack post-mutation domain validation.

**Required change:** Validate IDs, owner, timestamps, and forward deadlines at every boundary; align memory/file behavior; validate the complete next state before persistence so rejected operations are byte-stable.

**Required tests:** `TestRateParkingStores_RejectInvalidMutationWithoutStateChange`; run memory/file table cases and reopen after every accepted/rejected mutation.

**Six-surface impact:** ETL/reverse ETL=rate-parking durability; direct/binary surfaces=none unless parked.


## One fix wave: dependency and order

All fixes belong to one merge-blocking wave, executed in this order:

1. **Freeze red tests and invariants:** add source-lock, secret-boundary, terminal-receipt, input-closure, approval-materialization, stale-writer, and all-mode orchestration failures without changing production behavior. This prevents later generator churn from hiding current defects.
2. **Repair source authority and generation (`CF-B01`–`CF-B05`, `CF-W01`, `CF-W07`, `CF-W08`):** strict lock -> canonical descriptor -> generated bundle/CLI/docs/skills/gap ledger. This must precede runtime fixes because request/response schemas and installed commands are their authority.
3. **Establish one receipt and masking model (`CF-B08`–`CF-B12`, `CF-B16`, `CF-B18`, `CF-B23`, `CF-B26`):** complete internal receipt, credential-safe public clone, typed terminal errors, CLI emission. Downstream orchestration result collection depends on this type contract.
4. **Close request construction (`CF-B13`–`CF-B15`, `CF-B17`, `CF-B19`–`CF-B22`, `CF-W02`, `CF-W03`):** redirect/retry separation, safe previews, one-time materialization, exact numeric/occurrence/bound enforcement, closed direct reads, fail-closed transforms.
5. **Repair destination semantics and ownership (`CF-B06`, `CF-B07`, `CF-B24`–`CF-B27`, `CF-W04`–`CF-W06`, `CF-W09`):** reusable authorization state, real provider readback, stream fences, publication state machine, delivery compatibility, parking/store reconciliation, typed route decisions.
6. **Regenerate and certify all six surfaces:** installed binary commands and help, manuals, website, skills, source/projection drift checks, focused/race tests, App/CLI failure transcripts, and mode-complete evidence at one exact reviewed SHA.

No group may be declared complete while a later generated artifact still derives from pre-fix source metadata. No test should preserve raw configured credentials merely to assert provider-output completeness.

## What passed

The following narrow contracts passed and must remain regression-protected; none negates the blockers:

- Frozen SHA and source cleanliness at review intake; component ancestry/provenance checks were consistent with the input manifest.
- All 1,220 locked REST method/path identities exist in `api_surface.json` (field completeness does not pass).
- Operation-backed direct writes have declaration-owned request mappings, no raw HTTP/action escape hatch, sealed preview/digest, single-use approval immediately before I/O, and fail-closed redirect/retry behavior on the direct-operation path.
- Structured REST request bodies are recursively closed/bounded and multipart files are root-confined, capped, and digest-bound.
- Successful binary/text downloads are GET-only, bounded, root-confined, atomic, and return exact size/SHA-256; failure output remains blocked by `CF-B11`/`CF-B16`.
- Fixed GraphQL documents reject caller path/query/header overrides and closed variables are bounded; generated selection completeness remains blocked by `CF-B04`.
- Atomic transport definition registration, local warehouse WAL/Parquet durability, per-unit approval order, durable local acknowledgement validation, cancellation handling, bounded Arrow credits, and connector-neutral dispatch passed their focused tests.
- Rate-limit parking requires a typed reset-bearing cause before park/rearm; parking scheduling/store robustness remains warned.
- Website generation is deterministic; semantic parity remains warned.
- Focused package tests and vet commands recorded by the ledgers passed. The mapping ledger separately demonstrated the real checked-in source-import failure and 211-operation parameter drift. Passing existing tests is not correctness evidence where tests encode the unsafe behavior.

## Raw-to-canonical crosswalk (41/41 atomic claims)

| # | Atomic raw claim | Canonical ID | Disposition |
| ---: | --- | --- | --- |
| 1 | `MAP-BL-01a` — real checked-in lock exceeds reused grammar-position byte bound | `CF-B01` | BLOCKER |
| 2 | `MAP-BL-01b` — GraphQL lock section is silently ignored | `CF-B01` | BLOCKER |
| 3 | `MAP-BL-02` — descriptor output is orphaned from generation/validation | `CF-B02` | BLOCKER |
| 4 | `MAP-BL-03` — reverse actions omit provider inputs | `CF-B03` | BLOCKER |
| 5 | `MAP-BL-04` — GraphQL result/pagination surface is narrowed | `CF-B04` | BLOCKER |
| 6 | `MAP-BL-05` — binary upload is wrong JSON/origin/request | `CF-B05` | BLOCKER |
| 7 | `MAP-BL-06` — reusable authorization fails after effects | `CF-B06` | BLOCKER |
| 8 | `MAP-BL-07` — destination read-back is local-only | `CF-B07` | BLOCKER |
| 9 | `MAP-BL-08` — configured credentials survive public sanitizers | `CF-B08` | BLOCKER; duplicate of `RT-B01` |
| 10 | `MAP-BL-09` — CLI discards failed runs/receipts | `CF-B09` | BLOCKER |
| 11 | `MAP-BL-10` — no-response write loses attempted identity | `CF-B10` | BLOCKER |
| 12 | `MAP-BL-11a` — direct-read result/output is incomplete | `CF-B11` | BLOCKER |
| 13 | `MAP-BL-11b` — binary-download error receipt/output is incomplete | `CF-B11` | BLOCKER |
| 14 | `MAP-WR-01` — generated skill allowlist | `CF-W01` | WARNING |
| 15 | `RT-B01` — no-op public write sanitizers disclose credentials | `CF-B08` | BLOCKER; duplicate of `MAP-BL-08` |
| 16 | `RT-B02` — accepted-status mismatch drops receipt | `CF-B12` | BLOCKER |
| 17 | `RT-B03a` — idempotent write can redirect outside approval | `CF-B13` | BLOCKER |
| 18 | `RT-B03b` — idempotent write can lose terminal failure response | `CF-B13` | BLOCKER |
| 19 | `RT-B04` — buffered cross-origin redirects carry custom credentials | `CF-B14` | BLOCKER |
| 20 | `RT-B05` — preview leaks URL/path secrets | `CF-B15` | BLOCKER |
| 21 | `RT-B06` — error formatter exposes raw query/provider body | `CF-B16` | BLOCKER |
| 22 | `RT-B07` — write hook rematerializes after approval | `CF-B17` | BLOCKER |
| 23 | `RT-B08` — empty JSON fabricates body presence | `CF-B18` | BLOCKER |
| 24 | `RT-B09` — enum comparison loses >2^53 precision | `CF-B19` | BLOCKER |
| 25 | `RT-B10` — duplicate singleton flags are last-wins | `CF-B20` | BLOCKER |
| 26 | `RT-B11` — path/query values lack byte bounds | `CF-B21` | BLOCKER |
| 27 | `RT-B12` — direct reads admit undeclared query/body fields | `CF-B22` | BLOCKER |
| 28 | `RT-B13` — name-based redaction deletes ordinary IDs | `CF-B23` | BLOCKER |
| 29 | `RT-W14` — path filters are not post-validated | `CF-W02` | WARNING |
| 30 | `RT-W15` — declared secret transform has no executor | `CF-W03` | WARNING |
| 31 | `ORCH-B01` — side effects precede stale-writer CAS | `CF-B24` | BLOCKER |
| 32 | `ORCH-B02` — abort runs after publication | `CF-B25` | BLOCKER |
| 33 | `ORCH-B03` — full-overwrite/Arrow output is dropped | `CF-B26` | BLOCKER |
| 34 | `ORCH-B04` — delivery guarantees are unenforced | `CF-B27` | BLOCKER |
| 35 | `ORCH-W05` — parking failures lose retry timer | `CF-W04` | WARNING |
| 36 | `ORCH-W06` — same-scope parked runs resume concurrently | `CF-W05` | WARNING |
| 37 | `ORCH-W07` — route selection hides preflight failures | `CF-W06` | WARNING |
| 38 | `ORCH-W08` — generated website is semantically stale | `CF-W07` | WARNING |
| 39 | `ORCH-W09a` — planning/evidence gate states disagree | `CF-W08` | WARNING |
| 40 | `ORCH-W09b` — evidence overstates output-mode coverage and SHA | `CF-W08` | WARNING |
| 41 | `ORCH-W10` — store APIs can persist unloadable state | `CF-W09` | WARNING |

---

_Reviewed: 2026-08-20T18:02:34Z_
_Reviewer: the agent (gsd-code-reviewer)_
_Depth: deep_
