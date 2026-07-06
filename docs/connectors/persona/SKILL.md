---
name: pm-persona
description: Persona connector knowledge and safe action guide.
---

# pm-persona

## Purpose

Reads Persona inquiries, accounts, reports, transactions, and cases, and performs lifecycle mutations (redact, inquiry approve/decline/expire/resume, report re-run/pause/resume-monitoring, transaction biometrics redaction), through the Persona REST API.

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
- page_size
- api_key (secret)

## ETL Streams

- inquiries:
  - primary key: id
  - fields: attributes(), id(), relationships(), type()
- accounts:
  - primary key: id
  - fields: attributes(), id(), relationships(), type()
- reports:
  - primary key: id
  - fields: attributes(), id(), relationships(), type()
- transactions:
  - primary key: id
  - fields: attributes(), id(), relationships(), type()
- cases:
  - primary key: id
  - fields: attributes(), id(), relationships(), type()

## Sync Modes

- ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped

## Reverse ETL Actions

- redact_inquiry:
  - endpoint: DELETE /inquiries/{{ record.id }}
  - required fields: id
  - risk: permanently and irreversibly deletes all PII associated with this Inquiry (Persona's own docs: "This action cannot be undone"); approval required
- redact_account:
  - endpoint: DELETE /accounts/{{ record.id }}
  - required fields: id
  - risk: permanently and irreversibly deletes all PII associated with this Account and every Inquiry/Case/Report/Transaction linked to it; approval required
- redact_case:
  - endpoint: DELETE /cases/{{ record.id }}
  - required fields: id
  - risk: permanently and irreversibly deletes all PII associated with this Case; approval required
- redact_report:
  - endpoint: DELETE /reports/{{ record.id }}
  - required fields: id
  - risk: permanently and irreversibly deletes all PII associated with this Report; approval required
- redact_transaction:
  - endpoint: DELETE /transactions/{{ record.id }}
  - required fields: id
  - risk: permanently and irreversibly deletes all PII associated with this Transaction; approval required
- approve_inquiry:
  - endpoint: POST /inquiries/{{ record.id }}/approve
  - required fields: id
  - risk: finalizes an Inquiry's identity-verification decision as approved; triggers any workflows/webhooks associated with that transition; approval required
- decline_inquiry:
  - endpoint: POST /inquiries/{{ record.id }}/decline
  - required fields: id
  - risk: finalizes an Inquiry's identity-verification decision as declined; triggers any workflows/webhooks associated with that transition; approval required
- expire_inquiry:
  - endpoint: POST /inquiries/{{ record.id }}/expire
  - required fields: id
  - risk: ends an in-progress Inquiry's verification flow before completion, preventing the individual from continuing it; approval required
- resume_inquiry:
  - endpoint: POST /inquiries/{{ record.id }}/resume
  - required fields: id
  - risk: re-opens a previously-expired or paused Inquiry so the individual can continue its verification flow; approval required
- rerun_report:
  - endpoint: POST /reports/{{ record.id }}/run
  - required fields: id
  - risk: re-runs a continuously monitored Report immediately outside its normal recurrence schedule; a metered, billed external side-effecting action; approval required
- pause_report_monitoring:
  - endpoint: POST /reports/{{ record.id }}/pause
  - required fields: id
  - risk: pauses continuous monitoring on a Report (Persona's own docs: requires additional permissions); the report stops re-evaluating for new matches until resumed; approval required
- resume_report_monitoring:
  - endpoint: POST /reports/{{ record.id }}/resume
  - required fields: id
  - risk: resumes continuous monitoring on a previously paused Report (Persona's own docs: requires additional permissions); approval required
- redact_transaction_biometrics:
  - endpoint: POST /transactions/{{ record.id }}/redact-biometrics
  - required fields: id
  - risk: permanently and irreversibly deletes biometric data for a Transaction and all its associated objects (Persona's own docs: "This action cannot be undone"); narrower than redact_transaction (biometrics only, the rest of the transaction record is preserved); approval required

## Security

- read risk: external Persona API read of inquiry, account, report, transaction, and case data; identity-verification data may include PII (names, government IDs, selfies, addresses)
- write risk: external mutation of Persona identity-verification records: redact_inquiry/redact_account/redact_case/redact_report/redact_transaction permanently and irreversibly delete PII (Persona's own docs: "This action cannot be undone"); redact_transaction_biometrics does the same for biometric data only; approve_inquiry/decline_inquiry finalize an identity-verification decision and trigger any associated workflows/webhooks; expire_inquiry/resume_inquiry affect whether an individual can continue an in-progress verification flow; rerun_report triggers a new metered/billed report run; pause_report_monitoring/resume_report_monitoring toggle continuous monitoring on a report
- approval: required for all writes; the 6 redact_*/redact_transaction_biometrics actions are destructive and irreversible (PII deletion), approve_inquiry/decline_inquiry make a final identity decision with downstream workflow/webhook side effects, and rerun_report has a metered billing side effect
- Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

## Commands

### Inspect as a manual

```bash
pm connectors inspect persona
```

### Inspect as structured JSON

```bash
pm connectors inspect persona --json
```

## Agent Rules

- Run pm connectors inspect persona before creating credentials or plans.
- Use --json only when the caller needs structured output; use the manual for human-readable guidance.
- Never ask the user to paste secret values into chat.
- For reverse ETL writes, create a plan, show the preview, wait for explicit approval, then run with the approval token.
