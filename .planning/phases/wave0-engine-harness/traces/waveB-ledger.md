# Wave B TDD ledger — wave0-engine-harness

Executed by: gsd-loop-backend (sonnet), tasks T/B-07 -> T/B-05 -> T/B-06.

## T-07 (hooks registry)

Status: red-confirmed
Timestamp: 2026-07-02T00:00:00Z (session start)

Command: `go test ./internal/connectors/engine -run TestHooks -v`

Output:
```
# polymetrics.ai/internal/connectors/engine [polymetrics.ai/internal/connectors/engine.test]
internal/connectors/engine/hooks_test.go:56:94: undefined: Runtime
internal/connectors/engine/hooks_test.go:61:92: undefined: Runtime
internal/connectors/engine/hooks_test.go:66:77: undefined: Runtime
internal/connectors/engine/hooks_test.go:73:4: undefined: Hooks
internal/connectors/engine/hooks_test.go:74:4: undefined: AuthHook
internal/connectors/engine/hooks_test.go:75:4: undefined: RecordHook
internal/connectors/engine/hooks_test.go:76:4: undefined: StreamHook
internal/connectors/engine/hooks_test.go:77:4: undefined: WriteHook
internal/connectors/engine/hooks_test.go:78:4: undefined: CheckHook
internal/connectors/engine/hooks_test.go:82:21: undefined: unregisterHooks
internal/connectors/engine/hooks_test.go:82:21: too many errors
FAIL	polymetrics.ai/internal/connectors/engine [build failed]
FAIL
```

Test file authored per API-CONTRACT.md §2 (hooks.go) + design §B.7: 5 hook interfaces dispatched
via a single `fakeHooks` fake with compile-time assertions; `RegisterHooks`/`HooksFor` round-trip;
duplicate-name overwrite semantics (last registration wins); unknown name -> nil,nil-safe; factory
invoked fresh per `HooksFor` call (not cached).

Status: green
Timestamp: 2026-07-02T00:05:00Z

Implemented `internal/connectors/engine/hooks.go`: `Runtime` struct (Requester/Bundle/Config),
`Hooks`/`AuthHook`/`RecordHook`/`StreamHook`/`WriteHook`/`CheckHook` interfaces exactly per
API-CONTRACT.md §2 signatures, process-global `hookRegistry` (mutex-guarded map of
`name -> func() Hooks`), `RegisterHooks(name, factory)` (last-registration-wins overwrite,
matching `connectors.RegisterFactory`'s documented overwrite semantics), `HooksFor(name)` (nil
when unregistered; invokes the factory fresh per call), `unregisterHooks(name)` test-cleanup
helper (mirrors `connectors.unregisterFactory`). Also created
`internal/connectors/hooks/hookset/hookset_gen.go` placeholder (generated-file header, empty
import list — no wave0 golden needs a hook).

Command: `go test ./internal/connectors/engine -run 'TestRegisterHooks|TestHooksFor|TestAuthHook|TestRecordHook|TestStreamHook|TestWriteHook|TestCheckHook' -v`

Output: all 9 subtests PASS (TestRegisterHooksAndHooksForRoundTrip,
TestHooksForUnknownReturnsNilSafely, TestRegisterHooksDuplicateNameOverwrites,
TestRegisterHooksFactoryInvokedPerCall, TestAuthHookDispatch, TestRecordHookDispatch,
TestStreamHookDispatch, TestWriteHookDispatch, TestCheckHookDispatch).

`go build ./...`, `go vet ./internal/connectors/...`, `gofmt -l internal/connectors` all clean.

## T-05 (auth selection)

Status: red-confirmed
Timestamp: 2026-07-02T00:10:00Z

Command: `go test ./internal/connectors/engine -run TestSelectAuth -v`

