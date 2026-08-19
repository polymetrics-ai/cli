---
name: pm-wordpress
description: WordPress connector knowledge and safe action guide.
---

# pm-wordpress

## Purpose

Reads and writes WordPress REST API content: posts, pages, comments, media, users, categories, tags, taxonomies, post types, and post statuses.

## Icon

- id: simple-icons-wordpress
- asset: icons/simple-icons/wordpress.svg
- title: WordPress
- simple_icon_slug: wordpress
- simple_icon_hex: 21759B
- source: simple-icons
- license: CC0-1.0
- review_status: cc0_with_trademark_caveat
- review_url: https://simpleicons.org/?q=WordPress
- match: exact-name-or-slug
- matched_by: wordpress

## Capabilities

- check=true catalog=true read=true write=true query=false
- Integration type: api

## Authentication

- Use pm credentials add with --from-env or --value-stdin for secret fields.

## Configuration

- base_url (required)
- start_date
- password (secret)
- username (secret)

## ETL Streams

- posts:
  - primary key: id
  - cursor: date
  - fields: _links(object), author(integer), categories(array), comment_status(string), content(object), date(string), date_gmt(string), excerpt(object), featured_media(integer), format(string), guid(object), id(integer), link(string), modified(string), modified_gmt(string), ping_status(string), slug(string), status(string), sticky(boolean), tags(array), template(string), title(object), type(string)
- pages:
  - primary key: id
  - cursor: date
  - fields: _links(object), author(integer), comment_status(string), content(object), date(string), date_gmt(string), excerpt(object), featured_media(integer), guid(object), id(integer), link(string), menu_order(integer), modified(string), modified_gmt(string), parent(integer), ping_status(string), slug(string), status(string), template(string), title(object), type(string)
- comments:
  - primary key: id
  - cursor: date
  - fields: _links(object), author(integer), author_avatar_urls(object), author_name(string), author_url(string), content(object), date(string), date_gmt(string), id(integer), link(string), parent(integer), post(integer), status(string), type(string)
- media:
  - primary key: id
  - cursor: date
  - fields: _links(object), author(integer), comment_status(string), date(string), date_gmt(string), guid(object), id(integer), link(string), media_details(object), media_type(string), mime_type(string), modified(string), modified_gmt(string), ping_status(string), post(integer), slug(string), source_url(string), status(string), title(object), type(string)
- users:
  - primary key: id
  - fields: _links(object), avatar_urls(object), description(string), id(integer), link(string), name(string), slug(string), url(string)
- categories:
  - primary key: id
  - fields: _links(object), count(integer), description(string), id(integer), link(string), name(string), parent(integer), slug(string), taxonomy(string)
- tags:
  - primary key: id
  - fields: _links(object), count(integer), description(string), id(integer), link(string), name(string), slug(string), taxonomy(string)
- taxonomies:
  - primary key: slug
  - fields: description(string), hierarchical(boolean), name(string), rest_base(string), slug(string), types(array)
- types:
  - primary key: slug
  - fields: description(string), has_archive(boolean), hierarchical(boolean), name(string), rest_base(string), slug(string), taxonomies(array)
- statuses:
  - primary key: slug
  - fields: date_floating(boolean), name(string), public(boolean), queryable(boolean), slug(string)

## Sync Modes

- ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped, incremental_append, incremental_append_deduped

## Reverse ETL Actions

- create_post:
  - endpoint: POST /wp-json/wp/v2/posts
  - risk: external mutation; publishes/creates public site content; approval required
- update_post:
  - endpoint: POST /wp-json/wp/v2/posts/{{ record.id }}
  - required fields: id
  - risk: external mutation; edits public site content; approval required
- delete_post:
  - endpoint: DELETE /wp-json/wp/v2/posts/{{ record.id }}
  - required fields: id
  - risk: external deletion of public site content (moves to trash unless force=true is embedded in the path); approval required
- create_page:
  - endpoint: POST /wp-json/wp/v2/pages
  - risk: external mutation; publishes/creates public site content; approval required
