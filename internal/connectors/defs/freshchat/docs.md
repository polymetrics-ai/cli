# Overview

Freshchat is a customer-messaging product. This bundle reads Freshchat agents, users, groups,
channels, and roles through the Freshchat v2 REST API. It is read-only, matching legacy
`internal/connectors/freshchat` exactly (`Capabilities{Write: false}`).

## Auth setup

Provide a Freshchat API key via the `api_key` secret; it is sent as `Authorization: Bearer
<api_key>` and is never logged.

## Streams notes

All 5 streams (`agents`, `users`, `groups`, `channels`, `roles`) share the same shape: `GET`
against the Freshchat list endpoint, records at the top-level wrapper key matching the resource
name (`agents`, `users`, `groups`, `channels`, `roles`), primary key `["id"]`. `agents`, `users`,
and `channels` declare `x-cursor-field: updated_time` (matching legacy's `CursorFields:
["updated_time"]`); `groups` and `roles` have no cursor field, matching legacy (which declares no
`CursorFields` for those two streams). Pagination is `page_number` (`page`/`items_per_page`,
`start_page: 1`, `page_size: 50`), stopping on a short page — identical to legacy's
`connsdk.PageNumberPaginator` usage (legacy builds the exact same paginator inline).

## Write actions & risks

None. Freshchat is exposed read-only, matching legacy's `Capabilities{Write: false}` and its
`Write` method, which always returns `connectors.ErrUnsupportedOperation`.

## Known limits

- Full Freshchat API surface (conversations, messages, canned responses, business hours, webhooks,
  etc.) is out of scope for this wave; see `api_surface.json`'s `excluded: {category: out_of_scope,
  reason: "Pass B capability expansion"}` entries. Only the 5 legacy-parity read streams are
  implemented.
- **`base_url` is required config, not derived from `account_name`.** Legacy derives the base URL
  from an `account_name` config value when `base_url` is unset
  (`https://<account_name>.freshchat.com/v2`), validating that `account_name` is a bare subdomain
  with no `/:. ` characters. The engine's `spec.json` `"default"` materialization mechanism only
  supports a FIXED literal default value, not one derived from another config value at read/check
  time (see `docs/migration/conventions.md`'s `spec.json "default"` section) — there is no
  declarative way to express "derive base_url from account_name" without inventing Go for a single
  string-templating rule, which would be new undeclared logic outside the dialect. This bundle
  therefore requires the fully-formed `base_url` directly (e.g.
  `https://acme.freshchat.com/v2`), matching the same accepted precedent as chargebee and
  repairshopr in this repo. This is a documented config-surface narrowing (never a data-parity
  change): every record this bundle can read is byte-identical to what legacy would emit for the
  same account once the equivalent full URL is supplied.
