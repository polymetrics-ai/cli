# THREAT-MODEL — wave0-engine-harness

Scope: new engine/defs/conformance/certify code paths. Legacy behavior unchanged (see SPEC §2), so
the existing threat posture (encrypted credentials, plan/preview/approve writes, `safety`
redaction) is the baseline; this file covers the deltas.

## 1. Secrets in bundles, errors, and fixtures

- **Threat**: bundle authors (or fan-out agents) embed real tokens in `spec.json` defaults,
  `fixtures/`, or `docs.md`; engine errors echo request bodies/URLs containing secrets.
- **Controls**:
  - `x-secret: true` in `spec.json` is the single source of truth for the config/secret split
    (design §A); the loader partitions keys and the engine only reads secret values from
    `RuntimeConfig.Secrets` (never `Config`), mirroring `stripe/stripe.go:279`.
  - `engine.Error` messages pass through `safety.RedactErrorText`
    (`internal/safety/safety.go:50`); `connsdk.HTTPError.Error()` already applies it
    (`connsdk/http.go:46`). Engine never logs bodies.
  - `DryRunWrite` previews render resolved method/path with secret values replaced by
    `***` before inclusion in `WritePreview.Warnings` (T-09 asserts).
  - Static conformance check `secret_redaction` + `connectorgen validate` scan fixtures/docs for
    secret-shaped literals (`sk_live_`, `ghp_`, `Bearer <base64ish>`, key names flagged x-secret
    with non-placeholder values). Fixture values must be synthetic (`sk_test_fixture…`).
  - Certify harness scans ALL captured stdout/stderr + the report for planted secret values in
    exact, base64, and URL-encoded forms (T-12).
- **Residual**: entropy-based detection is heuristic; adversarial review checklist
  (orchestration-plan §Verification pyramid) covers fixture realism + redaction on samples.

## 2. Interpolation injection

- **Threat**: config values flow into URL paths/queries — `repository = "a/../admin"`,
  `"x?y=1#frag"`, header CRLF injection via `\r\n` in header templates.
- **Controls**:
  - `urlencode` is the DEFAULT filter for every path-segment insertion (design §B.3); T-02 has
    explicit traversal/metachar/double-encode cases.
  - Query values go through `url.Values.Encode()` in `connsdk.Requester.resolveURL`
    (`connsdk/http.go:145`) — never string-concatenated.
  - Header values: engine rejects interpolated header values containing CR/LF; empty value =
    header omitted (also functional requirement).
  - `when` conditions evaluate against parsed values with a closed grammar — no eval, no
    templating engine, no user functions.
  - `connectorgen validate` resolves every `{{ }}` at build time against `spec.json` properties;
    unknown keys fail merge.

## 3. SSRF via base_url and next_url

- **Threat**: connection config overrides `base_url` to an internal address; a hostile API
  response supplies a `next_url`/Link header pointing at internal services.
- **Controls**:
  - Engine base-URL rule mirrors `stripeBaseURL` (`stripe/stripe.go:289`): scheme must be
    http/https, host required; enforced once at requester build.
  - `next_url`/Link-header follow: engine requires same-host as the resolved base URL unless the
    bundle sets an explicit `allow_cross_host: true` escape (none of the goldens need it);
    loop guard errors when the same next URL repeats (T-06).
  - Tier-3 postgres keeps the legacy host validation (bare hostname, no scheme/path —
    `postgres.go:119` `resolveConfig`).
- **Residual**: local test servers use http://127.0.0.1 — allowed by design (same rule as legacy).

## 4. Fixture sanitization and replay integrity

- **Threat**: recorded fixture pages contain real PII/tokens; replay server accidentally serves a
  page for the wrong request masking request-construction bugs.
- **Controls**: fixtures are hand-authored/sanitized synthetic data in wave0 (recording tooling is
  a later-phase `httpx` deliverable); the replay envelope keys each page by expected method +
  path + query (DATA-MODEL §5) and an unmatched request is a conformance FAILURE (drift catch);
  conventions.md carries the sanitization checklist for fan-out waves.

## 5. Write-path safety

- No live writes exist in wave0: engine `Write` is exercised only against httptest; certify write
  stages are explicitly out of scope. The plan→preview→approve flow and approval-token handling
  are UNTOUCHED (`pm reverse` code paths not modified). `confirm: "destructive"` metadata is
  carried through `Definition` for later phases but enforces nothing new yet.
- Registration coexistence (SPEC §2) guarantees no engine connector is reachable from the CLI, so
  no new remote-write surface ships this phase.

## 6. Supply chain / tooling

- Zero new go.mod dependencies (draft-07 validator is internal; paginators wrap connsdk).
- golangci-lint acquisition is pinned by version whichever path the coordinator picks (SPEC §5);
  it runs as a local gate, not with repo write access.
- `connectorgen gen` outputs are deterministic (byte-stable) to keep generated-file diffs
  reviewable; wave-close path-guard (`git status --porcelain`) rejects out-of-scope writes.

## 7. Abuse of the certify harness

- Harness executes only the in-process CLI with an ephemeral `--root`; it never shells out; argv
  recorded in reports is redacted; ephemeral roots are deleted (leak of workdir = test failure).
  Budget/rate-limit machinery is deferred with the live tiers.
