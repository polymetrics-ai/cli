# TDD Ledger — Issue 4316

| Slice | State | Red evidence | Green evidence | Refactor / notes |
| --- | --- | --- | --- | --- |
| Declaration-owned resolver | planned | Add a real engine-path URL capture test before production edits; it must fail because operation-level route data is unavailable. | Targeted engine test passes only after resolution is hooked into request construction. | Keep request construction central and reject absent route data before I/O. |
| Missing/conflicting route diagnostics | planned | Add zero-I/O checks for absent and conflicting declaration bases. | Tests assert the existing source-traced missing-foundation/blocked-command shape. | No fallback or arbitrary URL source permitted. |
| Six execution surfaces | planned | Add cross-surface test coverage to prove one declaration resolves on direct read/write, binary download/upload, ETL, and reverse ETL. | Captured server requests prove every surface shares the resolver. | Preserve reverse ETL plan/preview/approval/execute gate. |
| Help Scout v3 acceptance | planned | Add five real Help Scout direct-read cases that fail to resolve before definitions/hook changes. | Cases assert exact v3 request URL, path values, and page behavior. | Regenerate only the canonical derived outputs. |

Red and green command output will be appended as each slice runs. No production `cmd/` or `internal/` change may precede its corresponding red evidence.
