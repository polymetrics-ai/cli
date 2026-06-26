---
name: pm-google-pagespeed-insights
description: Google PageSpeed Insights connector knowledge and safe action guide.
---

# pm-google-pagespeed-insights

## Purpose

Reads Lighthouse PageSpeed Insights reports (performance, accessibility, best-practices, SEO, PWA scores) for the configured URLs and strategies via the PageSpeed Insights v5 API.

## Capabilities

- check=true catalog=true read=true write=false query=false
- Integration type: api

## Authentication

- No secret authentication is required for this connector.

## Configuration

- No connector-specific config fields.

## Security

- read risk: connector-specific
- write risk: connector-specific
- approval: external mutations require preview and approval
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

