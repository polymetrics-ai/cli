# Google PageSpeed Insights Connector

## Overview

Reads Lighthouse PageSpeed Insights reports for a bounded Cartesian product of configured HTTPS URLs and mobile or desktop strategies through the fixed PageSpeed Insights v5 API.

Readable streams: `pagespeed_reports`.

Service API documentation: https://developers.google.com/speed/docs/insights/v5/get-started.

## Auth setup

Connection fields:

- `api_key` (optional, secret, string); Optional PageSpeed API key sent only as the declared key query parameter.
- `strategies` (required, string); Comma-separated, unique PageSpeed strategies: mobile and/or desktop.
- `urls` (required, string); Comma-separated, unique HTTPS URLs to analyse. The declared URL and strategy product is capped at twenty requests.

Authentication uses declared mode(s): `api_key_query`, `none`.

## Execution contract

Connection check: `GET /runPagespeed`
Check query: `category`=`performance`; `strategy`=`desktop`; `url`=`https://example.com`.

## Streams notes

- `pagespeed_reports`: `GET /runPagespeed`; records `.`

## Write actions & risks

This connector's write surface is declared separately in the rendered execution bundle.
