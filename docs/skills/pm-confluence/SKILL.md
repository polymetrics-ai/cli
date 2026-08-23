---
name: pm-confluence
description: Confluence connector knowledge and safe action guide.
---

# pm-confluence

## Purpose

Reads Confluence Cloud spaces, pages, blog posts, labels, attachments, comments, tasks, and custom content, and writes pages, blog posts, and comments through the Confluence Cloud REST API v2.

## Icon

- id: confluence
- asset: icons/confluence.svg
- source: upstream_registry
- review_status: upstream_seeded
- review_url: https://developer.atlassian.com/cloud/confluence/rest/v2/intro/

## Capabilities

- check=true catalog=true read=true write=true query=false
- Integration type: api

## Authentication

- Use pm credentials add with --from-env or --value-stdin for secret fields.

## Configuration

- base_url (required)
- custom_content_type
- email (required)
- mode
- api_token (secret) (required)

## ETL Streams

- spaces:
  - primary key: id
  - fields: authorId(string), createdAt(string), homepageId(string), id(string), key(string), name(string), status(string), type(string)
- pages:
  - primary key: id
  - cursor: createdAt
  - fields: authorId(string), createdAt(string), id(string), parentId(string), spaceId(string), status(string), title(string), version(integer)
- blogposts:
  - primary key: id
  - cursor: createdAt
  - fields: authorId(string), createdAt(string), id(string), spaceId(string), status(string), title(string), version(integer)
- labels:
  - primary key: id
  - fields: id(string), name(string), prefix(string)
- attachments:
  - primary key: id
  - cursor: createdAt
  - fields: createdAt(string), fileSize(integer), id(string), mediaType(string), pageId(string), status(string), title(string)
- footer_comments:
  - primary key: id
  - fields: blogPostId(string), id(string), pageId(string), parentCommentId(string), status(string), title(string), version(integer)
- inline_comments:
  - primary key: id
  - fields: blogPostId(string), id(string), pageId(string), parentCommentId(string), resolutionStatus(string), status(string), title(string), version(integer)
- tasks:
  - primary key: id
  - fields: assignedTo(string), blogPostId(string), completedAt(string), completedBy(string), createdAt(string), createdBy(string), dueAt(string), id(string), localId(string), pageId(string), spaceId(string), status(string), updatedAt(string)
- custom_content:
  - primary key: id
  - fields: authorId(string), blogPostId(string), createdAt(string), id(string), pageId(string), spaceId(string), status(string), title(string), type(string), version(integer)

## Sync Modes

- ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped

## Reverse ETL Actions

- create_page:
  - endpoint: POST /wiki/api/v2/pages
  - required fields: spaceId, title, body
  - risk: creates a new published or draft page in the target space; external mutation, no approval required
- update_page:
  - endpoint: PUT /wiki/api/v2/pages/{{ record.id }}
  - required fields: id, status, title, spaceId, version
  - risk: mutates an existing page's content/status; requires the caller to supply the next version.number (Confluence rejects a stale version number), external mutation, no approval required
- create_blogpost:
  - endpoint: POST /wiki/api/v2/blogposts
  - required fields: spaceId, title, body
  - risk: creates a new published or draft blog post in the target space; external mutation, no approval required
- create_footer_comment:
  - endpoint: POST /wiki/api/v2/footer-comments
  - required fields: pageId, body
  - risk: creates a new footer comment (or reply) on a page/blogpost; external mutation, no approval required
- create_inline_comment:
  - endpoint: POST /wiki/api/v2/inline-comments
  - required fields: pageId, body, inlineCommentProperties
  - risk: creates a new inline comment (or reply) anchored to a text selection on a page/blogpost; external mutation, no approval required

## Security

- read risk: external Confluence Cloud API read of space/content metadata
- write risk: external mutation: creates/updates Confluence pages, blog posts, and comments; no destructive (delete) actions are exposed
- approval: required for all write actions; read-only otherwise
- Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

## Commands

### Inspect as a manual

```bash
pm connectors inspect confluence
```

### Inspect as structured JSON

```bash
pm connectors inspect confluence --json
```

## Agent Rules

- Run pm connectors inspect confluence before creating credentials or plans.
- Use --json only when the caller needs structured output; use the manual for human-readable guidance.
- Never ask the user to paste secret values into chat.
- For reverse ETL writes, create a plan, show the preview, wait for explicit approval, then run with the approval token.
