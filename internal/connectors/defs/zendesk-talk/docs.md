# Overview

Reads Zendesk Talk phone numbers, greetings, greeting categories, IVRs, and agent activity
statistics through the Zendesk Talk (voice) REST API.

Readable streams: `phone_numbers`, `greetings`, `greeting_categories`, `ivrs`, `agents_activity`.

This connector is read-only; no write actions are declared.

Service API documentation: https://developer.zendesk.com/api-reference/voice/talk-api/introduction/.

## Auth setup

Connection fields:

- `access_token` (optional, secret, string); Zendesk OAuth2 access token. Sent as Authorization:
  Bearer <access_token>.
- `api_token` (optional, secret, string); Zendesk API token (Admin Center > Apps and integrations >
  APIs > Zendesk API). Sent via HTTP Basic as '<email>/token:<api_token>'. Requires email.
- `base_url` (required, string); format `uri`; Your Zendesk Talk API root, e.g.
  https://acme.zendesk.com/api/v2/channels/voice for subdomain 'acme'. Also usable as a base URL
  override for tests/proxies.
- `email` (optional, secret, string); Zendesk agent email address paired with api_token for
  API-token Basic auth (the '<email>/token' username half).

Secret fields are redacted in logs and write previews: `access_token`, `api_token`, `email`.

Authentication behavior:

- Bearer token authentication using `secrets.access_token` when `{{ secrets.access_token }}`.
- HTTP Basic authentication using `secrets.email`, `secrets.api_token` when `{{ secrets.api_token
  }}`.

Requests use the configured `base_url` value after applying defaults.

Connection checks call GET `/greeting_categories`.

## Streams notes

Default pagination: follows a next-page URL from the response body; URL path `next_page`; next URLs
stay on the configured API host.

- `phone_numbers`: GET `/phone_numbers` - records path `phone_numbers`; query `per_page`=`100`;
  follows a next-page URL from the response body; URL path `next_page`; next URLs stay on the
  configured API host.
- `greetings`: GET `/greetings` - records path `greetings`; query `per_page`=`100`; follows a
  next-page URL from the response body; URL path `next_page`; next URLs stay on the configured API
  host.
- `greeting_categories`: GET `/greeting_categories` - records path `greeting_categories`; query
  `per_page`=`100`; follows a next-page URL from the response body; URL path `next_page`; next URLs
  stay on the configured API host.
- `ivrs`: GET `/ivrs` - records path `ivrs`; query `per_page`=`100`; follows a next-page URL from
  the response body; URL path `next_page`; next URLs stay on the configured API host.
- `agents_activity`: GET `/stats/agents_activity` - records path `agents_activity`; query
  `per_page`=`100`; follows a next-page URL from the response body; URL path `next_page`; next URLs
  stay on the configured API host.

## Write actions & risks

This connector is read-only. Read behavior: external Zendesk Talk API read of phone number,
greeting, and agent activity data.

## Known limits

- Batch defaults: read_page_size=100.
- API coverage includes 5 stream-backed endpoint group(s).
- Other documented endpoints are not exposed by this connector where they are classified as
  non_data_endpoint=2, out_of_scope=4.
