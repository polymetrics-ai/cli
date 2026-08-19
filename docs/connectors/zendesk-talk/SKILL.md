---
name: pm-zendesk-talk
description: Zendesk Talk connector knowledge and safe action guide.
---

# pm-zendesk-talk

## Purpose

Reads Zendesk Talk phone numbers, greetings, greeting categories, IVRs, and agent activity statistics through the Zendesk Talk (voice) REST API.

## Icon

- id: zendesk-talk
- asset: icons/zendesk-talk.svg
- source: upstream_registry
- review_status: upstream_seeded
- review_url: https://support.zendesk.com/hc/en-us/sections/4405298889242-Developer-updates

## Capabilities

- check=true catalog=true read=true write=false query=false
- Integration type: api

## Authentication

- Use pm credentials add with --from-env or --value-stdin for secret fields.

## Configuration

- base_url (required)
- access_token (secret)
- api_token (secret)
- email (secret)

## ETL Streams

- phone_numbers:
  - primary key: id
  - cursor: created_at
  - fields: country_code(string), created_at(string), display_number(string), id(integer), nickname(string), number(string), recorded(boolean), sms_enabled(boolean), toll_free(boolean), voice_enabled(boolean)
- greetings:
  - primary key: id
  - fields: active(boolean), audio_name(string), audio_url(string), category_id(integer), default(boolean), has_sub_settings(boolean), id(integer), name(string)
- greeting_categories:
  - primary key: id
  - fields: id(integer), name(string)
- ivrs:
  - primary key: id
  - fields: id(integer), name(string), phone_number_ids(array), phone_number_names(array)
- agents_activity:
  - primary key: agent_id
  - fields: agent_id(integer), agent_state(string), available_time(integer), avatar_url(string), away_time(integer), call_status(string), calls_accepted(integer), calls_denied(integer), calls_missed(integer), forwarding_number(string), name(string), online_time(integer), total_call_duration(integer), total_talk_time(integer), total_wrap_up_time(integer), via(string)

## Sync Modes

- ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped

## Security

- read risk: external Zendesk Talk API read of phone number, greeting, and agent activity data
- Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

## Commands

### Inspect as a manual

```bash
pm connectors inspect zendesk-talk
```

### Inspect as structured JSON

```bash
pm connectors inspect zendesk-talk --json
```

## Agent Rules

- Run pm connectors inspect zendesk-talk before creating credentials or plans.
- Use --json only when the caller needs structured output; use the manual for human-readable guidance.
- Never ask the user to paste secret values into chat.
