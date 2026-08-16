# #4171 scale-proof TDD ledger

| Slice | Red | Green | Test contract coverage |
| --- | --- | --- | --- |
| Test-only bounded scale configuration | Complete: focused tests first failed to compile because the configuration parser/sentinel did not exist. | Complete: accepts the default `unlimited` path and positive page counts only. | Happy: default/full and one-page configuration. Bad: zero or non-numeric page count returns the named refusal before harness/binary I/O. Edge: one page and 900 pages map to 100 and 90,000 records. |
| Shipped binary route | Complete: built binary route guards production reachability rather than a hand-built component. | Complete at 100, 80,000, and 90,000 records: fresh binary independently verifies Parquet and PostgreSQL state. | Happy: exact 90,000 durable counts match. Bad: existing ineligible-stream/unapproved-plan tests refuse before I/O. Edge: one-page minimum, 800-page counterfactual, 900-page proof. |
| Durable PostgreSQL authorization | Complete: `TestPostgresManagedTargetAuthorizationLifetimeDefaultsAndRejectsOutOfRangeBeforePlanning`, `TestPostgresManagedTargetDurableAuthorizationValidationDoesNotReusePreviewSealLifetime`, and the built-binary no-token/replay checks failed or could not compile before the design existed. | Complete: token consumption atomically persists one 24h–48h `AuthorizationRecord`; managed binding scope is revalidated before each unit. | Happy: one-token run plus fresh binary same-shape continuation. Bad: malformed lifetime is a typed CLI validation refusal, replay is typed and side-effect-free. Edge: 24h/48h bounds, one-page continuation, pre-second-unit revocation. |
| Per-unit deadlines | Complete: a hanging real declarative GitHub page and a hanging second destination unit failed only under the new deadline implementation. | Complete: page request and apply/read-back units are deadline-bounded; the durable authorization itself is not cancelled. | Happy: normal one-page and 90k runs. Bad: HTTP/apply deadline returns `context.DeadlineExceeded`. Edge: first unit committed, second times out; source context survives; revocation is rejected before the second stage/apply. |
| Durable phase measurement | Complete: `TestRunETLTransportPersistsPartialPhaseMeasurementBeforeFailureCleanup` initially failed to compile without terminal measurement state. | Complete: partial results flow through all terminal failure writers; binary test reopens and logs state before cleanup. | Happy: 90k terminal record has exact non-zero counts/timings. Bad: injected second apply failure persists 2/2/1 before cleanup. Edge: no pre-stage execution persists explicit zero measurement. |

Manual-GSD fallback: the task started as a live-scale proof, then the captain's
in-flight two-clock safety decision required code changes. The red cases above
were written before their corresponding production changes; green evidence is
recorded in `VERIFICATION.md`. The six affected help transcripts are intentional
assertion updates for the documented new public flag, generated only after the
new help output was independently reviewed; no existing behavior was relaxed.
