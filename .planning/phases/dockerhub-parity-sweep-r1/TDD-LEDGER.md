# Docker Hub documented-operation parity — TDD ledger

## Baseline (before this phase)

`internal/connectors/defs/dockerhub/` was NOT a blank slate. It already had a complete,
`operation_ledger_version: 2` `api_surface.json` (54 rows: 4 `covered_by` the existing streams, 50
`operation.model: "disallowed"` rows with a `reason` but no `notes`), a 4-stream `streams.json`
(`"auth": [{"mode": "none"}]`), `metadata.json` (`capabilities.write: false`), `spec.json`
(`docker_username` only), `docs.md`, `schemas/*.json`, `fixtures/{check.json,streams/**}`. No
`operations.json` operations, `writes.json`, or `internal/connectors/hooks/dockerhub/` existed.

## RED — cmd/connectorgen/dockerhub_api_surface_test.go

Written against the target re-verified 2026-08-08 from the live artifact
(`https://docs.docker.com/reference/api/hub/latest.yaml`, 148,322 bytes, sha256
`99d9d53c2d93656a3c66d604885abd153dc5df285abc0ecb13802a3bc53d0756`): 54 operations (GET 24, POST
12, DELETE 6, PATCH 5, PUT 4, HEAD 3). This hash and (method, path) set independently re-derives
byte-for-byte identical to what `api_surface.json` already recorded — confirmed by direct
comparison, not assumed.

Modeled on `cmd/connectorgen/gorgias_api_surface_test.go`/`notion_api_surface_test.go`
(single-disposition-per-row, `named_dependency=` prefix check, per-method assertions,
duplicate-key detection), adapted for dockerhub's `operation_ledger_version: 2` shape (per-endpoint
`provenance{artifact,source_url}` block rather than v1's `operation.source_url`).

Asserted: `operation_ledger_version == 2`; `artifacts[]` non-empty with sha256 present; exactly 54
rows, each with `provenance.source_url`; the per-method split; zero legacy `excluded` rows; **zero
rows with `operation.model == "disallowed"`**; exactly one disposition per row; no duplicate
`method+path`; every blocked row carries a non-empty `reason` and a `notes` field prefixed
`named_dependency=`; a representative sample of expected endpoints present; and target
`covered`/`blocked` counts.

### Verbatim RED failure

No pre-implementation bundle-authoring work existed on disk when this test was written (confirmed
via `git status` immediately before running — nothing to stash, unlike gorgias). Run against the
genuine, disk-committed, pre-phase `api_surface.json`:

```
=== RUN   TestDockerhubAPISurfaceOperationLedger
    dockerhub_api_surface_test.go:118: GET /v2/access-tokens: operation.model is "disallowed" — the captain never authorised that classification; every not-yet-implementable operation must carry a real model (direct_read/binary_read/sensitive_reverse_etl/admin_reverse_etl/destructive_action/local_workflow/duplicate/deprecated) plus a named_dependency= note, never a bare disallowed refusal
    dockerhub_api_surface_test.go:127: GET /v2/access-tokens: blocked row must name its dependency (notes must start with "named_dependency=")
    [... identical pair repeated for all 50 disallowed rows: POST/GET/PATCH/DELETE /v2/access-tokens*, GET /v2/auditlogs/*, POST /v2/auth/token, POST/DELETE/PATCH /v2/invites/*, POST/HEAD/PATCH /v2/namespaces/*/repositories*, GET/POST/PUT/PATCH/DELETE /v2/orgs/*/access-tokens*, GET/PUT /v2/orgs/*/settings, GET/POST/PUT/PATCH/DELETE /v2/orgs/*/groups*, GET /v2/orgs/*/invites, GET/PUT/DELETE /v2/orgs/*/members*, POST /v2/repositories/*/groups, GET/POST/PUT /v2/scim/2.0/*, POST /v2/users/2fa-login, POST /v2/users/login ...]
    dockerhub_api_surface_test.go:149: 50 row(s) still carry operation.model "disallowed", want 0
    dockerhub_api_surface_test.go:201: covered = 4, want 39 (4 pre-existing streams + 35 newly implemented operations)
    dockerhub_api_surface_test.go:204: blocked = 50, want 15 (3 HEAD + 3 auth-exchange + 9 SCIM, each carrying a named_dependency= note)
--- FAIL: TestDockerhubAPISurfaceOperationLedger (0.00s)
FAIL
FAIL	polymetrics.ai/cmd/connectorgen	0.742s
FAIL
```

