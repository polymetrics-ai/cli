---
phase: cli-current-foundations-main-integration-r1
reviewed: 2026-08-21T03:15:01Z
depth: deep
source_sha: 8a8a866ff6d5282c28bda12acceed8a624218f01
diff_base: e62ae21d428f0d27225f9bff564dc2cd797f6b65
codegraph: unavailable
files_reviewed: 120
files_reviewed_list:
  - internal/app/app.go
  - internal/app/authorization.go
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
  - internal/app/transport_composition_test.go
  - internal/app/transport_dispatch.go
  - internal/app/transport_dispatch_test.go
  - internal/app/types.go
  - internal/app/util.go
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
  - internal/cli/skills.go
  - internal/cli/structured_rest_body_help_test.go
  - internal/connectors/command_surface.go
  - internal/connectors/commandrunner/content_preservation_test.go
  - internal/connectors/commandrunner/github_declared_parity_test.go
  - internal/connectors/commandrunner/runner.go
  - internal/connectors/commandrunner/runner_test.go
  - internal/connectors/conformance/dynamic.go
  - internal/connectors/conformance/dynamic_test.go
  - internal/connectors/conformance/github_exhaustive_proof_internal_test.go
  - internal/connectors/conformance/github_exhaustive_proof_test.go
  - internal/connectors/conformance/static.go
  - internal/connectors/conformance/static_test.go
  - internal/connectors/connectors.go
  - internal/connectors/connsdk/http.go
  - internal/connectors/connsdk/http_test.go
  - internal/connectors/connsdk/multipart_bounds_test.go
  - internal/connectors/connsdk/stream.go
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
  - internal/connectors/guide.go
  - internal/connectors/guide_test.go
  - internal/connectors/hooks/google-calendar/hooks_test.go
  - internal/connectors/native/ashby/connector_contract_test.go
  - internal/connectors/native/ashby/engine_delegate.go
  - internal/connectors/native/postgres/transport_source_test.go
  - internal/connectors/operation_headers.go
  - internal/connectors/sync_transport.go
  - internal/connectors/sync_transport_test.go
  - internal/connectors/transportpolicy/policy.go
  - internal/connectors/write_result_output_test.go
  - internal/coordination/durable_store.go
  - internal/coordination/durable_store_edges_test.go
  - internal/coordination/rate_parking.go
  - internal/coordination/rate_parking_test.go
  - internal/synccontract/commit.go
  - internal/synctransport/arrow_fast_path_controller.go
  - internal/synctransport/arrow_fast_path_pipeline.go
  - internal/synctransport/orchestrator.go
  - internal/synctransport/registry.go
  - internal/synctransport/transport_test.go
  - internal/synctransport/types.go
findings:
  critical: 13
  warning: 3
  info: 0
  total: 16
status: issues_found
verdict: blockers
---

# Foundation Post-Fix Runtime/Security Review

**Reviewed:** 2026-08-21T03:15:01Z
**Depth:** deep
**Frozen source:** `8a8a866ff6d5282c28bda12acceed8a624218f01`
**Diff base:** `e62ae21d428f0d27225f9bff564dc2cd797f6b65`
**Status:** issues_found — merge blocked

## Summary

The post-fix runtime is not releasable. This independent review froze **16 findings: 13 BLOCKER and 3 WARNING** before reading the prior review conclusions. The dominant failures are not missing surface polish: they are approval-to-wire drift, credential under- and over-redaction, provider-receipt loss, credential-bearing redirect behavior, stale authorization at destination effects, exact-number corruption, and a binary no-overwrite race that can overwrite or delete another writer's file.

The output contract remains violated in both directions. Concrete configured or declaration-classified credential material can survive in raw receipts and GraphQL metadata, while short configured secrets and keyword heuristics rewrite ordinary provider keys, occurrence IDs, header names, and error messages. Provider truth and masking therefore cannot both be trusted at the CLI boundary.

No production, test, generated, branch, commit, remote, or PR state was modified. This report is the only file written by this review.

## Scope Manifest

### Primary changed-file scope

The 120 Go source/test files in frontmatter are the filtered runtime/security scope changed between the diff base and frozen HEAD. Every primary file was read at deep-review depth or traced as part of its package/call path. Generated definitions were not counted as source in the frontmatter, but declaration rows were inspected where they prove a runtime operation is installed or a response field is classified.

### Supporting cross-module source inspected

The following supporting files were traced even when they were outside the filtered changed-file count: `internal/connectors/engine/hooks.go`, `internal/connectors/engine/write_gate.go`, `internal/connectors/hooks/github/hooks.go`, `internal/connectors/native/amazon-sqs/{connector.go,connection.go,direct_read.go}`, `internal/connectors/native/nativeset/factories.go`, `internal/connectors/native/postgres/{transport_destination.go,arrow_full_overwrite_transport.go}`, and relevant generated declarations under `internal/connectors/defs/{github,ashby,bahmni,google-search-console,lever-hiring,zendesk-support}`.

### Traced execution surfaces

| Surface | Traced path |
| --- | --- |
| REST direct read/status | CLI parsing -> `commandrunner.runDirectRead` / `runStatusCheck` -> connector interface -> engine operation/pagination/status executor -> `connsdk.Requester` -> response receipt -> CLI terminal envelope |
| GraphQL read/mutation | CLI typed flags -> commandrunner -> fixed-document variable construction -> engine GraphQL executor -> connsdk -> GraphQL metadata/receipt -> CLI output |
| Binary download/upload | commandrunner -> engine stream/multipart preparation -> redirect policy/file confinement/digest checks -> result receipt/artifact publication -> CLI output |
| Direct write | App/CLI plan -> prepared request/digest -> approval gate -> operation or hook executor -> connsdk -> provider receipt -> App/CLI sanitizer |
| ETL | App dispatch -> synctransport source page -> warehouse stage/reopen -> destination apply/readback -> checkpoint commit |
| Reverse ETL | App durable authorization -> destination plan -> per-unit authorization/evidence -> connector write/hook/native destination -> provider output persistence |

### CodeGraph

`.codegraph/` was absent at the repository root. Per project policy, CodeGraph was not invoked; review used read-only source tracing, `rg`, `git diff`, and Go tooling.

## Checks Run and Results

