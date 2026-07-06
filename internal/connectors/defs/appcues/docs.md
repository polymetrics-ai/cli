# Overview

Reads and manages Appcues in-app guidance experiences (flows, Flows 2.0, pins, mobile experiences,
launchpads, banners, checklists, embeds, NPS 2.0), audience data (segments, tags), operational
resources (offline jobs, SDK authentication keys), and individual end-user/group profiles through
the Appcues REST API v2.

Readable streams: `flows`, `flows_v2`, `segments`, `tags`, `checklists`, `banners`, `pins`,
`mobile_experiences`, `launchpads`, `embeds`, `nps`, `jobs`, `sdk_keys`.

Write actions: `publish_flow`, `unpublish_flow`, `publish_flow_v2`, `unpublish_flow_v2`,
`publish_pin`, `unpublish_pin`, `publish_mobile_experience`, `unpublish_mobile_experience`,
`publish_launchpad`, `unpublish_launchpad`, `publish_banner`, `unpublish_banner`,
`publish_checklist`, `unpublish_checklist`, `publish_embed`, `unpublish_embed`, `publish_nps`,
`unpublish_nps`, `create_segment`, `update_segment`, `delete_segment`, `add_segment_user_ids`,
`remove_segment_user_ids`, `update_user_profile`, `delete_user_profile`, `track_user_event`,
`update_group_profile`, `associate_group_users`, `create_sdk_key`, `update_sdk_key`,
`delete_sdk_key`, `enable_sdk_key_enforcement`, `disable_sdk_key_enforcement`,
`enable_sdk_key_secure_data_ingest`, `disable_sdk_key_secure_data_ingest`.

Service API documentation: https://docs.appcues.com/en_US/api.

## Auth setup

Connection fields:

- `account_id` (required, string); Appcues account ID; every resource path is scoped to
  accounts/{account_id}/.
- `base_url` (optional, string); default `https://api.appcues.com/v2`; format `uri`; Appcues API
  base URL override for tests or proxies.
- `max_pages` (optional, string); default `0`; Maximum pages; use 0, all, or unlimited to exhaust
  the stream.
- `mode` (optional, string).
- `page_size` (optional, string); default `100`; Records per page (1-1000).
- `password` (required, secret, string); Appcues API secret, used as the HTTP Basic auth password;
  never logged.
- `username` (required, string); Appcues API key, used as the HTTP Basic auth username.

Secret fields are redacted in logs and write previews: `password`.

Default configuration values: `base_url=https://api.appcues.com/v2`, `max_pages=0`, `page_size=100`.

Authentication behavior:

- HTTP Basic authentication using `config.username`, `secrets.password`.

Requests use the configured `base_url` value after applying defaults.

Connection checks call GET `/accounts/{{ config.account_id }}/flows`.

## Streams notes

Default pagination: page-number pagination; page parameter `page`; size parameter `limit`; starts at
1; page size 100.

- `flows`: GET `/accounts/{{ config.account_id }}/flows` - records at response root; page-number
  pagination; page parameter `page`; size parameter `limit`; starts at 1; page size 100.
- `flows_v2`: GET `/accounts/{{ config.account_id }}/flows-v2` - records at response root;
  page-number pagination; page parameter `page`; size parameter `limit`; starts at 1; page size 100.
- `segments`: GET `/accounts/{{ config.account_id }}/segments` - records at response root;
  page-number pagination; page parameter `page`; size parameter `limit`; starts at 1; page size 100.
- `tags`: GET `/accounts/{{ config.account_id }}/tags` - records at response root; page-number
  pagination; page parameter `page`; size parameter `limit`; starts at 1; page size 100.
- `checklists`: GET `/accounts/{{ config.account_id }}/checklists` - records at response root;
  page-number pagination; page parameter `page`; size parameter `limit`; starts at 1; page size 100.
- `banners`: GET `/accounts/{{ config.account_id }}/banners` - records at response root; page-number
  pagination; page parameter `page`; size parameter `limit`; starts at 1; page size 100.
- `pins`: GET `/accounts/{{ config.account_id }}/pins` - records at response root; page-number
  pagination; page parameter `page`; size parameter `limit`; starts at 1; page size 100.
