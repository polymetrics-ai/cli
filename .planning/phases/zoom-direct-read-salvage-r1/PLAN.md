# Zoom direct-read salvage — wave #4267

## Task Delivery Header

- Issue: Refs #4267 — feat(zoom): add reviewed direct-read salvage cohort
- Base branch: `fm/cli-zoom-parity-lane-r1`
- Merges into: `feat/4267-zoom-direct-reads` → `fm/cli-zoom-parity-lane-r1` → `main`
- Delivery: one stacked sub-PR after local gates and review. It is implemented-and-pending-certification because Zoom remains outside the firstmate-owned central certification scope; it must not integrate until that scope is added and the gate emits `PROCEED`.
- Task: selectively salvage the 70 reviewed Zoom `rest_read` declarations, matching command paths, endpoint ledger dispositions, and 52 sanitized direct-read fixtures from PR #3951. Do not copy write, reverse-ETL, binary, AI Services, shared-engine, or auth changes.
- Verification: red/green connector-local regression; real commandrunner preflight for every direct-read command; sanitized fixture inventory checks; `connectorgen validate`; `surface-sync --check`; generated sweep; runtime preflight; connector boundary; help parity; full `make verify` before push.

## Decisions

- This is selective file-content salvage, not a rebase, cherry-pick, or adoption of PR #3951. The source is limited to its reviewed Zoom `rest_read` records and direct-read fixtures; the parent remains based on post-foundation `main`.
- No authentication, engine, certification allowlist, or certification-status file may change. Certification is not claimed without accepted live proof.
- The four SCIM2 reads use the provider origin `/scim2/*` while Zoom's existing shared base URL defaults to `https://api.zoom.us/v2`. The former PR's unsupported operation-level `rest.base_url` is unavailable after foundation validation. Fixture preflight remains possible through the existing test base URL override; a default-correct live route requires a separately approved foundation capability and is recorded as a pending scope constraint, not patched here.
- The three existing ETL commands, streams, schemas, ledger inventory rows, and fixtures remain intact. No AI Services rows are added.
- Inline GSD fallback: prompts for `discuss-phase` and `plan-phase --tdd` were resolved through `scripts/gsd prompt`; the canonical contract forbids GSD role delegation.

## TDD slices

1. **Red:** a connector-local test asserts that 70 direct-read operations and commands, plus 52 sanitized fixture files, exist and all direct-read commands pass `commandrunner.Preflight`; it fails against the three-ETL-command baseline.
2. **Green:** mechanically salvage only the reviewed read declarations, command rows, endpoint dispositions, and fixtures from PR #3951, adapting unsupported per-operation transport fields out of the JSON schema.
3. **Refactor:** run derivation checks and make the evidence explicit: rows are implemented fixture proof but pending certification, including the SCIM2 default-origin constraint.

## CLI parity

- The wave adds declarative Zoom commands. Verify bare `pm connectors`, `pm zoom --help`, and representative group help. Generated surfaces and docs are checked for drift.
- No handwritten docs/website artifact is edited unless the repository generator/check reports a required drift; the added command summaries are contained in `cli_surface.json`.

## Required skills

`golang-how-to`, `golang-cli`, `golang-testing`, `golang-error-handling`, `golang-security`, `golang-safety`, `golang-design-patterns`, `golang-structs-interfaces`, and `golang-documentation`.
