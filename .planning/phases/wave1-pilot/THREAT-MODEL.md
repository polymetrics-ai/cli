# THREAT-MODEL — wave1-pilot (delta from wave0)

Baseline: `.planning/phases/wave0-engine-harness/THREAT-MODEL.md` remains authoritative for the
engine surface (interpolation/CRLF guards §2, pagination SSRF guards §3, secret redaction,
fixture secret scanning). This phase adds NO new engine attack surface except the two hook
packages below; bundles are data and inherit wave0's guards. Carried hardening notes N5
(`..%5C`/`x%2F..%2Fy` decode-before-route residuals) and N2 (digit-shaped non-unix start values)
remain open, unchanged in risk (no pilot feeds untrusted values into those shapes).

## Delta 1 — github_app AuthHook secret handling (`hooks/github/`)

- Assets: GitHub App RSA private key (`githubAppPrivateKey`/`githubAppPrivateKeyBase64` secrets),
  short-lived installation tokens.
- Requirements (parity with legacy `github/auth.go`, which never logs these): private key only
  ever parsed into `*rsa.PrivateKey` and used for JWT signing; JWT + installation token only flow
  into the Authorization header of requests to the configured base URL; token-exchange endpoint
  derived from the connector base URL (no attacker-controllable exchange host); errors wrap
  status text only, never the key/JWT/token (engine errors additionally pass
  `safety.RedactErrorText`); ctx-threaded exchange (wave0 F8) so cancellation can't leave a
  dangling request. Fixtures/docs contain no key material (`secret_literal` scan enforces).
- Review hook: P-11 line-by-line covers 100% of this file; hooks_test asserts no secret text in
  any error path.

## Delta 2 — gmail OAuth refresh AuthHook (`hooks/gmail/`) — KEPT per SPEC §5.7

- Assets: `client_secret`, `client_refresh_token` (long-lived!), cached access tokens.
- Requirements (parity with legacy `gmail/auth.go`): secrets flow ONLY into the token-request form
  and Authorization header; token_url default is Google's endpoint, config override must be an
  https URL validated like legacy's `validatedURL` (gmail.go:339) — prevents exfiltrating the
  refresh token to an attacker-chosen token endpoint (this is the ONE new SSRF-adjacent surface
  this phase; the hook must fail closed on non-https or unparseable token_url); cached token
  protected by mutex (race-tested); no secret in errors.
- Token acquisition/storage (3-legged consent) stays in the credentials layer — out of connector
  scope, unchanged from legacy's trust model.

## Non-deltas (explicitly considered, no change)

- monday StreamHook: sends the same secret header connsdk already sends; GraphQL query text is
  built from config values already covered by wave0 interpolation guards? — NO: the hook builds
  queries in Go, so the hook MUST NOT string-interpolate untrusted config into query text without
  the same validation legacy applies (legacy uses fixed query templates with integer
  limit/page — keep it that way; reviewer checklist item).
- Write risk (github): reverse-ETL writes remain gated by the app's plan-approval flow
  (github/manifest.go Approval note); DryRun previews redact `x-secret` fields (wave0 write.go
  behavior, unchanged).
- No registration flip: no pilot code is reachable from production paths until wave6; blast
  radius of a defective bundle/hook this phase is test-only.
- Parity/test servers: httptest on loopback only; no live API calls in CI (docs URLs are fetched
  only by the P-11 reviewer, read-only).
