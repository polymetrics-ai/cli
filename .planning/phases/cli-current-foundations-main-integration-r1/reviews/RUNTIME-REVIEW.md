---
reviewed_sha: 9e5329f34e015e39160bb8e951452bbd071a698a
depth: deep
scope_count: 30
status: issues_found
finding_count: 15
blocker_count: 13
warning_count: 2
---

# Runtime Deep Review

## Review boundary and method

This is a discovery-only adversarial review of the immutable foundation-rollup snapshot `9e5329f34e015e39160bb8e951452bbd071a698a`. The initial guard passed: `HEAD` matched exactly and `git status --porcelain=v1 --untracked-files=all` was empty before review work began. CodeGraph was invoked first as required, but this worktree has no `.codegraph` index, so the review continued with line-numbered source inspection, repository call-site tracing, combined-diff inspection, and focused Go diagnostics.

The required convergence, plan, TDD, verification, input-manifest, and evidence-manifest artifacts were read before the runtime audit. All 30 requested files were reviewed in full. Calls were followed read-only through App persistence, operation header policy, binary download, approval gating, hook contracts, and provider definitions where necessary. The review did not modify source, fix findings, commit, push, create or modify a PR, or merge.

The final implementation is not releasable as reviewed. Thirteen findings are correctness, secret-disclosure, target-authority, receipt-loss, or duplicate-mutation blockers. Two warnings identify a latent path-safety gap and a hard-coded transform contract with no executor.

## Architecture and call-path map

```text
installed CLI flags
  -> commandrunner validate + coerce + typed target maps
     -> connector interfaces in connectors.go
        -> engine.Base connector-neutral dispatch
           -> OperationDirectRead / fixed GraphQL / binary operation
           -> PreviewOperationDirectWrite
                -> preparedOperationDirectWrite -> PreparedWrite digest
                -> approval evidence -> OperationDirectWrite
           -> DryRunWrite
                -> prepareDeclarativeWrite -> PreparedWrite digest
                -> approval evidence -> Write -> executeApprovedWrite
                    -> scalar / JSON / structured JSON / form / SCIM /
                       multipart / GraphQL / base64 paths
              -> runtime.requesterFor -> connsdk.Requester
                 -> auth + declared headers + rate admission + retry/redirect
                 -> captured status + repeatable headers + bounded bytes
                    -> typed engine result/cause
                       -> connectors output sanitizer
                          -> App/CLI persistence and printable output
```

The principal trust boundaries are (1) source declaration to installed command metadata, (2) CLI occurrence/coercion to typed request maps, (3) prepared digest/approval to the actual transport request, (4) requester retry/redirect to the approved origin, and (5) complete internal provider receipts to secret-safe public output. The combined merge `808896a28873c5f0479fa10e2f798da56f885b5e` edited `connectors.go`, `connsdk/http.go`, `direct_write.go`, and `write.go` at those same seams. Several findings below are specifically composition regressions: exact receipt retention was combined with no-op public sanitization, accepted-status validation was placed before response capture, and retry eligibility was conflated with write redirect/terminal-response policy.

## Findings

### RT-B01 — Public write-result sanitizers deliberately persist credentials verbatim

**Severity:** BLOCKER

**Evidence:** `internal/connectors/connectors.go:921-932` labels the two App/CLI boundary functions as sanitizers but returns both result graphs unchanged and ignores the supplied secret map. `internal/connectors/engine/direct_write.go:1207-1212` disables declared response-value redaction for secret operations and explicitly states that secrets remain in responses, errors, logs, previews, reports, and fixtures. `internal/connectors/engine/direct_write.go:2226-2241` places raw body bytes, decoded body values, repeatable headers, and declared output-secret field names in the typed result. Read-only caller tracing shows those no-op sanitizers immediately precede persistence at `internal/app/app.go:2924-2929` and `internal/app/app.go:3149-3155`. The regression is locked in by `internal/connectors/write_result_output_test.go:12-65`, which requires configured credential literals in raw bodies, headers, nested decoded values, numeric values, base64 echoes, and GraphQL errors to survive the public sanitizers unchanged; `internal/app/rest_write_command_test.go:738-785` likewise requires a provider-returned `server-token` in the persisted result.

**Impact:** An API that echoes an access token, API key, cookie, generated credential, or submitted secret exposes it in durable reverse-run JSON and CLI output. The encrypted secret-store write does not protect those independent public copies. This violates the explicit boundary rule: retain the authorized internal receipt, preserve ordinary provider IDs/occurrence IDs, and mask only actual credentials at printable/persisted boundaries.

**Root cause:** Conflict resolution treated “complete internal provider receipt” and “complete public serialization” as the same contract. Secret safety was applied only to generated error strings, while functions named as public sanitizers were intentionally made no-ops.

