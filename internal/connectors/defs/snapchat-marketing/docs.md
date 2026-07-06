# Overview

Reads Snapchat Marketing (Ads API) organizations, ad accounts, campaigns, ad squads, and ads via the
OAuth2 refresh-token grant.

Readable streams: `organizations`, `adaccounts`, `campaigns`, `adsquads`, `ads`.

This connector is read-only; no write actions are declared.

Service API documentation: https://marketingapi.snapchat.com/docs/.

## Auth setup

Connection fields:

- `ad_account_ids` (optional, string); Comma-separated list of Snapchat ad account ids to read
  campaigns/adsquads/ads for (required for the campaigns, adsquads, and ads streams, e.g.
  "ACC1,ACC2").
- `base_url` (optional, string); default `https://adsapi.snapchat.com/v1`; format `uri`; Snapchat
  Ads API base URL override for tests or proxies.
- `client_id` (required, secret, string); Snapchat Marketing API OAuth 2.0 client ID for the
  refresh-token grant. Used only in the token-request form; never logged.
- `client_secret` (required, secret, string); Snapchat Marketing API OAuth 2.0 client secret. Used
  only in the token-request form; never logged.
- `organization_ids` (optional, string); Comma-separated list of Snapchat organization ids to read
  ad accounts for (required for the adaccounts stream, e.g. "org1,org2").
- `refresh_token` (required, secret, string); Long-lived Snapchat Marketing API OAuth 2.0 refresh
  token. Exchanged for a short-lived access token at token_url; never logged. The 3-legged
  consent/acquisition dance is out of scope for this connector (credentials layer already owns it).
- `token_url` (optional, string); default `https://accounts.snapchat.com/login/oauth2/access_token`;
  format `uri`; Snapchat OAuth 2.0 token endpoint override. MUST be http(s) with a host; the hook
  fails closed on an invalid value to prevent exfiltrating the refresh token/client secret to an
  attacker-chosen endpoint.

Secret fields are redacted in logs and write previews: `client_id`, `client_secret`,
`refresh_token`.

Default configuration values: `base_url=https://adsapi.snapchat.com/v1`,
`token_url=https://accounts.snapchat.com/login/oauth2/access_token`.

Authentication behavior:

- Connector-specific authentication using `secrets.refresh_token`, `config.token_url`,
  `secrets.client_id`, `secrets.client_secret`.

Requests use the configured `base_url` value after applying defaults.

Connection checks call GET `/organizations` with query `limit`=`1`.

## Streams notes

Default pagination: follows a next-page URL from the response body; URL path `paging.next_link`;
next URLs stay on the configured API host.

- `organizations`: GET `/organizations` - records path `organizations`; query `limit`=`50`; follows
  a next-page URL from the response body; URL path `paging.next_link`; next URLs stay on the
  configured API host; computed output fields `address_line_1`, `administrative_district_level_1`,
  `country`, `created_at`, `id`, `locality`, `name`, `postal_code`, `type`, `updated_at`.
- `adaccounts`: GET `/organizations/{{ fanout.id }}/adaccounts` - records path `adaccounts`; query
  `limit`=`50`; follows a next-page URL from the response body; URL path `paging.next_link`; next
  URLs stay on the configured API host; computed output fields `advertiser`, `created_at`,
  `currency`, `id`, `name`, `organization_id`, `status`, `timezone`, `type`, `updated_at`; fan-out;
  ids from config field `organization_ids`; id inserted into the request path.
- `campaigns`: GET `/adaccounts/{{ fanout.id }}/campaigns` - records path `campaigns`; query
  `limit`=`50`; follows a next-page URL from the response body; URL path `paging.next_link`; next
  URLs stay on the configured API host; computed output fields `ad_account_id`, `created_at`,
  `daily_budget_micro`, `end_time`, `id`, `lifetime_spend_cap_micro`, `name`, `objective`,
  `start_time`, `status`, `updated_at`; fan-out; ids from config field `ad_account_ids`; id inserted
  into the request path.
- `adsquads`: GET `/adaccounts/{{ fanout.id }}/adsquads` - records path `adsquads`; query
  `limit`=`50`; follows a next-page URL from the response body; URL path `paging.next_link`; next
  URLs stay on the configured API host; computed output fields `bid_micro`, `billing_event`,
  `campaign_id`, `created_at`, `daily_budget_micro`, `id`, `name`, `optimization_goal`, `status`,
  `type`, `updated_at`; fan-out; ids from config field `ad_account_ids`; id inserted into the
  request path.
- `ads`: GET `/adaccounts/{{ fanout.id }}/ads` - records path `ads`; query `limit`=`50`; follows a
  next-page URL from the response body; URL path `paging.next_link`; next URLs stay on the
  configured API host; computed output fields `ad_squad_id`, `created_at`, `creative_id`, `id`,
  `name`, `review_status`, `status`, `type`, `updated_at`; fan-out; ids from config field
  `ad_account_ids`; id inserted into the request path.

## Write actions & risks

This connector is read-only. Read behavior: external Snapchat Ads API read of organizations, ad
accounts, campaigns, ad squads, and ads under the configured organization/ad-account ids.

## Known limits

- Batch defaults: read_page_size=50.
- API coverage includes 5 stream-backed endpoint group(s).
- Other documented endpoints are not exposed by this connector where they are classified as
  out_of_scope=8.
