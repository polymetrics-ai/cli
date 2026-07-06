# Overview

Unknown Filter In Template is a deliberately invalid connectorgen validate corpus case: its
`streams.json` `widgets` stream's `computed_fields.tags` template references an unknown filter
name (`reverse`), which `engine.ResolveCheck`'s filter-name validation must reject.

## Auth setup

No auth required.

## Streams notes

`widgets` is a single stream, no pagination.

## Write actions & risks

None; read-only bundle.

## Known limits

None; this is test fixture data.