**Exact change plan:** Preserve the engine's internal typed result and wrapped provider cause byte-for-byte. At `SanitizeWriteResultForOutput` and `SanitizeOperationDirectWriteResultForOutput`, deep-clone the result and mask only concrete configured/selected-auth secret literals plus declared `OutputSecretFields` in `BodyRaw`, decoded maps/slices, repeatable header values, and GraphQL error strings. Mark affected response headers with `Masked`/`Redacted`; never delete ordinary fields merely because their name contains `token`, `secret`, or `credential`. Apply known credential encodings only where the selected auth/transform proves them. Do not mutate the internal receipt.

**Required tests:**

- Happy: ordinary IDs, occurrence IDs, provider fields named `token`, large numbers, duplicate-key raw JSON, binary bytes, and repeatable non-secret receipts survive unchanged.
- Bad: a configured bearer/API-key value echoed in a JSON/text/binary raw body, nested decoded value, response header, GraphQL error, and declared response-secret field is masked in persisted/printable output while remaining exact in the internal typed receipt and encrypted store.
- Edge: overlapping secret values, numeric secrets, known base64 encodings, empty values, secret-looking field names with non-secret values, and sanitizer non-mutation of the input object.

### RT-B02 — Accepted-success-status mismatch discards the entire terminal receipt

**Severity:** BLOCKER

**Evidence:** `internal/connectors/connsdk/http.go:1210-1212` closes the body and returns `(nil, *UnexpectedStatusError)` before reading bytes or cloning response metadata when a 2xx status is outside the declaration's `success_statuses`. Operation declarations install those ranges at `internal/connectors/engine/operation_headers.go:425-447`. The direct-write result is populated only when the requester returned a non-nil response (`internal/connectors/engine/direct_write.go:155-160`), then reports only an error at `internal/connectors/engine/direct_write.go:161-182`. Existing response-preservation tests cover terminal non-2xx and producer failures, but no scoped test combines `AcceptedStatuses` with an undeclared 2xx response.

**Impact:** A provider returning 200/201/202/204 outside a narrowly declared success set loses status, repeatable headers, exact body bytes, body-presence state, operation identity, and path—the precise terminal-error receipt the convergence contract requires.

**Root cause:** Accepted-status checking was composed into the requester before the shared terminal response capture step.

**Exact change plan:** Read at most the declared cap plus one, close the body, build the terminal `Response`, and then return that response together with `UnexpectedStatusError`. Keep it terminal and non-retryable. Ensure the error retains a typed cause without copying body content into printable diagnostics.

**Required tests:**

- Happy: each declared exact/ranged 2xx status succeeds with the full receipt.
- Bad: undeclared 202 returns both a typed result and `UnexpectedStatusError`, with exact raw body and both occurrences of a repeatable response header.
- Edge: empty 204, literal JSON `null`, binary body, and cap+1 body all preserve correct presence/size metadata and the typed terminal cause.

### RT-B03 — Idempotency-enabled reverse writes can follow unapproved redirects and lose final failures

**Severity:** BLOCKER

**Evidence:** `internal/connectors/engine/write.go:605-631` enables requester retries whenever a write action declares an idempotency-key header. `internal/connectors/connsdk/http.go:1138-1143` defines “strict write” as a non-safe method only when `DisableRetries` is true, so enabling retries also removes the mutation redirect refusal. `internal/connectors/connsdk/http.go:1253-1273` returns a final non-2xx response only for status checks, strict writes, or a destructive transport context; an idempotent, non-approval write therefore returns a nil response after its last 4xx/5xx. The approval gate marks the context destructive only for targets requiring approval (`internal/connectors/engine/write_gate.go:98-114`). The existing reverse-write redirect test at `internal/connectors/engine/write_test.go:120-154` covers a destructive no-retry write; it does not cover an action with an idempotency header.

**Impact:** A 307/308 can replay the approved method and body to a URL absent from the prepared digest; 301/302/303 can cause an unapproved follow-up request. Custom auth/default headers can accompany that redirect (see RT-B04). After retries, a terminal provider 4xx/5xx can also be omitted from `ProviderResponses`, losing status, headers, and exact bytes.

**Root cause:** One boolean (`DisableRetries`) simultaneously controls requester retries, transport replay protection, redirect refusal, and terminal-response retention. Idempotency is evidence for same-target retry, not authority to change targets or discard a final receipt.

**Exact change plan:** Introduce an explicit write transport policy independent of retry count. Every write must refuse redirects and return the final terminal response; only same-original-URL attempts may retry, and only when a stable provider-scoped idempotency key is bound into the prepared plan. Keep the generated key identical across those attempts and retain the final response with the error.

**Required tests:**

- Happy: an idempotency-declared transient failure retries the original endpoint with an identical key and records the final success receipt.
- Bad: 301/302/303/307/308, same-origin and cross-origin, never hit the redirect target; final 400/429/500 after retry budget returns the exact terminal receipt.
- Edge: multipart/bodyless writes, auth refresh, custom API-key headers, cancellation between attempts, and a provider that changes receipt headers on each attempt.

