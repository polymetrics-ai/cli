# Overview

MailerSend is a read-only declarative-HTTP connector that reads the legacy email activity, domains,
messages, and recipients streams plus additional documented MailerSend API collections: templates,
scheduled messages, sender identities, inbound routes, account users, invites, tokens, webhooks, and
the four email analytics views. Requests use the MailerSend REST API
(`https://api.mailersend.com/v1`). This bundle migrates `internal/connectors/mailersend` (the
hand-written legacy connector, which stays registered and unchanged until wave6's registry flip) to
a Tier-1 defs bundle. MailerSend's mutating endpoints send transactional email or alter account
configuration rather than perform safe reverse-ETL upserts, so the connector remains read-only.

## Auth setup

Provide a MailerSend API token via the `api_token` secret; it is used only for Bearer auth
(`Authorization: Bearer <api_token>`) and is never logged.

## Streams notes

Paginated collection streams share MailerSend's `{data:[...]}` envelope and `page_number`
pagination (`page_param: page`, `size_param: limit`, `start_page: 1`, `page_size: 25`) — matching
legacy's `connsdk.PageNumberPaginator{PageParam:"page", SizeParam:"limit", StartPage:1,
PageSize:pageSize}` short-page stop rule exactly for the original streams and matching the
documented page/limit shape for the added collection streams.

- `domains` (`GET /domains`) and `recipients` (`GET /recipients`) take no extra config.
- `messages` (`GET /messages`) optionally filters by `domain_id` when set (`omit_when_absent: true`
  — sent only when `domain_id` is configured, matching legacy's `if domain := ...; domain != ""`
  guard).
- `templates` optionally filters by `domain_id`; `webhooks` requires `domain_id`, matching
  MailerSend's documented webhook list parameters.
- `activity` (`GET /activity/{domain_id}`) requires `domain_id` (templated into the path, urlencoded
  by default like every path segment) and the `date_from`/`date_to` unix-seconds query window
  (both hard-required — an absent value is a hard interpolation error, matching legacy's explicit
  `errors.New("mailersend activity stream requires config domain_id")` /
  `"... requires config date_from and date_to"`).
- `analytics_by_date`, `analytics_country`, `analytics_user_agents`, and
  `analytics_reading_environment` require the same `date_from`/`date_to` unix-seconds window and
  read `data.stats` without pagination.

None of the streams declares an `incremental` block: legacy's `InitialState` always seeds an empty
cursor and `Read` never applies any cursor-based filter (server-side or client-side) to any request.
Declaring an `incremental`/`client_filtered` block here would introduce new record-dropping behavior
legacy never had, so this bundle omits it and keeps `x-cursor-field` values purely as documented
dedup-capable field names.

## Write actions & risks

None. MailerSend's mutating endpoints send transactional email, verify addresses, manage users and
tokens, change sending domains/routes/templates/webhooks, or delete scheduled messages;
`capabilities.write` is `false` and no `writes.json` is shipped, matching legacy's unsupported
`Write`.

## Known limits

- Legacy's `activity` stream accepts an optional `event` config (a CSV string split and sent as
  REPEATED `event[]` query params, one per listed event type) to filter which activity event types
  are returned. The engine's `stream.Query` dialect maps one JSON key to exactly one template-resolved
  string value — it has no repeated-key/array query param mechanism — so this optional filter cannot
  be expressed declaratively. It is out of scope here (a scope-narrowing, not a data-shape
  divergence: omitting the filter returns the full unfiltered activity set rather than a
  event-type-scoped subset, never a WRONG record shape for any record it does emit).
- Legacy's `date_from` also falls back to the `start_date` config alias when `date_from` itself is
  unset (`firstNonEmpty(cfg.Config["date_from"], cfg.Config["start_date"])`). The engine's
  `stream.Query` `default` dialect only supplies a FIXED literal fallback, not a second
  config-key reference, so this bundle requires `date_from` directly and does not implement the
  `start_date` alias for the activity stream specifically (the `start_date` property remains
  declared and documented as that fallback's intent, but is not wired for `activity`).
- Legacy's config-driven `page_size`/`max_pages` overrides have no declarative equivalent (the
  engine's `PaginationSpec.PageSize`/`MaxPages` are fixed values in `streams.json`, not
  runtime-config-driven); pagination is fixed at `limit=25` (legacy's own default) with unbounded
  pages.
- Detail endpoints that require a template, identity, inbound route, message, webhook, token, user,
  invite, domain, recipient, verification, or SMS object ID are excluded as ID-scoped lookups rather
  than global streams.
- SMS and email-verification resources are excluded because this connector's legacy/source contract
  is the Email API/account surface and those resources use separate scopes and operational risk.
