# Discussion log — Issue 3983

## Inline decisions

1. **Consumer boundary:** reuse `ChangeDeliveryWorkset`, `MappingContractV1`, `DeliveryReceiptV1`, `DatabaseWritePlan`, and `DatabaseWriteExecutor`; do not create a duplicate provider-specific payload or a generic writer.
2. **Delete meaning:** source/baseline comparison is solely #3980's workset derivation concern. This delivery consumes `ReadDelta` and `Tombstones`; physical absence carries no mutation authority.
3. **Approval:** native preview remains the source of one-shot approval. The higher-level workset delivery plan must make the workset identity part of its exact-match/staleness check before that approval is consumed.
4. **Recovery:** explicit committed receipt plus ledger evidence is the only baseline-promotion authority. Unknown commit is deliberately terminal/replay-required, never rolled back or retried by the controller.
5. **Scope:** one native PostgreSQL target. The existing warehouse-mediated/one-engine transport boundary is retained; this issue does not expose source-to-target dispatch or change capability claims.

## Manual GSD fallback

The generated `discuss-phase` prompt requires an interactive Pi discussion. This direct-PR worker has a complete issue contract and is prohibited from spawning GSD roles, so it recorded the bounded choices above inline before producing the TDD plan.
