# Overview

Reads Shutterstock media, collection, license, editorial, catalog, contributor, and subscription
metadata; writes collection/lightbox metadata through safe collection endpoints.

Readable streams: `images`, `videos`, `audio`, `list_similar_images`, `get_image_recommendations`,
`get_updated_images`, `get_image_list`, `get_image`, `get_image_license_list`,
`get_image_collection_list`, `get_image_collection`, `get_image_collection_items`,
`find_similar_videos`, `get_updated_videos`, `get_video_list`, `get_video`,
`get_video_license_list`, `get_video_collection_list`, `get_video_collection`,
`get_video_collection_items`, `get_track_list`, `get_track`, `get_track_license_list`,
`get_track_collection_list`, `get_track_collection`, `get_track_collection_items`, `search_sfx`,
`get_sfx`, `get_sfx_list`, `get_sfx_license_list`, `search_editorial_images`,
`get_updated_editorial_images`, `get_editorial_image`, `get_editorial_image_list`,
`get_editorial_image_license_list`, `get_editorial_image_livefeeds`, `get_editorial_image_livefeed`,
`get_editorial_image_livefeed_items`, `get_editorial`, `get_editorial_livefeeds`,
`get_editorial_livefeed`, `get_editorial_livefeed_items`, `search_editorial`,
`get_updated_editorial`, `search_editorial_videos`, `get_editorial_video`,
`get_editorial_video_list`, `get_editorial_video_license_list`, `get_similar_cv_images`,
`get_similar_cv_videos`, `search_catalog`, `get_collections`, `get_contributors`, `get_contributor`,
`get_contributor_collections`, `get_contributor_collection`, `get_contributor_collection_items`,
`get_user_subscriptions`.

Write actions: `create_image_collection`, `rename_image_collection`, `delete_image_collection`,
`add_image_collection_items`, `create_video_collection`, `rename_video_collection`,
`delete_video_collection`, `add_video_collection_items`, `create_audio_collection`,
`rename_audio_collection`, `delete_audio_collection`, `add_audio_collection_items`,
`create_catalog_collection`, `update_catalog_collection`, `delete_catalog_collection`,
`add_catalog_collection_items`.

Service API documentation: https://api-reference.shutterstock.com/.

## Auth setup

Connection fields:

- `access_token` (optional, secret, string); Shutterstock OAuth access token, sent as a Bearer
  token.
- `audio_collection_ids` (optional, string); Comma-separated audio collection IDs for collection
  detail and items fan-out streams.
- `audio_ids` (optional, string); Comma-separated Shutterstock audio track IDs for audio detail and
  audio-list fan-out streams.
- `base_url` (optional, string); default `https://api.shutterstock.com`; format `uri`; Shutterstock
  API base URL.
- `category` (optional, string); Optional category filter passed through to media search streams.
- `contributor_collection_ids` (optional, string); Comma-separated contributor collection IDs used
  with contributor_id for nested contributor collection detail streams.
- `contributor_id` (optional, string); Contributor ID used with contributor_collection_ids for
  nested contributor collection detail streams.
- `contributor_ids` (optional, string); Comma-separated contributor IDs for contributor fan-out
  streams.
- `editorial_ids` (optional, string); Comma-separated editorial asset IDs for generic editorial
  detail fan-out streams.
- `editorial_image_ids` (optional, string); Comma-separated editorial image IDs for editorial image
  fan-out streams.
- `editorial_image_livefeed_ids` (optional, string); Comma-separated editorial image livefeed IDs
  for livefeed fan-out streams.
- `editorial_livefeed_ids` (optional, string); Comma-separated editorial livefeed IDs for livefeed
  fan-out streams.
- `editorial_video_ids` (optional, string); Comma-separated editorial video IDs for editorial video
  fan-out streams.
- `image_collection_ids` (optional, string); Comma-separated image collection IDs for collection
  detail and items fan-out streams.
- `image_ids` (optional, string); Comma-separated Shutterstock image IDs for image detail,
  recommendations, similar, and image-list fan-out streams.
- `orientation` (optional, string); Optional orientation filter passed through to media search
  streams.