- update_page:
  - endpoint: POST /wp-json/wp/v2/pages/{{ record.id }}
  - required fields: id
  - risk: external mutation; edits public site content; approval required
- delete_page:
  - endpoint: DELETE /wp-json/wp/v2/pages/{{ record.id }}
  - required fields: id
  - risk: external deletion of public site content (moves to trash unless force=true is embedded in the path); approval required
- create_comment:
  - endpoint: POST /wp-json/wp/v2/comments
  - required fields: post, content
  - risk: external mutation; publishes a public-facing comment; approval required
- update_comment:
  - endpoint: POST /wp-json/wp/v2/comments/{{ record.id }}
  - required fields: id
  - risk: external mutation; edits/moderates a public-facing comment; approval required
- delete_comment:
  - endpoint: DELETE /wp-json/wp/v2/comments/{{ record.id }}
  - required fields: id
  - risk: external deletion of a comment (moves to trash unless force=true is embedded in the path); approval required
- update_media:
  - endpoint: POST /wp-json/wp/v2/media/{{ record.id }}
  - required fields: id
  - risk: external mutation; edits media-item metadata (title/alt text/caption/description); approval required
- delete_media:
  - endpoint: DELETE /wp-json/wp/v2/media/{{ record.id }}?force=true
  - required fields: id
  - risk: irreversible external deletion of a media/attachment item (WordPress core requires force=true; attachments do not support trashing); approval required
- create_user:
  - endpoint: POST /wp-json/wp/v2/users
  - required fields: username, email, password
  - risk: external mutation; creates a new site user account with a password; approval required
- update_user:
  - endpoint: POST /wp-json/wp/v2/users/{{ record.id }}
  - required fields: id
  - risk: external mutation; edits a site user account, including role/permission assignment; approval required
- delete_user:
  - endpoint: DELETE /wp-json/wp/v2/users/{{ record.id }}?force=true&reassign={{ record.reassign }}
  - required fields: id, reassign
  - risk: irreversible external deletion of a site user account (WordPress core requires force=true and a reassign target; users do not support trashing); approval required
- create_category:
  - endpoint: POST /wp-json/wp/v2/categories
  - required fields: name
  - risk: external mutation; approval required
- update_category:
  - endpoint: POST /wp-json/wp/v2/categories/{{ record.id }}
  - required fields: id
  - risk: external mutation; approval required
- delete_category:
  - endpoint: DELETE /wp-json/wp/v2/categories/{{ record.id }}?force=true
  - required fields: id
  - risk: irreversible external deletion of a category (WordPress core requires force=true; terms do not support trashing); approval required
- create_tag:
  - endpoint: POST /wp-json/wp/v2/tags
  - required fields: name
  - risk: external mutation; approval required
- update_tag:
  - endpoint: POST /wp-json/wp/v2/tags/{{ record.id }}
  - required fields: id
  - risk: external mutation; approval required
- delete_tag:
  - endpoint: DELETE /wp-json/wp/v2/tags/{{ record.id }}?force=true
  - required fields: id
  - risk: irreversible external deletion of a tag (WordPress core requires force=true; terms do not support trashing); approval required

## Security

- read risk: external WordPress site read of posts, pages, comments, media, users, categories, tags, taxonomies, post types, and post statuses
- write risk: external mutation of public site content and accounts (posts, pages, comments, media metadata, users, categories, tags); requires authenticated (Basic auth) credentials with sufficient WordPress capabilities; deletes are irreversible for users/categories/tags/media (WordPress core requires force=true, no trash) and approval-gated for all actions
- approval: read: none; write: required for every create/update/delete action
- Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

## Commands

### Inspect as a manual

```bash
pm connectors inspect wordpress
```

### Inspect as structured JSON

```bash
pm connectors inspect wordpress --json
```

## Agent Rules

- Run pm connectors inspect wordpress before creating credentials or plans.
- Use --json only when the caller needs structured output; use the manual for human-readable guidance.
- Never ask the user to paste secret values into chat.
- For reverse ETL writes, create a plan, show the preview, wait for explicit approval, then run with the approval token.
