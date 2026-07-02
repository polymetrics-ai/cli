# TEST-PLAN — wave0-engine-harness

All tests table-driven (`golang-testing`); HTTP behavior via `httptest.Server`; no network, no
credentials, no new deps. Test tasks are committed RED first (see `PLAN.md` pairing).

## 1. Engine unit tests (`internal/connectors/engine`)

### 1.1 Interpolation (`interpolate_test.go`, T-02)
| Case | Input | Expect |
|---|---|---|
| config/secrets/record/cursor resolution | `{{ config.base_url }}`, `{{ secrets.token }}`, `{{ record.user.login }}`, `{{ cursor }}` | resolved values; nested dotted path works |
| missing key | `{{ config.nope }}` | error naming key + namespace |
| path-segment injection | config value `a/../b` in `/repos/{{ config.repository }}` | urlencoded (`a%2F..%2Fb`), no traversal |
| query metachars | value `a?x=1&y=2` in a path | encoded |
| space + unicode | `a b`, `héllo` | percent-encoded |
| double-encode guard | value `%2e%2e` | encoded literally (`%252e%252e`) |
| explicit filters | `unix_seconds` on RFC3339, `base64` | correct output; bad input errors |
| when-grammar | `==`, `in ['auto','token']`, truthiness, unknown op | matches design §B.3 only; unknown op = compile error |
| ResolveCheck | template with unknown spec key | validation finding (consumed by connectorgen) |

### 1.2 Schema validator (`schema_test.go`, T-01)
- Keyword matrix: `type` (scalar + array w/ null), `required`, `properties`, `items`, `enum`,
  `pattern`, `minProperties`, `additionalProperties:false`.
- Annotations ignored-but-preserved: `format`, `default`, `title`, `description`.
- `x-secret` partition (`SecretKeys()`), `x-primary-key`, `x-cursor-field` accessors.
- Unknown keyword → compile error. Invalid instance → error naming path + rule.
- Negative corpus reused by bundle-validation tests.

### 1.3 Auth mode selection matrix (`auth_test.go`, T-05)
| Spec | cfg | Expect |
|---|---|---|
| bearer, when auth_type in [auto,token] | auth_type=auto, token set | `connsdk.Bearer` header on request |
| none, when auth_type==public | public | no auth header |
| basic | user+pass | Basic header |
| api_key_header (+prefix) | value | header set |
| api_key_query | value | query param set |
| oauth2_client_credentials | token endpoint (httptest) | token fetched, cached, refreshed |
| custom hook | fake AuthHook | hook authenticator applied |
| custom hook missing | — | typed error |
| no rule matches | — | typed error naming auth_type |
| ordering | two matching rules | first declared wins |

### 1.4 Pagination termination (`paginate_test.go`, T-06)
All 6 types against multi-page fixtures; each asserts: bounded page count, exact page order, each
page fetched exactly once, terminal condition honored:
- `link_header`: rel="next" chain of 3, stops at absent header.
- `page_number`: short-page stop (searxng shape, no size param); `max_pages` stop.
- `offset_limit`: short-page stop.
- `cursor(token_path)`: token exhausts to "".
- `cursor(last_record_field + stop_path)`: stripe `starting_after`/`has_more=false`; empty page
  with `has_more=true` (defensive stop — no infinite loop); missing id field stop.
- `next_url`: absolute URL followed; loop guard (same URL twice → error).
- `none`: exactly one request.

### 1.5 Read path (`read_test.go`, T-08)
- Cursor advance/resume: 2-page incremental fixture; post-read state cursor == max record cursor;
  re-read with state sends `request_param` (`since`/`created[gte]`) formatted per each
  `param_format` (`rfc3339`, `unix_seconds`, `date`, `github_date_range`).
- Projection vs passthrough: undeclared field dropped in `schema` mode, kept in `passthrough`.
- `computed_fields` nested extraction incl. missing intermediate (→ null, not panic).
- Filters: `field_absent` (issue-vs-PR), `field_equals`.
- `single_object` + records path `"."`.
- `client_filtered` incremental drops below-cursor records.
- Header omission when interpolated value empty.
- error_map: 401 → hint text surfaces; 403 + match_body "rate limit" → class `rate_limited`.
- Rate limit: injected sleeper called between requests per `requests_per_minute`.
- Hooks: RecordHook mutate/drop, StreamHook handled=true bypass, CheckHook.
- Limit/ctx: `connectors.LimitEmitter` interplay, ctx cancel mid-page.