| Check | Result |
| --- | --- |
| `git rev-parse HEAD` before review and before writing | PASS: exact frozen SHA `8a8a866ff6d5282c28bda12acceed8a624218f01`. |
| Source cleanliness preflight | PASS: no tracked or untracked source/test/generated drift. Only peer review Markdown appeared later and was treated as permitted non-source review output. |
| `git diff --check e62ae21d...8a8a866f` | PASS. |
| `go test -timeout 20m ./internal/connectors/connsdk ./internal/connectors/engine ./internal/connectors/commandrunner ./internal/connectors/conformance ./internal/connectors ./internal/cli ./internal/app ./internal/synctransport ./internal/coordination` | PASS. Notable uncached package times: CLI `1083.733s`, App `326.863s`, connsdk `6.648s`, connectors `4.041s`, synctransport `5.793s`, coordination `12.186s`. |
| `go vet` over the same packages | PASS. |
| `go test -race -timeout 20m ./internal/connectors/engine ./internal/connectors/commandrunner ./internal/synctransport` | PASS (`59.645s`, `177.044s`, `6.068s`). |
| `go test -race -timeout 20m ./internal/connectors/connsdk ./internal/coordination` | FAIL: connsdk race in `multipart_bounds_test.go`; coordination passed. The race is recorded as `PFR-WR-03`. |

Passing tests do not close the findings below: the missing regression cases sit at cross-package boundaries that the current suite does not assert.

## Narrative Findings (AI reviewer)

## Blockers

### PFR-BL-01 — Raw receipts and GraphQL metadata can disclose concrete and declared secrets

**Severity:** BLOCKER

**Evidence:** `internal/connectors/engine/response_receipt.go:30-47` (`providerResponseReceipt`) constructs and publicly sanitizes the complete raw/decoded receipt before operation-declared redaction is known. `internal/connectors/engine/direct_read.go:143-194` and `:490-532` redact only the convenience `Body`, leaving `Receipt.Body` and `Receipt.BodyRaw` untouched by `RedactFields`/`SensitivePolicy`. `internal/connectors/connectors.go:980-1007`, `:1094-1119`, and `:1192-1212` mask raw bodies with literal `strings.ReplaceAll` over raw/query/path/base64 variants; a JSON string such as `pa\"ss` appears on the wire as `pa\\\"ss` and is not matched. JSON HTML escapes and Unicode escapes have the same gap. Real declarations classify sensitive response locations, including `internal/connectors/defs/zendesk-support/operations.json:3461-3475` (`token`, `refresh_token`) and PII-bearing operations in Bahmni, Lever, and Google Search Console. `internal/connectors/engine/graphql_operation.go:665-710` builds the receipt before data redaction, while `boundedGraphQLErrorMetadata` at `:797-832` deliberately restores unsanitized provider text when `retainRuntimeErrors` is true and never receives configured secret values.

**Cross-file call path:** REST/GraphQL provider response -> `providerResponseReceipt*` -> `SanitizeProviderResponseReceiptForOutput` -> engine applies declared redaction only to `DirectReadResult.Body` -> `commandrunner.runOperationDirectRead` -> `internal/cli/cli.go:1415-1438` emits both convenience response and unredacted receipt. Direct-write raw bodies follow `OperationDirectWriteResult` -> App persistence (`internal/app/app.go:2927`, `:3150`) -> literal sanitizer.

**Behavior/security impact:** A provider echoing a configured secret using ordinary JSON escaping, or returning a declaration-classified secret different from configured credentials, can place it in durable/printable receipts. Partial GraphQL error messages can leak concrete credentials without any keyword such as “token.”

**Six-surface impact:** ETL=provider source/destination receipts unsafe; reverse ETL=write receipts unsafe; direct read=REST and GraphQL blocked; direct write=blocked; binary download=raw/error receipt sanitizer shared; binary upload=must inherit the same safe receipt boundary.

**Exact fix:** Build an immutable internal receipt, then create a syntax-aware public clone. Decode JSON with `UseNumber`, traverse values and declaration-classified paths, mask exact credential scalars, and re-encode canonical public JSON for `BodyRaw`; for invalid/binary bytes use byte-safe exact matching or an explicit withheld/masked representation. Pass concrete configured credentials and declared response-secret locations into GraphQL error/extension sanitization. Apply the public projection once, at App/CLI persistence/printing boundaries.

**Exact behavioral regression test:** Add a table-driven end-to-end test whose configured secret contains quote, backslash, `<`, `>`, `&`, non-ASCII, and base64-sensitive bytes; return each in REST decoded/raw JSON, GraphQL partial-data error text, declared Zendesk secret fields, and binary/error receipts. Assert no concrete or encoded variant appears in serialized output, every masked field remains present, ordinary provider fields and occurrence IDs remain byte-for-byte intact, and the internal receipt is not mutated.

### PFR-BL-02 — Public sanitizers corrupt ordinary provider fields, IDs, header names, and messages

**Severity:** BLOCKER

**Evidence:** `internal/connectors/connectors.go:1019-1031` rewrites response header names, and `sanitizeProviderOutputValue` at `:1058-1082` rewrites every map key, string, and matching `json.Number` by substring. `redactWriteResultString` at `:1208-1212` is unanchored. A valid configured secret `id` changes the key `occurrence_id` to `occurrence_[masked]` and can rewrite unrelated identifier values. `internal/connectors/engine/operation_headers.go:21-44` also masks header names by keyword rather than keeping the name and masking only classified values. `internal/connectors/engine/graphql_operation.go:817-832` replaces every provider message containing “token,” “secret,” or similar text, even when it contains no concrete credential. Existing preservation coverage (`internal/connectors/commandrunner/content_preservation_test.go:193-213`) does not exercise a short-secret collision.

**Cross-file call path:** Provider output -> engine receipt/result -> `SanitizeProviderResponseReceiptForOutput` / `SanitizeWriteResultForOutput` -> recursive key/name/value replacement -> App persisted output or `internal/cli/cli.go:1415-1438` terminal envelope.

**Behavior/security impact:** The public result ceases to be provider truth. Required IDs (including occurrence IDs), unfamiliar keys, repeatable header names, and ordinary diagnostic text can disappear or change, preventing audit, retry, reconciliation, and downstream automation. This violates the explicit contract to preserve all ordinary fields and mask only concrete configured/declared credential material.

