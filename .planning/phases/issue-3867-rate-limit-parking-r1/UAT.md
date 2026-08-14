# UAT — issue #3867 rate-limit parking and automatic resumption

GSD `verify-work` was executed inline. All deliverables have deterministic automated coverage in `SUMMARY.md`; no visual, product-choice, or credentialed-provider judgment is required for this connector-neutral coordination seam.

| ID | Deliverable | Automated evidence | Result |
| --- | --- | --- | --- |
| D1 | Durable, truthful `parked_rate_limit` record | Typed engine bridge test and restart persistence test | pass |
| D2 | Same-scope parking produces zero sends before reset while an unrelated scope remains available | Scope-isolation test plus 64-attempt concurrent zero-send proof | pass |
| D3 | Resume occurs only at/after reset with the exact committed checkpoint and no replayed apply | Restart scheduler test under required package race check | pass |
| D4 | Duplicate parking, cancellation, and callback failure are observable and safe | Lifecycle/idempotency test | pass |
