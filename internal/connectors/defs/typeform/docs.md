# Overview

Typeform reads Typeform forms, workspaces, themes, and images through the Typeform REST API. This
bundle migrates `internal/connectors/typeform` (the hand-written connector) to a declarative defs
bundle — 4 of its 5 legacy streams are at capability parity here; the 5th (`responses`) needs a
per-form-id sub-resource fan-out and is deferred (see Known limits). The legacy package stays
registered and unchanged until wave6's registry flip.

## Auth setup

Provide a Typeform personal access token (or OAuth access token) via the `access_token` secret; it
is used only for Bearer auth (`Authorization: Bearer <access_token>`) and is never logged.

## Streams notes

Four streams, all `GET`, all list endpoints reading `items[]`:

- `forms` (`GET /forms`, primary key `id`, cursor field `last_updated_at`). `computed_fields`
  derives `is_public` from the nested `settings.is_public` object field, and `theme_href` from
  `theme.href`. `self_href` is derived from `_links.display` — legacy actually prefers
  `self.href` over `_links.display` when BOTH are present on a form object (a second, overriding
  assignment in `typeformFormRecord`), but Typeform's real `GET /forms` list response only ever
  carries `_links.display`, not a `self.href` (that field appears on the single-form `GET
  /forms/{id}` response shape, which this stream never calls) — so the dropped fallback never
  changes behavior for any record the real list endpoint returns. See Known limits.
- `workspaces` (`GET /workspaces`, primary key `id`). `self_href` derived from `self.href`.
- `themes` (`GET /themes`, primary key `id`). No computed fields; every field schema-projects
  directly.
- `images` (`GET /images`, primary key `id`). No computed fields; every field schema-projects
  directly.

Pagination is `page_number` (`page`/`page_size` query params, page size 200, short-page stop) —
matches legacy's own page/page_size loop. Legacy additionally consults the response's `page_count`
field to stop early even on a full-size final page; the engine's `page_number` paginator only
implements the short-page stop (no `page_count` awareness), which is functionally equivalent for
every real Typeform response except the rare case where the very last page happens to be exactly
`page_size` records long (legacy would stop on `page_count` there; this bundle would issue one
extra request that returns 0 records and then stop) — documented as a minor, non-data-changing
scope narrowing.

## Write actions & risks

None. Typeform is read-only in both legacy and this bundle (`capabilities.write: false`) — creating
forms/webhooks are not meaningful reverse-ETL write targets.

## Known limits

- **`responses` stream is NOT implemented in this bundle.** Legacy's `responses` stream fans out
  across a configured `form_ids` list, issuing one `GET /forms/{form_id}/responses` request PER
  form id and stamping each emitted record with its source `form_id`. This is a sub-resource
  fan-out read — one of the named Tier-2 `StreamHook` triggers in
  `docs/migration/conventions.md` §1's Tier-2 table ("sub-resource fan-out reads … issue → comments
  per issue") — and cannot be expressed in `streams.json`/`spec.json` alone: the dialect has no
  mechanism for a stream to iterate a multi-valued config property (`form_ids`) and issue one
  request per value. Implementing it requires a `hooks/typeform/hooks.go` `StreamHook`, out of
  scope for this Tier-1-only wave. See `api_surface.json`'s `excluded` entry.
- `forms`' `self_href` computed field does not reproduce legacy's `self.href`-overrides-
  `_links.display` fallback (the dialect has no coalesce/priority combinator across two distinct
  record paths for a single output field) — `_links.display` alone is used, which is the only field
  Typeform's real `GET /forms` list response ever populates for this purpose (see Streams notes).
- `page_count`-based early pagination stop is not modeled; only the short-page stop is (see
  Streams notes) — functionally equivalent except for an exact-final-page-size edge case.
- Full Typeform API surface (webhooks, form/workspace mutation, etc.) is out of scope for this
  wave; see `api_surface.json`'s `excluded` entries.
