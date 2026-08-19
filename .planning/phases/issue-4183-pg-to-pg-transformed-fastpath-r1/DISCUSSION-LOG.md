# Discussion log — issue #4183

The captain supplied the design rather than requesting discovery. Inline/manual `discuss-phase` fallback is therefore recorded: no compatible Pi interactive runtime is available in this worker, and no open product choice remains.

| Topic | Decision | Source |
| --- | --- | --- |
| Delivery route | Direct PR to `integration/4015-mvp-flat-r1`; never `main`. | Launch brief; base verified locally at `94d69972c`. |
| Product slice | PostgreSQL-to-PostgreSQL transformed `full_overwrite` only. | Throughput report, “Next PR: one honest vertical slice”. |
| First risk | Characterize two-plus-page truncate behavior before optimizing it. | Launch brief; report in-scope item 1. |
| Transform language | Normalized, hashed `TransformPlanV1`; no arbitrary SQL. | Throughput report. |
| Port boundary | Generic segment/controller/transform/receipt/checkpoint contracts; PostgreSQL only implements range extraction and binary COPY/shadow publish. | Captain constraint, 2026-08-16. |
| Execution lifetime | No run deadline; bounded unit attempts and progress-bearing durable telemetry. | Unbounded-run architecture report and #4182 base increment. |
| Performance claim | ≥200 decimal MB/s only on the defined qualified host; report measured number and phase bottleneck otherwise. | Throughput report. |

