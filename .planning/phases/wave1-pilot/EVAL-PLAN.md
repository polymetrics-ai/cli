# EVAL-PLAN — wave1-pilot

Exit metrics (all must hold at P-14; measured, recorded in VERIFICATION.md/SUMMARY.md):

| # | Metric | Target | Measurement |
|---|---|---|---|
| 1 | Migration outcome | 10/10 connectors `migrated`, OR `partial`/`blocked` ONLY with typed blockers per result.schema.json (no silent approximations); quarantined ≤2 | result JSONs + quarantine.json |
| 2 | Parity | `go test ./internal/connectors/paritytest/...` green AND wave0 goldens still green (`go test ./internal/connectors/engine -run TestParity` + native/postgres) | P-14 gate output |
| 3 | Validate/lint findings | 0 findings: `connectorgen validate` clean for all 10 bundles; `make verify` (incl. golangci-lint) green | P-14 gate output |
| 4 | Conformance | `TestConformance/<name>` green for all migrated pilots (static + dynamic incl. `pagination_terminates`, `cursor_advances` with real wire-shape fixtures) | P-14 gate output |
| 5 | Fable review verdict | 10/10 reviewed line-by-line (100% coverage), verdicts recorded, zero unresolved blocker findings after ≤1 repair round per connector | traces/review-*.json |
| 6 | Cost report | `docs/migration/pilot-costs.json` complete: one row per connector with real token/duration data, totals, projection; cost-log.jsonl has one row per dispatch | file inspection vs dispatch count |
| 7 | Recipe patched | conventions.md + executor template updated with ≥ the pilot's confirmed learnings (parity-test location rule, hook patterns, envelope-unwrap example) and every new deviation ledgered | P-12 diff review |
| 8 | TDD honesty | TDD-GATE.json contains real task rows with red-first evidence for P-0 and all 10 bundles (no empty arrays — wave0 B3 lesson) | file inspection |
| 9 | Coexistence | Legacy tree zero-diff (git diff limited to `internal/connectors/<pilot>/` empty for all 10); registryset regen byte-identical | P-14 path guard + diff |
| 10 | Human gate honored | Pass B decision presented to user with the cost report; decision recorded in DECISIONS.md (not made autonomously) | DECISIONS.md entry |

## Prompt-eval notes (executor/reviewer template calibration)

The pilot doubles as the eval run for the fan-out prompts (waves 2–4 reuse them at ~100x scale):

- Per-connector, record in traces: did the executor deviate from conventions.md, and was the
  deviation traceable to prompt ambiguity vs agent error vs genuine engine gap? (input to P-12).
- Track: retries per connector, blocker precision (were typed blockers real?), self-reported
  status vs reviewer verdict agreement (a false "migrated" is the worst outcome — count = 0
  target), fixture-realism failures caught only at review (should be 0 after conventions §4).
- Success bar for the template: ≥7/10 connectors need ZERO prompt-attributable repairs; anything
  systemic (same misreading by ≥3 agents) becomes a mandatory template patch in P-12 and is
  called out in SUMMARY.md for the wave2 planner.

## Anti-goals (fail the phase even if metrics pass)

- Any weakened test/gate to reach green (reviewer dimension-5 check, wave0 discipline).
- Fixtures falsified away from real wire shapes (the B2 failure mode).
- Hook packages exceeding caps "to make it fit" instead of escalating tier or blocking.