(Full untruncated output — all 50 pairs, not just the elided sample above — is preserved in the git
history of this file's originating commit and in `RUN-STATE.json`'s `tdd.red_failure`.)

A follow-up whole-package run (`go test ./cmd/connectorgen/`, no `-run`) confirmed
`TestDockerhubAPISurfaceOperationLedger` was the ONLY failure in the package — the initial
`wantCovered=39`/`wantBlocked=15` targets in this red run reflected the classification plan at the
time (before the captain's rulings expanded scope to 54/0); the test's target constants were
updated twice more as the captain's rulings arrived (see PLAN.md), each time re-run to confirm the
new red/green transition was genuine, not a rewritten assertion papering over unfinished work.

## Planned assertions → GREEN evidence

| Assertion | Target | Evidence |
| --- | --- | --- |
| `operation_ledger_version == 2` | 2 | `api_surface.json` top-level field (unchanged from baseline) |
| Row count | 54 | generated `api_surface.json` |
| Method split | GET 24 / POST 12 / DELETE 6 / PATCH 5 / PUT 4 / HEAD 3 | cross-checked against the fetched OpenAPI document independently in Python |
| Zero `disallowed` rows | 0 | every row carries `covered_by`, final state |
| Zero legacy `excluded` rows | 0 | schema forbids it once `operation_ledger_version` is set |
| Every blocked row has `named_dependency=` note | 0 blocked rows remain | N/A — final state has none |
| `covered`/`blocked` | 54/0 | `api_surface.json` |

## GREEN

```
=== RUN   TestDockerhubAPISurfaceOperationLedger
--- PASS: TestDockerhubAPISurfaceOperationLedger (0.00s)
PASS
ok  	polymetrics.ai/cmd/connectorgen	0.751s
```

Full gate transcript:

- `go run ./cmd/connectorgen validate internal/connectors/defs/dockerhub` → 0 findings (several fix
  rounds along the way: `covered_by.direct_read` target must be the CLI command **path**, not the
  operation id; a HEAD row needs `operation.duplicate_of` when still blocked; a POST rest_read needs
  `rest.content_type`/`rest.body_schema`; SCIM write flags must map to the record_schema's real
  nested field names, `record.name.givenName`/`familyName`, not synthetic flattened names; a
  `binary_download` operation needs `output_policy: "binary_file_bounded"`).
- `go run ./cmd/connectorgen validate internal/connectors/defs` (whole repo) → 551 connectors
  checked, 0 findings.
- `go test ./cmd/connectorgen/` (whole package) → PASS.
- `go test ./internal/connectors/commandrunner/ -run TestEveryImplementedCommandPassesRuntimePreflight`
  → PASS across all 551 connectors' implemented commands (the authoritative fleet-wide proof the
  new HEAD-method foundation did not regress any other connector).
- `go test ./internal/connectors/engine/...` → PASS, including 5 new dedicated HEAD-support tests
  (`TestOperationDirectReadHEADReturnsStatusOnlyNoBodyDecode`,
  `TestOperationDirectReadHEADNonSuccessStatusIsError`,
  `TestOperationDirectReadSpecRejectsHEADWithoutJSONRedactedPolicy`,
  `TestOperationDirectReadSpecAcceptsHEADForRestRead`, `TestPreflightOperationDirectReadAcceptsHEAD`).
- `go test ./internal/connectors/hooks/dockerhub/...` → PASS, 22 tests (login exchange request
  shape, JWT-exp-driven caching/refresh with a 4-minute fallback, error handling, secret redaction,
  context cancellation, and 4 dedicated `dualAuth`/SCIM tests proving the SCIM path routes to the
  independent `scim_bearer_token` credential, never falls back to the session JWT even when one is
  cached, and fails closed with a named error when unconfigured).