- `mobile_experiences`: GET `/accounts/{{ config.account_id }}/mobile` - records at response root;
  page-number pagination; page parameter `page`; size parameter `limit`; starts at 1; page size 100.
- `launchpads`: GET `/accounts/{{ config.account_id }}/launchpads` - records at response root;
  page-number pagination; page parameter `page`; size parameter `limit`; starts at 1; page size 100.
- `embeds`: GET `/accounts/{{ config.account_id }}/embeds` - records at response root; page-number
  pagination; page parameter `page`; size parameter `limit`; starts at 1; page size 100.
- `nps`: GET `/accounts/{{ config.account_id }}/nps` - records at response root; page-number
  pagination; page parameter `page`; size parameter `limit`; starts at 1; page size 100.
- `jobs`: GET `/accounts/{{ config.account_id }}/jobs` - records at response root; page-number
  pagination; page parameter `page`; size parameter `limit`; starts at 1; page size 100.
- `sdk_keys`: GET `/accounts/{{ config.account_id }}/sdk_keys` - records at response root;
  page-number pagination; page parameter `page`; size parameter `limit`; starts at 1; page size 100.

## Write actions & risks

Overall write risk: external Appcues API mutation - publishes/unpublishes user-visible in-app
experiences, manages segments and SDK keys, and mutates individual end-user/group profiles and event
history.

Reverse ETL writes should be planned, previewed, approved, and then executed. Declared actions:

- `publish_flow`: POST `/accounts/{{ config.account_id }}/flows/{{ record.id }}/publish` - kind
  `update`; body type `none`; path fields `id`; required record fields `id`; accepted fields `id`;
  risk: publishes a flow, making it live to end users immediately.
- `unpublish_flow`: POST `/accounts/{{ config.account_id }}/flows/{{ record.id }}/unpublish` - kind
  `update`; body type `none`; path fields `id`; required record fields `id`; accepted fields `id`;
  risk: unpublishes a live flow, immediately hiding it from end users.
- `publish_flow_v2`: POST `/accounts/{{ config.account_id }}/flows-v2/{{ record.id }}/publish` -
  kind `update`; body type `none`; path fields `id`; required record fields `id`; accepted fields
  `id`; risk: publishes a Flows 2.0 experience, making it live to end users immediately.
- `unpublish_flow_v2`: POST `/accounts/{{ config.account_id }}/flows-v2/{{ record.id }}/unpublish` -
  kind `update`; body type `none`; path fields `id`; required record fields `id`; accepted fields
  `id`; risk: unpublishes a live Flows 2.0 experience, immediately hiding it from end users.
- `publish_pin`: POST `/accounts/{{ config.account_id }}/pins/{{ record.id }}/publish` - kind
  `update`; body type `none`; path fields `id`; required record fields `id`; accepted fields `id`;
  risk: publishes a pin, making it live to end users immediately.
- `unpublish_pin`: POST `/accounts/{{ config.account_id }}/pins/{{ record.id }}/unpublish` - kind
  `update`; body type `none`; path fields `id`; required record fields `id`; accepted fields `id`;
  risk: unpublishes a live pin, immediately hiding it from end users.
- `publish_mobile_experience`: POST `/accounts/{{ config.account_id }}/mobile/{{ record.id
  }}/publish` - kind `update`; body type `none`; path fields `id`; required record fields `id`;
  accepted fields `id`; risk: publishes a mobile experience, making it live to end users
  immediately.
- `unpublish_mobile_experience`: POST `/accounts/{{ config.account_id }}/mobile/{{ record.id
  }}/unpublish` - kind `update`; body type `none`; path fields `id`; required record fields `id`;
  accepted fields `id`; risk: unpublishes a live mobile experience, immediately hiding it from end
  users.
- `publish_launchpad`: POST `/accounts/{{ config.account_id }}/launchpads/{{ record.id }}/publish` -
  kind `update`; body type `none`; path fields `id`; required record fields `id`; accepted fields
  `id`; risk: publishes a launchpad, making it live to end users immediately.
