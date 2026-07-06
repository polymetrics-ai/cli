---
name: pm-castor-edc
description: Castor EDC connector knowledge and safe action guide.
---

# pm-castor-edc

## Purpose

Reads Castor EDC studies, users, countries, and audit-trail events through the Castor EDC OAuth2 REST API.

## Icon

- asset: icons/pm-sample.svg
- source: polymetrics
- review_status: polymetrics

## Capabilities

- check=true catalog=true read=true write=true query=false
- Integration type: api

## Authentication

- Use pm credentials add with --from-env or --value-stdin for secret fields.

## Configuration

- base_url
- study_id
- token_url
- client_id (secret)
- client_secret (secret)

## ETL Streams

- study:
  - primary key: study_id
  - cursor: updated_on
  - fields: created_on(), crf_id(), duration(), gcp_enabled(), institute_id(), live(), main_contact(), name(), premium_support_enabled(), randomization_enabled(), study_id(), surveys_enabled(), updated_on()
- user:
  - primary key: id
  - cursor: last_login
  - fields: email_address(), first_name(), full_name(), id(), institute(), is_active(), last_login(), last_name(), user_id()
- country:
  - primary key: id
  - fields: country_cca2(), country_cca3(), country_id(), country_name(), country_tld(), id()
- audit_trail:
  - primary key: uuid
  - cursor: datetime
  - fields: datetime(), event_details(), event_type(), user_email(), user_id(), user_name(), uuid()
- records:
  - primary key: record_id
  - fields: archived(), archived_reason(), ccr_patient_id(), created_on(), email_address(), institute_id(), locked(), progress(), randomization_datetime(), randomization_group(), randomization_group_name(), record_id(), status(), updated_on()
- fields:
  - primary key: id
  - fields: additional_config(), created_on(), exclude_on_index(), field_hidden(), field_id(), field_info(), field_label(), field_length(), field_max(), field_min(), field_name(), field_number(), field_required(), field_slider_step(), field_summary_template(), field_type(), field_units(), id(), parent_id(), report_id(), step_id(), updated_on()
- field_dependencies:
  - primary key: id
  - fields: child_field_id(), id(), operator(), parent_field_id(), value()
- field_optiongroups:
  - primary key: id
  - fields: id(), layout(), name(), options()
- field_validations:
  - primary key: id
  - fields: field_id(), id(), text(), type(), value()
- sites:
  - primary key: id
  - fields: abbreviation(), code(), country_id(), id(), name(), number_of_records()
- study_metadata:
  - primary key: id
  - fields: element_type(), external_field_id(), external_metadatatype_id(), id(), parent_id()
- metadata_types:
  - primary key: id
  - fields: description(), id(), name()
- phases:
  - primary key: id
  - fields: description(), duration(), id(), name(), phase_order()
- queries:
  - primary key: id
  - fields: created_by(), created_on(), field_id(), id(), instance_id(), query_text(), record_id(), status(), updated_on()
- reports:
  - primary key: id
  - fields: description(), id(), name(), type()
- report_instances:
  - primary key: id
  - fields: archived(), created_on(), id(), name(), name_custom(), parent_id(), record_id(), report_id(), updated_on()
- roles:
  - primary key: id
  - fields: description(), id(), name(), permissions()
- steps:
  - primary key: id
  - fields: id(), name(), phase_id(), step_description(), step_order()
- surveys:
  - primary key: id
  - fields: description(), id(), intro_text(), name(), outro_text()
- survey_packages:
  - primary key: id
  - fields: auto_lock_on_finish(), auto_send(), description(), expire_after_hours(), id(), name()
- survey_package_instances:
  - primary key: id
  - fields: ccr_patient_id(), created_on(), email_address(), finished_on(), id(), locked(), progress(), record_id(), started_on(), survey_package_id(), updated_on()
- study_users:
  - primary key: id
  - fields: email_address(), first_name(), full_name(), id(), institute_roles(), last_name(), manage_permission()
- verifications:
  - primary key: id
  - fields: entity_id(), entity_type(), id(), record_id(), verification_type(), verified_by(), verified_on()
- record_progress:
  - primary key: record_id
  - fields: progress(), record_id(), steps()

## Sync Modes

- ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped, incremental_append, incremental_append_deduped

## Reverse ETL Actions

- create_record:
  - endpoint: POST /study/{{ config.study_id }}/record
  - risk: external mutation creating a new clinical-trial study participant record; approval required
- create_site:
  - endpoint: POST /study/{{ config.study_id }}/site
  - risk: external mutation creating a new study site/institute; approval required
- create_role:
  - endpoint: POST /study/{{ config.study_id }}/role
  - risk: external mutation creating a new study access-control role; approval required
- create_survey_package_instance:
  - endpoint: POST /study/{{ config.study_id }}/surveypackageinstance
  - risk: external mutation dispatching a survey package invitation to a study participant; approval required
- create_report_instance:
  - endpoint: POST /study/{{ config.study_id }}/record/{{ record.record_id }}/report-instance
  - required fields: record_id
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