**Six-surface impact:** ETL=record/receipt identity corruption; reverse ETL=provider write output corrupted; direct read=REST/GraphQL output corrupted; direct write=blocked; binary download=header metadata affected; binary upload=receipt/header metadata affected.

**Exact fix:** Never mutate JSON object keys or HTTP header names. Mask only exact scalar occurrences proven to equal configured/declared secret material, with context-aware encoded variants, and preserve field/header presence using a value-level masked marker. Remove keyword-only GraphQL/header redaction; declaration metadata may classify a value location but not authorize deletion or key renaming.

**Exact behavioral regression test:** Configure secrets `id`, `token`, `0`, and a one-character value. Return `occurrence_id`, `trained_tokens`, `token_type`, `WWW-Authenticate`, duplicate ordinary headers, large numeric IDs, and an “Unknown token type” GraphQL message. Assert every key/name/value remains exact unless the scalar is the concrete secret; assert the concrete secret is masked at every public boundary and input objects remain unchanged.

### PFR-BL-03 — `WriteHook` executes unpreviewed requests, mutable records, and receipt-less compound mutations

**Severity:** BLOCKER

**Evidence:** `internal/connectors/engine/write_prepare.go:35-55` rejects a handled hook only when `target.RequiresApproval()` **and** `action.Hook` is nonempty. `DestructiveTarget.RequiresApproval` (`engine/write_gate.go:26-31`) recognizes delete/destructive/confirmation metadata, so ordinary create/update hooks and hook implementations whose action omits the `hook` field bypass the guard. `engine/write.go:317-340` approves a declarative `PreparedWrite`, but `executeApprovedWrite` at `:365-427` calls `WriteHook.ExecuteWrite` first with the original caller record (`:382-390`), before the pinned record/request equality check (`:394-403`); handled hooks merely increment a counter and append no `ProviderResponses`. GitHub's hook dispatch (`internal/connectors/hooks/github/hooks.go:257-286`) handles eight actions. `closeResource` (`:369-395`) can send a comment then a PATCH, and PR create/update (`:432-498`) can send up to three provider requests while discarding every response. Ashby's hook (`internal/connectors/native/ashby/engine_delegate.go:75-93`) handles all actions and discards its response; actions such as `delete_application` omit the declaration hook field, so even delete bypasses the preparation rejection.

**Cross-file call path:** App/CLI plan -> `engine.prepareDeclarativeWrite` creates/digests declarative `PreparedRequest` -> approval -> `engine.executeApprovedWrite` -> GitHub/Ashby `WriteHook` sends a different request sequence from the original mutable record -> no `WriteProviderResponse` -> App/CLI persists counts without provider receipts.

**Behavior/security impact:** Approval does not bind the method/path/query/body or number/order of actual provider mutations. Caller mutation while approval is pending can change the wire request. Successful and partially failed compound operations lose provider IDs, occurrence IDs, status, headers, bodies, and the exact nth request that committed.

**Six-surface impact:** ETL=hook-backed destinations unsafe; reverse ETL=blocked; direct read=none; direct write=hook-backed write commands blocked; binary download=none; binary upload=compound/file hook designs cannot safely reuse this seam.

**Exact fix:** Replace `WriteHook`'s execute override with a prepare hook that returns an exact ordered list of sealed `PreparedRequest` values plus a response projector, and execute those exact bytes only after approval. Alternatively, fail closed for every action a `WriteHookClassifier` says it handles, independent of declaration metadata, until such a plan exists. The hook result must return one bounded response receipt per attempted provider request, including prior successes and the terminal nth failure; execution must consume the pinned record only.

**Exact behavioral regression test:** For every GitHub handled action and representative Ashby create/delete, block at approval, mutate the original record, then release. Assert the exact previewed request sequence and bytes are sent or execution fails before I/O; assert success returns every response and ID, a failure on request N retains receipts 1..N, counts match committed effects, and no raw credential reaches output.

### PFR-BL-04 — Operation direct-write execution does not consume the immutable prepared bytes

**Severity:** BLOCKER

**Evidence:** `preparedOperationDirectWrite` stores live `RuntimeConfig`, maps, form values, and body at `internal/connectors/engine/direct_write.go:38-59`. REST preparation seals `PreparedRequest.Body` at `:1578-1599` but also retains the object at `:1600-1620`; GraphQL seals encoded bytes at `:1699-1718` but retains `payload` at `:1720-1738`. Execution at `:138-163` ignores the sealed body and calls `DoJSONLimited` with `prepared.body`, which marshals again in `internal/connectors/connsdk/http.go:625-634`; multipart rereads both the file and `prepared.cfg.ApprovedPayloadSHA256`. `materializeConfigDefaults` (`engine/read.go:521-537`) returns the caller's maps unchanged when defaults are absent and only clones `Config` otherwise; `Secrets` and `ApprovedPayloadSHA256` always remain aliased.

**Cross-file call path:** `OperationDirectWrite` prepares/digests bytes -> approval callback blocks -> caller mutates `RuntimeConfig.Secrets`, config, payload digest map, or swaps the approved file -> gated closure resolves auth/config and rematerializes body/multipart -> connsdk sends bytes not necessarily represented by the approved digest.

**Behavior/security impact:** An approval can authorize one credential/endpoint/file/body identity while execution sends another. In particular, changing both a multipart file and the aliased approved-digest map after preview defeats the file binding without modifying the signed preview.

**Six-surface impact:** ETL=operation-backed destination writes unsafe; reverse ETL=blocked; direct read=none; direct write=blocked for REST/GraphQL/multipart; binary download=none; binary upload=blocked wherever operation multipart/binary preparation is reused.

**Exact fix:** Deep-copy the complete `RuntimeConfig` security-relevant state during preparation, including Config, Secrets, `ApprovedPayloadSHA256`, header maps, and query/form values. Execute the exact sealed body bytes (add a bounded `DoBytesLimited` internal method) rather than re-marshalling an object. For file parts, bind stable file identity and digest in the plan and reverify that sealed identity immediately before send; never consult a caller-owned digest map after approval.

