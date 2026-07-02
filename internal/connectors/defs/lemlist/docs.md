# Overview

Lemlist is a cold outreach / sales engagement platform. This bundle reads its team (workspace),
campaigns, activities, and unsubscribes through the Lemlist REST API
(`https://api.lemlist.com/api`). It migrates `internal/connectors/lemlist` (the hand-written
connector); the legacy package stays registered and unchanged until wave6's registry flip.

## Auth setup

Provide a Lemlist API key via the `api_key` secret. It is sent as the `access_token` query
parameter on every request (`api_key_query` auth mode) — Lemlist's documented auth convention;
never logged.

## Streams notes

`campaigns`, `activities`, and `unsubscribes` are offset/limit list endpoints returning a
root-level JSON array (`records.path: ""`); pagination advances `offset` by 100 (`limit`/`offset`
query params) and stops on a short page (fewer than 100 records), matching legacy's
`harvest`/`RecordsAt(resp.Body, "")` loop exactly. `team` returns a single root-level JSON object
(no array) — modeled as `records.path: ""` with `single_object: true` and a stream-level
`pagination: {"type": "none"}` override, since the workspace endpoint is not paged upstream. No
stream has a legacy incremental filter mechanism — every legacy stream is full-refresh only (no
`created`/`updatedAt`-based server-side filtering), so no `incremental` block is declared anywhere
in this bundle. Every Lemlist object exposes a string `_id`, so every schema's `x-primary-key` is
`["_id"]`.

## Write actions & risks

None. This bundle is read-only (`capabilities.write: false`); Lemlist exposes no reverse-ETL write
surface here (matches legacy's `Write` returning `connectors.ErrUnsupportedOperation`
unconditionally).

## Known limits

- Legacy exposed runtime-configurable `page_size` (1-100) and `max_pages` (0/all/unlimited) config
  knobs. Neither is expressible in this dialect: `PaginationSpec`'s `page_size`/`max_pages` fields
  are fixed JSON literals in `streams.json`, never resolved from `RuntimeConfig.Config` at read
  time (no existing bundle in this codebase templates them from config; this mirrors the stripe
  golden's own documented `page_size`/`max_pages`-is-dead-config precedent, ledger item 3,
  `docs/migration/conventions.md` §5). This bundle fixes the page size at 100 (Lemlist's own
  documented max and legacy's own default) and leaves pagination unbounded (`max_pages` absent =
  no cap), matching legacy's own default behavior for a caller that never overrides either knob.
  Documented scope narrowing, not a data-shape deviation — the emitted records for any given page
  are identical either way.
- Full Lemlist API surface (lead add/remove-to-campaign, unsubscribe management, webhooks,
  schedules, senders) is out of scope for this wave; see `api_surface.json`'s `excluded:
  {category: out_of_scope, reason: "Pass B capability expansion"}` entries.
