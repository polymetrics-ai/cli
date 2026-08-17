---
name: pm-google-search-console
description: Google Search Console connector knowledge and safe action guide.
---

# pm-google-search-console

## Purpose

Reads Google Search Console sites, sitemaps, and Search Analytics performance reports (by date, query, page, country, and device) through the Search Console v3 REST API; submits/removes sites and sitemaps through explicit write actions.

## Icon

- asset: icons/googlesearchconsole.svg
- source: upstream_registry
- review_status: upstream_seeded
- review_url: https://developers.google.com/search/news

## Capabilities

- check=true catalog=true read=true write=true query=false
- Integration type: api

## Authentication

- Use pm credentials add with --from-env or --value-stdin for secret fields.

## Configuration

- base_url
- data_state
- end_date
- feedpath
- max_pages
- mode
- page_size
- search_type
- site_url
- site_urls
- start_date
- access_token (secret)

## ETL Streams

- sites:
  - primary key: site_url
  - fields: permission_level(), site_url()
- site_details:
  - primary key: site_url
  - fields: permission_level(), site_url()
- sitemaps:
  - primary key: site_url, path
  - fields: errors(), is_pending(), is_sitemaps_index(), last_downloaded(), last_submitted(), path(), site_url(), type(), warnings()
- sitemap_details:
  - primary key: site_url, path
  - fields: errors(), is_pending(), is_sitemaps_index(), last_downloaded(), last_submitted(), path(), site_url(), type(), warnings()
- search_analytics_by_date:
  - primary key: site_url, search_type, date
  - cursor: date
  - fields: clicks(), ctr(), date(), impressions(), position(), search_type(), site_url()
- search_analytics_by_country:
  - primary key: site_url, search_type, date, country
  - cursor: date
  - fields: clicks(), country(), ctr(), date(), impressions(), position(), search_type(), site_url()
- search_analytics_by_device:
  - primary key: site_url, search_type, date, device
  - cursor: date
  - fields: clicks(), ctr(), date(), device(), impressions(), position(), search_type(), site_url()
- search_analytics_by_page:
  - primary key: site_url, search_type, date, page
  - cursor: date
  - fields: clicks(), ctr(), date(), impressions(), page(), position(), search_type(), site_url()
- search_analytics_by_query:
  - primary key: site_url, search_type, date, query
  - cursor: date
  - fields: clicks(), ctr(), date(), impressions(), position(), query(), search_type(), site_url()

## Sync Modes

- ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped

## Reverse ETL Actions

- add_site:
  - endpoint: PUT /sites/{{ record.site_url | urlencode }}
  - required fields: site_url
  - risk: adds a site property to the authenticated Search Console account
- delete_site:
  - endpoint: DELETE /sites/{{ record.site_url | urlencode }}
  - required fields: site_url
  - risk: removes a site property from the authenticated Search Console account
- submit_sitemap:
  - endpoint: PUT /sites/{{ record.site_url | urlencode }}/sitemaps/{{ record.feedpath | urlencode }}
  - required fields: site_url, feedpath
  - risk: submits a sitemap URL for a Search Console site property
- delete_sitemap:
  - endpoint: DELETE /sites/{{ record.site_url | urlencode }}/sitemaps/{{ record.feedpath | urlencode }}
  - required fields: site_url, feedpath
  - risk: deletes a sitemap from a Search Console site property

## Security

- read risk: external Google Search Console API read of site/sitemap metadata and search analytics performance data
- write risk: adds or removes Search Console site properties and submits or deletes sitemap resources
- approval: reverse ETL writes require plan preview and approval token
- Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

## Commands

### Inspect as a manual

```bash
pm connectors inspect google-search-console
```

### Inspect as structured JSON

```bash
pm connectors inspect google-search-console --json
```

## Agent Rules

- Run pm connectors inspect google-search-console before creating credentials or plans.
- Use --json only when the caller needs structured output; use the manual for human-readable guidance.
- Never ask the user to paste secret values into chat.
- For reverse ETL writes, create a plan, show the preview, wait for explicit approval, then run with the approval token.
