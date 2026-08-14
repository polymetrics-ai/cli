# Discussion log — Issue 4126

The authoritative issue and parent correction settle the material product
questions: approvals are one-time proofs, standing authorization is
per-connection and revocable, and scope binds shape rather than payload. No
additional product choice is needed for this foundation.

The implementation adopts the already trusted `ReversePlan.ExpiresAt` / sealed
plan expiry for the record's expiry. This does not introduce a new duration
policy; a future durable-lifetime policy remains a consumer/product decision.
