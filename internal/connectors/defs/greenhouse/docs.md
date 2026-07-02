# Overview

Greenhouse is a Tier-1 declarative-HTTP wave2 fan-out migration. It reads Greenhouse candidates,
applications, jobs, offers, and users through the Greenhouse Harvest REST API. This bundle migrates
`internal/connectors/greenhouse` (the hand-written connector); the legacy package stays registered
and unchanged until wave6's registry flip.

## Auth setup

Provide the Greenhouse Harvest API token as the `api_key` secret. It flows into HTTP Basic auth as
the username with a blank password (`Authorization: Basic base64(api_key:)`, matching legacy's
`connsdk.Basic(secret, "")`) and is never logged. `base_url` defaults to
`https://harvest.greenhouse.io/v1` and can be overridden for tests or proxies.

## Streams notes

All 5 streams (`candidates`, `applications`, `jobs`, `offers`, `users`) are top-level JSON arrays
(`records.path: ""`) using the base's `link_header` pagination (RFC 5988 `Link: <url>; rel="next"`),
sending `per_page` from `config.page_size` (default `100`, matching legacy's
`greenhouseDefaultPageSize`).

No `incremental` block is declared for any stream — legacy never sends a date-filter query
parameter on any Greenhouse endpoint; every sync is a full, unfiltered fetch (legacy's own
`InitialState` cursor is never read back by the live `Read` path). `x-cursor-field` is still
declared on each schema (`updated_at`/`last_activity_at`/`updated_at`/`updated_at`/`updated_at`,
matching legacy's catalog `CursorFields`) as informational catalog metadata only.

## Write actions & risks

None. Greenhouse is a read-only source in this connector (legacy `Capabilities.Write` is `false`);
no `writes.json` file is present.

## Known limits

- Full Greenhouse Harvest API surface (candidate/application create/update, scorecards, scheduled
  interviews, departments, offices, custom fields, activity feed, etc.) is out of scope for wave2;
  see `api_surface.json`'s `excluded: {category: out_of_scope, reason: "Pass B capability
  expansion"}` entries. Only the 5 legacy-parity read streams are implemented.
- Every stream fixture ships a single page rather than the usual 2-page requirement (§4). This is
  the identical harness limitation the `gitlab` bundle documents for its own `link_header` streams: a
  fixture file has no field to declare a `Link:` response header, so a second page can never be
  expressed in a static fixture for `link_header` pagination. `pagination_terminates` still passes (a
  single-page fixture with no `Link` header terminates after exactly one request, the correct,
  honest outcome). No live parity test exists for Greenhouse's 2-page Link-header advance in this
  wave; a future wave could add one (bitly/calendly's `next_url` pattern) if this needs live proof.
- No incremental sync mode is derived for any stream, matching legacy's real unfiltered-every-sync
  behavior.
