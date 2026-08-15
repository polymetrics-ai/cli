# DISCUSSION LOG — issue #3865 verified-auth cohort fencing

The generated `scripts/gsd prompt discuss-phase issue-3865-cohort-fencing-r1` prompt was executed inline. The issue and parent contracts resolve the design choices below; no interactive product decision remains.

| Question | Decision | Evidence |
| --- | --- | --- |
| What may fence a cohort? | Only the new closed typed `verified_invalid` result. A status code alone, an unverified response, a timeout, transport failure, and a provider failure cannot change cohort health. | #3865 objective and acceptance. |
| What identity enters the coordinator? | Only #3863's `connectors.AuthCohortKey`; it is opaque and distinct from rate-budget scope keys. | `internal/connectors/coordination_identity.go`; #3863 verification. |
| How are siblings stopped? | Every admitted member receives a derived context. The coordinator changes health atomically, then cancels active same-epoch members; callers must pass the derived context to their normal admission/send boundary. | #3865 acceptance; `internal/coordination/rate_budget_coordinator.go` lifecycle pattern. |
| What proves repair? | A separate verified repair/test typed outcome starts a new healthy epoch. Existing members are cancelled and denied as stale. | #3865 objective. |
| What survives restart? | The coordinator uses a minimal opaque-key health store seam. The in-memory implementation enables deterministic reload/race proof here; durable application wiring is intentionally deferred under #3862/#4015. | #3862 boundary; #3865 restart acceptance. |
| Is provider, app, UDS, checkpoint, or parking integration part of this PR? | No. This is connector-neutral coordination foundation work only; #3867 owns persisted rate parking/resumption and production wiring remains a future #4015 child. | #3862/#3865/#3867 boundaries. |
