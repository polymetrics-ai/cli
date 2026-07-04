# Overview

Google Search Console reads site properties, submitted sitemaps, and Search Analytics performance
reports through the Search Console API. This Pass B bundle adds the documented site and sitemap
detail reads plus explicit write actions for adding/removing site properties and submitting/deleting
sitemaps. The five Search Analytics streams remain handled by the existing Tier-2 StreamHook because
the API is a semantic read carried in a POST body with in-body pagination.

## Auth setup

Provide a Google OAuth 2.0 access token with Search Console scope via the `access_token` secret; it
is used only for Bearer auth and is never logged. `base_url` defaults to
`https://www.googleapis.com/webmasters/v3`.

## Streams notes

`sites` lists accessible site properties, and `site_details` reads one configured `site_url`.
`sitemaps` is hook-dispatched to preserve legacy's `site_urls` else `site_url` fallback and its
stringification of the API's numeric warning/error counts. `sitemap_details` reads one configured
`site_url` plus `feedpath`. The Search Analytics streams fan out over `site_urls` (or legacy's
single `site_url` fallback) and are hook-dispatched because `searchAnalytics.query` requires a POST
body containing date range, dimensions, rowLimit, and startRow pagination state.

## Write actions & risks

`add_site` and `delete_site` add or remove Search Console site properties. `submit_sitemap` and
`delete_sitemap` submit or remove sitemap resources for a site. Deletes are marked destructive and
treat 404 as idempotent success. All reverse ETL writes require plan preview and approval.

## Known limits

- URL Inspection (`POST https://searchconsole.googleapis.com/v1/urlInspection/index:inspect`) is
  excluded: it is a read-like POST endpoint on a separate API root with request-body parameters, and
  the current hook only covers Search Analytics POST-body reads under the webmasters v3 base URL.
- `StreamSpec.Body` is still unwired in the generic declarative read path, so Search Analytics
  remains a Tier-2 StreamHook implementation rather than a pure JSON stream. The `sitemaps` stream
  is also hook-dispatched because the fan-out dialect cannot express legacy's `site_urls` else
  `site_url` fallback and the schema projection path cannot stringify numeric warning/error counts.
- `access_token` is the only declared secret key. Legacy also accepted
  `authorization.access_token`; callers should map credentials to this bundle's declared key.