- `query` (optional, string); Optional search query filter passed through to search-style streams.
- `sfx_ids` (optional, string); Comma-separated Shutterstock sound-effect IDs for SFX detail and
  SFX-list fan-out streams.
- `sort` (optional, string); Optional sort order filter passed through to search-style streams.
- `video_collection_ids` (optional, string); Comma-separated video collection IDs for collection
  detail and items fan-out streams.
- `video_ids` (optional, string); Comma-separated Shutterstock video IDs for video detail, similar,
  and video-list fan-out streams.
- `visual_asset_ids` (optional, string); Comma-separated asset IDs for computer-vision similar-media
  streams.

Secret fields are redacted in logs and write previews: `access_token`.

Default configuration values: `base_url=https://api.shutterstock.com`.

Authentication behavior:

- Bearer token authentication using `secrets.access_token`.

Requests use the configured `base_url` value after applying defaults.

Connection checks call GET `/v2/images/search`.

## Streams notes

Default pagination: page-number pagination; page parameter `page`; size parameter `per_page`; starts
at 1; page size 100.

Pagination by stream: none: `get_image_recommendations`, `get_image_list`, `get_image`,
`get_image_collection`, `get_video_list`, `get_video`, `get_video_collection`, `get_track_list`,
`get_track`, `get_track_collection`, `get_sfx`, `get_sfx_list`, `get_editorial_image`,
`get_editorial_image_list`, `get_editorial_image_livefeed`, `get_editorial`,
`get_editorial_livefeed`, `get_editorial_video`, `get_editorial_video_list`, `get_contributor`,
`get_contributor_collection`, `get_user_subscriptions`; page_number: `images`, `videos`, `audio`,
`list_similar_images`, `get_updated_images`, `get_image_license_list`, `get_image_collection_list`,
`get_image_collection_items`, `find_similar_videos`, `get_updated_videos`, `get_video_license_list`,
`get_video_collection_list`, `get_video_collection_items`, `get_track_license_list`,
`get_track_collection_list`, `get_track_collection_items`, `search_sfx`, `get_sfx_license_list`,
`search_editorial_images`, `get_updated_editorial_images`, `get_editorial_image_license_list`,
`get_editorial_image_livefeeds`, `get_editorial_image_livefeed_items`, `get_editorial_livefeeds`,
`get_editorial_livefeed_items`, `search_editorial`, `get_updated_editorial`,
`search_editorial_videos`, `get_editorial_video_license_list`, `get_similar_cv_images`,
`get_similar_cv_videos`, `search_catalog`, `get_collections`, `get_contributors`,
`get_contributor_collections`, `get_contributor_collection_items`.

- `images`: GET `/v2/images/search` - records path `data`; query `category` from template `{{
  config.category }}`, omitted when absent; `orientation` from template `{{ config.orientation }}`,
  omitted when absent; `query` from template `{{ config.query }}`, omitted when absent; `sort` from
  template `{{ config.sort }}`, omitted when absent; page-number pagination; page parameter `page`;
  size parameter `per_page`; starts at 1; page size 100; computed output fields `description`,
  `media_type`, `updated_at`.
- `videos`: GET `/v2/videos/search` - records path `data`; query `category` from template `{{
  config.category }}`, omitted when absent; `orientation` from template `{{ config.orientation }}`,
  omitted when absent; `query` from template `{{ config.query }}`, omitted when absent; `sort` from
  template `{{ config.sort }}`, omitted when absent; page-number pagination; page parameter `page`;
  size parameter `per_page`; starts at 1; page size 100; computed output fields `description`,
  `media_type`, `updated_at`.
- `audio`: GET `/v2/audio/search` - records path `data`; query `category` from template `{{
  config.category }}`, omitted when absent; `orientation` from template `{{ config.orientation }}`,
  omitted when absent; `query` from template `{{ config.query }}`, omitted when absent; `sort` from
  template `{{ config.sort }}`, omitted when absent; page-number pagination; page parameter `page`;
  size parameter `per_page`; starts at 1; page size 100; computed output fields `description`,
  `media_type`, `updated_at`.
- `list_similar_images`: GET `/v2/images/{{ fanout.id }}/similar` - records path `data`; page-number
  pagination; page parameter `page`; size parameter `per_page`; starts at 1; page size 100; fan-out;
  ids from config field `image_ids`; id inserted into the request path; stamps `image_id`; emits
  passthrough records.