### RT-B04 — Buffered declarative HTTP follows cross-origin redirects with custom credentials

**Severity:** BLOCKER

**Evidence:** Runtime construction leaves `Requester.RedirectPolicy` nil (`internal/connectors/engine/read.go:565-571`). `internal/connectors/connsdk/http.go:1139-1174` calculates every auth/default credential header key but passes a nil policy into the client. `internal/connectors/connsdk/stream.go:214-229` returns the ordinary client unchanged when that policy is nil, so the key list is never used; the credential stripping at `internal/connectors/connsdk/stream.go:232-250` runs only inside an explicit redirect callback. Declarative reads invoke this buffered requester directly at `internal/connectors/engine/read.go:355-368`. Operation-backed paths are safer because `operationRedirectPolicy` installs a non-nil fail-closed policy, but the ordinary declarative read/write runtime does not.

**Impact:** Go's default redirect handling protects a small standard sensitive-header set, but a connector's custom `X-API-Key`, vendor auth header, or other declaration-owned credential/default header can be copied to another origin. The same nil-policy condition enables the approved-target escape in RT-B03.

**Root cause:** The stream client implements a safe nil-policy fallback, but the buffered client's nil branch bypasses all redirect enforcement and credential stripping.

**Exact change plan:** Give buffered requests the same explicit default as the stream path: bounded hops, no scheme downgrade, same-origin only unless a declaration names allowed hosts, and credential/default header stripping before any permitted cross-origin hop. A caller should never need to opt in merely to obtain credential safety.

**Required tests:**

- Happy: a bounded same-origin GET redirect succeeds under an explicit/default same-origin policy.
- Bad: cross-origin and HTTPS-to-HTTP redirects carrying `X-API-Key`, custom vendor auth, `Authorization`, or cookies do not expose those values and fail unless the declaration authorizes the destination.
- Edge: subdomains, explicit ports, host case, redirect loops/max hops, relative locations, and allowed cross-origin downloads with all credential-derived headers stripped.

### RT-B05 — Reverse-write preview warnings print resolved credentials and declared sensitive identifiers

**Severity:** BLOCKER

**Evidence:** `internal/connectors/engine/write.go:100-109` returns the prepared warning list directly. The preparation path resolves the full URL and appends it verbatim to warnings at `internal/connectors/engine/write_prepare.go:62-69`. `internal/connectors/engine/write_test.go:895-928` explicitly requires a configured secret embedded in the base URL to appear in that warning. `internal/connectors/engine/write_test.go:1000-1057` similarly requires path fields declared in `RedactFields`, including nested patient UUIDs, to appear. App planning returns these previews at `internal/app/app.go:2125-2136` and `internal/app/app.go:2229-2254`.

**Impact:** Dry-run/plan output, persisted previews, logs, and evidence can contain URL credentials or specifically classified sensitive record values before execution. A preview intended as a safety boundary becomes a disclosure boundary.

**Root cause:** The private prepared request, digest projection, and public human warning share the same fully resolved URL representation.

**Exact change plan:** Keep the exact resolved URL, query, headers, and body only in the private prepared request/digest. Render warnings from declaration identity plus a separately redacted URL: replace concrete configured secrets and declaration-owned `RedactFields` values, strip userinfo/query secrets, and preserve ordinary non-sensitive IDs. Update tests that currently mandate leakage.

**Required tests:**

- Happy: method, operation name, declared route shape, and an ordinary provider ID remain useful in the warning.
- Bad: a base-URL secret, query secret, and declared sensitive top-level/nested path value never appear, while changing any of them still changes the prepared digest.
- Edge: repeated/overlapping literals, percent-encoded values, empty optional values, multi-record previews, and an ordinary field named `token` that is not classified sensitive.

### RT-B06 — REST and binary error formatting bypasses the safe HTTP error boundary

**Severity:** BLOCKER

**Evidence:** `internal/connectors/engine/direct_read.go:44-58` unwraps `connsdk.HTTPError` and reconstructs printable text from the raw URL, query, and provider body. That helper is used without redaction for operation reads at `internal/connectors/engine/direct_read.go:145-155`, generic direct reads at `internal/connectors/engine/direct_read.go:329-339`, and binary downloads at `internal/connectors/engine/binary_read.go:160-173`. This deliberately bypasses `HTTPError.Error`, whose safe boundary strips URL query/body patterns at `internal/connectors/connsdk/http.go:55-74`; the scoped regression at `internal/connectors/connsdk/http_test.go:1194-1203` proves that safe behavior only at the lower boundary. GraphQL adds `safety.RedactErrorText` at `internal/connectors/engine/graphql_operation.go:627-641`, demonstrating the divergence.

