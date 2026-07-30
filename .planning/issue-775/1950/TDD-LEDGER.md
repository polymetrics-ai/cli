# TDD Ledger — Issue #1950 Lucid ELD Operation Ledger

## Skill load record

- `gsd-core`: required GSD workflow; adapter command unavailable for `programming-loop`, so manual fallback recorded.
- `golang-how-to`: skill-routing orchestrator; connector runtime/operation ledger work routes to task-specific Go skills.
- `golang-cli`: CLI surface awareness; no runtime CLI code changes in this issue.
- `golang-spf13-cobra`: issue contract requires CLI command-tree rules; no Cobra edits.
- `golang-spf13-viper`: issue contract requires config layering rules; no Viper edits.
- `golang-testing`: Best Practices #1 named cases, #3 independent tests, #5 observable behavior; implemented as planning-only corpus validation, not production Go tests due fixed write scope.
- `golang-error-handling`: Best Practices #1 checked errors, #2 contextual failures, #3 lowercase error strings; planning validator returns explicit fixture failures.
- `golang-security`: Security Thinking #1-3 trust boundaries/attacker control/blast radius; no secrets, no credentialed calls, only public docs fetches.
- `golang-safety`: Best Practices #2 safe type checks, #4 initialize maps, #8 numeric conversion care; applies to planning validator and JSON parsing.
- `golang-design-patterns`: Best Practices #20 avoid new dependencies, #21 design for testability; planning validator uses stdlib only and stays outside production.
- `golang-structs-interfaces`: small/explicit data shapes; no new exported Go APIs.
- `golang-context`: Best Practices #5 cancel/timeout ownership; public fetches use bounded timeouts.
- `golang-documentation`: Writing principles — no invented context, preserve obligations; official-source evidence must cite source URLs/hashes.
- `caveman`: final handoff only; exact commands/output not compressed.

## Red/green/refactor ledger

| Time | Phase | Artifact/command | Expected | Actual |
|---|---|---|---|---|
| 2026-07-30 | RED | `python3 .planning/issue-775/1950/tools/validate_surface.py --surface .planning/issue-775/1950/fixtures/red/missing-endpoints.api_surface.json --openapi .planning/issue-775/1950/evidence/openapi-doc.json` | fail: missing known endpoint(s) | FAIL exit=1; `missing official endpoint GET /v2/company-info` |
| 2026-07-30 | NEGATIVE | duplicate/unknown-target/invalid-category/wildcard/stale fixtures | fail for each intended rule | all fixtures failed with intended rule(s); wildcard also reported missing official endpoint and extra wildcard endpoint |
| implementation | GREEN planned | same validator against `internal/connectors/defs/lucid-eld/api_surface.json` | pass: 8/8 official operations exactly once | pending |
| verification | REPO planned | issue verification command block | pass or documented incomplete-bundle failures | pending |

## TDD scope note

Fixed write scope forbids editing shared Go tests under `cmd/` or `internal/`. The failing test/corpus fixture for #1950 is therefore planning-owned under `.planning/issue-775/1950/**`; production edit starts only after red evidence is captured.
