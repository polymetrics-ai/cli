# Issue #4072 context: GitHub App shared-rate admission

**Issue:** #4072 under parent #3855

**Base:** `integration/4015-mvp-flat-r1` at `7eea99bae`, which includes the
#3754 rate-admission work merged as #4122.
**Status:** implemented and locally verified on 2026-08-14.

## Locked context

- #3754 admits endpoint-scoped policy at the `connsdk.Requester` physical-send
  boundary, based on the resolved request path. This work preserves that
  boundary; it does not re-introduce declared-path admission.
- A connector with a `require_shared` policy refuses an unresolved route using
  `*coordination.SharedRateLimitUnavailableError`; direct read, GraphQL, and
  binary formatters already preserve that error type.
- GitHub App authentication creates a real
  `POST /app/installations/{installation_id}/access_tokens` request while
  constructing its authenticator. It must use the existing requester rather
  than `http.DefaultClient` so the physical request goes through admission.
- JWT, private-key, and minted-token material remain confined to the hook and
  HTTP request. They are absent from coordination keys, errors, and test
  diagnostics.

## Chosen implementation

The engine creates its resolver and base requester before selecting custom
authentication. An opt-in custom auth hook gets a narrow declared-route JSON
request capability: it supplies its declaration plus escaped physical path,
headers, and JSON body; the engine selects the requester and calls
`Requester.Do`. The GitHub hook is the only current consumer.

## Delivery fallback

The project GSD adapter accepts numeric roadmap phases only, while this issue
uses a named issue phase. The delivery contract also prohibits role spawning in
this lane. Discuss, TDD plan, execution, verification, and review therefore
run inline and are recorded in this directory. Firstmate owns the subsequent
no-mistakes and PR gates.
