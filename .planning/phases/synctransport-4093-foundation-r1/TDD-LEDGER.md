# TDD Ledger — Refs #4093

| Slice | Red: test and observed failure | Green: implementation and passing assertion | Refactor / status |
| --- | --- | --- | --- |
| Neutral source | Pending: synthetic second connector with its own evidence cannot compose through the App-owned fixed evidence factory. | Pending | Planned first. |
| Typed destination | Pending: synthetic typed action destination cannot compose without the GitHub-only contract/App. | Pending | Planned second. |
| Fail closed | Pending: malformed/unknown role/executor/evidence must fail before all I/O. | Pending | Planned with each adapter. |
| Mode validation | Pending: `change_capture` can be represented as a destination mode. | Pending | Planned with descriptor validation. |
| Commit reconciliation | Pending: a commit returning an unknown outcome can replay a destination effect. | Pending | Planned before retiring legacy composition. |
| Stage bound / cleanup | Pending: owner stages have neither a safe quota nor bounded reclamation. | Pending | Planned before retiring legacy composition. |

Every Green entry must cite the exact focused `go test` command and observable counter/state assertion. No red result will be recorded as a pass.
