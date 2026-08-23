---
name: pm-payfit
description: PayFit connector knowledge and safe action guide.
---

# pm-payfit

## Purpose

Reads PayFit legacy /v1 resources and current company-scoped PayFit API resources; writes supported JSON customer-key mutations.

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

- absence_id
- base_url
- collaborator_id
- company_id
- contract_id
- document_id
- limit
- mode
- pay_period
- payslip_id
- api_key (secret) (required)

## ETL Streams

- employees:
  - primary key: id
  - cursor: updated_at
  - fields: email(string), first_name(string), id(string), last_name(string), updated_at(string)
- contracts:
  - primary key: id
  - cursor: updated_at
  - fields: employee_id(string), id(string), start_date(string), type(string), updated_at(string)
- companies:
  - primary key: id
  - cursor: updated_at
  - fields: country(string), id(string), name(string), updated_at(string)
- absences:
  - primary key: id
  - fields: id(string)
- collaborators:
  - primary key: id
  - fields: id(string)
- collaborator_meal_vouchers:
  - primary key: collaboratorId
  - fields: collaboratorId(string)
- collaborator_payslips:
  - primary key: payslipId
  - fields: payslipId(string)
- company_contracts:
  - primary key: contractId
  - fields: contractId(string)
- company_contracts_fr:
  - primary key: contractId
  - fields: contractId(string)
- worked_time_by_contract:
  - primary key: contractId
  - fields: contractId(string)
- health_insurance_contracts:
  - primary key: idContrat
  - fields: idContrat(string)
- provident_fund_contracts:
  - primary key: idContrat
  - fields: idContrat(string)
- auto_enrolment_documents:
  - primary key: documentId
  - fields: documentId(string)
- income_tax_documents:
  - primary key: documentId
  - fields: documentId(string)
- accounting_v2_entries:
  - primary key: operationDate, accountId, contractId
  - fields: accountId(string), contractId(string), operationDate(string)
- company:
  - primary key: id
  - fields: id(string)
- company_fr:
  - primary key: id
  - fields: id(string)
- collaborator:
  - primary key: id
  - fields: id(string)
- company_contract:
  - primary key: contractId
  - fields: contractId(string)
- company_contract_fr:
  - primary key: contractId
  - fields: contractId(string)
- payroll_status:
  - primary key: company_id
  - fields: company_id(string)

## Sync Modes

- ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped

## Reverse ETL Actions

- create_absence:
  - endpoint: POST /companies/{{ record.company_id }}/absences
  - required fields: company_id, contractId, type, startDate, endDate
  - risk: external PayFit mutation; create absence; approval required
- cancel_absence:
  - endpoint: DELETE /companies/{{ record.company_id }}/absences/{{ record.absence_id }}
  - required fields: company_id, absence_id
  - risk: external mutation; cancels an existing PayFit absence; approval required
- create_collaborator:
  - endpoint: POST /companies/{{ record.company_id }}/collaborators
  - required fields: company_id, firstName, lastName, personalEmail
  - risk: external PayFit mutation; create collaborator; approval required
- create_contract:
  - endpoint: POST /companies/{{ record.company_id }}/collaborators/{{ record.collaborator_id }}/contracts
  - required fields: company_id, collaborator_id, jobTitle, startDate
  - risk: external PayFit mutation; create contract; approval required
- update_contract_health_insurance:
  - endpoint: PUT /companies/{{ record.company_id }}/contracts-fr/{{ record.contract_id }}/health-insurance
  - required fields: company_id, contract_id, healthInsuranceContractIds
  - risk: external PayFit mutation; update contract health insurance; approval required
- update_contract_provident_fund:
  - endpoint: PUT /companies/{{ record.company_id }}/contracts-fr/{{ record.contract_id }}/provident-fund
  - required fields: company_id, contract_id, providentFundContractIds
  - risk: external PayFit mutation; update contract provident fund; approval required
- request_health_insurance_regularization:
  - endpoint: POST /companies/{{ record.company_id }}/contracts-fr/{{ record.contract_id }}/regularization
  - required fields: company_id, contract_id, healthInsuranceContractIds, effectiveDate
  - risk: external PayFit mutation; request health insurance regularization; approval required

## Security

- read risk: external PayFit API read of HR, contract, payroll-status, absence, accounting, and document-metadata data
- write risk: external PayFit API mutations for collaborator, contract, absence, and health-insurance/provident-fund workflows
- approval: write actions require explicit reverse-ETL approval; absence cancellation is idempotent-delete modeled
- Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

## Commands

### Inspect as a manual

```bash
pm connectors inspect payfit
```

### Inspect as structured JSON

```bash
pm connectors inspect payfit --json
```

## Agent Rules

- Run pm connectors inspect payfit before creating credentials or plans.
- Use --json only when the caller needs structured output; use the manual for human-readable guidance.
- Never ask the user to paste secret values into chat.
- For reverse ETL writes, create a plan, show the preview, wait for explicit approval, then run with the approval token.
