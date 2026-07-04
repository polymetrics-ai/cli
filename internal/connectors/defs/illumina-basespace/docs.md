# Overview

Illumina BaseSpace is a declarative HTTP connector for the documented BaseSpace v1pre3 REST API. This Pass B bundle preserves the five legacy user-scoped record projections for `projects`, `runs`, `samples`, `appsessions`, and `datasets`, then adds documented v1pre3 GET endpoints as passthrough streams plus direct documented POST, PUT, and DELETE endpoints as write actions.

## Auth setup

Provide a BaseSpace access token via the `access_token` secret; it is sent as the `x-access-token` header. `base_url` is required because BaseSpace is domain-scoped across regional hosts. `user` defaults to `current` and is used by the legacy user-scoped streams and check request.

## Streams notes

The legacy streams remain first and keep schema projection so their emitted records match `internal/connectors/illumina-basespace`. Newly added streams use `projection: passthrough` with permissive schemas. List endpoints read `Response.Items` and use BaseSpace `Offset`/`Limit` pagination at the fixed page size of 100; detail endpoints read `Response` and disable pagination. Path-scoped streams require the matching optional config key named in `spec.json` when selected.

## Write actions & risks

The bundle declares 20 write actions for documented POST, PUT, and DELETE endpoints that the Tier-1 dialect can express as one HTTP request per record. Record schemas require path parameters and allow documented JSON body fields to pass through. DELETE actions send no body and are marked `confirm: destructive`; other trash/stop-style actions are also marked destructive. Reverse ETL must still follow plan, preview, approval, and execute.

## Known limits

- The legacy `domain` config fallback is not modeled; `base_url` is required.
- `page_size` and `max_pages` remain legacy runtime concepts, but this declarative bundle uses a fixed pagination size of 100.
- Newly added stream schemas are permissive passthrough schemas derived from the documented `Response`/`Response.Items` envelopes, not hand-curated warehouse schemas.
- Legacy fixture-mode marker fields are not modeled; fixtures target live response shapes.
