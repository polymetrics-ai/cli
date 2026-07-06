# pm connectors inspect appcues

```text
NAME
  pm connectors inspect appcues - Appcues connector manual

SYNOPSIS
  pm connectors inspect appcues
  pm connectors inspect appcues --json
  pm credentials add <name> --connector appcues [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Reads and manages Appcues in-app guidance experiences (flows, Flows 2.0, pins, mobile experiences, launchpads, banners, checklists, embeds, NPS 2.0), audience data (segments, tags), operational resources (offline jobs, SDK authentication keys), and individual end-user/group profiles through the Appcues REST API v2.

ICON
  asset: icons/pm-sample.svg
  source: polymetrics
  review_status: polymetrics

CAPABILITIES
  check=true catalog=true read=true write=true query=false
  Integration type: api

AUTHENTICATION
  Use pm credentials add with --from-env or --value-stdin for secret fields.

CONFIGURATION
  account_id
  base_url
  max_pages
  mode
  page_size
  username
  password (secret)

ETL STREAMS
  flows:
    primary key: id
    cursor: updatedAt
    fields: createdAt(), createdBy(), id(), name(), published(), state(), updatedAt(), updatedBy()
  flows_v2:
    primary key: id
    cursor: updatedAt
    fields: createdAt(), createdBy(), frequency(), id(), name(), published(), tag_ids(), updatedAt(), updatedBy()
  segments:
    primary key: id
    cursor: updatedAt
    fields: createdAt(), description(), id(), name(), updatedAt()
  tags:
    primary key: id
    cursor: updatedAt
    fields: createdAt(), id(), name(), updatedAt()
  checklists:
    primary key: id
    cursor: updatedAt
    fields: createdAt(), id(), name(), published(), state(), updatedAt()
  banners:
    primary key: id
    cursor: updatedAt
    fields: createdAt(), id(), name(), published(), state(), updatedAt()
  pins:
    primary key: id
    cursor: updatedAt
    fields: createdAt(), frequency(), id(), name(), published(), tag_ids(), type(), updatedAt()
  mobile_experiences:
    primary key: id
    cursor: updatedAt
    fields: createdAt(), frequency(), id(), name(), platform(), published(), updatedAt()
  launchpads:
    primary key: id
    cursor: updatedAt
    fields: createdAt(), frequency(), id(), name(), published(), tag_ids(), type(), updatedAt()
  embeds:
    primary key: id
    cursor: updatedAt
    fields: createdAt(), id(), name(), published(), state(), updatedAt()
  nps:
    primary key: id
    cursor: updatedAt
    fields: createdAt(), id(), name(), published(), state(), updatedAt()
  jobs:
    primary key: id
    fields: id(), name(), started_at(), status(), url()
  sdk_keys:
    primary key: id
    fields: created_at(), id(), name(), tag_field()

SYNC MODES
  ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped

REVERSE ETL ACTIONS
  publish_flow:
    endpoint: POST /accounts/{{ config.account_id }}/flows/{{ record.id }}/publish
    required fields: id
    risk: publishes a flow, making it live to end users immediately
  unpublish_flow:
    endpoint: POST /accounts/{{ config.account_id }}/flows/{{ record.id }}/unpublish
    required fields: id
    risk: unpublishes a live flow, immediately hiding it from end users
  publish_flow_v2:
    endpoint: POST /accounts/{{ config.account_id }}/flows-v2/{{ record.id }}/publish
    required fields: id
    risk: publishes a Flows 2.0 experience, making it live to end users immediately
  unpublish_flow_v2:
    endpoint: POST /accounts/{{ config.account_id }}/flows-v2/{{ record.id }}/unpublish
    required fields: id
    risk: unpublishes a live Flows 2.0 experience, immediately hiding it from end users
  publish_pin:
    endpoint: POST /accounts/{{ config.account_id }}/pins/{{ record.id }}/publish
    required fields: id
    risk: publishes a pin, making it live to end users immediately
  unpublish_pin:
    endpoint: POST /accounts/{{ config.account_id }}/pins/{{ record.id }}/unpublish
    required fields: id
    risk: unpublishes a live pin, immediately hiding it from end users
  publish_mobile_experience:
    endpoint: POST /accounts/{{ config.account_id }}/mobile/{{ record.id }}/publish
    required fields: id
    risk: publishes a mobile experience, making it live to end users immediately
  unpublish_mobile_experience:
    endpoint: POST /accounts/{{ config.account_id }}/mobile/{{ record.id }}/unpublish
    required fields: id
    risk: unpublishes a live mobile experience, immediately hiding it from end users
  publish_launchpad:
    endpoint: POST /accounts/{{ config.account_id }}/launchpads/{{ record.id }}/publish
    required fields: id
    risk: publishes a launchpad, making it live to end users immediately
  unpublish_launchpad:
    endpoint: POST /accounts/{{ config.account_id }}/launchpads/{{ record.id }}/unpublish
    required fields: id
    risk: unpublishes a live launchpad, immediately hiding it from end users
  publish_banner:
    endpoint: POST /accounts/{{ config.account_id }}/banners/{{ record.id }}/publish
    required fields: id
    risk: publishes a banner, making it live to end users immediately
  unpublish_banner:
    endpoint: POST /accounts/{{ config.account_id }}/banners/{{ record.id }}/unpublish
    required fields: id
    risk: unpublishes a live banner, immediately hiding it from end users
  publish_checklist:
    endpoint: POST /accounts/{{ config.account_id }}/checklists/{{ record.id }}/publish
    required fields: id
    risk: publishes a checklist, making it live to end users immediately
  unpublish_checklist:
    endpoint: POST /accounts/{{ config.account_id }}/checklists/{{ record.id }}/unpublish
    required fields: id
    risk: unpublishes a live checklist, immediately hiding it from end users
  publish_embed:
    endpoint: POST /accounts/{{ config.account_id }}/embeds/{{ record.id }}/publish
    required fields: id
    risk: publishes an embed, making it live to end users immediately
  unpublish_embed:
    endpoint: POST /accounts/{{ config.account_id }}/embeds/{{ record.id }}/unpublish
    required fields: id
    risk: unpublishes a live embed, immediately hiding it from end users
  publish_nps:
    endpoint: POST /accounts/{{ config.account_id }}/nps/{{ record.id }}/publish
    required fields: id
    risk: publishes an NPS 2.0 survey, making it live to end users immediately
  unpublish_nps:
    endpoint: POST /accounts/{{ config.account_id }}/nps/{{ record.id }}/unpublish
    required fields: id
    risk: unpublishes a live NPS 2.0 survey, immediately hiding it from end users
  create_segment:
    endpoint: POST /accounts/{{ config.account_id }}/segments
    risk: creates a new user segment used to target flows/banners/checklists
  update_segment:
    endpoint: PATCH /accounts/{{ config.account_id }}/segments/{{ record.id }}
    required fields: id
    risk: mutates a user segment's definition, changing which users any flow/banner/checklist targeting it reaches
  delete_segment:
    endpoint: DELETE /accounts/{{ config.account_id }}/segments/{{ record.id }}
    required fields: id
    risk: permanently deletes a user segment; any flow/banner/checklist targeting rule referencing it stops matching
  add_segment_user_ids:
    endpoint: POST /accounts/{{ config.account_id }}/segments/{{ record.id }}/add_user_ids
    required fields: id
    optional fields: user_ids
    risk: adds specific end users to a segment (async job), changing who any targeting rule referencing it matches
  remove_segment_user_ids:
    endpoint: POST /accounts/{{ config.account_id }}/segments/{{ record.id }}/remove_user_ids
    required fields: id
    optional fields: user_ids
    risk: removes specific end users from a segment (async job), changing who any targeting rule referencing it matches
  update_user_profile:
    endpoint: PATCH /accounts/{{ config.account_id }}/users/{{ record.user_id }}/profile
    required fields: user_id
    risk: mutates an end user's profile attributes, changing which flows/segments they match
  delete_user_profile:
    endpoint: DELETE /accounts/{{ config.account_id }}/users/{{ record.user_id }}/profile
    required fields: user_id
    risk: permanently deletes an end user's profile, properties, and flow/banner completion history (async job)
  track_user_event:
    endpoint: POST /accounts/{{ config.account_id }}/users/{{ record.user_id }}/events
    required fields: user_id
    risk: injects a synthetic behavioral event into an end user's timeline, which may trigger flow/banner targeting rules
  update_group_profile:
    endpoint: PATCH /accounts/{{ config.account_id }}/groups/{{ record.group_id }}/profile
    required fields: group_id
    risk: mutates a group's profile attributes, changing which flows/segments its members match
  associate_group_users:
    endpoint: PATCH /accounts/{{ config.account_id }}/groups/{{ record.group_id }}/users
    required fields: group_id
    optional fields: user_ids
    risk: associates end users with a group, changing group-scoped targeting and analytics rollups
  create_sdk_key:
    endpoint: POST /accounts/{{ config.account_id }}/sdk_keys
    risk: creates a new SDK authentication key with production data-ingestion access
  update_sdk_key:
    endpoint: PATCH /accounts/{{ config.account_id }}/sdk_keys/{{ record.id }}
    required fields: id
    risk: changes an SDK key's tag field, altering how future ingested data is tagged
  delete_sdk_key:
    endpoint: DELETE /accounts/{{ config.account_id }}/sdk_keys/{{ record.id }}
    required fields: id
    risk: permanently revokes an SDK authentication key; any client still using it immediately loses ingestion access
  enable_sdk_key_enforcement:
    endpoint: POST /accounts/{{ config.account_id }}/sdk_keys/{{ record.id }}/enforcement_mode/enable
    required fields: id
    risk: enables strict enforcement mode on an SDK key, which can reject previously-accepted client requests
  disable_sdk_key_enforcement:
    endpoint: POST /accounts/{{ config.account_id }}/sdk_keys/{{ record.id }}/enforcement_mode/disable
    required fields: id
    risk: disables strict enforcement mode on an SDK key
  enable_sdk_key_secure_data_ingest:
    endpoint: POST /accounts/{{ config.account_id }}/sdk_keys/{{ record.id }}/secure_data_ingest/enable
    required fields: id
    risk: enables secure data ingest on an SDK key, which can reject unsigned client requests
  disable_sdk_key_secure_data_ingest:
    endpoint: POST /accounts/{{ config.account_id }}/sdk_keys/{{ record.id }}/secure_data_ingest/disable
    required fields: id
    risk: disables secure data ingest on an SDK key

SECURITY
  read risk: external Appcues API read of in-app guidance and audience data
  write risk: external Appcues API mutation — publishes/unpublishes user-visible in-app experiences, manages segments and SDK keys, and mutates individual end-user/group profiles and event history
  approval: reverse ETL plan approval required before writes
  Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

EXAMPLES
  # Inspect as a manual
  pm connectors inspect appcues

  # Inspect as structured JSON
  pm connectors inspect appcues --json

AGENT WORKFLOW
  - Run pm connectors inspect appcues before creating credentials or plans.
  - Use --json only when the caller needs structured output; use the manual for human-readable guidance.
  - Never ask the user to paste secret values into chat.
  - For reverse ETL writes, create a plan, show the preview, wait for explicit approval, then run with the approval token.

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
