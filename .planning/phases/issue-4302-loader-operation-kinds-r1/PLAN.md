# Plan — Issue #4302 loader operation-kind registration

## Task Delivery Header

- Issue: Refs #4302 — fix(engine): register `rest_status` and `text_export` operation kinds in the bundle loader.
- Base branch: `main`.
- Merges into: `main`.
- Delivery: Pull request open against `main`, with the local engine and repository verification gates green and the API-reported base confirmed as `main`.
- Working branch: `fm/cli-loader-kind-registration-r1`.
- Task: Make the ordinary bundle loader accept well-formed `rest_status` and `text_export` declarations, while preserving closed validation for status-only response behavior and bounded CSV exports. Do not edit connector definitions or the already-working executors, pagination, or SCIM handling.
- Verification: Add loader-path synthetic-bundle tests first; run their red and green commands, the full engine package, `go vet`, build, generator/schema checks, `connector-boundary`, and the applicable repository GSD workflow check.

## Evidence Table

| Acceptance criterion | Evidence | Observable assertion or fake reason |
| --- | --- | --- |
| A well-formed `rest_status` definition loads through `Load` | fake | Docker Hub’s in-progress definition is intentionally out of scope and unavailable in this worktree. A `fstest.MapFS` fixture exercises the real decoder, JSON schema, block map, and semantic validator, then asserts the parsed operation kind and REST block. |
| A well-formed bounded `text_export` definition loads through `Load` | fake | Same real-loader fixture rationale; it asserts the parsed binary block retains its positive `max_bytes` cap. |
| Invalid status and export declarations fail closed before I/O | fake | Bundle validation has no I/O seam; synthetic declarations assert `Load` rejects a JSON body/output contract, non-HEAD status request, unbounded CSV export, and the wrong execution block. |

## Lifecycle and scope

- GSD commands resolved: `scripts/gsd prompt discuss-phase 4302 --auto`, `plan-phase 4302 --tdd --auto`, `execute-phase 4302 --interactive --auto`, `verify-work 4302 --auto`, and `code-review 4302 --auto`.
- Manual-GSD fallback: this firstmate direct-PR lane is single-worker and the canonical contract prohibits role spawning here. The generated prompts are executed inline, with this plan, ledger, verification checklist, and review report as the durable evidence.
- Required skills loaded: `golang-how-to`, `golang-design-patterns`, `golang-structs-interfaces`, `golang-error-handling`, `golang-security`, `golang-safety`, and `golang-testing`.
- This is a shared engine foundation issue, not a connector lane. It changes no `internal/connectors/defs/**` file and includes no connector-name branch.

## TDD slices

1. **Loader reachability (red → green).** Add an ordinary `Load` test containing one HEAD/status REST operation and one bounded CSV/binary operation. Red must fail because `expectedOperationBlock` does not map either known kind; green adds only the two correct block mappings and proves both parsed declarations survive load.
2. **Closed declaration validation (red → green).** Add named loader-path cases for a `rest_status` JSON/body declaration and non-HEAD method, unbounded `text_export`, and an invalid block pairing. Green is the existing semantic validator reached through the corrected mapping; retain exact error assertions so the loader fails before executor/I/O construction.
3. **Regression verification and review.** Run focused and package tests, generated/schema checks, build/lint/boundary gates, execute manual `verify-work`, and perform a focused security/correctness review of the changed loader map and table-driven tests.
