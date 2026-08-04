# Verification — engine-refresh-token-auth-r1

**Verified by executing, not by reading.** Everything below is a recorded command output, not a
claim about what the code should do.

## 1. End-to-end harness against live local servers and a real encrypted vault

A scratch harness (`verifyharness/main.go`, deleted after the run — it was never committed) wrote a
**real connector bundle on disk** declaring `mode: "oauth2_refresh_token"`, loaded it through the
**real** `engine.Load`, and drove `engine.Connector.Read` against **live** `httptest` token and API
servers, plus a **real** `app.InitProject` project with a real `internal/vault`.

`go run ./verifyharness`:

```
[1] first exchange, and [2] reuse before expiry across a paginated read
  PASS  first exchange happened once — token calls=1
  PASS  read walked 8 authenticated pages — records=8 upstream requests=8
  PASS  one access token reused for every page — distinct upstream credentials: [Bearer AT-1]
  PASS  grant presented was the configured refresh token — grants=[rt-original]

[3] refresh at expiry, mid-read, on the real wall clock (expires_in=1)
  PASS  read completed across the expiry boundary — records=4
  PASS  token was renewed mid-read — token calls=2 (want >=2)
  PASS  later pages carried the NEW access token — upstream credentials=[Bearer AT-1 Bearer AT-1 Bearer AT-2 Bearer AT-2]

[4] refresh on 401 (out-of-band revocation) and retry
  PASS  401 recovered by one refresh — err=<nil> records=1
  PASS  exactly two upstream attempts, second with the new token — upstream credentials=[Bearer AT-1 Bearer AT-2]
  PASS  exactly two token exchanges — token calls=2

[4b] a provider that ALWAYS 401s terminates instead of hammering
  PASS  terminated — elapsed=1ms err type=*engine.Error
  PASS  exactly one reauth, then stop — token calls=2 api hits=2 (want 2 and 2)
  PASS  no credential in the surfaced error — checked error text for client_secret and refresh_token

[5] rotated refresh token persisted to the real encrypted vault, then used on the NEXT run
  PASS  resolved credential carries a SecretStore — RuntimeConfig.SecretStore is non-nil
  PASS  run 1 presented the original grant — grants=[rt-original]
  PASS  nothing stored in plaintext — files containing a secret verbatim: []
  PASS  next run loaded the ROTATED grant from the vault — refresh_token loaded = "rt-rotated-1"
  PASS  sibling secret untouched — client_secret preserved
  PASS  next run PRESENTED the rotated grant to the provider — grants across both runs=[rt-original rt-rotated-1]

[6] concurrent callers sharing one authenticator cause exactly one exchange
  PASS  exactly one exchange for 12 concurrent callers — token calls=1
  PASS  all 12 requests succeeded — no caller errored
  PASS  every caller used the shared token — 12 upstream requests, distinct credentials [Bearer AT-shared]

[7] authenticator lifetime is per engine Read (measured, not assumed)
  NOTE  3 sequential Reads of a 3-page stream = 9 upstream requests, 3 token exchanges
  PASS  one exchange per Read, not one per request — token calls=3 api hits=9

VERIFICATION PASSED: every check green
```

Every scenario the brief named is covered: first exchange (1), reuse before expiry (2), refresh at
expiry (3), refresh on 401 (4/4b), a rotated refresh token persisted and used on the next run (5),
and concurrent callers causing exactly one exchange (6).

### Measured finding — authenticator lifetime is per `Read`, not per sync

Scenario 7 measures something the brief assumed rather than stated: `engine.Read` calls `newRuntime`
(`read.go:48`), which calls `selectAuth`, so **each `Read` builds its own authenticator**. Measured:
3 Reads of a 3-page stream = 9 upstream requests and **3** token exchanges.

What that means in practice:

- Within one stream read, one exchange serves every page — the dominant win. A 500-page stream costs
  one exchange, not 500.
- Across streams, a sync costs one exchange per stream, not one in total.
- The authenticator itself **is** concurrency-safe and does collapse concurrent callers to one
  exchange (scenario 6, plus `-race` unit tests) — that guarantee lands wherever a caller shares one.