**Exact behavioral regression test:** Pause the approval callback, concurrently mutate nested request input, Config, Secrets, `ApprovedPayloadSHA256`, and replace the upload file. Release and assert either the original previewed URL/headers/body/file bytes are sent or the operation refuses before network I/O. Run under `-race` for REST JSON, GraphQL mutation, form, and multipart cases.

### PFR-BL-05 — Direct-read responses and receipts are erased at command/native adapter boundaries

**Severity:** BLOCKER

**Evidence:** Engine legacy direct reads return a populated result and receipt with errors (`internal/connectors/engine/direct_read.go:480-521`), and the connector adapter preserves it. `internal/connectors/commandrunner/runner.go:569-615` instead returns `Result{}` on every legacy provider error (`:604-605`) and navigation assertion error (`:607-608`). The operation branch preserves an error result only when a receipt already exists (`:667-698`) but also erases it when the post-call navigation assertion fails (`:695-696`). The CLI emits a failure envelope only when commandrunner returns a direct/binary receipt (`internal/cli/cli.go:1344-1355`). Ashby's wrapper similarly discards the engine result on logical success-envelope failure (`internal/connectors/native/ashby/engine_delegate.go:111-119`).

**Cross-file call path:** connsdk provider response -> engine `DirectReadResult{Receipt}` plus error -> connector adapter -> commandrunner zeroes result (or Ashby zeroes it) -> CLI sees only an error and emits no terminal result envelope.

**Behavior/security impact:** Final provider status, headers, body/presence, operation identity, partial GraphQL data, and pagination evidence vanish on common failure paths. The CLI's nonzero exit remains, but the promised single terminal envelope/audit receipt does not.

**Six-surface impact:** ETL=source diagnostic/recovery data can be lost through command paths; reverse ETL=none; direct read=blocked for legacy, operation navigation, and Ashby; direct write=none; binary download=analogous command boundary must retain result; binary upload=none.

**Exact fix:** Preserve `commandrunner.Result` for every post-provider error and every post-provider validation/navigation failure, regardless of whether the receipt is currently nonnil. Native wrappers must return `result, err`, not a zero result. Let CLI emit one masked terminal envelope and then return the categorized nonzero execution error.

**Exact behavioral regression test:** Exercise a legacy direct read with 4xx, malformed JSON, pagination failure, and navigation assertion failure; an operation navigation failure; and an Ashby `{success:false}`/invalid envelope. Assert exactly one JSON terminal envelope with complete receipt/operation identity and a nonzero exit; stderr must not duplicate the body or leak credentials.

### PFR-BL-06 — Redirect, retry, and cancellation transitions discard the last provider response

**Severity:** BLOCKER

**Evidence:** Stream requests return only `ctx.Err()` at loop entry (`internal/connectors/connsdk/stream.go:102-107`), ignore a nonnil `resp` returned with `client.Do` error (`:137-150`), and replace a captured retry `HTTPError` with only the sleep cancellation (`:153-161`). Redirect-policy refusal therefore loses the 3xx status/headers even though Go can return the last response with the redirect error. Buffered requests have the same top-of-loop loss (`internal/connectors/connsdk/http.go:1186-1189`): after a captured 429/503 and retry, a cancellation or later transport failure can return a nil/current response and overwrite `lastErr` (`:1223-1243`, `:1296-1301`). `engine/binary_read.go:162-200` can construct a receipt only from a returned `StreamResponse` or typed `HTTPError`.

**Cross-file call path:** REST/binary/status request -> connsdk attempt receives 3xx/429/503 -> retry sleep/next attempt/cancellation -> connsdk returns only cancellation/transport error -> engine has no terminal response -> commandrunner/CLI cannot emit provider receipt or rate-limit evidence.

**Behavior/security impact:** The same request can have a real provider response but surface as “no response.” This destroys auditability, rate-limit parking inputs, retry diagnosis, and the terminal receipt contract. Redirect targets remain uncalled in guarded paths, but the refused provider response is lost.

**Six-surface impact:** ETL=REST source/destination retries; reverse ETL=retry evidence; direct read=REST affected; direct write=shared buffered transition risk; binary download=blocked; binary upload=shared stream/multipart transport transitions.

**Exact fix:** Retain the latest bounded provider response/typed `HTTPError` across attempts. On cancellation or later transport failure, return that terminal response and `errors.Join(lastProviderError, terminalCause)`. For refused redirects, use a response-retaining policy (`http.ErrUseLastResponse` or equivalent), capture the 3xx, then return a typed refusal without following it. Close every superseded body deterministically.

**Exact behavioral regression test:** Add tests for 302/307 refusal (target untouched, 3xx receipt retained), 503 followed by backoff cancellation (`errors.As` still finds `HTTPError`), and 429/503 followed by connection reset/context cancellation (last provider receipt/rate headers retained). Assert CLI emits one nonzero terminal envelope and no credential crosses origin.

### PFR-BL-07 — Status-check operations have no complete receipt and lose all error responses

**Severity:** BLOCKER

**Evidence:** `OperationStatusCheckResult` at `internal/connectors/connectors.go:835-846` contains only status, body byte count, and declaration-admitted headers; it has no shared `ProviderResponseReceipt`. `engine.OperationStatusCheck` (`internal/connectors/engine/status_check.go:26-79`) returns a zero result for any transport/redirect/read/header error even when `DoStatusCheck` returns response metadata. `commandrunner.runStatusCheck` (`runner.go:2393-2408`) erases the result on error, and CLI failure dispatch (`cli.go:1344-1355`) recognizes only direct-read and binary receipts. Success output at `cli.go:1392-1413` also omits complete ordinary response headers.

**Cross-file call path:** declared HEAD command -> commandrunner -> engine -> `Requester.DoStatusCheck` -> status projection only or zero result on error -> CLI success-only status envelope / generic error.

**Behavior/security impact:** A supposedly complete provider-output foundation has a reachable operation kind that cannot report a complete bounded provider receipt on success, redirect refusal, retry cancellation, or transport error. Status/trace/request IDs and repeatable headers can be lost.

**Six-surface impact:** ETL=status-based source/destination checks; reverse ETL=preflight/status operations; direct read=rest_status command blocked; direct write=none; binary download=none; binary upload=none.

