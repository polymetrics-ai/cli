# Overview

Reads Microsoft Advertising (Bing Ads) accounts, users, campaigns, ad groups, and ads through the
v13 Customer Management and Campaign Management REST APIs. Read-only.

Readable streams: `accounts`, `users`, `campaigns`, `ad_groups`, `ads`.

This connector is read-only; no write actions are declared.

Service API documentation:
https://learn.microsoft.com/en-us/advertising/customer-management-service/customer-management-service-rest-api.

## Auth setup

Connection fields:

- `account_id` (optional, string); Single advertiser account id; used as the CustomerAccountId
  header fallback and as the sole AccountId when account_ids is unset.
- `account_ids` (optional, string); Comma-separated advertiser account ids. The campaigns stream
  fans out one QueryByAccountId request per id, stamping account_id onto every emitted record. Falls
  back to a single-element list from customer_account_id/account_id when unset.
- `ad_group_id` (optional, string); Ad group id used to scope the ads stream's Ads/QueryByAdGroupId
  request body.
- `base_url` (optional, string); default
  `https://clientcenter.api.bingads.microsoft.com/CustomerManagement/v13`; format `uri`; Customer
  Management REST service base URL override for tests or proxies.
- `campaign_base_url` (optional, string); default
  `https://campaign.api.bingads.microsoft.com/CampaignManagement/v13`; format `uri`; Campaign
  Management REST service base URL override for tests or proxies. Campaign-scoped streams
  (campaigns, ad_groups, ads) target this service instead of base_url.
- `campaign_id` (optional, string); Campaign id used to scope the ad_groups stream's
  AdGroups/QueryByCampaignId request body.
- `client_id` (required, secret, string); Microsoft identity platform (Azure AD) application
  (client) ID for the refresh-token grant. Used only in the token-request form; never logged.
- `client_secret` (optional, secret, string); Microsoft identity platform application client secret
  (optional for public/native client app registrations). Used only in the token-request form; never
  logged.
- `customer_account_id` (optional, string); Microsoft Advertising CustomerAccountId header value for
  Campaign Management requests. Falls back to account_id when unset.
- `customer_id` (optional, string); Microsoft Advertising CustomerId header/body scoping value, sent
  on Campaign Management requests and included in the accounts (AccountsInfo) query body when set.
- `developer_token` (required, secret, string); Microsoft Advertising developer token, sent as the
  DeveloperToken header on every Customer/Campaign Management request; never logged.
- `refresh_token` (required, secret, string); Long-lived Microsoft identity platform OAuth 2.0
  refresh token. Exchanged for a short-lived access token at token_url; never logged. The 3-legged
  consent/acquisition dance is out of scope for this connector (credentials layer already owns it).
- `tenant_id` (optional, secret, string); Azure AD tenant id used to build the default token_url
  (https://login.microsoftonline.com/<tenant_id>/oauth2/v2.0/token). Defaults to 'common' when
  unset.
- `token_url` (optional, string); format `uri`; Microsoft identity platform OAuth 2.0 token endpoint
  override. Defaults to https://login.microsoftonline.com/<tenant_id>/oauth2/v2.0/token. MUST be
  https in production; the hook fails closed on a non-https or unparseable value.

Secret fields are redacted in logs and write previews: `client_id`, `client_secret`,
`developer_token`, `refresh_token`, `tenant_id`.

Default configuration values:
`base_url=https://clientcenter.api.bingads.microsoft.com/CustomerManagement/v13`,
`campaign_base_url=https://campaign.api.bingads.microsoft.com/CampaignManagement/v13`.

Provide the secret fields listed above. Authentication is applied by the connector-specific
implementation for this service.

Requests use the configured `base_url` value after applying defaults.

Connection checks call POST `/AccountsInfo/Query`.

## Streams notes

Default pagination: single request; no pagination.

- `accounts`: POST `/AccountsInfo/Query` - records path `AccountsInfo`.
- `users`: POST `/User/Query` - records path `User`.
- `campaigns`: POST `/Campaigns/QueryByAccountId` - records path `Campaigns`.
- `ad_groups`: POST `/AdGroups/QueryByCampaignId` - records path `AdGroups`.
- `ads`: POST `/Ads/QueryByAdGroupId` - records path `Ads`.

## Write actions & risks

This connector is read-only. Read behavior: external Microsoft Advertising REST API read of
account/user/campaign/ad-group/ad metadata.

## Known limits

- Batch defaults: read_page_size=0.
- API coverage includes 5 stream-backed endpoint group(s).
- Other documented endpoints are not exposed by this connector where they are classified as
  binary_payload=1, destructive_admin=1, out_of_scope=5.
