# Overview

Reads Lighthouse PageSpeed Insights reports for each configured HTTPS URL and mobile or desktop strategy through the fixed PageSpeed Insights v5 API. Every request repeats the declared Lighthouse categories: accessibility, best-practices, performance, PWA, and SEO.

Readable streams: `pagespeed_reports`.

This connector is read-only; no write actions are declared.

Service API documentation: https://developers.google.com/speed/docs/insights/v5/get-started.

## Auth setup

Connection fields:

- `api_key` (optional, secret, string); optional PageSpeed API key sent only as the declared `key` query parameter.
- `strategies` (required, string); comma-separated, unique strategies: `mobile` and/or `desktop`.
- `urls` (required, string); comma-separated, unique HTTPS URLs to analyse.

Secret fields are redacted in logs and write previews: `api_key`.

The runtime applies only the source-declared optional API-key query authentication. It does not accept caller-provided origins, fixture modes, or category overrides.

Connection checks make one bounded request to the fixed PageSpeed origin.

## Streams notes

Default pagination: single report request; no pagination.

- `pagespeed_reports`: GET `/runPagespeed`; emits one flattened report for every URL and strategy pair.

## Write actions & risks

This connector is read-only; no reverse-ETL write actions are declared.

## Known limits

- The source lock permits at most twenty URL-and-strategy report requests per read.
- API coverage includes 1 stream-backed endpoint group.
