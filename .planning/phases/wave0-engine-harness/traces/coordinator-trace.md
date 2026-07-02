# Agent Trace: coordinator

Coordinator: main session (fable). Started 2026-07-02.

## Rendered Prompt Or Prompt Reference

docs/prompts/universal-programming-loop-prompts.md (kickoff) + .planning/phases/wave0-engine-harness/PROMPTS.md

## Lifecycle log

- Preflight run (`programming-loop.mjs run --phase wave0-engine-harness --subagents true --mode agents`)
  → initial `blocked_missing_artifact` (expected pre-PLAN).
- PLAN stage: gsd-loop-planner (fable) wrote SPEC/PLAN/TEST-PLAN/THREAT-MODEL/RUNBOOK/DATA-MODEL/
  API-CONTRACT/OBSERVABILITY/EVAL-PLAN + ADR-0001 + postmortem template. PRD coverage now PASSES
  (n/a markers: design-direction, release-notes, postmortem-template).
- Coordinator decisions recorded in DECISIONS.md (golangci-lint via brew binary 2.12.2; B-12 floats
  to Wave B if capacity; inventory loc = all .go; goldens get minimal honest api_surface).
- Plan-check stage: gsd-plan-checker (sonnet) dispatched → PLAN-CHECK.md (pending).

## TDD gate note

`tdd-gate.mjs` extracts checkbox-style tasks from PLAN.md; our PLAN uses heading-based tasks, so
the script gate passes trivially. Red-first is therefore enforced by orchestration: every executor
prompt requires T-N authored + run RED (output captured in the ledger) before B-N code; the
coordinator appends per-task ledger entries from agent trace files after each dispatch wave and
spot-checks the red evidence. Same discipline, enforced at the prompt/ledger level.

## Dispatch strategy (package-conflict aware)

Waves A–C all mutate the single `internal/connectors/engine` package — parallel agents there would
interfere via package-wide compilation during tests. Dispatch:

1. Run 1 (sonnet): Wave A — T/B-01..04 sequential in one run (engine foundations).
2. Run 2 (sonnet): Wave B — T/B-05,06,07.
3. Run 3 (sonnet): Wave C — T/B-08,09.
4. Runs 4–6 (parallel, disjoint packages): B-10 (engine/connector.go + connectors/definition.go),
   B-11 (cmd/connectorgen), B-12 (certify report/harness).
5. Runs 7–8 (parallel): B-13 (conformance), B-14 (certify source stages).
6. Runs 9–10 (parallel): B-15 (defs/stripe + parity), B-17 (native/postgres);
   then Run 11 (serial): B-16 (defs/searxng + parity + registrygen skip-map single legacy edit).
7. Runs 12–14 (parallel): B-18 (.golangci.yml/Makefile), B-19 (inventorygen), D-20 (docs, after F).
8. V-21: gsd-verifier + gsd-loop-reviewer (fable) + gsd-loop-security (sonnet).

Git: single-writer — agents do NOT commit; coordinator commits after each dispatch wave.
Agents write ledger evidence to traces/<task>-ledger.md (disjoint files); coordinator merges into
TDD-LEDGER.md.

## Execution log

- Wave A (T/B-01..04) + T/B-12 (floated) complete, gate green → commit c05edb4 (30 files, 75 tests).
- T/B-19 (floated, disjoint package): inventory.json generated — **557 connectors, real buckets
  S:137 / M:388 / L:31 / XL:1** (loc incl. tests). This is milder than orchestration-plan's
  estimate (S90/M294/L155/XL20): projected Pass A fan-out drops ~105 → **~77 bundle agents**.
  Feed into wave1 pilot-cost report + wave2-4 rosters. Commit edf40a6.
- Wave B (T/B-07, 05, 06) dispatched, in flight.

## Verification Evidence

- PRD coverage: passed=True after n/a markers (prd-coverage.mjs re-run).
- golangci-lint 2.12.2 installed (brew).

## Unresolved Risks

- Wave F parity tests share the engine package's test namespace — B-16 serialized after B-15/B-17.
- `install` check in resolve-verification has no Go mapping (script is JS-ecosystem-centric);
  coordinator runs `go mod download && go mod verify` manually at VERIFY and documents it.
