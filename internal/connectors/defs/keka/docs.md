# Overview

Keka is a read-only declarative-HTTP migration of `internal/connectors/keka` (the legacy
hand-written connector, which stays registered and unchanged until wave6's registry flip). It
reads Keka HRIS employees, attendance, leave types, leave requests, PSA clients, and projects
through the Keka REST API. Keka upstream supports full_refresh only; there is no write capability
and no `writes.json`.

## Auth setup

Keka authenticates with a custom OAuth2 client-credentials-shaped token exchange: `grant_type`
(default `kekaapi`, NOT the standard `client_credentials`), `scope` (default `kekaapi`),
`client_id`, `client_secret` (secret), and an `api_key` form field, POSTed to `token_url` (defaults
to `https://login.keka.com/connect/token`). The engine's declarative
`auth[].mode: oauth2_client_credentials` cannot express this: `connsdk.OAuth2ClientCredentials`
hard-codes `grant_type=client_credentials` in its token request form and has no override, and the
`extra_params` dialect only ADDS form fields (`form.Add`) rather than replacing an
already-`form.Set` field — declaring `extra_params: {"grant_type": "kekaapi"}` would send BOTH
`grant_type` values on the wire, not override the hard-coded one. This is a genuine `ENGINE_GAP`,
not a Tier-1-expressible shape (see `docs/migration/conventions.md` §1's escape-hatch decision
tree: token-exchange auth with a non-standard grant is a named legitimate Tier-2 `AuthHook`
trigger). `streams.json` declares `auth: [{"mode": "custom", "hook": "keka"}]`;
`internal/connectors/hooks/keka/hooks.go` (`AuthHook`, ~228 lines, one hook interface, well under
the ~300-line soft target) ports legacy's `kekaTokenAuth` exactly: same form fields, same 60s-early
token refresh, same 3600s fallback TTL when the token response omits `expires_in`. `base_url` is
company-specific (e.g. `https://<company>.keka.com/api/v1`) with no global default, matching
legacy's `kekaBaseURL` (which hard-errors on an unset `base_url`).

`api_key` is marked `x-secret` in `spec.json` even though it flows into a token-exchange form field
rather than a Bearer credential itself — the marker is about the field's credential-shaped nature
(conventions.md §2), not whether this bundle wires it as a header. `client_secret` is `x-secret`
and never logged; both flow only into the hook's token-exchange request body.

## Streams notes

All 6 streams (`employees`, `attendance`, `leave_types`, `leave_requests`, `clients`, `projects`)
share the identical shape: `GET` against the Keka list endpoint (`/hris/employees`,
`/time/attendance`, `/time/leavetypes`, `/time/leaverequests`, `/psa/clients`, `/psa/projects`),
records at the top-level `data` array, primary key `["id"]`. Pagination is `page_number`
(`page_param: pageNumber`, `size_param: pageSize`, `start_page: 1`, `page_size: 100`) — matches
legacy's `harvest` loop and its default `kekaDefaultPageSize` (100), stopping on a short/empty page.

Legacy's `harvest` treats the response body's `totalPages` field as authoritative when present,
falling back to the short-page check only when `totalPages` is absent/zero. The engine's
`page_number` paginator has no equivalent — it always stops purely on a short/empty page count,
never consulting a response-body field. This is an ACCEPTABLE parity deviation (never diverges for
any legacy-accepted input): the two stop conditions agree on every page shape Keka's real API
actually returns (a `totalPages`-driven stop and a short-page stop coincide whenever the last page
is genuinely short, which is the overwhelmingly common case); they would only disagree in the edge
case of a stream whose total record count is an exact multiple of 100, costing one extra
(harmless, empty) request — never a difference in which records are emitted. This mirrors the
already-accepted `aha` deviation for the identical `total_pages`-vs-short-page-stop class (see
`conventions.md`'s parity-deviation ledger).

No stream declares an `incremental` block: legacy's `kekaStreams()` catalog publishes no
`CursorFields` for any stream (Keka's public API exposes only full-refresh sync), so per
`conventions.md` §8's incremental truth table, this bundle correctly declares no `incremental`
block and no `x-cursor-field` anywhere — every sync is a full re-scan, matching legacy exactly.

`page_size` (default 100, legacy's `kekaDefaultPageSize`, max 200) and `max_pages` (default
0/unlimited) are **no longer configurable at runtime** — the same documented, deliberate
config-surface narrowing as `katana`'s identical finding (`page_number` pagination's `page_size`/
`max_pages` fields are load-time JSON literals with no `config.*`-driven override mechanism). The
bundle's fixed `page_size: 100` reproduces legacy's own default exactly; `max_pages` is omitted
(unbounded), matching legacy's default.

## Write actions & risks

None. Keka is read-only upstream (`capabilities.write: false`); there is no `writes.json`.

## Known limits

- **Conformance dynamic checks are marked `skip_dynamic` at the bundle level** (`metadata.json`'s
  `conformance` marker) — this bundle's sole auth candidate is `mode: custom` with no
  when-gated non-custom fallback, and the `AuthHook`'s real token-exchange target is
  `config.token_url`, a config key wholly separate from `base_url` (unlike github's hook, whose
  installation-token endpoint IS relative to `base_url`, which conformance's harness redirects to
  its own replay server). Conformance's synthetic non-secret config sets `token_url` to the literal
  string `"synthetic-conformance-value"`, never a real or replay-server URL, so the hook's real
  HTTP POST can never land anywhere meaningful — every auth-resolving dynamic check
  (`check_fixture`, every `read_fixture_nonempty:<stream>`, `pagination_terminates`,
  `records_match_schema`) would otherwise fail identically and uninformatively. This is the exact
  scenario `conventions.md` §4 documents as the sanctioned skip-marker case (gmail's identical
  shape). The skipped behavior is proven live instead by
  `internal/connectors/paritytest/keka/parity_test.go`, which drives both the legacy connector and
  the engine-backed connector (with the real registered `AuthHook`) against a shared `httptest`
  token-exchange server and a shared `httptest` data server, asserting RAW `connectors.Record`
  equality and identical `grant_type=kekaapi`/`api_key` token-request form fields.
- **`totalPages`-based early stop is approximated by short-page stop only** — see Streams notes
  above; the only observable difference is one extra empty-page request when a stream's total
  record count is an exact multiple of 100, never a difference in which records are emitted.
- **`page_size`/`max_pages` runtime overrides are not modeled** — see Streams notes above; the
  bundle's fixed defaults match legacy's own defaults exactly.
- Full Keka API surface (payroll, performance, recruitment, expenses, webhooks) is out of scope
  for this migration; see `api_surface.json`'s `excluded: {category: out_of_scope, reason: "Pass B
  capability expansion"}` entries. Only the 6 legacy-parity read streams are implemented.
