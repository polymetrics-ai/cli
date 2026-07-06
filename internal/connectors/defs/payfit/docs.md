# Overview

PayFit reads 21 stream(s), and writes through 7 action(s).

Readable streams: `employees`, `contracts`, `companies`, `absences`, `collaborators`,
`collaborator_meal_vouchers`, `collaborator_payslips`, `company_contracts`, `company_contracts_fr`,
`worked_time_by_contract`, `health_insurance_contracts`, `provident_fund_contracts`,
`auto_enrolment_documents`, `income_tax_documents`, `accounting_v2_entries`, `company`,
`company_fr`, `collaborator`, `company_contract`, `company_contract_fr`, `payroll_status`.

Write actions: `create_absence`, `cancel_absence`, `create_collaborator`, `create_contract`,
`update_contract_health_insurance`, `update_contract_provident_fund`,
`request_health_insurance_regularization`.

Service API documentation: https://developers.payfit.io/.

## Auth setup

Connection fields:

- `absence_id` (optional, string); Absence ID for cancellation writes.
- `api_key` (required, secret, string); PayFit partner API key. Used only for Bearer auth; never
  logged.
- `base_url` (optional, string); default `https://partner-api.payfit.com`; format `uri`; PayFit API
  base URL override.
- `collaborator_id` (optional, string); Collaborator ID for detail streams and contract creation.
- `company_id` (optional, string); PayFit company ID used by company-scoped current API endpoints.
  Discover it with the documented oauth.payfit.com/introspect helper outside this connector.
- `contract_id` (optional, string); Contract ID for detail streams and contract-scoped writes.
- `document_id` (optional, string); Document ID for binary document download endpoints, which are
  excluded from streams.
- `limit` (optional, string); default `100`; Records per page (1-100).
- `mode` (optional, string).
- `pay_period` (optional, string); YYYYMM pay period/date value required by PayFit accounting,
  worked-time, payroll-status, and meal-voucher endpoints.
- `payslip_id` (optional, string); Payslip ID for binary payslip download endpoints, which are
  excluded from streams.

Secret fields are redacted in logs and write previews: `api_key`.

Default configuration values: `base_url=https://partner-api.payfit.com`, `limit=100`.

Authentication behavior:

- Bearer token authentication using `secrets.api_key`.

Requests use the configured `base_url` value after applying defaults.

Connection checks call GET `/v1/employees` with query `limit`=`1`.

## Streams notes

Default pagination: cursor pagination; cursor parameter `offset`; next token from
`meta.next_offset`.

Pagination by stream: cursor: `employees`, `contracts`, `companies`, `absences`, `collaborators`,
`collaborator_meal_vouchers`, `company_contracts`, `company_contracts_fr`,
`worked_time_by_contract`; none: `collaborator_payslips`, `health_insurance_contracts`,
`provident_fund_contracts`, `auto_enrolment_documents`, `income_tax_documents`,
`accounting_v2_entries`, `company`, `company_fr`, `collaborator`, `company_contract`,
`company_contract_fr`, `payroll_status`.

- `employees`: GET `/v1/employees` - records path `data`; query `limit`=`{{ config.limit }}`; cursor
  pagination; cursor parameter `offset`; next token from `meta.next_offset`.
- `contracts`: GET `/v1/contracts` - records path `data`; query `limit`=`{{ config.limit }}`; cursor
  pagination; cursor parameter `offset`; next token from `meta.next_offset`.
- `companies`: GET `/v1/companies` - records path `data`; query `limit`=`{{ config.limit }}`; cursor
  pagination; cursor parameter `offset`; next token from `meta.next_offset`.
- `absences`: GET `/companies/{{ config.company_id }}/absences` - records path `absences`; query
  `maxResults`=`50`; cursor pagination; cursor parameter `nextPageToken`; next token from
  `meta.nextPageToken`; emits passthrough records.
- `collaborators`: GET `/companies/{{ config.company_id }}/collaborators` - records path
  `collaborators`; query `maxResults`=`50`; cursor pagination; cursor parameter `nextPageToken`;
  next token from `meta.nextPageToken`; emits passthrough records.
- `collaborator_meal_vouchers`: GET `/companies/{{ config.company_id }}/collaborators/meal-vouchers`
  - records path `mealVouchers`; query `date`=`{{ config.pay_period }}`; `maxResults`=`50`; cursor
  pagination; cursor parameter `nextPageToken`; next token from `meta.nextPageToken`; emits
  passthrough records.
- `collaborator_payslips`: GET `/companies/{{ config.company_id }}/collaborators/{{
  config.collaborator_id }}/payslips` - records path `payslips`; emits passthrough records.
- `company_contracts`: GET `/companies/{{ config.company_id }}/contracts` - records path
  `contracts`; query `maxResults`=`50`; cursor pagination; cursor parameter `nextPageToken`; next
  token from `meta.nextPageToken`; emits passthrough records.
- `company_contracts_fr`: GET `/companies/{{ config.company_id }}/contracts-fr` - records path
  `contracts`; query `maxResults`=`50`; cursor pagination; cursor parameter `nextPageToken`; next
  token from `meta.nextPageToken`; emits passthrough records.
- `worked_time_by_contract`: GET `/companies/{{ config.company_id }}/contracts/time` - records path
  `contracts`; query `date`=`{{ config.pay_period }}`; `maxResults`=`50`; cursor pagination; cursor
  parameter `nextPageToken`; next token from `meta.nextPageToken`; emits passthrough records.
