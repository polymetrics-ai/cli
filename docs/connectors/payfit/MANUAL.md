# pm connectors inspect payfit

```text
NAME
  pm connectors inspect payfit - PayFit connector manual

SYNOPSIS
  pm connectors inspect payfit
  pm connectors inspect payfit --json
  pm credentials add <name> --connector payfit [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Reads PayFit legacy /v1 resources and current company-scoped PayFit API resources; writes supported JSON customer-key mutations.

ICON
  id: pm-sample
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
  absence_id
  base_url
  collaborator_id
  company_id
  contract_id
  document_id
  limit
  mode
  pay_period
  payslip_id
  api_key (secret) (required)

ETL STREAMS
  employees:
    primary key: id
    cursor: updated_at
    fields: email(string), first_name(string), id(string), last_name(string), updated_at(string)
  contracts:
    primary key: id
    cursor: updated_at
    fields: employee_id(string), id(string), start_date(string), type(string), updated_at(string)
  companies:
    primary key: id
    cursor: updated_at
    fields: country(string), id(string), name(string), updated_at(string)
  absences:
    primary key: id
    fields: id(string)
  collaborators:
    primary key: id
    fields: id(string)
  collaborator_meal_vouchers:
    primary key: collaboratorId
    fields: collaboratorId(string)
  collaborator_payslips:
    primary key: payslipId
    fields: payslipId(string)
  company_contracts:
    primary key: contractId
    fields: contractId(string)
  company_contracts_fr:
    primary key: contractId
    fields: contractId(string)
  worked_time_by_contract:
    primary key: contractId
    fields: contractId(string)
  health_insurance_contracts:
    primary key: idContrat
    fields: idContrat(string)
  provident_fund_contracts:
    primary key: idContrat
    fields: idContrat(string)
  auto_enrolment_documents:
    primary key: documentId
    fields: documentId(string)
  income_tax_documents:
    primary key: documentId
    fields: documentId(string)
  accounting_v2_entries:
    primary key: operationDate, accountId, contractId
    fields: accountId(string), contractId(string), operationDate(string)
  company:
    primary key: id
    fields: id(string)
  company_fr:
    primary key: id
    fields: id(string)
  collaborator:
    primary key: id
    fields: id(string)
  company_contract:
    primary key: contractId
    fields: contractId(string)
  company_contract_fr:
    primary key: contractId
    fields: contractId(string)
  payroll_status:
    primary key: company_id
    fields: company_id(string)

SYNC MODES
  ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped

REVERSE ETL ACTIONS
  create_absence:
    endpoint: POST /companies/{{ record.company_id }}/absences
    required fields: company_id, contractId, type, startDate, endDate
    risk: external PayFit mutation; create absence; approval required
  cancel_absence:
    endpoint: DELETE /companies/{{ record.company_id }}/absences/{{ record.absence_id }}
    required fields: company_id, absence_id
    risk: external mutation; cancels an existing PayFit absence; approval required
  create_collaborator:
    endpoint: POST /companies/{{ record.company_id }}/collaborators
    required fields: company_id, firstName, lastName, personalEmail
    risk: external PayFit mutation; create collaborator; approval required
  create_contract:
    endpoint: POST /companies/{{ record.company_id }}/collaborators/{{ record.collaborator_id }}/contracts
    required fields: company_id, collaborator_id, jobTitle, startDate
    risk: external PayFit mutation; create contract; approval required
  update_contract_health_insurance:
    endpoint: PUT /companies/{{ record.company_id }}/contracts-fr/{{ record.contract_id }}/health-insurance
    required fields: company_id, contract_id, healthInsuranceContractIds
    risk: external PayFit mutation; update contract health insurance; approval required
  update_contract_provident_fund:
    endpoint: PUT /companies/{{ record.company_id }}/contracts-fr/{{ record.contract_id }}/provident-fund
    required fields: company_id, contract_id, providentFundContractIds
    risk: external PayFit mutation; update contract provident fund; approval required
  request_health_insurance_regularization:
    endpoint: POST /companies/{{ record.company_id }}/contracts-fr/{{ record.contract_id }}/regularization
    required fields: company_id, contract_id, healthInsuranceContractIds, effectiveDate
    risk: external PayFit mutation; request health insurance regularization; approval required

SECURITY
  read risk: external PayFit API read of HR, contract, payroll-status, absence, accounting, and document-metadata data
  write risk: external PayFit API mutations for collaborator, contract, absence, and health-insurance/provident-fund workflows
  approval: write actions require explicit reverse-ETL approval; absence cancellation is idempotent-delete modeled
  Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

EXAMPLES
  # Inspect as a manual
  pm connectors inspect payfit

  # Inspect as structured JSON
  pm connectors inspect payfit --json

AGENT WORKFLOW
  - Run pm connectors inspect payfit before creating credentials or plans.
  - Use --json only when the caller needs structured output; use the manual for human-readable guidance.
  - Never ask the user to paste secret values into chat.
  - For reverse ETL writes, create a plan, show the preview, wait for explicit approval, then run with the approval token.

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
