# pm connectors inspect google-webfonts

```text
NAME
  pm connectors inspect google-webfonts - Google Webfonts connector manual

SYNOPSIS
  pm connectors inspect google-webfonts
  pm connectors inspect google-webfonts --json
  pm credentials add <name> --connector google-webfonts [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Reads Google Web Fonts families (default, popular, trending, newest, and alphabetical views) through the Google Fonts Developer API. Read-only.

ICON
  id: googleworkpace
  asset: icons/googleworkpace.svg
  source: upstream_registry
  review_status: upstream_seeded

CAPABILITIES
  check=true catalog=true read=true write=false query=false
  Integration type: api

AUTHENTICATION
  Use pm credentials add with --from-env or --value-stdin for secret fields.

CONFIGURATION
  alt
  base_url
  capability
  category
  family
  pretty_print
  subset
  api_key (secret)

ETL STREAMS
  webfonts:
    primary key: family
    cursor: lastModified
    fields: axes(array), category(string), family(string), files(object), kind(string), lastModified(string), menu(string), subset_count(integer), subsets(array), variant_count(integer), variants(array), version(string)
  popular_fonts:
    primary key: family
    cursor: lastModified
    fields: axes(array), category(string), family(string), files(object), kind(string), lastModified(string), menu(string), subset_count(integer), subsets(array), variant_count(integer), variants(array), version(string)
  trending_fonts:
    primary key: family
    cursor: lastModified
    fields: axes(array), category(string), family(string), files(object), kind(string), lastModified(string), menu(string), subset_count(integer), subsets(array), variant_count(integer), variants(array), version(string)
  newest_fonts:
    primary key: family
    cursor: lastModified
    fields: axes(array), category(string), family(string), files(object), kind(string), lastModified(string), menu(string), subset_count(integer), subsets(array), variant_count(integer), variants(array), version(string)
  alpha_fonts:
    primary key: family
    cursor: lastModified
    fields: axes(array), category(string), family(string), files(object), kind(string), lastModified(string), menu(string), subset_count(integer), subsets(array), variant_count(integer), variants(array), version(string)

SYNC MODES
  ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped

SECURITY
  read risk: external Google Fonts Developer API read of public font metadata
  approval: none; read-only public font catalog API
  Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

COMMAND SURFACE
  Run Google Webfonts's declared streams and reverse-ETL actions.
  Usage: pm google-webfonts <command> [flags]
  Read streams
  Other Commands
    alpha fonts list - Run the alpha fonts ETL stream [intent=etl availability=implemented stream=alpha_fonts]
    newest fonts list - Run the newest fonts ETL stream [intent=etl availability=implemented stream=newest_fonts]
    popular fonts list - Run the popular fonts ETL stream [intent=etl availability=implemented stream=popular_fonts]
    trending fonts list - Run the trending fonts ETL stream [intent=etl availability=implemented stream=trending_fonts]
    webfonts list - Run the webfonts ETL stream [intent=etl availability=implemented stream=webfonts]

EXAMPLES
  # Inspect as a manual
  pm connectors inspect google-webfonts

  # Inspect as structured JSON
  pm connectors inspect google-webfonts --json

AGENT WORKFLOW
  - Run pm connectors inspect google-webfonts before creating credentials or plans.
  - Use --json only when the caller needs structured output; use the manual for human-readable guidance.
  - Never ask the user to paste secret values into chat.

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
