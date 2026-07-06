# Coordinator decisions — wave1-pilot

Answers to planner open questions (2026-07-02):

1. **DW-1 staggering**: dispatch github + monday FIRST (long poles: Tier-2 hooks), remaining 8 in a
   second parallel batch ~immediately after both long poles report their hook-package scaffolds
   compile (no need to wait for full completion — paths are disjoint).
2. **P-11 reviewer shape**: 2 parallel Fable reviewers, 5 connectors each (github+gmail+monday+
   sentry+chargebee / xkcd+vitally+bitly+calendly+zendesk-support) — hook-heavy connectors grouped
   with the stronger scrutiny batch.
3. **N2 validate-time guard**: YES — bundled into P-0 (all-digits non-unix start-value guard in
   connectorgen validate while the formatParam context is warm).
4. **Wave0 N7 (11MB blob in history)**: deferred to pre-push; branch is local-only. Recorded as a
   pre-push checklist item in RUNBOOK.md scope.
5. **monday approach**: Tier-2 StreamHook (GraphQL POST + in-body pagination is a documented Tier-2
   trigger; `StreamSpec.Body` stays unwired in wave1 — if ≥3 wave2+ connectors need POST-body
   streams, the ENGINE_GAP ≥3 rule triggers a native engine feature then). The pilot deliberately
   exercises all three hook seams: github (AuthHook github_app + WriteHook compound), gmail
   (AuthHook oauth2 refresh grant), monday (StreamHook GraphQL).
