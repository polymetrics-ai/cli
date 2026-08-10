# pm connectors inspect google-search-console

```text
NAME
  pm connectors inspect google-search-console - Google Search Console connector manual

SYNOPSIS
  pm connectors inspect google-search-console
  pm connectors inspect google-search-console --json
  pm credentials add <name> --connector google-search-console [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Reads Google Search Console sites, sitemaps, and Search Analytics performance reports (by date, query, page, country, and device) through the Search Console v3 REST API; submits/removes sites and sitemaps through explicit write actions.

ICON
  id: googlesearchconsole
  asset: icons/googlesearchconsole.svg
  source: upstream_registry
  review_status: upstream_seeded
  review_url: https://developers.google.com/search/news

CAPABILITIES
  check=true catalog=true read=true write=true query=false
  Integration type: api

AUTHENTICATION
  Use pm credentials add with --from-env or --value-stdin for secret fields.

CONFIGURATION
  base_url
  data_state
  end_date
  feedpath
  inspection_url
  max_pages
  mobile_test_url
  mode
  page_size
  search_type
  site_url
  site_urls
  start_date
  access_token (secret)

ETL STREAMS
  sites:
    primary key: site_url
    fields: permission_level(string), site_url(string)
  site_details:
    primary key: site_url
    fields: permission_level(string), site_url(string)
  sitemaps:
    primary key: site_url, path
    fields: errors(string), is_pending(boolean), is_sitemaps_index(boolean), last_downloaded(string), last_submitted(string), path(string), site_url(string), type(string), warnings(string)
  sitemap_details:
    primary key: site_url, path
    fields: errors(string), is_pending(boolean), is_sitemaps_index(boolean), last_downloaded(string), last_submitted(string), path(string), site_url(string), type(string), warnings(string)
  search_analytics_by_date:
    primary key: site_url, search_type, date
    cursor: date
    fields: clicks(number), ctr(number), date(string), impressions(number), position(number), search_type(string), site_url(string)
  search_analytics_by_country:
    primary key: site_url, search_type, date, country
    cursor: date
    fields: clicks(number), country(string), ctr(number), date(string), impressions(number), position(number), search_type(string), site_url(string)
  search_analytics_by_device:
    primary key: site_url, search_type, date, device
    cursor: date
    fields: clicks(number), ctr(number), date(string), device(string), impressions(number), position(number), search_type(string), site_url(string)
  search_analytics_by_page:
    primary key: site_url, search_type, date, page
    cursor: date
    fields: clicks(number), ctr(number), date(string), impressions(number), page(string), position(number), search_type(string), site_url(string)
  search_analytics_by_query:
    primary key: site_url, search_type, date, query
    cursor: date
    fields: clicks(number), ctr(number), date(string), impressions(number), position(number), query(string), search_type(string), site_url(string)

SYNC MODES
  ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped

REVERSE ETL ACTIONS
  add_site:
    endpoint: PUT /webmasters/v3/sites/{{ record.site_url | urlencode }}
    required fields: site_url
    risk: adds a site property to the authenticated Search Console account
  delete_site:
    endpoint: DELETE /webmasters/v3/sites/{{ record.site_url | urlencode }}
    required fields: site_url
    risk: removes a site property from the authenticated Search Console account
  submit_sitemap:
    endpoint: PUT /webmasters/v3/sites/{{ record.site_url | urlencode }}/sitemaps/{{ record.feedpath | urlencode }}
    required fields: site_url, feedpath
    risk: submits a sitemap URL for a Search Console site property
  delete_sitemap:
    endpoint: DELETE /webmasters/v3/sites/{{ record.site_url | urlencode }}/sitemaps/{{ record.feedpath | urlencode }}
    required fields: site_url, feedpath
    risk: deletes a sitemap from a Search Console site property

SECURITY
  read risk: external Google Search Console API read of site/sitemap metadata and search analytics performance data
  write risk: adds or removes Search Console site properties and submits or deletes sitemap resources
  approval: reverse ETL writes require plan preview and approval token
  Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

COMMAND SURFACE
  Run Google Search Console's declared streams and reverse-ETL actions.
  Usage: pm google-search-console <command> [flags]
  Read streams
  Reverse ETL writes
  Other Commands
    add site apply - Plan and execute the add site reverse-ETL action [intent=reverse_etl availability=implemented write=add_site]; approval: requires plan, preview, approval, and execute; risk: adds a site property to the authenticated Search Console account; flags: --site_url (required)
    delete site apply - Plan and execute the delete site reverse-ETL action [intent=reverse_etl availability=implemented write=delete_site]; approval: requires plan, preview, approval, and execute; risk: removes a site property from the authenticated Search Console account; flags: --site_url (required)
    delete sitemap apply - Plan and execute the delete sitemap reverse-ETL action [intent=reverse_etl availability=implemented write=delete_sitemap]; approval: requires plan, preview, approval, and execute; risk: deletes a sitemap from a Search Console site property; flags: --feedpath (required), --site_url (required)
    direct mobile-friendly-test run - Run the official Mobile Friendly Test operation as a typed bounded direct read. [intent=direct_read availability=implemented operation=google-search-console.mobile_friendly_test_run]; risk: low: bounded Search Console JSON read; fixed endpoint, closed request body schema, 1 MiB response cap, and redacted URL/screenshot-shaped fields; flags: --url (required), --request-screenshot, --page, --page-cursor
    direct url-inspection inspect - Run the official URL Inspection operation as a typed bounded direct read. [intent=direct_read availability=implemented operation=google-search-console.urlinspection_index_inspect]; risk: low: bounded Search Console JSON read; fixed endpoint, closed request body schema, 1 MiB response cap, and redacted URL-shaped fields; flags: --inspection-url (required), --site-url (required), --language-code, --page, --page-cursor
    search analytics by country list - Run the search analytics by country ETL stream [intent=etl availability=implemented stream=search_analytics_by_country]
    search analytics by date list - Run the search analytics by date ETL stream [intent=etl availability=implemented stream=search_analytics_by_date]
    search analytics by device list - Run the search analytics by device ETL stream [intent=etl availability=implemented stream=search_analytics_by_device]
    search analytics by page list - Run the search analytics by page ETL stream [intent=etl availability=implemented stream=search_analytics_by_page]
    search analytics by query list - Run the search analytics by query ETL stream [intent=etl availability=implemented stream=search_analytics_by_query]
    site details list - Run the site details ETL stream [intent=etl availability=implemented stream=site_details]
    sitemap details list - Run the sitemap details ETL stream [intent=etl availability=implemented stream=sitemap_details]
    sitemaps list - Run the sitemaps ETL stream [intent=etl availability=implemented stream=sitemaps]
    sites list - Run the sites ETL stream [intent=etl availability=implemented stream=sites]
    submit sitemap apply - Plan and execute the submit sitemap reverse-ETL action [intent=reverse_etl availability=implemented write=submit_sitemap]; approval: requires plan, preview, approval, and execute; risk: submits a sitemap URL for a Search Console site property; flags: --feedpath (required), --site_url (required)

EXAMPLES
  # Inspect as a manual
  pm connectors inspect google-search-console

  # Inspect as structured JSON
  pm connectors inspect google-search-console --json

AGENT WORKFLOW
  - Run pm connectors inspect google-search-console before creating credentials or plans.
  - Use --json only when the caller needs structured output; use the manual for human-readable guidance.
  - Never ask the user to paste secret values into chat.
  - For reverse ETL writes, create a plan, show the preview, wait for explicit approval, then run with the approval token.

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