- `health_insurance_contracts`: GET `/companies/{{ config.company_id }}/health-insurance-contracts`
  - records path `contracts`; emits passthrough records.
- `provident_fund_contracts`: GET `/companies/{{ config.company_id }}/provident-fund-contracts` -
  records path `contracts`; emits passthrough records.
- `auto_enrolment_documents`: GET `/companies/{{ config.company_id }}/auto-enrolment-documents` -
  records path `documents`; emits passthrough records.
- `income_tax_documents`: GET `/companies/{{ config.company_id }}/income-taxes-documents` - records
  path `documents`; emits passthrough records.
- `accounting_v2_entries`: GET `/companies/{{ config.company_id }}/accounting-v2` - records at
  response root; query `date`=`{{ config.pay_period }}`; emits passthrough records.
- `company`: GET `/companies/{{ config.company_id }}` - single-object response; records at response
  root; emits passthrough records.
- `company_fr`: GET `/companies-fr/{{ config.company_id }}` - single-object response; records at
  response root; emits passthrough records.
- `collaborator`: GET `/companies/{{ config.company_id }}/collaborators/{{ config.collaborator_id
  }}` - single-object response; records at response root; emits passthrough records.
- `company_contract`: GET `/companies/{{ config.company_id }}/contracts/{{ config.contract_id }}` -
  single-object response; records at response root; emits passthrough records.
- `company_contract_fr`: GET `/companies/{{ config.company_id }}/contracts-fr/{{ config.contract_id
  }}` - single-object response; records at response root; emits passthrough records.
- `payroll_status`: GET `/companies/{{ config.company_id }}/payroll-status` - single-object
  response; records at response root; query `date`=`{{ config.pay_period }}`; computed output fields
  `company_id`; emits passthrough records.

## Write actions & risks

Overall write risk: external PayFit API mutations for collaborator, contract, absence, and
health-insurance/provident-fund workflows.

Reverse ETL writes should be planned, previewed, approved, and then executed. Declared actions:

- `create_absence`: POST `/companies/{{ record.company_id }}/absences` - kind `create`; body type
  `json`; path fields `company_id`; required record fields `company_id`, `contractId`, `type`,
  `startDate`, `endDate`; accepted fields `company_id`, `contractId`, `endDate`, `startDate`,
  `type`; risk: external PayFit mutation; create absence; approval required.
- `cancel_absence`: DELETE `/companies/{{ record.company_id }}/absences/{{ record.absence_id }}` -
  kind `delete`; body type `json`; path fields `company_id`, `absence_id`; required record fields
  `company_id`, `absence_id`; accepted fields `absence_id`, `comment`, `company_id`; missing records
  treated as success for status `404`; risk: external mutation; cancels an existing PayFit absence;
  approval required.
- `create_collaborator`: POST `/companies/{{ record.company_id }}/collaborators` - kind `create`;
  body type `json`; path fields `company_id`; required record fields `company_id`, `firstName`,
  `lastName`, `personalEmail`; accepted fields `birthInformation`, `company_id`, `firstName`,
  `gender`, `inviteCollaborator`, `lastName`, `numberOfChildren`, `otherName`, `personalAddress`,
  `personalEmail`, `personalPhoneNumber`, `socialSecurityNumber`; risk: external PayFit mutation;
  create collaborator; approval required.
- `create_contract`: POST `/companies/{{ record.company_id }}/collaborators/{{
  record.collaborator_id }}/contracts` - kind `create`; body type `json`; path fields `company_id`,
  `collaborator_id`; required record fields `company_id`, `collaborator_id`, `jobTitle`,
  `startDate`; accepted fields `collaborator_id`, `company_id`, `jobTitle`, `startDate`; risk:
  external PayFit mutation; create contract; approval required.
- `update_contract_health_insurance`: PUT `/companies/{{ record.company_id }}/contracts-fr/{{
  record.contract_id }}/health-insurance` - kind `update`; body type `json`; path fields
  `company_id`, `contract_id`; required record fields `company_id`, `contract_id`,
  `healthInsuranceContractIds`; accepted fields `company_id`, `contract_id`, `employeeIsExempted`,
  `healthInsuranceContractIds`; risk: external PayFit mutation; update contract health insurance;
  approval required.
- `update_contract_provident_fund`: PUT `/companies/{{ record.company_id }}/contracts-fr/{{
  record.contract_id }}/provident-fund` - kind `update`; body type `json`; path fields `company_id`,
  `contract_id`; required record fields `company_id`, `contract_id`, `providentFundContractIds`;
  accepted fields `company_id`, `contract_id`, `providentFundContractIds`; risk: external PayFit
  mutation; update contract provident fund; approval required.
- `request_health_insurance_regularization`: POST `/companies/{{ record.company_id
  }}/contracts-fr/{{ record.contract_id }}/regularization` - kind `create`; body type `json`; path
  fields `company_id`, `contract_id`; required record fields `company_id`, `contract_id`,
  `healthInsuranceContractIds`, `effectiveDate`; accepted fields `company_id`, `contract_id`,
  `effectiveDate`, `healthInsuranceContractIds`; risk: external PayFit mutation; request health
  insurance regularization; approval required.

## Known limits

- Batch defaults: read_page_size=100.
- API coverage includes 21 stream-backed endpoint group(s), 7 write-backed endpoint group(s).
- Other documented endpoints are not exposed by this connector where they are classified as
  binary_payload=5, non_data_endpoint=2, requires_elevated_scope=1.
