# Discussion log — issue 4364

## 2026-08-27 — inline `discuss-phase 4364`

The Firstmate brief resolves the material decisions:

1. Use the 4,341-record source-operation mapping manifest, never the prior 1,908-row deferred matrix.
2. Keep deferred records discoverable and exact, but not executable. Their preflight ends in the exact `system/missing_foundation` terminal before transport setup, credential lookup, or provider I/O.
3. Do not synthesize `operations.json`, `streams.json`, or `writes.json` entries for a missing typed lane. Later foundation owners promote the same command path only with a real typed artifact and a credential-bound proof.
4. Preserve every source class—including delete/destructive, reverse ETL, binary, ETL, unsafe, and provider mutation—as a source-visible row rather than using a policy-only exclusion.
5. This is a shared foundation issue, not a connector lane: the generated data spans ten providers and must be driven from one reviewed manifest with closed validation.

The discussion has no unanswered product choice. The remaining work is technical validation through the TDD plan.