Making the *engine* share one authenticator across streams would need a cache keyed on credential
identity. That is a security-sensitive change — a wrong cache key is credential confusion between
two connectors — and it is deliberately **not** smuggled into this phase. Filed as
[#3708](https://github.com/polymetrics-ai/cli/issues/3708) rather than implemented here.

## 2. Unit tests, including `-race`

```
$ go test -race -run 'RefreshToken|Requester(Refresh|Reauth|DoesNotRefresh)' ./internal/connectors/connsdk/
ok  polymetrics.ai/internal/connectors/connsdk  1.389s
```

All 20 new connsdk tests pass under the race detector, including
`TestOAuth2RefreshTokenConcurrentApplyCausesExactlyOneExchange` (16 goroutines, exactly one
exchange) and `TestOAuth2RefreshTokenConcurrentRefreshAuthCollapses`.

## 3. Mutation checks — the assertions actually bite

Two tests were deliberately broken to prove they are not vacuous.

**Redaction.** Reinstating the token endpoint's error body in the 4xx error:

```
--- FAIL: TestOAuth2RefreshTokenErrorsNeverLeakCredentials/provider_4xx_echoing_every_credential_back_in_the_body
    error text leaked a credential ("rt-SUPERSECRET-refresh"): oauth2 refresh: token endpoint
    returned 400: {"error":"invalid_grant","error_description":"refresh_token=rt-SUPERSECRET-refresh
    client_secret=cs-SUPERSECRET-client access_token=at-SUPERSECRET-access"}
```

Restored → `ok`. The redaction test catches a real leak.

**Static validation.** Removing `{"refresh_token", spec.RefreshToken}` from
`ResolveCheckAuthSpec`'s field list:

```
--- FAIL: .../rejected_refresh_token_template:_unknown_filter
--- FAIL: .../rejected_refresh_token_template:_unknown_namespace
```

Restored → `ok`.

### Correction made during verification

The first version of the static-validation test asserted that a typo'd `{{ secrets.refersh_token }}`
would be rejected. It is not, and should not be: `checkNamespaceRef`
(`interpolate.go:785`) deliberately does not statically check the `secrets` namespace against
`specKeys` — `client_secret` behaves identically. The test was corrected to probe what genuinely is
checkable (an undeclared `config.*` key, an unknown filter, an unknown namespace), which still proves
the field reaches `ResolveCheck`. The production change stands; the test's premise was wrong.

## 4. Gates run locally

| Gate | Command | Result |
| --- | --- | --- |
| Format | `gofmt -l cmd internal` | clean (no output) |
| Vet | `go vet ./...` | clean |
| Build | `go build ./cmd/pm` | ok |
| Tidy | `make tidy-check` | clean — no dependency added (`x/sync` stays indirect) |
| Lint | `make lint` | `0 issues.` (two `errcheck` findings in new test helpers found and fixed) |
| Connector boundary | `go run ./cmd/connectorgen boundary . --json` | `outcome: clean`, 0 findings, 0 warnings, 550 connectors |
| connectorgen validate | via boundary + `go test ./cmd/connectorgen/` | ok |
| Docs | `make docs-check` | `Validated connector docs in docs/connectors` |
| Smoke | `make smoke-no-build` | `smoke ok` |
| Release workflow | `make release-workflow-check` | `homebrew release notification assertions passed` |
| Engine tests | `go test ./internal/connectors/engine/` | ok (10.6s) |
| connsdk tests | `go test ./internal/connectors/connsdk/` | ok |
| App tests | `go test ./internal/app/` | ok (28.7s) |
| connectors tests | `go test ./internal/connectors/` | ok |
| connectorgen tests | `go test ./cmd/connectorgen/` | ok (7.4s) |
| CLI tests | `go test ./internal/cli/` | ok (390s) |

`make verify` was **not** run as one command: it includes `go test ./...` over 550+ connectors,
which the brief flags as exceeding a single command's ceiling. Its constituent gates were run
individually above and CI carries the full suite.

## 5. Binary regression checks

```
$ ./pm connectors list --json | jq '.connectors | length'
554
$ ./pm connectors inspect reddit --json | head -5
{ "api_version": "polymetrics.ai/v1", "connector": { "name": "reddit", "display_name": "Reddit", ...
$ ./pm connectors                       # bare namespace renders help, exits 0
NAME
  pm connectors - inspect connector definitions, streams, and write actions
```

All 554 connectors still load. No CLI surface changed, so no `pm help`/manual/website parity work is
triggered (checked against `.agents/agentic-delivery/references/cli-help-docs-website-parity.md`).

## 6. Website catalog

```
$ cd website && npm run gen:catalog
$ git status --porcelain website/
(no output)
```

The generated catalog is byte-identical — expected, since no connector bundle was touched. Website
CI only triggers on `website/**`, `internal/connectors/icon_data.json` and `docs/connectors/icons/**`
(`.github/workflows/website.yml:4-9`), none of which this branch modifies.

## 7. Constraint compliance

| Constraint | Evidence |
| --- | --- |
| Credentials never in argv, logs, or error text | `TestOAuth2RefreshTokenErrorsNeverLeakCredentials` walks 6 error paths; mutation-checked (§3). No credential is ever passed as a process argument — the exchange is an in-process HTTP POST body. |
| Persisted state local and encrypted | Rotation writes only through `internal/vault` (AES-256-GCM, `0600`, `.polymetrics/vault`). `assertVaultCiphertext` walks the **entire** `.polymetrics` tree asserting the value appears nowhere in the clear; harness scenario 5 repeats it. No new storage path. |
| No raw request escape hatch | The only request added is a literal `POST` of a form-encoded `grant_type=refresh_token` body to the declared `token_url`. Method, path and body structure are Go literals; no spec field can alter them. |
| Engine only — `defs/**` untouched | `git diff --name-only origin/main...HEAD` lists no `internal/connectors/defs/**` path; boundary gate `clean`. |
| Strictly additive and opt-in | `TestExistingAuthModesAreNotRefreshers` asserts no pre-existing mode implements `connsdk.AuthRefresher`, so none changes behaviour on a 401; `TestRequesterDoesNotRefreshWhenAuthenticatorIsNotARefresher` proves the `Requester` path is unchanged for them; `TestBundleLoadStillRejectsUnknownAuthKey` proves the meta-schema widened by exactly two keys. |
