# Overview

Tavus is a declarative-HTTP migration of `internal/connectors/tavus` (the hand-written legacy
connector this bundle migrates; the legacy package stays registered and unchanged until wave6's
registry flip). It reads Tavus faces (replicas), videos, conversations, PALs, guardrails,
objectives, documents, pronunciation dictionaries, voices, and skills, and writes approved
video/conversation/PAL/guardrail/objective/document/pronunciation-dictionary mutations through the
Tavus REST API (`https://tavusapi.com/v2`).

This is a Pass B full-surface expansion re-researched against Tavus's CURRENT OpenAPI spec
(`docs.tavus.io/openapi.yaml`), not just its docs pages: **the entire "replica"/"persona" naming
Tavus used at wave2 has since been renamed account-wide to "face"/"PAL"**. `/v2/replicas` and
`replica_id`, and `/v2/personas` and `persona_id`/`default_replica_id`, are all still explicitly
documented as supported LEGACY ALIASES for the new `/v2/faces`/`pal_id` shapes — never deprecated,
never scheduled for removal. This bundle's `replicas` stream therefore keeps its wave2 path
unchanged (still fully correct and supported); every NEW stream/write added this pass uses the
CURRENT naming (`pals`, not `personas`) since there is no legacy alias to preserve for resources
that didn't exist at wave2.

## Auth setup

Provide a Tavus API key via the `api_key` secret; it is sent as the `x-api-key` header
(`api_key_header` auth mode, no prefix). Never logged. `base_url` defaults to
`https://tavusapi.com/v2` and may be overridden for tests/proxies.

## Streams notes

Pagination conventions are genuinely mixed across Tavus's own resources — each stream declares its
own `pagination` override rather than sharing `base.pagination`'s `page`/`page_size` (1-based, the
wave2 `replicas` shape):

- `replicas` (legacy-parity, unchanged from wave2): `GET /replicas`, `page`/`page_size`
  (1-based), records at `data`, primary key `["id"]`. `id`/`name` computed-field-renamed from the
  raw `replica_id`/`replica_name`.
