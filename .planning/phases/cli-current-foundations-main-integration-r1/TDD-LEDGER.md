# TDD Ledger — Current Foundations Main Integration r1

Every code or conflict correction begins with a focused production-shaped failing test. The component merge itself does not replace this rule.

| ID | Requirement | Red evidence | Green evidence | Refactor evidence | Status |
| --- | --- | --- | --- | --- | --- |
| INT-01 | Structured body + typed header + exact query/path binding | Pending exact #4305 and #4307 heads | Pending | Existing direct-write/scalar/form/SCIM regressions | blocked on #4305 head |
| INT-02 | Terminal status-only 4xx/5xx versus binary/text GET errors | Pending qualified #4308 head | Pending | Loader, retry, bounded download, and no-output-file regressions | blocked on #4308 terminal qualification |
| INT-03 | Source-imported declaration reaches installed generated surface | Pending combined #4306/#4307 head | Pending | Source-lock hash/count, malformed/unknown/oversized source regressions | ready after qualified merge inputs |
| INT-04 | Multiple reverse-ETL actions through persisted App/installed CLI | Pending exact #4303 and #4305 heads | Pending | Existing typed destination and GitHub regressions | blocked on #4303/#4305 heads |
| INT-05 | Cross-bound and raw-authority rejection happens before I/O | Pending exact combined head | Pending | Existing no-generic-HTTP/action authority regressions | blocked on all inputs |

No `Green` entry may be added solely from a component PR report: the test must pass from the final composed SHA.
