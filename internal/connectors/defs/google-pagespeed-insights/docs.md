# Overview

Reads Lighthouse PageSpeed Insights reports (performance, accessibility, best-practices, SEO, PWA
scores) for the configured URLs and strategies via the PageSpeed Insights v5 API.

Readable streams: `pagespeed_reports`.

This connector is read-only; no write actions are declared.

Service API documentation: https://developers.google.com/speed/docs/insights/v5/get-started.

## Auth setup

Connection fields:

- `api_key` (optional, secret, string); Google PageSpeed API Key. See <a
  href="https://developers.google.com/speed/docs/insights/v5/get-started#APIKey">here</a>. The key
  is optional - however the API is heavily rate limited when using without API Key. Creating and
  using the API key therefore is recommended. The key is case sensitive.
- `base_url` (optional, string).
- `categories` (required, string); Defines which Lighthouse category to run. One or many of:
  "accessibility", "best-practices", "performance", "pwa", "seo".
- `mode` (optional, string).
- `strategies` (required, string); The analyses strategy to use. Either "desktop" or "mobile".
- `urls` (required, string); The URLs to retrieve pagespeed information from. The connector will
  attempt to sync PageSpeed reports for all the defined URLs. Format: https://(www.)url.domain.

Secret fields are redacted in logs and write previews: `api_key`.

Provide the secret fields listed above. Authentication is applied by the connector-specific
implementation for this service.

Requests use the configured `base_url` value after applying defaults.

Connection checks use a connector-managed request.

## Streams notes

Default pagination: single request; no pagination.

- `pagespeed_reports`: GET connector-managed request path - records path `data`.

## Write actions & risks

This connector is read-only; no reverse-ETL write actions are declared.

## Known limits

- API coverage includes 1 stream-backed endpoint group(s).
