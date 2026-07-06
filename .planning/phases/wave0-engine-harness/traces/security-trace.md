# Agent Trace: security

## Rendered Prompt Or Prompt Reference

gsd-loop-security reviewer for phase `wave0-engine-harness`, branch `connector-architecture-v2`,
HEAD `b3f91af`. Read-only security pass over `git diff main...HEAD`, guided by
`.planning/phases/wave0-engine-harness/THREAT-MODEL.md`, verifying every claimed control against the
actual implementation and tests, per the 6-item checklist (secrets, injection, SSRF, auth semantics,
certify harness, supply chain).

## Files Inspected

- `.planning/phases/wave0-engine-harness/THREAT-MODEL.md`
- `internal/connectors/engine/interpolate.go`, `interpolate_test.go`
- `internal/connectors/engine/auth.go`, `auth_test.go`
- `internal/connectors/engine/errors.go`
- `internal/connectors/engine/write.go`, `write_test.go`
- `internal/connectors/engine/read.go`
- `internal/connectors/engine/paginate.go`, `paginate_test.go`
- `internal/connectors/engine/connector.go` (specJSON secret-metadata reconstruction)
- `internal/connectors/engine/schema.go` (SecretKeys/Properties)
- `internal/connectors/engine/bundle.go` (StreamSpec/WriteAction shapes)
- `internal/connectors/connsdk/http.go` (Requester.Do, resolveURL, client(), HTTPError)
- `internal/connectors/connsdk/paginate.go` (LinkHeaderPaginator — pre-existing, unmodified)
- `internal/connectors/connectors.go` (RuntimeConfig — confirmed zero diff vs main)
- `internal/connectors/certify/report.go`, `report_test.go`
- `internal/connectors/certify/cliharness.go`, `cliharness_test.go`
- `internal/connectors/certify/stages_source.go`, `stages_source_test.go`
- `internal/connectors/conformance/static.go` (checkSecretRedaction, secretLiteralPattern)
- `internal/connectors/conformance/dynamic.go` (SecretKeys usage for synthetic secret injection)
- `cmd/connectorgen/validate.go` (checkFixtureSecrets, checkWritePathFields, ResolveCheck wiring)
- `internal/connectors/defs/stripe/streams.json`, `stripe/fixtures/streams/customers/page_1.json`
- `cmd/connectorgen/testdata/{invalid/secret-literal-in-fixture,valid/goodconn}/fixtures/...`
- `.golangci.yml`, `Makefile` (diff vs main)
- `go.mod`, `go.sum` (diff vs main — confirmed empty)

## Actions Taken

- Traced the exact auth-fallthrough scenario named in the task (single auth spec, `when:
  {{ secrets.token }}`, token absent) through `authSpecMatches` -> `EvalWhen` ->
  `resolveRefForWhen` -> `selectAuth`'s no-match path -> `newRuntime`'s error propagation, to confirm
  no silent-unauthenticated dispatch is possible for a bundle that declares auth specs.
- Traced every `InterpolatePath`/`Interpolate`/`InterpolateHeader` call site across the engine
  package to map exactly which request-construction paths get urlencode-by-default and CRLF
  guarding, and which don't (found read.go's stream.Path never does).
- Traced the `next_url` SSRF guard's `BaseHost` wiring end-to-end from `read.go:82-84` through
  `paginate.go`'s `nextURL.Next`, and separately traced `pagination.type: "link_header"` to confirm
  `connsdk.LinkHeaderPaginator` has no equivalent guard.
- Traced `ScanForSecrets`/`finalizeSecretRedaction`/`redactArgv` call sites across
  `certify/stages_source.go` to count exactly how many of the ~13 harness stages have their
  stdout/stderr scanned vs. only their redacted argv.
- Confirmed `connectors.RuntimeConfig` (Config/Secrets split) has zero diff vs `main` — the
  Config/Secrets partitioning boundary itself is pre-existing/out of this phase's scope; this
  phase's `x-secret`/`SecretKeys()` is declarative metadata only.
- Confirmed Go stdlib redirect-header-stripping behavior (Authorization stripped on cross-host
  redirect since Go 1.8) to correctly scope the severity of the transport-redirect finding (m1) as
  SSRF-only, not credential-leak.

## Commands Run

- `git diff main...HEAD --stat`
- `git diff main...HEAD -- go.mod go.sum` (confirmed empty)
- `git diff main...HEAD -- .golangci.yml`
- `git diff main...HEAD -- Makefile`
- `git diff main...HEAD -- internal/connectors/connectors.go` (confirmed empty)
- `grep -rn "SecretKeys\b" internal cmd`
- `grep -rn "ScanForSecrets" internal/connectors/certify/*.go`
- `grep -n "rc.harness.Run|res.Stderr|ScanForSecrets" internal/connectors/certify/stages_source.go`
- `grep -rn "InterpolatePath(" internal/connectors/engine/*.go`
- `grep -rn "LinkHeaderPaginator|link_header" internal/connectors/engine/*.go`
- `grep -n "CheckRedirect|http.Client{" internal/connectors/connsdk/*.go`
- `go test ./internal/connectors/engine/... ./internal/connectors/certify/... ./internal/connectors/conformance/...` (all pass)

