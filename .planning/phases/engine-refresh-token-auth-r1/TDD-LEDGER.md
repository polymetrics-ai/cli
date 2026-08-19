# TDD ledger — engine-refresh-token-auth-r1

GSD programming loop, manual/inline fallback (see `PLAN.md`). Every behaviour-adding task starts
**red**: the test is written and run against the unmodified tree first, and the failure recorded
verbatim, before any production edit.

Red evidence is a real `go test` failure, not an assertion that one would occur.

---

## R1 — `connsdk.OAuth2RefreshToken` does not exist

**Test:** `TestOAuth2RefreshTokenFirstExchangeSetsBearer` (`connsdk/refresh_token_test.go`)

Red output: see the run log (Cycle 1).

---

## R2 — access token is not reused before expiry

**Test:** `TestOAuth2RefreshTokenReusesAccessTokenBeforeExpiry`,
`TestOAuth2RefreshTokenRefreshesAtExpiry`, `TestOAuth2RefreshTokenRenewsWithinSafetyMargin`

---

## R3 — a missing `expires_in` is not treated conservatively

**Test:** `TestOAuth2RefreshTokenMissingExpiresInUsesShortFallbackTTL`,
`TestOAuth2RefreshTokenShortLifetimeStillCaches`

---

## R4 — a rotated refresh token is dropped

**Test:** `TestOAuth2RefreshTokenRotatedTokenUsedOnNextExchange`,
`TestOAuth2RefreshTokenRotationCallbackInvoked`,
`TestOAuth2RefreshTokenRotationPersistFailureSurfaces`

---

## R5 — no 401 refresh, and no bound on one

**Test:** `TestRequesterRefreshesAuthOnceOn401AndRetries`,
`TestRequesterRefreshesAtMostOncePerRequestOnPersistent401`,
`TestRequesterRefreshFailureDoesNotRetryAgain`,
`TestRequesterReauthRetriesEvenWithNoRetryBudget`,
`TestRequesterDoesNotRefreshWhenAuthenticatorIsNotARefresher`,
`TestOAuth2RefreshTokenRefreshAuthReExchanges`

---

## R6 — concurrent callers all exchange

**Test:** `TestOAuth2RefreshTokenConcurrentApplyCausesExactlyOneExchange`,
`TestOAuth2RefreshTokenConcurrentRefreshAuthCollapses` (both run under `-race`)

---

## R7 — engine has no `oauth2_refresh_token` mode

**Test:** `TestSelectAuthOAuth2RefreshTokenMode`, `TestSelectAuthOAuth2RefreshTokenWhenGated`,
`TestBuildOAuth2RefreshTokenUnresolvedKeysError`,
`TestBuildOAuth2RefreshTokenReturnsRefresherAuthenticator`,
`TestBundleLoadAcceptsOAuth2RefreshTokenAuthSpec`, `TestBundleLoadStillRejectsUnknownAuthKey`,
`TestExistingAuthModesAreNotRefreshers`, `TestResolveCheckAuthSpecCoversRefreshTokenFields`

---

## R8 — credentials can reach error text

**Test:** `TestOAuth2RefreshTokenErrorsNeverLeakCredentials` — walks every error path (missing token
URL, provider 4xx whose body echoes the secret back, malformed JSON, response missing
`access_token`, transport failure against a closed listener) and asserts the refresh token, client
secret and access token appear in none of them.

---

## R9 — rotation is not persisted anywhere encrypted

**Test:** `TestRuntimeConfigSecretStorePersistsRotatedSecret`,
`TestRuntimeConfigSecretStoreRejectsInvalidKey`,
`TestRuntimeConfigSecretStoreErrorsAreRedacted` (all `internal/app`);
`TestBuildOAuth2RefreshTokenRotationWiredToSecretStore`,
`TestBuildOAuth2RefreshTokenNoRotationWithoutDeclaredKey`,
`TestBuildOAuth2RefreshTokenNoStoreConfiguredStillWorks`,
`TestBuildOAuth2RefreshTokenRejectsInvalidStoreKey` (all `internal/connectors/engine`)

---

## Run log

Every cycle below ran **red first against the unmodified tree**, and the failure output is quoted
verbatim.

### Cycle 1 — connsdk authenticator + 401 hook (R1–R6, R8)

Red (`go test ./internal/connectors/connsdk/`, before `refresh_token.go` existed):

```
internal/connectors/connsdk/reauth_test.go:214:11: undefined: OAuth2RefreshToken
internal/connectors/connsdk/refresh_token_test.go:69:11: undefined: OAuth2RefreshToken
...
FAIL	polymetrics.ai/internal/connectors/connsdk [build failed]
```

Green after `refresh_token.go` + the `doWithBody` reauth branch:

```
ok  	polymetrics.ai/internal/connectors/connsdk	0.930s
```

One genuine red-to-green correction inside the cycle: the `unparseable` sub-case of R3 failed
(`token endpoint returned a body that is not a valid token response`) because `expires_in` was typed
as `json.Number`, so a quoted or non-numeric value failed the WHOLE response decode instead of
falling back. Fixed by decoding it as `json.RawMessage` and parsing leniently — a real bug the test
caught before it could reach a provider.

Race run:

```
$ go test -race -run 'RefreshToken|Requester(Refresh|Reauth|DoesNotRefresh)' ./internal/connectors/connsdk/
ok  	polymetrics.ai/internal/connectors/connsdk	1.389s   (20 tests, all PASS)
```

### Cycle 2 — SecretStore seam (R9, app half)

Red (`go test -run SecretStore ./internal/app/`):

```
internal/app/secret_store_test.go:43:9: cfg.SecretStore undefined (type connectors.RuntimeConfig has no field or method SecretStore)
FAIL	polymetrics.ai/internal/app [build failed]
```

Green after `connectors.SecretStore` + `internal/app/secret_store.go` + the `resolveCredential`
wiring:

```
ok  	polymetrics.ai/internal/app	3.473s
```

### Cycle 3 — engine mode (R7, R9 engine half)

Red (`go test ./internal/connectors/engine/`):

```
internal/connectors/engine/auth_refresh_token_test.go:108:3: unknown field RefreshToken in struct literal of type AuthSpec
internal/connectors/engine/auth_refresh_token_test.go:224:14: undefined: buildOAuth2RefreshToken
internal/connectors/engine/auth_refresh_token_test.go:250:3: unknown field SecretStore in struct literal of type connectors.RuntimeConfig
...
FAIL	polymetrics.ai/internal/connectors/engine [build failed]
```

Green after the `oauth2_refresh_token` case, the two `AuthSpec` fields and the two meta-schema keys:

```
ok  	polymetrics.ai/internal/connectors/engine	13.167s
```

### Cycle 4 — static validation coverage (found during verification, not planned)

Verification surfaced that `ResolveCheckAuthSpec` — the function written to close exactly this class
of gap (F9) — did not cover `refresh_token` or `refresh_token_store_key`, so a typo would pass
`connectorgen validate` and fail only on a real sync.

Red:

```
--- FAIL: TestResolveCheckAuthSpecCoversRefreshTokenFields/typo'd_refresh_token_template_is_rejected
--- FAIL: TestResolveCheckAuthSpecCoversRefreshTokenFields/invalid_refresh_token_store_key_is_rejected_statically
```

Green after adding the field to the ResolveCheck list and an identifier check for the store key.
The first sub-case's premise was wrong and was corrected — see VERIFICATION.md §3.

### Mutation checks

Both the redaction assertion and the static-validation assertion were proven non-vacuous by
deliberately breaking the production code and observing the failure. Output in VERIFICATION.md §3.
