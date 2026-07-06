# Overview

Reads Tavus faces (replicas), videos, conversations, PALs, guardrails, objectives, documents,
pronunciation dictionaries, voices, and skills, and writes approved
video/conversation/PAL/guardrail/objective/document/pronunciation-dictionary create-delete mutations
through the Tavus API.

Readable streams: `replicas`, `videos`, `conversations`, `pals`, `guardrails`, `objectives`,
`documents`, `pronunciation_dictionaries`, `voices`, `skills`.

Write actions: `create_video`, `delete_video`, `create_conversation`, `end_conversation`,
`delete_conversation`, `create_pal`, `delete_pal`, `create_guardrail`, `delete_guardrail`,
`create_objective`, `delete_objective`, `create_document`, `delete_document`,
`create_pronunciation_dictionary`, `delete_pronunciation_dictionary`.

Service API documentation: https://docs.tavus.io/.

## Auth setup

Connection fields:

- `api_key` (required, secret, string); Tavus API key, sent as the x-api-key header. Never logged.
- `base_url` (optional, string); default `https://tavusapi.com/v2`; format `uri`; Tavus API base URL
  override for tests or proxies.

Secret fields are redacted in logs and write previews: `api_key`.

Default configuration values: `base_url=https://tavusapi.com/v2`.

Authentication behavior:

- API key authentication in `x-api-key` using `secrets.api_key`.

Requests use the configured `base_url` value after applying defaults.

Connection checks call GET `/replicas`.

## Streams notes

Default pagination: page-number pagination; page parameter `page`; size parameter `page_size`;
starts at 1; page size 100.

Pagination by stream: none: `skills`; page_number: `replicas`, `videos`, `conversations`, `pals`,
`guardrails`, `objectives`, `documents`, `pronunciation_dictionaries`, `voices`.

- `replicas`: GET `/replicas` - records path `data`; page-number pagination; page parameter `page`;
  size parameter `page_size`; starts at 1; page size 100; computed output fields `id`, `name`.
- `videos`: GET `/videos` - records path `data`; page-number pagination; page parameter `page`; size
  parameter `limit`; starts at 0; page size 100; computed output fields `id`, `name`.
- `conversations`: GET `/conversations` - records path `data`; page-number pagination; page
  parameter `page`; size parameter `limit`; starts at 1; page size 100; computed output fields `id`,
  `name`.
- `pals`: GET `/pals` - records path `data`; page-number pagination; page parameter `page`; size
  parameter `limit`; starts at 1; page size 100; computed output fields `id`, `name`.
- `guardrails`: GET `/guardrails` - records path `data`; page-number pagination; page parameter
  `page`; size parameter `limit`; starts at 0; page size 100; computed output fields `id`, `name`.
- `objectives`: GET `/objectives` - records path `data`; page-number pagination; page parameter
  `page`; size parameter `limit`; starts at 1; page size 100; computed output fields `id`, `name`.
- `documents`: GET `/documents` - records path `data`; page-number pagination; page parameter
  `page`; size parameter `limit`; starts at 0; page size 100; computed output fields `id`, `name`.
- `pronunciation_dictionaries`: GET `/pronunciation-dictionaries` - records path `data`; page-number
  pagination; page parameter `page`; size parameter `limit`; starts at 0; page size 100; computed
  output fields `id`.
- `voices`: GET `/voices` - records path `data`; page-number pagination; page parameter `page`; size
  parameter `limit`; starts at 1; page size 100.
- `skills`: GET `/skills` - records path `data`.

## Write actions & risks

Overall write risk: external Tavus API mutation (create/delete videos, conversations, PALs,
guardrails, objectives, documents, pronunciation dictionaries; end conversations);
create_video/create_conversation consume billed generation/conversational minutes.

Reverse ETL writes should be planned, previewed, approved, and then executed. Declared actions:

- `create_video`: POST `/videos` - kind `create`; body type `json`; required record fields
  `replica_id`; accepted fields `audio_url`, `background_url`, `callback_url`, `replica_id`,
  `script`, `video_name`; risk: generates a new async video render from a face and script/audio;
  consumes video-generation minutes on the account.
- `delete_video`: DELETE `/videos/{{ record.id }}` - kind `delete`; body type `none`; path fields
  `id`; required record fields `id`; accepted fields `id`; missing records treated as success for
  status `400`; risk: permanently deletes a generated video and its hosted/download URLs.
