## TDD Ledger

This is a live-certification evidence task, not a production behavior change. The required red/green cycle applies to every retained certification assertion.

| Slice | Red | Green | Refactor / outcome |
| --- | --- | --- | --- |
| Each mutation | Before execution, a read-back predicate rejects a plausible wrong result (missing object, mismatched tag/value, or unchanged state). | The provider read-back proves the planned produced value. | Direct provider DELETE plus independent absence read-back completes cleanup; only then is the result eligible for a certified record. |
| Evidence record | Candidate record before validation is untrusted. | `certification-matrix --check` accepts the record. | Delete any rejected candidate rather than retaining invalid evidence. |
| Large numeric path IDs | A raw exact-integer provider control succeeds while pm emits scientific notation and fails. | The provider effect is independently read back, then directly deleted and proven absent. | Record `class=integer_id_scientific_notation` once and tag later instances without re-investigating the class. |
| Organization variable create | A predicate rejects a missing variable or a mismatched unique name/value. | Independent GET matched both name and value after pm execution. | Direct DELETE returned 204 and independent GET returned 404; schema-v2 evidence validates. |
| Runner-group create | A predicate rejects a missing group or mismatched name, visibility, or public-repository policy. | Independent collection GET matched the unique selected/private fixture values. | Direct DELETE returned 204 and collection read-back proved absence; schema-v2 evidence validates. |

The current binary’s connector-command lifecycle was confirmed to require `--plan`, `--preview`, and the bare stdin approval-token marker; `--approve` is explicitly rejected.
