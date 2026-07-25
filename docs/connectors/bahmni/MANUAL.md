# pm connectors inspect bahmni

```text
NAME
  pm connectors inspect bahmni - Bahmni connector manual

SYNOPSIS
  pm connectors inspect bahmni
  pm connectors inspect bahmni --json
  pm credentials add <name> --connector bahmni [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Reads clinical EMR data from a Bahmni deployment, including local Bahmni/bahmni-docker setups, through the OpenMRS REST, Bahmni-core REST, and OpenMRS FHIR2 R4 APIs — patients, encounters, observations, visits, concepts, locations, providers, orders, lab results, appointments, and diagnoses — executes bounded direct reads and a schema-gated typed POST patient search, and models OpenMRS/Bahmni clinical mutations, a top-level JSON array observation upload, and a bounded multipart patient-document upload as typed reverse-ETL actions. Output is bounded; secret-shaped fields and file-path inputs are redacted by the current runtime, while broad clinical PHI field redaction remains a separate engine policy decision.

ICON
  asset: icons/pm-sample.svg
  source: polymetrics
  review_status: polymetrics
  review_url: https://github.com/polymetrics-ai/cli

CAPABILITIES
  check=true catalog=true read=true write=true query=false
  Integration type: api

AUTHENTICATION
  Use pm credentials add with --from-env or --value-stdin for secret fields.

CONFIGURATION
  base_url
  patient_query
  patient_uuid
  username
  password (secret)

ETL STREAMS
  patients:
    primary key: uuid
    fields: display(), identifiers(), person(), uuid(), voided()
  encounters:
    primary key: uuid
    fields: display(), encounterDatetime(), encounterType(), patient(), uuid(), visit()
  observations:
    primary key: uuid
    fields: concept(), display(), obsDatetime(), uuid(), value()
  visits:
    primary key: uuid
    fields: display(), location(), startDatetime(), stopDatetime(), uuid(), visitType()
  concepts:
    primary key: uuid
    fields: conceptClass(), datatype(), display(), name(), uuid()
  locations:
    primary key: uuid
    fields: description(), display(), name(), tags(), uuid()
  providers:
    primary key: uuid
    fields: attributes(), display(), identifier(), person(), uuid()
  drug_orders:
    primary key: uuid
    fields: dateActivated(), display(), dose(), doseUnits(), drug(), uuid()
  lab_orders:
    primary key: uuid
    fields: accessionNumber(), concept(), dateActivated(), display(), orderType(), uuid()
  lab_results:
    primary key: uuid
    fields: concept(), display(), obsDatetime(), uuid(), value()
  appointments:
    primary key: uuid
    fields: display(), endDateTime(), patient(), service(), startDateTime(), status(), uuid()
  diagnoses:
    primary key: existingObs
    fields: certainty(), codedAnswer(), diagnosisDateTime(), display(), existingObs(), order()

SYNC MODES
  ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped

REVERSE ETL ACTIONS
  create_patient:
    endpoint: POST /ws/rest/v1/patient
    required fields: identifiers, person
    risk: high: creates a clinical patient record with PHI in Bahmni/OpenMRS; requires reverse ETL approval
  update_patient:
    endpoint: POST /ws/rest/v1/patient/{{ record.uuid }}
    required fields: uuid
    risk: high: overwrites existing patient PHI in Bahmni/OpenMRS; requires reverse ETL approval and destructive confirmation
  create_encounter:
    endpoint: POST /ws/rest/v1/encounter
    required fields: patient, encounterType
    risk: high: writes a clinical encounter with linked observations for a patient; requires reverse ETL approval
  create_observation:
    endpoint: POST /ws/rest/v1/obs
    required fields: person, concept, value
    risk: high: records a clinical observation (vitals/diagnostic value) for a patient; requires reverse ETL approval
  create_appointment:
    endpoint: POST /ws/rest/v1/appointments
    required fields: patientUuid, serviceUuid, startDateTime
    risk: medium: books a Bahmni patient appointment; requires reverse ETL approval
  create_drug_order:
    endpoint: POST /ws/rest/v1/order
    required fields: patient, concept, careSetting
    risk: critical: places a medication (drug) order for a patient; requires reverse ETL approval and destructive confirmation
  create_diagnosis:
    endpoint: POST /ws/rest/v1/bahmnicore/diagnosis
    required fields: patientUuid, codedAnswer
    risk: critical: records a clinical diagnosis for a patient; requires reverse ETL approval and destructive confirmation
  create_observations_bulk:
    endpoint: POST /ws/rest/v1/bahmnicore/observations
    required fields: observations
    risk: critical: uploads a bounded top-level JSON array of clinical observations; use only reviewed observation payloads with reverse ETL approval and destructive confirmation
  upload_patient_document:
    endpoint: POST /ws/rest/v1/attachment
    required fields: document_file_path, patient_uuid
    risk: critical: uploads a patient document from a bounded local file path to Bahmni/OpenMRS; local file path is redacted in plans and the action requires reverse ETL approval and destructive confirmation

SECURITY
  read risk: external Bahmni/OpenMRS clinical PHI read of patient, encounter, observation, visit, order, lab, appointment, and diagnosis data; reads are bounded and secret-shaped fields are redacted by existing output policies, but clinical PHI fields are not generally field-redacted by the current engine
  write risk: typed Bahmni/OpenMRS reverse ETL clinical mutations for patients, encounters, observations, appointments, drug orders, diagnoses, a bounded top-level JSON array observation upload, and a bounded multipart patient-document upload
  approval: reverse ETL writes require plan, preview, approval, execute; clinical/destructive actions require --confirm destructive; multipart document upload binds a SHA-256 payload identity that is snapshot-verified before network send and redacts the local file path in plans
  Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

COMMAND SURFACE
  Inspect, read, and safely plan typed Bahmni clinical operations.
  Usage: pm bahmni <command> [flags]
  Source CLI: Bahmni / OpenMRS REST + FHIR2 (OpenMRS REST v1, Bahmni-core REST, OpenMRS FHIR2 R4)
  Global flags:
    --credential (string): Credential name to use for the Bahmni request.
    --connection (string): Alias for --credential.
    --config (string_array): Connector config override as key=value.
    --json (boolean): Emit machine-readable JSON output.
    --limit (integer): Maximum ETL records to emit for stream commands.
    --max-bytes (integer): Maximum direct-read response bytes; typed operations declare their own lower cap.
    --plan (string): Execute an approved reverse-ETL plan by id.
    --preview (boolean): Preview a reverse-ETL write command without making a network mutation.
    --approve (string): Approval token required to execute a reverse-ETL plan.
    --confirm (string): Typed confirmation challenge for destructive reverse-ETL writes.
  Clinical data
    patients list - List Bahmni/OpenMRS patients as ETL records (requires a patient search term). [intent=etl availability=implemented stream=patients]
    patients create - Create a Bahmni/OpenMRS patient record (clinical PHI). [intent=reverse_etl availability=partial write=create_patient]; approval: reverse ETL plan -> preview -> approval -> execute; clinical/destructive actions require --confirm destructive.; risk: clinical reverse-ETL mutation of Bahmni/OpenMRS PHI; runs only through plan, preview, approval, execute.; notes: Typed reverse-ETL action; no raw HTTP body, raw JSON, generic shell, or SQL write is exposed.; flags: --identifiers, --person
    patients update - Update an existing Bahmni/OpenMRS patient record (clinical PHI). [intent=reverse_etl availability=partial write=update_patient]; approval: reverse ETL plan -> preview -> approval -> execute; clinical/destructive actions require --confirm destructive.; risk: clinical reverse-ETL mutation of Bahmni/OpenMRS PHI; runs only through plan, preview, approval, execute.; notes: Typed reverse-ETL action; no raw HTTP body, raw JSON, generic shell, or SQL write is exposed.; flags: --uuid, --person
    encounters list - List patient encounters as ETL records (scoped by patient_uuid). [intent=etl availability=implemented stream=encounters]
    encounters create - Create a clinical encounter with linked observations. [intent=reverse_etl availability=partial write=create_encounter]; approval: reverse ETL plan -> preview -> approval -> execute; clinical/destructive actions require --confirm destructive.; risk: clinical reverse-ETL mutation of Bahmni/OpenMRS PHI; runs only through plan, preview, approval, execute.; notes: Typed reverse-ETL action; no raw HTTP body, raw JSON, generic shell, or SQL write is exposed.; flags: --patient, --encounter-type
    observations list - List patient observations (obs) as ETL records (scoped by patient_uuid). [intent=etl availability=implemented stream=observations]
    observations create - Record a clinical observation (obs) for a patient. [intent=reverse_etl availability=partial write=create_observation]; approval: reverse ETL plan -> preview -> approval -> execute; clinical/destructive actions require --confirm destructive.; risk: clinical reverse-ETL mutation of Bahmni/OpenMRS PHI; runs only through plan, preview, approval, execute.; notes: Typed reverse-ETL action; no raw HTTP body, raw JSON, generic shell, or SQL write is exposed.; flags: --person, --concept, --value
    observations upload-bulk - Upload a bounded top-level JSON array of Bahmni observations. [intent=reverse_etl availability=partial write=create_observations_bulk]; approval: reverse ETL plan -> preview -> approval -> execute; clinical/destructive actions require --confirm destructive.; risk: clinical reverse-ETL mutation of Bahmni/OpenMRS PHI; runs only through plan, preview, approval, execute.; notes: Typed reverse-ETL action; no raw HTTP body, raw JSON, generic shell, or SQL write is exposed.; flags: --observations
    visits list - List patient visits as ETL records (scoped by patient_uuid). [intent=etl availability=implemented stream=visits]
    diagnoses list - List Bahmni-core patient diagnoses as ETL records (scoped by patient_uuid). [intent=etl availability=implemented stream=diagnoses]
    diagnoses create - Record a clinical diagnosis for a patient. [intent=reverse_etl availability=partial write=create_diagnosis]; approval: reverse ETL plan -> preview -> approval -> execute; clinical/destructive actions require --confirm destructive.; risk: clinical reverse-ETL mutation of Bahmni/OpenMRS PHI; runs only through plan, preview, approval, execute.; notes: Typed reverse-ETL action; no raw HTTP body, raw JSON, generic shell, or SQL write is exposed.; flags: --patient-uuid, --coded-answer
  Catalog & reference
    concepts list - List OpenMRS concepts as ETL records. [intent=etl availability=implemented stream=concepts]
    locations list - List OpenMRS locations as ETL records. [intent=etl availability=implemented stream=locations]
    providers list - List OpenMRS providers as ETL records. [intent=etl availability=implemented stream=providers]
  Orders & labs
    drug_orders list - List active Bahmni drug orders as ETL records (scoped by patient_uuid). [intent=etl availability=implemented stream=drug_orders]
    drug_orders create - Place a medication (drug) order for a patient. [intent=reverse_etl availability=partial write=create_drug_order]; approval: reverse ETL plan -> preview -> approval -> execute; clinical/destructive actions require --confirm destructive.; risk: clinical reverse-ETL mutation of Bahmni/OpenMRS PHI; runs only through plan, preview, approval, execute.; notes: Typed reverse-ETL action; no raw HTTP body, raw JSON, generic shell, or SQL write is exposed.; flags: --patient, --concept, --care-setting
    lab_orders list - List lab (test) orders as ETL records (scoped by patient_uuid). [intent=etl availability=implemented stream=lab_orders]
    lab_results list - List Bahmni-core lab result observations as ETL records (scoped by patient_uuid). [intent=etl availability=implemented stream=lab_results]
  Scheduling
    appointments list - List Bahmni appointments as ETL records. [intent=etl availability=implemented stream=appointments]
    appointments create - Book a Bahmni patient appointment. [intent=reverse_etl availability=partial write=create_appointment]; approval: reverse ETL plan -> preview -> approval -> execute; clinical/destructive actions require --confirm destructive.; risk: clinical reverse-ETL mutation of Bahmni/OpenMRS PHI; runs only through plan, preview, approval, execute.; notes: Typed reverse-ETL action; no raw HTTP body, raw JSON, generic shell, or SQL write is exposed.; flags: --patient-uuid, --service-uuid, --start-date-time
  Direct reads
    patient get - Retrieve a single patient resource by UUID. [intent=direct_read availability=implemented]; risk: bounded Bahmni/OpenMRS JSON read; the response is size-limited and secret-shaped fields are redacted.; flags: --uuid
    encounter get - Retrieve a single encounter resource by UUID. [intent=direct_read availability=implemented]; risk: bounded Bahmni/OpenMRS JSON read; the response is size-limited and secret-shaped fields are redacted.; flags: --uuid
    visit get - Retrieve a single visit resource by UUID. [intent=direct_read availability=implemented]; risk: bounded Bahmni/OpenMRS JSON read; the response is size-limited and secret-shaped fields are redacted.; flags: --uuid
    concept get - Retrieve a single concept resource by UUID. [intent=direct_read availability=implemented]; risk: bounded Bahmni/OpenMRS JSON read; the response is size-limited and secret-shaped fields are redacted.; flags: --uuid
    provider get - Retrieve a single provider resource by UUID. [intent=direct_read availability=implemented]; risk: bounded Bahmni/OpenMRS JSON read; the response is size-limited and secret-shaped fields are redacted.; flags: --uuid
    location get - Retrieve a single location resource by UUID. [intent=direct_read availability=implemented]; risk: bounded Bahmni/OpenMRS JSON read; the response is size-limited and secret-shaped fields are redacted.; flags: --uuid
    fhir patient-read - Read a FHIR R4 Patient resource by id. [intent=direct_read availability=implemented]; risk: bounded Bahmni/OpenMRS JSON read; the response is size-limited and secret-shaped fields are redacted.; flags: --id
    fhir observation-read - Read a FHIR R4 Observation resource by id. [intent=direct_read availability=implemented]; risk: bounded Bahmni/OpenMRS JSON read; the response is size-limited and secret-shaped fields are redacted.; flags: --id
    fhir encounter-read - Read a FHIR R4 Encounter resource by id. [intent=direct_read availability=implemented]; risk: bounded Bahmni/OpenMRS JSON read; the response is size-limited and secret-shaped fields are redacted.; flags: --id
    fhir condition-read - Read a FHIR R4 Condition resource by id. [intent=direct_read availability=implemented]; risk: bounded Bahmni/OpenMRS JSON read; the response is size-limited and secret-shaped fields are redacted.; flags: --id
    bahmnicore patient-detail - Read a consolidated Bahmni-core patient profile by UUID. [intent=direct_read availability=implemented]; risk: bounded Bahmni/OpenMRS JSON read; the response is size-limited and secret-shaped fields are redacted.; flags: --uuid
    bahmnicore patient-search - Schema-gated typed Bahmni patient search (POST /ws/rest/v1/bahmnicore/search/patient). [intent=direct_read availability=implemented]; approval: none: read-only POST query with a schema-gated body.; risk: bounded typed POST read-query; the response is size-capped and PHI/identifier fields are redacted.; notes: Executes through the typed operation direct-read engine; no raw request body or generic HTTP flag is exposed.; flags: --q, --identifier, --address-field-name, --address-field-value, --login-location-uuid, --start-index
  Documents
    documents upload - Upload a bounded local patient document via multipart. [intent=reverse_etl availability=partial write=upload_patient_document]; approval: reverse ETL plan -> preview -> approval -> execute; clinical/destructive actions require --confirm destructive.; risk: clinical reverse-ETL mutation of Bahmni/OpenMRS PHI; runs only through plan, preview, approval, execute.; notes: Typed reverse-ETL action; no raw HTTP body, raw JSON, generic shell, or SQL write is exposed.; flags: --document-file-path, --patient-uuid, --caption
  Help topics:
    bahmni-auth - Point base_url at a Bahmni OpenMRS instance, including a local Bahmni/bahmni-docker deployment, and supply username/password via credentials; never pass secrets in command text.
    bahmni-writes - Bahmni clinical mutations are typed reverse-ETL actions with plan, preview, approval, execute gates; clinical/destructive actions require --confirm destructive.
    bahmni-direct-read - Bahmni direct reads are bounded JSON GET-by-uuid, FHIR read-by-id, or a schema-gated typed POST patient search, all with a response byte cap; clinical PHI fields are not generally field-redacted by the current engine.
    bahmni-phi - Bahmni reads and writes can include clinical PHI. The current runtime bounds output and redacts secret-shaped fields/file-path inputs, but broad clinical PHI field redaction remains a separate engine policy decision.

EXAMPLES
  # Inspect as a manual
  pm connectors inspect bahmni

  # Inspect as structured JSON
  pm connectors inspect bahmni --json

AGENT WORKFLOW
  - Run pm connectors inspect bahmni before creating credentials or plans.
  - Use --json only when the caller needs structured output; use the manual for human-readable guidance.
  - Never ask the user to paste secret values into chat.
  - For reverse ETL writes, create a plan, show the preview, wait for explicit approval, then run with the approval token.

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
