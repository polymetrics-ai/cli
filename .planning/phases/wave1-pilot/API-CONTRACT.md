# API-CONTRACT — wave1-pilot

No exported engine API changes this phase. The contracts pilots consume exist since wave0
(`.planning/phases/wave0-engine-harness/API-CONTRACT.md`); this file pins the exact hook
interfaces the three Tier-2 pilots implement and the shapes they must honor.

## Hook interfaces consumed (defined `internal/connectors/engine/hooks.go` — DO NOT MODIFY)

```go
type AuthHook interface {
    Authenticator(ctx context.Context, cfg connectors.RuntimeConfig, spec AuthSpec) (connsdk.Authenticator, error)
}
type StreamHook interface {
    ReadStream(ctx context.Context, stream StreamSpec, req connectors.ReadRequest, rt *Runtime, emit func(connectors.Record) error) (handled bool, err error)
}
type WriteHook interface {
    ExecuteWrite(ctx context.Context, action WriteAction, rec connectors.Record, rt *Runtime) (handled bool, err error)
}
type CheckHook interface { Check(ctx context.Context, cfg connectors.RuntimeConfig, rt *Runtime) (handled bool, err error) }
type Hooks interface { ConnectorName() string }
```

Registration: `engine.RegisterHooks(name, factory)` from each `hooks/<name>` package `init()`;
discovery via `hooks/hookset/hookset_gen.go` regenerated ONLY by `go run ./cmd/connectorgen gen`
(orchestrator, P-14). Parity tests blank-import their own hooks package directly (SPEC §6).

## github_app AuthHook shape (design doc §B.7 pattern; port of `github/auth.go`)

- Bundle side (`defs/github/streams.json` base.auth candidates, evaluated in order):
  bearer-token candidate(s) `when`-gated on `{{ secrets.token }}` truthiness, then
  `{"mode": "custom", "hook": "github", "when": "..."}` for the github_app path, then/or
  `{"mode": "none"}` for public mode — final order must reproduce legacy's `auto` resolution
  (auth.go:60-80: token wins, then app config, then public).
- Hook side: reads `app_id`/`installation_id` (+ optional scoping keys) from `cfg.Config` and the
  private key from `cfg.Secrets` (`githubAppPrivateKey`/`githubAppPrivateKeyBase64`,
  auth.go:211-218); signs an RS256 JWT (stdlib crypto, exactly like auth.go:154+); POSTs
  `/app/installations/{id}/access_tokens` on the connector base URL with ctx; returns a
  `connsdk.Authenticator` that sets `Authorization: Bearer <installation token>` and re-exchanges
  on expiry. Contract obligations: ctx honored, errors secret-free, no logging.

## gmail AuthHook shape (port of `gmail/auth.go` oauthRefreshAuth)

- Bundle side: `{"mode": "custom", "hook": "gmail", "token_url": "{{ config.token_url }}",
  "client_id": "{{ secrets.client_id }}", "client_secret": "{{ secrets.client_secret }}",
  "token": "{{ secrets.client_refresh_token }}", "scopes": "..."}` — the existing templated
  AuthSpec fields (bundle.go:103-106; all nine validated by `engine.ResolveCheckAuthSpec` via
  connectorgen) carry the wiring; the hook interprets `Token` as the refresh token. Exact field
  mapping is the agent's call, but it MUST flow through AuthSpec templates (validate-time
  checked), not ad-hoc cfg reads, wherever an equivalent field exists.
- Hook side: refresh-token grant POST, cached token with expiry-60s refresh, mutex-guarded,
  injectable clock, https-validated token_url (THREAT-MODEL Delta 2).

## monday StreamHook / CheckHook shape

- `ReadStream` handles ALL declared streams (returns handled=true always; declarative fallback
  unreachable by design — document in docs.md Known limits); uses `rt.Requester` for POSTs so
  auth/retry/UA plumbing stays engine-owned. GraphQL query text: fixed templates with integer
  limit/page only (THREAT-MODEL non-delta). Emits records matching `schemas/<stream>.json`
  projection semantics (hook output is post-projection by contract — mirror legacy shape
  exactly).

## Stability promises

- No pilot changes any engine-exported identifier; `connectorgen validate` LOC/interface-count
  report enforces hook caps (≤300 lines, ≤2 interfaces — monday's StreamHook+CheckHook and
  github's AuthHook+WriteHook are both exactly at 2).
- `Definition().Spec` remains verbatim `RawSpec` bytes (wave0 F5) — spec.json IS the config
  contract for each pilot.
