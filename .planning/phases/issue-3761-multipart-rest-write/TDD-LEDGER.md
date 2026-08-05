# TDD LEDGER — issue #3761 multipart `rest_write`

| ID | Requirement | RED evidence | GREEN evidence | Refactor / verification |
| --- | --- | --- | --- | --- |
| R1 | Closed `rest.multipart` contract only loads for safe `rest_write` declarations | `go test ./internal/connectors/engine -run 'TestBundleLoadAcceptsTypedMultipartRestWriteContract|TestBundleLoadRejectsUnsafeMultipartRestWriteContracts' -count=1` failed before production changes: `operations.json: /operations/0/rest/multipart: additional property not allowed`; the valid fixture could not load and unsafe cases could not reach their intended semantic guards. | The same focused command passed after the schema/type/semantic contract; `go vet ./internal/connectors/engine` also passed. The retained metadata test remains red until the dispatcher recognizes the new format. | Pending |
| R2 | Preview binds fields, source identity, and every approved file SHA-256 without network I/O | `go test ./internal/connectors/engine -run 'TestOperationDirectWriteMetadataRecognizesTypedMultipartRestWrite|TestOperationDirectWriteMultipart' -count=1` failed before direct-write production edits with `rest_write content_type "multipart/form-data" is not supported by the typed executor`; previews could not represent the approved payload. | Focused engine tests now prove zero preview calls, a required digest per file, different previews for field/path changes, a stale preview refusal, and the full canonical multipart declaration in the prepared definition. `go test ./internal/connectors/engine -count=1` passed. | Existing writes.json canonical multipart tests still pass in the full engine package. |
| R3 | Direct multipart dispatch is root-confined, bounded, media-checked, response-bounded, and exactly once | The same focused engine command failed before dispatch changes; the connsdk bounded-response test was added first and intentionally did not compile because `Requester.DoMultipartLimited` did not yet exist. | Focused engine/connsdk tests now prove one approved loopback request, no pre-approval call, changed/missing/root-escaping/oversize/disallowed-media refusal before network, aggregate cap, 429/500 single attempts, redirect refusal, and `rest.max_bytes + 1` bounded response capture. `go test ./internal/connectors/connsdk -count=1` and `go vet ./internal/connectors/engine ./internal/connectors/connsdk` passed. | Existing requester root/snapshot/symlink/media tests remain green in the full connsdk package. |
| R4 | Docs distinguish shared capability from connector adoption and legacy `file_upload` | Pending | Pending | Pending |
| R5 | Implemented command claim reaches real preflight and plan → preview → approval → execute loopback path | Pending | Pending | Pending |

## Evidence discipline

Each RED entry records the exact focused command and its observed failure before
production code changes. The final GREEN entries record the same command after
the smallest implementation and then the broader changed-package gates. Tests
use `httptest` and temporary files only; no provider request or credential is
permitted.
