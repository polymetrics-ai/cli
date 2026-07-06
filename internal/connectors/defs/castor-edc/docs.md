# Overview

Reads Castor EDC studies, users, countries, and audit-trail events through the Castor EDC OAuth2
REST API.

Readable streams: `study`, `user`, `country`, `audit_trail`, `records`, `fields`,
`field_dependencies`, `field_optiongroups`, `field_validations`, `sites`, `study_metadata`,
`metadata_types`, `phases`, `queries`, `reports`, `report_instances`, `roles`, `steps`, `surveys`,
`survey_packages`, `survey_package_instances`, `study_users`, `verifications`, `record_progress`.

Write actions: `create_record`, `create_site`, `create_role`, `create_survey_package_instance`,
`create_report_instance`, `create_randomization`.

Service API documentation: https://data.castoredc.com/api.

## Auth setup

Connection fields:

- `base_url` (optional, string); default `https://data.castoredc.com/api`; format `uri`; Castor EDC
  API base URL override for a regional host, tests, or proxies.
- `client_id` (required, secret, string); Castor EDC OAuth2 client-credentials client ID. Never
  logged.
- `client_secret` (required, secret, string); Castor EDC OAuth2 client-credentials client secret.
  Never logged.
- `study_id` (optional, string); Castor EDC study id to scope every study-level resource (records,
  fields, forms, visits/phases, sites, surveys, survey packages, roles, steps, metadata, queries,
  verifications, and the create_* write actions). Optional: the study/user/country/audit_trail
  streams and the account-level users stream do not require it.
- `token_url` (optional, string); default `https://data.castoredc.com/oauth/token`; format `uri`.

Secret fields are redacted in logs and write previews: `client_id`, `client_secret`.

Default configuration values: `base_url=https://data.castoredc.com/api`,
`token_url=https://data.castoredc.com/oauth/token`.

Authentication behavior:

- OAuth 2.0 client credentials authentication using `config.token_url`, `secrets.client_id`,
  `secrets.client_secret`.

Requests use the configured `base_url` value after applying defaults.

Connection checks call GET `/user`.

## Streams notes

Default pagination: page-number pagination; page parameter `page`; size parameter `page_size`;
starts at 1; page size 100.

Pagination by stream: none: `country`, `roles`, `study_users`; page_number: `study`, `user`,
`audit_trail`, `records`, `fields`, `field_dependencies`, `field_optiongroups`, `field_validations`,
`sites`, `study_metadata`, `metadata_types`, `phases`, `queries`, `reports`, `report_instances`,
`steps`, `surveys`, `survey_packages`, `survey_package_instances`, `verifications`,
`record_progress`.

Incremental streams use their declared cursor fields and send lower-bound parameters only when a
lower bound is available.

- `study`: GET `/study` - records path `_embedded.study`; page-number pagination; page parameter
  `page`; size parameter `page_size`; starts at 1; page size 100; incremental cursor `updated_on`;
  formatted as `rfc3339`.
- `user`: GET `/user` - records path `_embedded.user`; page-number pagination; page parameter
  `page`; size parameter `page_size`; starts at 1; page size 100; incremental cursor `last_login`;
  formatted as `rfc3339`.
- `country`: GET `/country` - records path `_embedded.countries`.
- `audit_trail`: GET `/audit-trail` - records path `_embedded.audit_trail`; page-number pagination;
  page parameter `page`; size parameter `page_size`; starts at 1; page size 100; incremental cursor
  `datetime`; formatted as `rfc3339`.
- `records`: GET `/study/{{ config.study_id }}/record` - records path `_embedded.records`;
  page-number pagination; page parameter `page`; size parameter `page_size`; starts at 1; page size
  100.
- `fields`: GET `/study/{{ config.study_id }}/field` - records path `_embedded.fields`; query
  `include`=`metadata,validations,optiongroup`; page-number pagination; page parameter `page`; size
  parameter `page_size`; starts at 1; page size 100.
- `field_dependencies`: GET `/study/{{ config.study_id }}/field-dependency` - records path
  `_embedded.fieldDependencies`; page-number pagination; page parameter `page`; size parameter
  `page_size`; starts at 1; page size 100.
- `field_optiongroups`: GET `/study/{{ config.study_id }}/field-optiongroup` - records path
  `_embedded.fieldOptionGroups`; page-number pagination; page parameter `page`; size parameter
  `page_size`; starts at 1; page size 100.
- `field_validations`: GET `/study/{{ config.study_id }}/field-validation` - records path
  `_embedded.fieldValidations`; page-number pagination; page parameter `page`; size parameter
  `page_size`; starts at 1; page size 100.
- `sites`: GET `/study/{{ config.study_id }}/site` - records path `_embedded.sites`; page-number
  pagination; page parameter `page`; size parameter `page_size`; starts at 1; page size 100.
- `study_metadata`: GET `/study/{{ config.study_id }}/metadata` - records path
  `_embedded.metadatas`; page-number pagination; page parameter `page`; size parameter `page_size`;
  starts at 1; page size 100.
