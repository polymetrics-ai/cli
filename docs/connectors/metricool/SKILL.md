---
name: pm-metricool
description: Metricool connector knowledge and safe action guide.
---

# pm-metricool

## Purpose

Reads Metricool brand profiles and per-brand Instagram, Facebook, LinkedIn, and TikTok post analytics through the Metricool REST API.

## Icon

- asset: icons/pm-sample.svg
- source: polymetrics
- review_status: polymetrics

## Capabilities

- check=true catalog=true read=true write=false query=false
- Integration type: api

## Authentication

- Use pm credentials add with --from-env or --value-stdin for secret fields.

## Configuration

- base_url
- blog_ids
- end_date
- start_date
- user_id
- user_token (secret)

## ETL Streams

- brands:
  - primary key: id
  - fields: id(), label(), timezone(), title(), url(), userId()
- instagram_posts:
  - primary key: blogId, postId
  - fields: blogId(), comments(), impressions(), interactions(), likes(), postId(), publishDate(), reach(), saved(), text(), type(), url()
- facebook_posts:
  - primary key: blogId, postId
  - fields: blogId(), comments(), impressions(), interactions(), likes(), postId(), publishDate(), reach(), shares(), text(), type(), url()
- linkedin_posts:
  - primary key: blogId, postId
  - fields: blogId(), clicks(), comments(), impressions(), interactions(), likes(), postId(), publishDate(), shares(), text(), type(), url()
- tiktok_posts:
  - primary key: blogId, videoId
  - fields: blogId(), comments(), engagement(), likes(), publishDate(), reach(), shares(), text(), url(), videoId(), views()

## Sync Modes

- ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped

## Security

- read risk: external Metricool API read of brand-scoped social analytics for the configured user_id/blog_ids
- approval: none; read-only, no reverse-ETL write surface
- Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

## Commands

### Inspect as a manual

```bash
pm connectors inspect metricool
```

### Inspect as structured JSON

```bash
pm connectors inspect metricool --json
```

## Agent Rules

- Run pm connectors inspect metricool before creating credentials or plans.
- Use --json only when the caller needs structured output; use the manual for human-readable guidance.
- Never ask the user to paste secret values into chat.
