# Overview

Wufoo is a wave2 fan-out declarative-HTTP migration. It reads forms, entries, and reports through
the Wufoo API v3 (`GET {{ config.base_url }}/...`). This bundle is migrated from
`internal/connectors/wufoo` (the hand-written connector it replaces); the legacy package stays
registered and unchanged until wave6's registry flip. Read-only (`capabilities.write` is `false`,
matching legacy's `Write` returning `connectors.ErrUnsupportedOperation`).

## Auth setup

Provide a Wufoo API key via the `api_key` secret; it is sent as the HTTP Basic username with the
literal password `pass` (`mode: basic`), matching legacy's `connsdk.Basic(apiKey, "pass")` exactly
— Wufoo's API convention accepts any password value when authenticating with an API key as the
username. `base_url` defaults to `https://example.wufoo.com/api/v3` (legacy's `defaultBaseURL`, a
placeholder subdomain) and must be overridden with the account's actual subdomain
(`https://<subdomain>.wufoo.com/api/v3`).

## Streams notes

All 3 streams share `page_number` pagination (`page`/`pageSize` query params, matching legacy's
`PageNumberPaginator{PageParam: "page", SizeParam: "pageSize", StartPage: 1}`). `page_size`
defaults to 100 (legacy's `defaultPageSize`); legacy bounds it to a max of 1000 (`maxPageSize`) and
`max_pages` defaults to 1 (legacy's `readMaxPages` default) when unset.

`forms` (`GET /forms.json`) emits `Hash`/`Name`/`DateUpdated` from the `Forms` array, matching
legacy's field set and PascalCase naming exactly (Wufoo's own API convention; no renaming via
`computed_fields` is needed since legacy itself preserves the raw field names verbatim). `reports`
(`GET /reports.json`) emits the identical shape from the `Reports` array. `entries` (`GET
/forms/{{ config.form_hash }}/entries.json`) emits `EntryId`/`DateCreated`/`DateUpdated` from the
`Entries` array — the path requires the `form_hash` config value to resolve, matching legacy's
`resolveResource`, which errors with `"wufoo config form_hash is required for entries and must be
a path segment"` for the identical missing/invalid case (this bundle's path-templating error is a
more generic engine message, not legacy's specific wording — see Known limits). Primary key is
`Hash` for `forms`/`reports` and `EntryId` for `entries`; `DateUpdated` is declared as the
incremental cursor field for manifest-surface parity on all 3 streams, matching legacy's
`cursorFields`, though neither legacy nor this bundle actually issues a server-side incremental
filter — legacy's `Read` performs a full stream read every time regardless of any prior cursor.

All 3 streams declare `"projection": "passthrough"`: legacy's `Read` hands `connsdk.Harvest` a
callback that does `emit(connectors.Record(rec))` verbatim for every stream — no field-built
`connectors.Record{...}` mapping anywhere in `wufoo.go` — so schema-mode projection would silently
drop any Wufoo field beyond the three currently declared per stream; passthrough reproduces
legacy's actual raw-emission behavior exactly, and the schema remains a documentation surface of
the known shape.

## Write actions & risks

None. Legacy `wufoo.go`'s `Write` returns `connectors.ErrUnsupportedOperation` unconditionally;
`capabilities.write` is `false` and this bundle ships no `writes.json`.

## Known limits

- **`page_size`/`max_pages` config-driven overrides are not modeled.** Legacy reads
  `config["page_size"]` (bounded 1-1000) and `config["max_pages"]` (default 1) at request time via
  `boundedInt`/`readMaxPages`. The engine's `page_number` paginator reads `PaginationSpec.PageSize`
  from the static `streams.json` `base.pagination` block only — there is no per-request
  config-driven override mechanism for either value in the current dialect. `page_size`/`max_pages`
  remain declared in `spec.json` as documentation of legacy's accepted config surface, but neither
  is wired into any template in this bundle.
- **`form_hash` path-segment validation error wording is generic, not legacy's specific
  message.** Legacy's `resolveResource` rejects an empty or `/?#`-containing `form_hash` with a
  named error (`"wufoo config form_hash is required for entries and must be a path segment"`); the
  engine's path interpolation instead surfaces its own generic unresolved-key or
  traversal-rejection error for the same inputs (see `docs/migration/conventions.md`'s path
  interpolation rules: an absent `config.*` key is a hard error naming the key/namespace; a
  path-traversal-shaped value is rejected outright). The class of rejected input is identical;
  only the error message differs. This is not a behavior change for any accepted input.
- **No incremental filter is modeled**, matching legacy: `DateUpdated` is declared as
  `x-cursor-field` for manifest parity, but Wufoo's `/forms.json`, `/forms/<hash>/entries.json`,
  and `/reports.json` endpoints (as legacy calls them) accept no time-range query parameter — both
  connectors always perform a full stream read on every sync.
- The full Wufoo API surface (form/entry mutation, widgets, comments, webhooks) is out of scope
  for this wave; see `api_surface.json`'s `excluded` entries.
