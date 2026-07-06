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
  api_key
  base_url
  mode
  page_size
  api_key_secret (secret)

ETL STREAMS
  contacts:
    primary key: ID
    fields: CreatedAt(), DeliveredCount(), Email(), ID(), IsExcludedFromCampaigns(), IsOptInPending(), IsSpamComplaining(), LastActivityAt(), LastUpdateAt(), Name()
  contactslists:
    primary key: ID
    fields: Address(), CreatedAt(), ID(), IsDeleted(), Name(), SubscriberCount()
  messages:
    primary key: ID
    fields: ArrivedAt(), AttemptCount(), CampaignID(), ContactID(), ID(), IsClickTracked(), IsOpenTracked(), MessageSize(), Status()
  campaigns:
    primary key: ID
    fields: CreatedAt(), FromEmail(), FromName(), ID(), IsDeleted(), IsStarred(), SendStartAt(), Status(), Subject()
  stats:
    primary key: ID
    fields: ID(), MessageBouncedCount(), MessageClickedCount(), MessageDeliveredCount(), MessageOpenedCount(), MessageSentCount(), MessageSpamCount(), MessageUnsubscribedCount()

SYNC MODES
  ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped

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