## Findings

Full detail in `.planning/phases/wave0-engine-harness/SECURITY-REVIEW.md`. Summary:

- **Major (2)**: `link_header` pagination has no SSRF same-host guard (M1); certify's
  `secret_redaction` capability only scans argv, not the "ALL captured stdout/stderr" the
  THREAT-MODEL and a test comment both claim — most stages (including every `etl run`) are
  unscanned, and stderr is never scanned anywhere (M2).
- **Minor (5)**: no `CheckRedirect` policy on the shared HTTP client, so 3xx transport redirects
  bypass the SSRF guard entirely (m1, pre-existing/inherited); same-host guard compares host+port
  but not scheme, allowing an https->http downgrade on a matching host (m2); stream READ paths never
  go through `InterpolatePath` at all — the urlencode-by-default control is unreachable from `Read`,
  latent since no shipped bundle templates a read path today (m3); `ScanForSecrets` doesn't check a
  JSON-escaped secret form (m4); certification report/history files are unconditionally
  world-readable (m5, compounds M2 if unfixed).
- **Info (3)**: duplicated `secretLiteralPattern` regex between connectorgen and conformance (i1);
  THREAT-MODEL's own fixture example (`sk_test_fixture…`) would trip its own validator (i2, prose
  bug, shipped fixtures are correct); no end-to-end write-path traversal regression test through
  `Write()` itself, though the composition is sound by inspection (i3).
- **No finding** on auth-semantics fallthrough (item 4) — this was the checklist's flagged "severity
  major if silent-unauthenticated is possible" scenario, and it is confirmed NOT possible:
  `selectAuth` errors explicitly when no spec matches for a bundle that declares >=1 auth spec.
- **No finding** on supply chain (item 6) — zero new go.mod deps, `.golangci.yml` is net-new and
  purely additive, no gate weakened.

## Handoff Summary

Two majors need owner decisions before this engine backs a live-write-capable connector:
1. Backend/engine owner: add an engine-local `linkHeaderPaginator` wrapper applying the same
   `BaseHost`/`allow_cross_host`/loop-guard checks as `nextURL` (M1). Straightforward, mirrors
   existing code exactly.
2. Backend/certify owner: decide between (a) threading captured stdout/stderr into every stage for
   full-output secret scanning in `finalizeSecretRedaction`, or (b) narrowing the
   `secret_redaction` capability's documented scope to match its actual behavior (argv-only) until
   full scanning lands (M2). Either is acceptable; leaving the mismatch undocumented is not.

Minor items (m1-m5) and info items (i1-i3) are backlog-appropriate; none block this phase's merge on
their own, but m1-m3 should be tracked before any wave that adds record-path-driven reads or
live-write dispatch through this engine, since they compound with write-capable connectors.

## Verification Evidence

- `go test ./internal/connectors/engine/... ./internal/connectors/certify/... ./internal/connectors/conformance/...`
  → all packages pass (cached, pre-existing green state; no test changes made by this review since
  it is read-only per scope).
- `TestSelectAuthNoRuleMatchesIsTypedError`, `TestSelectAuthGithubStyleAutoTokenPublicTable`,
  `TestSelectAuthSecretsNeverInError` (auth_test.go) directly evidence the item-4 scenario resolves
  safely (typed error, never silent-none).
- `TestInterpolatePathDefaultURLEncode`, `TestInterpolateHeaderCRLFInjectionRejected`
  (interpolate_test.go) directly evidence item-2 traversal/CRLF controls.
- `TestNewPaginatorNextURLSSRFGuardDifferentHostRejected`,
  `TestNewPaginatorNextURLAllowCrossHostEscape` (paginate_test.go) evidence the `next_url` guard
  works, but no equivalent test exists (or could exist, since no wrapper exists) for `link_header`
  — this absence is itself part of the M1 evidence.
- `TestScanForSecretsDetectsExactBase64AndURLEncodedForms`, `TestHarnessRunRedactsArgvSecrets`
  (cliharness_test.go) evidence the argv-redaction and per-form-scan mechanics work as far as they
  go; no test exists asserting stdout of a non-manual_json/credentials_test stage is scanned, which
  is consistent with M2's finding (absence of coverage matches absence of implementation).

## Unresolved Risks

- M1, M2 as stated above — require an owner decision/fix, not just documentation, before this engine
  backs any live connector with real credentials in a production certify run.
- m1 (transport redirect bypass) is inherited from pre-existing `connsdk` and not newly introduced,
  but is currently undocumented as a residual risk anywhere in THREAT-MODEL §3.
