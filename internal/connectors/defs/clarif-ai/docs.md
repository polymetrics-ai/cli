# Overview

Clarif-ai reads Clarifai applications, datasets, models, model versions, and workflows, and
writes application/dataset lifecycle mutations, through the Clarifai v2 REST API
(`https://api.clarifai.com/v2`). It began as a wave2 fan-out migration of
`internal/connectors/clarif-ai` (the hand-written connector it migrates; the legacy package stays
registered and unchanged until wave6's registry flip) and was expanded with 4 write actions in a
Pass B pass. `capabilities.write` is now `true`.

## Auth setup

Provide a Clarifai personal access token (PAT) via the `api_key` secret; it is sent as
`Authorization: Key <api_key>` (`mode: api_key_header`, `header: Authorization`, `prefix: "Key "`),
matching legacy's `connsdk.APIKeyHeader("Authorization", secret, "Key ")` exactly. Every stream is
scoped under `users/{user_id}/...`, so the required `user_id` config value must also be set
(legacy's `clarifaiUserID`). `base_url` defaults to `https://api.clarifai.com/v2` and may be
overridden for tests/proxies. The `create_dataset`/`delete_dataset` write actions additionally
require the (new, optional-in-spec-but-required-for-those-2-actions) `app_id` config value, since
Clarifai scopes datasets under a specific application (`users/{user_id}/apps/{app_id}/datasets`).

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

4 write actions, added in this Pass B pass. Clarifai's public docs site is gRPC/SDK-first with no
single browsable REST reference (unlike bitly/clickup's OpenAPI-backed sites), so each action
below is grounded in an independently-confirmed request-body example rather than guessed from
convention alone — see `api_surface.json`'s `scope` note for what was excluded because it could
NOT be confirmed to the same bar.

- **`create_application`** (`POST /users/{{ config.user_id }}/apps`, `body_fields: ["apps"]`):
  Clarifai's plural-array POST convention — the record itself must carry the wrapper key, e.g.
  `{"apps": [{"id": "my-app", "description": "..."}]}` (confirmed via a direct curl example:
  `--data-raw '{"apps": [{"id": "test-application"}]}'`). Low-risk (additive).
- **`update_application`** (`PATCH /users/{{ config.user_id }}/apps`,
  `body_fields: ["action", "apps"]`): Clarifai's PATCH convention requires a top-level `"action"`
  field (`"merge"` to update-in-place, `"overwrite"` to fully replace the named fields) alongside
  the same plural `"apps"` array (confirmed via a direct documented example:
  `{"action": "overwrite", "apps": [{"id": "...", "default_workflow_id": "..."}]}`).
  **Approval required** — `action=overwrite` fully replaces the named fields rather than merging,
  so an operator should review the resolved action value before use.
- **`create_dataset`** (`POST /users/{{ config.user_id }}/apps/{{ config.app_id }}/datasets`,
  `body_fields: ["datasets"]`): the identical plural-array convention scoped to a specific app
  (confirmed via a direct curl example: `{"datasets": [{"id": "dataset-1633032323", "description":
  "...", "metadata": {...}}]}`). Low-risk (additive).
- **`delete_dataset`** (`DELETE /users/{{ config.user_id }}/apps/{{ config.app_id }}/datasets`,
  `body_type: none`, `body_fields: ["dataset_ids"]`): Clarifai's DELETE requests carry a JSON body
  (confirmed via direct documentation: `{"dataset_ids": ["YOUR_DATASET_ID_HERE"]}` against the
  plural `datasets` collection path, not a per-id path segment) — the engine's `body_type: none`
  + `body_fields` combination sends exactly this body with no other record fields leaking in.
  **Approval required** — deleting a dataset also deletes all its inputs/annotations and cannot be
  undone (Clarifai's own docs: "deleted datasets cannot be recovered").

Whole-application deletion (`DELETE /v2/users/{user_id}/apps`) is deliberately NOT implemented:
research found no independently-confirmed exact request-body shape (candidates seen were `ids`,
an `app_ids` field, or a `delete_all` flag, none confirmed to the same bar as the 4 actions above)
and the blast radius of a wrong shape — or of the wrong records reaching a correctly-shaped
request — is irreversible whole-application data loss; see `api_surface.json`'s
`destructive_admin` exclusion.

## Known limits

- **Check request is not bounded to 1 result.** Legacy's `Check` sends a bounded
  `GET /users/{user_id}/apps?per_page=1` read to confirm auth/connectivity cheaply. The engine's
  `check` dialect (`RequestSpec`) has only `method`/`path` fields — no `query` field at all
  (`engine.Check` calls `rt.Requester.Do(ctx, method, checkPath, nil, nil)` with a literal `nil`
  query) — so this bundle's check request has no `per_page` bound. This changes the check
  request's shape (an unbounded list read instead of a 1-item bound) but never emits records or
  mutates anything either way, so it is data-neutral; documented here per conventions.md §5's
  meta-rule (deviation acceptable iff it never changes emitted record data).
- Model/workflow prediction, training, and input/annotation ingestion remain out of scope — those
  require image/text/video payload encoding and concept-vector bodies this pass did not
  independently re-verify to the confidence bar the 4 implemented writes meet; see
  `api_surface.json`'s per-endpoint `excluded` entries (each carries a specific category and
  reason, not a blanket "Pass B" bucket).
- `app_id` is a new, optional `spec.json` property required only by `create_dataset`/
  `delete_dataset` (every read stream and the other 2 write actions do not need it) — an operator
  who only uses the read streams or the 2 app-level writes never needs to set it.