- `get_image_recommendations`: GET `/v2/images/recommendations` - records path `data`; fan-out; ids
  from config field `image_ids`; id sent as query parameter `id`; stamps `image_id`; emits
  passthrough records.
- `get_updated_images`: GET `/v2/images/updated` - records path `data`; page-number pagination; page
  parameter `page`; size parameter `per_page`; starts at 1; page size 100; emits passthrough
  records.
- `get_image_list`: GET `/v2/images` - records path `data`; fan-out; ids from config field
  `image_ids`; id sent as query parameter `id`; stamps `image_id`; emits passthrough records.
- `get_image`: GET `/v2/images/{{ fanout.id }}` - single-object response; records at response root;
  fan-out; ids from config field `image_ids`; id inserted into the request path; stamps `image_id`;
  emits passthrough records.
- `get_image_license_list`: GET `/v2/images/licenses` - records path `data`; page-number pagination;
  page parameter `page`; size parameter `per_page`; starts at 1; page size 100; emits passthrough
  records.
- `get_image_collection_list`: GET `/v2/images/collections` - records path `data`; page-number
  pagination; page parameter `page`; size parameter `per_page`; starts at 1; page size 100; emits
  passthrough records.
- `get_image_collection`: GET `/v2/images/collections/{{ fanout.id }}` - single-object response;
  records at response root; fan-out; ids from config field `image_collection_ids`; id inserted into
  the request path; stamps `collection_id`; emits passthrough records.
- `get_image_collection_items`: GET `/v2/images/collections/{{ fanout.id }}/items` - records path
  `data`; page-number pagination; page parameter `page`; size parameter `per_page`; starts at 1;
  page size 100; fan-out; ids from config field `image_collection_ids`; id inserted into the request
  path; stamps `collection_id`; emits passthrough records.
- `find_similar_videos`: GET `/v2/videos/{{ fanout.id }}/similar` - records path `data`; page-number
  pagination; page parameter `page`; size parameter `per_page`; starts at 1; page size 100; fan-out;
  ids from config field `video_ids`; id inserted into the request path; stamps `video_id`; emits
  passthrough records.
- `get_updated_videos`: GET `/v2/videos/updated` - records path `data`; page-number pagination; page
  parameter `page`; size parameter `per_page`; starts at 1; page size 100; emits passthrough
  records.
- `get_video_list`: GET `/v2/videos` - records path `data`; fan-out; ids from config field
  `video_ids`; id sent as query parameter `id`; stamps `video_id`; emits passthrough records.
- `get_video`: GET `/v2/videos/{{ fanout.id }}` - single-object response; records at response root;
  fan-out; ids from config field `video_ids`; id inserted into the request path; stamps `video_id`;
  emits passthrough records.
- `get_video_license_list`: GET `/v2/videos/licenses` - records path `data`; page-number pagination;
  page parameter `page`; size parameter `per_page`; starts at 1; page size 100; emits passthrough
  records.
- `get_video_collection_list`: GET `/v2/videos/collections` - records path `data`; page-number
  pagination; page parameter `page`; size parameter `per_page`; starts at 1; page size 100; emits
  passthrough records.
- `get_video_collection`: GET `/v2/videos/collections/{{ fanout.id }}` - single-object response;
  records at response root; fan-out; ids from config field `video_collection_ids`; id inserted into
  the request path; stamps `collection_id`; emits passthrough records.
- `get_video_collection_items`: GET `/v2/videos/collections/{{ fanout.id }}/items` - records path
  `data`; page-number pagination; page parameter `page`; size parameter `per_page`; starts at 1;
  page size 100; fan-out; ids from config field `video_collection_ids`; id inserted into the request
  path; stamps `collection_id`; emits passthrough records.
- `get_track_list`: GET `/v2/audio` - records path `data`; fan-out; ids from config field
  `audio_ids`; id sent as query parameter `id`; stamps `audio_id`; emits passthrough records.
