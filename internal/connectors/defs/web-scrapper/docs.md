# Overview

Web Scrapper is a read-only declarative-HTTP connector migrated from
`internal/connectors/web-scrapper` (legacy wave2 fan-out). It reads sitemap and scraping job
metadata from the Web Scraper Cloud API. This bundle is capability-parity with the legacy
hand-written connector; the legacy package stays registered and unchanged until wave6's registry
flip.

## Auth setup

Provide a Web Scraper Cloud API token via the `api_token` secret; it is sent as the `api_token`
query parameter on every request (`auth: [{"mode": "api_key_query", "param": "api_token", ...}]`)
and is never logged. `base_url` defaults to `https://api.webscraper.io/api/v1` and may be
overridden for tests or proxies.

## Streams notes

2 streams: `sitemaps` (`GET /sitemap`, records at `data`) and `jobs` (`GET /scraping-job`, records
at `data`). Primary key is `["id"]` for both; neither stream is incremental or paginated — legacy's
`Read` issues exactly one unparameterized request per stream (`r.Do(ctx, http.MethodGet,
endpoint.resource, nil, nil)`) and this bundle mirrors that exactly (no `query` block, no
`pagination` block).

## Write actions & risks

None. This bundle covers only the read surface legacy implemented; `capabilities.write` is `false`
and this bundle ships no `writes.json`.

## Known limits

- Only the 2 legacy-parity read streams are implemented; other Web Scraper Cloud endpoints
  (single-sitemap lookup, per-job scraped-data download, sitemap/job creation, job deletion) are
  out of scope for this migration wave — see `api_surface.json`'s `excluded: {category:
  out_of_scope, reason: "Pass B capability expansion"}` entries.
- Neither stream declares pagination: legacy issues exactly one request per stream with no paging
  loop, and this bundle mirrors that (no `pagination` block in `streams.json`).
