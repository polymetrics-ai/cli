# SECURITY-REVIEW — wave0-engine-harness

Reviewer: gsd-loop-security · Branch: connector-architecture-v2 · HEAD: b3f91af
Scope: `git diff main...HEAD` (engine/defs/conformance/certify/connectorgen additions), verified
against `.planning/phases/wave0-engine-harness/THREAT-MODEL.md`.

Verdict: **pass-with-findings** (no blockers; two major findings require a fix or an explicit
accepted-risk sign-off before this engine is wired to a live-write-capable connector; the rest are
minor/info hardening and doc-accuracy items).

## Findings

### MAJOR

**M1 — `link_header` pagination has no SSRF same-host guard**
`internal/connectors/engine/paginate.go:33-34` (`newPaginator`, `case "link_header"`) returns a bare
`&connsdk.LinkHeaderPaginator{}` with no wrapper. `connsdk.LinkHeaderPaginator.Next`
(`internal/connectors/connsdk/paginate.go:129-138`) follows whatever `Link: rel="next"` URL the
response supplies with zero host validation — no `BaseHost`, no `allow_cross_host` check, no loop
guard. This is the exact SSRF vector THREAT-MODEL §3 claims is covered ("`next_url`/Link-header
follow: engine requires same-host... unless the bundle sets `allow_cross_host: true`"), but the
control is implemented ONLY for `pagination.type: "next_url"` (via the engine-local `nextURL`
wrapper, paginate.go:174-239) — never for `pagination.type: "link_header"`. GitHub-style connectors
(the paginator's own docstring example) use `link_header` pagination, so a compromised/malicious
upstream API can redirect the paginator to an arbitrary internal host (e.g. a cloud metadata
endpoint) via a crafted `Link` response header, and the engine will request it with the connector's
configured auth applied (same-request-cycle `Requester.Auth.Apply`, not a transport-level redirect,
so `Authorization` is NOT stripped the way it would be on an actual cross-origin HTTP redirect).
- File: `internal/connectors/engine/paginate.go:33-34`
- Recommendation: add an engine-local `linkHeaderPaginator` wrapper (mirroring `nextURL`) that
  applies the same `BaseHost`/`allow_cross_host`/loop-guard checks before returning a `NextPage{URL:
  next}`, and wire its `BaseHost` from `requester.BaseURL` the same way `read.go:82-84` does for
  `next_url`. Add a parity test analogous to
  `TestNewPaginatorNextURLSSRFGuardDifferentHostRejected`.

**M2 — `secret_redaction` capability does not scan "ALL captured stdout/stderr"; only 2 of ~13
stages' stdout are scanned, and stderr is never scanned anywhere**
THREAT-MODEL §1/§7 and a code comment in `stages_source_test.go:177` both claim the harness "scans
ALL captured stdout/stderr + the report for planted secret values." The actual implementation:
- `finalizeSecretRedaction` (`internal/connectors/certify/stages_source.go:850-859`), which is what
  sets `Capabilities.SecretRedaction.Result` (the field asserted by tests and read by consumers of
  the persisted report), only calls `ScanForSecrets(stage.CLI.ArgvRedacted, ...)` — i.e. it scans
  the ALREADY-REDACTED argv string, never raw `Stdout`/`Stderr`.
- Only two stages (`stageManualJSON` at `stages_source.go:324`, `stageCredentialsTest` at
  `stages_source.go:370`) separately call `ScanForSecrets(res.Stdout, ...)` inline, and a failure
  there is folded into that stage's own pass/fail, not into `SecretRedaction`.
- The ~11 other stages that call `rc.harness.Run` (`connection_create`, `catalog_refresh`, every
  `etl run` invocation across full_refresh/incremental/resume/query stages —
  `stages_source.go:384,398,454,478,561,607,616,657,678,692,724,796,814`) never scan their stdout.
  `etl run` is the highest-risk call (it exercises the live connector against a real/fixture API and
  is the most likely place for a secret to leak into an error message or verbose output).
- `CLIResult.Stderr` (`cliharness.go:53`) is captured but is never passed to `ScanForSecrets`
  anywhere in the codebase.
- Net effect: a secret leaking into stdout of an `etl run` stage, or into ANY stage's stderr, is
  silently missed by the `secret_redaction` capability, which will still report `"pass"`.
- File: `internal/connectors/certify/stages_source.go:850-859` (finalizeSecretRedaction),
  `cliharness.go:169-187` (ScanForSecrets call sites)
- Recommendation: either (a) have every `recordStage` call scan both `res.Stdout` and `res.Stderr`
  and roll a hit into `finalizeSecretRedaction`'s aggregate result (requires threading captured
  stdout/stderr into `StageResult`, not just the redacted argv), or (b) at minimum extend
  `finalizeSecretRedaction`'s doc/name to accurately describe what it checks today (argv only) and
  track full-output scanning as an explicit backlog item — do not leave the capability's stated
  scope and actual scope diverged, since operators will treat `secret_redaction: pass` as a
  stronger guarantee than it is.

### MINOR

**m1 — Base-request transport-level redirects are not blocked; SSRF guard only covers
paginator-supplied next URLs, not 3xx `Location` redirects on the primary request**
`connsdk.Requester.client()` (`internal/connectors/connsdk/http.go:78-83`) returns a plain
`&http.Client{Timeout: ...}` with no `CheckRedirect` override, so Go's default (follow up to 10
redirects, any host) applies to every `Do`/`DoForm` call — base URL, OAuth2 token fetch, write
actions, everything. THREAT-MODEL §3's SSRF control is scoped to the paginator's own follow-the-body
logic and doesn't mention transport-level redirects at all. Mitigating factor: Go's stdlib client
strips `Authorization`/`Cookie`/auth-sensitive headers on a cross-host redirect since Go 1.8, so
credential leakage to the redirect target is not automatic — but SSRF (the ability to make the
engine issue a request, with no credential, to an attacker-chosen internal host) is still possible
via a malicious/compromised upstream returning a 3xx. This is pre-existing `connsdk` behavior (not
new in this diff — `connsdk/http.go` shows no diff vs `main`), so it is the inherited baseline
posture, but the THREAT-MODEL doesn't call it out as a residual risk and a reviewer reading §3 could
reasonably conclude redirects are covered by the "SSRF guard."
- File: `internal/connectors/connsdk/http.go:78-83`
- Recommendation: add `CheckRedirect` (deny cross-host, or replicate the same allow_cross_host
  policy) to the shared `Requester` default client, or explicitly document this as an accepted
  residual risk in THREAT-MODEL §3 (it currently only mentions the local test-server exception).

