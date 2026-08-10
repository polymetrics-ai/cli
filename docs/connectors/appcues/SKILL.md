---
name: pm-appcues
description: Appcues connector knowledge and safe action guide.
---

# pm-appcues

## Purpose

Reads and manages Appcues in-app guidance experiences (flows, Flows 2.0, pins, mobile experiences, launchpads, banners, checklists, embeds, NPS 2.0), audience data (segments, tags), operational resources (offline jobs, SDK authentication keys), and individual end-user/group profiles through the Appcues REST API v2.

## Icon

- id: pm-sample
- asset: icons/pm-sample.svg
- source: polymetrics
- review_status: polymetrics
- review_url: https://github.com/polymetrics-ai/cli

## Capabilities

- check=true catalog=true read=true write=true query=false
- Integration type: api

## Authentication

- Use pm credentials add with --from-env or --value-stdin for secret fields.

## Configuration

- account_id (required)
- base_url
- max_pages
- mode
- page_size
- username (required)
- password (secret) (required)

## ETL Streams

- flows:
  - primary key: id
  - cursor: updatedAt
  - fields: createdAt(string), createdBy(string), id(string), name(string), published(boolean), state(string), updatedAt(string), updatedBy(string)
- flows_v2:
  - primary key: id
  - cursor: updatedAt
  - fields: createdAt(string), createdBy(string), frequency(string), id(string), name(string), published(boolean), tag_ids(array), updatedAt(string), updatedBy(string)
- segments:
  - primary key: id
  - cursor: updatedAt
  - fields: createdAt(string), description(string), id(string), name(string), updatedAt(string)
- tags:
  - primary key: id
  - cursor: updatedAt
  - fields: createdAt(string), id(string), name(string), updatedAt(string)
- checklists:
  - primary key: id
  - cursor: updatedAt
  - fields: createdAt(string), id(string), name(string), published(boolean), state(string), updatedAt(string)
- banners:
  - primary key: id
  - cursor: updatedAt
  - fields: createdAt(string), id(string), name(string), published(boolean), state(string), updatedAt(string)
- pins:
  - primary key: id
  - cursor: updatedAt
  - fields: createdAt(string), frequency(string), id(string), name(string), published(boolean), tag_ids(array), type(string), updatedAt(string)
- mobile_experiences:
  - primary key: id
  - cursor: updatedAt
  - fields: createdAt(string), frequency(string), id(string), name(string), platform(string), published(boolean), updatedAt(string)
- launchpads:
  - primary key: id
  - cursor: updatedAt
  - fields: createdAt(string), frequency(string), id(string), name(string), published(boolean), tag_ids(array), type(string), updatedAt(string)
- embeds:
  - primary key: id
  - cursor: updatedAt
  - fields: createdAt(string), id(string), name(string), published(boolean), state(string), updatedAt(string)
- nps:
  - primary key: id
  - cursor: updatedAt
  - fields: createdAt(string), id(string), name(string), published(boolean), state(string), updatedAt(string)
- jobs:
  - primary key: id
  - fields: id(string), name(string), started_at(string), status(string), url(string)
- sdk_keys:
  - primary key: id
  - fields: created_at(string), id(string), name(string), tag_field(string)

## Sync Modes

- ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped

## Reverse ETL Actions

- publish_flow:
  - endpoint: POST /accounts/{{ config.account_id }}/flows/{{ record.id }}/publish
  - required fields: id
  - risk: publishes a flow, making it live to end users immediately
- unpublish_flow:
  - endpoint: POST /accounts/{{ config.account_id }}/flows/{{ record.id }}/unpublish
  - required fields: id
  - risk: unpublishes a live flow, immediately hiding it from end users
- publish_flow_v2:
  - endpoint: POST /accounts/{{ config.account_id }}/flows-v2/{{ record.id }}/publish
  - required fields: id
  - risk: publishes a Flows 2.0 experience, making it live to end users immediately
