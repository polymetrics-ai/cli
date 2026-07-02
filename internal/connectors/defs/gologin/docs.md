# Overview

GoLogin is a browser-profile management API. This bundle reads GoLogin browser profiles, folders,
tags, and account information through the GoLogin REST API (`https://api.gologin.com`). It is a
Tier-1 declarative migration of `internal/connectors/gologin` (the legacy hand-written connector,
which stays registered and unchanged until wave6's registry flip). GoLogin is read-only in pm.

## Auth setup

Provide a GoLogin API key via the `api_key` secret; it is used only for Bearer auth
(`Authorization: Bearer <api_key>`) and is never logged.

## Streams notes

- `profiles` (`GET /browser/v2`, records at `profiles`): paginated with `pagination.type:
  page_number`, `page_param: page`, `size_param: ""` (GoLogin's profiles list accepts no
  page-size query parameter at all — legacy always requests `?page=N` with no size override), and
  `page_size: 30` (matches legacy's `gologinDefaultPageSize`, used purely as the client-side
  short-page stop threshold: a page returning fewer than 30 records ends the walk, identical to
  legacy's `harvest` loop). Primary key `id`, cursor field `updatedAt`.
- `folders` (`GET /folders`, records at the response root `.`): a bare JSON array, not paginated
  (`pagination.type: none`), matching legacy's `readSinglePage`. Primary key `id`.
- `user` (`GET /user`, records at the response root `.`): a single JSON object (not an array), not
  paginated. Primary key `_id`, cursor field `createdAt`.
- `tags` (`GET /tags/all`, records at `tags`): a bare list, not paginated. Primary key `_id`.

Every field is a direct schema projection matching legacy's `mapRecord` functions field-for-field
(no renames were needed — GoLogin's raw field names already match legacy's emitted keys), each
expressed as an explicit bare `{{ record.<path> }}` `computed_fields` entry so typed fields
(`profilesCount`, an integer) survive with their native JSON type rather than being stringified.

## Write actions & risks

None. GoLogin is read-only in pm (`capabilities.write: false`, no `writes.json`) — matching
legacy's `Write` stub, which unconditionally returns `connectors.ErrUnsupportedOperation`.

## Known limits

- `page_size`/`max_pages` config validation (legacy's numeric-range checks and `all`/`unlimited`
  keyword parsing for `max_pages`) is not reproduced at the bundle-config level: the engine's
  `pagination.page_size`/`pagination.max_pages` fields are static JSON literals, not templated from
  `config.*`, so there is no runtime-config-driven override mechanism for either value (matching
  the documented `orb`/`stripe` precedent for this exact gap class). `page_size: 30` and unbounded
  `max_pages` (absent, i.e. unlimited) are baked into `streams.json`'s `profiles` pagination block
  instead. A caller-supplied `page_size`/`max_pages` config value is accepted by `spec.json` for
  documentation/parity-of-intent but has no runtime effect — this never changes emitted record
  DATA for any legacy-valid input; it only narrows client-side config-surface flexibility, out of
  scope for this wave (Pass B if ever needed).
- Full GoLogin API surface (creating/updating/deleting profiles, proxy management, plan details) is
  out of scope; see `api_surface.json`'s `excluded` entries. Only the 4 legacy-parity read streams
  are implemented.
