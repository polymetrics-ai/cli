# UAT — Issue 3981: managed-target ownership

Manual UAT is not required: every deliverable is a connector-neutral internal
state transition with deterministic fake-driver or persisted-state proof.

| Deliverable | Evidence | Result |
| --- | --- | --- |
| Owned namespace can create a second stream relation | `TestManagedTargetProvisioningTruthTable/owned_namespace_allows_second_stream_relation` red then green | pass |
| Namespace and relation ownership are independently fail-closed | table-driven managed-target truth-table tests | pass |
| Rename-stable immutable identity | managed-target rename assertion and `TestStreamIDIsPersistedAndSurvivesStreamRename` | pass |
| Concurrent/cancel-safe provisioning | managed-target concurrent/cancellation tests under `-race` | pass |

No browser, external service, credential, SQL, or human-judgment check applies.