- `go test ./internal/connectors/conformance/...` → PASS (dockerhub's dynamic checks are
  `skip_dynamic: true` with a documented reason — conformance's synthetic-secret harness always
  synthesizes a non-empty `docker_pat`, which forces the custom-hook path even for
  unauthenticated-by-design public reads; the hook's own real behavior is proven by its dedicated
  test file instead, per `docs/migration/conventions.md`'s conformance-skip-marker convention).
- `go run ./cmd/connectorgen surface-sync --check` → 551 connectors scanned, 0 corrections (the
  hand-authored bundle content already matched what the tool derives).
- `go run ./cmd/connectorgen surface-sync` (write mode) → regenerated
  `internal/connectors/defs/operation_endpoint_ledger.json`, diff confined to the `"dockerhub": [...]`
  key, purely additive.
- `go run ./cmd/connectorgen gen` → regenerated `hookset_gen.go` to register the new `dockerhub`
  AuthHook (1-line diff, additive).
- `make connector-boundary` → `"outcome": "clean"`.
- `make lint` → 0 issues (caught and fixed one `unused` finding: a leftover unused test helper).
- `make tidy-check`, `make agent-contract-check`, `make smoke-no-build`, `make release-workflow-check`
  → all pass.
- `gofmt -l cmd internal` and `go vet ./cmd/... ./internal/connectors/...` → clean.
- `go build ./...` → clean (whole repo).
- Full `go test -timeout 20m ./internal/cli/...` → PASS, including `TestGoldenTranscripts` after
  regenerating (`POLYMETRICS_UPDATE_GOLDEN_TRANSCRIPTS=1`) — diff scoped to the 9 root-listing
  subtests, each changed only in the single `pm dockerhub <command> - Docker Hub: ...` tagline line.
- Binary built and exercised live:
  - `pm connectors inspect dockerhub --json` → `capabilities.write: true`.
  - Bare `pm dockerhub` → contextual help, exit 0.
  - **All 54 `cli_surface.json` commands' `--help`** → exit 0 (scripted sweep over the real binary,
    not a diff read).
  - **Live loopback HTTP round trip**: `pm dockerhub repository check --namespace library
    --repository alpine --credential dockerhub-test --json` against a local Python `http.server`
    implementing only `do_HEAD` (any other method would 501) → the server logged
    `GOT HEAD /v2/namespaces/library/repositories/alpine`, and the CLI returned
    `{"method":"HEAD","response":{"status_code":200},"status":200}` — genuine proof the new
    status-only HEAD path executes correctly end-to-end, not just in unit tests.
- Website catalog regenerated (`gen-connector-bundles.mjs`, `gen-connector-catalog.mjs`,
  `gen-connectors.mjs`); diffs checked **by object** (Python dict comparison), not by line — only
  the `"Docker Hub"` entry differs in `connectors.generated.json` and
  `connectors.catalog.data.generated.json`; `connectors.catalog.generated.ts`'s single-line diff is
  the aggregate write-capability counter (236→237).
- `docs/cli` and `docs/connectors` regenerated via `./pm docs generate --dir docs/cli`; reverted
  every non-dockerhub path under `docs/connectors/` (1,027 files of pre-existing `main` drift, each
  regeneration pass) — kept only `docs/connectors/dockerhub/{MANUAL.md,SKILL.md}`.

See `RUN-STATE.json`'s `tdd.green`/`tdd.green_evidence` fields for the compact form.

## Review-fix round — write-action request line and SCIM auth routing

Automated review found the 26 new `writes.json` actions unreachable and the SCIM auth route
base-URL-fragile. Both were confirmed against the real execution paths before any edit.

### Red

- **Write paths never substituted, and doubled the base path.** `engine.InterpolatePath`'s
  `templatePattern` (`internal/connectors/engine/interpolate.go:54`) only expands `{{ … }}`, so the
  single-brace `{namespace}` form copied from Docker Hub's API surface survived verbatim;
  `path_fields` only *excludes* keys from the JSON body (`write.go:437`), it substitutes nothing.
  Separately, the write path has no `normalizeDirectReadPathForBaseURL` equivalent — `joinURL`
  (`write.go:268`) and `connsdk`'s `resolveURL` (`connsdk/http.go:371`) both concatenate — so the
  `/v2`-prefixed action paths doubled against `base_url`'s own `/v2`.
  Red evidence: resolving every action through the real `resolveWriteRequestLine` against the
  disk-committed bundle produced e.g.
  `POST https://hub.docker.com/v2/v2/namespaces/{namespace}/repositories` — a literal brace pair in
  a doubled path, for all 26 actions.
