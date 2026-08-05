# TDD LEDGER — issue #3761 multipart `rest_write`

| ID | Requirement | RED evidence | GREEN evidence | Refactor / verification |
| --- | --- | --- | --- | --- |
| R1 | Closed `rest.multipart` contract only loads for safe `rest_write` declarations | `go test ./internal/connectors/engine -run 'TestBundleLoadAcceptsTypedMultipartRestWriteContract|TestBundleLoadRejectsUnsafeMultipartRestWriteContracts' -count=1` failed before production changes: `operations.json: /operations/0/rest/multipart: additional property not allowed`; the valid fixture could not load and unsafe cases could not reach their intended semantic guards. | The same focused command passed after the schema/type/semantic contract; `go vet ./internal/connectors/engine` also passed. The retained metadata test remains red until the dispatcher recognizes the new format. | Pending |
| R2 | Preview binds fields, source identity, and every approved file SHA-256 without network I/O | Pending | Pending | Pending |
| R3 | Direct multipart dispatch is root-confined, bounded, media-checked, response-bounded, and exactly once | Pending | Pending | Pending |
| R4 | Docs distinguish shared capability from connector adoption and legacy `file_upload` | Pending | Pending | Pending |
| R5 | Implemented command claim reaches real preflight and plan → preview → approval → execute loopback path | Pending | Pending | Pending |

## Evidence discipline

Each RED entry records the exact focused command and its observed failure before
production code changes. The final GREEN entries record the same command after
the smallest implementation and then the broader changed-package gates. Tests
use `httptest` and temporary files only; no provider request or credential is
permitted.