**Exact fix:** Add `Receipt *ProviderResponseReceipt` to `OperationStatusCheckResult`, populate it from any received response/typed HTTP error, and preserve result-plus-error through commandrunner and CLI. HEAD bodies remain non-decoded but presence/bytes/raw policy must be explicit and bounded; ordinary headers stay complete while concrete credentials are masked at the public boundary.

**Exact behavioral regression test:** Cover 204 success, final 404 returned as status, redirect refusal, retry cancellation, and response-read error. Assert complete repeatable headers/status/body metadata on every received response, one terminal envelope on error, nonzero exit for transport/refusal failures, and no body decoding.

### PFR-BL-08 — Destination authorization can expire or be revoked long before the actual side effect

**Severity:** BLOCKER

**Evidence:** The contract explicitly says `Apply` must revalidate after warehouse reopen and before each mutation (`internal/synctransport/types.go:396-443`). Ordinary orchestration calls `AuthorizeNextUnit` before staging at `internal/synctransport/orchestrator.go:167-171`, then performs stage/reopen/cloning and invokes `ApplyDestination` at `:181-243` without another check. Full overwrite authorizes at `:396-400`, then stages/reopens and calls `ApplyFullOverwrite` at `:435-450`; only publication gets a fresh check (`:466-473`). Arrow serial authorizes before credit acquisition, transform, and segment storage (`arrow_fast_path_controller.go:96-151`), while the pipeline producer authorizes before queueing (`arrow_fast_path_pipeline.go:202-230`) and the consumer later transforms/stores/applies with no check (`:242-287`). App supplies a live revocation/expiry callback (`internal/app/declarative_typed_destination_approval.go:192-215`), and the generic declarative adapter correctly calls `IssueWriteEvidence` close to its write (`issue_label_warehouse_transport.go:228-249`), but native PostgreSQL validates only static evidence/expiry at apply entry (`native/postgres/transport_destination.go:119-139`) before resolution/provisioning, and Arrow segment apply never invokes either callback (`native/postgres/arrow_full_overwrite_transport.go:82-114`).

**Cross-file call path:** durable App authorization -> synctransport early `AuthorizeNextUnit` -> potentially blocking stage/reopen/queue/transform/provision -> authorization revoked or expires -> native/generic destination side effect proceeds -> checkpoint/output records a write performed without live authority.

**Behavior/security impact:** Revocation and expiry are not effective at the security boundary they claim to guard. Slow storage, queued Arrow work, database provisioning, or a blocked adapter can extend the gap arbitrarily and allow unauthorized destination mutation.

**Six-surface impact:** ETL=blocked for ordinary/full-overwrite/Arrow destinations; reverse ETL=destination writes can outlive authorization; direct read=none; direct write=adapter evidence path is the correct model; binary download=none; binary upload=affected when used as a destination side effect.

**Exact fix:** Require every destination adapter/session to invoke a non-serializable live authorization callback immediately before each actual mutation/publish, after all blocking preparation. The orchestrator may retain early admission for efficiency, but it cannot substitute for the adapter-side check. Full-overwrite and Arrow session APIs must carry the callback; native adapters must defensively enforce it.

**Exact behavioral regression test:** For ordinary, full-overwrite, Arrow serial, and Arrow pipeline paths, block after stage/reopen/queue/resolve but before provider/DB I/O, revoke or expire authorization, then release. Assert zero destination mutation, zero publish, and no checkpoint advance. Include multi-segment revocation between segments and verify earlier acknowledged effects are retained without replay.

### PFR-BL-09 — CLI numeric coercion destroys exact provider values and minimum semantics

**Severity:** BLOCKER

**Evidence:** The engine schema correctly compares exact numbers with `json.Number`/`big.Rat` (`internal/connectors/engine/schema.go:604-678`), but command metadata stores `Minimum *float64` in both `internal/connectors/command_surface.go:29-53` and `engine/bundle.go:1051-1069`. `commandrunner.validateFlagMinimum` converts even parsed `int64` to float64 (`runner.go:1433-1461`), while `coerceFlagValue` parses declared `number` flags with `strconv.ParseFloat` and integer with platform-sized `Atoi` (`:2089-2124`). Shipped GitHub number flags are present in `internal/connectors/defs/github/cli_surface.json:41095-41100`.

**Cross-file call path:** generated CLI numeric lexeme -> `CommandSurfaceFlag` float metadata -> commandrunner float coercion/minimum check -> request body/query map -> engine sees an already-rounded float -> prepared preview and wire JSON contain a different value from the user's input.

**Behavior/security impact:** Decimals such as `0.10000000000000001`, adjacent integers above 2^53 at minimum boundaries, and architecture-sized integers can be rounded, wrongly admitted/rejected, or sent as a different provider value. Approval then faithfully binds the wrong bytes.

**Six-surface impact:** ETL=typed numeric source/destination parameters; reverse ETL=blocked numeric writes; direct read=typed REST/GraphQL values; direct write=blocked; binary download=numeric path/query parameters; binary upload=numeric metadata parameters.

**Exact fix:** Carry numeric declaration and input lexemes as `json.Number`; compile minimum as exact rational/decimal metadata and compare with `big.Rat` before mapping. Use explicit int64/uint64 bounds where the provider declares integer range; never round through float64 before request construction.

**Exact behavioral regression test:** Feed `0.10000000000000001`, `9007199254740992`, `9007199254740993`, negative/exponent forms, and exact minimum/minimum-epsilon values through a real command. Assert preview JSON and server wire bytes preserve the lexeme's exact numeric value, exact boundary decisions, and consistent REST/GraphQL mapping on 32- and 64-bit builds.

### PFR-BL-10 — Binary no-overwrite publication can overwrite or delete a competing file

**Severity:** BLOCKER

**Evidence:** `streamBinaryDownloadToRoot` reserves the target with `O_CREATE|O_EXCL` (`internal/connectors/engine/binary_read.go:461-485`) and writes a temp file, but every failure cleanup unconditionally removes the target (`:487-500`, `:519-535`) and success uses `os.Root.Rename(tempName, fileName)` (`:530`), which replaces an existing destination. If another same-user process removes the zero-length placeholder and creates a real file while the download is in progress, success overwrites it; any later failure deletes it.

**Cross-file call path:** binary command -> engine stream response -> reserve empty destination -> concurrent process replaces placeholder -> engine success rename or error cleanup -> competing file overwritten/deleted -> CLI reports success/failure without the lost-file event.

