---
name: pm-bahmni
description: Bahmni connector knowledge and safe action guide.
---

# pm-bahmni

## Purpose

Reads clinical EMR data from a Bahmni deployment, including local Bahmni/bahmni-docker setups, through the OpenMRS REST, Bahmni-core REST, and OpenMRS FHIR2 R4 APIs — patients, encounters, observations, visits, concepts, locations, providers, orders, lab results, appointments, and diagnoses — executes bounded direct reads and a schema-gated typed GET patient search, and models retained OpenMRS/Bahmni clinical mutations as approval-gated, schema-bound reverse-ETL actions. Output is bounded; secret-shaped fields and typed patient-search declared sensitive fields are redacted by the current runtime, while broad clinical PHI field redaction remains a separate engine policy decision.

## Icon

- id: pm-sample
- asset: icons/pm-sample.svg
- source: polymetrics
- review_status: polymetrics
- review_url: https://github.com/polymetrics-ai/cli

## Capabilities

- check=true catalog=true read=true write=true query=false
- Integration type: api

## Authentication

- Use pm credentials add with --from-env or --value-stdin for secret fields.

## Configuration

- appointment_date (required)
- base_url (required)
- lab_result_concepts (required)
- patient_query (required)
- patient_uuid (required)
- username (required)
- password (secret) (required)

## ETL Streams

- patients:
  - primary key: uuid
  - fields: display(string), identifiers(array), person(object), uuid(string), voided(boolean)
- encounters:
  - primary key: uuid
  - fields: display(string), encounterDatetime(string), encounterType(object), patient(object), uuid(string), visit(object)
- observations:
  - primary key: uuid
  - fields: concept(object), display(string), obsDatetime(string), uuid(string), value(string)
- visits:
  - primary key: uuid
  - fields: display(string), location(object), startDatetime(string), stopDatetime(string), uuid(string), visitType(object)
- concepts:
  - primary key: uuid
  - fields: conceptClass(object), datatype(object), display(string), name(object), uuid(string)
- locations:
  - primary key: uuid
  - fields: description(string), display(string), name(string), tags(array), uuid(string)
- providers:
  - primary key: uuid
  - fields: attributes(array), display(string), identifier(string), person(object), uuid(string)
- drug_orders:
  - primary key: uuid
  - fields: dateActivated(string), display(string), dose(number), doseUnits(object), drug(object), uuid(string)
- lab_orders:
  - primary key: uuid
  - fields: accessionNumber(string), concept(object), dateActivated(string), display(string), orderType(object), uuid(string)
- lab_results:
  - primary key: uuid
  - fields: concept(object), display(string), obsDatetime(string), uuid(string), value(string)
- appointments:
  - primary key: uuid
  - fields: display(string), endDateTime(string), patient(object), service(object), startDateTime(string), status(string), uuid(string)
- diagnoses:
  - fields: certainty(string), codedAnswer(object), diagnosisDateTime(string), display(string), existingObs(string), order(string)

## Sync Modes

- ETL sync modes: full_refresh_append, full_refresh_overwrite

## Reverse ETL Actions

- create_patient:
  - endpoint: POST /ws/rest/v1/patient
  - required fields: identifiers, person
  - risk: high: writes clinical Bahmni/OpenMRS data; requires reverse ETL plan, preview, approval, execute against synthetic/disposable records unless explicitly approved
- update_patient:
  - endpoint: POST /ws/rest/v1/patient/{{ record.uuid }}
  - required fields: uuid, person
  - risk: critical: mutates clinical Bahmni/OpenMRS records; requires reverse ETL approval and destructive confirmation for non-disposable use
- create_encounter:
  - endpoint: POST /ws/rest/v1/encounter
  - required fields: patient, encounterType, encounterDatetime, location
  - risk: high: writes clinical Bahmni/OpenMRS data; requires reverse ETL plan, preview, approval, execute against synthetic/disposable records unless explicitly approved
- create_observation:
  - endpoint: POST /ws/rest/v1/obs
  - required fields: person, obsDatetime, concept, value
  - risk: high: writes clinical Bahmni/OpenMRS data; requires reverse ETL plan, preview, approval, execute against synthetic/disposable records unless explicitly approved
- create_visit:
  - endpoint: POST /ws/rest/v1/visit
  - required fields: patient, visitType, startDatetime, location
  - risk: high: writes clinical Bahmni/OpenMRS data; requires reverse ETL plan, preview, approval, execute against synthetic/disposable records unless explicitly approved