- `create_conversation`: POST `/conversations` - kind `create`; body type `json`; accepted fields
  `audio_only`, `callback_url`, `conversation_name`, `conversational_context`, `face_id`, `pal_id`;
  risk: starts a real-time video conversation, which begins consuming conversational-minutes billing
  immediately and (unless test_mode) places a live call.
- `end_conversation`: POST `/conversations/{{ record.id }}/end` - kind `update`; body type `none`;
  path fields `id`; required record fields `id`; accepted fields `id`; risk: ends an active
  conversation for every participant; routine call cleanup, not destructive to conversation history
  (compare delete_conversation).
- `delete_conversation`: DELETE `/conversations/{{ record.id }}` - kind `delete`; body type `none`;
  path fields `id`; required record fields `id`; accepted fields `id`; missing records treated as
  success for status `400`; risk: permanently deletes a conversation and its recorded history; use
  end_conversation instead for routine call cleanup.
- `create_pal`: POST `/pals` - kind `create`; body type `json`; required record fields
  `default_face_id`; accepted fields `default_face_id`, `document_ids`, `guardrail_ids`, `pal_name`,
  `pipeline_mode`, `system_prompt`; risk: creates a new PAL persona; low-risk external mutation, no
  approval required.
- `delete_pal`: DELETE `/pals/{{ record.id }}` - kind `delete`; body type `none`; path fields `id`;
  required record fields `id`; accepted fields `id`; missing records treated as success for status
  `400`; risk: permanently deletes a PAL; any conversation still referencing its pal_id will fail to
  start.
- `create_guardrail`: POST `/guardrails` - kind `create`; body type `json`; required record fields
  `guardrail_name`, `guardrail_prompt`; accepted fields `guardrail_name`, `guardrail_prompt`,
  `modality`, `tags`; risk: creates a new behavioral guardrail; low-risk external mutation, no
  approval required.
- `delete_guardrail`: DELETE `/guardrails/{{ record.id }}` - kind `delete`; body type `none`; path
  fields `id`; required record fields `id`; accepted fields `id`; missing records treated as success
  for status `400`; risk: permanently deletes a guardrail; any PAL referencing it via guardrail_ids
  loses that behavioral boundary immediately.
- `create_objective`: POST `/objectives` - kind `create`; body type `json`; required record fields
  `data`; accepted fields `data`; risk: creates one or more new PAL objectives; low-risk external
  mutation, no approval required.
- `delete_objective`: DELETE `/objectives/{{ record.id }}` - kind `delete`; body type `none`; path
  fields `id`; required record fields `id`; accepted fields `id`; missing records treated as success
  for status `400`; risk: permanently deletes an objective; any PAL referencing it via objectives_id
  loses that goal-oriented instruction immediately.
- `create_document`: POST `/documents` - kind `create`; body type `json`; required record fields
  `document_url`; accepted fields `callback_url`, `document_name`, `document_url`, `tags`; risk:
  uploads a document to the knowledge base; processing is asynchronous and the document becomes
  available to PALs only once status reaches ready.
- `delete_document`: DELETE `/documents/{{ record.id }}` - kind `delete`; body type `none`; path
  fields `id`; required record fields `id`; accepted fields `id`; missing records treated as success
  for status `400`; risk: permanently deletes a knowledge-base document and its processed data; any
  PAL referencing it via document_ids loses that knowledge source immediately.
- `create_pronunciation_dictionary`: POST `/pronunciation-dictionaries` - kind `create`; body type
  `json`; required record fields `name`; accepted fields `name`, `rules`; risk: creates a new
  pronunciation dictionary; low-risk external mutation, no approval required.
- `delete_pronunciation_dictionary`: DELETE `/pronunciation-dictionaries/{{ record.id }}` - kind
  `delete`; body type `none`; path fields `id`; required record fields `id`; accepted fields `id`;
  missing records treated as success for status `400`; risk: permanently deletes a pronunciation
  dictionary and removes it from every linked PAL.

## Known limits

- Batch defaults: read_page_size=100.
- API coverage includes 10 stream-backed endpoint group(s), 15 write-backed endpoint group(s).
- Other documented endpoints are not exposed by this connector where they are classified as
  deprecated=8, destructive_admin=6, duplicate_of=15, non_data_endpoint=3, out_of_scope=26.