- unpublish_flow_v2:
  - endpoint: POST /accounts/{{ config.account_id }}/flows-v2/{{ record.id }}/unpublish
  - required fields: id
  - risk: unpublishes a live Flows 2.0 experience, immediately hiding it from end users
- publish_pin:
  - endpoint: POST /accounts/{{ config.account_id }}/pins/{{ record.id }}/publish
  - required fields: id
  - risk: publishes a pin, making it live to end users immediately
- unpublish_pin:
  - endpoint: POST /accounts/{{ config.account_id }}/pins/{{ record.id }}/unpublish
  - required fields: id
  - risk: unpublishes a live pin, immediately hiding it from end users
- publish_mobile_experience:
  - endpoint: POST /accounts/{{ config.account_id }}/mobile/{{ record.id }}/publish
  - required fields: id
  - risk: publishes a mobile experience, making it live to end users immediately
- unpublish_mobile_experience:
  - endpoint: POST /accounts/{{ config.account_id }}/mobile/{{ record.id }}/unpublish
  - required fields: id
  - risk: unpublishes a live mobile experience, immediately hiding it from end users
- publish_launchpad:
  - endpoint: POST /accounts/{{ config.account_id }}/launchpads/{{ record.id }}/publish
  - required fields: id
  - risk: publishes a launchpad, making it live to end users immediately
- unpublish_launchpad:
  - endpoint: POST /accounts/{{ config.account_id }}/launchpads/{{ record.id }}/unpublish
  - required fields: id
  - risk: unpublishes a live launchpad, immediately hiding it from end users
- publish_banner:
  - endpoint: POST /accounts/{{ config.account_id }}/banners/{{ record.id }}/publish
  - required fields: id
  - risk: publishes a banner, making it live to end users immediately
- unpublish_banner:
  - endpoint: POST /accounts/{{ config.account_id }}/banners/{{ record.id }}/unpublish
  - required fields: id
  - risk: unpublishes a live banner, immediately hiding it from end users
- publish_checklist:
  - endpoint: POST /accounts/{{ config.account_id }}/checklists/{{ record.id }}/publish
  - required fields: id
  - risk: publishes a checklist, making it live to end users immediately
- unpublish_checklist:
  - endpoint: POST /accounts/{{ config.account_id }}/checklists/{{ record.id }}/unpublish
  - required fields: id
  - risk: unpublishes a live checklist, immediately hiding it from end users
- publish_embed:
  - endpoint: POST /accounts/{{ config.account_id }}/embeds/{{ record.id }}/publish
  - required fields: id
  - risk: publishes an embed, making it live to end users immediately
- unpublish_embed:
  - endpoint: POST /accounts/{{ config.account_id }}/embeds/{{ record.id }}/unpublish
  - required fields: id
  - risk: unpublishes a live embed, immediately hiding it from end users
- publish_nps:
  - endpoint: POST /accounts/{{ config.account_id }}/nps/{{ record.id }}/publish
  - required fields: id
  - risk: publishes an NPS 2.0 survey, making it live to end users immediately
- unpublish_nps:
  - endpoint: POST /accounts/{{ config.account_id }}/nps/{{ record.id }}/unpublish
  - required fields: id
  - risk: unpublishes a live NPS 2.0 survey, immediately hiding it from end users
- create_segment:
  - endpoint: POST /accounts/{{ config.account_id }}/segments
  - required fields: name
  - risk: creates a new user segment used to target flows/banners/checklists
- update_segment:
  - endpoint: PATCH /accounts/{{ config.account_id }}/segments/{{ record.id }}
  - required fields: id
  - risk: mutates a user segment's definition, changing which users any flow/banner/checklist targeting it reaches
- delete_segment:
  - endpoint: DELETE /accounts/{{ config.account_id }}/segments/{{ record.id }}
  - required fields: id
  - risk: permanently deletes a user segment; any flow/banner/checklist targeting rule referencing it stops matching
- add_segment_user_ids:
  - endpoint: POST /accounts/{{ config.account_id }}/segments/{{ record.id }}/add_user_ids
  - required fields: id, user_ids
  - risk: adds specific end users to a segment (async job), changing who any targeting rule referencing it matches
