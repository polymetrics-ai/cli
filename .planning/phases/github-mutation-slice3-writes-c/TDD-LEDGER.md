# GitHub mutation slice 3 TDD ledger

| Slice | Red | Green | Refactor / outcome |
| --- | --- | --- | --- |
| Each mutation | Before execution, the agent-derived predicate rejects an absent object, mismatched tag/value, or unchanged provider state. | An independent GitHub read-back proves the produced value after the approved `pm` run. | Direct provider cleanup plus independent absence read-back completes containment; only then may the command be certified. |
| Evidence record | A captured lifecycle is not accepted as published evidence before schema validation. | `go run ./cmd/connectorgen certification-matrix --check` accepts the schema-v2 record. | Invalid candidate records are removed rather than weakened or retained. |

The first green slice assigned the disposable `pm-cert-slice3-1787024791` team as a security-manager team, rejected an absent/different slug as a plausible wrong result, removed the assignment through GitHub directly, proved the assignment absent, deleted the disposable team container, and received an independent `404` read-back.
