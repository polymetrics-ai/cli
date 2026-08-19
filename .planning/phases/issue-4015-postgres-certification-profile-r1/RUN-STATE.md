# Run state — PostgreSQL certification profile

- Status: implementation and local verification complete — inline review and direct PR pending for #4192
- Lifecycle: inline/manual GSD fallback, prompted with `--auto`; no compatible isolated GSD role runtime is available and the autonomous launch brief forbids waiting.
- Base verified: `404536538038e20c3010692ec8fb31e87b11f72f` (`origin/integration/4015-mvp-flat-r1`).
- Decision: The delivery owns a single unstacked foundation-and-PostgreSQL PR for #4192. `allStagesPassed` is exclusively owned here. Implement only the exact declared PostgreSQL polling-watermark → managed-target transport adapter and its evidence path; preserve benign environment/safety skips but never let an unexecutable declared pair pass. Escalate rather than silently generalizing if another connector requires a broader framework.