- create_lab_order:
  - endpoint: POST /ws/rest/v1/order
  - required fields: type, encounter, patient, concept, careSetting, orderer, orderType, action
  - risk: high: writes clinical Bahmni/OpenMRS data; requires reverse ETL plan, preview, approval, execute against synthetic/disposable records unless explicitly approved
- create_patient_diagnosis:
  - endpoint: POST /ws/rest/v1/patientdiagnoses
  - required fields: diagnosis, encounter, condition, certainty, patient, rank
  - risk: critical: mutates clinical Bahmni/OpenMRS records; requires reverse ETL approval and destructive confirmation for non-disposable use
- create_appointment:
  - endpoint: POST /ws/rest/v1/appointments
  - required fields: patientUuid, serviceUuid, startDateTime, endDateTime, appointmentKind, status
  - risk: high: writes clinical Bahmni/OpenMRS data; requires reverse ETL plan, preview, approval, execute against synthetic/disposable records unless explicitly approved
- update_appointment_status:
  - endpoint: POST /ws/rest/v1/appointments/{{ record.appointmentUuid }}/status-change
  - required fields: appointmentUuid, toStatus, onDate
  - risk: critical: mutates clinical Bahmni/OpenMRS records; requires reverse ETL approval and destructive confirmation for non-disposable use
- update_appointment_provider_response:
  - endpoint: POST /ws/rest/v1/appointments/{{ record.appointmentUuid }}/providerResponse
  - required fields: appointmentUuid, uuid, response
  - risk: high: writes clinical Bahmni/OpenMRS data; requires reverse ETL plan, preview, approval, execute against synthetic/disposable records unless explicitly approved
- create_note:
  - endpoint: POST /ws/rest/v1/notes
  - required fields: notes
  - risk: high: writes clinical Bahmni/OpenMRS data; requires reverse ETL plan, preview, approval, execute against synthetic/disposable records unless explicitly approved

## Security

- read risk: external Bahmni/OpenMRS clinical PHI read of patient, encounter, observation, visit, order, lab, appointment, and diagnosis data; reads are bounded, secret-shaped fields are redacted by existing output policies, and typed patient search redacts declared identifier/address/name/date fields, but remaining clinical PHI fields are not generally field-redacted by the current engine
- write risk: typed Bahmni/OpenMRS reverse ETL clinical mutations for retained, live-proven patient, encounter, observation, visit, lab order, diagnosis, appointment create/status/provider-response, and note routes; drug order, appointment reschedule, bulk observation, and document upload surfaces are blocked unless separately typed and live-proven
- approval: reverse ETL writes require plan, preview, approval, execute; clinical/destructive actions require --confirm destructive
- Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

## Command Surface

- Inspect, read, and safely plan typed Bahmni clinical operations.
- Usage: pm bahmni <command> [flags]
- Source CLI: Bahmni / OpenMRS REST + FHIR2 (OpenMRS REST v1, Bahmni-core REST, OpenMRS FHIR2 R4)
- Global flags:
  - --credential (string): Credential name to use for the Bahmni request.
  - --connection (string): Alias for --credential.
  - --config (string_array): Connector config override as key=value.
  - --json (boolean): Emit machine-readable JSON output.
  - --limit (integer): Maximum ETL records to emit for stream commands.
  - --max-bytes (integer): Maximum direct-read response bytes; typed operations declare their own lower cap.
  - --plan (string): Execute an approved reverse-ETL plan by id.
  - --preview (boolean): Preview a reverse-ETL write command without making a network mutation.
  - --approval-token-stdin (boolean): Read the approval token as one bounded line from standard input.
  - --confirm (string): Typed confirmation challenge for destructive reverse-ETL writes.