- `get_track`: GET `/v2/audio/{{ fanout.id }}` - single-object response; records at response root;
  fan-out; ids from config field `audio_ids`; id inserted into the request path; stamps `audio_id`;
  emits passthrough records.
- `get_track_license_list`: GET `/v2/audio/licenses` - records path `data`; page-number pagination;
  page parameter `page`; size parameter `per_page`; starts at 1; page size 100; emits passthrough
  records.
- `get_track_collection_list`: GET `/v2/audio/collections` - records path `data`; page-number
  pagination; page parameter `page`; size parameter `per_page`; starts at 1; page size 100; emits
  passthrough records.
- `get_track_collection`: GET `/v2/audio/collections/{{ fanout.id }}` - single-object response;
  records at response root; fan-out; ids from config field `audio_collection_ids`; id inserted into
  the request path; stamps `collection_id`; emits passthrough records.
- `get_track_collection_items`: GET `/v2/audio/collections/{{ fanout.id }}/items` - records path
  `data`; page-number pagination; page parameter `page`; size parameter `per_page`; starts at 1;
  page size 100; fan-out; ids from config field `audio_collection_ids`; id inserted into the request
  path; stamps `collection_id`; emits passthrough records.
- `search_sfx`: GET `/v2/sfx/search` - records path `data`; query `category` from template `{{
  config.category }}`, omitted when absent; `orientation` from template `{{ config.orientation }}`,
  omitted when absent; `query` from template `{{ config.query }}`, omitted when absent; `sort` from
  template `{{ config.sort }}`, omitted when absent; page-number pagination; page parameter `page`;
  size parameter `per_page`; starts at 1; page size 100; emits passthrough records.
- `get_sfx`: GET `/v2/sfx/{{ fanout.id }}` - single-object response; records at response root;
  fan-out; ids from config field `sfx_ids`; id inserted into the request path; stamps `sfx_id`;
  emits passthrough records.
- `get_sfx_list`: GET `/v2/sfx` - records path `data`; fan-out; ids from config field `sfx_ids`; id
  sent as query parameter `id`; stamps `sfx_id`; emits passthrough records.
- `get_sfx_license_list`: GET `/v2/sfx/licenses` - records path `data`; page-number pagination; page
  parameter `page`; size parameter `per_page`; starts at 1; page size 100; emits passthrough
  records.
- `search_editorial_images`: GET `/v2/editorial/images/search` - records path `data`; query
  `category` from template `{{ config.category }}`, omitted when absent; `orientation` from template
  `{{ config.orientation }}`, omitted when absent; `query` from template `{{ config.query }}`,
  omitted when absent; `sort` from template `{{ config.sort }}`, omitted when absent; page-number
  pagination; page parameter `page`; size parameter `per_page`; starts at 1; page size 100; emits
  passthrough records.
- `get_updated_editorial_images`: GET `/v2/editorial/images/updated` - records path `data`;
  page-number pagination; page parameter `page`; size parameter `per_page`; starts at 1; page size
  100; emits passthrough records.
- `get_editorial_image`: GET `/v2/editorial/images/{{ fanout.id }}` - single-object response;
  records at response root; fan-out; ids from config field `editorial_image_ids`; id inserted into
  the request path; stamps `editorial_image_id`; emits passthrough records.
- `get_editorial_image_list`: GET `/v2/editorial/images` - records path `data`; fan-out; ids from
  config field `editorial_image_ids`; id sent as query parameter `id`; stamps `editorial_image_id`;
  emits passthrough records.
- `get_editorial_image_license_list`: GET `/v2/editorial/images/licenses` - records path `data`;
  page-number pagination; page parameter `page`; size parameter `per_page`; starts at 1; page size
  100; emits passthrough records.
- `get_editorial_image_livefeeds`: GET `/v2/editorial/images/livefeeds` - records path `data`;
  page-number pagination; page parameter `page`; size parameter `per_page`; starts at 1; page size
  100; emits passthrough records.
- `get_editorial_image_livefeed`: GET `/v2/editorial/images/livefeeds/{{ fanout.id }}` -
  single-object response; records at response root; fan-out; ids from config field
  `editorial_image_livefeed_ids`; id inserted into the request path; stamps `livefeed_id`; emits
  passthrough records.
