---
name: pm-metricool
description: Metricool connector knowledge and safe action guide.
---

# pm-metricool

## Purpose

Reads Metricool brand profiles and per-brand Instagram, Facebook, LinkedIn, and TikTok post analytics through the Metricool REST API.

## Icon

- id: pm-sample
- asset: icons/pm-sample.svg
- source: polymetrics
- review_status: polymetrics
- review_url: https://github.com/polymetrics-ai/cli

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
- user_id (required)
- user_token (secret)

## ETL Streams

- brands:
  - primary key: id
  - fields: id(integer), label(string), timezone(string), title(string), url(string), userId(integer)
- instagram_posts:
  - primary key: blogId, postId
  - fields: blogId(string), comments(number), impressions(number), interactions(number), likes(number), postId(string), publishDate(string), reach(number), saved(number), text(string), type(string), url(string)
- facebook_posts:
  - primary key: blogId, postId
  - fields: blogId(string), comments(number), impressions(number), interactions(number), likes(number), postId(string), publishDate(string), reach(number), shares(number), text(string), type(string), url(string)
- linkedin_posts:
  - primary key: blogId, postId
  - fields: blogId(string), clicks(number), comments(number), impressions(number), interactions(number), likes(number), postId(string), publishDate(string), shares(number), text(string), type(string), url(string)
- tiktok_posts:
  - primary key: blogId, videoId
  - fields: blogId(string), comments(number), engagement(number), likes(number), publishDate(string), reach(number), shares(number), text(string), url(string), videoId(string), views(number)

## Sync Modes

- ETL sync modes: full_refresh_append, full_refresh_overwrite

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