### 1.6 Write path (`write_test.go`, T-09)
- Body construction: json default (record minus path_fields), `form`
  (url-encoded ordering-stable), `none`; `body_fields` subset for delete-with-body.
- `record_schema` invalid → error with record index; valid passes.
- DryRun preview: resolved method/path, secret absent from warnings.
- `missing_ok_status`: 404 on idempotent delete = written; 404 without listing = failed;
  non-listed status = failed.
- Accounting parity with legacy semantics (`stripe/write.go:66`): fail-fast, `RecordsFailed`
  = remainder.
- WriteHook handled bypass.

## 2. Bundle validation negative tests (T-03, T-11)

Seeded-invalid bundle corpus (each = one defect class, used by loader tests AND
`connectorgen validate` self-tests; classes enumerated in `EVAL-PLAN.md` §3):
missing metadata.json · spec not valid draft-07 subset · unresolvable `{{ config.x }}` ·
`schema` ref to missing file · `x-primary-key` field absent from schema ·
`incremental.cursor_field` absent from schema · write `path_fields` not in record_schema ·
api_surface endpoint with both/neither of covered_by/excluded · stream not in api_surface ·
bad connector name (`Source-GitHub`) · secret literal planted in fixture · docs.md missing
required heading. Expect: `connectorgen validate` exit≠0 with finding naming file+rule;
good-bundle control passes.

## 3. Conformance v2 self-tests (T-13)

- Each static check: one targeted failing corpus bundle + passing control.
- Dynamic: `pagination_terminates` (2-page fixture, each served exactly once — replay server
  counts hits); `records_match_schema` fails on seeded type drift; `cursor_advances` +
  re-read request assertion; `write_request_shape` mismatch vs `expect` fails; `delete_semantics`.
- `TestConformance` over `defs.FS`: subtest per bundle; empty defs tree passes (pre-Wave-F).

## 4. Certify harness self-test (T-12, T-14)

- Harness: kind/exit assertions, envelope parse failure path, secret scan (exact, base64,
  URL-encoded planted values), argv redaction.
- Source stages vs `sample` end-to-end in ephemeral root (flags mirrored from `Makefile:41` smoke):
  stages 0–11 green; 5 sync-mode matrix rows with `data_source` live(local)/capture; capture
  replays via built-in `file` connector; dedup verified via `pm query` (no duplicate PK tuples);
  resume: run2 records ≤ run1, cursor monotonic; sabotage test → `passed=false`, failing stage
  named; report file written + history appended; workdir cleanup.

## 5. Golden parity tests (T-15/16/17)

| Golden | Assertions |
|---|---|
| stripe | Per stream (5): identical record slices engine vs legacy against one httptest server (order-sensitive, json.Number-normalized); 2-page customers pagination; incremental `created[gte]` unix-seconds propagation from state and from `start_date`; write create/update: captured method+path+form body identical; manifest surface equal (streams, PKs, cursor fields, action names) vs `connectors.ManifestOf(stripe.New())` |
| searxng | Both streams: identical records; `q` templating incl. `site:reddit.com` scoping; `pageno` pagination stop at max_pages=1; no size param sent; optional bearer applied when token present; manifest surface equal |
| postgres | Fixture-mode Check/Catalog/Read equality vs legacy; config-validation error table (missing host/db/user, bad port, bad sslmode, host-with-scheme SSRF guard per `postgres.go:119`); `Definition()` from bundle; guard test: no `RegisterFactory` call in `native/postgres` |

Plus regression: `registryset` still registers legacy stripe/searxng/postgres (registry list
contains all three with legacy metadata).

## 6. Lint gate

`golangci-lint run` clean over the repo with the new `.golangci.yml` (generated files excluded);
`gofmt -l cmd internal` empty; `go vet ./...` clean. Wired into `make verify` (T-18).

## 7. Coverage

`go test -cover ./internal/connectors/engine` ≥ 85% statements (phase exit metric, EVAL-PLAN §1).
