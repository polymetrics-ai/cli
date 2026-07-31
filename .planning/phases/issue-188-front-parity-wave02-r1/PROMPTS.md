# PROMPTS — issue #188 Front parity

## Kickoff snapshot

- Command path: `scripts/gsd prompt quick --full 'Implement issue #188 Front full documented connector parity in internal/connectors/defs/front only; update GSD plan, TDD ledger, verification, conformance evidence; no live provider calls, no no-mistakes, no push/PR.'`
- Trace artifact: `.planning/traces/issue-188-front-parity-wave02-r1-gsd-quick.prompt.md`
- Runtime fallback: `scripts/gsd prompt programming-loop ...` is unavailable in this adapter, so the Pi programming-loop prompt `.pi/prompts/pm-gsd-loop.md` and `.agents/agentic-delivery/workflows/gsd-universal-runtime-loop.md` were followed manually.
- Downstream artifact: `PLAN.md`, `TDD-LEDGER.md`, `VERIFICATION.md`, `RUN-STATE.json`, `SUMMARY.md`.
- Verification result: issue-local gates passed; see `VERIFICATION.md`.
