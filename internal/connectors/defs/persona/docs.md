# Overview

Reads Persona inquiries, accounts, reports, transactions, and cases, and performs lifecycle
mutations (redact, inquiry approve/decline/expire/resume, report re-run/pause/resume-monitoring,
transaction biometrics redaction), through the Persona REST API.

Readable streams: `inquiries`, `accounts`, `reports`, `transactions`, `cases`.

Write actions: `redact_inquiry`, `redact_account`, `redact_case`, `redact_report`,
`redact_transaction`, `approve_inquiry`, `decline_inquiry`, `expire_inquiry`, `resume_inquiry`,
`rerun_report`, `pause_report_monitoring`, `resume_report_monitoring`,
`redact_transaction_biometrics`.

Service API documentation: https://docs.withpersona.com/api-reference.

## Auth setup

Connection fields:

- `api_key` (required, secret, string); Persona API key, sent as a Bearer token (Authorization:
  Bearer <api_key>). Never logged.
- `base_url` (optional, string); default `https://api.withpersona.com/api/v1`; format `uri`; Persona
  API base URL override for tests or proxies.
- `page_size` (optional, string); default `50`; Records per page (1-100).

Secret fields are redacted in logs and write previews: `api_key`.

Default configuration values: `base_url=https://api.withpersona.com/api/v1`, `page_size=50`.

Authentication behavior:

- Bearer token authentication using `secrets.api_key`.

Requests use the configured `base_url` value after applying defaults.

Connection checks call GET `/inquiries`.

## Streams notes

Default pagination: single request; no pagination.

- `inquiries`: GET `/inquiries` - records path `data`; query `page[size]`=`50`; follows a next-page
  URL from the response body; URL path `links.next`; next URLs stay on the configured API host;
  emits passthrough records.
- `accounts`: GET `/accounts` - records path `data`; query `page[size]`=`50`; follows a next-page
  URL from the response body; URL path `links.next`; next URLs stay on the configured API host;
  emits passthrough records.
- `reports`: GET `/reports` - records path `data`; query `page[size]`=`50`; follows a next-page URL
  from the response body; URL path `links.next`; next URLs stay on the configured API host; emits
  passthrough records.
- `transactions`: GET `/transactions` - records path `data`; query `page[size]`=`50`; follows a
  next-page URL from the response body; URL path `links.next`; next URLs stay on the configured API
  host; emits passthrough records.
- `cases`: GET `/cases` - records path `data`; query `page[size]`=`50`; follows a next-page URL from
  the response body; URL path `links.next`; next URLs stay on the configured API host; emits
  passthrough records.

## Write actions & risks