**Behavior/security impact:** `allowOverwrite=false` does not provide no-overwrite semantics and can cause local data loss. Root confinement prevents path escape but does not establish ownership of a replaced directory entry.

**Six-surface impact:** ETL=none; reverse ETL=none; direct read=none; direct write=none; binary download=blocked data-loss risk; binary upload=the same atomic file-identity pattern must be avoided for staged artifacts.

**Exact fix:** Publish with an atomic no-replace primitive inside the opened root (`renameat2(RENAME_NOREPLACE)` where available or link/unlink using an owned inode), and remove only directory entries whose inode/handle identity is proven to be the reservation created by this call. Prefer keeping the final name absent until atomic no-replace publication rather than a replaceable placeholder.

**Exact behavioral regression test:** Coordinate a test hook after reservation and before publish/failure; replace the placeholder with a sentinel file. On both successful download and injected read/sync/close failure, assert the operation fails, sentinel contents/inode survive, temp files are cleaned, and no outside-root path is touched.

### PFR-BL-11 — Page cursors bypass request-field admission and encoded-size caps

**Severity:** BLOCKER

**Evidence:** CLI `connectorCommandPage` accepts `--page-cursor` without dangerous-character or byte checks (`internal/cli/cli.go:1740-1766`), and commandrunner deliberately deletes page flags before ordinary flag validation (`internal/connectors/commandrunner/runner.go:560-570`). Opaque cursor insertion at `engine/direct_read_paginate.go:599-609` is unbounded. For `next_url`/`link_header`, `admitDirectReadCursorURL` (`:466-501`) checks absolute URL, userinfo, same origin, and same path but explicitly allows any query; a caller can therefore provide the admitted path with undeclared query controls such as `?admin=true`. Native SQS inserts `req.PageCursor` into the signed form without bounds or character admission (`internal/connectors/native/amazon-sqs/direct_read.go:76-100`). GraphQL happens to reject dangerous characters and then relies on schema/body caps (`graphql_operation.go:563-619`), but the global CLI boundary does not guarantee it.

**Cross-file call path:** unvalidated CLI cursor -> commandrunner strips it from typed validation -> engine pagination treats it as an opaque token/absolute URL -> connsdk or native SQS authenticates and sends undeclared/unbounded query/form bytes.

**Behavior/security impact:** The post-fix closed query/body contract can be bypassed after initial request admission. Same-origin authenticated requests may carry arbitrary undeclared query parameters, and oversized/control-bearing cursor material can reach signing, allocation, logging, or provider I/O.

**Six-surface impact:** ETL=pagination inputs can escape declaration bounds; reverse ETL=none; direct read=blocked across REST/native/GraphQL global boundary; direct write=none; binary download=pagination-style continuation must not reuse this channel; binary upload=none.

**Exact fix:** Treat cursor provenance as authority: preferably emit an opaque signed CLI token binding connector, operation, exact allowed continuation query, and size, then accept only that token. At minimum, admit only declaration/paginator-owned query keys, reject collisions/unknown keys, and enforce dangerous-character plus exact percent/form-encoded byte caps on every cursor type before auth/runtime construction.

**Exact behavioral regression test:** Return a legitimate next cursor, replay it successfully, then test a same-origin/same-path URL with an undeclared query field, oversized UTF-8/percent-expanded cursor, CR/LF/control characters, duplicate paginator keys, and an oversized SQS `NextToken`. Assert all invalid cases fail before any server request while the exact emitted cursor remains usable.

### PFR-BL-12 — Native Amazon SQS direct reads never emit provider receipts

**Severity:** BLOCKER

**Evidence:** Amazon SQS is reachable from the production native factory (`internal/connectors/native/nativeset/factories.go:20-29`). Its `OperationDirectRead` returns a zero result on `doService` or XML decode errors and returns success without `Operation`, headers, or `Receipt` (`internal/connectors/native/amazon-sqs/direct_read.go:19-49`). The transport response type holds only status/body (`connection.go:166-169`), and `doEndpoint` discards every response on read, oversize, and non-2xx errors (`:201-238`). This implementation therefore does not satisfy the complete receipt contract added at `internal/connectors/connectors.go:474-513` even though commandrunner/CLI expose it through the same interface.

**Cross-file call path:** installed SQS command -> commandrunner operation direct read -> native SigV4/XML executor -> provider response -> native code discards metadata/error response -> commandrunner gets zero result -> CLI cannot emit receipt or terminal envelope.

**Behavior/security impact:** Successful SQS operations omit provider request/trace IDs and raw/body-presence truth; 4xx AWS XML errors, malformed XML, oversize bodies, and read failures lose the complete response. This creates a provider-specific hole in a shared public contract and makes SQS failures unauditable.

**Six-surface impact:** ETL=SQS source observability/parity affected; reverse ETL=SQS write path should share the fixed receipt model; direct read=blocked for six installed SQS operations; direct write=none; binary download=none; binary upload=none.

**Exact fix:** Extend `sqsHTTPResponse` with cloned headers, raw bytes, presence/byte metadata, and response-received state; return it alongside every post-I/O error. Populate and publicly sanitize `ProviderResponseReceipt`, set the operation identity, and return result-plus-error through commandrunner/CLI. Mask exact access/session credential material without deleting ordinary AWS IDs or headers.

**Exact behavioral regression test:** For each installed SQS direct-read class, cover 200 success, AWS 4xx XML, malformed XML, cap+1, and injected read error. Assert status, duplicate headers/request ID, raw/decoded bounded body, operation/path, and receipt survive; CLI emits one envelope and nonzero exit on error; concrete AWS credentials never appear.

### PFR-BL-13 — Native Amazon SQS can forward a SigV4 session credential across origins

**Severity:** BLOCKER

**Evidence:** `internal/connectors/native/amazon-sqs/connection.go:201-223` sets `X-Amz-Security-Token`, signs the request, and sends with a default client. `transportpolicy.HTTPClient` at `internal/connectors/transportpolicy/policy.go:22-41` returns the ambient client unchanged unless the context is marked destructive. `OperationDirectRead` is logically a read and does not mark the context, although the SQS protocol uses POST. Go's default redirect copier treats `Authorization` specially but `X-Amz-Security-Token` is a custom header and can be copied to a cross-origin redirect. The provider-specific client has no redirect policy, credential-header stripping, or retained redirect receipt.

