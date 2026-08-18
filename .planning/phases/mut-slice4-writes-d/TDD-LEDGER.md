# TDD ledger — GitHub mutation certification slice 4 writes-d

## Red

- Before recording a `certified` outcome, each command's acceptance predicate begins false: the provider read-back does not yet show the new `pm-cert-` tagged state.
- Cleanup is independently red until a direct provider DELETE has completed and a separate read returns 404 or omits the tagged object. A successful `pm` delete is explicitly insufficient because #4221 reproduces a false success.

## Green

- A command reaches green only after `pm ... plan`, `preview`, and approved `run` produce a provider-visible change that satisfies the record's declared or `agent_derived` predicate.
- Cleanup reaches green only when the independent direct-provider read-back proves absence.
- Evidence is retained only after `go run ./cmd/connectorgen certification-matrix --check` succeeds; invalid evidence is deleted rather than repaired into a fabricated exchange.

## Resolved surface finding — 2026-08-18

- `branches apps create` was planned successfully as `rplan_46dedc2add0a381b` with the contained `cert-classic` credential and a fixture-repository branch path. No provider write was dispatched.
- The stdin-only attempt was made twice and rejected before dispatch because connector commands use `--approve <token>` rather than the ETL transport's stdin-token path.
- Fleet ruling confirms that the locally minted, single-use, time-bounded approval token is not a GitHub credential and is authorized on argv. This is therefore a one-line surface finding, withdrawn as a product defect; mint a fresh plan/token rather than reuse either prior plan.
- Raw provider control: authenticated `GET /repos/Polymetrics-Cert/pm-cert-3993-20260810-wz0fru` returned HTTP 200.

## Refactor

- No product code, generated shared certification artifacts, command metadata, or CLI/help/docs surface is changed in this lane. The only committed changes are planning/verification records and valid proof-bearing evidence.
