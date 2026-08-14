# pm connectors inspect algolia

```text
NAME
  pm connectors inspect algolia - Algolia connector manual

SYNOPSIS
  pm connectors inspect algolia
  pm connectors inspect algolia --json
  pm credentials add <name> --connector algolia [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Reads Algolia indices, API keys, index settings, dictionaries, security sources, and logs, and writes index settings/API keys, through the Algolia Search REST API.

ICON
  id: simple-icons-algolia
  asset: icons/simple-icons/algolia.svg
  title: Algolia
  simple_icon_slug: algolia
  simple_icon_hex: 003DFF
  source: simple-icons
  license: CC0-1.0
  review_status: cc0_with_trademark_caveat
  review_url: https://simpleicons.org/?q=Algolia
  match: exact-name-or-slug
  matched_by: algolia

CAPABILITIES
  check=true catalog=true read=true write=true query=false
  Integration type: api

AUTHENTICATION
  Use pm credentials add with --from-env or --value-stdin for secret fields.

CONFIGURATION
  application_id (required)
  base_url (required)
  index_name
  api_key (secret) (required)

ETL STREAMS
  indices:
    primary key: name
    fields: created_at(string), data_size(integer), entries(integer), file_size(integer), last_build_time_s(integer), name(string), number_of_pending_tasks(integer), pending_task(boolean), primary(string), replicas(array), updated_at(string)
  api_keys:
    primary key: value
    fields: acl(array), created_at(integer), description(string), indexes(array), max_hits_per_query(integer), max_queries_per_ip_per_hour(integer), referers(array), validity(integer), value(string)
  index_settings:
    primary key: index_name
    fields: attributes_for_faceting(array), custom_ranking(array), hits_per_page(integer), index_name(string), pagination_limited_to(integer), ranking(array), replicas(array), searchable_attributes(array)
  vault_sources:
    primary key: source
    fields: description(string), source(string)
  dictionary_settings:
    primary key: id
    fields: disable_standard_entries(object), id(string)
  dictionary_languages:
    primary key: language
    fields: compounds(object), language(string), plurals(object), stopwords(object)
  logs:
    primary key: id
    fields: answer(string), answer_code(string), id(string), index(string), ip(string), method(string), nb_api_calls(string), processing_time_ms(string), query_body(string), query_headers(string), query_nb_hits(string), query_params(string), sha1(string), timestamp(string), url(string)

SYNC MODES
  ETL sync modes: full_refresh_append, full_refresh_overwrite

REVERSE ETL ACTIONS
  update_index_settings:
    endpoint: PUT /1/indexes/{{ record.index_name }}/settings
    required fields: index_name
    risk: overwrites the named index's search settings (ranking, faceting, searchable attributes); settings not included in the submitted record are left unchanged, but any included field replaces its current value immediately for live search traffic
  create_api_key:
    endpoint: POST /1/keys
    required fields: acl
    risk: creates a new live Algolia API key with the requested ACL/index scope; a broadly-scoped key (e.g. admin-level ACLs) is a new standing credential that must be tracked and rotated like any other secret

SECURITY
  read risk: external Algolia API read of index/key/dictionary/security/log configuration metadata
  write risk: external mutation: overwrites live index search settings (update_index_settings) or creates a new standing API key credential (create_api_key); approval required
  approval: required for both write actions
  Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

EXAMPLES
  # Inspect as a manual
  pm connectors inspect algolia

  # Inspect as structured JSON
  pm connectors inspect algolia --json

AGENT WORKFLOW
  - Run pm connectors inspect algolia before creating credentials or plans.
  - Use --json only when the caller needs structured output; use the manual for human-readable guidance.
  - Never ask the user to paste secret values into chat.
  - For reverse ETL writes, create a plan, show the preview, wait for explicit approval, then run with the approval token.

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
