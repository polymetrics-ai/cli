# GitHub mutation slice 3 — writes-c

## Task Delivery Header

- Issue: Refs #4015 — Production MVP certification program.
- Base branch: `integration/4015-mvp-flat-r1`.
- Merges into: `integration/4015-mvp-flat-r1 → main`.
- Delivery: A PR from `fm/cli-mut-slice3-writes-c` is open against the stated base with committed, schema-v2 live-certification evidence.
- Working branch: `fm/cli-mut-slice3-writes-c`.
- Task: Attempt all 146 assigned GitHub mutation commands serially. Each certified command must execute plan → preview → approval-token run, prove its provider state, use a direct provider DELETE for cleanup, and independently prove absence. Effects are confined to `Polymetrics-Cert`, `Polymetrics-Cert/pm-cert-3993-20260810-wz0fru`, and `polymetrics-ai-certification`; the captain's 2026-08-18 authorization governs contained writes.
- Verification: `go run ./cmd/connectorgen certification-matrix --check`, targeted `cmd/connectorgen` tests, `git diff --check`, repository verification sub-gates, and GitHub API base read-back after opening the PR.

## GSD and skills

- Inline/manual GSD fallback: this non-Pi worker generated and followed `scripts/gsd prompt discuss-phase github-mutation-slice3-writes-c`, `plan-phase ... --tdd`, `execute-phase`, `verify-work`, and `code-review`; compatible isolated GSD runtime workers are unavailable and the repository contract forbids role spawning.
- Required skills loaded: `golang-how-to`, `golang-cli`, `golang-testing`, `golang-error-handling`, `golang-security`, `golang-safety`, `golang-design-patterns`, and `golang-structs-interfaces`.
- CLI help/manual/website parity: not applicable; this is evidence-only live certification and changes no command, flag, help text, docs, or generated public command surface.

## Delivery rulings

- Firstmate decision: author schema-v2 evidence directly from each captured live lifecycle; do not wait for an importer.
- Firstmate decision: where GitHub cannot delete an object, restore the strongest provider-supported benign state, independently verify that state, and record `contained_closed`; the disposable fixture repository is the cleanup container. Never label a surviving object `verified_absent`.

## TDD ledger

- Red: before the live run, no slice-3 write evidence records exist on this branch and `certification-matrix --check` cannot count this slice as certified.
- Green: each accepted schema-v2 evidence record contains sanitized plan/preview/run/readback/direct-cleanup/readback exchanges and the matrix check accepts it.
- Refactor: retain only validated records; do not regenerate shared matrix artifacts while parallel lanes are active.

## Evidence Table

| Acceptance criterion | Evidence | Observable assertion or fake reason |
| --- | --- | --- |
| A certified mutation changes provider state | live | Independent read-back sees the run's `pm-cert-` value; a pre-mutation or unrelated value fails the assertion. |
| Cleanup is real | live | A direct GitHub DELETE followed by a provider read-back returns 404 or excludes the identifier; CLI exit status is not used. |
| Published evidence remains safe and valid | live | Schema-v2 proof fingerprints credentials and protocol values; `certification-matrix --check` validates every retained record. |

## Serial execution plan

1. Work the supplied JSON order one command at a time, resolving fixture identifiers from provider reads and deriving an explicit assertion where none is declared.
2. Use `pm` plan/preview/run with the plan-minted token supplied through stdin. Never print or persist credentials or tokens.
3. Direct-delete each disposable resource using GitHub's provider API, read it back independently, write a schema-v2 evidence record immediately, and validate it.
4. After one retry, record the specific honest bucket and continue. Stop only for the four captain escapes.