- `metadata_types`: GET `/study/{{ config.study_id }}/metadatatype` - records path
  `_embedded.metadatatypes`; page-number pagination; page parameter `page`; size parameter
  `page_size`; starts at 1; page size 100.
- `phases`: GET `/study/{{ config.study_id }}/phase` - records path `_embedded.phases`; page-number
  pagination; page parameter `page`; size parameter `page_size`; starts at 1; page size 100.
- `queries`: GET `/study/{{ config.study_id }}/query` - records path `_embedded.queries`;
  page-number pagination; page parameter `page`; size parameter `page_size`; starts at 1; page size
  100.
- `reports`: GET `/study/{{ config.study_id }}/report` - records path `_embedded.reports`;
  page-number pagination; page parameter `page`; size parameter `page_size`; starts at 1; page size
  100.
- `report_instances`: GET `/study/{{ config.study_id }}/report-instance` - records path
  `_embedded.reportInstances`; query `archived`=`0`; page-number pagination; page parameter `page`;
  size parameter `page_size`; starts at 1; page size 100.
- `roles`: GET `/study/{{ config.study_id }}/role` - records path `_embedded.roles`.
- `steps`: GET `/study/{{ config.study_id }}/step` - records path `_embedded.steps`; page-number
  pagination; page parameter `page`; size parameter `page_size`; starts at 1; page size 100.
- `surveys`: GET `/study/{{ config.study_id }}/survey` - records path `_embedded.surveys`;
  page-number pagination; page parameter `page`; size parameter `page_size`; starts at 1; page size
  100.
- `survey_packages`: GET `/study/{{ config.study_id }}/surveypackage` - records path
  `_embedded.survey_packages`; page-number pagination; page parameter `page`; size parameter
  `page_size`; starts at 1; page size 100.
- `survey_package_instances`: GET `/study/{{ config.study_id }}/surveypackageinstance` - records
  path `_embedded.surveypackageinstance`; page-number pagination; page parameter `page`; size
  parameter `page_size`; starts at 1; page size 100.
- `study_users`: GET `/study/{{ config.study_id }}/user` - records path `_embedded.studyUsers`.
- `verifications`: GET `/study/{{ config.study_id }}/verification` - records path
  `_embedded.verifications`; page-number pagination; page parameter `page`; size parameter
  `page_size`; starts at 1; page size 100.
- `record_progress`: GET `/study/{{ config.study_id }}/record-progress/steps` - records path
  `_embedded.records`; page-number pagination; page parameter `page`; size parameter `page_size`;
  starts at 1; page size 100.

## Write actions & risks

Overall write risk: external mutations creating clinical-trial study participant records, sites,
roles, survey-package invitations, report instances, and record randomization; every write action
requires approval.

Reverse ETL writes should be planned, previewed, approved, and then executed. Declared actions:

- `create_record`: POST `/study/{{ config.study_id }}/record` - kind `create`; body type `json`;
  required record fields `institute_id`, `email_address`; accepted fields `ccr_record_id`,
  `email_address`, `institute_id`, `record_id`; risk: external mutation creating a new
  clinical-trial study participant record; approval required.
- `create_site`: POST `/study/{{ config.study_id }}/site` - kind `create`; body type `json`;
  required record fields `name`, `abbreviation`, `code`, `country_id`; accepted fields
  `abbreviation`, `code`, `country_id`, `name`; risk: external mutation creating a new study
  site/institute; approval required.
- `create_role`: POST `/study/{{ config.study_id }}/role` - kind `create`; body type `json`;
  required record fields `name`, `description`, `permissions`; accepted fields `description`,
  `name`, `permissions`; risk: external mutation creating a new study access-control role; approval
  required.
- `create_survey_package_instance`: POST `/study/{{ config.study_id }}/surveypackageinstance` - kind
  `create`; body type `json`; required record fields `survey_package_id`, `record_id`,
  `email_address`; accepted fields `auto_lock_on_finish`, `auto_send`, `ccr_patient_id`,
  `email_address`, `package_invitation`, `package_invitation_subject`, `record_id`,
  `survey_package_id`; risk: external mutation dispatching a survey package invitation to a study
  participant; approval required.
- `create_report_instance`: POST `/study/{{ config.study_id }}/record/{{ record.record_id
  }}/report-instance` - kind `create`; body type `json`; path fields `record_id`; required record
  fields `record_id`, `report_id`; accepted fields `parent_id`, `record_id`, `report_id`,
  `report_name_custom`; risk: external mutation creating a new report instance for a study
  participant record; approval required.
- `create_randomization`: POST `/study/{{ config.study_id }}/record/{{ record.record_id
  }}/randomization` - kind `create`; body type `none`; path fields `record_id`; required record
  fields `record_id`; accepted fields `record_id`; risk: irreversible external mutation randomizing
  a clinical-trial study participant; approval required.

## Known limits

- Batch defaults: read_page_size=100.
- API coverage includes 24 stream-backed endpoint group(s), 6 write-backed endpoint group(s).
- Other documented endpoints are not exposed by this connector where they are classified as
  binary_payload=3, destructive_admin=1, duplicate_of=24, non_data_endpoint=2, out_of_scope=36,
  requires_elevated_scope=2.