**m2 — `next_url`/Link-header same-host check compares host (with port) but not scheme; a
`https://` base can be silently downgraded to `http://` on the same host and still pass the guard**
`urlHost`/`requesterHost` (`paginate.go:233-238`, `read.go:468-474`) both derive comparison values
from `url.URL.Host` only, never checking `url.URL.Scheme`. A hostile API could return `next:
"http://api.example.com/..."` when the base was `https://api.example.com` — same `Host` string,
guard passes, and the follow-up request is sent in plaintext (on-path attacker can then read/tamper
with a request that may carry the connector's Authorization header, since this is a same-host
follow, not a cross-host block). No test exercises a same-host/different-scheme case (all
`paginate_test.go` SSRF tests use the same `httptest.Server` for both "same-host" and "cross-host"
so scheme is identical throughout).
- File: `internal/connectors/engine/paginate.go:210-215`, `read.go:468-474`
- Recommendation: compare `(scheme, host)` pairs, not host alone; reject a next_url whose scheme
  downgrades from https to http even when the host matches, unless allow_cross_host is set. Add a
  same-host/different-scheme regression test alongside
  `TestNewPaginatorNextURLSSRFGuardDifferentHostRejected`.

**m3 — Stream read paths are never passed through `InterpolatePath`; the urlencode-by-default
control is unreachable from `Read`**
`readDeclarative` (`internal/connectors/engine/read.go:123-129`) uses `stream.Path` (loaded raw from
`streams.json`) directly as the request path — `InterpolatePath` is called only from `write.go`
(lines 123, 202), never from `read.go`. Today this is not exploitable: every shipped bundle
(stripe, searxng, postgres) uses static/non-templated `stream.path` values, and the JSON Schema for
`streams.json` (`schema/streams.schema.json:30`) does not forbid `{{ }}` templates in `path`, nor
does `connectorgen validate`'s `ResolveCheck` (which only checks that a referenced `config.*` key is
declared — it does not distinguish "declared but never runtime-interpolated"). If a future
bundle/fan-out author writes a stream `path` containing `{{ record.x }}` or `{{ config.x }}`, it
will pass `connectorgen validate` but be sent to the live API as the literal uninterpolated string
at runtime (a correctness bug, and it means the injection control the THREAT-MODEL describes for
"path segments are the primary injection surface" does not apply uniformly to reads vs. writes).
- File: `internal/connectors/engine/read.go:123-129`
- Recommendation: either add `connectorgen validate` enforcement that `streams.json` `path` fields
  contain no `{{ }}` templates (documenting reads as static-path-only by design), or wire
  `InterpolatePath` into the read path for parity with writes if templated read paths are meant to
  be supported in a later wave. Track as a backlog item for whichever wave adds record-path-driven
  reads.