**Cross-file call path:** installed SQS direct read -> `doEndpoint` adds session token/SigV4 -> `transportpolicy.HTTPClient` no-ops for non-destructive context -> server returns 302/307 to another origin -> default `http.Client` follows and copies the session token -> native result/receipt path loses the redirect history.

**Behavior/security impact:** A malicious/compromised endpoint or redirect can exfiltrate temporary AWS session credentials to another origin. Even where the signature becomes invalid for the new host, the session token itself is concrete reusable credential material.

**Six-surface impact:** ETL=SQS source reads can leak credentials; reverse ETL=SQS writes are protected only when marked destructive and need a uniform native policy; direct read=blocked; direct write=none; binary download=none; binary upload=none.

**Exact fix:** Clone the client for every SQS request and fail closed on all redirects (or admit an explicitly declared same-origin policy only). Strip every auth-added header on any allowed origin change, including `X-Amz-Security-Token`, `Authorization`, `X-Amz-Date`, and related signing headers. Retain the 3xx response as a typed, bounded receipt without contacting the target.

**Exact behavioral regression test:** Start an origin server that 302/307 redirects to a second server. Use session credentials and the production `New()` connector. Assert the second server receives no request (or at minimum no credential/signing header under an explicit allowed policy), the first 3xx is retained in the result, CLI exits nonzero with one safe envelope, and same-origin non-redirect success still works.

## Warnings

### PFR-WR-01 — Declaration-owned idempotency headers are accepted, previewed, then silently stripped

**Severity:** WARNING

**Evidence:** `IsProtectedOperationHeaderName` (`internal/connectors/operation_headers.go:22-56`) does not protect `Idempotency-Key` or `X-Idempotency-Key`. Engine validation/admission accepts any non-protected declared bounded header (`internal/connectors/engine/operation_headers.go:46-108`, `:201-309`) and direct-write preparation binds it into the preview. Execution sets `DisableRetries=true`, and `disableTransportReplay` silently deletes both headers (`internal/connectors/connsdk/http.go:297-303`). Current operation declarations do not appear to expose these as caller headers, so the defect is latent rather than a currently demonstrated installed command failure; reverse-write generated idempotency uses a separate action-owned mechanism.

**Cross-file call path:** future/generated operation header declaration -> CLI flag -> engine validates and digest-binds exact value -> direct-write requester disables retries -> connsdk deletes header -> provider receives a request different from preview.

**Behavior/security impact:** The type system advertises an executable header that the wire layer removes, creating approval drift and potentially removing provider deduplication.

**Six-surface impact:** ETL=latent destination operation risk; reverse ETL=separate generated mechanism currently unaffected; direct read=header deletion applies when retries are disabled; direct write=latent drift; binary download=none; binary upload=latent.

**Exact fix:** Either reject these names as runtime-owned during bundle validation until the operation executor can preserve a declaration-owned key safely, or preserve the exact preview-bound header while independently disabling automatic replay. Never silently mutate a prepared request.

**Exact behavioral regression test:** Add an operation declaring each header. Assert bundle preflight rejects it before I/O, or—if supported—the exact value appears in preview and reaches the server once with redirects/retries disabled. Assert duplicate/alias forms remain rejected.

### PFR-WR-02 — Valid structured REST schemas with `minLength` can be rejected as unreachable

**Severity:** WARNING

**Evidence:** `compileStructuredRESTBodySchema` computes a minimum valid JSON witness before accepting the schema (`internal/connectors/engine/structured_rest_body.go:682-768`). Its string witness generator returns only `""` when no pattern exists (or `"x:"` for URI) at `:1785-1814`; it never accounts for supported `minLength`. The shared schema compiler explicitly parses and validates `minLength` (`internal/connectors/engine/schema.go:315-339`, `:434-447`). Thus a valid required string with `minLength: 1` can yield “cannot prove a schema-valid string witness,” making an otherwise supported structured operation unreachable. Current generated structured operations did not prove an installed occurrence, so this is a correctness/quality warning rather than a blocker.

**Cross-file call path:** generated structured body schema -> engine bundle/operation compilation -> minimum byte witness -> empty candidate fails `minLength` -> operation rejected before command execution despite runtime validator supporting the value.

**Behavior/security impact:** Valid provider operations can be falsely blocked, and future generation may require provider-specific workarounds or broadened schemas to become executable.

**Six-surface impact:** ETL=latent typed destination schema issue; reverse ETL=structured writes affected; direct read=structured POST body compiler reuse risk; direct write=affected; binary download=none; binary upload=multipart JSON metadata if routed through structured bodies.

**Exact fix:** Synthesize a bounded witness meeting `minLength` and `maxLength`, then combine it with pattern/format witnesses and validate the final candidate. Count Unicode code points for schema cardinality and encoded JSON bytes for the request cap.

**Exact behavioral regression test:** Compile and execute an object schema with a required `minLength:1` string plus a nested object/array that selects the structured path. Cover `minLength==maxLength`, Unicode, URI+minimum, pattern+minimum, and impossible min>max. Assert valid preview/wire bytes and existing fail-closed bounds.

### PFR-WR-03 — Multipart symlink-boundary regression test has a data race and can false-pass

**Severity:** WARNING

**Evidence:** `internal/connectors/connsdk/multipart_bounds_test.go:24-45` writes through `*got` from the HTTP handler goroutine, while `TestRequesterDoMultipartRefusesEscapingSymlinkSwappedAfterValidation` reads `sawFile` at `:105-107` after `DoMultipart` may return early, before the handler has finished. `go test -race` reproduced the conflict: handler write at line 43 versus test read at line 105.

**Cross-file call path:** test starts server -> multipart request can fail/return while handler still parses body -> assertion reads unsynchronized boolean -> later handler observes leaked file bytes and writes true after assertion.

**Behavior/security impact:** The security regression test can miss an actual outside-root upload and pass spuriously; the race also prevents a clean race-gate signal for connsdk.

**Six-surface impact:** ETL=none; reverse ETL=multipart destination test confidence; direct read=none; direct write=multipart safety test unreliable; binary download=none; binary upload=security boundary coverage unreliable.

