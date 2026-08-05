# Google Search Console provider-field matrix

Researched 2026-08-05 against Google's current provider-published Discovery
description, revision `20260802`, and the linked REST reference pages. This matrix is
the citation research required by the connector-resume contract; it is intentionally
not a competing machine-readable bundle format while the shared citation convention is
landing.

## Operation inventory and reconciliation

Primary inventory source: <https://www.googleapis.com/discovery/v1/apis/searchconsole/v1/rest>.
It enumerates 11 methods:

| Google operation ID | Method and path | Reachable connector surface |
| --- | --- | --- |
| `webmasters.sites.list` | `GET /webmasters/v3/sites` | `sites list` ETL |
| `webmasters.sites.get` | `GET /webmasters/v3/sites/{siteUrl}` | `sites get` ETL |
| `webmasters.sites.add` | `PUT /webmasters/v3/sites/{siteUrl}` | `sites add` typed reverse ETL |
| `webmasters.sites.delete` | `DELETE /webmasters/v3/sites/{siteUrl}` | `sites delete` typed destructive reverse ETL |
| `webmasters.sitemaps.list` | `GET /webmasters/v3/sites/{siteUrl}/sitemaps` | `sitemaps list` ETL |
| `webmasters.sitemaps.get` | `GET /webmasters/v3/sites/{siteUrl}/sitemaps/{feedpath}` | `sitemaps get` ETL |
| `webmasters.sitemaps.submit` | `PUT /webmasters/v3/sites/{siteUrl}/sitemaps/{feedpath}` | `sitemaps submit` typed reverse ETL |
| `webmasters.sitemaps.delete` | `DELETE /webmasters/v3/sites/{siteUrl}/sitemaps/{feedpath}` | `sitemaps delete` typed destructive reverse ETL |
| `webmasters.searchanalytics.query` | `POST /webmasters/v3/sites/{siteUrl}/searchAnalytics/query` | five `search-analytics by-*` ETL conveniences |
| `searchconsole.urlInspection.index.inspect` | `POST /v1/urlInspection/index:inspect` | `direct url-inspection inspect` |
| `searchconsole.urlTestingTools.mobileFriendlyTest.run` | `POST /v1/urlTestingTools/mobileFriendlyTest:run` | `direct mobile-friendly-test run` |

The human REST reference index at
<https://developers.google.com/webmaster-tools/v1/api_reference_index> visibly lists
the first ten methods (Search Analytics 1, Sitemaps 4, Sites 4, URL Inspection 1).
Its navigation no longer links a Mobile-Friendly Test page, but Google's current
machine-readable Discovery document still publishes
`searchconsole.urlTestingTools.mobileFriendlyTest.run`; its old prose URL returns 404.
For parity this is therefore a provider-published operation, not an invented one. The
Discovery document is the source of record for the **11-operation** total and the
Mobile-Friendly Test field citation below.

The preserved 15-row API ledger was not a 15-operation inventory: five ledger rows
describe one Search Analytics POST endpoint so every dimension-specific ETL stream can
be covered. That yields four duplicate ledger rows. The preserved 16 generated commands
also included a direct Search Analytics command, but current-main runtime safety rejects
the documented URL-valued `siteUrl` path before dispatch. It was a redundant, false
implemented claim because the operation already reaches the five ETL stream conveniences;
the connector therefore retires it rather than carrying a hollow command. The resulting
generated surface has 15 honest commands.

## Citation key

- **T1-equivalent**: Google's provider-published machine-readable Discovery operation
  or request schema. Google does not publish an OpenAPI document for this service; this
  is the closest first-party machine-readable request definition.
- **T3**: the provider's REST reference page anchored to the named operation and
  request/parameters section.
- **T4**: same-operation description used only where a field's explicit required marker
  is absent.
- Requiredness is stated as **provider-required**, **provider-optional**, or
  **connector-required**. The last form is a deliberate bounded-command constraint,
  never a claim that Google marked the field required.

Generic Google transport parameters (`alt`, `fields`, `key`, `oauth_token`,
`prettyPrint`, `quotaUser`, and similar common parameters) are not service-operation
fields in this connector. OAuth is supplied only via the connection's secret-bearing
auth configuration; the connector intentionally exposes no generic query or raw-body
escape hatch.

## Sites and Sitemaps

