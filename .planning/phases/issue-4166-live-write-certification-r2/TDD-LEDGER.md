# TDD Ledger — Issue 4166 Live Write Certification

| Slice | Red | Green | Refactor / evidence |
| --- | --- | --- | --- |
| Scope gate | The prior proof’s 607 `DryRunWrite` preparations were falsely treated as live action passes. | `SCOPE.md` proves the safe set is only 28/607, classifies the other 579 actions, and blocks harness construction pending an infrastructure decision. | No production code changed. A future invariant test must reject inventory drift or an unclassified action. |
| Report truthfulness | Pending decision | Pending decision | Prepared-only and non-live outcomes must be distinct from `pass`; batch aggregation must expose the boundary. |
| Full parity | Pending decision | Pending decision | The exact external-child CLI path must prove it enables full/write stages and refuses skipped applicable work. |
| Live scenarios | Pending decision | Pending decision | Every action needs a production-entry-point mutation, independent read-back, ownership-checked cleanup, and resume assertion. |
| Fault proof | Pending decision | Pending decision | A post-schema deliberately broken request action must fail the real certification run by exact action name. |
