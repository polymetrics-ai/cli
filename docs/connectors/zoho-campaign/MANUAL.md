# pm connectors inspect zoho-campaign

```text
NAME
  pm connectors inspect zoho-campaign - Zoho Campaign connector manual

SYNOPSIS
  pm connectors inspect zoho-campaign
  pm connectors inspect zoho-campaign --json
  pm credentials add <name> --connector zoho-campaign [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Reads Zoho Campaigns lists, campaigns, and contacts through the Zoho Campaigns REST API.

ICON
  id: simple-icons-zoho-campaign
  asset: icons/simple-icons/zoho-campaign.svg
  title: Zoho
  simple_icon_slug: zoho
  simple_icon_hex: E42527
  source: simple-icons
  license: CC0-1.0
  review_status: cc0_with_trademark_caveat
  review_url: https://simpleicons.org/?q=Zoho
  match: curated-alias
  matched_by: zoho

CAPABILITIES
  check=true catalog=true read=true write=false query=false
  Integration type: api

AUTHENTICATION
  Use pm credentials add with --from-env or --value-stdin for secret fields.

CONFIGURATION
  base_url
  max_pages
  mode
  page_size
  access_token (secret)

ETL STREAMS
  lists:
    primary key: id
    cursor: updated_at
    fields: createdtime(), id(), list_name(), listkey(), listname(), listtype(), modified_time(), name(), tag(), updated_at()
  campaigns:
    primary key: id
    cursor: updated_at
    fields: campaign_key(), campaign_name(), campaignkey(), campaignname(), from_email(), id(), modified_time(), name(), sent_time(), status(), subject(), updated_at()
  contacts:
    primary key: id
    cursor: updated_at
    fields: contact_id(), contact_key(), email(), first_name(), id(), last_name(), modified_time(), name(), status(), updated_at()

SYNC MODES
  ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped

SECURITY
  read risk: external Zoho Campaigns API read of email marketing data
  approval: none; read-only
  Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

EXAMPLES
  # Inspect as a manual
  pm connectors inspect zoho-campaign

  # Inspect as structured JSON
  pm connectors inspect zoho-campaign --json

AGENT WORKFLOW
  - Run pm connectors inspect zoho-campaign before creating credentials or plans.
  - Use --json only when the caller needs structured output; use the manual for human-readable guidance.
  - Never ask the user to paste secret values into chat.

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
