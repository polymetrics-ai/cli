---
name: pm-google-pagespeed-insights
description: Google PageSpeed Insights connector knowledge and safe action guide.
---

# pm-google-pagespeed-insights

## Purpose

Reads Lighthouse PageSpeed Insights reports (performance, accessibility, best-practices, SEO, PWA scores) for the configured URLs and strategies via the PageSpeed Insights v5 API. In architecture v2 this quarantine bundle dispatches live reads through a Tier-2 hook that delegates to the legacy connector until the wave 6 cutover.

## Icon

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
- categories
- mode
- strategies
- urls
- api_key (secret)

## ETL Streams

- pagespeed_reports:
  - primary key: url, strategy
  - fields: accessibility_score(), analysis_utc_timestamp(), best_practices_score(), fetch_time(), final_url(), id(), kind(), lighthouse_version(), overall_loading_experience(), performance_score(), pwa_score(), requested_url(), seo_score(), strategy(), url()

## Sync Modes

- ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped

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
