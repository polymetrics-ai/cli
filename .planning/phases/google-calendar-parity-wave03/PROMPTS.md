# Prompt Snapshots

## Resume R1 manual-GSD kickoff

- Agent role: coordinator and connector implementer (local critical path)
- Adapter result: `scripts/gsd prompt programming-loop ...` is unavailable (`unknown GSD command`)
- Downstream artifact: `PLAN.md`, `SPEC.md`, `TEST-PLAN.md`, and `TDD-LEDGER.md` updated after recovery/rebase.
- Verification result: red evidence captured; implementation, bundle validation, runtime activation, and generated-surface refresh complete; focused contract gates pending.

```text
Recover PR #3554 on current main, retain only connector-owned behavior, audit every official Google Calendar v3 operation, make only runtime-reachable reads executable, and record each rest_write mutation as blocked with provider-field evidence.
```

## 2026-08-05 shared-contract correction

Section 3b confirms that record-driven reverse-ETL writes execute through `writes.json` today. The superseded kickoff instruction above was corrected before final authoring: all 26 Google Calendar mutations are record-shaped, so they are implemented as typed write actions with fixtures, confirmations where destructive, and provider-field citations. No Google Calendar operation uses or remains blocked on `rest_write`.
