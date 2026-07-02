# Overview

Split.io is a wave2 fan-out migration from `internal/connectors/split-io` (the legacy
hand-written connector this bundle replaces at capability parity). It reads Split.io workspaces,
environments, feature flags (splits), and segments through the Split Admin API. Read-only; the
legacy package stays registered and unchanged until wave6's registry flip.

## Auth setup

Provide a Split.io Admin API key via the `api_key` secret; it is used only for Bearer auth
(`Authorization: Bearer <api_key>`) and is never logged, matching legacy's
`connsdk.Bearer(secret)`.

## Streams notes

`workspaces` needs no path parameter. `environments`, `splits`, and `segments` each require
`config.workspace_id` to substitute the `{workspace_id}` path segment (`path:
"/internal/api/v2/environments/ws/{{ config.workspace_id }}"`, etc.) — legacy's
`resolveResource` hard-errors with `"stream requires config workspace_id"` when unset; this bundle
reproduces the identical requirement by declaring `workspace_id` as a plain (non-required-at-spec-
level, since `workspaces` doesn't need it) config property referenced directly in the three
path-scoped streams' `path` templates — an absent `config.workspace_id` is a hard interpolation
error on those three streams' reads exactly like legacy's explicit check, just surfaced through
the engine's own path-interpolation error rather than a bespoke message.

All 4 streams share the identical pagination shape: `offset_limit` (`limit_param: limit`,
`offset_param: offset`, `page_size: 100`), records at `objects` — matches legacy's
`connsdk.OffsetPaginator{LimitParam: "limit", OffsetParam: "offset", PageSize: pageSize}` reading
`endpoint.recordsPath` (`"objects"` for every stream) with `defaultPageSize = 100`.
`page_size`/`max_pages` were legacy config knobs with no equivalent config-driven mechanism in
this dialect (`PaginationSpec.PageSize`/`MaxPages` are static `streams.json` fields, not
`{{ }}`-templated) — see Known limits.

Legacy performs no incremental/state-cursor filtering during `Read` (no persisted cursor is read
or sent as a request filter) even though `splits`/`segments` declare `updatedAt` as a
`CursorFields` catalog hint; this bundle matches that exactly — `schemas/splits.json` and
`schemas/segments.json` declare `x-cursor-field: updatedAt` (descriptive, matching legacy's
catalog metadata) but no stream declares an `incremental` block, since legacy never actually
applies one during a read.

## Write actions & risks

None. Legacy `split_io.go`'s `Write` returns `connectors.ErrUnsupportedOperation` unconditionally;
`capabilities.write` is `false` and this bundle ships no `writes.json`.

## Known limits

- **`page_size`/`max_pages` are not exposed as config properties.** Legacy accepts
  `config.page_size` (bounded 1-1000, default 100) and `config.max_pages` (default unbounded) at
  read time. The engine's `PaginationSpec.PageSize`/`MaxPages` fields are static values baked into
  `streams.json`'s pagination block, with no `{{ }}` templating support from `config.*` — there is
  no per-run override mechanism at all (matching searxng's `page_size`/`max_pages` precedent, F6
  REVIEW.md). This bundle hard-codes `page_size: 100` (legacy's own default) and declares no
  `max_pages` (unbounded, matching legacy's own default when the config value is unset/`all`/
  `unlimited`). A caller that previously overrode either value per-run loses that capability;
  every default-config caller sees byte-identical behavior.
- Only the 4 legacy-parity streams are implemented; the wider Split Admin API (change requests,
  attribute definitions, API key management, restrictions) is out of scope for this wave — see
  `api_surface.json`'s `excluded: {category: out_of_scope}` entries.
