# pm connectors inspect mailersend

```text
NAME
  pm connectors inspect mailersend - MailerSend connector manual

SYNOPSIS
  pm connectors inspect mailersend
  pm connectors inspect mailersend --json
  pm credentials add <name> --connector mailersend [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Reads MailerSend email activity, analytics, domains, messages, recipients, templates, scheduled messages, sender identities, inbound routes, users, invites, tokens, and webhooks through the MailerSend REST API.

ICON
  id: mailersend
  asset: icons/mailersend.svg
  source: upstream_registry
  review_status: upstream_seeded
  review_url: https://developers.mailersend.com/

CAPABILITIES
  check=true catalog=true read=true write=false query=false
  Integration type: api

AUTHENTICATION
  Use pm credentials add with --from-env or --value-stdin for secret fields.

CONFIGURATION
  base_url
  date_from
  date_to
  domain_id
  mode
  start_date
  api_token (secret) (required)

ETL STREAMS
  activity:
    primary key: id
    cursor: created_at
    fields: created_at(string), email(object), id(string), type(string), updated_at(string)
  domains:
    primary key: id
    cursor: updated_at
    fields: created_at(string), dkim(boolean), id(string), is_dns_active(boolean), is_verified(boolean), name(string), spf(boolean), tracking(boolean), updated_at(string)
  messages:
    primary key: id
    cursor: created_at
    fields: created_at(string), id(string), updated_at(string)
  recipients:
    primary key: id
    cursor: updated_at
    fields: created_at(string), deleted_at(string), email(string), id(string), updated_at(string)
  templates:
    primary key: id
    cursor: updated_at
    fields: created_at(string), description(string), id(string), image_path(string), name(string), tags(array), type(string), updated_at(string), variables(object)
  scheduled_messages:
    primary key: message_id
    cursor: created_at
    fields: created_at(string), domain(object), message(object), message_id(string), send_at(string), status(string), status_message(string), subject(string)
  sender_identities:
    primary key: id
    fields: add_note(boolean), domain(object), email(string), id(string), is_verified(boolean), name(string), personal_note(string), reply_to_email(string), reply_to_name(string), resends(integer)
  inbound_routes:
    primary key: id
    fields: address(string), dns_checked_at(string), domain(string), enabled(boolean), filters(array), forwards(array), id(string), mxValues(object), name(string), priority(integer)
  account_users:
    primary key: id
    fields: created_at(string), email(string), id(string), name(string), permissions(array), role(string), status(string), updated_at(string)
  invites:
    primary key: id
    cursor: updated_at
    fields: created_at(string), data(object), email(string), id(string), permissions(array), requires_periodic_password_change(boolean), role(string), updated_at(string)
  tokens:
    primary key: id
    fields: created_at(string), id(string), name(string), scopes(array), status(string)
  webhooks:
    primary key: id
    fields: created_at(string), domain(object), enabled(boolean), events(array), id(string), name(string), updated_at(string), url(string)
  analytics_by_date:
    primary key: date
    fields: clicked(integer), clicked_unique(integer), date(string), delivered(integer), hard_bounced(integer), opened(integer), opened_unique(integer), queued(integer), sent(integer), soft_bounced(integer), spam_complaints(integer), survey_opened(integer), survey_submitted(integer), unsubscribed(integer)
  analytics_country:
    primary key: name
    fields: count(integer), name(string)
  analytics_user_agents:
    primary key: name
    fields: count(integer), name(string)
  analytics_reading_environment:
    primary key: name
    fields: count(integer), name(string)

SYNC MODES
  ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped

SECURITY
  read risk: external MailerSend API read of email activity, analytics, domain, message, recipient, template, schedule, identity, inbound-route, account-user, token, invite, and webhook data
  approval: none; read-only
  Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

EXAMPLES
  # Inspect as a manual
  pm connectors inspect mailersend

  # Inspect as structured JSON
  pm connectors inspect mailersend --json

AGENT WORKFLOW
  - Run pm connectors inspect mailersend before creating credentials or plans.
  - Use --json only when the caller needs structured output; use the manual for human-readable guidance.
  - Never ask the user to paste secret values into chat.

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