**Impact:** A secret in a request query or a plaintext provider error body can reach CLI/log output. Binary failures have the same leak despite correctly producing no output file. The wrapped response formatter also intentionally exposes only selected typed causes, so callers get unsafe text without a generally usable retained HTTP receipt.

**Root cause:** A convenience formatter was introduced specifically to recover fields that the SDK's public `Error()` intentionally redacts, and callers then treated those fields as printable diagnostics.

**Exact change plan:** Never render typed `HTTPError.Body` or raw query text. Produce a bounded message from status plus declaration-owned method/path identity, apply concrete configured-secret masking at the public boundary, and retain the typed HTTP cause privately for classification/parking. Share one formatter across REST, GraphQL, and binary paths.

**Required tests:**

- Happy: diagnostics retain operation identity, declared path template, status, class, and safe hint.
- Bad: REST-read and binary 4xx/5xx responses containing a plaintext secret, JSON credential, and secret query parameter never print those values; `errors.As` still reaches the typed internal cause where authorized.
- Edge: 401 classification, 429 reset evidence, huge/truncated bodies, invalid UTF-8, and provider text that resembles a URL or terminal control sequence.

### RT-B07 — Declarative reverse writes execute a second materialization after approval

**Severity:** BLOCKER

**Evidence:** `internal/connectors/engine/write_prepare.go:57-76` invokes `applyWriteRecordHook`, then materializes and digest-binds each mapped record. `internal/connectors/engine/write.go:302-320` authorizes that `PreparedWrite`, but the execution closure calls `executeApprovedWrite` with the original records. `internal/connectors/engine/write.go:328-347` documents the mapper, and `internal/connectors/engine/write.go:350-381` invokes it again immediately before rebuilding/sending the request. The scoped tests only verify the first prepared body and non-mutation (`internal/connectors/engine/write_record_hook_test.go:89-133`); none executes a stateful mapper through approval.

**Impact:** A time-dependent, counter-based, random, external-state, or accidentally stateful mapper can send a path/body different from the request digest the operator approved. For safe actions the second materialization can diverge without approval evidence at all; for destructive actions a mapper that is stable during preview/revalidation but changes during execution bypasses the digest binding and can duplicate or misdirect a mutation.

**Root cause:** `PreparedWrite` is used as a digest description, not as the immutable execution plan. The runtime rebuilds request material from the original records and assumes hooks are pure, although the hook interface provides no purity/determinism contract.

**Exact change plan:** Materialize once into an immutable execution plan and have the gated closure consume that plan. If transport setup requires rematerialization, canonicalize and compare method, URL/query, headers, content type, body bytes, multipart content identities, and hook identity to the approved projection immediately before every send; refuse before I/O on any mismatch. Do not invoke `MapWriteRecord` twice.

**Required tests:**

- Happy: one mapper invocation per record and execution bytes equal the approved prepared bytes for scalar, form, SCIM, GraphQL, multipart, and base64 routes.
- Bad: counter/random/time-dependent mappers cannot produce an unapproved request and no network call occurs on divergence.
- Edge: multiple records/order, `handled=false`, mapper error, caller record mutation attempts, multipart file replacement, and cancellation after approval but before dispatch.

### RT-B08 — Both write receipt parsers fabricate body presence for empty JSON

**Severity:** BLOCKER

**Evidence:** Reverse-write parsing sets `BodyPresent=true` solely because a response declares JSON at `internal/connectors/engine/write.go:520-551`, even when `len(response.Body)==0`. Operation-direct-write parsing similarly initializes `present:true` whenever policy/content type implies JSON at `internal/connectors/engine/direct_write.go:2244-2269`. The direct-write test at `internal/connectors/engine/direct_write_test.go:305-409` explicitly requires an empty `Content-Type: application/json` response to have `BodyPresent=true`; reverse-write cases at `internal/connectors/engine/write_test.go:1382-1483` check decoding/status but omit the presence assertion.

**Impact:** Durable receipts cannot distinguish “provider sent no body” from “provider sent a present JSON value.” This breaks exact response-presence preservation and makes an empty invalid JSON response look partly like a present zero-byte payload. The two independently implemented parsers also drift in their empty/non-JSON handling.

**Root cause:** Body presence is inferred from declared parsing policy rather than transport bytes, and response materialization is duplicated across the reverse and operation-direct paths.

**Exact change plan:** Derive presence only from actual captured transport bytes (`len(raw)>0`, or an explicit transport presence bit if later required), then parse according to content type/policy. Share a single receipt materializer so both result types agree on empty, whitespace, null, text, binary, oversize, and malformed JSON semantics. An empty JSON response may still produce a decode error, but its retained receipt must say `BodyPresent=false`.

**Required tests:**

- Happy: non-empty JSON/text/binary and literal JSON `null` are present with exact size/bytes in both paths.
- Bad: zero-byte JSON with status 200 returns a decode error plus a receipt with `BodyPresent=false`; zero-byte 204/non-JSON succeeds with false presence.
- Edge: whitespace-only JSON, invalid UTF-8, cap/cap+1, omitted/invalid content type, and repeatable headers on each case.