- Clinical data
  - patients list - List Bahmni/OpenMRS patients as ETL records (requires a patient search term). [intent=etl availability=implemented stream=patients]
  - patients create - Create a Bahmni/OpenMRS patient record (clinical PHI). [intent=reverse_etl availability=implemented write=create_patient]; approval: reverse ETL plan -> preview -> approval -> execute; clinical/destructive actions require --confirm destructive.; risk: clinical reverse-ETL mutation of Bahmni/OpenMRS PHI; runs only through plan, preview, approval, execute.; notes: Typed reverse-ETL action; no raw HTTP body, raw JSON, generic shell, or SQL write is exposed.; flags: --identifier, --identifier-type, --identifier-location, --identifier-preferred, --given-name, --family-name, --middle-name, --gender, --birthdate, --address1, --city-village, --state-province, --country, --postal-code
  - patients update - Update an existing Bahmni/OpenMRS patient record (clinical PHI). [intent=reverse_etl availability=implemented write=update_patient]; approval: reverse ETL plan -> preview -> approval -> execute; clinical/destructive actions require --confirm destructive.; risk: clinical reverse-ETL mutation of Bahmni/OpenMRS PHI; runs only through plan, preview, approval, execute.; notes: Typed reverse-ETL action; no raw HTTP body, raw JSON, generic shell, or SQL write is exposed.; flags: --uuid, --given-name, --family-name, --middle-name, --gender, --birthdate, --address1, --city-village, --state-province, --country, --postal-code
  - encounters list - List patient encounters as ETL records (scoped by patient_uuid). [intent=etl availability=implemented stream=encounters]
  - encounters create - Create a clinical encounter with explicit typed scalar fields. [intent=reverse_etl availability=implemented write=create_encounter]; approval: reverse ETL plan -> preview -> approval -> execute; clinical/destructive actions require --confirm destructive.; risk: clinical reverse-ETL mutation of Bahmni/OpenMRS PHI; runs only through plan, preview, approval, execute.; notes: Typed reverse-ETL action; no raw HTTP body, raw JSON, generic shell, or SQL write is exposed.; flags: --patient, --encounter-type, --encounter-datetime, --location, --visit
  - observations list - List patient observations (obs) as ETL records (scoped by patient_uuid). [intent=etl availability=implemented stream=observations]
  - observations create - Record a clinical observation (obs) for a patient. [intent=reverse_etl availability=implemented write=create_observation]; approval: reverse ETL plan -> preview -> approval -> execute; clinical/destructive actions require --confirm destructive.; risk: clinical reverse-ETL mutation of Bahmni/OpenMRS PHI; runs only through plan, preview, approval, execute.; notes: Typed reverse-ETL action; no raw HTTP body, raw JSON, generic shell, or SQL write is exposed.; flags: --person, --obs-datetime, --concept, --value
  - visits list - List patient visits as ETL records (scoped by patient_uuid). [intent=etl availability=implemented stream=visits]
  - visits create - Create an OpenMRS visit. [intent=reverse_etl availability=implemented write=create_visit]; approval: reverse ETL writes require plan, preview, approval, execute; risk: high; notes: Version-pinned supported write route; execute only through reverse ETL approval against disposable synthetic records unless explicitly approved.; flags: --patient, --visit-type, --start-datetime, --stop-datetime, --location
  - diagnoses list - List Bahmni-core patient diagnoses as ETL records (scoped by patient_uuid). [intent=etl availability=implemented stream=diagnoses]
  - diagnoses create - Create an OpenMRS patient diagnosis through the supported patientdiagnoses resource. [intent=reverse_etl availability=implemented write=create_patient_diagnosis]; approval: reverse ETL writes require plan, preview, approval, execute; risk: critical; notes: Version-pinned supported write route; execute only through reverse ETL approval against disposable synthetic records unless explicitly approved.; flags: --patient, --encounter, --condition, --certainty, --rank, --non-coded-diagnosis, --coded-diagnosis, --specific-name
- Catalog & reference
  - concepts list - List OpenMRS concepts as ETL records. [intent=etl availability=implemented stream=concepts]
  - locations list - List OpenMRS locations as ETL records. [intent=etl availability=implemented stream=locations]
  - providers list - List OpenMRS providers as ETL records. [intent=etl availability=implemented stream=providers]
- Orders & labs
  - drug_orders list - List active Bahmni drug orders as ETL records (scoped by patient_uuid). [intent=etl availability=implemented stream=drug_orders]
  - drug_orders create - Drug order creation is not exposed in this checkpoint; the local pinned Bahmni/OpenMRS lab rejected the pinned drug-order body at execute and the direct counterfactual failed against the same endpoint. [intent=reverse_etl availability=unsupported_local unsupported local workflow]; approval: reverse ETL plan -> preview -> approval -> execute; clinical/destructive actions require --confirm destructive.; risk: unsupported local drug-order mutation; no live proof, not retained as an executable write; notes: Blocked until a concrete upstream-required field or lab capability is proven. Diagnostic evidence: typed plan/preview reached execute with HTTP 400 safe fields action,drug; direct counterfactual to POST /ws/rest/v1/order returned HTTP 400 with upstream code webservices.rest.error.invalid.submission. Pinned source: OpenMRS webservices REST OrderResource1_10 + DrugOrderSubclassHandler1_10/1_12 define the route/properties, but current lab did not accept the generated pinned-shape body.
  - lab_orders list - List lab (test) orders as ETL records (scoped by patient_uuid). [intent=etl availability=implemented stream=lab_orders]
  - lab_orders create - Create an OpenMRS test/lab order through the typed order resource. [intent=reverse_etl availability=implemented write=create_lab_order]; approval: reverse ETL writes require plan, preview, approval, execute; risk: high; notes: Version-pinned supported write route; execute only through reverse ETL approval against disposable synthetic records unless explicitly approved.; flags: --type, --patient, --encounter, --concept, --care-setting, --orderer, --order-type, --action, --clinical-history
  - lab_results list - List Bahmni-core lab result observations as ETL records (scoped by patient_uuid). [intent=etl availability=implemented stream=lab_results]