Output:
```
# polymetrics.ai/internal/connectors/engine [polymetrics.ai/internal/connectors/engine.test]
internal/connectors/engine/auth_test.go:30:15: undefined: selectAuth
internal/connectors/engine/auth_test.go:49:15: undefined: selectAuth
internal/connectors/engine/auth_test.go:67:15: undefined: selectAuth
internal/connectors/engine/auth_test.go:86:15: undefined: selectAuth
internal/connectors/engine/auth_test.go:104:15: undefined: selectAuth
internal/connectors/engine/auth_test.go:138:15: undefined: selectAuth
internal/connectors/engine/auth_test.go:178:15: undefined: selectAuth
internal/connectors/engine/auth_test.go:194:12: undefined: selectAuth
internal/connectors/engine/auth_test.go:200:13: undefined: selectAuth
internal/connectors/engine/auth_test.go:218:12: undefined: selectAuth
internal/connectors/engine/auth_test.go:218:12: too many errors
FAIL	polymetrics.ai/internal/connectors/engine [build failed]
FAIL
```

Test file authored per TEST-PLAN.md §1.3 + API-CONTRACT.md `selectAuth(cfg, specs, h) (connsdk.Authenticator, error)`:
all 6 modes (bearer/none/basic/api_key_header/api_key_query/oauth2_client_credentials), `when`
ordering (first match wins, two-always-match-rules case), github-style auto/token/public/github_app
table, custom->AuthHook via a fake Hooks implementing AuthHook, custom with Hooks=nil and with
Hooks-not-implementing-AuthHook (both typed errors), no-match typed error, empty-specs typed error,
secrets-never-leak-into-error assertion (interpolation error message must not contain a planted
secret value from an unrelated config key).

Status: green
Timestamp: 2026-07-02T00:20:00Z

Implemented `internal/connectors/engine/auth.go`: `selectAuth(cfg, specs, h)` iterates specs in
declared order, evaluating `EvalWhen` (empty `when` = always matches), and builds the
`connsdk.Authenticator` for the first match via `buildAuthenticator` — a switch over
`AuthSpec.Mode` mapping onto the existing connsdk constructors (`Bearer`, `Basic`, `APIKeyHeader`,
`APIKeyQuery`, `OAuth2ClientCredentials`) plus a no-op `AuthFunc` for `none`. `custom` resolves an
`AuthHook` via the `Hooks` passed in (`HooksFor` is the caller's job, not selectAuth's — matches
API-CONTRACT.md signature taking `h Hooks` directly); both "no hooks provided" and "hooks present
but doesn't implement AuthHook" are typed errors naming the hook. Empty spec list and no-rule-match
are typed errors. All templated AuthSpec fields (Token/Username/Password/Value/TokenURL/ClientID/
ClientSecret/Scopes) are resolved via `Interpolate` against a `Vars{Config, Secrets}` built only
from `RuntimeConfig.Config`/`.Secrets` — never reading secrets from Config, per THREAT-MODEL §1.

Command: `go test ./internal/connectors/engine -run TestSelectAuth -v`

Output: all 17 test functions/subtests PASS (bearer, none/public, basic, api_key_header w/prefix,
api_key_query, oauth2_client_credentials fetch+cache-within-window, custom hook resolve, custom
missing-hook x2 (nil Hooks / non-AuthHook Hooks), no-rule-match, empty-specs, first-rule-wins
ordering, github-style auto/token/public/github_app 4-case table, secrets-never-leak).

`go build ./...`, `go vet ./internal/connectors/...`, `gofmt -l internal/connectors` all clean;
full `go test ./internal/connectors/engine -v` green (no regressions in schema/interpolate/bundle/
errors/hooks suites).

## T-06 (paginator construction)

Status: red-confirmed
Timestamp: 2026-07-02T00:30:00Z

Command: `go test ./internal/connectors/engine -run TestNewPaginator -v`

Output:
```
# polymetrics.ai/internal/connectors/engine [polymetrics.ai/internal/connectors/engine.test]
internal/connectors/engine/paginate_test.go:78:12: undefined: newPaginator
internal/connectors/engine/paginate_test.go:119:12: undefined: newPaginator
internal/connectors/engine/paginate_test.go:142:12: undefined: newPaginator
internal/connectors/engine/paginate_test.go:184:12: undefined: newPaginator
internal/connectors/engine/paginate_test.go:218:12: undefined: newPaginator
internal/connectors/engine/paginate_test.go:252:12: undefined: newPaginator
internal/connectors/engine/paginate_test.go:287:12: undefined: newPaginator
internal/connectors/engine/paginate_test.go:321:12: undefined: newPaginator
internal/connectors/engine/paginate_test.go:361:12: undefined: newPaginator
internal/connectors/engine/paginate_test.go:382:12: undefined: newPaginator
internal/connectors/engine/paginate_test.go:382:12: too many errors
FAIL	polymetrics.ai/internal/connectors/engine [build failed]
FAIL
```

