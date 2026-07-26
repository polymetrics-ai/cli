# Verification — Bahmni final verified parity corrections

## Custody before edits

- Branch: `feat/bahmni-docker-connector`.
- Local head: `4f5822533d880fbe294a4d33969c11b12ed2675e`.
- no-mistakes run `01KYDK7BF4TMSYKJHA8GA8BF6P` remains as a post-check CI monitor.
- `no-mistakes axi sync --check` reported local, remote, and pipeline heads equal, relation `equal`, safety `already_synchronized`, PR state `open`.
- Worktree clean before this phase's planning artifacts.

## Required local verification

Pending implementation:

- [ ] `go test ./internal/connectors/engine -run 'Bahmni|Redact|Sensitive|Write|Request' -count=1`
- [ ] `go test ./internal/connectors/conformance -run '^TestConformance$/^bahmni$' -count=1 -v`
- [ ] `go test ./cmd/connectorgen ./internal/connectors/bundleregistry ./internal/connectors/conformance ./internal/cli -count=1`
- [ ] `go run ./cmd/connectorgen validate internal/connectors/defs`
- [ ] `go vet ./...`
- [ ] `go test -timeout 20m ./...`
- [ ] `go build ./cmd/pm`
- [ ] `./pm docs validate --connectors-dir docs/connectors`
- [ ] `website: npm run gen:website-data` in a way that keeps generated outputs in sync.
- [ ] Check no unrelated 446 connector MANUAL/SKILL churn is present.

## Required non-local validation / audit gates

- [ ] Codex-only validation/review against exact candidate head.
- [ ] Fresh independent exact-head audit before any “full parity” claim.
- [ ] Live STANDARD synthetic read verification and disposable write verification only if captain explicitly provides/authorizes synthetic lab credentials and write execution.

## Merge boundary

Do not merge PR #533.
Do not claim full functionality/parity until all required validation and fresh exact-head audit pass.

## Local preview checkpoint addendum - 2026-07-26

Pending at next coherent buildable correction checkpoint:

```bash
go build -o ./pm-bahmni ./cmd/pm
ln -sfn "$(pwd)/pm-bahmni" /Users/karthiksivadas/.local/bin/pm-bahmni
command -v pm-bahmni
./pm-bahmni connectors inspect bahmni --json
pm-bahmni help bahmni
shasum -a 256 ./pm-bahmni
wc -c < ./pm-bahmni
```

Also run credential-free fixture/conformance smoke and safe synthetic-lab bounded reads without printing secrets.

## Typed write verification addendum - 2026-07-26

- [ ] Static check: Bahmni implemented write CLI commands expose no generic `json`, raw-body, method/path, or arbitrary nested-object flags.
- [ ] Static check: structured required write fields are represented by explicit named scalar/enum/file flags mapped to schema-bound target paths/builders.
- [ ] Unit/conformance checks prove schema-bound nested object/array builders and rejection of unsupported structured string/generic JSON flags.
- [ ] Live synthetic proof: for every retained implemented write, run plan -> preview -> approval -> execute against disposable synthetic lab records only, without printing/persisting secret values.
- [ ] For any write not live-proven, update definitions/API surface to mark it unavailable or remove it with source/live evidence before claiming completion.

## PHI protection verification addendum - 2026-07-26

- [ ] Normal Bahmni read command output applies explicit `redact_fields` for patient, encounter, observation, diagnosis, lab, and appointment clinical identifiers/values.
- [ ] Bahmni write-preview output redacts note payloads and other clinical write fields.
- [ ] Bahmni error output is redacted for command-level clinical identifiers/values before reaching normal CLI output.
- [ ] Help/inspect output does not print clinical record values and does not claim PHI caveats as production readiness.
- [ ] If any PHI protection test remains incomplete, connector stays alpha and PHI protection is reported as a production blocker.

## Live synthetic write verification authorization - 2026-07-26

