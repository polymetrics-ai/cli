# Research — GitHub live-operation proof and rate policy

**Date:** 2026-08-08  
**Mode:** Inline/manual GSD research fallback. The local adapter is healthy, but this session has
no compatible isolated GSD worker and is not authorised to delegate. Research was performed against
the checked-out code and GitHub's primary documentation before planning.

## Findings

### Existing runtime is the only limiter to use

- `internal/connectors/engine/rate_limit_runtime.go` creates the resolver only for a declared
  bundle policy; `Runtime.RequesterFor` attaches `Admit` before and `Observe` after every
  request. Endpoint-specific requesters serve checks, ETL, direct reads, direct writes and binary
  downloads; whole-connector/auth-type selectors also attach to hook requesters.
- `internal/connectors/connsdk/rate_limits.go` already provides `selector.auth_types`, scoped
  non-secret registry keys, fixed/sliding windows and request/point budgets. A new limiter would
  bypass this single choke point and is explicitly out of scope.
- `internal/connectors/defs/defs.go` does not currently embed `*/rate_limits.json`; the first
  production declaration must add it alongside the GitHub declaration.
- The deferred declaration branch's GitHub file is only an honest `unknown` placeholder. Its claim
  that active auth cannot be selected is no longer sufficient: auth types are declarative selector
  input, as the schema and runtime show.

### Provider policy to cite

- GitHub documents 60 unauthenticated REST requests/hour by originating IP and 5,000
  authenticated-user REST requests/hour. GitHub App installation access tokens have a minimum of
  5,000 requests/hour. `GITHUB_TOKEN` has a distinct per-repository limit. Source:
  `https://docs.github.com/en/rest/using-the-rest-api/rate-limits-for-the-rest-api`
  (retrieved 2026-08-08).
- GitHub sends `x-ratelimit-limit`, `x-ratelimit-remaining`, `x-ratelimit-used`,
  `x-ratelimit-reset`, and `x-ratelimit-resource` on responses. Primary exhaustion reports 403 or
  429 with zero remaining; a secondary limit may include `retry-after`. The same source explains
  that `GET /rate_limit` does not consume primary quota, but may count against secondary quota.
- GitHub documents REST secondary limits in points (most GET/HEAD/OPTIONS are one point, most
  mutations five). The declaration must only encode budgets that the existing selector/cost schema
  can state honestly; do not invent endpoint costs.

### Scope and auth flow

- The engine intentionally refuses a declared policy when its non-secret `scope.subject_config` is
  absent. A GitHub policy therefore needs an explicit non-secret account/IP subject, not a raw
  token, approval revision, or credential binding.
- `spec.json` currently has `auth_type`, `owner`, `repo`, and `installation_id`, but no generic
  authenticated-account or public-IP subject. The rate declaration slice must add the smallest
  documented non-secret selector/scope fields and test their fail-closed behavior. It must not
  pretend `owner` identifies the authenticated user when a token can target another account.
- The live harness must set the active auth type and non-secret subject explicitly. It may use a
  credential name but never accepts, prints, or stores the credential value.

### Existing GitHub documentation work

- Current branch already contains derived `CONFIRMATION` help and tests that remove the false
  GitHub `--allow-destructive` notes. Re-verify rather than duplicate it. The latest captain scope
  allows only a debt count outside GitHub: 21 commands across Ashby (12), YouTube Analytics (4),
  Recurly (3), Gong (1), and Gorgias (1).
- `repo delete` and `repo delete-2` share the destructive runtime path. `repo create`, archive,
  unarchive, and secret set remain non-destructive approval writes; issue delete remains blocked
  pending a GraphQL mutation executor.

### Live proof implementation shape

- The sweep must invoke the built `pm` binary, not replay fixtures. A GitHub-specific script is
  appropriate now: it can enumerate the `implemented` command surface, keep raw output in process,
  write only redacted outcome/assertion records, and refuse any terminal state outside proven,
  concrete-untestable, or failed.
- The harness needs a deterministic, no-network self-test fixture to cover enumeration, redaction,
  and state accounting before it contacts GitHub.
- A live command needs a data assertion, not just exit code. For JSON output, record a schema/path
  assertion and status; for a mutation, verify provider-visible state in the dedicated private test
  repository; for binary download, assert the bounded destination artifact. Do not persist response
  bodies or approval values.
- The earlier codeload failure must be traced through the binary-download redirect path. Any fix
  must preserve bounded destination semantics and avoid forwarding credentials across an external
  redirect.

## Risks and controls

| Risk | Control |
|---|---|
| Accidental quota exhaustion / secondary throttle | Real declared policy, serial sustained sweep, free `/rate_limit` observations, fail run on GitHub 429. |
| Credential or grant leakage | Use credential names, in-memory parsing, redacted records, and tests asserting no secret-like output persists. |
| Write outside test scope | Require the dedicated private repository identity for every live write and reject other target configurations. |
| Incorrectly classifying unsupported operations | Require a concrete provider/permission/state reason for each untestable result; no fixture-only terminal state. |
| Unrelated connector churn | GitHub-only code/data changes; regenerate and object-diff shared artifacts before commit. |

## Required skills recorded

- `golang-how-to`, `golang-cli`, `golang-testing`, `golang-error-handling`,
  `golang-security`, `golang-safety`, `golang-design-patterns`,
  `golang-structs-interfaces`, `golang-context`, `golang-concurrency`, and
  `golang-documentation`.

