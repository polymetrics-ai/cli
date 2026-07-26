# TDD ledger — Bahmni final verified parity corrections

## Red / baseline at start

Exact starting head: `4f5822533d880fbe294a4d33969c11b12ed2675e`.

Independent final audit found green internal validation but material operation-parity failures:

- Appointments stream sends unsupported `patientUuid`, accepts incorrectly documented date shape, and may return an unscoped appointment book.
- Patients list can omit `q` and fails live on unqualified get-all.
- Lab results omit required `concept` list.
- Diagnoses use nonexistent `/bahmnicore/diagnosis/getDiagnoses` and still declare nullable `existingObs` as primary key.
- Bahmni patient profile GET-by-UUID and POST patient-search route are unsupported by pinned STANDARD modules.
- Diagnosis and bulk-observation write actions point at unsupported POST endpoints.
- PHI-sensitive connector fields are documented but not enforced by runtime tests.
- No Bahmni-specific tests cover exact request shapes, invalid/missing filters, PHI redaction, root arrays, two-page pagination, unsupported route rejection, or retained writes.

## Planned red tests

Before/alongside fixes, add tests that fail against `4f582253` and pass after correction:

1. Bahmni bundle contract tests for exact stream request method/path/query:
   - `patients` requires `patient_query` and includes `q`.
   - `lab_results` requires `patient_uuid` and `lab_result_concepts`/concept list.
   - `diagnoses` uses `/ws/rest/v1/bahmnicore/diagnosis/search`.
   - `appointments` does not send unsupported `patientUuid`; date shape is exact `yyyy-MM-dd'T'HH:mm:ss.SSS` or command is blocked when absent.
2. Unsupported-route tests asserting removed/blocked patient-profile GET and unsupported patient-search POST cannot be advertised as implemented.
3. Write capability tests proving unsupported diagnosis/bulk-observation writes are removed/blocked and metadata `write` reflects retained proof.
4. Diagnosis schema test proving nullable `existingObs` is not an `x-primary-key`.
5. PHI redaction tests proving declared sensitive fields are redacted from direct operation output and reverse-plan previews.
6. Fixture replay tests for root arrays and at least one two-page offset stream.

## Green criteria

- Bundle validator reports 0 findings.
- Bahmni-specific tests pass and fail if audited bad routes/queries are reintroduced.
- Existing conformance passes for the corrected connector.
- Focused CLI/help/docs/catalog and website generation are in sync without unrelated 446-doc churn.
- Full local gates pass.
- Fresh exact-head independent audit is requested only after implementation and validation.

## Results

Review-fix slice green for Bahmni scope text, generated artifact parity, and write-preview path redaction. Broader full-phase gates remain pending.

## Local preview checkpoint addendum - 2026-07-26

| Slice | Test first? | Expected evidence | Status |
|---|---|---|---|
| pm-bahmni local preview | Yes | Build `./pm-bahmni`, symlink command path, SHA-256/size/source identity, credential-free smoke tests, safe synthetic-lab bounded reads | pending until coherent buildable checkpoint |

GSD fallback note: `scripts/gsd prompt programming-loop ...` unavailable (`unknown GSD command: programming-loop`); manual red/green loop used for following slices.

## Typed write directive addendum - 2026-07-26

Captain clarified that scalar-only live-write checks are partial evidence only. Completion requires every retained write to be executable through strict typed CLI flags and to receive live synthetic proof. Add/keep red tests that fail if:

- any Bahmni write command exposes generic `json`, raw-body, method/path, or arbitrary nested-object escape hatches;
- structured write inputs remain as string-valued `person`, `identifiers`, `notes`, or equivalent JSON-in-string flags;
- a retained write lacks explicit schema-bound flag mappings/builders for required nested objects/arrays;
- a retained write is advertised as implemented but cannot pass plan -> preview -> approval -> execute against the pinned synthetic lab.

If an operation cannot be typed/proven against the pinned STANDARD alpha stack, mark it unavailable or remove it with source/live evidence instead of leaving an unusable advertised write.

## PHI protection addendum - 2026-07-26

Captain clarified that default PHI protection is a production-readiness gate, not a documentation caveat. Add/keep red tests proving normal Bahmni read/help/error/preview paths either emit opaque references or apply explicit command-level `redact_fields` for clinical identifiers and values. Required coverage: patient, encounter, observation, diagnosis, lab, appointment, note preview, and error output. Raw clinical content may flow only through trusted typed execution with deliberate authorization. If this cannot be completed here, keep release stage alpha and record PHI protection as a production blocker.

## Live synthetic write verification addendum - 2026-07-26

Captain authorized live write proof after current typed/schema tests pass and `pm-bahmni` is rebuilt. Red tests/verification must fail if any retained write lacks typed CLI plan -> preview -> explicit approval -> execute proof against a unique disposable `SYN-CONN-*` record, or if a failed write remains advertised without source/live evidence. Evidence must be status/count/opaque IDs only; no credentials/PHI.

## Review-fix slice - 2026-07-26

Baseline checks at `4e89af9ea5436088f5cef8e9f14e6eee0696b290` reproduce the prior findings:

- `docs/connectors/catalog/all-connectors.json` and `website/data/connectors.generated.json` still list `create_drug_order`, `create_diagnosis`, `create_observations_bulk`, and `upload_patient_document`.
- `internal/connectors/defs/bahmni/spec.json` and `api_surface.json` still mention the old mixed-stack support claim.
- `internal/connectors/defs/bahmni/cli_surface.json` still describes appointment reads as `appointment_date and/or patient_uuid`.
- `internal/connectors/defs/bahmni/operations.json` and command help still call patient search a POST/body read.
- `engine.DryRunWrite` resolves path fields from raw records in preview warnings.

Planned red coverage: extend engine write-preview tests so configured write-action redaction fields scrub resolved path identifiers, and use the existing Bahmni contract tests plus generated-data checks to keep the frozen scope truthful.

## Post-merge corrective slice - 2026-07-26

GSD fallback remains in effect for `scripts/gsd prompt programming-loop ...` (`unknown GSD command: programming-loop`). Added failing regression `TestWriteErrorRedactsConfiguredRecordFieldsInHTTPPathAndBody` after PR #533 merged; red state proved execute-time `*connsdk.HTTPError` display leaked encoded clinical path identifiers. Green target: preserve `errors.As`/error-map behavior while redacting configured write-action record literals from operator-visible write errors.
