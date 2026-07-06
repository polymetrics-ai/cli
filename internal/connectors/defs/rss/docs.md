# Overview

Reads RSS channel metadata and feed items from any RSS 2.0 feed URL. Read-only and credential-free.

Readable streams: `items`, `channel`.

This connector is read-only; no write actions are declared.

Service API documentation: https://www.rssboard.org/rss-specification.

## Auth setup

Connection fields:

- `feed_url` (optional, string); default `https://xkcd.com/rss.xml`; format `uri`; The RSS feed URL
  to read.

Default configuration values: `feed_url=https://xkcd.com/rss.xml`.

Authentication behavior:

- No authentication.

Requests use the configured `feed_url` value after applying defaults and send an
`Accept: application/rss+xml, application/xml, text/xml, */*` header.

Connection checks use a connector-managed request.

## Streams notes

Default pagination: single request; no pagination.

Incremental streams use their declared cursor fields and send lower-bound parameters only when a
lower bound is available. Note that although the `published_at` cursor field is advertised, every
read is a full feed read — reads are not incremental.

- `items`: GET connector-managed request path - records path `channel.item`; incremental cursor
  `published_at`; formatted as `rfc3339`. Record ids are selected as `guid`, falling back to
  `link`, then `title`. `published_at` carries the raw RSS `pubDate` value unreformatted.
- `channel`: GET connector-managed request path - records path `channel`. Record ids are selected
  as `link`, falling back to `title`. `updated_at` carries the raw `lastBuildDate` value
  unreformatted.

## Write actions & risks

This connector is read-only. Read behavior: external RSS feed read (XML over HTTP/HTTPS).

## Known limits

- Batch defaults: read_page_size=100.
- API coverage includes 2 stream-backed endpoint group(s).