- `videos`: `GET /videos`, `page`/`limit` (0-BASED — Tavus's own docs default `page` to 0),
  records at `data`, primary key `["id"]`. `id`/`name` renamed from `video_id`/`video_name`.
- `conversations`: `GET /conversations`, `page`/`limit` (1-based), records at `data`, primary key
  `["id"]`, `x-cursor-field: created_at`. `id`/`name` renamed from `conversation_id`/
  `conversation_name`.
- `pals`: `GET /pals`, `page`/`limit` (1-based), records at `data`, primary key `["id"]`.
  `id`/`name` renamed from `pal_id`/`pal_name`.
- `guardrails`: `GET /guardrails?legacy=false` (the static `legacy=false` query param is REQUIRED —
  omitting it defaults to Tavus's old legacy-guardrail-SET list shape, a different, incompatible
  response envelope; `legacy=false` returns individual guardrails, "recommended for all new
  integrations" per Tavus's own docs), `page`/`limit` (0-based), records at `data`, primary key
  `["id"]`. `id`/`name` renamed from `uuid`/`guardrail_name`.
- `objectives`: `GET /objectives`, `page`/`limit` (1-based), records at `data`, primary key
  `["id"]`. `id`/`name` renamed from `objectives_id`/`objective_name`.
- `documents`: `GET /documents`, `page`/`limit` (0-based), records at `data`, primary key `["id"]`.
  `id`/`name` renamed from `document_id`/`document_name`.
- `pronunciation_dictionaries`: `GET /pronunciation-dictionaries`, `page`/`limit` (0-based),
  records at `data`, primary key `["id"]`. `id` renamed from `pronunciation_dictionary_id`; no
  `name` rename needed (Tavus's own `name` field already matches).
- `voices`: `GET /voices`, `page`/`limit` (1-based), records at `data`. NO single natural id field
  (a stock voice slug can be linked to more than one face); primary key is the composite
  `["voice_name", "face_id"]`.
- `skills`: `GET /skills`, records at `data`, `pagination.type: none` — Tavus's own docs show no
  `page`/`limit` query parameters at all for this endpoint (a small, static-ish registry), unlike
  every other list endpoint in this bundle. Primary key `["skill_id"]`.

## Write actions & risks

- `create_video`/`delete_video`: generates/deletes an async rendered video. `create_video`
  consumes billed video-generation minutes; `delete_video`'s `missing_ok_status: [400]` matches
  Tavus's own convention of answering an invalid/missing id with HTTP 400, never 404 (confirmed
  against the OpenAPI spec's documented response codes for every delete endpoint in this bundle).
- `create_conversation`/`end_conversation`/`delete_conversation`: starts, ends, or destructively
  deletes a real-time video conversation. `create_conversation` begins consuming
  conversational-minutes billing immediately and places a live call (unless `test_mode`);
  `end_conversation` is routine call cleanup (not destructive to history); `delete_conversation` is
  the destructive variant Tavus's own docs explicitly warn to prefer `end_conversation` over for
  routine cleanup.
- `create_pal`/`delete_pal`: creates/deletes a PAL persona; low-risk creation, but deleting a PAL
  breaks any conversation still referencing its `pal_id`.
- `create_guardrail`/`delete_guardrail`: creates/deletes a behavioral guardrail; deleting one
  immediately removes that boundary from every PAL referencing it via `guardrail_ids`.
- `create_objective`/`delete_objective`: creates (as an array — Tavus's `POST /objectives` accepts
  one or more objectives per call in a single `data` array) or deletes a PAL objective.
- `create_document`/`delete_document`: uploads a knowledge-base document (processing is
  asynchronous; the document becomes usable by PALs only once `status` reaches `ready`) or deletes
  one (removing it from every PAL referencing it via `document_ids`).
- `create_pronunciation_dictionary`/`delete_pronunciation_dictionary`: creates/deletes a
  pronunciation-rule dictionary.

`metadata.json` now declares `capabilities.write: true`.

## Known limits

- Every rename/PATCH endpoint in this API (face rename, video rename, PAL/guardrail/objective/
  pronunciation-dictionary PATCH) uses a JSON Patch (RFC 6902) array request body — this dialect's
  `body_type` (`json`/`form`/`none`) has no JSON-Patch-array shape, so these are excluded as
  `out_of_scope`/`ENGINE_GAP`-flavored, not silently approximated with a flat-object PATCH that
  would not match the real API's expected body shape.
- `lipsync` and `replacements` (both full endpoint groups) are excluded as `deprecated` — Tavus's
  own OpenAPI spec marks every operation in both groups `deprecated: true` with "no longer
  supported by Tavus" in the Lipsync group's own description.
- Face/video training and generation writes that create billable, asynchronous, or rights-sensitive
  work (`POST /faces` face training) are excluded as `out_of_scope` — training a face requires the
  caller to attest to footage/likeness rights per Tavus's own platform policy, not a plain
  reverse-ETL create action this pass models.
- Tools, PAL-tool/PAL-skill attachment sub-resources, deployments (widget/embed delivery
  configuration), canvas interactions (real-time Magic Canvas protocol calls), and transcriptions
  are excluded as `out_of_scope` — each is a separate feature-specific object domain distinct from
  the core face/video/conversation/PAL data this pass prioritizes; Pass B breadth-vs-cost triage.
- **`page_size` is not runtime-configurable** on `replicas` (carried over from the wave2 golden):
  the engine's `page_number` paginator constructor reads `PaginationSpec.PageSize` as a static
  bundle-level integer, not a config-templated field.
- All fixtures (`fixtures/streams/**`, `fixtures/writes/**`, `fixtures/check.json`) represent
  Tavus's real wire shape, including each stream's own distinct 0-based-vs-1-based page-numbering
  convention documented above.