**m4 — `ScanForSecrets` does not check a JSON-escaped form of the secret**
`containsSecretForm` (`cliharness.go:189-206`) checks exact, base64 (standard + raw/no-padding), and
URL-encoded (`url.QueryEscape` + path-escaped) forms, but not a JSON-string-escaped form (e.g. a
secret containing `"`, `\`, or control characters, rendered inside a `--json` envelope, would be
escaped per RFC 8259 and could evade all four current checks while still being visible verbatim to
a human reading the JSON). The existing test
(`TestScanForSecretsDetectsExactBase64AndURLEncodedForms`) uses a secret
(`sk_test_topsecret12345`) with no JSON-special characters, so it can't surface this gap. Narrow in
practice (most generated secret values are alphanumeric/underscore and would be unaffected), but the
THREAT-MODEL explicitly enumerates "exact, base64, and URL-encoded" forms without JSON-escaped, so
this is a known, documented-but-incomplete enumeration rather than a silent gap.
- File: `internal/connectors/certify/cliharness.go:189-206`
- Recommendation: add a `strconv.Quote`-derived (or `encoding/json`-marshaled string) form to
  `containsSecretForm`'s checks; add a regression test with a secret containing a `"` or backslash.

**m5 — Certification report/history files use world-readable `0o644`/`0o755`, unconditionally**
`Report.Save` (`report.go:112,122,127,131`) always writes with `0o644` files / `0o755` dirs
regardless of the caller's umask context. Consistent with the design intent that reports never
carry secret values, but that guarantee is only as strong as M2's redaction coverage above — if a
future stage leaks a secret into an argv string that isn't caught by `finalizeSecretRedaction`
before `Save` is called, the leaked value would land in a world-readable file. Low severity on its
own; flagged because it compounds M2.
- File: `internal/connectors/certify/report.go:112,122,127,131`
- Recommendation: no change required if M2 is fixed; otherwise consider `0o600`/`0o700` for
  certification artifacts as defense in depth, matching the `0o600` already used for
  `writeCaptureFile` (`stages_source.go:529`).

### INFO

**i1 — `secretLiteralPattern` regex is duplicated verbatim between `cmd/connectorgen/validate.go:73`
and `internal/connectors/conformance/static.go:82`.** Both copies are currently identical, so no
functional drift exists yet, but a future edit to one without the other would silently create an
inconsistency between `connectorgen validate` and the conformance `secret_redaction` check. Consider
extracting to a shared internal package. Not a vulnerability today.

**i2 — THREAT-MODEL's own fixture example contradicts its validator.** THREAT-MODEL §1 says
"Fixture values must be synthetic (`sk_test_fixture…`)" as the recommended synthetic-secret
convention, but `secretLiteralPattern` (`\bsk_(live|test)_[a-z0-9]{10,}\b`) would flag exactly that
literal as a secret-shaped violation (alphanumeric suffix "test_fixture..." satisfies the 10+-char
class). The actual shipped fixtures (stripe, searxng, goodconn testdata) correctly avoid `sk_`
prefixes (using `cus_fixture_1`-style IDs instead), so the validator behaves correctly and no real
fixture violates it — this is purely a stale/inaccurate example in THREAT-MODEL prose.
- Recommendation: fix the THREAT-MODEL wording to match the actual convention used in shipped
  fixtures (non-`sk_`-prefixed synthetic IDs), or loosen the regex if `sk_test_` fixture literals
  are meant to be an accepted convention.

**i3 — No test asserts write-path record-value traversal/metachar encoding end-to-end through
`Write`/`executeWriteRecord`.** `TestInterpolatePathDefaultURLEncode` proves `InterpolatePath` itself
neutralizes traversal at the unit level, and `executeWriteRecord` (`write.go:199-205`) correctly
calls `InterpolatePath(action.Path, vars)` with `vars.Record` populated from the write record, so the
composition is sound by inspection — but no `write_test.go` case feeds a `record.*` value containing
`/`, `..`, or CRLF through the full `Write()` entrypoint to lock in the composition as a regression
guard (existing tests use clean values like `"a.txt"`, `"cus_1"`). Recommend adding one.

## Checklist verification summary (against THREAT-MODEL claims)

1. **Secrets** — `x-secret` partitioning is metadata-only in this phase (declarative; the actual
   Config/Secrets split is pre-existing/out-of-scope, confirmed `connectors.go` RuntimeConfig has
   zero diff vs `main`). `engine.Error` → `safety.RedactErrorText` confirmed
   (`errors.go:51`, `connsdk/http.go:46`, both pre-existing redaction path, reused correctly).
   `DryRunWrite` redaction confirmed structurally correct (`write.go:112-117`) though its own test
   doesn't template a secret into the previewed path (see i3-adjacent note above; not a finding on
   its own since the redaction map-swap happens unconditionally before interpolation). Certify scan
   completeness gap: see **M2**, **m4**.
