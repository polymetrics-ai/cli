# GitHub mutation slice 3 TDD ledger

| Slice | Red | Green | Refactor / outcome |
| --- | --- | --- | --- |
| Each mutation | Before execution, the agent-derived predicate rejects an absent object, mismatched tag/value, or unchanged provider state. | An independent GitHub read-back proves the produced value after the approved `pm` run. | Direct provider cleanup, or the provider-native terminal disposal where REST has no DELETE, plus an independent read-back completes containment; only then may the command be certified. |
| Empty parent collection | A missing child identifier alone cannot justify `no_object`. | The parent collection is listed; an existing identifier is used, or a fresh `pm-cert-` fixture is created, mutated, and disposed. | `no_object` remains only where GitHub exposes no way to create the required identity/object inside the captain's boundary. |
| Evidence record | A captured lifecycle is not accepted as published evidence before schema validation. | `go run ./cmd/connectorgen certification-matrix --check` accepts the schema-v2 record. | Invalid candidate records are removed rather than weakened or retained. |

The first green slice assigned the disposable `pm-cert-slice3-1787024791` team as a security-manager team, rejected an absent/different slug as a plausible wrong result, removed the assignment through GitHub directly, proved the assignment absent, deleted the disposable team container, and received an independent `404` read-back.
