# Discussion log — bounded ordered pipeline throughput R1

The brief and `cli-parallelism-caps-audit-r1` resolve the product choices. Inline/manual `discuss-phase 4183 --auto` is used because no compatible Pi role runtime is available and the task requires autonomous execution.

| Topic | Decision | Source |
| --- | --- | --- |
| Throughput lever | Decouple only the transformed Arrow full-overwrite `emit` callback through a bounded ordered queue. | Parallelism audit, Executive Answer and Recommended Delivery Sequence. |
| Delivery semantics | Keep #4184's run-scoped shadow publication, receipt/read-back, and exactly one final checkpoint. | Task brief; existing #4183 evidence. |
| Depth compatibility | Depth one uses the existing serial path exactly. | Task acceptance. |
| Capability gate | Both declarative endpoints must opt in to the ordered pipeline before any endpoint I/O. | Task brief; audit Concrete CLI Proposal. |
| Target worker policy | Persist a bounded PostgreSQL full-overwrite COPY capacity on the connection and show it in plan/preview; no multi-COPY implementation in this slice. | Task brief and audit scope fence. |
| Benchmark reporting | Re-run before and after only on a quiet machine, retain prior values verbatim, and report a missed gate plainly with stage timings. | Task brief. |
