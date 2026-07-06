# Overview

Reads the authenticated Strava athlete's profile, activities, lifetime stats, and clubs through the
Strava v3 REST API.

Readable streams: `activities`, `athlete`, `athlete_stats`, `clubs`.

This connector is read-only; no write actions are declared.

Service API documentation: https://developers.strava.com/docs/reference/.

## Auth setup

Connection fields:

- `athlete_id` (optional, string); The authenticated athlete's numeric Strava ID. Required only for
  the athlete_stats stream, whose resource path substitutes {athlete_id}.
- `base_url` (optional, string); default `https://www.strava.com/api/v3`; format `uri`; Strava API
  base URL override for tests or proxies.
- `client_id` (required, string); Strava OAuth 2.0 client ID for the refresh-token grant. Used only
  in the token-request form; never logged.
- `client_secret` (required, secret, string); Strava OAuth 2.0 client secret. Used only in the
  token-request form; never logged.
- `refresh_token` (required, secret, string); Long-lived Strava OAuth 2.0 refresh token. Exchanged
  for a short-lived access token at token_url; never logged. The 3-legged consent/acquisition dance
  is out of scope for this connector (credentials layer already owns it).
- `token_url` (optional, string); default `https://www.strava.com/oauth/token`; format `uri`; Strava
  OAuth 2.0 token endpoint override. MUST be http(s) with a host; the hook fails closed on an
  invalid value to prevent exfiltrating the refresh token/client secret to an attacker-chosen
  endpoint.

Secret fields are redacted in logs and write previews: `client_secret`, `refresh_token`.

Default configuration values: `base_url=https://www.strava.com/api/v3`,
`token_url=https://www.strava.com/oauth/token`.

Authentication behavior:

- Connector-specific authentication using `secrets.refresh_token`, `config.token_url`,
  `config.client_id`, `secrets.client_secret`.

Requests use the configured `base_url` value after applying defaults.

Connection checks call GET `/athlete`.

## Streams notes

Default pagination: page-number pagination; page parameter `page`; size parameter `per_page`; starts
at 1; page size 100.

Pagination by stream: none: `athlete`, `athlete_stats`; page_number: `activities`, `clubs`.

- `activities`: GET `/athlete/activities` - records path `.`; page-number pagination; page parameter
  `page`; size parameter `per_page`; starts at 1; page size 100.
- `athlete`: GET `/athlete` - single-object response; records path `.`.
- `athlete_stats`: GET `/athletes/{{ config.athlete_id }}/stats` - single-object response; records
  path `.`; computed output fields `id`.
- `clubs`: GET `/athlete/clubs` - records path `.`; page-number pagination; page parameter `page`;
  size parameter `per_page`; starts at 1; page size 100.

## Write actions & risks

This connector is read-only. Read behavior: external Strava API read of the authenticated athlete's
profile, activity, stats, and club data.

## Known limits

- Batch defaults: read_page_size=100.
- API coverage includes 4 stream-backed endpoint group(s).
- Other documented endpoints are not exposed by this connector where they are classified as
  out_of_scope=8.
