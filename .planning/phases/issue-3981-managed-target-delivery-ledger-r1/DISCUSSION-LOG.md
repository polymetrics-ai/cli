# Discussion log — Issue 3981: target delivery ledger

The issue body, its 2026-08-14 Production MVP amendment, and the dedicated
verification report were treated as the authoritative discussion record. No
unresolved product decision remains:

1. The existing owner/provisioning kernel is already delivered and out of scope.
2. The residual is a delivery ledger keyed by structural target identity.
3. It must survive mutable artifact-display/table rename and process restart.
4. It must isolate records for distinct immutable streams under one owner.
5. A native driver, DDL, SQL, actual delivery/mode application, source checkpoint,
   and CLI are expressly deferred to #3973, #3982, and #3983.

`scripts/gsd prompt discuss-phase issue-3981-managed-target-delivery-ledger-r1`
was generated and executed inline. The task is an issue foundation rather than
a numbered roadmap phase, and the canonical single-worker contract forbids GSD
role spawning; this file is the documented manual/inline fallback.
