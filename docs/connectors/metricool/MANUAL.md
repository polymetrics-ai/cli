# pm connectors inspect metricool

```text
NAME
  pm connectors inspect metricool - Metricool connector manual

SYNOPSIS
  pm connectors inspect metricool
  pm connectors inspect metricool --json
  pm credentials add <name> --connector metricool [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Reads Metricool brand profiles and per-brand Instagram, Facebook, LinkedIn, and TikTok post analytics through the Metricool REST API.

ICON
  asset: icons/pm-sample.svg
  source: polymetrics
  review_status: polymetrics

CAPABILITIES
  check=true catalog=true read=true write=false query=false
  Integration type: api

AUTHENTICATION
  Use pm credentials add with --from-env or --value-stdin for secret fields.

CONFIGURATION
  base_url
  blog_ids
  end_date
  start_date
  user_id
  user_token (secret)

ETL STREAMS
  brands:
    primary key: id
    fields: id(), label(), timezone(), title(), url(), userId()
  instagram_posts:
    primary key: blogId, postId
    fields: blogId(), comments(), impressions(), interactions(), likes(), postId(), publishDate(), reach(), saved(), text(), type(), url()
  facebook_posts:
    primary key: blogId, postId
    fields: blogId(), comments(), impressions(), interactions(), likes(), postId(), publishDate(), reach(), shares(), text(), type(), url()
  linkedin_posts:
    primary key: blogId, postId
    fields: blogId(), clicks(), comments(), impressions(), interactions(), likes(), postId(), publishDate(), shares(), text(), type(), url()
  tiktok_posts:
    primary key: blogId, videoId
    fields: blogId(), comments(), engagement(), likes(), publishDate(), reach(), shares(), text(), url(), videoId(), views()

SYNC MODES
  ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped

SECURITY
  read risk: external Metricool API read of brand-scoped social analytics for the configured user_id/blog_ids
  approval: none; read-only, no reverse-ETL write surface
  Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

EXAMPLES
  # Inspect as a manual
  pm connectors inspect metricool

  # Inspect as structured JSON
  pm connectors inspect metricool --json

AGENT WORKFLOW
  - Run pm connectors inspect metricool before creating credentials or plans.
  - Use --json only when the caller needs structured output; use the manual for human-readable guidance.
  - Never ask the user to paste secret values into chat.

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
