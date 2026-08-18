# UAT — #3859 native database apply strategies

Manual `verify-work` fallback: this is a backend-only native target phase, so
acceptance is wholly observable through the required Docker PostgreSQL test.
No browser or subjective UI judgment applies.

| Deliverable | Automated evidence | Result |
| --- | --- | --- |
| Registered native target accepts only a resolved bounded page | `internal/connectors/engine/polling_apply_test.go` and live oversized-page refusal with a PostgreSQL re-read | passed |
| Six target strategies mutate durable state correctly | `TestPostgresManagedTargetWorksetDeliveryLive` under `databaseintegration` | passed |
| Source acknowledgement is post-commit/ledger only | `DatabasePollingApplyExecutor` uses `DatabaseWriteExecutor` receipt/ledger acknowledgement; focused and live tests passed | passed |
| Public generic-write/CLI surface remains unchanged | diff and connector surface-sync check | passed |
