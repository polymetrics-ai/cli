# Overview

HoorayHR reads users (employees), time-off requests, leave types, and sick-leave records through
the HoorayHR REST API (`https://api.hooray.nl`). This bundle is a full capability-parity migration
of the legacy hand-written connector (`internal/connectors/hoorayhr`), which stays registered and
unchanged until wave6's registry flip. Read-only: legacy itself sets `Capabilities.Write = false`
with no reverse-ETL write path.

## Auth setup

HoorayHR uses a custom session-token exchange, not a declarative auth mode: the connector POSTs
`{"email": <hoorayhrusername>, "password": <hoorayhrpassword>, "strategy": "local"}` to
`/authentication`, reads `accessToken` back from the JSON response, and injects that token **raw**
(no `Bearer ` prefix — matching legacy's `sessionTokenAuth.Apply`,
`internal/connectors/hoorayhr/hoorayhr.go:242-289`, which sets
`req.Header.Set("Authorization", token)` verbatim) into the `Authorization` header of every
subsequent request. This is a genuine `AUTH_COMPLEX` shape the declarative `auth` dialect cannot
express (none of `bearer`/`basic`/`api_key_header`/`api_key_query`/`oauth2_client_credentials`
prefixes a token-exchange response's field with nothing, and none performs a token-exchange
round-trip against a fixed path with a non-standard JSON body at all) — it is implemented as a
Tier-2 `AuthHook` (`internal/connectors/hooks/hoorayhr/hooks.go`), following the same shape as
`hooks/keka` and `hooks/github`'s token-exchange hooks (conventions.md Section 1's Tier-2 table:
"token-exchange auth" is a named legitimate trigger). The token is fetched once and cached for the
lifetime of the authenticator (mirrors legacy's `sync.Mutex`-guarded `cached` field exactly — no
expiry/refresh logic, since legacy has none either).

Provide the HoorayHR account email via the `hoorayhrusername` config value and the account
password via the `hoorayhrpassword` secret (never logged; used only in the `/authentication`
request body). `base_url` defaults to `https://api.hooray.nl` and may be overridden for
tests/proxies (validated as an absolute http/https URL with a host).

## Streams notes

All 4 streams are single-request, unpaginated top-level JSON arrays (`pagination.type: none`,
`records.path: ""`), exactly matching legacy's own read path (`Connector.Read`,
`hoorayhr.go:100-140`): one `GET`, decoded via `connsdk.RecordsAt(resp.Body, "")`, no next-page
concept at all.

- `users` (`GET /users`) — HoorayHR employees; primary key `id`.
- `time_off` (`GET /time-off`) — leave/time-off requests; primary key `id`.
- `leave_types` (`GET /leave-types`) — company-configured leave types; primary key `id`.
- `sick_leaves` (`GET /sick-leave`) — sick-leave records; catalog stream name is `sick_leaves`
  (snake_case, per conventions.md Section 2's naming rule — legacy's own catalog uses the
  hyphenated `sick-leaves`), underlying API path `/sick-leave` (matching legacy's
  `hoorayhrStreamEndpoints["sick-leaves"].resource`).

Every published field name matches the raw HoorayHR wire shape verbatim (`id`, `email`,
`firstName`, `lastName`, `status`, `jobTitle`, `companyId`, `isAdmin`, `companyStartDate`,
`userId`, `leaveTypeId`, `timeOffType`, `leaveUnit`, `start`, `end`, `notes`, `name`, `icon`,
`color`, `budget`, `default`, `unpaidLeave`, `leaveInDays`, `percentage`, `actualStart`,
`actualReturn`, `reportedStart`, `reportedReturn`, `createdAt`, `updatedAt`) — no
`computed_fields` renames are needed, mirroring legacy's own record mappers
(`hoorayhrUserRecord`/`hoorayhrTimeOffRecord`/`hoorayhrLeaveTypeRecord`/`hoorayhrSickLeaveRecord`),
which all copy fields through under their original HoorayHR key names.

None of the 4 resources exposes a legacy-recognized incremental cursor field — matching legacy's
own catalog, which publishes no `CursorFields` for any stream (every legacy `connectors.Stream`
entry in `hoorayhr/streams.go` omits `CursorFields` entirely). All 4 streams are full-refresh only.

## Write actions & risks

None. `capabilities.write` is `false`; no `writes.json` is shipped. Legacy itself implements no
write path for HoorayHR (`Write` is a stub returning `ErrUnsupportedOperation`).

## Known limits

- **Stream names are snake_case, not legacy's hyphenated catalog names**: legacy publishes
  `time-off`/`leave-types`/`sick-leaves` (`internal/connectors/hoorayhr/streams.go`); this bundle
  publishes `time_off`/`leave_types`/`sick_leaves` instead, since the engine's `namePattern`
  (`^[a-z][a-z0-9_]*$`, conventions.md Section 2) rejects hyphens in stream names. The underlying
  HTTP paths (`/time-off`, `/leave-types`, `/sick-leave`) are unchanged; only the catalog stream
  identifier differs. This is a naming-surface change only, never a data-shape change.
- **Conformance dynamic checks are marked `skip_dynamic` at the bundle level** (`metadata.json`'s
  `conformance` marker) — this bundle's sole auth candidate is `mode: custom` with no
  when-gated non-custom fallback, and the `AuthHook`'s real session-token exchange POSTs live
  `hoorayhrusername`/`hoorayhrpassword` values to a fixed `/authentication` path. Conformance's
  synthetic non-secret config populates every non-`x-secret` spec property (e.g.
  `hoorayhrusername`) with the literal string `"synthetic-conformance-value"` and provides no live
  server for that POST to land on, so the hook's real HTTP round-trip can never succeed against the
  replay harness — every auth-resolving dynamic check (`check_fixture`, every
  `read_fixture_nonempty:<stream>`, `records_match_schema`) would otherwise fail identically and
  uninformatively. This is the exact scenario conventions.md Section 4 documents as the sanctioned
  bundle-level skip-marker case (gmail's/keka's identical shape). The skipped behavior is proven
  live instead by `internal/connectors/paritytest/hoorayhr/parity_test.go`, which drives both the
  legacy connector and the engine-backed connector (with the real registered `AuthHook`) against a
  shared `httptest` authentication+data server, asserting RAW `connectors.Record` equality and the
  identical raw (non-`Bearer`-prefixed) `Authorization` header value on every data request.
- `pagination_terminates` is consequently not exercised dynamically for this bundle either (every
  stream is unpaginated in any case, so there is nothing for it to prove beyond what
  `records_match_schema`/the parity suite already cover).
- Full HoorayHR API surface (payroll, documents, company settings) is out of scope for this
  migration; see `api_surface.json`'s single `non_data_endpoint` exclusion for `/authentication`
  itself. Only the 4 legacy-parity read streams are implemented, matching legacy's own scope
  exactly.
