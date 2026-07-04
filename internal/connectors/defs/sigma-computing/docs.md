# Overview

Sigma Computing is a wave2 fan-out declarative-HTTP migration. It reads Sigma workbooks, datasets,
teams, and members through the Sigma REST API (`GET https://api.sigmacomputing.com/v2/...`). This
bundle migrates `internal/connectors/sigma-computing` (the hand-written legacy connector) to a
declarative defs bundle at capability parity; the legacy package stays registered and unchanged
until wave6's registry flip.

## Auth setup

Provide a Sigma Computing OAuth access token via the `access_token` secret; it is sent as a Bearer
token (`Authorization: Bearer <access_token>`, matching legacy's `connsdk.Bearer(secret)` at
`sigma_computing.go:131`) and is never logged. `base_url` defaults to
`https://api.sigmacomputing.com` (legacy's `sigmaDefaultBaseURL`) and may be overridden for
tests/proxies.

## Streams notes

All four streams (`workbooks`, `datasets`, `teams`, `members`) share an identical shape: `GET
/v2/<stream>`, records at the top-level `entries` key, and a `page`-cursor pagination convention
(`pagination.type: cursor`, `cursor_param: page`, `token_path: nextPage` — the response's own
`nextPage` field is echoed back as the next request's `page` query param), matching legacy's
`harvest` loop at `sigma_computing.go:89-121` exactly: no `stop_path` is declared, since legacy
stops purely on an absent/empty `nextPage` value with no separate boolean stop signal. `limit` is
a config-driven per-page-size override (`{{ config.page_size }}`, defaulting to legacy's own
default of 100 via `spec.json`'s `page_size` default and the query param's own `default: "100"`),
matching legacy's `pageSize` resolution (`sigma_computing.go:213-226`).

Legacy's `sigmaRecord` mapper (`sigma_computing.go:147-149`) is shared **verbatim** by all four
endpoints — every stream emits the identical 4-field record shape (`id`, `name`, `email`,
`updated_at`), even though `email` is a members-only concept and workbooks/datasets/teams are
unlikely to ever populate it; this bundle reproduces that exact shared-mapper shape rather than
narrowing any stream's schema, matching legacy's actual emitted data field-for-field. `name` and
`updated_at` are mapped via `coalesce` computed fields to reproduce legacy's fallback chains:
`name` then `displayName`, and `updatedAt` then `updated_at`.
None of the four streams expose a real server-side incremental filter in legacy (no
date-range/updated-since query parameter is ever sent); `x-cursor-field: updated_at` is declared
purely as catalog/sort-key metadata matching legacy's own `CursorFields` declaration
(`sigma_computing.go:168`), and no `incremental` block is declared on any stream, matching legacy's
full-refresh-only read behavior exactly.

## Write actions & risks

None. Sigma Computing's legacy connector is read-only (`Write` returns
`connectors.ErrUnsupportedOperation`); `capabilities.write` is `false` and this bundle ships no
`writes.json`.

## Known limits

- **`max_pages` is not runtime-configurable.** Legacy exposes a `max_pages` config override
  (`0`/`all`/`unlimited` for unbounded, or a positive integer hard cap,
  `sigma_computing.go:227-237`). The engine's `PaginationSpec.MaxPages` is a fixed
  bundle-authored literal, not config-driven; this bundle leaves it unset (unbounded), matching
  legacy's own default.
- **Legacy's fixture-mode-only stamped fields are not modeled.** Legacy's `readFixture` path
  (only reached when `config.mode == "fixture"`) stamps a `fixture: true` marker onto every emitted
  record (`sigma_computing.go:157`); this is a credential-free conformance-harness affordance, not
  part of the live record shape, and is intentionally not reproduced — the engine's own
  `internal/connectors/conformance` fixture-replay harness provides the equivalent test
  affordance.
