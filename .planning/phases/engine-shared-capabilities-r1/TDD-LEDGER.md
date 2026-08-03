# TDD Ledger — engine-shared-capabilities-r1

GSD programming loop, manual-GSD fallback (Claude Code session, no Pi runtime subagents). Every
behaviour-adding task starts with a failing test. Red evidence is recorded before the implementing
commit; green evidence after.

Status legend: `planned` → `RED` (failing test committed/observed) → `GREEN` (implementation passes).

## Task 1 — `connsdk` bounded streaming + redirect origin policy

| # | Test | Asserts | Status |
| --- | --- | --- | --- |
| 1.1 | `TestDoStreamReturnsOpenBody` | body is streamed, not buffered; caller closes | planned |
| 1.2 | `TestDoStreamRejectsCrossOriginRedirect` | fail-closed on a redirect to another host | planned |
| 1.3 | `TestDoStreamStripsCustomAuthHeaderCrossOrigin` | a custom auth header (the 71-connector case Go does **not** strip) is absent on the cross-origin hop when it is explicitly allowed | planned |
| 1.4 | `TestDoStreamKeepsAuthSameOrigin` | same-origin redirect still carries credentials | planned |
| 1.5 | `TestDoStreamRejectsSchemeDowngrade` | https→http on the same host is refused | planned |
| 1.6 | `TestDoStreamRetryDiscardsPartialBody` | a retried attempt never concatenates partial bytes | planned |
| 1.7 | `TestDoStreamDoesNotMutateSharedClient` | `CheckRedirect` is set on a clone, not `r.Client` | planned |

## Task 2 — engine bounded binary download executor

| # | Test | Asserts | Status |
| --- | --- | --- | --- |
| 2.1 | `TestBinaryDownloadWritesBoundedFile` | happy path: file on disk, correct size, correct SHA-256 | planned |
| 2.2 | `TestBinaryDownloadRejectsOverflow` | a body one byte past the limit is rejected, not silently truncated | planned |
| 2.3 | `TestBinaryDownloadClampsMaxBytes` | request → spec → ceiling clamping order | planned |
| 2.4 | `TestBinaryDownloadRejectsExtractArchives` | `extract_archives: true` is a hard execution error | planned |
| 2.5 | `TestBinaryDownloadRejectsNonBinaryKind` | a `rest_read` operation cannot be run as a download | planned |
| 2.6 | `TestBinaryDownloadRejectsNonGET` | non-GET refused | planned |
| 2.7 | `TestBinaryDownloadContainsPathTraversal` | `../` and absolute filenames cannot escape the root | planned |
| 2.8 | `TestBinaryDownloadRejectsSymlinkEscape` | a symlink inside the root pointing out of it is refused (`os.Root`) | planned |
| 2.9 | `TestBinaryDownloadHonoursAllowOverwrite` | `O_CREATE\|O_EXCL` unless `allow_overwrite` | planned |
| 2.10 | `TestBinaryDownloadRecordIsFlatAndUnredacted` | flat scalars only; uses `source_ref`, and the record survives `shouldRedactJSONField` unredacted | planned |
| 2.11 | `TestBinaryDownloadSniffsContentType` | sniffed type recorded alongside the provider's claim; mismatch surfaced, not rejected | planned |
| 2.12 | `TestBinaryDownloadFilePermissions` | files `0o600`, dirs `0o700` | planned |
| 2.13 | `TestBinaryDownloadProviderFilenameSanitized` | `Content-Disposition` with `..\..\etc\passwd` and with an RFC 5987 `filename*` is sanitized | planned |
| 2.14 | `TestBinaryDownloadRequiresDeclaredEndpoint` | an endpoint absent from `api_surface` is refused | planned |

## Task 3 — query parameters on write actions

| # | Test | Asserts | Status |
| --- | --- | --- | --- |
| 3.1 | `TestWriteActionQueryPlainString` | plain-string entry resolves; unresolved key is a hard error | planned |
| 3.2 | `TestWriteActionQueryOmitWhenAbsent` | object form drops the param instead of erroring | planned |
| 3.3 | `TestWriteActionQueryDefault` | object form sends the literal default | planned |
| 3.4 | `TestWriteActionQueryAllBodyTypes` | query reaches the wire for all six body types | planned |
| 3.5 | `TestWriteActionNoQueryUnchanged` | **regression guard**: no `query` declared → no query string, byte-identical to today | planned |

## Task 4 — typed dynamic-key write bodies

| # | Test | Asserts | Status |
| --- | --- | --- | --- |
| 4.1 | `TestDynamicFieldsAcceptsTypedScalars` | declared region accepts scalar values under matching keys | planned |
| 4.2 | `TestDynamicFieldsRejectsNestedValue` | object/array value is a hard error — the anti-escape-hatch invariant | planned |
| 4.3 | `TestDynamicFieldsRejectsUnmatchedKey` | key failing `key_pattern` is refused | planned |
| 4.4 | `TestDynamicFieldsRejectsCollision` | a key shadowing `path_fields`/`body_fields` is refused | planned |
| 4.5 | `TestDynamicFieldsEnforcesBounds` | `max_keys` and `max_value_bytes` enforced | planned |
| 4.6 | `TestDynamicFieldsAbsentUnchanged` | **regression guard**: no `dynamic_fields` → closed schema behaviour unchanged | planned |
| 4.7 | `TestDynamicFieldsRedactionApplies` | `redact_fields` still redacts dynamic values | planned |

## Regression guards (must stay green, unmodified)

- `internal/connectors/engine/direct_read_test.go` — the bounded-JSON path is untouched.
- `internal/connectors/engine/write_test.go` — current write behaviour is untouched.
- `internal/connectors/engine/bundle_test.go` — existing bundle validation is untouched.
- `internal/connectors/certify/**` — the binary gate keeps passing without modification.
