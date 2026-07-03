# Overview

Typeform reads Typeform forms, responses, workspaces, themes, and images through the Typeform
REST API. This bundle migrates `internal/connectors/typeform` (the hand-written connector) to a
declarative defs bundle — all 5 legacy streams are now at capability parity here. `responses`
(previously deferred as needing a per-form-id sub-resource fan-out) is implemented via
`streams.json`'s `fan_out` dialect (S4 engine mini-wave item 2, `docs/migration/conventions.md`
§3) — no Tier-2 `StreamHook` was needed. The legacy package stays registered and unchanged until
wave6's registry flip.

## Auth setup

Provide a Typeform personal access token (or OAuth access token) via the `access_token` secret; it
is used only for Bearer auth (`Authorization: Bearer <access_token>`) and is never logged.

## Streams notes

Five streams, all `GET`, all list endpoints reading `items[]`:

- `forms` (`GET /forms`, primary key `id`, cursor field `last_updated_at`). `computed_fields`
  derives `is_public` from the nested `settings.is_public` object field, and `theme_href` from
  `theme.href`. `self_href` is derived from `_links.display` — legacy actually prefers
  `self.href` over `_links.display` when BOTH are present on a form object (a second, overriding
  assignment in `typeformFormRecord`), but Typeform's real `GET /forms` list response only ever
  carries `_links.display`, not a `self.href` (that field appears on the single-form `GET
  /forms/{id}` response shape, which this stream never calls) — so the dropped fallback never
  changes behavior for any record the real list endpoint returns. See Known limits.
- `responses` (`GET /forms/{form_id}/responses`, primary key `response_id`, cursor field
  `submitted_at`). Legacy's `readResponses` fans out across a configured `form_ids` list, issuing
  one request PER form id and stamping each emitted record with its source `form_id`
  (`internal/connectors/typeform/typeform.go`'s `harvest`/`perForm` branch). This bundle reproduces
  that exact shape via `streams.json`'s `fan_out`: `ids_from.config_key: "form_ids"` (a
  comma-separated config value, split/trimmed exactly like legacy's own `typeformFormIDs` helper),
  `into.path_var` (the resolved id is referenced as `{{ fanout.id }}` in `stream.path`:
  `/forms/{{ fanout.id }}/responses`), and `stamp_field: "form_id"` (stamps the fanned-out id onto
  every emitted record after projection — matching legacy's stamp-if-absent behavior; legacy's own
  `mapRecord` also copies a `form_id` field straight off the raw item when present, so the stamp is
  a no-op whenever Typeform's real response already carries one). Every other field
  (`response_id`/`token`/`landing_id`/`landed_at`/`submitted_at`/`answers`/`hidden`/`calculated`/
  `metadata`) is a direct schema-projection passthrough, matching legacy's `typeformResponseRecord`
  exactly (no `computed_fields` needed). Each configured form id runs its own independent
  pagination/incremental sequence (the engine's fan-out contract), matching legacy's per-form-id
  HTTP loop. `form_ids` is a required config value for this stream (matching legacy's own hard
  error — `"typeform responses stream requires config form_ids"` — when the list is empty).
- `workspaces` (`GET /workspaces`, primary key `id`). `self_href` derived from `self.href`.
- `themes` (`GET /themes`, primary key `id`). No computed fields; every field schema-projects
  directly.
- `images` (`GET /images`, primary key `id`). No computed fields; every field schema-projects
  directly.

Pagination is `page_number` (`page`/`page_size` query params, page size 200, short-page stop) —
matches legacy's own page/page_size loop, for every stream including the fanned-out `responses`
(each form id's own request sequence uses the identical base pagination spec). Legacy additionally
consults the response's `page_count` field to stop early even on a full-size final page; the
engine's `page_number` paginator only implements the short-page stop (no `page_count` awareness),
which is functionally equivalent for every real Typeform response except the rare case where the
very last page happens to be exactly `page_size` records long (legacy would stop on `page_count`
there; this bundle would issue one extra request that returns 0 records and then stop) —
documented as a minor, non-data-changing scope narrowing.

## Write actions & risks

None. Typeform is read-only in both legacy and this bundle (`capabilities.write: false`) — creating
forms/webhooks are not meaningful reverse-ETL write targets.

## Known limits

- **RESOLVED — `responses` stream is now implemented** via `streams.json`'s `fan_out` dialect (see
  Streams notes above); previously deferred pending a Tier-2 `StreamHook`, now closed by the S4
  engine mini-wave's `fan_out` addition (`docs/migration/conventions.md` §3). No Go hook was
  needed.
- `forms`' `self_href` computed field does not reproduce legacy's `self.href`-overrides-
  `_links.display` fallback (the dialect has no coalesce/priority combinator across two distinct
  record paths for a single output field) — `_links.display` alone is used, which is the only field
  Typeform's real `GET /forms` list response ever populates for this purpose (see Streams notes).
- `page_count`-based early pagination stop is not modeled; only the short-page stop is (see
  Streams notes) — functionally equivalent except for an exact-final-page-size edge case. This also
  applies to `responses`' per-form-id request sequence.
- Full Typeform API surface (webhooks, form/workspace mutation, etc.) is out of scope for this
  wave; see `api_surface.json`'s `excluded` entries.
