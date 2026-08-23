# TDD ledger — issue #4288

| Slice | Red | Green | Refactor / observable proof |
| --- | --- | --- | --- |
| Certification scope admission | `go run ./cmd/connectorgen certification-matrix --connector jira --check`, `... asana ...`, and `... notion ...` each exit 2 with `is not certification-allowlisted`; no evidence can be generated. | Add the three scope entries without changing their definition bundles, regenerate through `certification-matrix --all`, and rerun the scoped checks. | The generator's scope/status projections include exactly the original entries plus Jira, Asana, and Notion; `certification-matrix --check` validates the committed generated state. |
| Capability-cell classification | Before live runs, each new matrix has zero accepted records and the candidate/sweep outputs must not be interpreted as a certification result. | For every cell, the harness emits either an immediately accepted proof record or a concrete sanitized non-pass receipt. | A successful process exit alone is rejected as proof; the matrix accepts only records containing observed, assertion-backed protocol evidence. |
| Provider live read | Before a valid credential and in-scope fixture, a provider request cannot create accepted evidence. | A successful live provider operation writes an accepted normal-format record, and the immediate matrix check passes. | Each proof is count/fingerprint-only; raw provider values and credentials remain absent from the committed diff. |
| Live write containment | A command lacking a scratch resource or reverse-ETL lifecycle approval is not run as a write proof. | A permitted scratch-owned mutation is independently read back after cleanup; otherwise it stays explicitly uncertified. | No write to user-owned or pre-existing provider data is treated as evidence. |

## Red execution record

- Red: all three scoped matrix checks failed as expected on the initial branch because the generator
  rejects unallowlisted connectors. The exact safe error was
  `connector "<target>" is not certification-allowlisted` (exit status 2).
- Green: pending the narrow allowlist addition and normal generated-artifact refresh.

## CLI/docs parity

Not applicable: no `pm` command, flag, help topic, output, documentation, website page, or connector
definition command surface changes in this slice. `connectorgen`'s existing generated scope artifacts
are verified instead.