- `unpublish_launchpad`: POST `/accounts/{{ config.account_id }}/launchpads/{{ record.id
  }}/unpublish` - kind `update`; body type `none`; path fields `id`; required record fields `id`;
  accepted fields `id`; risk: unpublishes a live launchpad, immediately hiding it from end users.
- `publish_banner`: POST `/accounts/{{ config.account_id }}/banners/{{ record.id }}/publish` - kind
  `update`; body type `none`; path fields `id`; required record fields `id`; accepted fields `id`;
  risk: publishes a banner, making it live to end users immediately.
- `unpublish_banner`: POST `/accounts/{{ config.account_id }}/banners/{{ record.id }}/unpublish` -
  kind `update`; body type `none`; path fields `id`; required record fields `id`; accepted fields
  `id`; risk: unpublishes a live banner, immediately hiding it from end users.
- `publish_checklist`: POST `/accounts/{{ config.account_id }}/checklists/{{ record.id }}/publish` -
  kind `update`; body type `none`; path fields `id`; required record fields `id`; accepted fields
  `id`; risk: publishes a checklist, making it live to end users immediately.
- `unpublish_checklist`: POST `/accounts/{{ config.account_id }}/checklists/{{ record.id
  }}/unpublish` - kind `update`; body type `none`; path fields `id`; required record fields `id`;
  accepted fields `id`; risk: unpublishes a live checklist, immediately hiding it from end users.
- `publish_embed`: POST `/accounts/{{ config.account_id }}/embeds/{{ record.id }}/publish` - kind
  `update`; body type `none`; path fields `id`; required record fields `id`; accepted fields `id`;
  risk: publishes an embed, making it live to end users immediately.
- `unpublish_embed`: POST `/accounts/{{ config.account_id }}/embeds/{{ record.id }}/unpublish` -
  kind `update`; body type `none`; path fields `id`; required record fields `id`; accepted fields
  `id`; risk: unpublishes a live embed, immediately hiding it from end users.
- `publish_nps`: POST `/accounts/{{ config.account_id }}/nps/{{ record.id }}/publish` - kind
  `update`; body type `none`; path fields `id`; required record fields `id`; accepted fields `id`;
  risk: publishes an NPS 2.0 survey, making it live to end users immediately.
- `unpublish_nps`: POST `/accounts/{{ config.account_id }}/nps/{{ record.id }}/unpublish` - kind
  `update`; body type `none`; path fields `id`; required record fields `id`; accepted fields `id`;
  risk: unpublishes a live NPS 2.0 survey, immediately hiding it from end users.
- `create_segment`: POST `/accounts/{{ config.account_id }}/segments` - kind `create`; body type
  `json`; required record fields `name`; accepted fields `description`, `name`; risk: creates a new
  user segment used to target flows/banners/checklists.
- `update_segment`: PATCH `/accounts/{{ config.account_id }}/segments/{{ record.id }}` - kind
  `update`; body type `json`; path fields `id`; required record fields `id`; accepted fields
  `description`, `id`, `name`; risk: mutates a user segment's definition, changing which users any
  flow/banner/checklist targeting it reaches.
- `delete_segment`: DELETE `/accounts/{{ config.account_id }}/segments/{{ record.id }}` - kind
  `delete`; body type `none`; path fields `id`; required record fields `id`; accepted fields `id`;
  missing records treated as success for status `404`; risk: permanently deletes a user segment; any
  flow/banner/checklist targeting rule referencing it stops matching.
- `add_segment_user_ids`: POST `/accounts/{{ config.account_id }}/segments/{{ record.id
  }}/add_user_ids` - kind `update`; body type `json`; path fields `id`; body fields `user_ids`;
  required record fields `id`, `user_ids`; accepted fields `id`, `user_ids`; risk: adds specific end
  users to a segment (async job), changing who any targeting rule referencing it matches.
- `remove_segment_user_ids`: POST `/accounts/{{ config.account_id }}/segments/{{ record.id
  }}/remove_user_ids` - kind `update`; body type `json`; path fields `id`; body fields `user_ids`;
  required record fields `id`, `user_ids`; accepted fields `id`, `user_ids`; risk: removes specific
  end users from a segment (async job), changing who any targeting rule referencing it matches.
