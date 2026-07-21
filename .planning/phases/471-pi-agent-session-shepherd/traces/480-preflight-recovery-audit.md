# Issue #480 preflight recovery and cutover audit

Date: 2026-07-21
Parent head audited: `0a7fc179`
Mode: read-only planning/review (`openai-codex/gpt-5.6-sol`, `xhigh`)

## Outcome

#480 is a recovery, audit, and reversible-cutover preparation slice. It cannot truthfully activate
legacy-shell deprecation before #481 proves the replacement, so the activation belongs to a
parent-owned post-canary finalization commit. A failed or interrupted canary leaves the rollback
path and its documentation unchanged.

## Required #479 seams

#480 depends on these concrete #479 contracts rather than repairing them after wiring:

1. validated autonomous state v2 round-trips all scheduler, review, decision, and effect truth;
2. lease-fenced revision/CAS persistence behaves as one durable transaction boundary;
3. every external mutation has a prepared/observed/consumed/applied effect-journal entry;
4. bounded redacted audit events are emitted through an explicit outbox/port;
5. reconciliation obtains one stable, complete authoritative truth snapshot;
6. recovery is a hard pre-schedule barrier, never a best-effort background task;
7. execution epoch is separate from logical retry/correction generation;
8. every external port accepts `AbortSignal`, has a deadline, and exposes join/settlement truth;
9. stop, retry, rate-limit, and human-wait state is durable; and
10. no controller port can merge the parent PR to `main`.

If any seam is absent from #479, #480 must fail at planning/RED rather than create a parallel
recovery-only state model.

## RED slices

1. **Recovery classifier and ordering:** enumerate prepared-only, externally-observed,
   locally-applied, stale-generation, corrupt, and impossible states; complete reconciliation before
   the scheduler can reserve work.
2. **Per-effect idempotent reconciliation:** comments, commits, pushes, PR creation/update,
   integration, review requests, and human-decision consumption each reconcile by exact durable
   key and authoritative observation.
3. **Audit schema/store:** bounded causal IDs, parent/child generation, effect key, stage transition,
   redacted typed outcome, rotation/retention, and fail-closed validation.
4. **Fault matrix:** crash before/after every journal checkpoint and audit/state write; prove one
   outcome and no duplicate mutation after restart.
5. **Operational failures:** stop/process loss, network timeout, retry/rate-limit windows, stale or
   force-moved heads, conflicts, changed reviews/check policy, and incomplete child settlement.
6. **Operator UX:** deterministic status/recovery explanations, safe resume/stop syntax, trusted
   local boundary, and exact authenticated human-decision instructions.
7. **Cutover preparation:** generate and verify the rollback/deprecation delta without applying it;
   preserve the legacy shell route and historical Go artifacts.

## Activation sequence

```text
#480 recovery/audit green
  -> stage and review reversible cutover delta (unapplied)
  -> #481 deterministic and designated canary pass
  -> parent orchestrator applies the exact deprecation delta
  -> rerun affected docs/help/Shepherd gates
  -> final exact-head parent review
  -> wait for explicit human merge approval
```

The activation commit is invalid if the canary evidence is missing, stale, interrupted, or bound to
a different parent head. It is not a worker inference and never authorizes parent-to-main merge.