### RT-B09 — Numeric enum validation collapses distinct integers above 2^53

**Severity:** BLOCKER

**Evidence:** `internal/connectors/engine/schema.go:627-665` converts `json.Number`, `int`, and `int64` enum operands to `float64` and compares the rounded values. Thus adjacent integers such as `9007199254740992` and `9007199254740993` normalize to the same IEEE-754 value. `internal/connectors/engine/schema_test.go:127-136` tests only string enum equality and has no large-number regression.

**Impact:** A typed structured body/GraphQL variable can pass a declaration-owned numeric enum while carrying a different provider identifier or control value. Validation claims exact source-declaration membership but can authorize an undeclared mutation value.

**Root cause:** The generic enum comparator chose float normalization despite the decoder's deliberate `UseNumber` fidelity contract.

**Exact change plan:** Parse numeric operands into exact canonical integers/rationals (for example `math/big.Int`/`big.Rat`) and define JSON-number equivalence explicitly, including whether `1`, `1.0`, and `1e0` are equal. Reject non-finite programmatic floats before comparison.

**Required tests:**

- Happy: exactly equal small/large integers and declared decimal/exponent equivalents match according to the documented rule.
- Bad: adjacent integers above and below 2^53, large negatives, and nearby decimals do not match and fail before I/O.
- Edge: int/int64/json.Number/float64 pairs, exponent notation, negative zero, overflow, NaN, and infinities.

### RT-B10 — Singleton typed CLI flags silently use their last occurrence

**Severity:** BLOCKER

**Evidence:** `internal/connectors/commandrunner/runner.go:1497-1515` performs occurrence checks only for direct-write query flags and non-repeatable headers. Other singleton targets reach `coerceFlagValue`, which validates every input then unconditionally selects the last at `internal/connectors/commandrunner/runner.go:2023-2037`; literal text bodies do the same at `internal/connectors/commandrunner/runner.go:2005-2020`. Only body mappings from different flags are rejected later (`internal/connectors/commandrunner/runner.go:1609-1615`). The scoped test `internal/connectors/commandrunner/runner_test.go:2208-2259` covers duplicate query occurrences but not path, scalar body/form/GraphQL variables, or raw text body.

**Impact:** Repeated command-line values for a supposedly exact path/body mapping are silently last-wins. That creates ambiguous previews, makes wrapper/alias mistakes hard to detect, and violates the required duplicate-before-I/O rejection contract.

**Root cause:** Repeatability is interpreted only for header and array assembly; the general coercer treats a slice of occurrences as a convenience input even for scalar targets.

**Exact change plan:** Before coercion, enforce exactly one occurrence for every non-repeatable path, query, scalar body, structured body, GraphQL variable, form, and literal body flag. Canonicalize target names and reject two aliases mapped to the same singleton target. Permit multiple occurrences only for explicitly repeatable headers and declared array flags.

**Required tests:**

- Happy: one value for each scalar target; multiple values for a repeatable header and `string_array` retain declared order.
- Bad: repeated path/body/form/GraphQL/raw-body flags and two aliases for the same target are rejected before preview or I/O.
- Edge: identical duplicates, boolean true/false, empty values, case-normalized header aliases, and repeated structured JSON.

### RT-B11 — Caller-controlled path and query values have no byte-bound contract

**Severity:** BLOCKER

**Evidence:** `OperationParameter` exposes `MaxBytes`, but its declaration comment and enforcement cover headers only (`internal/connectors/engine/bundle.go:785-800`); the command-surface projection has no byte-limit field (`internal/connectors/engine/connector.go:959-974`). General flag coercion has no string byte cap (`internal/connectors/commandrunner/runner.go:2023-2037`). Direct-write query validation checks controls/type/enum but never length at `internal/connectors/engine/direct_write.go:1993-2024`. Path encoding validates identifiers/relative paths without a limit at `internal/connectors/engine/direct_read.go:909-978`. This directly contradicts the phase's oversized path/query rejection criterion.

**Impact:** Arbitrarily large source-imported CLI values can enter URL construction and provider I/O despite the declaration claiming bounded typed mappings. Limits on body, response, and headers do not bound URL components.

**Root cause:** `max_bytes` was added narrowly for header values and response bodies; path/query parameter metadata was never propagated through the command surface or enforced after wire encoding.

**Exact change plan:** Require a source-owned `max_bytes` for caller-bound path/query parameters, or define a conservative engine ceiling plus optional tighter declared cap. Propagate it into command metadata and enforce bytes after the exact path/query wire encoding, before preparation/preview/runtime construction. Validate fixed query declarations against the same cap at bundle load.

**Required tests:**

