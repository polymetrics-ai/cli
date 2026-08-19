# Discussion log — issue #4273

`scripts/gsd prompt discuss-phase 4273 --auto` was resolved manually and inline. No human decision was left open because the Firstmate task fixes the batch size, source policy, placement, and non-goals.

| Area | Decision | Rationale |
| --- | --- | --- |
| Batch selection | 20 eligible entries from the external provider-artifact ledger | The pipeline validates the evidence and produces a deterministic manifest. |
| Source material | Read-only corpus for source bundles; materializer for cited public artifacts | Reuses the existing pipeline and avoids a scraper/browser-first approach. |
| Unsupported parity classes | Leave absent/unimplemented and log a foundation gap | Transport/direct-read/binary claims require real executors and flow proof. |
| Delivery | One direct PR to `main`, with a committed checkpoint before materialization | The ledger remains resumable if a later materialization drops a candidate. |