2. **Injection** — CRLF rejection is centralized in `resolveExpr` (interpolate.go:90-92) and thus
   applies uniformly to `Interpolate`/`InterpolatePath`/`InterpolateHeader`; both header and path
   CRLF cases are tested (`interpolate_test.go:166-180`). `urlencode`-by-default for path segments
   confirmed via `TestInterpolatePathDefaultURLEncode` (traversal `a/../b` → `a%2F..%2Fb`, double-
   encode guard for literal `%`). Record-sourced write-path values ARE urlencoded (`write.go:202`
   calls `InterpolatePath`, not `Interpolate`) — confirmed correct. Stream READ paths never go
   through `InterpolatePath` at all: see **m3**.
3. **SSRF** — `next_url` same-host guard implemented and tested
   (`paginate.go:174-239`, `read.go:82-84`); host comparison includes port (uses `url.URL.Host`
   symmetrically) but not scheme (**m2**); `link_header` pagination has NO guard at all (**M1**,
   the most significant finding of this review); transport-level redirect-following is unguarded for
   every request type, not just pagination (**m1**, pre-existing/inherited, not new in this diff).
4. **Auth semantics** — Traced the exact scenario requested: a bundle with a single auth spec whose
   only `when` references an absent secret. `EvalWhen`'s absent-key-falsy tolerance
   (`interpolate.go:272-297`, scoped to `resolveRefForWhen` only) makes `authSpecMatches` return
   `false`, and `selectAuth`'s loop (`auth.go:34-45`) simply `continue`s past every non-matching
   spec; when no spec matches, it explicitly returns `fmt.Errorf("select auth: no auth spec matched
   for auth_type %q", ...)` — **never** a nil/silent-none Authenticator. `newRuntime`
   (`read.go:214-220`) propagates that error rather than falling back; the ONLY way `auth` ends up
   nil is the explicit, separate `len(b.HTTP.Auth) == 0` branch (zero declared auth specs, a
   documented distinct case for genuinely public APIs). Confirmed by
   `TestSelectAuthNoRuleMatchesIsTypedError` and `TestSelectAuthGithubStyleAutoTokenPublicTable`.
   **No finding** — the threat model's design intent holds; silent-unauthenticated dispatch for a
   bundle that DOES declare auth is not possible.
5. **Certify** — Ephemeral workdir cleanup via `defer os.RemoveAll(root)` unless `opts.KeepWork`
   (`stages_source.go:100-107`), tested (`stages_source_test.go:245-269`). `KeepWork`'s residual
   risk (plaintext secrets surviving on disk) is mitigated by the pre-existing AES-GCM vault
   encryption for `credentials add` (`internal/cli/docs.go:129`), out of this phase's scope. Argv
   redaction confirmed (`redactArgv`/`redactSecretsInText`, `cliharness.go:150-167`), tested
   (`TestHarnessRunRedactsArgvSecrets`). Report file permissions: see **m5**. Secret-scan
   completeness: see **M2**, **m4**.
6. **Supply chain** — `git diff main...HEAD go.mod go.sum` is empty: confirmed zero new
   dependencies. `.golangci.yml` is a net-new file (no prior repo baseline existed) that is purely
   additive (`govet`, `staticcheck`, `errcheck`, `ineffassign`, `unused`, `misspell`), scoped via the
   Makefile `lint` target to the new engine tree only, explicitly to avoid loosening the linter set
   to accommodate the legacy ~560-connector tree — no gate weakening. **No finding.**

## Residual risks carried forward (not new to this phase)

- `connsdk.Requester`'s HTTP client has no `CheckRedirect` policy (pre-existing; see m1).
- Live-write approval-token handling and `pm reverse` are untouched by this phase, per
  THREAT-MODEL §5 — verified no new write dispatch path exists outside test harnesses
  (`certify` write/flow stages are explicitly out of scope per THREAT-MODEL §5, confirmed:
  `report.go` Capabilities struct has no `write_actions` field yet).