- Happy: values at the exact byte cap reach preview and execution.
- Bad: cap+1 path/query/fixed-query values fail before auth/runtime/network construction.
- Edge: multibyte UTF-8 versus rune count, percent-encoding expansion, repeated separators, long numeric values, and config-fallback path parameters.

### RT-B12 — REST operation direct reads remain a generic query/body escape hatch

**Severity:** BLOCKER

**Evidence:** `internal/connectors/engine/direct_read.go:61-108` merges every caller `req.Query` key over the fixed declaration without checking `REST.Parameters`, and it resolves path templates without rejecting extra path keys. `internal/connectors/engine/direct_read.go:431-474` merges every caller body key and relies on the general schema compiler, whose open-object default can admit undeclared properties. The operation preflight interface contains only operation/method/path/response cap/output policy (`internal/connectors/connectors.go:599-605`), and command validation passes only those fields at `internal/connectors/commandrunner/runner.go:748-775`. By contrast, direct writes enumerate and preflight exact path/query/body bindings at `internal/connectors/commandrunner/runner.go:830-862` and `internal/connectors/engine/direct_write.go:1875-1941`.

**Impact:** A hand-authored or conflict-merged installed command—or an internal caller—can add arbitrary query keys and, under an open schema, arbitrary JSON body fields to a named read operation. The operation ID ceases to be a closed provider contract and becomes a partial generic HTTP request surface.

**Root cause:** Closed binding preflight/materialization was implemented for operation writes but not carried across the older operation-read path.

**Exact change plan:** Extend operation-read preflight with exact declared path/query/body bindings. Reject unknown, fixed-field collisions, duplicate aliases, cross-bound mappings, and unused path values. Require caller-overridable JSON body schemas to be recursively closed/bounded using the structured-body compiler. Preserve the literal raw body only for the existing exact `text/plain` root-string declaration.

**Required tests:**

- Happy: source-declared query/path/body fields and reserved pagination keys materialize exactly.
- Bad: unknown query, unknown/extra path key, unknown body property, fixed/caller collision, and cross-bound mapping fail before I/O.
- Edge: required query groups, config-backed path variables, pagination-controlled parameters, text/plain empty/present bodies, and a GraphQL query rejecting all path/query/header overrides.

### RT-B13 — Name-based JSON redaction deletes ordinary provider identifiers and counters

**Severity:** BLOCKER

**Evidence:** The `json_redacted` policies route all results through `redactJSONValue` (`internal/connectors/engine/direct_read.go:671-697`). That function deletes any non-null field selected by `shouldRedactJSONField` (`internal/connectors/engine/direct_read.go:744-755`), whose substring rule matches every key containing `token`, `secret`, `credential`, and related markers (`internal/connectors/engine/direct_read.go:805-822`). Real source schemas demonstrate the collision: Braintree declares ordinary primary key `token` at `internal/connectors/defs/braintree/schemas/payment_methods.json:5-8`, and Nebius declares the ordinary counter `trained_tokens` at `internal/connectors/defs/nebius-ai/schemas/fine_tuning_jobs.json:233-235`.

**Impact:** Provider IDs, occurrence tokens, pagination tokens, analytics counters, and other ordinary output silently disappear and are replaced with `<name>_redacted=true`. This loses provider truth and violates the explicit instruction to preserve IDs/occurrence IDs while masking only actual credentials/secrets.

**Root cause:** Field-name heuristics substitute for declaration-owned secret classification or concrete secret-value matching.

**Exact change plan:** Make redaction declaration/value aware. Preserve every ordinary provider field by default, regardless of name. Redact only explicitly classified response fields or concrete credential literals at the public boundary, and retain a presence marker where the schema requires one. Migrate broad legacy policies to explicit field lists rather than substring guessing.

**Required tests:**

- Happy: `token` primary keys, occurrence/pagination tokens, `trained_tokens`, and ordinary credential-looking labels survive exactly.
- Bad: an explicitly declared credential response field and a concrete configured secret literal are masked.
- Edge: null secret fields, nested arrays/maps, case/punctuation variants, a secret value under an ordinary key, and an ordinary value under a secret-looking key.

### RT-W14 — Path interpolation validates values before filters, not after them

**Severity:** WARNING

**Evidence:** `internal/connectors/engine/interpolate.go:154-200` rejects controls/dangerous Unicode only on the original resolved value. Filter output is returned without a second context check at `internal/connectors/engine/interpolate.go:212-226`. `last_path_segment`, `join:`, and `const:` can synthesize or expose raw output at `internal/connectors/engine/interpolate.go:492-518`; `lastPathSegment` returns the trailing text verbatim at `internal/connectors/engine/interpolate.go:562-581`. `InterpolatePath` only checks slash-delimited output segments that decode exactly to `..` (`internal/connectors/engine/interpolate.go:73-97`), so a filtered value such as a URI tail containing `?`, `#`, encoded separators, controls from a `const:`, or joined path separators is not guaranteed to remain one encoded path segment. Existing last-segment tests at `internal/connectors/engine/interpolate_test.go:304-359` cover benign values only.