Overall write risk: external mutation of Persona identity-verification records:
redact_inquiry/redact_account/redact_case/redact_report/redact_transaction permanently and
irreversibly delete PII (Persona's own docs: "This action cannot be undone");
redact_transaction_biometrics does the same for biometric data only; approve_inquiry/decline_inquiry
finalize an identity-verification decision and trigger any associated workflows/webhooks;
expire_inquiry/resume_inquiry affect whether an individual can continue an in-progress verification
flow; rerun_report triggers a new metered/billed report run;
pause_report_monitoring/resume_report_monitoring toggle continuous monitoring on a report.

Reverse ETL writes should be planned, previewed, approved, and then executed. Declared actions:

- `redact_inquiry`: DELETE `/inquiries/{{ record.id }}` - kind `delete`; body type `none`; path
  fields `id`; required record fields `id`; accepted fields `id`; missing records treated as success
  for status `404`; confirmation `destructive`; risk: permanently and irreversibly deletes all PII
  associated with this Inquiry (Persona's own docs: "This action cannot be undone"); approval
  required.
- `redact_account`: DELETE `/accounts/{{ record.id }}` - kind `delete`; body type `none`; path
  fields `id`; required record fields `id`; accepted fields `id`; missing records treated as success
  for status `404`; confirmation `destructive`; risk: permanently and irreversibly deletes all PII
  associated with this Account and every Inquiry/Case/Report/Transaction linked to it; approval
  required.
- `redact_case`: DELETE `/cases/{{ record.id }}` - kind `delete`; body type `none`; path fields
  `id`; required record fields `id`; accepted fields `id`; missing records treated as success for
  status `404`; confirmation `destructive`; risk: permanently and irreversibly deletes all PII
  associated with this Case; approval required.
- `redact_report`: DELETE `/reports/{{ record.id }}` - kind `delete`; body type `none`; path fields
  `id`; required record fields `id`; accepted fields `id`; missing records treated as success for
  status `404`; confirmation `destructive`; risk: permanently and irreversibly deletes all PII
  associated with this Report; approval required.
- `redact_transaction`: DELETE `/transactions/{{ record.id }}` - kind `delete`; body type `none`;
  path fields `id`; required record fields `id`; accepted fields `id`; missing records treated as
  success for status `404`; confirmation `destructive`; risk: permanently and irreversibly deletes
  all PII associated with this Transaction; approval required.
- `approve_inquiry`: POST `/inquiries/{{ record.id }}/approve` - kind `update`; body type `none`;
  path fields `id`; required record fields `id`; accepted fields `id`; risk: finalizes an Inquiry's
  identity-verification decision as approved; triggers any workflows/webhooks associated with that
  transition; approval required.
- `decline_inquiry`: POST `/inquiries/{{ record.id }}/decline` - kind `update`; body type `none`;
  path fields `id`; required record fields `id`; accepted fields `id`; risk: finalizes an Inquiry's
  identity-verification decision as declined; triggers any workflows/webhooks associated with that
  transition; approval required.
- `expire_inquiry`: POST `/inquiries/{{ record.id }}/expire` - kind `update`; body type `none`; path
  fields `id`; required record fields `id`; accepted fields `id`; risk: ends an in-progress
  Inquiry's verification flow before completion, preventing the individual from continuing it;
  approval required.
- `resume_inquiry`: POST `/inquiries/{{ record.id }}/resume` - kind `update`; body type `none`; path
  fields `id`; required record fields `id`; accepted fields `id`; risk: re-opens a
  previously-expired or paused Inquiry so the individual can continue its verification flow;
  approval required.
- `rerun_report`: POST `/reports/{{ record.id }}/run` - kind `update`; body type `none`; path fields
  `id`; required record fields `id`; accepted fields `id`; risk: re-runs a continuously monitored
  Report immediately outside its normal recurrence schedule; a metered, billed external
  side-effecting action; approval required.
- `pause_report_monitoring`: POST `/reports/{{ record.id }}/pause` - kind `update`; body type
  `none`; path fields `id`; required record fields `id`; accepted fields `id`; risk: pauses
  continuous monitoring on a Report (Persona's own docs: requires additional permissions); the
  report stops re-evaluating for new matches until resumed; approval required.
- `resume_report_monitoring`: POST `/reports/{{ record.id }}/resume` - kind `update`; body type
  `none`; path fields `id`; required record fields `id`; accepted fields `id`; risk: resumes
  continuous monitoring on a previously paused Report (Persona's own docs: requires additional
  permissions); approval required.
- `redact_transaction_biometrics`: POST `/transactions/{{ record.id }}/redact-biometrics` - kind
  `update`; body type `none`; path fields `id`; required record fields `id`; accepted fields `id`;
  confirmation `destructive`; risk: permanently and irreversibly deletes biometric data for a
  Transaction and all its associated objects (Persona's own docs: "This action cannot be undone");
  narrower than redact_transaction (biometrics only, the rest of the transaction record is
  preserved); approval required.

## Known limits

- Batch defaults: read_page_size=50.
- API coverage includes 5 stream-backed endpoint group(s), 13 write-backed endpoint group(s).
- Other documented endpoints are not exposed by this connector where they are classified as
  binary_payload=3, duplicate_of=39, non_data_endpoint=2, out_of_scope=125,
  requires_elevated_scope=17.
