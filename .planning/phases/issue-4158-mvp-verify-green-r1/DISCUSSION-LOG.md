# Discussion Log — issue #4158 / Production MVP verify green

`discuss-phase --auto` ran inline. Firstmate explicitly instructed autonomous execution; no unresolved product decision remains.

| Area | Decision | Reason |
| --- | --- | --- |
| Test truth | Treat both failures as defects until falsified. | The task and #4158 prohibit accepting a route refusal as a replacement for durable acknowledgement. |
| Scope | The GitHub external-binary fixture is the corrected scope; PostgreSQL #4158 is independent and remains separately owned. | Its execution path never creates a managed-target driver or history route. |
| Safety | Retain non-PostgreSQL refusal before I/O with its existing typed error. | #4158 makes that safety property explicit. |
| Evidence | Keep original test evidence, an added production-path regression, trigger/mask/symptom, divergent path, counterfactual, and falsifier results. | The launch brief requires causal rather than proximity-based attribution. |
| Contract selection | Migrate the fixture to the #4168 job-only contract; do not restore legacy inline action scope. | Firstmate's option-1 decision preserves approved-job authorization as the execution-time boundary. |