- `get_editorial_image_livefeed_items`: GET `/v2/editorial/images/livefeeds/{{ fanout.id }}/items` -
  records path `data`; page-number pagination; page parameter `page`; size parameter `per_page`;
  starts at 1; page size 100; fan-out; ids from config field `editorial_image_livefeed_ids`; id
  inserted into the request path; stamps `livefeed_id`; emits passthrough records.
- `get_editorial`: GET `/v2/editorial/{{ fanout.id }}` - single-object response; records at response
  root; fan-out; ids from config field `editorial_ids`; id inserted into the request path; stamps
  `editorial_id`; emits passthrough records.
- `get_editorial_livefeeds`: GET `/v2/editorial/livefeeds` - records path `data`; page-number
  pagination; page parameter `page`; size parameter `per_page`; starts at 1; page size 100; emits
  passthrough records.
- `get_editorial_livefeed`: GET `/v2/editorial/livefeeds/{{ fanout.id }}` - single-object response;
  records at response root; fan-out; ids from config field `editorial_livefeed_ids`; id inserted
  into the request path; stamps `livefeed_id`; emits passthrough records.
- `get_editorial_livefeed_items`: GET `/v2/editorial/livefeeds/{{ fanout.id }}/items` - records path
  `data`; page-number pagination; page parameter `page`; size parameter `per_page`; starts at 1;
  page size 100; fan-out; ids from config field `editorial_livefeed_ids`; id inserted into the
  request path; stamps `livefeed_id`; emits passthrough records.
- `search_editorial`: GET `/v2/editorial/search` - records path `data`; query `category` from
  template `{{ config.category }}`, omitted when absent; `orientation` from template `{{
  config.orientation }}`, omitted when absent; `query` from template `{{ config.query }}`, omitted
  when absent; `sort` from template `{{ config.sort }}`, omitted when absent; page-number
  pagination; page parameter `page`; size parameter `per_page`; starts at 1; page size 100; emits
  passthrough records.
- `get_updated_editorial`: GET `/v2/editorial/updated` - records path `data`; page-number
  pagination; page parameter `page`; size parameter `per_page`; starts at 1; page size 100; emits
  passthrough records.
- `search_editorial_videos`: GET `/v2/editorial/videos/search` - records path `data`; query
  `category` from template `{{ config.category }}`, omitted when absent; `orientation` from template
  `{{ config.orientation }}`, omitted when absent; `query` from template `{{ config.query }}`,
  omitted when absent; `sort` from template `{{ config.sort }}`, omitted when absent; page-number
  pagination; page parameter `page`; size parameter `per_page`; starts at 1; page size 100; emits
  passthrough records.
- `get_editorial_video`: GET `/v2/editorial/videos/{{ fanout.id }}` - single-object response;
  records at response root; fan-out; ids from config field `editorial_video_ids`; id inserted into
  the request path; stamps `editorial_video_id`; emits passthrough records.
- `get_editorial_video_list`: GET `/v2/editorial/videos` - records path `data`; fan-out; ids from
  config field `editorial_video_ids`; id sent as query parameter `id`; stamps `editorial_video_id`;
  emits passthrough records.
- `get_editorial_video_license_list`: GET `/v2/editorial/videos/licenses` - records path `data`;
  page-number pagination; page parameter `page`; size parameter `per_page`; starts at 1; page size
  100; emits passthrough records.
- `get_similar_cv_images`: GET `/v2/cv/similar/images` - records path `data`; page-number
  pagination; page parameter `page`; size parameter `per_page`; starts at 1; page size 100; fan-out;
  ids from config field `visual_asset_ids`; id sent as query parameter `asset_id`; stamps
  `asset_id`; emits passthrough records.
- `get_similar_cv_videos`: GET `/v2/cv/similar/videos` - records path `data`; page-number
  pagination; page parameter `page`; size parameter `per_page`; starts at 1; page size 100; fan-out;
  ids from config field `visual_asset_ids`; id sent as query parameter `asset_id`; stamps
  `asset_id`; emits passthrough records.