**Impact:** A future/current declaration that uses these general filters in a path can turn record data into query/fragment syntax or additional/traversal-like segments. Current repository uses are primarily computed fields/query values, which limits immediate exploitability, but the public path API admits the unsafe combination.

**Root cause:** Supplying an explicit filter disables path's default `urlencode`, and validation is tied to the pre-filter value rather than the final context-specific string.

**Exact change plan:** Make filter admission context aware. For path interpolation, encode the final dynamic result as one segment (or reject filters that intentionally produce multi-segment output), then reject decoded separators/traversal, controls, bidi/invisible characters, query, and fragment syntax. Validate `const:` and `join:` arguments at bundle load and revalidate final output.

**Required tests:**

- Happy: benign `last_path_segment` and explicitly encoded IDs remain stable; computed-field/query uses keep existing semantics.
- Bad: `?`, `#`, `%2f..%2f`, CR/LF/C1/bidi from last-segment/const/join cannot alter a path.
- Edge: percent signs, empty segments, trailing slash, multi-character join separators, and chained filters.

### RT-W15 — A GitHub-specific secret transform is accepted but never executed

**Severity:** WARNING

**Evidence:** The neutral operation model advertises a provider transform at `internal/connectors/engine/bundle.go:874-897`, and bundle validation hard-codes `github_secret_encryption` as an accepted value at `internal/connectors/engine/bundle.go:2949-2983`. Repository-wide call tracing finds no runtime read of `SensitivePolicySpec.Transform`; only loader/schema/tests and GitHub declarations reference it. The scoped test `internal/connectors/engine/bundle_test.go:1689-1700` verifies only rejection of unknown transform names.

**Impact:** A declaration can validate with a claimed preprocessing contract that the executor silently ignores, sending untransformed material if the surrounding operation later becomes executable. It also embeds provider-specific behavior in the connector-neutral engine while providing no implementation/registration seam.

**Root cause:** Schema/validator enablement landed without a corresponding transform executor or fail-closed runtime gate.

**Exact change plan:** Replace the string switch with a connector-neutral transform registry/interface bound by declaration identity, include transform identity/output in the prepared digest, and reject any executable operation whose declared transform has no registered implementation. Until implemented, reject `github_secret_encryption` for executable commands rather than ignoring it.

**Required tests:**

- Happy: `none` is byte-stable; a registered transform changes the exact prepared body and digest deterministically.
- Bad: known-but-unimplemented and unknown transforms fail during load/preflight before preview/I/O.
- Edge: transform error, changed public key/input, secret zeroization, retry determinism, and no connector-name switch in dispatch.

## Explicit PASS evidence

The following areas were traced and did not produce an additional finding beyond the boundary defects above:

