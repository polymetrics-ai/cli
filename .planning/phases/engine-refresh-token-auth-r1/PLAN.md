# Phase: engine-refresh-token-auth-r1

**GSD command:** `/gsd-plan-phase engine-refresh-token-auth-r1`, generated through the repo-local Pi
adapter with `scripts/gsd prompt plan-phase engine-refresh-token-auth-r1` (`scripts/gsd doctor`
green, 69 commands registered). The generated prompt is recorded at
`.planning/traces/gsd-plan-phase-engine-refresh-token-auth-r1-prompt.md`.

**Runtime fallback:** the adapter's `/gsd-plan-phase` expects Pi runtime subagents. This session runs
in Claude Code, so the documented inline/manual fallback was taken: the official workflow was
executed inline by the single session agent instead of via spawned Pi researcher/planner/checker
subagents. AGENTS.md permits this and requires it be recorded — it is recorded here. The TDD
lifecycle is **not** waived: every behaviour-adding task starts red, and the red/green evidence is
in `TDD-LEDGER.md`.

**Required skills loaded** (per `.agents/agentic-delivery/references/required-skills-routing.md`,
"Connector runtime and architecture"): `golang-how-to` (orchestrator), `golang-concurrency`,
`golang-security`. Applied concretely: a single mutex covering check-and-exchange so concurrent
streams collapse to one token request (`golang-concurrency`'s "caching expensive computations"
row — deduplicate concurrent calls); `-race` on the concurrency test; secrets never in error text,
generic error messages with details withheld rather than echoed (`golang-security`'s "Returning
detailed errors" and "Logging Security" rows); `crypto/rand`-keyed AES-256-GCM reused from the
existing vault rather than any new crypto.

**Issue tree:** parent #3702 with sub-issues #3703 (mode), #3704 (expiry-aware reuse), #3705
(rotation persisted), #3706 (401 refresh once), #3707 (concurrency). Created before any code, under
`karthik-sivadas` — see `ISSUE-TREE.md` for the identity note and the captain's bounded exception.

## Scope

Foundation phase. One shared auth capability that connector lanes cannot add themselves: the OAuth2
**refresh-token grant**, so a connector whose provider issues short-lived user-context access tokens
can run unattended.

Strictly additive and opt-in. A connector declaring none of this behaves byte-for-byte as it does
today: the new mode is a new `case` in an existing switch, the new `AuthSpec` fields are `omitempty`
and unread by every other mode, and the new 401-refresh hook is an **optional interface** that only
the new authenticator implements — `oauth2_client_credentials`, `bearer`, `basic`, the two api-key
modes, `none` and `custom` are all untouched.

**Paths owned by this phase:**

- `internal/connectors/engine/**` — the mode, the spec fields, the meta-schema.
- `internal/connectors/connsdk/**` — the authenticator and the `Requester` 401 hook.
- `internal/connectors/connectors.go` — the `SecretStore` seam on `RuntimeConfig`.
- `internal/app/**` — the vault-backed implementation of that seam.
- `docs/migration/conventions.md`, `docs/architecture/connector-architecture-v2-design.md`.
- `.planning/**`.

**Explicitly NOT touched:** `internal/connectors/defs/**`. The path-ownership guardrail is live on
`main` and rejects connector bundle edits from a foundation branch. Reddit, and every other
connector with the same need, adopts the mode in its own lane afterwards. `cmd/**` and
`internal/connectors/certify/**` are also out of scope.

## Verified starting state

Every claim below was read directly out of this repo before planning.

| Claim | Evidence |
| --- | --- |
| The engine dispatches exactly seven auth modes | `engine/auth.go:64-108` (`buildAuthenticator` switch) |
| `oauth2_client_credentials` is one of them, and the only OAuth2 one | `engine/auth.go:100-101` |
| There is no refresh-token mode anywhere in the engine or connsdk | `rg 'refresh_token\|RefreshToken'` over `internal/connectors/{engine,connsdk}` returns nothing |
| An unknown mode is a hard error, so a new mode is additive and fail-closed | `engine/auth.go:106-107` |
| `AuthSpec` is the templated field carrier; every field goes through `Interpolate` | `engine/bundle.go:129-159`, `engine/auth.go:111-146` |
| The auth meta-schema is `additionalProperties: false`, so a new field must be declared | `engine/schema/streams.schema.json`, `properties.base.properties.auth.items` |
| `OAuth2ClientCredentials` caches with a 60s pre-expiry margin | `connsdk/auth.go:114-116` |
| …and assumes **3600s** when `expires_in` is absent | `connsdk/auth.go:168-171` |
| …and holds its mutex across the token HTTP call, giving one-exchange-for-many-callers | `connsdk/auth.go:112-113` + `146` |
| `Requester` retries 429/500/502/503/504 — **401 is not retryable and returns immediately** | `connsdk/http.go:135-149`, `585-587` |
| `Requester` applies `Auth` per attempt, so a refreshed credential is picked up on retry | `connsdk/http.go:546-552` |
| `HTTPError.Error()` already runs through `safety.RedactErrorText` | `connsdk/http.go:46-55` |
| `RedactErrorText` strips URL query/fragment and `key=value` secret assignments | `internal/safety/safety.go:50-74` |
| The project's encrypted local credential store is `internal/vault`: AES-256-GCM, `crypto/rand` 32-byte per-project key, `0600` files under `.polymetrics/vault/<id>.enc`, id-bound AAD | `internal/vault/vault.go:22-78`, `140-142` |
| Credential secrets already reach a run through `vault.Get` → `RuntimeConfig.Secrets` | `internal/app/app.go:1110-1124` |
| `RuntimeConfig` has `ProjectDir`/`Config`/`Secrets` and **no** write-back seam | `internal/connectors/connectors.go:78-83` |
| `CredentialMeta.SecretFields` records secret **key names** (never values) in plain state | `internal/app/types.go:16-25`, `app.go:205-212` |
| Reddit's spec states token acquisition/refresh is out of scope, against 1-hour tokens | `internal/connectors/defs/reddit/spec.json`, `access_token.description` |
| `golang.org/x/sync` is an **indirect** dependency — using `singleflight` would promote it | `go.mod:60` |
| Go toolchain is 1.25.4 | `go.mod:3` |
| GSD evidence gate fires whenever `cmd/` or `internal/` changes | `scripts/verify-gsd-workflow` |
| Website CI triggers only on `website/**`, `internal/connectors/icon_data.json`, `docs/connectors/icons/**` — none of which this phase touches | `.github/workflows/website.yml:4-9` |

### Correction to the brief

The brief cites "81 moderator operations" on Reddit's ledger. That figure is **not** verifiable from
this repo: `defs/reddit/api_surface.json` on `main` declares 9 endpoints scoped to two legacy-parity
read streams, and the pushed lane branch `origin/fm/cli-reddit-connector-defects-r1` carries the
same 9. The figure is recorded here as a **lane report**, not as a repo-verified count. It does not
change the phase: the verifiable driver — a spec that documents refresh as out of scope, against a
provider whose tokens expire in an hour — is sufficient on its own, and the capability is not
Reddit-specific either way.

## Design decisions

### D1 — Mode name: `oauth2_refresh_token`

Named for the mechanism (the RFC 6749 §6 grant type), not for a provider, matching every existing
mode name and the brief's explicit instruction.

### D2 — Field shape follows `oauth2_client_credentials` exactly

`token_url`, `client_id`, `client_secret`, `scopes`, `extra_params` are the **same** `AuthSpec`
fields the client-credentials mode already uses — no parallel vocabulary. Only two fields are new:

- `refresh_token` — templated, normally `{{ secrets.refresh_token }}`.
- `refresh_token_store_key` — a plain secret **key name** (not a template, not a value) naming where
  a rotated refresh token is persisted.

Every templated field resolves through the existing `Interpolate`, so an unresolved `config.*` or
`secrets.*` key is a hard error exactly as it is for `client_id`/`client_secret` today.

### D3 — Rotation persists to the existing vault, and the key is declared not guessed

`refresh_token_store_key` is explicit because the alternative is worse. The engine cannot recover a
key name from a resolved template (`Interpolate` returns the *value*), and pattern-matching
`{{ secrets.X }}` back to `X` would be magic that silently overwrites a caller's secret whenever the
guess is right and silently does nothing whenever it is not. Declaring the key means:

- a connector whose provider does **not** rotate omits it and nothing is ever written;
- a connector whose provider **does** rotate names the same key its `refresh_token` reads from, so
  the next run picks the rotated value up through the ordinary `Secrets` path with no new machinery;
- the engine never writes to a secret it was not explicitly told to write to.

The write goes through a new `connectors.SecretStore` seam on `RuntimeConfig`, implemented in
`internal/app` over `internal/vault`. The engine never learns what a vault is. A nil store (test
harnesses, the conformance runner, any caller with no credential store) degrades to in-memory
rotation for the process lifetime — never to a plaintext file. **No new storage path is invented and
nothing leaves the machine.**

### D4 — Concurrency: one mutex across check-and-exchange, not `singleflight`

`x/sync/singleflight` would deduplicate too, but it is an indirect dependency today and the mutex
approach is what `OAuth2ClientCredentials` already does three files away. Holding a mutex across
I/O is normally an anti-pattern; here it *is* the feature — a goroutine that arrives mid-exchange
must block and then observe the fresh token, which is precisely "one refresh, shared result". The
critical section is one bounded HTTP call against a client with a 30s timeout, and it is on the
credential path, not a hot path. Recorded as a deliberate deviation, not an oversight.

### D5 — 401 refresh via an optional interface, bounded before the attempt

`connsdk.AuthRefresher` is a new optional interface. `Requester.doWithBody` calls it on a 401 at most
once per request, and the guard flag is set **before** the refresh is attempted — so a refresh that
itself errors cannot yield a second attempt. The reauth retry does not consume the transient-failure
budget (otherwise a `MaxRetries: 0` requester would refresh and then never retry), and because the
flag makes it strictly once-per-request, the loop still terminates.

Authenticators that do not implement the interface are unaffected: `oauth2_client_credentials` and
every other mode keep their exact current behaviour on a 401.

### D6 — Missing `expires_in` means five minutes, not an hour

`OAuth2ClientCredentials` assumes 3600s. For a *user-context* token that guess is unsafe — it is the
exact interval at which Reddit's tokens die. This mode assumes 5 minutes and re-exchanges. The
safety margin is additionally clamped to half the token lifetime, so a provider handing out a
lifetime shorter than the margin still caches instead of exchanging on literally every request.

### D7 — No raw request escape hatch

The only request this phase can emit is a fixed `POST` of a form-encoded
`grant_type=refresh_token` body to the declared `token_url`. Method, path and body structure are all
literals in the implementation; nothing in the spec dialect can change them.

## Task 1 — `connsdk.OAuth2RefreshToken` (issues #3703, #3704, #3707)

`internal/connectors/connsdk/auth.go`.

New authenticator with the client-credentials field shape plus `RefreshToken`,
`OnRefreshTokenRotated`, `ExpirySafetyMargin` and `FallbackTTL`. `Apply` sets
`Authorization: Bearer <token>`; the cached token is reused until `renewAt`.

- `renewAt = expiresAt - min(margin, ttl/2)`, so caching is never zero-length.
- `expires_in` absent / zero / negative / unparseable → `FallbackTTL` (default 300s).
- A `refresh_token` in the token response replaces the in-memory one **and** is handed to
  `OnRefreshTokenRotated` while the lock is held, so the persisted value can never be older than the
  in-memory one.
- All state under one `sync.Mutex`; `exchangeLocked` asserts the lock by construction (unexported,
  only called from locked paths).
- Errors: no URL, no response body, no credential — `oauth2 refresh: token endpoint returned 401
  Unauthorized` and nothing more. Transport errors are wrapped through `safety.RedactErrorText`
  because `url.Error` carries the request URL.

## Task 2 — 401 refresh hook (issue #3706)

`internal/connectors/connsdk/auth.go` (interface) and `connsdk/http.go` (`doWithBody`).

```go
type AuthRefresher interface {
    RefreshAuth(ctx context.Context, req *http.Request) error
}
```

`OAuth2RefreshToken.RefreshAuth` reads the bearer credential off the failed request. If the cached
token has already moved on, another in-flight request refreshed it first and this one returns
immediately — its retry picks up the fresher token with no second exchange. Otherwise it invalidates
and exchanges under the same lock.

## Task 3 — engine mode (issue #3703)

`internal/connectors/engine/auth.go`, `bundle.go`, `schema/streams.schema.json`.

`case "oauth2_refresh_token": return buildOAuth2RefreshToken(cfg, spec, vars)`. Interpolates every
templated field, resolves `extra_params` through the existing shared helper, validates
`refresh_token_store_key` as a secret key name, and wires `OnRefreshTokenRotated` to `cfg.SecretStore`
when both are present.

## Task 4 — vault-backed `SecretStore` (issue #3705)

`internal/connectors/connectors.go` (interface + `RuntimeConfig` field) and `internal/app/app.go`
(implementation). The implementation re-reads the credential's secret bundle, merges the one rotated
key, re-encrypts, and keeps `CredentialMeta.SecretFields`/`UpdatedAt` consistent. It carries its own
mutex: rotation can in principle be driven from a connector goroutine, and `App` has none.

## Task 5 — docs

`docs/migration/conventions.md` (the authoring recipe: how a bundle declares the mode) and
`docs/architecture/connector-architecture-v2-design.md` (the mode list). No CLI surface changes, so
no `pm help`/manual/website parity work is triggered — verified against
`.agents/agentic-delivery/references/cli-help-docs-website-parity.md`.

## Risks

| Risk | Mitigation |
| --- | --- |
| Mutex held across a network call could stall many streams behind one slow token endpoint | Bounded by the token client's 30s timeout; this is the same trade-off `OAuth2ClientCredentials` already makes, and it is what makes one-exchange-for-many true (D4) |
| A refresh loop hammering a permanently-401 provider | Guard flag set before the attempt, once per request; proven by counting token-endpoint hits in a test, not by reading the code (D5) |
| Rotated token lost on a store failure | The access token obtained in the same exchange is still returned and used; the store error is surfaced, not swallowed, so the run fails loudly rather than silently drifting |
| Overwriting a caller's secret with a bad rotated value | Only ever writes the key the bundle explicitly declares; omitting the key means nothing is written |
| Secrets leaking into error text | Dedicated test walks every error path, including a token endpoint that echoes the secret back in a 4xx body |
