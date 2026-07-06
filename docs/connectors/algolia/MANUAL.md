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
  asset: icons/pm-sample.svg
  source: polymetrics
  review_status: polymetrics

CAPABILITIES
  check=true catalog=true read=true write=true query=false
  Integration type: api

AUTHENTICATION
  Use pm credentials add with --from-env or --value-stdin for secret fields.

CONFIGURATION
  application_id
  base_url
  index_name
  api_key (secret)

ETL STREAMS
  indices:
    primary key: name
    fields: created_at(), data_size(), entries(), file_size(), last_build_time_s(), name(), number_of_pending_tasks(), pending_task(), primary(), replicas(), updated_at()
  api_keys:
    primary key: value
    fields: acl(), created_at(), description(), indexes(), max_hits_per_query(), max_queries_per_ip_per_hour(), referers(), validity(), value()
  index_settings:
    primary key: index_name
    fields: attributes_for_faceting(), custom_ranking(), hits_per_page(), index_name(), pagination_limited_to(), ranking(), replicas(), searchable_attributes()
  vault_sources:
    primary key: source
    fields: description(), source()
  dictionary_settings:
    primary key: id
    fields: disable_standard_entries(), id()
  dictionary_languages:
    primary key: language
    fields: compounds(), language(), plurals(), stopwords()
  logs:
    primary key: id
    fields: answer(), answer_code(), id(), index(), ip(), method(), nb_api_calls(), processing_time_ms(), query_body(), query_headers(), query_nb_hits(), query_params(), sha1(), timestamp(), url()

SYNC MODES
  ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped

REVERSE ETL ACTIONS
  update_index_settings:
    endpoint: PUT /1/indexes/{{ record.index_name }}/settings
    required fields: index_name
    risk: overwrites the named index's search settings (ranking, faceting, searchable attributes); settings not included in the submitted record are left unchanged, but any included field replaces its current value immediately for live search traffic
  create_api_key:
    endpoint: POST /1/keys
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
