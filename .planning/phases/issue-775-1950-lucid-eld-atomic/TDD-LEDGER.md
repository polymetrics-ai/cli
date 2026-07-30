# TDD Ledger — Issue #1950 Lucid ELD Atomic Pilot Bundle

## Skill load record

- `gsd-core`: implementation workflow steps 1-9; manual fallback allowed only with recorded adapter failure.
- `caveman`: compact handoff only; exact commands/output/security warnings not compressed.
- `golang-how-to`: Go orchestration routing loaded first; connector runtime/architecture maps to testing, error handling, security, safety, lint, design, structs/interfaces, CLI, context, documentation.
- `golang-cli`: stdout/stderr/JSON/help parity awareness; no shared Cobra code edits planned.
- `golang-spf13-cobra`: Best Practices #1 RunE, #3 Args validators, #4 command output via `cmd.OutOrStdout`; no Cobra edits.
- `golang-spf13-viper`: Best Practices #4 test isolation, #5 bind flags before execute; no Viper edits.
- `golang-testing`: Best Practices #1 named cases, #3 independent tests, #5 observable behavior; use existing conformance/CLI tests and connector fixtures as executable specs.
- `golang-error-handling`: Best Practices #1 check errors, #2 wrap context, #3 lowercase errors; applies to validation interpretation and no swallowed failures.
- `golang-security`: Security Thinking #1 trust boundaries, #2 attacker-controlled inputs, #3 blast radius; no secrets/live data/raw write surfaces.
- `golang-safety`: Best Practices #2 safe type assertions, #4 initialized maps, #8 numeric conversion care; applies to JSON fixture/schema correctness and not inventing types.
- `golang-lint`: Development Workflow #1 run linters/vet after significant changes, #3 format before committing.
- `golang-design-patterns`: Best Practices #20 prefer recode over dependency, #21 design for testability; zero new dependencies and Tier-1 data-only bundle.
- `golang-structs-interfaces`: interface principle “keep interfaces small” and “don’t create interfaces prematurely”; no new Go API/interfaces.
- `golang-context`: Best Practices #1 propagate cancellation and #5 call cancel; no live external loops added.
- `golang-documentation`: Writing principles “no invented context” and “preserve meaning”; docs state schema limits and no-write truth.

Note: `.pi/skills/go-implementation/SKILL.md` is not present in this checkout; loaded repo-routed Go skills above are the applicable implementation guidance.

## Red / green / refactor ledger

| Time (UTC) | Phase | Artifact/command | Expected | Actual |
|---|---|---|---|---|
| 2026-07-30 (prior cycle) | RED | `python3 .planning/issue-775/1950/tools/validate_surface.py --surface .planning/issue-775/1950/fixtures/red/missing-endpoints.api_surface.json --openapi .planning/issue-775/1950/evidence/openapi-doc.json` | fail: missing official endpoint(s) | FAIL exit=1; `missing official endpoint GET /v2/company-info` |
| 2026-07-30 (prior cycle) | GREEN-LEDGER | `python3 .planning/issue-775/1950/tools/validate_surface.py --surface internal/connectors/defs/lucid-eld/api_surface.json --openapi .planning/issue-775/1950/evidence/openapi-doc.json` | pass: 8/8 official operations exactly once | PASS; `8 endpoint(s) match official OpenAPI` |
| 2026-07-30T02:52Z | RED-CI | `scripts/verify-gsd-workflow feat/775-lucid-eld-full-parity` | fail before relocation: recognized phase evidence missing | FAIL exit=1; `verify-gsd-workflow: cmd/internal changed, but no GSD planning evidence changed.` |
| 2026-07-30T02:52Z | RED-CI | `make connector-boundary` | fail before bundle completion: metadata missing | FAIL exit=2; `connectorgen boundary: load connector metadata lucid-eld: open .../internal/connectors/defs/lucid-eld/metadata.json: no such file or directory` |
| pending | GREEN | `go run ./cmd/connectorgen validate internal/connectors/defs/lucid-eld` | pass, includes secret-literal scanner for fixtures/docs/operations | pending |
| pending | GREEN | `go test ./internal/connectors/conformance -run 'TestConformance/lucid-eld' -count=1` | pass | pending |
| pending | GREEN | focused CLI/vet/build/boundary/GSD/diff/gofmt gates | pass | pending |
| pending | BROAD | `make verify` | pass or exact unrelated baseline blocker | pending |

## Manual GSD fallback evidence

```bash
scripts/gsd doctor
scripts/gsd list
scripts/gsd prompt programming-loop init --phase issue-775-1950-lucid-eld-atomic --dry-run
```

`doctor` passed, `list` showed 69 commands, and the programming-loop prompt returned exactly:

```text
scripts/gsd: unknown GSD command: programming-loop
```

## Test-first decision

No new production Go test is added: this corrective cycle is definition-only under `internal/connectors/defs/lucid-eld/**`. Red evidence is current branch/CI failure from incomplete atomic bundle and unrecognized GSD phase evidence, plus existing planning validator red fixtures. Production bundle edits start only after this ledger/plan/verification update.