- **Prepared-plan digest and approval binding:** `internal/connectors/engine/prepared_write.go:97-189` binds method, full URL/query, target, media/body format, body bytes, ordered repeatable header values (via SHA-256), definition, hook identity, credential revision, configuration digest, batchability, action, and record count. `internal/connectors/engine/prepared_write.go:298-310` recomputes the preview digest and passes the exact target/digest into the approval gate. RT-B07 is the remaining execution-materialization break, not a digest omission.
- **Operation direct-write dispatch policy:** `internal/connectors/engine/direct_write.go:92-151` marks every REST/GraphQL mutation as a write transport, disables requester retries, and dispatches only the prepared closed format. Existing direct-write redirect/retry tests at `internal/connectors/engine/direct_write_test.go:581-648` and `internal/connectors/engine/direct_write_test.go:651-719` preserve the one-call terminal receipt behavior. RT-B03 is confined to the separate declarative reverse-write requester.
- **Structured REST body closure and bounds:** `internal/connectors/engine/structured_rest_body.go:682-804` compiles and materializes only recursively closed/bounded schemas; `internal/connectors/engine/structured_rest_body.go:965-1009` bounds bytes/nodes and rejects cycles; `internal/connectors/engine/structured_rest_body.go:1420-1536` requires closed objects, bounded arrays, declared required properties, and bounded depth/field count. Scoped structured-body tests exercise static/caller composition, overlap, unknown fields, malformed structures, and no-I/O failures.
- **Raw-body exclusion:** `internal/connectors/commandrunner/runner.go:1593-1606` admits exact raw `body` only for direct-read string flags, and `internal/connectors/commandrunner/runner.go:1632-1635` rejects mixing it with JSON mappings. `internal/connectors/engine/direct_read.go:431-474` further restricts it to declared `text/plain` POST with a root string schema and a request byte cap. RT-B12 concerns the still-open query/JSON-object bindings, not this raw-body branch.
- **Multipart and binary file confinement:** Operation multipart uses the same root-confined canonical builder at `internal/connectors/engine/direct_write.go:136-151` and digest-binds the canonical declaration/body at `internal/connectors/engine/direct_write.go:1495-1566`. Scoped multipart tests cover regular files, roots, aggregate/per-file caps, digest changes, terminal receipts, and redirect/retry refusal. Read-only tracing through `binary_read.go` confirms declared binary/text downloads use bounded `DoStream`, a destination root, temp/atomic file handling, response media validation, and no output file on error; RT-B06 is limited to printable error content.
- **Fixed GraphQL authority and binary/text response behavior:** `internal/connectors/engine/direct_write.go:1592-1699` rejects caller path/query/header overrides, fixes document/operation/path, validates closed variables, caps the payload, and digests exact encoded bytes. `internal/connectors/engine/graphql_operation.go:691-753` uses one JSON value, preserves partial-data metadata for reads, and bounds response parsing. `internal/connectors/engine/direct_write.go:200-223` correctly leaves explicitly non-JSON provider bytes as a receipt rather than fabricating an envelope.
- **Repeatable headers and exact receipt bytes:** `internal/connectors/engine/direct_write.go:2226-2241` and `internal/connectors/engine/write.go:516-602` retain status, ordered repeatable header values, raw text or base64 bytes, byte count, decoded view, record/operation identity, and path. Scoped tests cover duplicate JSON keys, invalid UTF-8, terminal non-2xx, and repeated receipt headers. RT-B01 is specifically the public secret projection; RT-B02/RT-B08 cover the remaining loss/presence cases.
- **Status checks:** Read-only tracing through the operation status-check call path confirms `DoStatusCheck` uses a bodyless HEAD, normal retry/status handling, a cap+1 capture, declared response headers, and final non-2xx metadata rather than a write result. It receives the operation's explicit redirect policy. No additional status-check finding was found.
- **Rate-limit parking:** `internal/connectors/engine/rate_limit_parking.go:23-95` parks/rearms only an `errors.As`-reachable typed rate-limit error with non-zero authoritative reset evidence and a known observation source, checks context first, clones the checkpoint, and does not mutate the coordinator for generic/no-reset failures.
- **Record-schema promotion:** `internal/connectors/engine/record_schema_promotion.go:24-165` expands top-level unions before measuring fields, rejects hollow closed records, validates exact case-sensitive named mappings, and requires all schema-required fields; scoped promotion tests cover nested unions, wrapper/arm fields, empty objects, required mapping, and duplicates.
- **Connector-neutral reachability:** `internal/connectors/engine/connector.go:222-392` exposes direct read/write, binary download, status check, structured-body materialization, metadata/preflight, and ordinary reverse write through interfaces without connector-name dispatch. RT-W15 is the one provider-specific declarative exception found.

## Verification run

The following diagnostics passed from the reviewed SHA after all findings were frozen:

```text
go test -timeout 20m ./internal/connectors/commandrunner ./internal/connectors/connsdk ./internal/connectors/engine
go vet ./internal/connectors/commandrunner ./internal/connectors/connsdk ./internal/connectors/engine
```

Passing tests/vet are compilation and regression evidence only; several scoped tests explicitly encode the unsafe output/presence behavior described above.

## Reviewed scope (30/30)

1. `internal/connectors/commandrunner/runner.go`
2. `internal/connectors/commandrunner/runner_test.go`
3. `internal/connectors/connectors.go`
4. `internal/connectors/connsdk/http.go`
5. `internal/connectors/connsdk/http_test.go`
6. `internal/connectors/connsdk/stream.go`
7. `internal/connectors/engine/bundle.go`
8. `internal/connectors/engine/bundle_test.go`
9. `internal/connectors/engine/connector.go`
10. `internal/connectors/engine/connector_test.go`
11. `internal/connectors/engine/direct_read.go`
12. `internal/connectors/engine/direct_write.go`
13. `internal/connectors/engine/direct_write_multipart_test.go`
14. `internal/connectors/engine/direct_write_test.go`
15. `internal/connectors/engine/graphql_operation.go`
16. `internal/connectors/engine/graphql_operation_test.go`
17. `internal/connectors/engine/interpolate.go`
18. `internal/connectors/engine/interpolate_test.go`
19. `internal/connectors/engine/prepared_write.go`
20. `internal/connectors/engine/rate_limit_parking.go`
21. `internal/connectors/engine/read.go`
22. `internal/connectors/engine/record_schema_promotion.go`
23. `internal/connectors/engine/record_schema_promotion_test.go`
24. `internal/connectors/engine/schema.go`
25. `internal/connectors/engine/schema_test.go`
26. `internal/connectors/engine/structured_rest_body.go`
27. `internal/connectors/engine/structured_rest_body_test.go`
28. `internal/connectors/engine/write.go`
29. `internal/connectors/engine/write_query_test.go`
30. `internal/connectors/engine/write_test.go`