- **SCIM routing failed open.** `dualAuth.Apply` matched the absolute `req.URL.Path` against the
  hardcoded `"/v2/scim/2.0/"`. That prefix only holds while `base_url`'s path is exactly `/v2` —
  and `spec.json` advertises `base_url` as a self-hosted-proxy override. Under any other path the
  match misses and the request silently falls through to the account session JWT, the exact
  credential substitution the type exists to prevent. Red evidence: a request composed the way
  `connsdk.resolveURL` composes one, from base `https://proxy.internal/dockerhub/api/v2` plus
  `/scim/2.0/Users`, took the session-JWT branch.

### Green

- `writes.json`: all 26 paths rewritten to the base-relative `{{ record.<field> }}` form used by the
  other 219 connectors with write actions, matching `streams.json`'s existing `/namespaces/…`
  convention in this same bundle. Every referenced field was already declared in that action's
  `record_schema` (and, where the API omits it from the body, in `path_fields`), so no schema
  changed. Re-resolving through `resolveWriteRequestLine` now yields fully-substituted, single-`/v2`
  URLs for all 26 — e.g. `PATCH https://hub.docker.com/v2/orgs/acme/access-tokens/t9`,
  `DELETE https://hub.docker.com/v2/orgs/acme/groups/devs/members/bob`.
- `hooks.go`: `dualAuth` now matches the base-relative `"/scim/2.0/"` segment pair anywhere in
  `req.URL.EscapedPath()`, so any `base_url` path component keeps routing SCIM to
  `scim_bearer_token`. Matching on the *escaped* path is what keeps the widened match safe:
  `InterpolatePath` urlencodes every resolved segment, so a record value containing a literal
  `/scim/2.0/` arrives percent-encoded and cannot steer a non-SCIM request onto the SCIM credential.
  Widening also fails closed by construction — the dangerous direction is a SCIM route *missing* the
  match. The fail-closed unconfigured-credential branch is unchanged.
- `hooks_test.go`: 3 new tests — `TestAuthenticator_SCIMRoutingSurvivesBaseURLPathOverride` (16
  subtests: 4 `base_url` values × 4 SCIM paths, composed the way `connsdk.resolveURL` composes them,
  each asserting the SCIM token and zero login-endpoint hits),
  `TestAuthenticator_SCIMRoutingFailsClosedUnderBaseURLPathOverride`, and
  `TestAuthenticator_SCIMMatchIsWholeSegmentAndEscapeSafe` (look-alike `/orgs/myscim/2.0/groups`
  and a percent-encoded `evil/scim/2.0/x` path segment both keep the session JWT).

**`operations.json` deliberately keeps its `/v2` prefix.** The review's suggestion to strip it there
too was tested and rejected on evidence: operation `rest.path` is the join key to
`api_surface.json`'s provenance rows, which must stay at Docker Hub's real `/v2/…` paths. Stripping
it produced `connectorgen surface-sync --check` → *"1 connector(s) out of sync … 23 field(s)
divergent, runtime endpoint ledger drift=true"*, which after a sync would fail `validate`'s
`cli_surface_unknown_target` rule and drop every dockerhub row from the operation endpoint ledger.
Operation direct reads do not have the write path's defect: they route through
`normalizeDirectReadPathForBaseURL` (`direct_read.go:507`), which strips `base_url`'s own path
prefix before the request — the same reason `gong` and `notion` carry the prefix. The experiment was
reverted; `operations.json` is unchanged.

### Verification (scoped to the changed area)

- `go build ./...` → clean; `gofmt -l internal/connectors/hooks/dockerhub` → clean.
- `go test ./internal/connectors/hooks/dockerhub/...` → PASS (now 25 tests).
- `go test ./internal/connectors/engine/...` → PASS.
- `go test ./internal/connectors/commandrunner/...` → PASS, including
  `TestEveryImplementedCommandPassesRuntimePreflight` across all 551 connectors.
