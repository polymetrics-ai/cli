# pm connectors inspect confluence

```text
NAME
  pm connectors inspect confluence - Confluence connector manual

SYNOPSIS
  pm connectors inspect confluence
  pm connectors inspect confluence --json
  pm credentials add <name> --connector confluence [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Reads Confluence Cloud spaces, pages, blog posts, labels, attachments, comments, tasks, and custom content, and writes pages, blog posts, and comments through the Confluence Cloud REST API v2.

ICON
  asset: icons/confluence.svg
  source: upstream_registry
  review_status: upstream_seeded
  review_url: https://developer.atlassian.com/cloud/confluence/rest/v2/intro/

CAPABILITIES
  check=true catalog=true read=true write=true query=false
  Integration type: api

AUTHENTICATION
  Use pm credentials add with --from-env or --value-stdin for secret fields.

CONFIGURATION
  base_url
  custom_content_type
  email
  mode
  api_token (secret)

ETL STREAMS
  spaces:
    primary key: id
    fields: authorId(), createdAt(), homepageId(), id(), key(), name(), status(), type()
  pages:
    primary key: id
    cursor: createdAt
    fields: authorId(), createdAt(), id(), parentId(), spaceId(), status(), title(), version()
  blogposts:
    primary key: id
    cursor: createdAt
    fields: authorId(), createdAt(), id(), spaceId(), status(), title(), version()
  labels:
    primary key: id
    fields: id(), name(), prefix()
  attachments:
    primary key: id
    cursor: createdAt
    fields: createdAt(), fileSize(), id(), mediaType(), pageId(), status(), title()
  footer_comments:
    primary key: id
    fields: blogPostId(), id(), pageId(), parentCommentId(), status(), title(), version()
  inline_comments:
    primary key: id
    fields: blogPostId(), id(), pageId(), parentCommentId(), resolutionStatus(), status(), title(), version()
  tasks:
    primary key: id
    fields: assignedTo(), blogPostId(), completedAt(), completedBy(), createdAt(), createdBy(), dueAt(), id(), localId(), pageId(), spaceId(), status(), updatedAt()
  custom_content:
    primary key: id
    fields: authorId(), blogPostId(), createdAt(), id(), pageId(), spaceId(), status(), title(), type(), version()

SYNC MODES
  ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped

REVERSE ETL ACTIONS
  create_page:
    endpoint: POST /wiki/api/v2/pages
    risk: creates a new published or draft page in the target space; external mutation, no approval required
  update_page:
    endpoint: PUT /wiki/api/v2/pages/{{ record.id }}
    required fields: id
    risk: mutates an existing page's content/status; requires the caller to supply the next version.number (Confluence rejects a stale version number), external mutation, no approval required
  create_blogpost:
    endpoint: POST /wiki/api/v2/blogposts
    risk: creates a new published or draft blog post in the target space; external mutation, no approval required
  create_footer_comment:
    endpoint: POST /wiki/api/v2/footer-comments
    risk: creates a new footer comment (or reply) on a page/blogpost; external mutation, no approval required
  create_inline_comment:
    endpoint: POST /wiki/api/v2/inline-comments
    risk: creates a new inline comment (or reply) anchored to a text selection on a page/blogpost; external mutation, no approval required

SECURITY
  read risk: external Confluence Cloud API read of space/content metadata
  write risk: external mutation: creates/updates Confluence pages, blog posts, and comments; no destructive (delete) actions are exposed
  approval: required for all write actions; read-only otherwise
  Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

EXAMPLES
  # Inspect as a manual
  pm connectors inspect confluence

  # Inspect as structured JSON
  pm connectors inspect confluence --json

AGENT WORKFLOW
  - Run pm connectors inspect confluence before creating credentials or plans.
  - Use --json only when the caller needs structured output; use the manual for human-readable guidance.
  - Never ask the user to paste secret values into chat.
  - For reverse ETL writes, create a plan, show the preview, wait for explicit approval, then run with the approval token.

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