- Scheduling
  - appointments list - List Bahmni appointments as ETL records scoped by appointment_date only. [intent=etl availability=implemented stream=appointments]
  - appointments create - Book a Bahmni patient appointment. [intent=reverse_etl availability=implemented write=create_appointment]; approval: reverse ETL plan -> preview -> approval -> execute; clinical/destructive actions require --confirm destructive.; risk: clinical reverse-ETL mutation of Bahmni/OpenMRS PHI; runs only through plan, preview, approval, execute.; notes: Typed reverse-ETL action; no raw HTTP body, raw JSON, generic shell, or SQL write is exposed.; flags: --patient-uuid, --service-uuid, --start-date-time, --end-date-time, --appointment-kind, --status, --provider-uuid, --provider-response, --location-uuid, --comments
  - appointments status-change - Change an appointment status. [intent=reverse_etl availability=implemented write=update_appointment_status]; approval: reverse ETL writes require plan, preview, approval, execute; risk: critical; notes: Version-pinned supported write route; execute only through reverse ETL approval against disposable synthetic records unless explicitly approved.; flags: --appointment-uuid, --to-status, --on-date
  - appointments provider-response - Update an appointment provider response. [intent=reverse_etl availability=implemented write=update_appointment_provider_response]; approval: reverse ETL writes require plan, preview, approval, execute; risk: high; notes: Version-pinned supported write route; execute only through reverse ETL approval against disposable synthetic records unless explicitly approved.; flags: --appointment-uuid, --provider-detail-uuid, --response
  - appointments reschedule - Appointment reschedule is not exposed in this checkpoint; the local pinned Bahmni appointments lab rejected the singular-controller AppointmentRequest-shaped body. [intent=reverse_etl availability=unsupported_local unsupported local workflow]; approval: reverse ETL writes require plan, preview, approval, execute; risk: unsupported local appointment reschedule mutation; no live proof, not retained as an executable write; notes: Blocked until the singular reschedule route is live-proven or a concrete upstream-required field is identified. Diagnostic evidence: typed plan/preview reached execute with HTTP 400 safe field uuid; direct counterfactual to POST /ws/rest/v1/appointment/{uuid}/reschedule returned HTTP 400 with upstream code org.hibernate.exception.internal.SQLExceptionTypeDelegate:59. Pinned source: AppointmentController.rescheduleAppointment accepts AppointmentRequest and calls AppointmentsServiceImpl.reschedule, but current lab did not accept the generated contract-shaped body.
- Direct reads
  - patient get - Retrieve a single patient resource by UUID. [intent=direct_read availability=implemented]; risk: bounded Bahmni/OpenMRS JSON read; the response is size-limited and secret-shaped fields are redacted.; flags: --uuid (max 4096 bytes), --page, --page-cursor
  - encounter get - Retrieve a single encounter resource by UUID. [intent=direct_read availability=implemented]; risk: bounded Bahmni/OpenMRS JSON read; the response is size-limited and secret-shaped fields are redacted.; flags: --uuid (max 4096 bytes), --page, --page-cursor
  - visit get - Retrieve a single visit resource by UUID. [intent=direct_read availability=implemented]; risk: bounded Bahmni/OpenMRS JSON read; the response is size-limited and secret-shaped fields are redacted.; flags: --uuid (max 4096 bytes), --page, --page-cursor
  - concept get - Retrieve a single concept resource by UUID. [intent=direct_read availability=implemented]; risk: bounded Bahmni/OpenMRS JSON read; the response is size-limited and secret-shaped fields are redacted.; flags: --uuid (max 4096 bytes), --page, --page-cursor
  - provider get - Retrieve a single provider resource by UUID. [intent=direct_read availability=implemented]; risk: bounded Bahmni/OpenMRS JSON read; the response is size-limited and secret-shaped fields are redacted.; flags: --uuid (max 4096 bytes), --page, --page-cursor
  - location get - Retrieve a single location resource by UUID. [intent=direct_read availability=implemented]; risk: bounded Bahmni/OpenMRS JSON read; the response is size-limited and secret-shaped fields are redacted.; flags: --uuid (max 4096 bytes), --page, --page-cursor
  - fhir patient-read - Read a FHIR R4 Patient resource by id. [intent=direct_read availability=implemented]; risk: bounded Bahmni/OpenMRS JSON read; the response is size-limited and secret-shaped fields are redacted.; flags: --id (max 4096 bytes), --page, --page-cursor
  - fhir observation-read - Read a FHIR R4 Observation resource by id. [intent=direct_read availability=implemented]; risk: bounded Bahmni/OpenMRS JSON read; the response is size-limited and secret-shaped fields are redacted.; flags: --id (max 4096 bytes), --page, --page-cursor
  - fhir encounter-read - Read a FHIR R4 Encounter resource by id. [intent=direct_read availability=implemented]; risk: bounded Bahmni/OpenMRS JSON read; the response is size-limited and secret-shaped fields are redacted.; flags: --id (max 4096 bytes), --page, --page-cursor
  - fhir condition-read - Read a FHIR R4 Condition resource by id. [intent=direct_read availability=implemented]; risk: bounded Bahmni/OpenMRS JSON read; the response is size-limited and secret-shaped fields are redacted.; flags: --id (max 4096 bytes), --page, --page-cursor
  - bahmnicore patient-search - Search patients with the pinned bahmni-commons GET patient search route. [intent=direct_read availability=implemented operation=bahmni.patient_search]; approval: none: read-only GET query with allow-listed query parameters and bounded JSON output.; risk: bounded typed GET read-query; the response is size-capped, secret-shaped fields are redacted, and the operation direct-read engine redacts declared patient-search sensitive fields including identifier, addressFieldValue, display, givenName, middleName, familyName, birthDate, and deathDate. Remaining clinical PHI fields should still be treated as clinical data.; notes: Executes through the typed operation direct-read engine; no raw request body or generic HTTP flag is exposed.; flags: --page, --page-cursor
