# Overview

The `bahmni` connector reads clinical EMR data from a Bahmni deployment (including the local
Bahmni/bahmni-docker docker-compose setup at https://github.com/Bahmni/bahmni-docker) through the
OpenMRS REST v1, Bahmni-core REST, and OpenMRS FHIR2 R4 APIs exposed by the deployment. It is
config-driven and usable locally: point `base_url` at a running instance (for example
`http://localhost/openmrs`), supply an OpenMRS username/password via credentials, and inspect it with
`pm connectors inspect bahmni --json` without reading or printing any secret values.

Executable ETL streams: `patients`, `encounters`, `observations`, `visits`, `concepts`, `locations`,
`providers`, `drug_orders`, `lab_orders`, `lab_results`, `appointments`, `diagnoses`.

Bounded direct-read commands cover GET-by-UUID for patient/encounter/visit/concept/provider/location,
FHIR R4 read-by-id for Patient/Observation/Encounter/Condition, and a schema-gated typed GET Bahmni
patient search. Retained clinical mutations are modeled only as approval-gated, schema-bound
reverse-ETL write actions. Drug-order create, appointment reschedule, bulk observation upload, and
visit-document upload remain blocked until a safe typed surface is implemented and live-proven.

Bahmni reads and writes can include clinical PHI. The current runtime bounds output, redacts
secret-shaped fields, redacts configured write path identifiers, and redacts the typed patient search
fields declared in `operations.json` (`identifier`, `addressFieldValue`, display/name, and
birth/death dates). Broad clinical PHI field redaction remains a separate engine policy decision.

## Auth setup

Connection fields:

- `base_url` (required, string, format uri); default `http://localhost/openmrs`; base URL of the
  Bahmni OpenMRS instance, including local Bahmni/bahmni-docker deployments.
- `username` (required, string); OpenMRS username for HTTP Basic authentication.
- `password` (required, secret, string); OpenMRS password for HTTP Basic authentication; never
  logged or printed.
- `patient_query` (required, string); identifier or name search term used to enumerate the
  `patients` stream (OpenMRS patient search `q`).
- `patient_uuid` (required, string); patient UUID context used to scope patient-linked streams
  (encounters, observations, visits, orders, lab results, diagnoses).
- `lab_result_concepts` (required, string); comma-separated Bahmni concept display names for the
  Bahmni lab-results observation route.
- `appointment_date` (required, string); `yyyy-MM-ddTHH:mm:ss.SSS` appointment day scope used by the
  pinned appointments module; that route does not honor `patientUuid` scoping.

Secret fields are redacted in logs and write previews: `password`.

Authentication behavior: HTTP Basic authentication using `config.username` and `secrets.password`.
Connection checks call an authenticated bounded provider-list endpoint (`GET /ws/rest/v1/provider`)
with `limit=1`, so bad basic-auth credentials fail instead of passing on an unauthenticated session
probe.

## Streams notes

OpenMRS REST list streams return a `{ "results": [...] }` envelope and page with the OpenMRS
`limit`/`startIndex` (offset) convention. Bahmni-core and appointment streams (`drug_orders`,
`lab_results`, `appointments`, `diagnoses`) return top-level JSON arrays and explicitly disable the
inherited offset paginator because those endpoints do not honor `limit`/`startIndex`. Patient-linked
streams require a `patient_uuid` (or, for `patients`, a `patient_query`) config value to scope the
request; see Known limits. Because the appointment endpoint does not page or honor `patientUuid`, the
connector requires `appointment_date` rather than advertising patient-scoped appointment reads.

## Write actions & risks

Write actions are declared in `writes.json` as typed Bahmni/OpenMRS reverse-ETL mutations:
`create_patient`, `update_patient`, `create_encounter`, `create_observation`, `create_visit`,
`create_lab_order`, `create_patient_diagnosis`, `create_appointment`,
`update_appointment_status`, `update_appointment_provider_response`, and `create_note`.

Safety gates:

- Use reverse ETL plan -> preview -> approval -> execute.
- Clinical/destructive actions declare `confirm: destructive` where the retained route mutates an
  existing clinical record or high-risk workflow state.
- No generic raw HTTP write, raw JSON body, arbitrary OpenMRS resource method/path/body escape hatch,
  generic shell write, or SQL write is exposed.
- Structured write payloads are built from connector-authored scalar/enum/boolean flag mappings such
  as `record.person.names.0.givenName`; no raw JSON CLI flag is exposed.
- Drug-order create is not retained in this PR: pinned OpenMRS webservices REST exposes the
  `drugorder` subclass fields, but the local pinned lab rejected both the typed generated body and a
  diagnostic-only direct counterfactual against `POST /ws/rest/v1/order`.
- Appointment reschedule is not retained in this PR: pinned appointments source exposes the singular
  reschedule controller, but the local pinned lab rejected both the typed generated body and a
  diagnostic-only direct counterfactual against `POST /ws/rest/v1/appointment/{uuid}/reschedule`.
- Visit-document upload is not retained in this PR because the previous inline-content surface lacked
  the claimed file snapshot/SHA-256 approval binding.
- The unsupported top-level bulk-observation route remains blocked; single observations use the
  typed `POST /ws/rest/v1/obs` action.

PHI note: the typed Bahmni patient search redacts its declared identifier, address value,
display/name, birth-date, and death-date fields. Other patient identifiers, names, addresses, and
clinical observation/diagnosis values are not generally field-redacted by the current connector
engine. Treat command output and write plans as clinical data unless and until a broader engine PHI
redaction policy is authorized.

Read risk: external Bahmni/OpenMRS clinical PHI read; direct reads are bounded, generic direct reads
redact secret-shaped fields, and typed patient search redacts declared identifier/address/name/date
fields.

Write risk: typed Bahmni/OpenMRS reverse ETL clinical mutations.

Approval: reverse ETL writes require plan, preview, approval, execute; clinical/destructive actions
require `--confirm destructive`.

## Known limits

- Batch defaults: read_page_size=50; only OpenMRS envelope streams use `limit`/`startIndex` paging.
- OpenMRS/Bahmni REST does not document a fixed public rate limit; the target is a self-hosted local
  deployment. The connector still bounds response sizes and paginates where the endpoint supports it.
- OpenMRS/Bahmni REST lacks a universal "modified since" cursor across these list endpoints, so
  streams are full-refresh rather than incremental.
- `patients` enumeration requires a search term: set `patient_query` (OpenMRS patient list is a
  search endpoint). Patient-linked streams (encounters/observations/visits/orders/lab results/
  diagnoses) require a `patient_uuid` context. `appointments` requires `appointment_date`; the pinned
  appointment controller ignores `patientUuid`.
- The typed Bahmni patient search is modeled as a schema-gated GET read-query with connector-authored
  query flags and operation-level redaction for declared identifier/address/name/date fields;
  arbitrary/raw request bodies remain intentionally unavailable.
- Patient-document binary download and patient image binary reads are blocked by default rather than
  exposed as generic byte-stream downloads. Permanent patient deletion/purge, OpenMRS server
  administration/global-property helpers, bulk observation upload, visit-document upload, and
  discharge/ADT mutations are blocked or excluded with recorded evidence in `execution bundle`.
- Live clinical writes are human-gated against any non-local/non-disposable deployment.
