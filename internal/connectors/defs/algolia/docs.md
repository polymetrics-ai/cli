# Overview

Reads Algolia indices, API keys, index settings, dictionaries, security sources, and logs, and
writes index settings/API keys, through the Algolia Search REST API.

Readable streams: `indices`, `api_keys`, `index_settings`, `vault_sources`, `dictionary_settings`,
`dictionary_languages`, `logs`.

Write actions: `update_index_settings`, `create_api_key`.

Service API documentation: https://www.algolia.com/doc/rest-api/search/.

## Auth setup

Connection fields:

- `api_key` (required, secret, string); Algolia Admin API key, sent as the X-Algolia-API-Key header.
  Never logged.
- `application_id` (required, string); Algolia application id, sent as the X-Algolia-Application-Id
  header and used to derive base_url when not overridden.
- `base_url` (required, string); format `uri`; Algolia REST API base URL, e.g.
  https://<application_id>.algolia.net.
- `index_name` (optional, string); Algolia index name (required for the 'index_settings' stream).

Secret fields are redacted in logs and write previews: `api_key`.

Authentication behavior:

- API key authentication in `X-Algolia-API-Key` using `secrets.api_key`.

Requests use the configured `base_url` value after applying defaults.

Connection checks call GET `/1/indexes`.

## Streams notes

Default pagination: single request; no pagination.

Pagination by stream: none: `api_keys`, `index_settings`, `vault_sources`, `dictionary_settings`,
`dictionary_languages`; offset_limit: `logs`; page_number: `indices`.

- `indices`: GET `/1/indexes` - records path `items`; page-number pagination; page parameter `page`;
  no page-size parameter; starts at 0; page size 100; maximum 100 page(s); computed output fields
  `created_at`, `data_size`, `file_size`, `last_build_time_s`, `number_of_pending_tasks`,
  `pending_task`, `updated_at`.
- `api_keys`: GET `/1/keys` - records path `keys`; computed output fields `created_at`,
  `max_hits_per_query`, `max_queries_per_ip_per_hour`.
- `index_settings`: GET `/1/indexes/{{ config.index_name }}/settings` - records path `.`; computed
  output fields `attributes_for_faceting`, `custom_ranking`, `hits_per_page`, `index_name`,
  `pagination_limited_to`, `searchable_attributes`.
- `vault_sources`: GET `/1/security/sources` - records path `.`.
- `dictionary_settings`: GET `/1/dictionaries/*/settings` - records path `.`; computed output fields
  `disable_standard_entries`, `id`.
- `dictionary_languages`: GET `/1/dictionaries/*/languages` - records path `.`; flattens keyed
  objects; key field `language`.
- `logs`: GET `/1/logs` - records path `logs`; offset/limit pagination; offset parameter `offset`;
  limit parameter `length`; page size 10; computed output fields `id`.

## Write actions & risks

Overall write risk: external mutation: overwrites live index search settings (update_index_settings)
or creates a new standing API key credential (create_api_key); approval required.

Reverse ETL writes should be planned, previewed, approved, and then executed. Declared actions:

- `update_index_settings`: PUT `/1/indexes/{{ record.index_name }}/settings` - kind `update`; body
  type `json`; path fields `index_name`; required record fields `index_name`; accepted fields
  `attributesForFaceting`, `customRanking`, `hitsPerPage`, `index_name`, `searchableAttributes`;
  risk: overwrites the named index's search settings (ranking, faceting, searchable attributes);
  settings not included in the submitted record are left unchanged, but any included field replaces
  its current value immediately for live search traffic.
- `create_api_key`: POST `/1/keys` - kind `create`; body type `json`; required record fields `acl`;
  accepted fields `acl`, `description`, `indexes`, `maxHitsPerQuery`, `maxQueriesPerIPPerHour`,
  `referers`, `validity`; risk: creates a new live Algolia API key with the requested ACL/index
  scope; a broadly-scoped key (e.g. admin-level ACLs) is a new standing credential that must be
  tracked and rotated like any other secret.

## Known limits

- Batch defaults: read_page_size=100.
- API coverage includes 7 stream-backed endpoint group(s), 2 write-backed endpoint group(s).
- Other documented endpoints are not exposed by this connector where they are classified as
  deprecated=9, destructive_admin=11, non_data_endpoint=1, out_of_scope=25,
  requires_elevated_scope=3.
