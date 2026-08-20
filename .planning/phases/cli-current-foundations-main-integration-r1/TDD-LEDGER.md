# TDD Ledger — Current Foundations Main Integration r1

Every code or conflict correction begins with a focused production-shaped failing test. The component merge itself does not replace this rule.

| ID | Requirement | Red evidence | Green evidence | Refactor evidence | Status |
| --- | --- | --- | --- | --- | --- |
| INT-01 | Structured body + typed header + exact query/path binding | `TestDirectWriteCommandBindsStructuredBodyHeadersQueryAndPath` failed before I/O: a declaration-fixed `payload.fixed` field was treated as missing when the caller supplied only `payload.name`. | Pending merge and conflict resolution of `55ddb650aa5594ddd156b0939cb1df6027a31d56` | Existing direct-write/scalar/form/SCIM regressions | red |
| INT-02 | Terminal status-only 4xx/5xx versus binary/text GET errors | `TestOperationStatusCheckPreservesTerminalNon2xxMetadataAndDeclaredHeaders` failed: ordinary requester returned `http 404` as an error before the final declared header/result could be retained. | `OperationStatusCheck` now uses the closed `DoStatusCheck` response path while retaining declaration-owned request/response headers; focused engine, connsdk, commandrunner, and CLI regressions pass. | Loader, retry, bounded download, no-output-file, and deterministic no-header CLI status-output regressions | green |
| INT-03 | Source-imported declaration reaches installed generated surface | Pending combined #4306/#4307 head | Pending | Source-lock hash/count, malformed/unknown/oversized source regressions | ready after qualified merge inputs |
| INT-04 | Multiple reverse-ETL actions through persisted App/installed CLI | Pending exact #4303 and #4305 heads | Pending | Existing typed destination and GitHub regressions | blocked on #4303/#4305 heads |
| INT-05 | Cross-bound and raw-authority rejection happens before I/O | Pending exact combined head | Pending | Existing no-generic-HTTP/action authority regressions | blocked on all inputs |

No `Green` entry may be added solely from a component PR report: the test must pass from the final composed SHA.
