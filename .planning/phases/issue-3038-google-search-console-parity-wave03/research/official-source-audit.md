# Google Search Console official source audit — wave03

- Source: https://www.googleapis.com/discovery/v1/apis/searchconsole/v1/rest
- Title/version/revision: Google Search Console API / v1 / 20260729
- Root/base: `https://searchconsole.googleapis.com/` + ``; baseUrl `https://searchconsole.googleapis.com/`
- Official operation count: 11
- Lane counts: etl_read=4, reverse_etl_write=4, direct_read_query_search=3, binary_file=0, cdc_changefeed=0, excluded_not_applicable=0

| Lane | Method | Path | Operation ID | Notes |
| --- | --- | --- | --- | --- |
| direct_read_query_search | POST | `/v1/urlInspection/index:inspect` | `searchconsole.urlInspection.index.inspect` | Index inspection. |
| direct_read_query_search | POST | `/v1/urlTestingTools/mobileFriendlyTest:run` | `searchconsole.urlTestingTools.mobileFriendlyTest.run` | Runs Mobile-Friendly Test for a given URL. |
| direct_read_query_search | POST | `/webmasters/v3/sites/{siteUrl}/searchAnalytics/query` | `webmasters.searchanalytics.query` | Query your data with filters and parameters that you define. Returns zero or more rows grouped by the row keys that you define. You must define a date range of one or more days. When date is one of the group by values, any days without data are omitted from the result list. If you need to know which days have data, issue a broad date range query grouped by date for any metric, and see which day rows are returned. |
| etl_read | GET | `/webmasters/v3/sites` | `webmasters.sites.list` | Lists the user's Search Console sites. |
| etl_read | GET | `/webmasters/v3/sites/{siteUrl}` | `webmasters.sites.get` | Retrieves information about specific site. |
| etl_read | GET | `/webmasters/v3/sites/{siteUrl}/sitemaps` | `webmasters.sitemaps.list` | Lists the [sitemaps-entries](/webmaster-tools/v3/sitemaps) submitted for this site, or included in the sitemap index file (if `sitemapIndex` is specified in the request). |
| etl_read | GET | `/webmasters/v3/sites/{siteUrl}/sitemaps/{feedpath}` | `webmasters.sitemaps.get` | Retrieves information about a specific sitemap. |
| reverse_etl_write | DELETE | `/webmasters/v3/sites/{siteUrl}` | `webmasters.sites.delete` | Removes a site from the set of the user's Search Console sites. |
| reverse_etl_write | PUT | `/webmasters/v3/sites/{siteUrl}` | `webmasters.sites.add` | Adds a site to the set of the user's sites in Search Console. |
| reverse_etl_write | DELETE | `/webmasters/v3/sites/{siteUrl}/sitemaps/{feedpath}` | `webmasters.sitemaps.delete` | Deletes a sitemap from the Sitemaps report. Does not stop Google from crawling this sitemap or the URLs that were previously crawled in the deleted sitemap. |
| reverse_etl_write | PUT | `/webmasters/v3/sites/{siteUrl}/sitemaps/{feedpath}` | `webmasters.sitemaps.submit` | Submits a sitemap for a site. |

Audit conclusion: the official discovery document still contains exactly 11 operations. No official binary or CDC surfaces are present. URL Inspection and Mobile Friendly Test are Search Console API roots outside `/webmasters/v3`; both are read-like POST operations and require bounded typed direct-read command metadata rather than a generic query/raw API escape hatch.

Implementation note (wave03 r1): local bundle metadata now implements all 11 official operations with 0 exclusions. `searchanalytics.query`, `urlInspection.index.inspect`, and `urlTestingTools.mobileFriendlyTest.run` are represented by closed-schema, `json_redacted`, 1 MiB capped direct-read operations; the four write operations are typed reverse-ETL actions, with destructive deletes redacted and idempotent on HTTP 404. Existing dimension-specific Search Analytics ETL streams remain documented as stream conveniences over the one official Search Analytics POST operation. This audit note is fixture-only and does not claim live provider certification.