- Documents
  - documents upload - Visit-document upload is not advertised until a file-backed bounded multipart/hash-gated typed surface is implemented and live-proven. [intent=reverse_etl availability=unsafe_or_disallowed]; approval: blocked: inline content upload is not a retained typed write surface; risk: critical; notes: Blocked for PR #533: current inline JSON content surface lacks the claimed file snapshot/SHA-256 approval binding; do not use as an advertised write.
- Other Commands
  - notes create - Create one Bahmni note through a schema-bound top-level JSON array body. [intent=reverse_etl availability=implemented write=create_note]; approval: reverse ETL writes require plan, preview, approval, execute; risk: high; notes: Version-pinned supported write route; execute only through reverse ETL approval against disposable synthetic records unless explicitly approved.; flags: --note-type-name, --note-text, --note-date
- Help topics:
  - bahmni-auth - Point base_url at a Bahmni OpenMRS instance, including a local Bahmni/bahmni-docker deployment, and supply username/password via credentials; never pass secrets in command text.
  - bahmni-writes - Bahmni clinical mutations are typed reverse-ETL actions with plan, preview, approval, execute gates; clinical/destructive actions require --confirm destructive.
  - bahmni-direct-read - Bahmni direct reads are bounded JSON GET-by-uuid, FHIR read-by-id, or a schema-gated typed GET patient search, all with a response byte cap; typed patient search redacts its declared identifier, address, name, birth-date, and death-date fields.
  - bahmni-phi - Bahmni reads and writes can include clinical PHI. The current runtime bounds output, redacts secret-shaped fields, configured write path identifiers, and typed patient-search declared sensitive fields; broad clinical PHI field redaction remains a separate engine policy decision.

## Commands

### Inspect as a manual

```bash
pm connectors inspect bahmni
```

### Inspect as structured JSON

```bash
pm connectors inspect bahmni --json
```

### Command discovery

```bash
pm bahmni --help
pm bahmni appointments --help
pm bahmni appointments create --help
```

### Synthetic appointment read

```bash
pm bahmni appointments list --credential bahmni-local --config appointment_date=2026-01-01T00:00:00.000 --limit 10 --json
```

### Synthetic patient create plan

```bash
pm bahmni patients create --credential bahmni-local --identifier SYN-CONN-EXAMPLE-001 --identifier-type <identifier-type-uuid> --identifier-location <location-uuid> --given-name Synthetic --family-name Connector --gender O --birthdate 1990-01-01 --preview --json
```

### Unsupported retained as blocked

```bash
pm bahmni appointments reschedule --help
pm bahmni drug_orders create --help
```

## Agent Rules

- Run pm connectors inspect bahmni before creating credentials or plans.
- Use --json only when the caller needs structured output; use the manual for human-readable guidance.
- Never ask the user to paste secret values into chat.
- For reverse ETL writes, create a plan, show the preview, wait for explicit approval, then run with the approval token.
