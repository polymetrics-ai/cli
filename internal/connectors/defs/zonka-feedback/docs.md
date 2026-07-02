# Overview

Zonka Feedback is a wave2 fan-out declarative-HTTP migration of
`internal/connectors/zonka-feedback` (the hand-written legacy connector this bundle migrates;
the legacy package stays registered and unchanged until wave6's registry flip). It reads Zonka
Feedback responses, surveys, and contacts through the Zonka Feedback REST API (`GET
https://api.zonkafeedback.com/...`). Read-only.

## Auth setup

Provide a Zonka Feedback API token via either the `auth_token` secret (preferred) or the
`access_token` secret (fallback), sent as a Bearer token (`Authorization: Bearer <token>`) and
never logged. This reproduces legacy's exact first-match-wins precedence
(`zonka_feedback.go:169-176`: `auth_token` is tried first, `access_token` only when `auth_token`
is unset) via the dialect's dual-auth-candidate-list ordering (conventions.md §3): `streams.json`
`base.auth` declares the `auth_token`-gated bearer candidate FIRST, the `access_token`-gated
bearer candidate second. `base_url` defaults to `https://api.zonkafeedback.com` and may be
overridden for tests/proxies.

## Streams notes

Three streams, identical shape: `responses` (`GET /responses`, records at `responses`),
`surveys` (`GET /surveys`, records at `surveys`), `contacts` (`GET /contacts`, records at
`contacts`). Pagination is `page_number` (`page_param: page`, `size_param: per_page`,
`start_page: 1`, `page_size: 100`), matching legacy's own `page`/`per_page` query params and
100-record default page size (`zonka_feedback.go:243-253`'s `pageSize`); the engine's built-in
page_number paginator stops on a short/final page exactly like legacy's own `len(records) <
size` check (`zonka_feedback.go:134`).

Legacy's `mapRecord` (`zonka_feedback.go:179-194`) copies every raw field verbatim onto the
emitted record, then backfills `id`/`name`/`updated_at` from an ordered list of alternate raw key
names (`idKeys`/`nameKeys`/`cursorKeys`) only when the primary key is absent on that record — but
legacy's own `Catalog()` (`zonka_feedback.go:70-80`) advertises exactly four fields (`id`, `name`,
`rating`, `updated_at`) as the stream's schema-facing contract for every one of the three
streams, identically. This bundle's schemas model that four-field contract (`"schema"`
projection mode, the default — conventions.md §2's schema-as-projection rule: only
schema-declared properties survive projection, and the legacy connector's own advertised Catalog
fields are the authoritative "what this stream emits" surface, not the unbounded raw passthrough
fields `mapRecord` additionally copies through at the Go-struct level but never advertises via
Catalog).

## Write actions & risks

None. Legacy `zonka-feedback` is read-only (`Write` returns
`connectors.ErrUnsupportedOperation`); `metadata.json` declares `capabilities.write: false` and
this bundle ships no `writes.json`.

## Known limits

- **The `idKeys`/`nameKeys`/`cursorKeys` alternate-key fallback chains are not modeled.** Legacy's
  `mapRecord` falls back to `response_id`/`survey_id`/`contact_id` for `id`, to
  `respondent_name`/`email`/`name` (responses) or `name`/`title` (surveys) or `name`/`email`
  (contacts) for `name`, and to `modified_at`/`created_at` for `updated_at`, but ONLY when the
  primary key (`id`, `name`, or `updated_at` respectively) is absent from the raw record. The
  engine's `computed_fields` dialect has no coalesce-across-multiple-alternate-keys filter (only
  a single-reference bare copy, a rename, a join, or a static literal — conventions.md §3); this
  fallback would require either a new engine filter or a RecordHook, neither of which is
  available in Tier 1. Since legacy's own advertised `Catalog()` schema guarantees `id` on every
  emitted record regardless of raw shape, and the live Zonka Feedback API's own documented
  response/survey/contact objects always include a top-level `id` field (this bundle's read of
  the documented wire shape — the alternate keys exist in legacy purely as defensive handling for
  a raw shape that does not occur in practice), this narrowing only diverges from legacy on a raw
  API response that omits `id` entirely, which the documented API surface never produces.
  Documented scope narrowing, not silent divergence.
- **`page_size` is not runtime-configurable.** Legacy exposes a config-driven `page_size`
  override (1-200, default 100, `zonka_feedback.go:243-253`). The engine's `page_number`
  paginator's `PageSize` is a static bundle-level integer (`PaginationSpec.PageSize`), not a
  config-templated field, so there is no mechanism to make it runtime-configurable from
  `config.page_size` without inventing Go. This bundle sends legacy's own default (`per_page=100`)
  as a static pagination value; `page_size` is not declared in `spec.json` (F6, REVIEW.md: a
  declared-but-unwireable config key is worse than an absent one).
- **`max_pages` is not modeled.** Legacy exposes a config-driven `max_pages` override (`0`/`all`/
  `unlimited` meaning unbounded, or a positive integer hard cap, `zonka_feedback.go:255-265`). The
  engine's `PaginationSpec.MaxPages` is a static bundle-level integer, not config-templated, so
  there is no mechanism to make it runtime-configurable. This bundle omits `max_pages` entirely,
  which is unbounded — legacy's own default when unset/`0`/`all`/`unlimited` (the common case) —
  so every input legacy itself defaults to behaves identically; only an operator who explicitly
  set a positive `max_pages` override loses that cap here.
- **Legacy's fixture-mode-only fields are not modeled.** Legacy's `readFixture` path (only
  reached when `config.mode == "fixture"`, a credential-free conformance-harness affordance)
  stamps an additional `previous_cursor` field (echoing `req.State["cursor"]`) onto emitted
  fixture-mode records. This is not part of the live record shape; this bundle's schemas and
  fixtures target the live path only. The engine's own conformance/fixture-replay harness provides
  the credential-free test affordance this bundle needs.
- The connector's own documented `docs_url` page render dynamically and could not be fetched by
  automated tooling during this migration; legacy Go source (`internal/connectors/zonka-feedback/
  zonka_feedback.go`) is the ground truth this bundle was built from, per conventions.md's
  "legacy is ground truth over any doc" rule — this did not block the migration.
