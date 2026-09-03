# Google Analytics 4 (GA4) Connector

## Overview

Reads fixed Google Analytics 4 reports from the Analytics Data API runReport endpoint through declared response-header projection.

Readable streams: `daily_active_users`, `website_overview`, `traffic_sources`, `devices`, `pages`.

Service API documentation: https://developers.google.com/analytics/devguides/reporting/data/v1/rest/v1beta/properties/runReport.

## Auth setup

Connection fields:

- `access_token` (required, secret, string); Google Analytics OAuth2 access token.
- `end_date` (optional, string); GA4 report end date or relative date token.
- `property_ids` (required, string); One numeric Google Analytics property ID.
- `start_date` (optional, string); GA4 report start date or relative date token.

Authentication uses declared mode(s): `bearer`.

## Execution contract

Connection check: `POST /v1beta/properties/{{ config.property_ids }}:runReport`
Check JSON body: `dateRanges`=[{'endDate': '{{ config.end_date }}', 'startDate': '{{ config.start_date }}'}], `dimensions`=[{'name': 'date'}], `keepEmptyRows`=false, `limit`=1, `metrics`=[{'name': 'activeUsers'}, {'name': 'newUsers'}, {'name': 'sessions'}], `offset`=0.

## Streams notes

- `daily_active_users`: `POST /v1beta/properties/{{ config.property_ids }}:runReport`; records `rows`
  - JSON body: `dateRanges`=[{'endDate': '{{ config.end_date }}', 'startDate': '{{ config.start_date }}'}], `dimensions`=[{'name': 'date'}], `keepEmptyRows`=false, `limit`=10000, `metrics`=[{'name': 'activeUsers'}, {'name': 'newUsers'}, {'name': 'sessions'}], `offset`=0.
- `website_overview`: `POST /v1beta/properties/{{ config.property_ids }}:runReport`; records `rows`
  - JSON body: `dateRanges`=[{'endDate': '{{ config.end_date }}', 'startDate': '{{ config.start_date }}'}], `dimensions`=[{'name': 'date'}], `keepEmptyRows`=false, `limit`=10000, `metrics`=[{'name': 'activeUsers'}, {'name': 'newUsers'}, {'name': 'sessions'}, {'name': 'screenPageViews'}, {'name': 'averageSessionDuration'}, {'name': 'bounceRate'}], `offset`=0.
- `traffic_sources`: `POST /v1beta/properties/{{ config.property_ids }}:runReport`; records `rows`
  - JSON body: `dateRanges`=[{'endDate': '{{ config.end_date }}', 'startDate': '{{ config.start_date }}'}], `dimensions`=[{'name': 'date'}, {'name': 'sessionSource'}, {'name': 'sessionMedium'}], `keepEmptyRows`=false, `limit`=10000, `metrics`=[{'name': 'sessions'}, {'name': 'activeUsers'}, {'name': 'newUsers'}, {'name': 'engagedSessions'}], `offset`=0.
- `devices`: `POST /v1beta/properties/{{ config.property_ids }}:runReport`; records `rows`
  - JSON body: `dateRanges`=[{'endDate': '{{ config.end_date }}', 'startDate': '{{ config.start_date }}'}], `dimensions`=[{'name': 'date'}, {'name': 'deviceCategory'}, {'name': 'operatingSystem'}, {'name': 'browser'}], `keepEmptyRows`=false, `limit`=10000, `metrics`=[{'name': 'activeUsers'}, {'name': 'sessions'}, {'name': 'screenPageViews'}], `offset`=0.
- `pages`: `POST /v1beta/properties/{{ config.property_ids }}:runReport`; records `rows`
  - JSON body: `dateRanges`=[{'endDate': '{{ config.end_date }}', 'startDate': '{{ config.start_date }}'}], `dimensions`=[{'name': 'date'}, {'name': 'pagePath'}, {'name': 'pageTitle'}], `keepEmptyRows`=false, `limit`=10000, `metrics`=[{'name': 'screenPageViews'}, {'name': 'activeUsers'}, {'name': 'averageSessionDuration'}], `offset`=0.

## Write actions & risks

This connector's write surface is declared separately in the rendered execution bundle.
