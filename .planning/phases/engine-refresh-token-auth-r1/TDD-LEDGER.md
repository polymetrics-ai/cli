# TDD ledger — engine-refresh-token-auth-r1

GSD programming loop, manual/inline fallback (see `PLAN.md`). Every behaviour-adding task starts
**red**: the test is written and run against the unmodified tree first, and the failure recorded
verbatim, before any production edit.

Red evidence is a real `go test` failure, not an assertion that one would occur.

---

## R1 — `connsdk.OAuth2RefreshToken` does not exist

**Test:** `TestOAuth2RefreshTokenFetchesAndCachesAccessToken` (`connsdk/auth_test.go`)

Red command / output: recorded below in the run log.

---

## R2 — access token is not reused before expiry

**Test:** `TestOAuth2RefreshTokenReusesAccessTokenBeforeExpiry`,
`TestOAuth2RefreshTokenRefreshesAtExpiry`

---

## R3 — a missing `expires_in` is not treated conservatively

**Test:** `TestOAuth2RefreshTokenMissingExpiresInUsesShortFallbackTTL`,
`TestOAuth2RefreshTokenShortLifetimeStillCaches`

---

## R4 — a rotated refresh token is dropped

**Test:** `TestOAuth2RefreshTokenRotatedTokenUsedOnNextExchange`,
`TestOAuth2RefreshTokenRotationCallbackInvoked`

---

## R5 — no 401 refresh, and no bound on one

**Test:** `TestRequesterRefreshesAuthOnceOn401AndRetries`,
`TestRequesterRefreshesAtMostOncePerRequestOnPersistent401`,
`TestRequesterDoesNotRefreshWhenAuthenticatorIsNotARefresher`

---

## R6 — concurrent callers all exchange

**Test:** `TestOAuth2RefreshTokenConcurrentApplyCausesExactlyOneExchange` (run under `-race`)

---

## R7 — engine has no `oauth2_refresh_token` mode

**Test:** `TestSelectAuthOAuth2RefreshTokenMode`,
`TestBuildOAuth2RefreshTokenInterpolatesEveryField`,
`TestBuildOAuth2RefreshTokenUnresolvedKeyErrors`

---

## R8 — credentials can reach error text

**Test:** `TestOAuth2RefreshTokenErrorsNeverLeakCredentials` — walks every error path (missing token
URL, provider 4xx whose body echoes the secret back, malformed JSON, response missing
`access_token`, transport failure against a closed listener) and asserts the refresh token, client
secret and access token appear in none of them.

---

## R9 — rotation is not persisted anywhere encrypted

**Test:** `TestCredentialSecretStorePersistsRotatedSecret` (`internal/app`),
`TestRuntimeConfigSecretStoreWiredIntoRefreshTokenMode` (`internal/connectors/engine`)

---

## Run log

Filled in as each red → green cycle completes.
