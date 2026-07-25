# Issue #516 Bahmni official-operation parity matrix

## Scope and sources

Connector id: `bahmni`.

Official sources reviewed for this follow-up:

- Bahmni product site: https://www.bahmni.org/
- Bahmni deployment repository: https://github.com/Bahmni/bahmni-docker
- Bahmni core backend: https://github.com/Bahmni/bahmni-core
- Bahmni/OpenMRS appointments backend: https://github.com/Bahmni/openmrs-module-appointments
- OpenMRS REST module docs/API: https://rest.openmrs.org/ and https://github.com/openmrs/openmrs-module-webservices.rest
- OpenMRS FHIR2 module: https://github.com/openmrs/openmrs-module-fhir2

The old Bahmni Confluence API pages linked from search results returned page-not-found/cookie pages
when fetched non-interactively, so source-code controller annotations and current OpenMRS REST/FHIR
references are the auditable source of truth for endpoint discovery.

## Issue hierarchy

- #516 — Bahmni connector CLI feature parity parent roadmap
  - #517 — Bahmni: CLI surface metadata (CLI parity)
  - #518 — Bahmni: help renderer (CLI parity)
  - #519 — Bahmni: stream runner (CLI parity)
  - #520 — Bahmni: API surface inventory + exclusion ledger (CLI parity)
  - #521 — Bahmni: direct read (CLI parity)
  - #522 — Bahmni: bounded binary and blocked-operation policy
  - #523 — Bahmni: typed reverse-ETL writes
  - #524 — Bahmni: typed POST read-query operation execution
  - #525 — Bahmni: schema-gated top-level JSON array request bodies
  - #526 — Bahmni: bounded typed multipart upload support

## Operation family matrix

