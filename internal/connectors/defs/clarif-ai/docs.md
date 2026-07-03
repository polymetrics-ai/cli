# Overview

Clarif-ai reads Clarifai applications, datasets, models, model versions, and workflows through
the Clarifai v2 REST API (`https://api.clarifai.com/v2`). This bundle is a wave2 fan-out
migration of `internal/connectors/clarif-ai` (the hand-written connector it migrates); the legacy
package stays registered and unchanged until wave6's registry flip. Clarifai is read-only
(full refresh only); `capabilities.write` is `false` and this bundle ships no `writes.json`.

## Auth setup

Provide a Clarifai personal access token (PAT) via the `api_key` secret; it is sent as
`Authorization: Key <api_key>` (`mode: api_key_header`, `header: Authorization`, `prefix: "Key "`),
matching legacy's `connsdk.APIKeyHeader("Authorization", secret, "Key ")` exactly. Every stream is
scoped under `users/{user_id}/...`, so the required `user_id` config value must also be set
(legacy's `clarifaiUserID`). `base_url` defaults to `https://api.clarifai.com/v2` and may be
overridden for tests/proxies.

## Streams notes

All 5 streams (`applications`, `datasets`, `models`, `model_versions`, `workflows`) share the
identical shape: `GET /users/{{ config.user_id }}/<resource>`, records at a resource-named
top-level key (`apps`/`datasets`/`models`/`model_versions`/`workflows`), primary key `["id"]`, no
incremental cursor (legacy is full-refresh only — Clarifai list responses expose no incremental
filter parameter). `model_versions` is intentionally NOT scoped under a specific model id; it
reads the flat `users/{user_id}/models/versions` collection endpoint, exactly matching legacy's
own `clarifaiStreamEndpoints["model_versions"].resource = "models/versions"`.

Pagination is `page`/`per_page` (`pagination.type: page_number`, `page_param: page`,
`size_param: per_page`, `start_page: 1`) with the engine's standard short-page stop (a page
returning fewer than `page_size` records — including an empty page — ends the stream), matching
legacy's own loop (`len(records) < pageSize`) exactly. Clarifai's list envelope has no
total/has_more flag, so this short-page heuristic is legacy's own termination signal, not an
approximation.

Legacy exposes `page_size`/`max_pages` as config-driven overrides (`clarifaiPageSize`/
`clarifaiMaxPages`, default 100, max 1000). The engine's `PaginationSpec.PageSize`/`MaxPages`
fields are static JSON integers (no `{{ config.* }}` templating support), so these cannot be wired
as runtime-configurable knobs — this bundle declares neither `page_size` nor `max_pages` in
`spec.json` at all (a declared-but-unwireable key is worse than an absent one, per
conventions.md F6). This mirrors bitly's identical documented "`page_size` is not
runtime-configurable" limitation. `streams.json`'s base pagination pins `page_size: 100`, matching
legacy's real default (`clarifaiDefaultPageSize`) exactly — a live request always asks for the same
per-page size legacy would have asked for absent an explicit override. The required first-stream
2-page fixture proof (conventions.md §4) is built the same way as chargify's identical precedent:
`applications` page 1 returns exactly 100 records (a full page, so the paginator advances) and page
2 returns 1 (a short page, so the paginator terminates); every other stream's single fixture page
returns 1 record against a declared `per_page=100`, correctly read as an already-short page.

## Write actions & risks

None. Clarif-ai is a read-only source connector (legacy's own package doc: "The Clarifai source is
read-only (full refresh)"); `capabilities.write` is `false` and this bundle ships no
`writes.json`, matching legacy's `Write` returning `connectors.ErrUnsupportedOperation`.

## Known limits

- **Check request is not bounded to 1 result.** Legacy's `Check` sends a bounded
  `GET /users/{user_id}/apps?per_page=1` read to confirm auth/connectivity cheaply. The engine's
  `check` dialect (`RequestSpec`) has only `method`/`path` fields — no `query` field at all
  (`engine.Check` calls `rt.Requester.Do(ctx, method, checkPath, nil, nil)` with a literal `nil`
  query) — so this bundle's check request has no `per_page` bound. This changes the check
  request's shape (an unbounded list read instead of a 1-item bound) but never emits records or
  mutates anything either way, so it is data-neutral; documented here per conventions.md §5's
  meta-rule (deviation acceptable iff it never changes emitted record data).
- Full Clarifai API surface (predictions, training, input/annotation ingestion, async
  training-status polling) is out of scope; see `api_surface.json`'s `excluded:
  {category: out_of_scope, reason: "Pass B capability expansion"}` entries. Only the 5
  legacy-parity read streams are implemented.
