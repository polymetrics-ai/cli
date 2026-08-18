## TDD Ledger

This is a live-certification evidence task, not a production behavior change. The required red/green cycle applies to every retained certification assertion.

| Slice | Red | Green | Refactor / outcome |
| --- | --- | --- | --- |
| Each mutation | Before execution, a read-back predicate rejects a plausible wrong result (missing object, mismatched tag/value, or unchanged state). | The provider read-back proves the planned produced value. | Direct provider DELETE plus independent absence read-back completes cleanup; only then is the result eligible for a certified record. |
| Evidence record | Candidate record before validation is untrusted. | `certification-matrix --check` accepts the record. | Delete any rejected candidate rather than retaining invalid evidence. |

The current binary’s connector-command lifecycle was confirmed to require `--plan`, `--preview`, and the bare stdin approval-token marker; `--approve` is explicitly rejected.