- `go test ./internal/connectors/conformance/...` → PASS.
- `go test ./cmd/connectorgen/...` → PASS.
- `go run ./cmd/connectorgen validate internal/connectors/defs/dockerhub` → 0 findings.
- `go run ./cmd/connectorgen surface-sync --check` → 551 scanned, 0 corrections, no ledger drift.
- `make docs-check` → `Validated connector docs in docs/connectors`.

No generated artifact needed regeneration: `api_surface.json`, `operations.json`, `cli_surface.json`,
`operation_endpoint_ledger.json`, the website catalogs and `docs/connectors/dockerhub/` all key off
the `/v2`-prefixed documented paths, which did not change.

## Review-fix round 2 — SCIM matcher over-match

The round-1 fix widened the SCIM match to `strings.Contains(EscapedPath, "/scim/2.0/")` so it would
survive a `base_url` path override. Review found that widening reaches too far in the opposite
direction, and it is right: sending the Enterprise SCIM admin token to a non-SCIM endpoint is the
same credential substitution as sending the session JWT to a SCIM one, and `dualAuth` documents that
it performs neither.

### Red

`assign_repository_group` (`writes.json:175`,
`/repositories/{{ record.namespace }}/{{ record.repository }}/groups`) is the bundle's only action
with two adjacent path parameters, and its `record_schema` constrains neither to a pattern — both are
bare `"type": "string"`. `urlencodeSegment` is `url.QueryEscape` with `+`→`%20`
(`interpolate.go:498`), and `QueryEscape` leaves unreserved bytes — alphanumerics and `-_.~` — alone,
so `scim` and `2.0` pass through verbatim as two already-distinct segments. Escaping therefore does
not help here, unlike the `%2F`-encoded case round 1 relied on.

Red evidence: `pm dockerhub repository group assign --namespace scim --repository 2.0 …` builds
`/v2/repositories/scim/2.0/groups`, which satisfies the substring test, so the request is signed with
`scim_bearer_token` instead of the account session JWT.

### Green

`isSCIMRequest` now compares whole segments of the escaped path, requiring the consecutive triple
`scim` / `2.0` / *declared SCIM resource* (`Users`, `Schemas`, `ResourceTypes`,
`ServiceProviderConfig`). The resource segment is the anchor that makes the match positional without
depending on `base_url`'s path component: the segment following the colliding parameter pair is the
route's own literal (`groups`), which is not a SCIM resource. All 9 declared SCIM routes still match;
the fail-closed unconfigured-credential branch is untouched.

The resource allowlist introduces its own drift risk — a SCIM resource added to the bundle but not to
the list would fall through to the session JWT — so the fix is paired with a test that derives the
set from the shipped bundle rather than trusting the constant:

- `TestSCIMResourceSegmentsCoverEveryDeclaredSCIMRoute` loads `defs.FS`'s dockerhub bundle, collects
  every `operations.json`/`writes.json` path carrying the `scim`/`2.0` segment pair, and asserts both
  directions: every declared resource is listed, and every listed resource is actually used.
- `TestAuthenticator_EveryDeclaredSCIMRouteUsesScimBearerToken` drives those same bundle-derived
  paths through the authenticator, composed against `base_url` the way `connsdk.resolveURL` composes
  them, asserting the SCIM token and zero login-endpoint hits — the routing contract asserted against
  shipped declarations rather than hand-written paths.
- `TestAuthenticator_SCIMMatchIsWholeSegmentAndEscapeSafe` gains the collision case
  `/v2/repositories/scim/2.0/groups` (plus a deeper variant), which fails against the round-1
  substring matcher and passes against the segment matcher.

The `isSCIMRequest` doc comment's round-1 claim that "matching wider than necessary fails closed" was
removed — it was wrong. Over-matching emits the wrong credential; neither direction is safe, which is
why the match is now exact rather than deliberately loose.

### Verification (scoped to the changed area)

- `go build ./...` → clean; `gofmt -l internal/connectors/hooks/dockerhub` → clean.
- `go test ./internal/connectors/hooks/dockerhub/... -count=1` → PASS.
- `go run ./cmd/connectorgen validate internal/connectors/defs/dockerhub` → 0 findings.

No bundle JSON changed in this round; the fix is confined to the connector-scoped hook and its tests.
