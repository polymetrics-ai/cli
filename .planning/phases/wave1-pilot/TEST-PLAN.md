# TEST-PLAN — wave1-pilot

Test strategy per task; parity is THE acceptance mechanism (wave0 golden pattern:
`internal/connectors/engine/parity_stripe_test.go`, `parity_searxng_test.go` — both connectors
driven LIVE against the same httptest server; RAW `reflect.DeepEqual`; legacy-side "test fixture
bug" sanity fatals so a dead comparison can't pass silently).

## 1. Parity matrix (per connector; ✓ = required test, — = n/a with reason)

Legacy write survey (verified): `grep -l write.go internal/connectors/*` → only
amazon-seller-partner, dropbox-sign, hubspot, stripe have write.go; among pilots ONLY github has
real writes (implemented inline: `github.go:236 Write`, 16 actions `github.go:1759+`). All other
pilots return `ErrUnsupportedOperation` (e.g. xkcd.go:108-109, gmail.go:191-192) → no write parity
tests, `capabilities.write: false`, no writes.json.

| Connector | Streams parity | Pagination parity | Incremental parity | Auth parity | Writes parity | Error-map parity |
|---|---|---|---|---|---|---|
| xkcd | ✓ latest + comic (single_object) | — (none) | — (none) | — (no auth) | — (no writes) | ✓ non-2xx; hostile comic_number fail-closed both sides |
| vitally | ✓ all legacy streams | ✓ per legacy | per legacy (read code) | ✓ byte-exact Authorization header | — | ✓ |
| bitly | ✓ all legacy streams | ✓ next_url 2-page (absolute URLs; exhaustion on empty `pagination.next`) | — (legacy: full refresh, bitly.go:36) | ✓ Bearer | — | ✓ |
| calendly | ✓ all legacy streams (records at `collection`) | ✓ next_url 2-page via pagination.next_page; null next_page terminates | ✓ start_date-raised lower bound | ✓ Bearer | — | ✓ |
| sentry | ✓ all legacy streams (top-level array records) | ✓ link_header 2-page WITH `results="false"` final page — termination + request-count assertion (SPEC §5.3 ladder outcome recorded) | per legacy cursor handling | ✓ Bearer | — | ✓ |
| chargebee | ✓ all legacy streams (envelope-unwrapped fields, field-for-field) | ✓ cursor offset/next_offset 2-page | ✓ updated_at unix_seconds + app-persisted digit cursor round-trip | ✓ Basic (key as username, empty password) | — | ✓ |
| zendesk-support | ✓ all legacy streams | ✓ cursor page[after]/meta.after_cursor 2-page; `stop_path: meta.has_more` stop (incl. has_more:false-with-non-null-cursor regression) | N/A — legacy sends no incremental filter (gap-loop cycle-1 fix; original "start_date-raised" row was a planning error — no such wire behavior exists to parity-test) | ✓ BOTH modes: Basic `<email>/token` AND OAuth Bearer (when-gated candidates) | — | ✓ |
| monday | ✓ boards/users/teams/tags/items via StreamHook | ✓ GraphQL in-body page loop (short page stop) + items next_items_page cursor, 2 pages each | ✓ updated_at cursor fields per legacy | ✓ raw-token Authorization + API-Version header | — | ✓ |
| github | ✓ all 19 streams (streams.go:12-30) incl. flattened computed fields + `repository` stamp + issues PR-filter | ✓ page_number per_page/page 2-page short-page stop | ✓ `since` param forwarding (github.go:91-92,152-162) + cursor round-trip on updated_at streams | ✓ token bearer AND github_app JWT→installation-token (httptest token endpoint; AuthHook) | ✓ **only pilot with writes**: parity floor create_issue, update_issue, comment_issue, create_pull_request (compound follow-ups asserted as separate requests), merge_pull_request, delete_label (missing_ok 404 semantics); fail-fast accounting; DryRun redaction | ✓ |
| gmail | ✓ messages/threads/drafts/labels (streams.go:35-62) | ✓ pageToken/nextPageToken cursor 2-page | — (legacy publishes no cursor fields, streams.go:31-34) | ✓ AuthHook: refresh-token exchange against httptest token endpoint; token caching (2nd request reuses token); expiry refresh (injectable clock) | — | ✓ incl. token-endpoint failure surfaces as auth error, not silent unauth request |

Cross-cutting parity assertions (every connector): request path/query/header shape captured
server-side and compared legacy-vs-engine; record ORDER equality (not just set equality) where
legacy guarantees order; no secret value ever appears in test server logs or error strings
(reuse wave0's parity redaction assertions).

## 2. Conformance auto-coverage

`TestConformance/<name>` (10 static + 8 dynamic checks, `internal/connectors/conformance/`) runs
automatically once `defs/<name>/` exists — each agent's self-verify runs it; P-14 runs the full
matrix. Notably exercised by pilots: `pagination_terminates` (2-page fixtures REQUIRED —
conventions §4), `cursor_advances` (numeric cursors for chargebee/github; `github_date_range`
mirror fixed in P-0), `check_fixture`, `write_validate`/`write_request_shape` (github only),
`secret_redaction` + `secret_literal` fixture scans, `interpolations_resolve`
(`ResolveCheck`/`ResolveCheckAuthSpec` over every template incl. hook AuthSpec fields),
`docs_heading`/`docs_present`, `primary_key_missing`/`cursor_field_missing`.

## 3. Hook unit tests (Tier-2 pilots; live in `hooks/<name>/hooks_test.go`)

- gmail: refresh grant form encoding (`grant_type=refresh_token`, optional client_secret/scope
  omission), cache-hit path, 60s-early refresh, non-2xx token endpoint → error, missing
  refresh_token/client_id → error naming the key, ctx cancellation honored (wave0 F8), zero
  secret text in errors.
- github: JWT claims/exp shape (fixed clock), installation-token request path
  `/app/installations/<id>/access_tokens`, private_key vs private_key_base64 secret resolution
  (auth.go:211-218 parity), scoping payload (repositories/ids/permissions), WriteHook
  `handled=false` fallback for simple actions, compound create_pull_request request sequence.
- monday: StreamHook page loop termination (short page), items cursor loop (next_items_page null
  stop), max_pages cap, GraphQL query text golden-matched to legacy's, `handled=true` for every
  declared stream.

## 4. Fixture rules (restating the non-negotiables, conventions §4)

- **Real wire shapes** — chargebee `created_at`/`updated_at` and github numeric ids/timestamps as
  the API actually sends them (numbers as bare JSON numbers; B2 lesson: NEVER stringify a cursor
  to appease tooling — the tooling was fixed).
- 2-page fixtures for every paginated stream; page-2 request must carry the expected
  cursor/offset/page param (stripe golden shape).
- Sanitized synthetic values; nothing secret-shaped (`secret_literal` scanner is a hard fail);
  gmail/github fixtures carry NO tokens (auth is engine/hook-side, fixtures never carry auth).
- Legacy fixture-mode artifacts (bitly/xkcd `fixture: true` fields) must NOT leak into bundle
  fixtures — bundles model the LIVE wire shape only.

## 5. Red-first evidence protocol

- P-0: RED transcript (failing `cursor_advances` on the new self-test bundle) pasted into
  `TDD-LEDGER.md` before the fix commit; GREEN transcript after. Wave0 ledger format
  (`.planning/phases/wave0-engine-harness/TDD-LEDGER.md` → traces/) reused.
- P-1..P-10: each agent records (a) the initial `go test ./internal/connectors/paritytest/<name>`
  RED failure (bundle-not-found), (b) at least one MEANINGFUL red→green transition (a parity
  assertion that failed against a wrong/incomplete bundle — not only the trivial load failure),
  (c) final green self-verify block output. Goes into the agent's trace file
  `.planning/phases/wave1-pilot/traces/executor-<name>.md`; P-14 aggregates into TDD-GATE.json
  task rows (the wave0 B3 lesson: the machine-readable gate must contain real rows, not empty
  arrays).
- Repairs (post-review): repair ledger entries per wave0's `traces/gaploop-r*-ledger.md` format.

## 6. Suites & commands

| Layer | Command | When |
|---|---|---|
| Agent self-check | conventions §7 block + `go test ./internal/connectors/paritytest/<name> -v` | end of each P-1..P-10 |
| Engine/conformance regression | `go test ./internal/connectors/engine ./internal/connectors/conformance` | P-0, P-14 |
| Full parity | `go test ./internal/connectors/paritytest/... ./internal/connectors/engine -run 'TestParity\|Test.*Parity'` (wave0 goldens included) | P-14 |
| Full conformance | `go test ./internal/connectors/conformance -run TestConformance` | P-14 |
| Whole repo | `make verify` (build, tests, lint, connectorgen-validate, smoke) | P-0, P-14, and any repair |
| Race | `go test -race ./internal/connectors/hooks/... ./internal/connectors/paritytest/...` (gmail token cache is mutex-guarded — prove it) | P-14 |

Coverage: no new numeric gate this phase (engine's ≥85% gate stays; hooks packages target ≥80%
line coverage informally — reviewer flags, doesn't block).
