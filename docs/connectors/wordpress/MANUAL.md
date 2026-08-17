# pm connectors inspect wordpress

```text
NAME
  pm connectors inspect wordpress - WordPress connector manual

SYNOPSIS
  pm connectors inspect wordpress
  pm connectors inspect wordpress --json
  pm credentials add <name> --connector wordpress [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Reads and writes WordPress REST API content: posts, pages, comments, media, users, categories, tags, taxonomies, post types, and post statuses.

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
  base_url
  start_date
  password (secret)
  username (secret)

ETL STREAMS
  posts:
    primary key: id
    cursor: date
    fields: _links(), author(), categories(), comment_status(), content(), date(), date_gmt(), excerpt(), featured_media(), format(), guid(), id(), link(), modified(), modified_gmt(), ping_status(), slug(), status(), sticky(), tags(), template(), title(), type()
  pages:
    primary key: id
    cursor: date
    fields: _links(), author(), comment_status(), content(), date(), date_gmt(), excerpt(), featured_media(), guid(), id(), link(), menu_order(), modified(), modified_gmt(), parent(), ping_status(), slug(), status(), template(), title(), type()
  comments:
    primary key: id
    cursor: date
    fields: _links(), author(), author_avatar_urls(), author_name(), author_url(), content(), date(), date_gmt(), id(), link(), parent(), post(), status(), type()
  media:
    primary key: id
    cursor: date
    fields: _links(), author(), comment_status(), date(), date_gmt(), guid(), id(), link(), media_details(), media_type(), mime_type(), modified(), modified_gmt(), ping_status(), post(), slug(), source_url(), status(), title(), type()
  users:
    primary key: id
    fields: _links(), avatar_urls(), description(), id(), link(), name(), slug(), url()
  categories:
    primary key: id
    fields: _links(), count(), description(), id(), link(), name(), parent(), slug(), taxonomy()
  tags:
    primary key: id
    fields: _links(), count(), description(), id(), link(), name(), slug(), taxonomy()
  taxonomies:
    primary key: slug
    fields: description(), hierarchical(), name(), rest_base(), slug(), types()
  types:
    primary key: slug
    fields: description(), has_archive(), hierarchical(), name(), rest_base(), slug(), taxonomies()
  statuses:
    primary key: slug
    fields: date_floating(), name(), public(), queryable(), slug()

SYNC MODES
  ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped, incremental_append, incremental_append_deduped

REVERSE ETL ACTIONS
  create_post:
    endpoint: POST /wp-json/wp/v2/posts
    risk: external mutation; publishes/creates public site content; approval required
  update_post:
    endpoint: POST /wp-json/wp/v2/posts/{{ record.id }}
    required fields: id
    risk: external mutation; edits public site content; approval required
  delete_post:
    endpoint: DELETE /wp-json/wp/v2/posts/{{ record.id }}
    required fields: id
    risk: external deletion of public site content (moves to trash unless force=true is embedded in the path); approval required
  create_page:
    endpoint: POST /wp-json/wp/v2/pages
    risk: external mutation; publishes/creates public site content; approval required
  update_page:
    endpoint: POST /wp-json/wp/v2/pages/{{ record.id }}
    required fields: id
    risk: external mutation; edits public site content; approval required
  delete_page:
    endpoint: DELETE /wp-json/wp/v2/pages/{{ record.id }}
    required fields: id
    risk: external deletion of public site content (moves to trash unless force=true is embedded in the path); approval required
  create_comment:
    endpoint: POST /wp-json/wp/v2/comments
    risk: external mutation; publishes a public-facing comment; approval required
  update_comment:
    endpoint: POST /wp-json/wp/v2/comments/{{ record.id }}
    required fields: id
    risk: external mutation; edits/moderates a public-facing comment; approval required
  delete_comment:
    endpoint: DELETE /wp-json/wp/v2/comments/{{ record.id }}
    required fields: id
    risk: external deletion of a comment (moves to trash unless force=true is embedded in the path); approval required
  update_media:
    endpoint: POST /wp-json/wp/v2/media/{{ record.id }}
    required fields: id
    risk: external mutation; edits media-item metadata (title/alt text/caption/description); approval required
  delete_media:
    endpoint: DELETE /wp-json/wp/v2/media/{{ record.id }}?force=true
    required fields: id
    risk: irreversible external deletion of a media/attachment item (WordPress core requires force=true; attachments do not support trashing); approval required
  create_user:
    endpoint: POST /wp-json/wp/v2/users
    risk: external mutation; creates a new site user account with a password; approval required
  update_user:
    endpoint: POST /wp-json/wp/v2/users/{{ record.id }}
    required fields: id
    risk: external mutation; edits a site user account, including role/permission assignment; approval required
  delete_user:
    endpoint: DELETE /wp-json/wp/v2/users/{{ record.id }}?force=true&reassign={{ record.reassign }}
    required fields: id, reassign
    risk: irreversible external deletion of a site user account (WordPress core requires force=true and a reassign target; users do not support trashing); approval required
  create_category:
    endpoint: POST /wp-json/wp/v2/categories
    risk: external mutation; approval required
  update_category:
    endpoint: POST /wp-json/wp/v2/categories/{{ record.id }}
    required fields: id
    risk: external mutation; approval required
  delete_category:
    endpoint: DELETE /wp-json/wp/v2/categories/{{ record.id }}?force=true
    required fields: id
    risk: irreversible external deletion of a category (WordPress core requires force=true; terms do not support trashing); approval required
  create_tag:
    endpoint: POST /wp-json/wp/v2/tags
    risk: external mutation; approval required
  update_tag:
    endpoint: POST /wp-json/wp/v2/tags/{{ record.id }}
    required fields: id
    risk: external mutation; approval required
  delete_tag:
    endpoint: DELETE /wp-json/wp/v2/tags/{{ record.id }}?force=true
    required fields: id
    risk: irreversible external deletion of a tag (WordPress core requires force=true; terms do not support trashing); approval required

SECURITY
  read risk: external WordPress site read of posts, pages, comments, media, users, categories, tags, taxonomies, post types, and post statuses
  write risk: external mutation of public site content and accounts (posts, pages, comments, media metadata, users, categories, tags); requires authenticated (Basic auth) credentials with sufficient WordPress capabilities; deletes are irreversible for users/categories/tags/media (WordPress core requires force=true, no trash) and approval-gated for all actions
  approval: read: none; write: required for every create/update/delete action
  Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

EXAMPLES
  # Inspect as a manual
  pm connectors inspect wordpress

  # Inspect as structured JSON
  pm connectors inspect wordpress --json

AGENT WORKFLOW
  - Run pm connectors inspect wordpress before creating credentials or plans.
  - Use --json only when the caller needs structured output; use the manual for human-readable guidance.
  - Never ask the user to paste secret values into chat.
  - For reverse ETL writes, create a plan, show the preview, wait for explicit approval, then run with the approval token.

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