- `update_user_profile`: PATCH `/accounts/{{ config.account_id }}/users/{{ record.user_id
  }}/profile` - kind `update`; body type `json`; path fields `user_id`; required record fields
  `user_id`; accepted fields `user_id`; risk: mutates an end user's profile attributes, changing
  which flows/segments they match.
- `delete_user_profile`: DELETE `/accounts/{{ config.account_id }}/users/{{ record.user_id
  }}/profile` - kind `delete`; body type `none`; path fields `user_id`; required record fields
  `user_id`; accepted fields `user_id`; missing records treated as success for status `404`; risk:
  permanently deletes an end user's profile, properties, and flow/banner completion history (async
  job).
- `track_user_event`: POST `/accounts/{{ config.account_id }}/users/{{ record.user_id }}/events` -
  kind `create`; body type `json`; path fields `user_id`; required record fields `user_id`, `name`;
  accepted fields `attributes`, `group_id`, `name`, `timestamp`, `user_id`.
- `update_group_profile`: PATCH `/accounts/{{ config.account_id }}/groups/{{ record.group_id
  }}/profile` - kind `update`; body type `json`; path fields `group_id`; required record fields
  `group_id`; accepted fields `group_id`; risk: mutates a group's profile attributes, changing which
  flows/segments its members match.
- `associate_group_users`: PATCH `/accounts/{{ config.account_id }}/groups/{{ record.group_id
  }}/users` - kind `update`; body type `json`; path fields `group_id`; body fields `user_ids`;
  required record fields `group_id`, `user_ids`; accepted fields `group_id`, `user_ids`; risk:
  associates end users with a group, changing group-scoped targeting and analytics rollups.
- `create_sdk_key`: POST `/accounts/{{ config.account_id }}/sdk_keys` - kind `create`; body type
  `json`; required record fields `name`; accepted fields `name`, `tag_field`; risk: creates a new
  SDK authentication key with production data-ingestion access.
- `update_sdk_key`: PATCH `/accounts/{{ config.account_id }}/sdk_keys/{{ record.id }}` - kind
  `update`; body type `json`; path fields `id`; required record fields `id`, `tag_field`; accepted
  fields `id`, `tag_field`; risk: changes an SDK key's tag field, altering how future ingested data
  is tagged.
- `delete_sdk_key`: DELETE `/accounts/{{ config.account_id }}/sdk_keys/{{ record.id }}` - kind
  `delete`; body type `none`; path fields `id`; required record fields `id`; accepted fields `id`;
  missing records treated as success for status `404`; risk: permanently revokes an SDK
  authentication key; any client still using it immediately loses ingestion access.
- `enable_sdk_key_enforcement`: POST `/accounts/{{ config.account_id }}/sdk_keys/{{ record.id
  }}/enforcement_mode/enable` - kind `update`; body type `none`; path fields `id`; required record
  fields `id`; accepted fields `id`; risk: enables strict enforcement mode on an SDK key, which can
  reject previously-accepted client requests.
- `disable_sdk_key_enforcement`: POST `/accounts/{{ config.account_id }}/sdk_keys/{{ record.id
  }}/enforcement_mode/disable` - kind `update`; body type `none`; path fields `id`; required record
  fields `id`; accepted fields `id`; risk: disables strict enforcement mode on an SDK key.
- `enable_sdk_key_secure_data_ingest`: POST `/accounts/{{ config.account_id }}/sdk_keys/{{ record.id
  }}/secure_data_ingest/enable` - kind `update`; body type `none`; path fields `id`; required record
  fields `id`; accepted fields `id`; risk: enables secure data ingest on an SDK key, which can
  reject unsigned client requests.
- `disable_sdk_key_secure_data_ingest`: POST `/accounts/{{ config.account_id }}/sdk_keys/{{
  record.id }}/secure_data_ingest/disable` - kind `update`; body type `none`; path fields `id`;
  required record fields `id`; accepted fields `id`; risk: disables secure data ingest on an SDK
  key.

## Known limits

- Batch defaults: read_page_size=100.
- API coverage includes 13 stream-backed endpoint group(s), 35 write-backed endpoint group(s).
- Other documented endpoints are not exposed by this connector where they are classified as
  duplicate_of=12, out_of_scope=11.
