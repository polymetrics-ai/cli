# OBSERVABILITY — wave1-pilot

Runtime observability is unchanged from wave0 (engine error mapping through
`safety.RedactErrorText`; no new production paths this phase). The phase-specific observability
deliverable is the **pilot cost capture** feeding the Pass B decision (PLAN.md P-13).

## Pilot cost capture method

- **Who records**: the coordinator, at dispatch boundaries — agents cannot introspect their own
  token usage. Wave0 established the baseline numbers this method must reproduce
  (`.planning/phases/wave0-engine-harness/SUMMARY.md` "Numbers for the pilot cost model":
  Sonnet executors ~100k–460k tokens; Fable reviewer ~130k–200k).
- **What, per dispatch**: agent role/model, connector, start/end wall clock, input/output/total
  tokens from the agent-run usage stats, exit status. Repairs and re-reviews are separate rows
  attributed to the same connector (`dispatches` count).
- **Where**: appended to a working log
  `.planning/phases/wave1-pilot/traces/cost-log.jsonl` (one JSON object per dispatch, written by
  the coordinator immediately after each agent returns — not reconstructed at phase end), then
  aggregated into `docs/migration/pilot-costs.json` in P-13.
- **pilot-costs.json shape** (consumed by the Pass B decision conversation):

```json
{
  "phase": "wave1-pilot",
  "recorded_at": "<RFC3339>",
  "connectors": [
    {"name": "...", "bucket": "S|M|L|XL", "tier": 1, "model": "sonnet",
     "input_tokens": 0, "output_tokens": 0, "total_tokens": 0,
     "wall_clock_minutes": 0, "dispatches": 1, "status": "migrated|partial|blocked",
     "deviations_count": 0, "blockers": [], "reviewer_verdict": "pass|fail"}
  ],
  "review_overhead": {"model": "fable", "total_tokens": 0, "verdicts": 10},
  "totals": {"executor_tokens": 0, "review_tokens": 0, "repair_tokens": 0},
  "projection": {
    "method": "per-bucket mean total_tokens x remaining inventory histogram S137/M388/L31/XL1",
    "pass_a_remaining_estimate_tokens": 0,
    "pass_b_note": "expansion multiplier decision input; see orchestration-plan Budget truth"
  }
}
```

- **Quality signals captured alongside cost** (same rows): deviations per connector
  (result JSON `parity_deviations` count), blocker types, repair count — so the Pass B decision
  sees cost AND defect-rate per bucket, not cost alone.
- **Verification observability**: P-14 records every gate command + output + HEAD in
  VERIFICATION.md (wave0 B3 lesson — a gate that isn't recorded didn't happen).
