# Overview

MailerLite is a read-only declarative-HTTP connector that reads subscribers, campaigns, groups,
segments, and automations through the MailerLite v2 REST API (`https://connect.mailerlite.com/api`).
This bundle migrates `internal/connectors/mailerlite` (the hand-written legacy connector, which
stays registered and unchanged until wave6's registry flip) to a Tier-1 defs bundle at capability
parity.

## Auth setup

Provide a MailerLite API token via the `api_token` secret; it is used only for Bearer auth
(`Authorization: Bearer <api_token>`) and is never logged.

## Streams notes

All 5 streams share MailerLite's `{data:[...], meta:{next_cursor:...}}` envelope and cursor
pagination (`pagination.type: cursor`, `cursor_param: cursor`, `token_path: meta.next_cursor`, no
`stop_path` declared) — matching legacy's own stop rule exactly: legacy stops when
`next_cursor` is null/empty OR the cursor fails to advance OR a page returns zero records; the
engine's token-path cursor paginator's built-in empty-token stop and same-token loop guard
reproduce the first two conditions, and a zero-record page naturally yields no further pagination
progress. Every request sends `limit=100`, matching legacy's default `page_size`.

None of the 5 streams declares an `incremental` block: legacy declares `CursorFields` on its
`Stream` catalog entries (`updated_at`/`created_at`, informational for downstream dedup) but its
`harvest` loop never applies any cursor filter — server-side or client-side — to any request; every
MailerLite read is a full sync regardless of prior state. Declaring `incremental.client_filtered:
true` here would introduce NEW record-dropping behavior legacy never had (the meta-rule's exact
"changes emitted-record-data for an accepted input" test), so this bundle instead keeps each
schema's `x-cursor-field` (documenting the dedup-capable field per stream) without an enforced
`incremental` block — `full_refresh_append`/`full_refresh_overwrite`(`_deduped`) sync modes apply,
`incremental_append` does not, matching legacy's real behavior precisely.

## Write actions & risks

None. MailerLite is wired read-only (`Capabilities.Write: false` in legacy); `capabilities.write` is
`false` and no `writes.json` is shipped.

## Known limits

- Legacy's config-driven `page_size`/`max_pages` overrides have no declarative equivalent: the
  engine's `PaginationSpec.PageSize`/`MaxPages` are fixed values in `streams.json`'s
  `base.pagination` block, not runtime-config-driven (same dead-config class as searxng's wave0
  finding). Neither is declared in `spec.json`; pagination is fixed at `limit=100` (legacy's own
  default) with unbounded pages (legacy's own default `max_pages` behavior).
- Only the 5 legacy-parity read streams are implemented; the broader MailerLite API surface
  (single-subscriber CRUD, webhooks, forms, fields, campaign content) is out of scope until Pass B —
  see `api_surface.json`.
