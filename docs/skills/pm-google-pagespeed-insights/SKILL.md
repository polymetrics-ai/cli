---
name: pm-google-pagespeed-insights
description: Google PageSpeed Insights connector knowledge and safe action guide.
---

# pm-google-pagespeed-insights

## Purpose

Reads Lighthouse PageSpeed Insights reports (performance, accessibility, best-practices, SEO, PWA scores) for the configured URLs and strategies via the PageSpeed Insights v5 API. In architecture v2 this quarantine bundle dispatches live reads through a Tier-2 hook that delegates to the legacy connector until the wave 6 cutover.

## Icon

- id: google-pagespeed-insights
- asset: icons/google-pagespeed-insights.svg
- source: upstream_registry
- review_status: upstream_seeded
- review_url: https://developers.google.com/speed/docs/insights/v5/get-started

## Capabilities

- check=true catalog=true read=true write=false query=false
- Integration type: api

## Authentication

- Use pm credentials add with --from-env or --value-stdin for secret fields.

## Configuration

- base_url
- categories (required)
- mode
- strategies (required)
- urls (required)
- api_key (secret)

## ETL Streams

- pagespeed_reports:
  - primary key: url, strategy
  - fields: accessibility_score(number), analysis_utc_timestamp(string), best_practices_score(number), fetch_time(string), final_url(string), id(string), kind(string), lighthouse_version(string), overall_loading_experience(string), performance_score(number), pwa_score(number), requested_url(string), seo_score(number), strategy(string), url(string)

## Sync Modes

- ETL sync modes: full_refresh_append, full_refresh_overwrite

## Security

- read risk: external Google PageSpeed Insights API reads performed by the legacy connector via a Tier-2 hook
- write risk: unsupported
- approval: none; read-only
- Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

## Commands

### Inspect as a manual

```bash
pm connectors inspect google-pagespeed-insights
```

### Inspect as structured JSON

```bash
pm connectors inspect google-pagespeed-insights --json
```

## Agent Rules

- Run pm connectors inspect google-pagespeed-insights before creating credentials or plans.
- Use --json only when the caller needs structured output; use the manual for human-readable guidance.
- Never ask the user to paste secret values into chat.