Test file authored per TEST-PLAN.md §1.4 + PLAN.md T-06 against multi-page `httptest.Server`
fixtures (using `connsdk.Harvest` as the driving loop, matching how read.go will call
paginators): all 6 types (link_header 3-page chain terminating at absent Link header; page_number
short-page stop AND max_pages stop, searxng-shape with no size param sent; offset_limit short-page
stop; cursor(token_path) exhausting to ""; cursor(last_record_field+stop_path) stripe
starting_after/has_more shape incl. empty-page-with-has_more:true defensive stop and
missing-id-field stop; next_url absolute-URL follow, loop guard on repeated URL, SSRF guard
rejecting a cross-host next URL, and the allow_cross_host escape; none = exactly one request) +
malformed-spec table (unknown type, cursor with both token_path+last_record_field, cursor with
neither, next_url missing next_url_path). Every httptest handler asserts each distinct
query-string/path key is hit at most once via a shared hitCounter (no duplicate page fetches).

Status: green
Timestamp: 2026-07-02T00:45:00Z

Implemented `internal/connectors/engine/paginate.go`: `newPaginator(spec, pageSize)` maps
`link_header`/`page_number`/`offset_limit` directly onto the existing connsdk constructors
(`LinkHeaderPaginator`, `PageNumberPaginator` with `StartPage` defaulted to 1, `OffsetPaginator`),
`cursor` dispatches on which of `token_path`/`last_record_field` is set (both or neither ->
typed error) to either `connsdk.CursorPaginator` or the new local `lastRecordCursor`, `next_url`
builds the new local `nextURL` (error when `next_url_path` is empty), `none`/`""` builds a local
`nonePaginator` issuing exactly one request.

Design note surfaced during implementation (not a blocker, resolved locally): `connsdk.Response`
does not expose the resolved request URL (the `requestURL` field is unexported with no accessor),
so a `next_url` paginator constructed from `newPaginator(spec, pageSize)` alone cannot learn the
base URL's host to enforce the THREAT-MODEL §3 same-host guard. Resolved by adding an exported
`BaseHost` field on the unexported `*nextURL` type (mirrors connsdk's own pattern of post-
construction-settable fields, e.g. `OAuth2ClientCredentials.Client`); the wave0 test harness sets
it directly, and read.go (wave C) will set it from `requester.BaseURL`'s parsed host before the
first `Harvest` call. `nextURL.Next` records guard violations (cross-host block, loop detection)
on a sticky `lastErr` surfaced via `Err()`, since `connsdk.Paginator.Next` has no error return —
Harvest itself returns nil in that case (paginator just stops), so callers must check `Err()` to
distinguish a benign stop from a blocked guard. connsdk was NOT modified.

`lastRecordCursor.Next` treats an empty page, a `stop_path` value other than `"true"`, or a last
record missing/null on `last_record_field` as unconditional stops — this is what makes the
"empty page with has_more:true" and "missing id field" cases terminate instead of looping, matching
stripe's real semantics defensively.

Command: `go test ./internal/connectors/engine -run TestNewPaginator -v`

Output: all 21 test functions/subtests PASS covering all 6 pagination types + malformed-spec table
(unknown type, cursor both/neither token source, next_url missing path) + the SSRF guard (blocked
cross-host, `allow_cross_host: true` escape, loop guard).

`go build ./...`, `go vet ./internal/connectors/...`, `gofmt -l internal/connectors` all clean.
Full `go test ./internal/connectors/engine -v`: 76 PASS / 0 FAIL (no regressions across schema/
interpolate/bundle/errors/hooks/auth/paginate suites). `go test ./internal/connectors/engine -race
-run TestNewPaginator`: clean, no data races.
