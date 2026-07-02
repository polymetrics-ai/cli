# Overview

GitLab is a Tier-1 declarative-HTTP wave2 fan-out migration. It reads GitLab projects, groups,
users, and issues through the GitLab REST API v4. This bundle migrates
`internal/connectors/gitlab` (the hand-written connector); the legacy package stays registered and
unchanged until wave6's registry flip.

## Auth setup

Provide a GitLab personal access token or OAuth access token via the `access_token` secret; it is
used only for Bearer auth (`Authorization: Bearer <access_token>`) and is never logged. `base_url`
defaults to `https://gitlab.com/api/v4` and can be overridden for self-managed instances, tests, or
proxies.

## Streams notes

All 4 streams (`projects`, `groups`, `users`, `issues`) are top-level JSON arrays (`records.path:
""`) using the base's `link_header` pagination (RFC 5988 `Link: <url>; rel="next"`, GitLab's own
`page`/`per_page` convention), sending `per_page=50` (matches legacy's default `page_size`).

- `projects` sends `last_activity_after` when `start_date` is configured (via the optional-query
  object form, `omit_when_absent: true` — absent entirely on a run with no `start_date`).
- `groups` has no matching since-filter upstream (legacy's own `gitlabSinceParam["groups"]` is
  empty), so no conditional query param is declared for it.
- `users` sends `created_after` when `start_date` is configured.
- `issues` sends `updated_after` when `start_date` is configured, and renames the raw nested
  `author.id` object to a flat `author_id` via `computed_fields` (legacy's `gitlabAuthorID` helper).

**`start_date` is a static per-request filter, not a stateful incremental cursor** — this is a
deliberate, legacy-preserving choice, not an oversight. Legacy's live `Read` path applies the
configured `start_date` config value verbatim to every single request; it never reads
`req.State["cursor"]` to compute a lower bound on repeat syncs (that state-read only appears in
legacy's `readFixture` helper, unreachable outside fixture mode). Declaring the engine's proper
`incremental` block (`cursor_field`/`request_param`/`start_config_key`) would compute the lower
bound as "persisted state cursor, falling back to start_date" — genuinely stateful cursor
advancement across syncs that legacy's live implementation does not perform. Doing so would emit
FEWER records on a second-or-later sync than legacy would (legacy re-sends the exact same
`start_date` filter every time), which the meta-rule (`docs/migration/conventions.md` §5) forbids as
an accepted-input-behavior change. Instead, `start_date` is wired as a plain `config.*`-templated
optional query param on each stream's `query` block, reproducing legacy's exact "always the
configured value, never advances" behavior. `x-cursor-field` is still declared on each schema
(`last_activity_at`/`created_at`/`created_at`/`updated_at`, matching legacy's catalog
`CursorFields`) as informational catalog metadata; no `streams.json` `incremental` block accompanies
it, so no `incremental_append` sync mode is derived for these streams (by design — see
`conventions.md` §2's sync-mode-derivation rule).

`base_url`'s self-managed derivation is narrowed: legacy accepts either a fully-formed `base_url`
override OR a bare `api_url` host/URL that it auto-normalizes and suffixes with `/api/v4` if not
already present. The engine's `spec.json` `"default"` mechanism only materializes a FIXED literal
default, not one derived from another config value at read time, and there is no templating
primitive for conditional suffix-append across a 3-way branch (full override / bare-host
convenience / default). This bundle therefore only accepts a fully-formed `base_url` (e.g.
`https://gitlab.example.com/api/v4` for self-managed) — the bare-`api_url`-host convenience is
dropped as documented config-surface narrowing (never a data-parity change; once `base_url` is
correctly configured, every stream behaves identically).

## Write actions & risks

None. GitLab is a read-only source in this connector (legacy `Capabilities.Write` is `false`); no
`writes.json` file is present.

## Known limits

- Full GitLab API surface (merge requests, pipelines, commits, branches, webhooks, project/group
  members, snippets, etc.) is out of scope for wave2; see `api_surface.json`'s `excluded:
  {category: out_of_scope, reason: "Pass B capability expansion"}` entries. Only the 4
  legacy-parity read streams are implemented.
- Every stream fixture ships a single page rather than the usual 2-page requirement (§4). This is
  the same harness limitation as the sanctioned `next_url` exception, applied to `link_header`: a
  fixture file has no field to declare a `Link:` response header, so a second page can never be
  expressed in a static fixture for `link_header` pagination (the freshdesk bundle in this repo
  ships the identical single-page-only shape for the same reason). `pagination_terminates` still
  passes (a single-page fixture with no `Link` header terminates after exactly one request, which
  is the correct, honest outcome — not a corner cut). No live parity test exists for GitLab's
  2-page Link-header advance in this wave; a future wave could add one (bitly/calendly's
  `next_url` pattern) if GitLab's pagination behavior needs to be proven live.
- The bare-`api_url`-host self-managed convenience is dropped (see Streams notes above);
  `base_url` must be the fully-formed API root.
- `start_date` is a static since-filter re-applied on every sync, matching legacy's real (not
  fully-stateful) behavior — see Streams notes above for why the engine's stateful `incremental`
  mechanism was deliberately not used here.