| Operation / field | Source URL and section | Evidence | Confidence | Requiredness rationale | Connector treatment |
| --- | --- | --- | --- | --- | --- |
| `sites.get`, `sites.add`, `sites.delete` — `siteUrl` path | <https://developers.google.com/webmaster-tools/v1/sites/get#parameters>, <https://developers.google.com/webmaster-tools/v1/sites/add#parameters>, <https://developers.google.com/webmaster-tools/v1/sites/delete#parameters> | T3 path-parameter tables; T1-equivalent operation parameters | High | Each operation declares `siteUrl` as a required path parameter. | `site_details` config/path or typed write `record.site_url`; required by write schemas. |
| `sites.list` | <https://developers.google.com/webmaster-tools/v1/sites/list> | T3 operation page and T1-equivalent operation | High | No service-specific request field. | Account-scoped `sites` stream. |
| `sitemaps.list` — `siteUrl` path | <https://developers.google.com/webmaster-tools/v1/sitemaps/list#parameters> | T3 parameter table; T1-equivalent operation parameters | High | Provider-required path parameter. | `sitemaps` stream fan-out site property. |
| `sitemaps.list` — `sitemapIndex` query | <https://developers.google.com/webmaster-tools/v1/sitemaps/list#parameters> | T3 parameter table; T1-equivalent optional query parameter | High | Provider-optional; list without it returns submitted sitemaps. | Not exposed: operation remains reachable without the optional index filter; raw arbitrary query is prohibited. |
| `sitemaps.get` — `siteUrl`, `feedpath` paths | <https://developers.google.com/webmaster-tools/v1/sitemaps/get#parameters> | T3 parameter table; T1-equivalent operation parameters | High | Both are provider-required path parameters. | `sitemap_details` config/path. |
| `sitemaps.submit` — `siteUrl`, `feedpath` paths | <https://developers.google.com/webmaster-tools/v1/sitemaps/submit#parameters> | T3 parameter table; T1-equivalent operation parameters | High | Both are provider-required path parameters. | Closed `submit_sitemap` write record. |
| `sitemaps.delete` — `siteUrl`, `feedpath` paths | <https://developers.google.com/webmaster-tools/v1/sitemaps/delete#parameters> | T3 parameter table; T1-equivalent operation parameters | High | Both are provider-required path parameters. | Closed destructive `delete_sitemap` write record. |

## Search Analytics query

Source operation: <https://developers.google.com/webmaster-tools/v1/searchanalytics/query>.
The provider's current prose page has the request body and explicit required markers for
`startDate` and `endDate`; the Discovery schema records the current full field set,
including `dataState`, `searchType`, hourly dimensions, and the filter-object schema.

| Field | Source URL and section | Evidence | Confidence | Requiredness rationale | Connector treatment |
| --- | --- | --- | --- | --- | --- |
| `siteUrl` path | `searchanalytics/query#parameters` | T3 parameter table; T1-equivalent operation parameters | High | Provider-required path parameter. | Required ETL fan-out property, safely path-escaped by the connector hook. |
| `startDate` body | `searchanalytics/query#request-body` | T3 marks Required; T1-equivalent request schema | High | Provider-required. | Stream config/cursor supplies it. |
| `endDate` body | `searchanalytics/query#request-body` | T3 marks Required; T1-equivalent request schema | High | Provider-required. | Stream config/default supplies it. |
| `dimensions[]` body | `searchanalytics/query#request-body` | T3 marks Optional; T1-equivalent enum schema | High | Provider-optional. | The by-date stream requests `date`; country/device/page/query streams request `date` plus their named dimension, producing per-date-per-dimension records while preserving `date` as cursor and primary-key component. |
| `type` body | `searchanalytics/query#request-body` | T3 Optional/default web; T1-equivalent enum schema | High | Provider-optional. | Streams use configured search type. |
| `searchType` body | Discovery `SearchAnalyticsQueryRequest.properties.searchType` | T1-equivalent field and enum schema | High | Provider-optional/default web. | Not separately exposed: current REST prose documents `type`; accepting both aliases would permit conflicting stream requests. |
| `dimensionFilterGroups[]` body | `searchanalytics/query#request-body` | T3 body structure; T1-equivalent `ApiDimensionFilterGroup` schema | High | Provider-optional. | Deferred from the bounded direct command: current safe CLI flags cannot express nested objects; raw JSON/body input is prohibited. This is a field-shape limitation, not an operation gap. |
| `dimensionFilterGroups[].groupType` | Discovery `ApiDimensionFilterGroup.properties.groupType` | T1-equivalent enum (`AND`) | High | Child of provider-optional group. | Covered by the same safe nested-object deferment. |
| `dimensionFilterGroups[].filters[]` | Discovery `ApiDimensionFilterGroup.properties.filters` | T1-equivalent array schema | High | Child of provider-optional group. | Covered by the same safe nested-object deferment. |
| `dimensionFilterGroups[].filters[].dimension` | Discovery `ApiDimensionFilter.properties.dimension` | T1-equivalent enum schema | High | Required by semantic filter object; Discovery does not carry a formal body-required list. | Covered by the same safe nested-object deferment. |
| `dimensionFilterGroups[].filters[].operator` | Discovery `ApiDimensionFilter.properties.operator` | T1-equivalent enum schema | High | Required by semantic filter object; Discovery does not carry a formal body-required list. | Covered by the same safe nested-object deferment. |
| `dimensionFilterGroups[].filters[].expression` | Discovery `ApiDimensionFilter.properties.expression` | T1-equivalent field schema | High | Required by semantic filter object; Discovery does not carry a formal body-required list. | Covered by the same safe nested-object deferment. |
| `aggregationType` body | `searchanalytics/query#request-body` | T3 Optional/default auto; T1-equivalent enum schema | High | Provider-optional. | Not exposed: per-dimension ETL streams use Google’s default aggregation; raw arbitrary body input is prohibited. |
| `rowLimit` body | `searchanalytics/query#request-body` | T3 Optional/default 1000 and max 25,000; T1-equivalent schema | High | Provider-optional. | Bounded `page_size` configuration, capped at 25,000. |
| `startRow` body | `searchanalytics/query#request-body` | T3 Optional/default 0; T1-equivalent schema | High | Provider-optional. | The ETL hook starts at zero and advances it safely while paginating. |
| `dataState` body | Discovery `SearchAnalyticsQueryRequest.properties.dataState` | T1-equivalent enum schema | High | Provider-optional (the schema does not label it required). `DATA_STATE_UNSPECIFIED` is deliberately not exposed because Google's enum description says it should not be used. | Stream `data_state` configuration; no raw default override. |