| Official operation family | Official evidence | Owning issue | Connector mapping | Validation evidence |
| --- | --- | --- | --- | --- |
| OpenMRS patients | `GET/POST /ws/rest/v1/patient`, `GET/POST/DELETE /ws/rest/v1/patient/{uuid}` from OpenMRS REST resources | #519, #521, #523, #520 | Stream `patients`; direct read `patient get`; writes `create_patient`, `update_patient`; destructive delete/purge is blocked in `api_surface.json` | `connectorgen validate`; `TestConformance/bahmni` stream fixture; docs/CLI inspection |
| OpenMRS encounters | `GET/POST /ws/rest/v1/encounter`; Bahmni-core `bahmniencounter` find/get/delete controllers | #519, #521, #523, #520 | Stream `encounters`; direct read `encounter get`; write `create_encounter`; Bahmni-core find/delete endpoints recorded as non-generic/excluded workflow surface in `api_surface.json` | `connectorgen validate`; `TestConformance/bahmni` encounter fixture |
| OpenMRS observations and Bahmni lab observations | `GET/POST /ws/rest/v1/obs`; `GET/POST /ws/rest/v1/bahmnicore/observations`; `flowSheet` display controller | #519, #521, #523, #525, #520 | Streams `observations`, `lab_results`; direct read `fhir observation-read`; writes `create_observation`, `create_observations_bulk`; display/flowsheet helper excluded as UI/reporting surface | `connectorgen validate`; `TestConformance/bahmni`; root-array pagination disabled for `lab_results` |
| Visits | `GET /ws/rest/v1/visit`, `GET /ws/rest/v1/visit/{uuid}`, Bahmni-core visit summary/endVisit controllers | #519, #521, #520 | Stream `visits`; direct read `visit get`; endVisit/ADT workflow writes are blocked/excluded in `api_surface.json` pending dedicated typed workflow design | `connectorgen validate`; direct-read CLI surface inspection |
| Concepts and terminology | OpenMRS `concept` resource; Bahmni `GET /ws/rest/v1/bahmni/terminologies/concepts` | #519, #521, #520 | Stream `concepts`; direct read `concept get`; terminology helper covered by concepts/reference family and recorded in parity audit | `connectorgen validate`; `TestConformance/bahmni` concept fixture |
| Locations | OpenMRS `location`; Bahmni visit/facility-location helpers | #519, #521, #520 | Stream `locations`; direct read `location get`; visit/facility-location helpers are reference/display helpers excluded from generic direct reads | `connectorgen validate`; `TestConformance/bahmni` location fixture |
| Providers | OpenMRS `provider`; authenticated provider list | #519, #521, #520 | Stream `providers`; direct read `provider get`; connector check uses authenticated `GET /ws/rest/v1/provider?v=default&limit=1` instead of `/session` | `connectorgen validate`; `fixtures/check.json`; `TestConformance/bahmni` check + provider fixture |
| Orders, drug orders, lab orders | OpenMRS `order`; Bahmni-core `drugOrders/active` and `orders` controllers | #519, #523, #520 | Streams `drug_orders`, `lab_orders`; write `create_drug_order`; group/command metadata consistently uses `drug_orders` | `connectorgen validate`; CLI manual shows `drug_orders list` and `drug_orders create`; root-array pagination disabled for `drug_orders` |
| Diagnoses | Bahmni-core `diagnosis/getDiagnoses`, `diagnosis/search`, `diagnosis`, `diagnosis/delete` | #519, #523, #520 | Stream `diagnoses`; write `create_diagnosis`; delete endpoint is blocked/destructive in `api_surface.json` | `connectorgen validate`; root-array pagination disabled for `diagnoses`; nullable primary-key decision remains captain-owned |
| Appointments | Appointments module plural and singular controllers: `/appointments`, `/appointment`, `/appointment/search`, status/providing/reschedule helpers | #519, #523, #520 | Stream `appointments` (scoped by `appointment_date`/`patient_uuid`); write `create_appointment` against the plural `POST /ws/rest/v1/appointments` controller, with the singular `POST /ws/rest/v1/appointment` save controller recorded as a blocked `duplicate` ledger row rather than a second covered endpoint; status-change/provider-response/reschedule workflow mutations are blocked until dedicated typed schemas exist | `connectorgen validate`; `TestConformance/bahmni` appointment fixture without offset pagination; docs/CLI surface |
| Appointment services/reference data | Appointments module `appointment-services`, service types, speciality, unavailability controllers | #520 (gap), #519 if later promoted | Recorded in `api_surface.json` as `GET /ws/rest/v1/appointment-services` blocked/excluded pending stream; not silently omitted | Parity matrix + `api_surface.json`; remaining executable gap reported |
| Bahmni patient search/profile/context | `POST /ws/rest/v1/bahmnicore/search/patient`, patient profile/context controllers | #521, #524, #520 | Direct read `bahmnicore patient-search`; direct read `bahmnicore patient-detail`; patient context helper covered by patient profile family/exclusion ledger | `connectorgen validate`; direct-read CLI surface; operation schema in `operations.json` |
| Attachments/documents | OpenMRS/Bahmni attachment endpoints including upload and bytes download | #522, #526, #520 | Write `upload_patient_document` with bounded multipart upload and redacted `document_file_path`; binary download blocked as `binary_read` | `connectorgen validate`; `writes.json` multipart schema; docs/CLI flag `--document-file-path` |
| FHIR2 clinical resources | OpenMRS FHIR2 R4 resources including Patient, Observation, Encounter, Condition | #521, #520 | Direct reads `fhir patient-read`, `fhir observation-read`, `fhir encounter-read`, `fhir condition-read`; broader FHIR write/search surface is not exposed generically | `connectorgen validate`; CLI direct-read surface |
| Auth/session | `/ws/rest/v1/session`, Bahmni-core `whoami`, authenticated REST reads | #520 | `/session` is explicitly excluded for health checks because it may return 200 with `authenticated:false`; check uses authenticated provider list | `fixtures/check.json`; `api_surface.json`; conformance check fixture |
| UI/app configuration and global properties | Bahmni-core `config/*`, `sql/globalproperty`, OpenMRS `systemsetting`, admin import/export/tasks | #520 | Blocked/excluded as admin/config surface; no generic config/global-property read/write command | `api_surface.json` blocked rows; `connectorgen validate` |
| Notes/free-text clinical notes | Bahmni-core `GET/POST /ws/rest/v1/notes`, `POST/DELETE /notes/{id}` | #520 (gap) | Blocked/excluded pending dedicated typed schema and redaction design; no generic notes write/read command | `api_surface.json` blocked rows; remaining executable gap reported |
| Patient image and binary media reads | Bahmni-core `GET /ws/rest/v2/patientImage`; attachment bytes | #522, #520 | Blocked as binary reads; current connector exposes bounded multipart upload but no generic binary download | `api_surface.json`; docs known limits |
| Discharge/ADT and workflow mutations | Bahmni-core discharge/endVisit/ADT style controllers | #520 (gap) | Blocked/excluded pending dedicated workflow-specific typed write commands | `api_surface.json` blocked rows; remaining executable gap reported |
| Other UI/reporting/display helpers | Disease summaries, disposition, obs relationships, forms details, teleconsult link, display controls | #520 | Excluded as UI/reporting/control helper surface unless a future issue promotes one to typed read/write | Parity audit records the family; no generic raw HTTP escape hatch |

## Remaining executable parity gaps

The audit found official operation families that are CLI-relevant but not safely executable in this
connector PR. They are not silently omitted; they are recorded in `api_surface.json` and remain as
explicit gaps or blocked exclusions:

1. Appointment service/reference-data reads (`appointment-services`, service types, speciality,
   unavailability) could become streams/direct reads in a future slice.
2. Clinical notes require a dedicated typed schema and a product decision around clinical free-text
   redaction before read/write exposure.
3. Appointment workflow mutations (status transitions, provider response, reschedule) require typed
   schemas/approval text before write exposure.
4. Discharge/ADT/end-visit workflow mutations require dedicated workflow-specific write commands.
5. Patient image / document binary downloads remain blocked until the engine has an approved bounded
   binary-read UX.
6. Admin/config/global-property/import/export/task helpers remain out of scope and blocked to avoid
   generic admin/config escape hatches.

## Captain-owned decisions preserved

- Broad clinical PHI field redaction remains unresolved. This PR does **not** add an engine-level PHI
  redaction policy and does not claim one exists. It only makes the local multipart file-path field
  (`document_file_path`) match existing reverse-plan redaction markers and softens false documentation
  claims.
- The nullable `diagnoses.existingObs` primary-key finding remains unresolved. The connector keeps the
  existing metadata unless the captain chooses to drop or replace that primary key.