- `search_catalog`: GET `/v2/catalog/search` - records path `data`; query `category` from template
  `{{ config.category }}`, omitted when absent; `orientation` from template `{{ config.orientation
  }}`, omitted when absent; `query` from template `{{ config.query }}`, omitted when absent; `sort`
  from template `{{ config.sort }}`, omitted when absent; page-number pagination; page parameter
  `page`; size parameter `per_page`; starts at 1; page size 100; emits passthrough records.
- `get_collections`: GET `/v2/catalog/collections` - records path `data`; page-number pagination;
  page parameter `page`; size parameter `per_page`; starts at 1; page size 100; emits passthrough
  records.
- `get_contributors`: GET `/v2/contributors` - records path `data`; page-number pagination; page
  parameter `page`; size parameter `per_page`; starts at 1; page size 100; emits passthrough
  records.
- `get_contributor`: GET `/v2/contributors/{{ fanout.id }}` - single-object response; records at
  response root; fan-out; ids from config field `contributor_ids`; id inserted into the request
  path; stamps `contributor_id`; emits passthrough records.
- `get_contributor_collections`: GET `/v2/contributors/{{ fanout.id }}/collections` - records path
  `data`; page-number pagination; page parameter `page`; size parameter `per_page`; starts at 1;
  page size 100; fan-out; ids from config field `contributor_ids`; id inserted into the request
  path; stamps `contributor_id`; emits passthrough records.
- `get_contributor_collection`: GET `/v2/contributors/{{ config.contributor_id }}/collections/{{
  fanout.id }}` - single-object response; records at response root; fan-out; ids from config field
  `contributor_collection_ids`; id inserted into the request path; stamps `collection_id`; emits
  passthrough records.
- `get_contributor_collection_items`: GET `/v2/contributors/{{ config.contributor_id
  }}/collections/{{ fanout.id }}/items` - records path `data`; page-number pagination; page
  parameter `page`; size parameter `per_page`; starts at 1; page size 100; fan-out; ids from config
  field `contributor_collection_ids`; id inserted into the request path; stamps `collection_id`;
  emits passthrough records.
- `get_user_subscriptions`: GET `/v2/user/subscriptions` - records path `data`; emits passthrough
  records.

## Write actions & risks

Overall write risk: external Shutterstock collection/lightbox create, rename, delete, and item-add
mutations; licensing/download writes are intentionally excluded.

Reverse ETL writes should be planned, previewed, approved, and then executed. Declared actions:

- `create_image_collection`: POST `/v2/images/collections` - kind `create`; body type `json`;
  required record fields `name`; accepted fields `name`; risk: external Shutterstock POST
  /v2/images/collections; approval required.
- `rename_image_collection`: POST `/v2/images/collections/{{ record.image_collection_id }}` - kind
  `update`; body type `json`; path fields `image_collection_id`; required record fields
  `image_collection_id`, `name`; accepted fields `image_collection_id`, `name`; risk: external
  Shutterstock POST /v2/images/collections/{{ record.image_collection_id }}; approval required.
- `delete_image_collection`: DELETE `/v2/images/collections/{{ record.image_collection_id }}` - kind
  `delete`; body type `none`; path fields `image_collection_id`; required record fields
  `image_collection_id`; accepted fields `image_collection_id`; missing records treated as success
  for status `404`; confirmation `destructive`; risk: destructive external Shutterstock DELETE
  /v2/images/collections/{{ record.image_collection_id }}; approval required.
- `add_image_collection_items`: POST `/v2/images/collections/{{ record.image_collection_id }}/items`
  - kind `update`; body type `json`; path fields `image_collection_id`; required record fields
  `image_collection_id`, `items`; accepted fields `image_collection_id`, `items`; risk: external
  Shutterstock POST /v2/images/collections/{{ record.image_collection_id }}/items; approval
  required.
- `create_video_collection`: POST `/v2/videos/collections` - kind `create`; body type `json`;
  required record fields `name`; accepted fields `name`; risk: external Shutterstock POST
  /v2/videos/collections; approval required.
- `rename_video_collection`: POST `/v2/videos/collections/{{ record.video_collection_id }}` - kind
  `update`; body type `json`; path fields `video_collection_id`; required record fields
  `video_collection_id`, `name`; accepted fields `name`, `video_collection_id`; risk: external
  Shutterstock POST /v2/videos/collections/{{ record.video_collection_id }}; approval required.
