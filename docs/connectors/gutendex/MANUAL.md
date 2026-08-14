# pm connectors inspect gutendex

```text
NAME
  pm connectors inspect gutendex - Gutendex connector manual

SYNOPSIS
  pm connectors inspect gutendex
  pm connectors inspect gutendex --json
  pm credentials add <name> --connector gutendex [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Reads Project Gutenberg books from the free, public Gutendex JSON API (books, popular, latest, and English-language views). Read-only; no credentials required.

ICON
  id: source-gutendex
  asset: icons/source-gutendex.svg
  source: upstream_registry
  review_status: upstream_seeded

CAPABILITIES
  check=true catalog=true read=true write=false query=false
  Integration type: api

AUTHENTICATION
  No secret authentication is required for this connector.

CONFIGURATION
  author_year_end
  author_year_start
  base_url
  copyright
  ids
  languages
  mode
  search
  sort
  topic

ETL STREAMS
  books:
    primary key: id
    fields: bookshelves(string), copyright(boolean), download_count(integer), id(integer), languages(string), media_type(string), subjects(string), title(string)
  popular_books:
    primary key: id
    fields: bookshelves(string), copyright(boolean), download_count(integer), id(integer), languages(string), media_type(string), subjects(string), title(string)
  latest_books:
    primary key: id
    fields: bookshelves(string), copyright(boolean), download_count(integer), id(integer), languages(string), media_type(string), subjects(string), title(string)
  english_books:
    primary key: id
    fields: bookshelves(string), copyright(boolean), download_count(integer), id(integer), languages(string), media_type(string), subjects(string), title(string)

SYNC MODES
  ETL sync modes: full_refresh_append, full_refresh_overwrite

SECURITY
  read risk: external read of the public, unauthenticated Gutendex book catalog
  approval: none; read-only public API, no credentials
  Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

EXAMPLES
  # Inspect as a manual
  pm connectors inspect gutendex

  # Inspect as structured JSON
  pm connectors inspect gutendex --json

AGENT WORKFLOW
  - Run pm connectors inspect gutendex before creating credentials or plans.
  - Use --json only when the caller needs structured output; use the manual for human-readable guidance.
  - Never ask the user to paste secret values into chat.

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
