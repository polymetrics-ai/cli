# TDD LEDGER — issue #3761 multipart `rest_write`

| ID | Requirement | RED evidence | GREEN evidence | Refactor / verification |
| --- | --- | --- | --- | --- |
| R1 | Closed `rest.multipart` contract only loads for safe `rest_write` declarations | Pending: add focused bundle/schema test and run against current source | Pending | Pending |
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