- `delete_video_collection`: DELETE `/v2/videos/collections/{{ record.video_collection_id }}` - kind
  `delete`; body type `none`; path fields `video_collection_id`; required record fields
  `video_collection_id`; accepted fields `video_collection_id`; missing records treated as success
  for status `404`; confirmation `destructive`; risk: destructive external Shutterstock DELETE
  /v2/videos/collections/{{ record.video_collection_id }}; approval required.
- `add_video_collection_items`: POST `/v2/videos/collections/{{ record.video_collection_id }}/items`
  - kind `update`; body type `json`; path fields `video_collection_id`; required record fields
  `video_collection_id`, `items`; accepted fields `items`, `video_collection_id`; risk: external
  Shutterstock POST /v2/videos/collections/{{ record.video_collection_id }}/items; approval
  required.
- `create_audio_collection`: POST `/v2/audio/collections` - kind `create`; body type `json`;
  required record fields `name`; accepted fields `name`; risk: external Shutterstock POST
  /v2/audio/collections; approval required.
- `rename_audio_collection`: POST `/v2/audio/collections/{{ record.audio_collection_id }}` - kind
  `update`; body type `json`; path fields `audio_collection_id`; required record fields
  `audio_collection_id`, `name`; accepted fields `audio_collection_id`, `name`; risk: external
  Shutterstock POST /v2/audio/collections/{{ record.audio_collection_id }}; approval required.
- `delete_audio_collection`: DELETE `/v2/audio/collections/{{ record.audio_collection_id }}` - kind
  `delete`; body type `none`; path fields `audio_collection_id`; required record fields
  `audio_collection_id`; accepted fields `audio_collection_id`; missing records treated as success
  for status `404`; confirmation `destructive`; risk: destructive external Shutterstock DELETE
  /v2/audio/collections/{{ record.audio_collection_id }}; approval required.
- `add_audio_collection_items`: POST `/v2/audio/collections/{{ record.audio_collection_id }}/items`
  - kind `update`; body type `json`; path fields `audio_collection_id`; required record fields
  `audio_collection_id`, `items`; accepted fields `audio_collection_id`, `items`; risk: external
  Shutterstock POST /v2/audio/collections/{{ record.audio_collection_id }}/items; approval required.
- `create_catalog_collection`: POST `/v2/catalog/collections` - kind `create`; body type `json`;
  required record fields `name`; accepted fields `metadata`, `name`, `visibility`; risk: external
  Shutterstock POST /v2/catalog/collections; approval required.
- `update_catalog_collection`: PATCH `/v2/catalog/collections/{{ record.catalog_collection_id }}` -
  kind `update`; body type `json`; path fields `catalog_collection_id`; required record fields
  `catalog_collection_id`; accepted fields `catalog_collection_id`, `metadata`, `name`,
  `visibility`; risk: external Shutterstock PATCH /v2/catalog/collections/{{
  record.catalog_collection_id }}; approval required.
- `delete_catalog_collection`: DELETE `/v2/catalog/collections/{{ record.catalog_collection_id }}` -
  kind `delete`; body type `none`; path fields `catalog_collection_id`; required record fields
  `catalog_collection_id`; accepted fields `catalog_collection_id`; missing records treated as
  success for status `404`; confirmation `destructive`; risk: destructive external Shutterstock
  DELETE /v2/catalog/collections/{{ record.catalog_collection_id }}; approval required.
- `add_catalog_collection_items`: POST `/v2/catalog/collections/{{ record.catalog_collection_id
  }}/items` - kind `update`; body type `json`; path fields `catalog_collection_id`; required record
  fields `catalog_collection_id`, `items`; accepted fields `catalog_collection_id`, `items`; risk:
  external Shutterstock POST /v2/catalog/collections/{{ record.catalog_collection_id }}/items;
  approval required.

## Known limits

- Batch defaults: read_page_size=100.
- API coverage includes 58 stream-backed endpoint group(s), 16 write-backed endpoint group(s).
- Other documented endpoints are not exposed by this connector where they are classified as
  binary_payload=5, non_data_endpoint=19, out_of_scope=4, requires_elevated_scope=7.