Captain authorized parallel live verification of every retained Bahmni write against the existing loopback-only Podman lab after typed/schema tests pass and `pm-bahmni` is rebuilt. Constraints: unique `SYN-CONN-*` disposable identifiers, typed CLI plan -> preview -> explicit approval -> execute only, no raw JSON/method/path write escape hatches, no credential/PHI printing or persistence, no Karthik/Rohit records, no reseed/restart/container mutation, no lane collisions, safe opaque evidence/status only. Independent chains may run in parallel; dependencies inside each lane remain serial. Any failed operation must stop, be fixed or marked unavailable with source/live evidence, then rerun before readiness is claimed.

## Review-fix slice - 2026-07-26

Required focused verification for this slice:

- [x] `go test ./internal/connectors/engine ./internal/connectors/conformance -run 'Test(DryRunWritePreviewResolvedPathRedactsConfiguredRecordFields|DryRunWritePreviewResolvedPathRedactionCopiesNestedRecord|BahmniVersionPinnedReadContracts|BahmniVersionPinnedDirectOperationContracts|BahmniVersionPinnedWriteContracts|BahmniFrozenScopeTextContracts)$' -count=1`
- [x] Bahmni-only generated-data assertion confirms no stale write claims, no non-STANDARD support claim, appointment date-only help, and GET patient-search wording in connector catalog and website data.
- [x] `go run ./cmd/pm docs generate --dir docs/cli --connectors-dir docs/connectors`
- [x] `npm --prefix website run gen:website-data`
- [x] Unrelated generated connector MANUAL/SKILL churn restored out of the worktree.

### Live lane attempt 1 safe summary

- Patient lane: `create_patient`, `update_patient` completed through typed CLI plan -> preview -> approval -> execute.
- Appointment lane: `create_patient`, `create_appointment`, `update_appointment_status` completed; `update_appointment_provider_response` stopped with upstream provider-self authorization failure. This is a test-credential limitation, not unsupported-route evidence; provider-linked synthetic credentials are pending.
- Clinical chain lane: `create_patient`, `create_visit` completed; `create_encounter` stopped with upstream HTTP 400 on a future-dated encounter payload. Next rerun uses a past disposable timestamp before proceeding to observation, diagnosis, note, and order writes.
- No credentials, PHI, or record identifiers were printed or persisted in this summary.

### Live lane attempt 2 safe summary

- Observation chain: `create_patient`, `create_visit`, `create_encounter`, and `create_observation` completed after switching to past disposable timestamps.
- Order chain: `create_patient`, `create_visit`, `create_encounter`, and `create_lab_order` completed; `create_drug_order` stopped with upstream HTTP 400 on the typed order payload and needs source/live diagnosis before rerun or availability change.
- Note/diagnosis chain: `create_patient`, `create_visit`, and `create_encounter` completed; `create_note` stopped with upstream HTTP 500/HTML on `POST /ws/rest/v1/notes` and needs source/live diagnosis before rerun or availability change. `create_patient_diagnosis` remains untested because the dependency chain stopped before it.
- Appointment reschedule chain: `create_patient` and `create_appointment` completed; `reschedule_appointment` did not execute because the CLI surface lacked a typed `--end-date-time` flag. This is a connector mapping defect to fix and rerun.
- Provider response remains supported-upstream-but-not-yet-live-proven with current admin credential; do not mark unavailable solely for this test-credential limitation.

### Live lane attempt 3 safe summary

- Provider-authenticated appointment chain: bounded provider/service check failed because the synthetic provider is not yet linked to an appointment service; `update_appointment_provider_response` remains deferred until lab-side service linkage is ready. This is not unsupported-route evidence.
- Clinical chain: `create_patient`, `create_visit`, `create_encounter`, and `create_note` completed through typed CLI plan -> preview -> approval -> execute with past disposable timestamps.
- Clinical failures to diagnose/rerun: `create_patient_diagnosis` and `create_drug_order` failed through the typed path and remain unproven.
- Omitted operations are not proof: `create_observation`, `create_lab_order`, `reschedule_appointment`, and `update_appointment_provider_response` require explicit rerun/accounting before readiness is claimed.
- No credentials, PHI, or record identifiers were printed or persisted in this summary.

### Live drug-order correction checkpoint

