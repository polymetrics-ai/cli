# Discussion Log — issue #4158 / Production MVP verify green

`discuss-phase --auto` ran inline. Firstmate explicitly instructed autonomous execution; no unresolved product decision remains.

| Area | Decision | Reason |
| --- | --- | --- |
| Test truth | Treat both failures as defects until falsified. | The task and #4158 prohibit accepting a route refusal as a replacement for durable acknowledgement. |
| Scope | PostgreSQL managed-target admission is the one target connector; GitHub is an external binary reproducer. | This honors the connector ownership guard while testing the claimed common cause. |
| Safety | Retain non-PostgreSQL refusal before I/O with its existing typed error. | #4158 makes that safety property explicit. |
| Evidence | Keep original test evidence, an added production-path regression, trigger/mask/symptom, divergent path, counterfactual, and falsifier results. | The launch brief requires causal rather than proximity-based attribution. |