## URL Inspection and Mobile-Friendly Test

| Operation / field | Source URL and section | Evidence | Confidence | Requiredness rationale | Connector treatment |
| --- | --- | --- | --- | --- | --- |
| `urlInspection.index.inspect` — `inspectionUrl` body | <https://developers.google.com/webmaster-tools/v1/urlInspection.index/inspect#request-body> | T3 marks Required; T1-equivalent `InspectUrlIndexRequest` schema | High | Provider-required. | Required `--inspection-url`. |
| `urlInspection.index.inspect` — `siteUrl` body | <https://developers.google.com/webmaster-tools/v1/urlInspection.index/inspect#request-body> | T3 marks Required; T1-equivalent `InspectUrlIndexRequest` schema | High | Provider-required. | Required `--site-url`. |
| `urlInspection.index.inspect` — `languageCode` body | <https://developers.google.com/webmaster-tools/v1/urlInspection.index/inspect#request-body> | T3 marks Optional/default `en-US`; T1-equivalent schema | High | Provider-optional. | Optional `--language-code`. |
| `urlTestingTools.mobileFriendlyTest.run` — `url` body | Discovery `RunMobileFriendlyTestRequest.properties.url` | T1-equivalent request schema plus same-operation description "Runs Mobile-Friendly Test for a given URL" | Medium | Google no longer provides the old prose operation page and Discovery does not render an explicit Required marker. The only request subject is the URL, so the connector treats it as connector-required for a meaningful invocation; the inference is explicitly T4, not a fabricated provider-required marker. | Required `--url`; source availability is a documented prose-page Tier-5 gap, field citation itself is present. |
| `urlTestingTools.mobileFriendlyTest.run` — `requestScreenshot` body | Discovery `RunMobileFriendlyTestRequest.properties.requestScreenshot` | T1-equivalent schema, default false | High | Provider-optional/default false. | Optional boolean `--request-screenshot`. |

## Citation coverage result

- **32/32 operation-specific field paths have a provider-owned citation record** in this
  matrix, including nested Search Analytics filter fields and the optional `sitemapIndex`
  query parameter that the bounded connector intentionally does not expose.
- **0 fields are uncited.** The Mobile-Friendly Test field has a source-availability
  caveat: Google’s prose page is 404, but the current Google Discovery document supplies
  the machine-readable request field. Its provider-required marker remains unverified;
  the connector-required flag is explicitly recorded as a medium-confidence inference.
- **No documented operation remains planned.** Optional unsupported fields are explicitly
  named above and do not create an unrepresented API operation.
