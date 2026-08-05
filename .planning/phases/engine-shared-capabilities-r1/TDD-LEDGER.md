# TDD Ledger — engine-shared-capabilities-r1

GSD programming loop, manual-GSD fallback (Claude Code session, no Pi runtime subagents). Every
behaviour-adding task starts with a failing test. Red evidence is recorded before the implementing
commit; green evidence after.

Status legend: `planned` → `RED` (failing test committed/observed) → `GREEN` (implementation passes).

## Task 1 — `connsdk` bounded streaming + redirect origin policy

| # | Test | Asserts | Status |
| --- | --- | --- | --- |
| 1.1 | `TestDoStreamReturnsOpenBody` | body is streamed, not buffered; caller closes | GREEN |
| 1.2 | `TestDoStreamRejectsCrossOriginRedirect` | fail-closed on a redirect to another host | GREEN |
| 1.3 | `TestDoStreamStripsCustomAuthHeaderCrossOrigin` | a custom auth header (the 71-connector case Go does **not** strip), a default header, and Authorization are all absent on a permitted cross-origin hop | GREEN |
| 1.4 | `TestDoStreamKeepsAuthSameOrigin` | same-origin redirect still carries credentials | GREEN |
| 1.5 | `TestDoStreamAllowedHostsPermitsNamedHost` | per-operation allowlist admits exactly that host, still credential-free; others stay refused | GREEN |
| 1.6 | `TestDoStreamRetryDiscardsPartialBody` | a retried attempt never concatenates partial bytes | GREEN |
| 1.7 | `TestDoStreamDoesNotMutateSharedClient` | `CheckRedirect` is set on a clone, not `r.Client` | GREEN |
| 1.8 | `TestDoStreamHTTPErrorClosesBody` | a terminal 4xx returns `*HTTPError` and no dangling reader | GREEN |
| 1.9 | `TestDoStreamRejectsRedirectLoop` | hop cap re-stated (installing CheckRedirect replaces Go's default 10) | GREEN |

**Red evidence (task 1):** `r.DoStream undefined (type *Requester has no field or method DoStream)`
and `undefined: StreamOptions` across stream_test.go. **Green evidence:** all 9 pass.

**Design note:** the first implementation smuggled the credential-header list to the redirect policy
through an internal `X-Polymetrics-Internal-Credential-Headers` request header. That worked but put
an implementation artifact on the wire. Replaced with a closure variable captured by the per-call
client clone — nothing extra is ever sent.

## Task 2 — engine bounded binary download executor

| # | Test | Asserts | Status |
| --- | --- | --- | --- |
| 2.1 | `TestBinaryDownloadWritesBoundedFile` | file on disk, correct size and SHA-256, contained in root | GREEN |
| 2.2 | `TestBinaryDownloadRejectsOverflow` | one byte past the limit is rejected, and leaves no file behind | GREEN |
| 2.3 | `TestBinaryDownloadExactLimitSucceeds` | read-one-past introduces no off-by-one rejection | GREEN |
| 2.4 | `TestBinaryDownloadClampsMaxBytes` | request clamps below spec; request cannot raise spec | GREEN |
| 2.5 | `TestBinaryDownloadRejectsExtractArchives` | hard execution error naming `extract_archives` | GREEN |
| 2.6 | `TestBinaryDownloadRejectsWrongKind` | a `rest_read` operation cannot be run as a download | GREEN |
| 2.7 | `TestBinaryDownloadRejectsNonGET` | non-GET refused | GREEN |
| 2.8 | `TestBinaryDownloadRejectsAbsoluteEndpoint` | connector-relative invariant preserved | GREEN |
| 2.9 | `TestBinaryDownloadRequiresDeclaredEndpoint` | endpoint absent from `api_surface` refused | GREEN |
| 2.10 | `TestBinaryDownloadFilenameSanitized` | 5 hostile `Content-Disposition` values contained | GREEN |
| 2.11 | `TestBinaryDownloadRFC5987Filename` | `filename*` decoded from the **unstarred** key | GREEN |
| 2.12 | `TestBinaryDownloadCallerFileNameContained` | traversing caller names refused | GREEN |
| 2.13 | `TestBinaryDownloadCallerFileNameNotRewritten` | refused, **not** silently basename-ed | GREEN |
| 2.14 | `TestBinaryDownloadRejectsSymlinkEscape` | never writes *through* an escaping symlink (`os.Root`) | GREEN |
| 2.15 | `TestBinaryDownloadHonoursAllowOverwrite` | refuses by default without clobbering; replaces when declared | GREEN |
| 2.16 | `TestBinaryDownloadFilePermissions` | file mode `0o600` | GREEN |
| 2.17 | `TestBinaryDownloadSniffsContentType` | mismatch recorded on both fields, download NOT rejected | GREEN |
| 2.18 | `TestBinaryDownloadRecordIsFlatAndSurvivesRedaction` | flat scalars only; no field trips `shouldRedactJSONField`; all 10 fields present; no `download_url` | GREEN |
| 2.19 | `TestBinaryDownloadRequiresDestRoot` | no implicit destination | GREEN |
| 2.20 | `TestBinaryDownloadHTTPErrorLeavesNoFile` | a 403 leaves the destination empty | GREEN |
| 2.21 | `TestOperationsSchemaAcceptsBinaryPolicyFields` | the real meta-schema is COMPILED and RUN against a document using the new fields, and still rejects an undeclared one | GREEN |
| 2.22 | `TestBinaryDownloadRefusesCrossHostRedirectByDefault` | executed end-to-end: an undeclared cross-host redirect is refused and the CDN sees no credential | GREEN |
| 2.23 | `TestBinaryDownloadAllowCrossHostIsEnforced` | declaring `allow_cross_host` actually changes behaviour; credential still does not travel | GREEN |
| 2.24 | `TestBinaryDownloadAllowedHostsIsEnforced` | `allowed_hosts` admits exactly the named host, refuses others, sends no credential | GREEN |
| 2.25 | `TestBinaryDownloadStallTimeoutIsEnforced` | a stalled transfer is actually aborted and leaves no file | GREEN |

**Third real defect, caught by insisting on execution over schema-reading.** The first version of
test 2.21 only did `strings.Contains(operationsSchemaJSON, field)` — it would have passed on a field
name appearing anywhere in the file, with no enforcement whatsoever. Replaced with a real
compile-and-validate, plus negative coverage that the block stays `additionalProperties:false`.

Rewriting it that way then exposed a genuine bug in `stall_timeout_seconds`: `newStallReader` derived
its OWN context and cancelled that, so the watchdog fired correctly and aborted nothing — the read
hung until connsdk's 60s client timeout. The declared field was inert. Fixed by passing the cancel
func of the context the HTTP request was actually built with; the test now completes in ~1s instead
of 60s.

**Red evidence (task 2):** `undefined: OperationBinaryDownload`, `undefined:
BinaryDownloadRequest`.

**Two real defects the tests caught, worth recording:**

1. The first sanitizer applied `filepath.Base` to a CALLER-supplied name, silently rewriting
   `../escape.txt` to `escape.txt`. That hides a traversal attempt instead of reporting it. Split
   into a strict `isLocalSingleSegment` for caller input (refuses, never rewrites) and the lenient
   sanitizer for untrusted provider text, where falling back to a safe name beats failing the
   download.
2. The first symlink test only exercised name validation, not containment. Rewritten to plant an
   escaping symlink under a perfectly valid single-segment name and assert the file outside the root
   is never written through.

## Task 3 — query parameters on write actions

| # | Test | Asserts | Status |
| --- | --- | --- | --- |
| 3.1 | `TestWriteActionQueryPlainString` | plain-string entry resolves; unresolved key is a hard error | GREEN |
| 3.2 | `TestWriteActionQueryOmitWhenAbsent` | object form drops the param instead of erroring | GREEN |
| 3.3 | `TestWriteActionQueryDefault` | object form sends the literal default | GREEN |
| 3.4 | `TestWriteActionQueryAllBodyTypes` | query reaches the wire for json/form/none/json_array/graphql | GREEN |
| 3.5 | `TestWriteActionNoQueryUnchanged` | **regression guard**: no `query` declared → no query string, byte-identical to today | GREEN |
| 3.6 | `TestWriteActionQueryMultipartBodyType` | the sixth branch (needs a real file, so separate from the table) | GREEN |
| 3.7 | `TestWriteActionQueryFromRecordField` | query templates see the same `Vars` the path does | GREEN |
| 3.8 | `TestWriteActionQueryParsesBothDialects` | both dialects parse via the existing `QueryParam.UnmarshalJSON`, no second parser | GREEN |
| 3.9 | `TestWritesSchemaAcceptsQuery` | `writes.schema.json` accepts the optional `query` object | GREEN |
| 3.10 | `TestBundleLoadWiresWriteQueryAndDynamicFields` | both write capabilities survive the REAL loader — meta-schema validation, decode, semantic validation — not just a Go-constructed action | GREEN |
| 3.11 | `TestBundleLoadRejectsInvalidDynamicFields` | the loader ENFORCES the declaration contract (`dynamic_fields` on an unsupported `body_type` is rejected at load) | GREEN |

**Red evidence (task 3):** before implementation the package did not compile —
`unknown field Query in struct literal of type WriteAction` at write_query_test.go lines
46, 74, 95, 116, 149, 205, 208. **Green evidence:** `go test ./internal/connectors/engine/ -run
'TestWriteAction|TestWritesSchema'` → `ok`. Full-package regression `go test
./internal/connectors/engine/ ./internal/connectors/connsdk/` → `ok`, `ok`.

## Task 4 — typed dynamic-key write bodies

| # | Test | Asserts | Status |
| --- | --- | --- | --- |
| 4.1 | `TestDynamicFieldsAcceptsTypedScalars` | declared region accepts scalar values under matching keys; inline merge | GREEN |
| 4.2 | `TestDynamicFieldsRejectsNestedValue` | object/array value is a hard error — the anti-escape-hatch invariant | GREEN |
| 4.3 | `TestDynamicFieldsRejectsUnmatchedKey` | keys failing `key_pattern` refused (leading digit, dash, space, empty, overlong) | GREEN |
| 4.4 | `TestDynamicFieldsRejectsCollision` | a key shadowing a `path_field` or an existing body key is refused | GREEN |
| 4.5 | `TestDynamicFieldsEnforcesBounds` | `max_keys` and `max_value_bytes` enforced | GREEN |
| 4.6 | `TestDynamicFieldsAbsentUnchanged` | **regression guard**: no `dynamic_fields` → closed schema still rejects undeclared fields | GREEN |
| 4.7 | `TestDynamicFieldsRedactionApplies` | `redact_fields` still redacts dynamic values out of error text | GREEN |
| 4.8 | `TestDynamicFieldsNestedTarget` | `target: nested` keeps the region under its container | GREEN |
| 4.9 | `TestDynamicFieldsRejectsDisallowedValueType` | `value_types` is an allow-list, not a suggestion | GREEN |
| 4.10 | `TestDynamicFieldsBundleValidation` | declaration-time rejection of 8 malformed specs; valid and absent both accepted | GREEN |
| 4.11 | `TestWritesSchemaAcceptsDynamicFields` | `writes.schema.json` accepts the block | GREEN |

**Red evidence (task 4):** package did not compile — `undefined: DynamicFieldsSpec` and `unknown
field DynamicFields in struct literal of type WriteAction` at write_dynamic_fields_test.go lines 17,
23, 36, 37, 235-239. **Green evidence:** `go test ./internal/connectors/engine/ -run
'TestDynamicFields|TestWritesSchema|TestWriteAction'` → `ok`.

**Caught by test 4.11, worth recording:** the first schema draft used JSON Schema `minimum`, which
this repo's `CompileSchema` does not implement (`unknown keyword "minimum"`). Because the writes
meta-schema is compiled at bundle load, shipping it would have broken loading for **every**
connector, not just ones using `dynamic_fields`. Replaced with plain `integer`; the non-negative
bound is enforced in `validateDynamicFields` instead.

## Regression guards (must stay green, unmodified)

- `internal/connectors/engine/direct_read_test.go` — the bounded-JSON path is untouched.
- `internal/connectors/engine/write_test.go` — current write behaviour is untouched.
- `internal/connectors/engine/bundle_test.go` — existing bundle validation is untouched.
- `internal/connectors/certify/**` — the binary gate keeps passing without modification.