- remove_segment_user_ids:
  - endpoint: POST /accounts/{{ config.account_id }}/segments/{{ record.id }}/remove_user_ids
  - required fields: id, user_ids
  - risk: removes specific end users from a segment (async job), changing who any targeting rule referencing it matches
- update_user_profile:
  - endpoint: PATCH /accounts/{{ config.account_id }}/users/{{ record.user_id }}/profile
  - required fields: user_id
  - risk: mutates an end user's profile attributes, changing which flows/segments they match
- delete_user_profile:
  - endpoint: DELETE /accounts/{{ config.account_id }}/users/{{ record.user_id }}/profile
  - required fields: user_id
  - risk: permanently deletes an end user's profile, properties, and flow/banner completion history (async job)
- track_user_event:
  - endpoint: POST /accounts/{{ config.account_id }}/users/{{ record.user_id }}/events
  - required fields: user_id, name
  - risk: injects a synthetic behavioral event into an end user's timeline, which may trigger flow/banner targeting rules
- update_group_profile:
  - endpoint: PATCH /accounts/{{ config.account_id }}/groups/{{ record.group_id }}/profile
  - required fields: group_id
  - risk: mutates a group's profile attributes, changing which flows/segments its members match
- associate_group_users:
  - endpoint: PATCH /accounts/{{ config.account_id }}/groups/{{ record.group_id }}/users
  - required fields: group_id, user_ids
  - risk: associates end users with a group, changing group-scoped targeting and analytics rollups
- create_sdk_key:
  - endpoint: POST /accounts/{{ config.account_id }}/sdk_keys
  - required fields: name
  - risk: creates a new SDK authentication key with production data-ingestion access
- update_sdk_key:
  - endpoint: PATCH /accounts/{{ config.account_id }}/sdk_keys/{{ record.id }}
  - required fields: id, tag_field
  - risk: changes an SDK key's tag field, altering how future ingested data is tagged
- delete_sdk_key:
  - endpoint: DELETE /accounts/{{ config.account_id }}/sdk_keys/{{ record.id }}
  - required fields: id
  - risk: permanently revokes an SDK authentication key; any client still using it immediately loses ingestion access
- enable_sdk_key_enforcement:
  - endpoint: POST /accounts/{{ config.account_id }}/sdk_keys/{{ record.id }}/enforcement_mode/enable
  - required fields: id
  - risk: enables strict enforcement mode on an SDK key, which can reject previously-accepted client requests
- disable_sdk_key_enforcement:
  - endpoint: POST /accounts/{{ config.account_id }}/sdk_keys/{{ record.id }}/enforcement_mode/disable
  - required fields: id
  - risk: disables strict enforcement mode on an SDK key
- enable_sdk_key_secure_data_ingest:
  - endpoint: POST /accounts/{{ config.account_id }}/sdk_keys/{{ record.id }}/secure_data_ingest/enable
  - required fields: id
  - risk: enables secure data ingest on an SDK key, which can reject unsigned client requests
- disable_sdk_key_secure_data_ingest:
  - endpoint: POST /accounts/{{ config.account_id }}/sdk_keys/{{ record.id }}/secure_data_ingest/disable
  - required fields: id
  - risk: disables secure data ingest on an SDK key

## Security

- read risk: external Appcues API read of in-app guidance and audience data
- write risk: external Appcues API mutation — publishes/unpublishes user-visible in-app experiences, manages segments and SDK keys, and mutates individual end-user/group profiles and event history
- approval: reverse ETL plan approval required before writes
- Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

## Commands

### Inspect as a manual

```bash
pm connectors inspect appcues
```

### Inspect as structured JSON

```bash
pm connectors inspect appcues --json
```

## Agent Rules

- Run pm connectors inspect appcues before creating credentials or plans.
- Use --json only when the caller needs structured output; use the manual for human-readable guidance.
- Never ask the user to paste secret values into chat.
- For reverse ETL writes, create a plan, show the preview, wait for explicit approval, then run with the approval token.
