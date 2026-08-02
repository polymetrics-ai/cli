# TDD LEDGER — issue 3586 generated/unrelated connector remediation

## Loaded skill rule notes

- `gsd-core`: implementation workflow steps 1-9 loaded; use plan-before-code, red/green/refactor evidence, and verification before PR.
- `golang-how-to`: skill-loading matrix loaded; documentation/CLI/test work routes through `golang-cli`, `golang-testing`, `golang-security`, and `golang-documentation`.
- `golang-cli`: I/O and CLI docs parity rules loaded; no CLI runtime behavior edits planned in this worker.
- `golang-testing`: best-practice rules #1 table-driven named cases, #3 independent tests, #5 observable behavior, #10 race in CI loaded; this worker uses validation artifacts rather than new shared Go tests because shared guard code is owned by #3581.
- `golang-error-handling`: best-practice rules #1 check errors, #2 wrap with context, #7 log-or-return loaded; no Go error-handling edits planned.
- `golang-security`: threat-model questions #1-3 and quick-reference command/path injection defenses loaded; no secrets, credentialed checks, generic HTTP/shell/SQL write tools, or reverse ETL execution.
- `golang-safety`: rules #6 defensive copies, #8 conversion safety, #10 useful zero values loaded; no Go data-structure edits planned.
- `golang-documentation`: writing principles and application/CLI documentation checklist loaded; ledger must avoid invented context and preserve safety wording.
- `no-mistakes`: active validation-step/nested gate boundary loaded; run scoped validation only after committed work, stop on daemon or ask-user blocker.
- `caveman`: compact status only; exact commands/test output remain exact.

## Planned red/green evidence

| Slice | Red evidence before production/remediation edit | Green evidence | Status |
| --- | --- | --- | --- |
| A planning | Missing issue ledger/proof validation fails because `REMEDIATION-LEDGER.md` is absent/incomplete. | Plan/TDD/verification artifacts exist and record GSD fallback + skills. | Planned |
| B dispositions | Ledger completeness validation fails until #3532/#3535 audited paths are present with dispositions. | `REMEDIATION-LEDGER.md` covers every audited generated/unrelated/manual/website path with PR URL, merge SHA, class, disposition, evidence. | Planned |
| C guard fixture | Target-aware guard evidence fails because current branch lacks `connectorgen ownership`; fixture/proof absent until created. | `OWNERSHIP-GUARD-FIXTURE.md` lists target scopes and reject/allow examples for #3581/#3587 integration; current focused tests still pass. | Planned |
| D validation | Focused/broader gates pending. | Required verification commands recorded in `VERIFICATION.md`. | Planned |

## Actual evidence log

- 2026-08-02: isolation verified with `pwd -P`, `git rev-parse --show-toplevel`, and `git status --short --branch`; branch is `fix/3586-generated-unrelated-connector-remediation`.
- 2026-08-02: `scripts/gsd doctor` passed.
- 2026-08-02: `scripts/gsd prompt programming-loop init --phase connector-guardrail-remediation-r1 --dry-run` failed with `scripts/gsd: unknown GSD command: programming-loop`; manual GSD fallback recorded.
- 2026-08-02: `scripts/gsd prompt execute-phase connector-guardrail-remediation-r1 --dry-run` generated prompt successfully.
- 2026-08-02 RED: worker ledger completeness validation failed before remediation ledger creation: `ledger incomplete: 12 missing disposition(s)` for audited #3532/#3535 paths.
- 2026-08-02 RED: target-aware guard validation failed before guard fixture/proof because #3581 command is not integrated yet: `go run ./cmd/connectorgen ownership --help` -> `connectorgen: unknown subcommand "ownership"` / exit status 2.
- 2026-08-02 GREEN: worker ledger/fixture completeness validation passed: `ledger complete: 29 audited path references found`; `fixture complete: 8 rejection path references found`.
- 2026-08-02 GREEN: connector validation passed for target/unrelated preserved defs: `go run ./cmd/connectorgen validate internal/connectors/defs/zendesk-support`, `.../google-ads`, and `.../gong` each reported `1 connector(s) checked, 0 findings`.
- 2026-08-02 EVIDENCE: temporary pre-#3535 Gong `cli_surface.json` revert probe failed with 19 `cli_surface_safety` findings, proving the current unrelated Gong metadata is valid safety state to preserve rather than delete.
- 2026-08-02 GREEN: `gofmt -w cmd internal` completed with no Go file diff.
- 2026-08-02 GREEN: `go test ./internal/connectors/boundary ./cmd/connectorgen` passed (`boundary` 54.444s, `cmd/connectorgen` 18.320s).
- 2026-08-02 GREEN: required `rg -n "zendesk-support|google-ads|gong|hubspot|bitbucket|xero|bahmni" docs/connectors website internal/connectors/defs` completed with 31,225 matches.
- 2026-08-02 GATE: first `go vet ./...` attempt failed with transient Go build-cache/std-package lookup errors; immediate retry passed with no output.
- 2026-08-02 GREEN: `go build ./cmd/pm` passed.
- 2026-08-02 GREEN: `make verify` passed, including `go test -timeout 20m ./...`, `pm docs validate --connectors-dir docs/connectors`, `golangci-lint`, `connectorgen validate internal/connectors/defs`, `connectorgen boundary . --json`, and Homebrew notification assertions.
