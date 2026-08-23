---
name: pm-castor-edc
description: Castor EDC connector knowledge and safe action guide.
---

# pm-castor-edc

## Purpose

Reads Castor EDC studies, users, countries, and audit-trail events through the Castor EDC OAuth2 REST API.

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

- base_url
- study_id
- token_url
- client_id (secret) (required)
- client_secret (secret) (required)

## ETL Streams

- study:
  - primary key: study_id
  - cursor: updated_on
  - fields: created_on(string), crf_id(string), duration(integer), gcp_enabled(boolean), institute_id(string), live(boolean), main_contact(string), name(string), premium_support_enabled(boolean), randomization_enabled(boolean), study_id(string), surveys_enabled(boolean), updated_on(string)
- user:
  - primary key: id
  - cursor: last_login
  - fields: email_address(string), first_name(string), full_name(string), id(string), institute(string), is_active(boolean), last_login(string), last_name(string), user_id(string)
- country:
  - primary key: id
  - fields: country_cca2(string), country_cca3(string), country_id(string), country_name(string), country_tld(string), id(integer)
- audit_trail:
  - primary key: uuid
  - cursor: datetime
  - fields: datetime(string), event_details(object), event_type(string), user_email(string), user_id(string), user_name(string), uuid(string)
- records:
  - primary key: record_id
  - fields: archived(boolean), archived_reason(string), ccr_patient_id(string), created_on(string), email_address(string), institute_id(string), locked(boolean), progress(integer), randomization_datetime(string), randomization_group(string), randomization_group_name(string), record_id(string), status(string), updated_on(string)
- fields:
  - primary key: id
  - fields: additional_config(string), created_on(string), exclude_on_index(boolean), field_hidden(boolean), field_id(string), field_info(string), field_label(string), field_length(integer), field_max(number), field_min(number), field_name(string), field_number(integer), field_required(boolean), field_slider_step(number), field_summary_template(string), field_type(string), field_units(string), id(string), parent_id(string), report_id(string), step_id(string), updated_on(string)
- field_dependencies:
  - primary key: id
  - fields: child_field_id(string), id(string), operator(string), parent_field_id(string), value(string)
- field_optiongroups:
  - primary key: id
  - fields: id(string), layout(boolean), name(string), options(array)
- field_validations:
  - primary key: id
  - fields: field_id(string), id(string), text(string), type(string), value(string)
- sites:
  - primary key: id
  - fields: abbreviation(string), code(string), country_id(string), id(string), name(string), number_of_records(integer)
- study_metadata:
  - primary key: id
  - fields: element_type(string), external_field_id(string), external_metadatatype_id(string), id(string), parent_id(string)
- metadata_types:
  - primary key: id
  - fields: description(string), id(string), name(string)
- phases:
  - primary key: id
  - fields: description(string), duration(integer), id(string), name(string), phase_order(integer)
- queries:
  - primary key: id
  - fields: created_by(string), created_on(string), field_id(string), id(string), instance_id(string), query_text(string), record_id(string), status(string), updated_on(string)
- reports:
  - primary key: id
  - fields: description(string), id(string), name(string), type(string)
- report_instances:
  - primary key: id
  - fields: archived(boolean), created_on(string), id(string), name(string), name_custom(string), parent_id(string), record_id(string), report_id(string), updated_on(string)
- roles:
  - primary key: id
  - fields: description(string), id(string), name(string), permissions(object)
- steps:
  - primary key: id
  - fields: id(string), name(string), phase_id(string), step_description(string), step_order(integer)
- surveys:
  - primary key: id
  - fields: description(string), id(string), intro_text(string), name(string), outro_text(string)
- survey_packages:
  - primary key: id
  - fields: auto_lock_on_finish(boolean), auto_send(boolean), description(string), expire_after_hours(integer), id(string), name(string)
- survey_package_instances:
  - primary key: id
  - fields: ccr_patient_id(string), created_on(string), email_address(string), finished_on(string), id(string), locked(boolean), progress(integer), record_id(string), started_on(string), survey_package_id(string), updated_on(string)
- study_users:
  - primary key: id
  - fields: email_address(string), first_name(string), full_name(string), id(string), institute_roles(array), last_name(string), manage_permission(string)
- verifications:
  - primary key: id
  - fields: entity_id(string), entity_type(string), id(string), record_id(string), verification_type(string), verified_by(string), verified_on(string)
- record_progress:
  - primary key: record_id
  - fields: progress(integer), record_id(string), steps(array)

## Sync Modes

- ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped, incremental_append, incremental_append_deduped

## Reverse ETL Actions

- create_record:
  - endpoint: POST /study/{{ config.study_id }}/record
  - required fields: institute_id, email_address
  - risk: external mutation creating a new clinical-trial study participant record; approval required
- create_site:
  - endpoint: POST /study/{{ config.study_id }}/site
  - required fields: name, abbreviation, code, country_id
  - risk: external mutation creating a new study site/institute; approval required
- create_role:
  - endpoint: POST /study/{{ config.study_id }}/role
  - required fields: name, description, permissions
  - risk: external mutation creating a new study access-control role; approval required
- create_survey_package_instance:
  - endpoint: POST /study/{{ config.study_id }}/surveypackageinstance
  - required fields: survey_package_id, record_id, email_address
  - risk: external mutation dispatching a survey package invitation to a study participant; approval required
- create_report_instance:
  - endpoint: POST /study/{{ config.study_id }}/record/{{ record.record_id }}/report-instance
  - required fields: record_id, report_id
  - risk: external mutation creating a new report instance for a study participant record; approval required
- create_randomization:
  - endpoint: POST /study/{{ config.study_id }}/record/{{ record.record_id }}/randomization
  - required fields: record_id
  - risk: irreversible external mutation randomizing a clinical-trial study participant; approval required

## Security

- read risk: external Castor EDC API read of clinical-trial study/user/audit-trail/record/field/form/survey/site/role data
- write risk: external mutations creating clinical-trial study participant records, sites, roles, survey-package invitations, report instances, and record randomization; every write action requires approval
- approval: read: none; write: required for every action
- Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

## Commands

### Inspect as a manual

```bash
pm connectors inspect castor-edc
```

### Inspect as structured JSON

```bash
pm connectors inspect castor-edc --json
```

## Agent Rules

- Run pm connectors inspect castor-edc before creating credentials or plans.
- Use --json only when the caller needs structured output; use the manual for human-readable guidance.
- Never ask the user to paste secret values into chat.
- For reverse ETL writes, create a plan, show the preview, wait for explicit approval, then run with the approval token.
