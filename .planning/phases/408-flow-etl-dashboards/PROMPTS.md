# Prompts — Phase 408 flow/ETL dashboards

## Kickoff snapshot

Command path:

```bash
scripts/gsd prompt plan-phase 408 --skip-research
scripts/gsd prompt programming-loop init --phase 408-flow-etl-dashboards --dry-run
```

Result:

- `plan-phase` prompt generated successfully at `/tmp/gsd-plan-408.txt`.
- `programming-loop` prompt unavailable: `scripts/gsd: unknown GSD command: programming-loop`.
- Manual universal-loop fallback active per `.agents/agentic-delivery/workflows/gsd-universal-runtime-loop.md`.

Downstream artifact:

- `.planning/phases/408-flow-etl-dashboards/PLAN.md`
- `.planning/phases/408-flow-etl-dashboards/TDD-LEDGER.md`
- `.planning/phases/408-flow-etl-dashboards/VERIFICATION.md`
- `.planning/phases/408-flow-etl-dashboards/SUMMARY.md`
- `.planning/phases/408-flow-etl-dashboards/RUN-STATE.json`

Verification result:

- Planning preflight complete.
- Production verification pending.
