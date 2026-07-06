# Overview

Reads Zenefits people, companies, departments, locations, employments, custom fields/values, bank
accounts, labor groups, and time-off data.

Readable streams: `people`, `companies`, `departments`, `locations`, `employments`, `custom_fields`,
`custom_field_values`, `company_banks`, `employee_banks`, `labor_group_types`, `labor_groups`,
`vacation_types`, `vacation_requests`.

This connector is read-only; no write actions are declared.

Service API documentation: https://developers.zenefits.com/.

## Auth setup

Connection fields:

- `base_url` (optional, string); default `https://api.zenefits.com`; format `uri`; Zenefits API root
  URL.
- `token` (required, secret, string); Zenefits API token, sent as an 'Authorization: Bearer <token>'
  header.

Secret fields are redacted in logs and write previews: `token`.

Default configuration values: `base_url=https://api.zenefits.com`.

Authentication behavior:

- Bearer token authentication using `secrets.token`.

Requests use the configured `base_url` value after applying defaults.

## Streams notes

Default pagination: single request; no pagination.

Pagination by stream: next_url: `locations`, `employments`, `custom_fields`, `custom_field_values`,
`company_banks`, `employee_banks`, `labor_group_types`, `labor_groups`, `vacation_types`,
`vacation_requests`; none: `people`, `companies`, `departments`.

- `people`: GET `/core/people` - records path `data`.
- `companies`: GET `/core/companies` - records path `data`.
- `departments`: GET `/core/departments` - records path `data`.
- `locations`: GET `/core/locations` - records path `data.data`; follows a next-page URL from the
  response body; URL path `data.next_url`; next URLs stay on the configured API host.
- `employments`: GET `/core/employments` - records path `data.data`; follows a next-page URL from
  the response body; URL path `data.next_url`; next URLs stay on the configured API host.
- `custom_fields`: GET `/core/custom_fields` - records path `data.data`; follows a next-page URL
  from the response body; URL path `data.next_url`; next URLs stay on the configured API host.
- `custom_field_values`: GET `/core/custom_field_values` - records path `data.data`; follows a
  next-page URL from the response body; URL path `data.next_url`; next URLs stay on the configured
  API host.
- `company_banks`: GET `/core/company_banks` - records path `data.data`; follows a next-page URL
  from the response body; URL path `data.next_url`; next URLs stay on the configured API host.
- `employee_banks`: GET `/core/banks` - records path `data.data`; follows a next-page URL from the
  response body; URL path `data.next_url`; next URLs stay on the configured API host.
- `labor_group_types`: GET `/core/labor_group_types` - records path `data.data`; follows a next-page
  URL from the response body; URL path `data.next_url`; next URLs stay on the configured API host.
- `labor_groups`: GET `/core/labor_groups` - records path `data.data`; follows a next-page URL from
  the response body; URL path `data.next_url`; next URLs stay on the configured API host.
- `vacation_types`: GET `/time_off/vacation_types` - records path `data.data`; follows a next-page
  URL from the response body; URL path `data.next_url`; next URLs stay on the configured API host.
- `vacation_requests`: GET `/time_off/vacation_requests` - records path `data.data`; follows a
  next-page URL from the response body; URL path `data.next_url`; next URLs stay on the configured
  API host.

## Write actions & risks

This connector is read-only. Read behavior: external Zenefits account read of people, companies,
departments, locations, employments, custom field definitions/values, company and employee bank
account details, labor groups, and time-off vacation types/requests.

## Known limits

- API coverage includes 13 stream-backed endpoint group(s).
- Other documented endpoints are not exposed by this connector where they are classified as
  duplicate_of=22, non_data_endpoint=2, out_of_scope=3.
