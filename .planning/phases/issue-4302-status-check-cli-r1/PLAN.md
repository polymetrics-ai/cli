# Plan — PR #4308 status-check result preservation

## Task Delivery Header

- Issue: Refs #4302 — `fix(engine): register rest_status and text_export operation kinds in the bundle loader`; PR #4308 remediation adds the missing shared CLI result-preservation boundary.
- Base branch: `main`.
- Merges into: `main`.
- Delivery: Existing PR #4308 remains open against `main`; its exact updated head is pushed, its current-head checks are terminal green or an explicit CI HOLD is recorded, and Firstmate alone retains merge authority.
- Working branch: `fm/cli-loader-kind-registration-r1`.
- Task: Preserve every declared `StatusCheck` result through the installed CLI in dedicated JSON and deterministic human representations, while retaining the proven binary-download behavior and the loader's pre-I/O closed validation.
- Verification: Record red/green focused CLI tests, run changed-package tests, formatting, vet, build, lint, applicable repository gates, no-mistakes validation, and a source-locked credential-free GitHub Pages HEAD/GET proof through the installed binary.

## Evidence Table

| Acceptance criterion | Evidence | Observable assertion or fake reason |
| --- | --- | --- |
| Status result uses a dedicated stable JSON envelope | fake | A deterministic CLI unit fixture injects the typed result at the real output boundary because no shipped connector declaration currently exposes this new operation kind. It asserts exact JSON identity and rejects the empty ETL-read fallback. |
| Human status output is non-empty and complete | fake | The same output-boundary fixture asserts operation, method, path, non-200 status, and zero body bytes. It is deterministic and does not need a network service. |
| Result classification remains closed and binary behavior remains intact | fake | Table-driven output-boundary cases assert the StatusCheck branch wins only for its typed field, while existing binary-download tests continue to cover exact manifest fields. |
| Malformed declaration/input fails before I/O | fake | Existing loader-path tests have no I/O seam; they assert malformed `rest_status` declarations are rejected by `engine.Load` before executor construction. |
| Ordinary loader and installed CLI reach the real provider without data loss | live | A source-locked temporary declaration is injected only for the local proof, then removed. Independent readback compares HEAD metadata and GET bytes/SHA-256 with GitHub's immutable source. |

## Lifecycle and skills

- GSD commands resolved: `scripts/gsd prompt discuss-phase 4302 --auto`, `plan-phase 4302 --tdd --auto`, `execute-phase 4302 --interactive --auto`, `verify-work 4302 --auto`, and `code-review 4302 --auto`.
- Manual-GSD fallback: the Firstmate direct-PR lane is single-worker and prohibits GSD role spawning. The generated workflow prompts are executed inline with durable context, plan, TDD ledger, verification, summary, and review records.
- Required skills: `golang-how-to`, `golang-cli`, `golang-design-patterns`, `golang-structs-interfaces`, `golang-error-handling`, `golang-security`, `golang-safety`, `golang-testing`, `golang-lint`, `no-mistakes`, `gsd-discuss-phase`, `gsd-plan-phase`, `gsd-execute-phase`, `gsd-verify-work`, and `gsd-code-review`.
- CLI parity: no command/flag/help text or public connector declaration changes. `docs/cli/**`, `website/**`, generated manuals, namespace help, and completions are not applicable. JSON and human behavior receive focused CLI tests and are documented in the PR body and ship report.

## TDD slices

1. **Classify and preserve status result (red → green).** Add focused JSON and human output tests at the installed CLI result boundary. Red proves the current fall-through produces `ConnectorCommandRead` / empty records and silent human success. Green adds one generic typed status branch and a stable typed envelope.
2. **Edge contracts (red → green).** Cover a non-200 status and zero-byte HEAD result, exact field presence, and a guard that StatusCheck cannot be serialized as an empty ETL result. Preserve the existing binary-download assertions unchanged.
3. **Source-locked real proof and review.** Build the installed binary; inject no persistent declaration or credential material; compare the source-locked GitHub Pages HEAD status and export bytes/hash against the exact upstream commit; remove all proof-only state, verify the diff, run local gates, no-mistakes, and focused review.
