# Discussion Log

## Inputs resolved from the launch brief

- Route order is locked: PostgreSQL → GitHub, GitHub → PostgreSQL, GitHub → GitHub, then flow/schedule.
- Correctness precedes scale; the optional 5 GB run is attempted only after all four verdicts are proven and time remains.
- A passing command is insufficient. Every route needs an independently read destination, exact count, and named content sample.
- Each route is run a second time and described using its declared sync mode.
- GitHub authorization and containment are fully specified by the brief; no human clarification is required.
- Destructive cleanup is explicitly authorized only for task-created `pm-cert-` fixtures in the named repository, and absence must be observed as HTTP 404.
- Product defects are recorded, not fixed.

## Implementation decisions

1. Use a fresh binary integration test so credentials remain in memory and provider interactions can be cleaned in `t.Cleanup` even after an assertion fails.
2. Reuse the repository's PostgreSQL `dbtest` harness; do not start or restart a shared runtime.
3. Use typed GitHub connector actions for pipeline writes. Independent HTTP clients are read-back and cleanup observers only, never substitutes for the pipeline under test.
4. Keep all route data small and uniquely named. The test may reuse an existing fixture issue only as a target; every label/comment/object it creates is task-owned and removed.
5. Use the installed crontab payload as the scheduling entry point, matching the production scheduler contract, then inspect durable schedule state and destination content.
6. Execute GSD inline because the canonical single-worker contract forbids spawning delivery roles.

## Red state

At task start the repository has deterministic/provider-double cross-system coverage and historical live tests outside the authorized fixture boundary, but no run-specific live evidence for the three requested routes against `Polymetrics-Cert/pm-cert-3993-20260810-wz0fru`, and no run-specific scheduled cross-system proof.
