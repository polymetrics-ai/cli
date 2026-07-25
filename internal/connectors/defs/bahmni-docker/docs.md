# Overview

The `bahmni-docker` connector reads clinical EMR data from a local Bahmni deployment (the
open-source hospital/EMR distribution deployed via https://github.com/Bahmni/bahmni-docker) through
the OpenMRS REST v1, Bahmni-core REST, and OpenMRS FHIR2 R4 APIs exposed by the deployment. It is
config-driven and usable locally: point `base_url` at a running instance (for example
`http://localhost/openmrs`), supply an OpenMRS username/password via credentials, and inspect it with
`pm connectors inspect bahmni-docker --json` without reading or printing any secret values.

Executable ETL streams: `patients`, `encounters`, `observations`, `visits`, `concepts`, `locations`,
`providers`, `drug_orders`, `lab_orders`, `lab_results`, `appointments`, `diagnoses`.

Bounded direct-read commands cover GET-by-UUID for patient/encounter/visit/concept/provider/location,
FHIR R4 read-by-id for Patient/Observation/Encounter/Condition, a consolidated Bahmni-core patient
profile detail, and a schema-gated typed POST patient search. Clinical mutations, a top-level JSON
array observation upload, and a bounded multipart patient-document upload are modeled as typed
reverse-ETL write actions.

Bahmni handles clinical PHI, so patient-linked data is redacted by default on reads and in write
plans.

## Auth setup

Connection fields:

- `base_url` (required, string, format uri); default `http://localhost/openmrs`; base URL of the
  local bahmni-docker OpenMRS instance.
- `username` (required, string); OpenMRS username for HTTP Basic authentication.
- `password` (required, secret, string); OpenMRS password for HTTP Basic authentication; never
  logged or printed.
- `patient_query` (optional, string); identifier or name search term used to enumerate the
  `patients` stream (OpenMRS patient search `q`).
- `patient_uuid` (optional, string); patient UUID context used to scope patient-linked streams
  (encounters, observations, visits, orders, lab results, diagnoses).

Secret fields are redacted in logs and write previews: `password`.

Authentication behavior: HTTP Basic authentication using `config.username` and `secrets.password`.
Connection checks call `GET /ws/rest/v1/session`.

## Streams notes

All OpenMRS REST list streams return a `{ "results": [...] }` envelope and page with the OpenMRS
`limit`/`startIndex` (offset) convention. Bahmni-core and appointment streams (`drug_orders`,
`lab_results`, `appointments`, `diagnoses`) return top-level JSON arrays and are read from the
response root. Patient-linked streams require a `patient_uuid` (or, for `patients`, a
`patient_query`) config value to scope the request; see Known limits.

## Write actions & risks

Write actions are declared in `writes.json` as typed Bahmni/OpenMRS reverse-ETL mutations:
`create_patient`, `update_patient`, `create_encounter`, `create_observation`, `create_appointment`,
`create_drug_order`, `create_diagnosis`, the top-level JSON array `create_observations_bulk`
observation upload, and the bounded multipart `upload_patient_document` upload.

Safety gates:

- Use reverse ETL plan -> preview -> approval -> execute.
- Clinical/destructive actions declare `confirm: destructive` (update patient, drug order, diagnosis,
  bulk observation upload, and document upload).
- No generic raw HTTP write, raw JSON body, arbitrary OpenMRS resource method/path/body escape hatch,
  generic shell write, or SQL write is exposed.
- The multipart document upload accepts only a declared project-local file path field, binds approval
  to a SHA-256 content digest, snapshots and verifies the approved bytes before any HTTP request, and
  enforces byte limits; file/path/content and patient identifiers are redacted in command plans.
- The top-level JSON array observation upload uses a declared `body_field` and `body_schema`; no raw
  JSON CLI flag is exposed.

PHI redaction: patient identifiers, names, addresses, and clinical observation/diagnosis values are
redacted by default on patient-linked reads and in write plans.

Read risk: external Bahmni/OpenMRS clinical PHI read; direct reads are bounded and PHI-redacted.

Write risk: typed Bahmni/OpenMRS reverse ETL clinical mutations.

Approval: reverse ETL writes require plan, preview, approval, execute; clinical/destructive actions
require `--confirm destructive`.

## Known limits

- Batch defaults: read_page_size=50; streams page with OpenMRS `limit`/`startIndex`.
- OpenMRS/Bahmni REST does not document a fixed public rate limit; the target is a self-hosted local
  deployment. The connector still bounds response sizes and paginates.
- OpenMRS/Bahmni REST lacks a universal "modified since" cursor across these list endpoints, so
  streams are full-refresh (offset/limit) rather than incremental.
- `patients` enumeration requires a search term: set `patient_query` (OpenMRS patient list is a
  search endpoint). Patient-linked streams (encounters/observations/visits/orders/lab results/
  diagnoses) require a `patient_uuid` context.
- The typed Bahmni patient search is modeled as a schema-gated JSON POST read-query with
  connector-authored flags; arbitrary/raw request bodies remain intentionally unavailable.
- Patient-document binary download is modeled as bounded binary metadata with a bounded local-output
  policy (api_surface `binary_read`), not exposed as a generic HTTP byte-stream download. Permanent
  patient deletion/purge and OpenMRS server administration are blocked with recorded evidence.
- Live clinical writes and patient-document payload tests beyond fixtures are human-gated against any
  non-local/non-disposable deployment.
