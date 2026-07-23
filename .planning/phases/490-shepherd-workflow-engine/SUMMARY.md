# Summary: #490

Status: implementation checkpoint green; exact-head review pending.

The decision is partial adoption: exact `pi-workflow-engine@0.12.0` is the project-local bounded analysis/review orchestrator, while Shepherd's `ProductionAgentSessionPort` and all durable authority remain unchanged. Pi compatibility is narrowed to stable `>=0.80.10 <0.80.11`; prompt settlement plus a validated bound terminal handoff is authoritative and unknown raw events are inert telemetry. Focused tests, the pre-review full suite, strict typecheck, exact runtime-family verifier, and offline Shepherd/workflow-engine/co-load RPC gates pass. Exactly one xhigh exact-head review, its dispositions, the final full gate, PR, and issue comments remain.