- Current `create_drug_order` evidence remains failing/unproven: typed plan -> preview reached execute, upstream returned HTTP 400 classified as payload validation.
- The preceding local schema-only correction is not proof; final proof still requires a successful typed live execute.
- Next diagnostic is restricted to operation, stage, HTTP status, upstream machine code/class, safe field names, and generated request key/type shape only; no response bodies, messages, stack traces, identifiers, credentials, or PHI.
- Constrained drug-order diagnostic captured generated request key/type shape vs pinned successful create-test shape; missing typed fields were corrected in the connector and the rerun used the full typed shape.
- `create_drug_order` remains failing/unproven after the corrected-shape rerun: stage `execute`, HTTP `400`, safe fields `action,drug`, upstream code unavailable from the wrapped response. Do not count schema validation or the corrected fixture as live proof.
- `create_patient_diagnosis` was rerun after the lab identity received the exact temporary `Edit Diagnoses` grant and completed through typed plan -> preview -> approval -> execute.

### Live lane attempt 4 safe summary

- Independent clinical reruns: `create_observation` PASS, `create_lab_order` PASS, `create_patient_diagnosis` PASS.
- Provider-authenticated appointment rerun: `create_appointment` PASS, `update_appointment_provider_response` PASS, `reschedule_appointment` FAIL.
- Still unproven and not counted as proof: `create_drug_order` and `reschedule_appointment`.
- No credentials, PHI, response bodies, stack traces, record identifiers, or secret values were printed or persisted in this summary.

### Direct counterfactual diagnostic checkpoint

- Diagnostic-only direct counterfactuals were run with the same disposable reference style and exact in-memory generated request bodies; these direct calls are not connector capability surfaces and do not count as proof.
- `create_drug_order`: direct `POST /ws/rest/v1/order` returned HTTP `400`, upstream code `webservices.rest.error.invalid.submission`, safe fields `none`. Typed flow also failed at execute with HTTP `400`, safe fields `action,drug`. Result: removed from retained executable writes and marked `unsupported_local` in CLI/docs until a concrete upstream-required field or lab capability is proven.
- `reschedule_appointment`: direct `POST /ws/rest/v1/appointment/{uuid}/reschedule` returned HTTP `400`, upstream code `org.hibernate.exception.internal.SQLExceptionTypeDelegate:59`, safe fields `action,dose,type,uuid`; typed flow failed at execute with HTTP `400`, safe field `uuid`, while generated key/type shape matched the pinned singular `AppointmentRequest` controller contract. Result: removed from retained executable writes and marked `unsupported_local` in CLI/API/docs until a concrete upstream-required field or route capability is proven.
- Retained executable write matrix after reconciliation:
  - `create_patient`: PASS live synthetic proof.
  - `update_patient`: PASS live synthetic proof.
  - `create_visit`: PASS live synthetic proof.
  - `create_encounter`: PASS live synthetic proof with past timestamps.
  - `create_observation`: PASS live synthetic proof.
  - `create_lab_order`: PASS live synthetic proof.
  - `create_patient_diagnosis`: PASS live synthetic proof after exact temporary `Edit Diagnoses` grant.
  - `create_appointment`: PASS live synthetic proof.
  - `update_appointment_status`: PASS live synthetic proof.
  - `update_appointment_provider_response`: PASS live synthetic proof after provider-linked temporary credential/service mapping.
  - `create_note`: PASS live synthetic proof.
- Blocked/unretained write-like commands: `create_drug_order`, `reschedule_appointment`, `upload_visit_document`, document upload, top-level bulk observations, and other destructive/admin routes in `api_surface.json`.
- No operation is omitted from the retained-write matrix; unretained operations are explicitly not proof.

## Post-merge corrective verification - 2026-07-26

- `scripts/gsd doctor`: pass.
- `scripts/gsd prompt programming-loop init --phase issue-535-bahmni-post-merge-corrections --dry-run`: unavailable (`unknown GSD command: programming-loop`); manual GSD fallback recorded.
- Red: `go test ./internal/connectors/engine -run TestWriteErrorRedactsConfiguredRecordFieldsInHTTPPathAndBody -count=1` failed because the write error contained the encoded clinical path identifier.
- Green: `go test ./internal/connectors/engine -run 'TestDryRunWritePreviewResolvedPath|TestWriteErrorRedactsConfiguredRecordFieldsInHTTPPathAndBody' -count=1` passed after write-error literal redaction preserved `errors.As` reachability.
