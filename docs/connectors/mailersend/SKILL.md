---
name: pm-mailersend
description: MailerSend connector knowledge and safe action guide.
---

# pm-mailersend

## Purpose

Reads MailerSend email activity, analytics, domains, messages, recipients, templates, scheduled messages, sender identities, inbound routes, users, invites, tokens, and webhooks through the MailerSend REST API.

## Icon

- asset: icons/mailersend.svg
- source: upstream_registry
- review_status: upstream_seeded
- review_url: https://developers.mailersend.com/

## Capabilities

- check=true catalog=true read=true write=false query=false
- Integration type: api

## Authentication

- Use pm credentials add with --from-env or --value-stdin for secret fields.

## Configuration

- base_url
- date_from
- date_to
- domain_id
- mode
- start_date
- api_token (secret)

## ETL Streams

- activity:
  - primary key: id
  - cursor: created_at
  - fields: created_at(), email(), id(), type(), updated_at()
- domains:
  - primary key: id
  - cursor: updated_at
  - fields: created_at(), dkim(), id(), is_dns_active(), is_verified(), name(), spf(), tracking(), updated_at()
- messages:
  - primary key: id
  - cursor: created_at
  - fields: created_at(), id(), updated_at()
- recipients:
  - primary key: id
  - cursor: updated_at
  - fields: created_at(), deleted_at(), email(), id(), updated_at()
- templates:
  - primary key: id
  - cursor: updated_at
  - fields: created_at(), description(), id(), image_path(), name(), tags(), type(), updated_at(), variables()
- scheduled_messages:
  - primary key: message_id
  - cursor: created_at
  - fields: created_at(), domain(), message(), message_id(), send_at(), status(), status_message(), subject()
- sender_identities:
  - primary key: id
  - fields: add_note(), domain(), email(), id(), is_verified(), name(), personal_note(), reply_to_email(), reply_to_name(), resends()
- inbound_routes:
  - primary key: id
  - fields: address(), dns_checked_at(), domain(), enabled(), filters(), forwards(), id(), mxValues(), name(), priority()
- account_users:
  - primary key: id
  - fields: created_at(), email(), id(), name(), permissions(), role(), status(), updated_at()
- invites:
  - primary key: id
  - cursor: updated_at
  - fields: created_at(), data(), email(), id(), permissions(), requires_periodic_password_change(), role(), updated_at()
- tokens:
  - primary key: id
  - fields: created_at(), id(), name(), scopes(), status()
- webhooks:
  - primary key: id
  - fields: created_at(), domain(), enabled(), events(), id(), name(), updated_at(), url()
- analytics_by_date:
  - primary key: date
  - fields: clicked(), clicked_unique(), date(), delivered(), hard_bounced(), opened(), opened_unique(), queued(), sent(), soft_bounced(), spam_complaints(), survey_opened(), survey_submitted(), unsubscribed()
- analytics_country:
  - primary key: name
  - fields: count(), name()
- analytics_user_agents:
  - primary key: name
  - fields: count(), name()
- analytics_reading_environment:
  - primary key: name
  - fields: count(), name()

## Sync Modes

- ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped

## Security

- read risk: external MailerSend API read of email activity, analytics, domain, message, recipient, template, schedule, identity, inbound-route, account-user, token, invite, and webhook data
- approval: none; read-only
- Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

## Commands

### Inspect as a manual

```bash
pm connectors inspect mailersend
```

### Inspect as structured JSON

```bash
pm connectors inspect mailersend --json
```

## Agent Rules

- Run pm connectors inspect mailersend before creating credentials or plans.
- Use --json only when the caller needs structured output; use the manual for human-readable guidance.
- Never ask the user to paste secret values into chat.
