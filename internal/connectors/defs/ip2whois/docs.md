# Overview

Looks up WHOIS records for configured domains via the IP2WHOIS API, exposing a flattened whois
stream and per-role contact streams (registrant, admin, tech, billing).

Readable streams: `whois`, `contacts_registrant`, `contacts_admin`, `contacts_tech`,
`contacts_billing`.

This connector is read-only; no write actions are declared.

Service API documentation: https://www.ip2whois.com/developers-api.

## Auth setup

Connection fields:

- `api_key` (required, secret, string); IP2WHOIS API key, sent as the 'key' query parameter on every
  lookup. Never logged.
- `base_url` (optional, string); default `https://api.ip2whois.com/v2`; format `uri`; IP2WHOIS API
  base URL override for tests or proxies.
- `domains` (required, string); Comma-separated list of domains to look up (one WHOIS lookup per
  domain). A single domain is a length-1 list.
- `mode` (optional, string).

Secret fields are redacted in logs and write previews: `api_key`.

Default configuration values: `base_url=https://api.ip2whois.com/v2`.

Authentication behavior:

- API key authentication in query parameter `key` using `secrets.api_key`.

Requests use the configured `base_url` value after applying defaults.

Connection checks use a connector-managed request with query `domain`=`{{ config.domains }}`.

## Streams notes

Default pagination: single request; no pagination.

- `whois`: GET connector-managed request path - records at response root; computed output fields
  `admin_email`, `admin_name`, `billing_email`, `billing_name`, `nameservers`, `registrant_country`,
  `registrant_email`, `registrant_name`, `registrant_organization`, `registrar_iana_id`,
  `registrar_name`, `registrar_url`, `tech_email`, `tech_name`; fan-out; ids from config field
  `domains`; id sent as query parameter `domain`.
- `contacts_registrant`: GET connector-managed request path - records path `registrant`; computed
  output fields `role`; fan-out; ids from config field `domains`; id sent as query parameter
  `domain`; stamps `domain`.
- `contacts_admin`: GET connector-managed request path - records path `admin`; computed output
  fields `role`; fan-out; ids from config field `domains`; id sent as query parameter `domain`;
  stamps `domain`.
- `contacts_tech`: GET connector-managed request path - records path `tech`; computed output fields
  `role`; fan-out; ids from config field `domains`; id sent as query parameter `domain`; stamps
  `domain`.
- `contacts_billing`: GET connector-managed request path - records path `billing`; computed output
  fields `role`; fan-out; ids from config field `domains`; id sent as query parameter `domain`;
  stamps `domain`.

## Write actions & risks

This connector is read-only. Read behavior: external IP2WHOIS API read of WHOIS records for the
configured domain set.

## Known limits

- API coverage includes 5 stream-backed endpoint group(s).
- Other documented endpoints are not exposed by this connector where they are classified as
  out_of_scope=1.