**Exact fix:** Report handler observation through a channel or guarded/atomic value and wait for handler completion (or server shutdown) before asserting. Preserve deterministic request cancellation and do not infer completion from the requester returning.

**Exact behavioral regression test:** Make `go test -race -run TestRequesterDoMultipartRefusesEscapingSymlinkSwappedAfterValidation ./internal/connectors/connsdk` pass repeatedly while synchronizing handler completion; prove the server never observes the outside-root sentinel bytes.

## Prior-Finding Closure Cross-Check

The independent finding set above was frozen before `.planning/phases/cli-current-foundations-main-integration-r1/REVIEW.md` and `REVIEW-FIX.md` were read. The fix report claims all 36 canonical findings closed at `2d656b8b632defbae4f43e08c3f4aaf6fab96942`; the statuses below assess the relevant runtime/security claims at the current frozen HEAD.

| Prior finding | Claimed fix | Current closure status | Current evidence |
| --- | --- | --- | --- |
| `CF-B08` public receipts disclose credentials | Exact configured-secret masking at public boundaries | **PARTIAL / REOPENED** | Sanitizers exist, but escaped raw JSON and GraphQL metadata leak while short secrets corrupt provider truth (`PFR-BL-01`, `PFR-BL-02`). |
| `CF-B09` CLI discards failed ETL/reverse runs | Persist/render terminal runs | **CLOSED for the cited run paths** | No regression found in the cited ETL/reverse terminal-run flow; direct-read/status loss is a different command-result path (`PFR-BL-05`, `PFR-BL-07`). |
| `CF-B10` no-response direct write loses attempted identity | Initialize identity before I/O | **CLOSED for operation direct write** | `direct_write.go:126-136` now initializes attempt identity before transport. |
| `CF-B11` direct-read/binary receipts incomplete | Shared complete receipt on success/error | **PARTIAL** | Shared type exists, but raw declared secrets, runner/native erasure, retry transitions, status checks, and SQS omit/drop receipts (`PFR-BL-01`, `PFR-BL-05`–`07`, `PFR-BL-12`). |
| `CF-B12` accepted-success mismatch drops direct-write receipt | Return response plus typed mismatch | **CLOSED** | Buffered requester returns terminal response with `UnexpectedStatusError` at `http.go:1253-1260`; no contrary direct-write path found. |
| `CF-B13` idempotent writes redirect/retry outside approval or lose final response | Separate redirect/retry and preserve final response | **CLOSED for the cited declarative mutation path; broader transition gap remains** | Strict write redirect policy is present, but stream/buffer retry cancellation can still lose the last provider response (`PFR-BL-06`). |
| `CF-B14` buffered declarative cross-origin redirect carries credentials | Fail-closed/default credential stripping | **PARTIAL across runtime** | Declarative requester is repaired; native Amazon SQS bypasses that policy and can forward `X-Amz-Security-Token` (`PFR-BL-13`). |
| `CF-B15` previews disclose resolved credentials | Safe public preview projection | **CLOSED for reviewed declarative preview path** | No concrete preview disclosure regression proved. Approval still binds the wrong executor/bytes in hook/direct-operation paths (`PFR-BL-03`, `PFR-BL-04`). |
| `CF-B16` REST/binary diagnostics expose provider body/query | Safe typed diagnostics | **CLOSED for cited formatters** | No raw query/body reconstruction regression proved in the cited engine formatter path. |
| `CF-B17` writes rematerialize after approval | Execute sealed prepared plan once | **PARTIAL / REOPENED** | Ordinary declarative record mapper is pinned, but `WriteHook` executes original records/different sequences and operation direct writes re-marshal live plan/config state (`PFR-BL-03`, `PFR-BL-04`). |
| `CF-B18` empty JSON fabricates body presence | Derive presence from transport bytes | **CLOSED** | Shared receipt materialization derives presence from captured bytes; no contradictory case proved. |
| `CF-B19` numeric enum loses >2^53 precision | Exact `big.Rat` schema comparison | **PARTIAL at CLI boundary** | Engine enum is exact, but CLI metadata/coercion still rounds number inputs and minimums (`PFR-BL-09`). |
| `CF-B20` duplicate singleton flags are last-wins | Reject duplicate canonical targets | **CLOSED for ordinary typed flags** | No duplicate-target regression proved; page cursor is removed from this validation and has separate admission gaps (`PFR-BL-11`). |
| `CF-B21` path/query values lack byte caps | Project/enforce encoded caps | **PARTIAL** | Ordinary declared parameters are bounded, but `--page-cursor` bypasses flag validation and cursor query/form insertion is unbounded (`PFR-BL-11`). |
| `CF-B22` direct reads admit undeclared query/body fields | Closed exact bindings | **PARTIAL** | Initial operation bindings are closed; next-url/link cursors can introduce arbitrary same-path query fields after admission (`PFR-BL-11`). |
| `CF-B23` name-based redaction deletes ordinary IDs | Remove key-name heuristics | **REOPENED** | Exact-secret substring replacement mutates keys/header names, and GraphQL retains a keyword heuristic (`PFR-BL-02`). |
| `CF-B24` side effects precede ownership/authorization guard | CAS/authorization before provider I/O | **PARTIAL for authorization freshness** | Early CAS/admission does not satisfy the explicit live revalidation requirement after stage/queue/provision and immediately before each destination effect (`PFR-BL-08`). |
| `CF-W02` path filters not post-validated | Revalidate final path output | **CLOSED in reviewed path interpolation** | No contrary filtered-path case proved. |
| `CF-W03` accepted secret transform has no executor | Registered implementation/fail closed | **CLOSED in reviewed transform path** | Implementation is present and no unsupported accepted transform was proved. |

## Final Verdict

**Verdict: `blockers`.** Thirteen release-blocking correctness/security failures remain at the frozen HEAD. Merge should remain blocked until every `PFR-BL-*` item is fixed with the exact behavioral regression coverage above, `PFR-WR-03` is race-clean, and the full targeted package/race suite passes at one unchanged source SHA.

---

_Reviewed: 2026-08-21T03:15:01Z_
_Reviewer: the agent (gsd-code-reviewer)_
_Depth: deep_
