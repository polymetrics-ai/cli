# Zoom ETL certification parity — wave #4266

## Task Delivery Header

- Issue: Refs #4266 — test(zoom): retain ETL under generated certification parity
- Base branch: `fm/cli-zoom-parity-lane-r1`
- Merges into: `test/4266-zoom-etl-certification-parity` → `fm/cli-zoom-parity-lane-r1` → `main`
- Delivery: A sub-PR to the parent after local gates and required CI/review coverage. Zoom is deliberately outside the centrally owned certification scope, so this wave records fixture proof and pending certification; it must not be integrated until firstmate adds scope and the gate returns `PROCEED`. Parent merge remains captain-gated.
- Working branch: `test/4266-zoom-etl-certification-parity`
- Task: Generate Zoom's existing ETL/capability certification sweep and add connector-local regression evidence without changing the three ETL commands, streams, schemas, fixtures, or provider ledger.
- Verification: targeted Zoom definition/conformance and connectorgen tests; `connectorgen validate`; `surface-sync --check`; `certification-sweep --check`; `make connector-runtime-preflight`; `make connector-boundary`; docs checks; and `make verify` before pushing.

## Decisions

- The source report and parent issue fix scope: preserve existing ETL behavior exactly.
- The sweep is generated inventory, not a live-certification claim. No credential or provider call is used.
- Captain decision (2026-08-19): do not modify authentication, engine, `cmd/connectorgen/certificationallowlist.go`, or certification `status.json`. `capability/zoom/missing` is an expected external scope gate owned centrally by firstmate, not a defect in this wave. Every cell remains implemented-and-pending-certification until fixture and live proof can be accepted through that gate.
- Inline GSD fallback: `scripts/gsd prompt discuss-phase zoom-etl-certification-parity-r1` and `plan-phase zoom-etl-certification-parity-r1 --tdd` were generated and executed manually because the canonical contract forbids GSD-role delegation.

## TDD slices

1. **Red:** connector-local test requires a committed generated sweep with exactly three ETL rows and one capability-read row; it fails while the file is absent.
2. **Green:** generate Zoom's sweep with the repository command; do not edit generated JSON by hand.
3. **Refactor:** run derivation/check commands and confirm no non-Zoom production path changed.

## CLI parity

- Existing command paths are unchanged; runtime help, bare `pm connectors`, and Zoom connector help are verified as unchanged/not applicable to an added command.
- Generated connector docs/catalog are checked for drift; no hand-authored CLI documentation change is expected.

## Required skills

`golang-how-to`, `golang-cli`, `golang-testing`, `golang-error-handling`, `golang-security`, `golang-safety`, `golang-design-patterns`, `golang-structs-interfaces`, and `golang-documentation`.
