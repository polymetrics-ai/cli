---
name: pm-google-pagespeed-insights
description: Google PageSpeed Insights connector knowledge and safe action guide.
---

# pm-google-pagespeed-insights

## Purpose

Reads Lighthouse PageSpeed Insights reports for a bounded Cartesian product of configured HTTPS URLs and mobile or desktop strategies through the fixed PageSpeed Insights v5 API.

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

- read risk: Each sync sends at most twenty bounded GET requests to the fixed PageSpeed Insights API origin; one report is emitted per configured URL and strategy pair.
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
