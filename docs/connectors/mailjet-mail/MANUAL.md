# pm connectors inspect mailjet-mail

```text
NAME
  pm connectors inspect mailjet-mail - Mailjet Mail connector manual

SYNOPSIS
  pm connectors inspect mailjet-mail
  pm connectors inspect mailjet-mail --json
  pm credentials add <name> --connector mailjet-mail [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Reads Mailjet contacts, contact lists, messages, campaigns, and statistics through the Mailjet Email REST API (v3).

ICON
  id: mailjetmail
  asset: icons/mailjetmail.svg
  source: upstream_registry
  review_status: upstream_seeded
  review_url: https://dev.mailjet.com/email/reference/

CAPABILITIES
  check=true catalog=true read=true write=false query=false
  Integration type: api

AUTHENTICATION
  Use pm credentials add with --from-env or --value-stdin for secret fields.

CONFIGURATION
  api_key (required)
  base_url
  mode
  page_size
  api_key_secret (secret) (required)

ETL STREAMS
  contacts:
    primary key: ID
    fields: CreatedAt(string), DeliveredCount(integer), Email(string), ID(integer), IsExcludedFromCampaigns(boolean), IsOptInPending(boolean), IsSpamComplaining(boolean), LastActivityAt(string), LastUpdateAt(string), Name(string)
  contactslists:
    primary key: ID
    fields: Address(string), CreatedAt(string), ID(integer), IsDeleted(boolean), Name(string), SubscriberCount(integer)
  messages:
    primary key: ID
    fields: ArrivedAt(string), AttemptCount(integer), CampaignID(integer), ContactID(integer), ID(integer), IsClickTracked(boolean), IsOpenTracked(boolean), MessageSize(integer), Status(string)
  campaigns:
    primary key: ID
    fields: CreatedAt(string), FromEmail(string), FromName(string), ID(integer), IsDeleted(boolean), IsStarred(boolean), SendStartAt(string), Status(integer), Subject(string)
  stats:
    primary key: ID
    fields: ID(integer), MessageBouncedCount(integer), MessageClickedCount(integer), MessageDeliveredCount(integer), MessageOpenedCount(integer), MessageSentCount(integer), MessageSpamCount(integer), MessageUnsubscribedCount(integer)

SYNC MODES
  ETL sync modes: full_refresh_append, full_refresh_overwrite

SECURITY
  read risk: external Mailjet API read of contact, list, message, campaign, and statistics data
  approval: none; read-only source connector
  Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

EXAMPLES
  # Inspect as a manual
  pm connectors inspect mailjet-mail

  # Inspect as structured JSON
  pm connectors inspect mailjet-mail --json

AGENT WORKFLOW
  - Run pm connectors inspect mailjet-mail before creating credentials or plans.
  - Use --json only when the caller needs structured output; use the manual for human-readable guidance.
  - Never ask the user to paste secret values into chat.

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
