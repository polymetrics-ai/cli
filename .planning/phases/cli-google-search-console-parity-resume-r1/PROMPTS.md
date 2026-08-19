# GSD Prompt Snapshot — Google Search Console documented-operation parity resume

## Kickoff

Rehydrate PR #3555’s Google Search Console bundle on current main, fix the nine validator findings,
research every request-field citation from Google-owned documentation, and make every documented
operation reachable through safe typed streams, direct reads, or `writes.json` actions. Never
restore stale shared runtime changes, add a generic raw API surface, execute live provider calls, or
mark `rest_write` as implemented.

## Inputs

- `AGENTS.md`
- `docs/migration/conventions.md`
- `docs/architecture/connector-architecture-v2-design.md`
- `/Users/karthiksivadas/karthik-agent-workspace/data/shared/connector-parity-resume-contract.md`
- `docs/plans/universal-programming-loop-prd.md`
- `docs/prompts/universal-programming-loop-prompts.md`

## Downstream artifacts

- `PLAN.md`, `TDD-LEDGER.md`, `VERIFICATION.md`, `RUN-STATE.json`, and the provider-field
  research matrix in this phase directory.

## Verification result

Completed through the manual-GSD fallback: all focused current-main red/green cycles pass.
The final provider inventory is 11 unique operations, all genuinely reachable; the generated
surface has 15 commands and zero planned operations. `REQUEST-FIELD-MATRIX.md` contains 32/32
provider-owned field citations while the shared canonical bundle convention remains unlanded on
current `origin/main`.
