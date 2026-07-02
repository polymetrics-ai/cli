# Overview

Pingdom is a read-only declarative migration of `internal/connectors/pingdom` (legacy Go
connector). It reads Pingdom checks, probes, actions, maintenance windows, and reference data
through Pingdom's REST API 3.1. This bundle is capability-parity with legacy; legacy stays
registered and unchanged until wave6's registry flip.

## Auth setup

Provide a Pingdom API 3.1 Bearer token via the `api_key` secret; it is used only for Bearer auth
(`Authorization: Bearer <api_key>`) and is never logged. There is no unauthenticated fallback mode —
legacy hard-errors when `api_key` is unset (`pingdom connector requires secret api_key`), matching
this bundle's Bearer `auth` candidate and `api_key`'s `x-secret` presence check at request time.

## Streams notes

All 5 streams (`checks`, `probes`, `actions`, `maintenance`, `reference`) share the identical
`offset_limit` pagination shape (`limit`/`offset` query params, `page_size: 100` matching legacy's
`defaultPageSize`), and all pass through the RAW record with no field-filtering — legacy's `Read`
calls `connsdk.Harvest` directly with no `mapRecord`-style projection function at all, so every
stream declares `"projection": "passthrough"` to preserve that exact behavior. None of the five
streams declares an `incremental` block or an `x-cursor-field`: legacy never filters or advances
reads by any timestamp field on any of these resources (`lasttesttime`/`time` are ordinary
passthrough fields, not cursors) — every read is a full stream read, matching the sync-mode
derivation rule (an `x-cursor-field` with no backing `incremental` block would misrepresent
capability this connector does not have).

`checks`/`probes`/`actions`/`maintenance` extract records from their eponymous top-level array key
(`records.path: "checks"`, etc.). `reference` is different: legacy's `streamEndpoints["reference"]`
declares an EMPTY `recordsPath` (`""`), so `connsdk.RecordsAt` wraps the entire response body as a
single record (Pingdom's `/reference` endpoint returns one lookup bundle — check types, regions,
probe metadata — not a keyed collection). This bundle mirrors that exactly via
`"records": {"path": ".", "single_object": true}`; `schemas/reference.json` declares no
`x-primary-key` since there is no natural per-record identity for a singleton reference-data dump.

## Write actions & risks

None. Legacy `Write` always returns `connectors.ErrUnsupportedOperation`; `metadata.json` declares
`capabilities.write: false` and no `writes.json` file exists, matching legacy exactly.

## Known limits

- `page_size` config validation (legacy's 1-25000 numeric-range check) is not reproduced at the
  bundle level; the engine treats `page_size` as an opaque value substituted directly into the
  `limit` query param, sent to Pingdom as-is rather than rejected client-side the way legacy's
  `strconv.Atoi` range check would. This never changes emitted record DATA for any legacy-valid
  input; it only narrows client-side input validation, out of scope for wave2 fan-out (Pass B).
- Legacy's `max_pages` config (a non-negative integer, or the keywords `all`/`unlimited` for
  unbounded) has no bundle-level equivalent — this bundle relies solely on the offset paginator's own
  short-page stop signal (no `MaxPages` hard-cap declared), matching legacy's own unbounded default
  (`maxPages() == 0` when `max_pages` is unset) exactly for the common case. Out of scope for wave2
  fan-out (Pass B).
- The 2-page conformance fixture (`fixtures/streams/checks/page_1.json` /`page_2.json`) uses a
  synthetic 100-record first page purely to exercise the real `offset_limit` short-page stop
  condition against the bundle's real `page_size: 100` default (the engine's `OffsetPaginator` stops
  only when a page returns fewer records than `page_size`); this is a fixture-authoring artifact
  proving pagination correctness, not a claim about Pingdom's actual typical result-set size.
