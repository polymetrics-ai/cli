# pm connectors inspect snapchat-marketing

```text
NAME
  pm connectors inspect snapchat-marketing - Snapchat Marketing connector manual

SYNOPSIS
  pm connectors inspect snapchat-marketing
  pm connectors inspect snapchat-marketing --json
  pm credentials add <name> --connector snapchat-marketing [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Reads Snapchat Marketing (Ads API) organizations, ad accounts, campaigns, ad squads, and ads via the OAuth2 refresh-token grant.

ICON
  asset: icons/snapchat.svg
  source: upstream_registry
  review_status: upstream_seeded
  review_url: https://developers.snap.com/api/marketing-api/Ads-API/announcements

CAPABILITIES
  check=true catalog=true read=true write=false query=false
  Integration type: api

AUTHENTICATION
  Use pm credentials add with --from-env or --value-stdin for secret fields.

CONFIGURATION
  ad_account_ids
  base_url
  organization_ids
  token_url
  client_id (secret)
  client_secret (secret)
  refresh_token (secret)

ETL STREAMS
  organizations:
    primary key: id
    cursor: updated_at
    fields: address_line_1(), administrative_district_level_1(), country(), created_at(), id(), locality(), name(), postal_code(), type(), updated_at()
  adaccounts:
    primary key: id
    cursor: updated_at
    fields: advertiser(), created_at(), currency(), id(), name(), organization_id(), status(), timezone(), type(), updated_at()
  campaigns:
    primary key: id
    cursor: updated_at
    fields: ad_account_id(), created_at(), daily_budget_micro(), end_time(), id(), lifetime_spend_cap_micro(), name(), objective(), start_time(), status(), updated_at()
  adsquads:
    primary key: id
    cursor: updated_at
    fields: bid_micro(), billing_event(), campaign_id(), created_at(), daily_budget_micro(), id(), name(), optimization_goal(), status(), type(), updated_at()
  ads:
    primary key: id
    cursor: updated_at
    fields: ad_squad_id(), created_at(), creative_id(), id(), name(), review_status(), status(), type(), updated_at()

SYNC MODES
  ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped

SECURITY
  read risk: external Snapchat Ads API read of organizations, ad accounts, campaigns, ad squads, and ads under the configured organization/ad-account ids
  approval: none; read-only, no reverse-ETL write surface
  Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

EXAMPLES
  # Inspect as a manual
  pm connectors inspect snapchat-marketing

  # Inspect as structured JSON
  pm connectors inspect snapchat-marketing --json

AGENT WORKFLOW
  - Run pm connectors inspect snapchat-marketing before creating credentials or plans.
  - Use --json only when the caller needs structured output; use the manual for human-readable guidance.
  - Never ask the user to paste secret values into chat.

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
