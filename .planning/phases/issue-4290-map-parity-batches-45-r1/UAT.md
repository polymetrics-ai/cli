# UAT — Issue #4290

Manual GSD verification fallback: the project-local Pi adapter cannot provide the
isolated verifier runtime. This documentation-only deliverable has no
judgment-dependent UI or credentialed provider behavior; its acceptance criteria
are therefore covered by deterministic local assertions.

| Deliverable | Evidence | Result |
| --- | --- | --- |
| Twenty complete source-locked maps | `materialize-parity-maps.mjs check batch4` and `check batch5` | pass |
| Exact public-source pins or honest browser skips | checker requires SHA-256/positive bytes for fetched sources and no pin for TikTok/eBay skips | pass |
| Correct direct-write/reverse-ETL model | checker forbids `reverse_etl` endpoint classes and requires the exact destination-executor eligibility evidence on typed direct writes | pass |
| Definition compatibility | `connectorgen validate`, `surface-sync --check`, and `connector-boundary` | pass |

No provider credentials, provider requests, or destructive exercises were used.
