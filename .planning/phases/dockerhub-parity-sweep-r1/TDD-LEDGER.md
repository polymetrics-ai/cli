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

## Planned RED — namespace override regression (2026-08-08)

Live evidence showed `--config namespace=library` was accepted but did not affect any Docker Hub
ETL route: each stream and the health check read `config.docker_username`. The red test will use
the real declarative engine, not a duplicated template rule:

```text
TestDockerhubNamespaceOverrideDrivesAllStreamAndCheckPaths
  config docker_username=auth-identity, namespace=target-namespace
  engine.Read repositories      => /namespaces/target-namespace/repositories
  engine.Read tags              => /namespaces/target-namespace/repositories/fixture-repository/tags
  engine.Read repository_detail => /namespaces/target-namespace/repositories/fixture-repository
  engine.Read tag_detail        => /namespaces/target-namespace/repositories/fixture-repository/tags/fixture-tag
  engine.Check                  => /namespaces/target-namespace/repositories
```

The committed pre-fix bundle still substitutes `auth-identity`; therefore the focused test must
fail before `spec.json` or `streams.json` is edited. Its failure and command transcript will be
recorded below verbatim, then the test will become the green regression guard after `namespace`
is declared required and all five templates interpolate it.

The engine's credential-boundary validator intentionally checks only constraints expressible over
flat string maps; it does not enforce JSON Schema `required`. The green assertion therefore proves
the relevant runtime safety property instead: a read without `namespace` returns its local
unresolved-configuration error before any HTTP request is sent. This preserves existing shared
validation semantics while ensuring the Docker Hub target is never silently substituted.

### Verbatim RED failure

```text
--- FAIL: TestDockerhubNamespaceOverrideDrivesAllStreamAndCheckPaths (0.01s)
    --- FAIL: TestDockerhubNamespaceOverrideDrivesAllStreamAndCheckPaths/repositories (0.00s)
        namespace_override_test.go:65: request path = "/namespaces/auth-identity/repositories", want namespace override path "/namespaces/target-namespace/repositories"
    --- FAIL: TestDockerhubNamespaceOverrideDrivesAllStreamAndCheckPaths/tags (0.00s)
        namespace_override_test.go:65: request path = "/namespaces/auth-identity/repositories/fixture-repository/tags", want namespace override path "/namespaces/target-namespace/repositories/fixture-repository/tags"
    --- FAIL: TestDockerhubNamespaceOverrideDrivesAllStreamAndCheckPaths/repository_detail (0.00s)
        namespace_override_test.go:65: request path = "/namespaces/auth-identity/repositories/fixture-repository", want namespace override path "/namespaces/target-namespace/repositories/fixture-repository"
    --- FAIL: TestDockerhubNamespaceOverrideDrivesAllStreamAndCheckPaths/tag_detail (0.00s)
        namespace_override_test.go:65: request path = "/namespaces/auth-identity/repositories/fixture-repository/tags/fixture-tag", want namespace override path "/namespaces/target-namespace/repositories/fixture-repository/tags/fixture-tag"
    --- FAIL: TestDockerhubNamespaceOverrideDrivesAllStreamAndCheckPaths/check (0.00s)
        namespace_override_test.go:78: request path = "/namespaces/auth-identity/repositories", want namespace override path "/namespaces/target-namespace/repositories"
FAIL
FAIL	polymetrics.ai/internal/connectors/defs/dockerhub	0.721s
FAIL
```

Command: `go test ./internal/connectors/defs/dockerhub -run
TestDockerhubNamespaceOverrideDrivesAllStreamAndCheckPaths -count=1`. No production bundle file
had changed when this command ran; `namespace_override_test.go` plus this ledger entry are the
committed red state.

### GREEN — distinct target namespace is honored

`namespace_override_test.go` now proves all five runtime routes use an explicitly supplied target
namespace, while an omitted namespace causes the `repositories` stream to fail locally before the
`httptest` server sees any request. The green command was:

```text
go test -timeout 20m ./internal/connectors/defs/dockerhub
ok   polymetrics.ai/internal/connectors/defs/dockerhub  0.757s
```

The adjacent contract gates also passed:

```text
go run ./cmd/connectorgen validate internal/connectors/defs/dockerhub
connectorgen validate: 1 connector(s) checked, 0 findings

go run ./cmd/connectorgen surface-sync --check
connectorgen surface-sync: 551 connector(s) scanned, 0 field(s) filled and 0 field(s) corrected across 0 connector(s)

go test -timeout 20m ./internal/connectors/commandrunner -run TestEveryImplementedCommandPassesRuntimePreflight -count=1
ok   polymetrics.ai/internal/connectors/commandrunner  2.265s

make connector-boundary
go run ./cmd/connectorgen boundary . --json
```

After rebuilding `pm`, `pm docs generate --dir docs/cli` regenerated the Docker Hub manual and
skill. It also exposed 1,027 unrelated baseline documentation changes; those exact non-Docker Hub
generated paths were restored. The retained website generated artifacts were compared as parsed
objects against `HEAD`: both `website/data/connectors.generated.json` and
`website/lib/connectors.catalog.data.generated.json` changed **only** the `dockerhub` object
(`description` and `docs_md`/`docsMd`). No golden transcript changed.

## Live E2E gap — reverse-ETL paths (2026-08-08)

### Reproduction before production edits

Using the real built `pm` binary, a captain-authorized credential in an ignored
local project root, and the required connector-command plan → preview flow:

```text
pm dockerhub repository create --credential dockerhub-live --root <isolated-root> \
  --name <unique-test-repository> --namespace polymetrics --is-private=true
pm dockerhub repository create --root <isolated-root> --plan <plan-id> --preview

- resolved request: POST https://hub.docker.com/v2/v2/namespaces/{namespace}/repositories
```

No approval was used and no external repository mutation was dispatched. The preview
proves both defects mechanically: duplicate `/v2` base/path composition and a raw
OpenAPI placeholder escaping the engine's `{{ record.* }}` interpolator.

### RED — `internal/connectors/defs/dockerhub/write_paths_test.go`

Added the regression test before altering `writes.json` and ran:

```text
go test ./internal/connectors/defs/dockerhub -run TestDockerhubReverseETLWritePathsAreEngineRelativeAndInterpolated -count=1
```

Verbatim failure:

```text
--- FAIL: TestDockerhubReverseETLWritePathsAreEngineRelativeAndInterpolated (0.01s)
    --- FAIL: TestDockerhubReverseETLWritePathsAreEngineRelativeAndInterpolated/create_auth_token (0.00s)
        write_paths_test.go:28: path = "/v2/auth/token", want engine-relative path without the base URL's /v2 prefix
    --- FAIL: TestDockerhubReverseETLWritePathsAreEngineRelativeAndInterpolated/create_user_login (0.00s)
        write_paths_test.go:28: path = "/v2/users/login", want engine-relative path without the base URL's /v2 prefix
    --- FAIL: TestDockerhubReverseETLWritePathsAreEngineRelativeAndInterpolated/create_2fa_login (0.00s)
        write_paths_test.go:28: path = "/v2/users/2fa-login", want engine-relative path without the base URL's /v2 prefix
    --- FAIL: TestDockerhubReverseETLWritePathsAreEngineRelativeAndInterpolated/create_repository (0.00s)
        write_paths_test.go:28: path = "/v2/namespaces/{namespace}/repositories", want engine-relative path without the base URL's /v2 prefix
        write_paths_test.go:31: path = "/v2/namespaces/{namespace}/repositories" retains raw OpenAPI parameter "{namespace}", want a {{ record.* }} template
    --- FAIL: TestDockerhubReverseETLWritePathsAreEngineRelativeAndInterpolated/update_repository_immutable_tags (0.00s)
        write_paths_test.go:28: path = "/v2/namespaces/{namespace}/repositories/{repository}/immutabletags", want engine-relative path without the base URL's /v2 prefix
        write_paths_test.go:31: path = "/v2/namespaces/{namespace}/repositories/{repository}/immutabletags" retains raw OpenAPI parameter "{namespace}", want a {{ record.* }} template
    --- FAIL: TestDockerhubReverseETLWritePathsAreEngineRelativeAndInterpolated/assign_repository_group (0.00s)
        write_paths_test.go:28: path = "/v2/repositories/{namespace}/{repository}/groups", want engine-relative path without the base URL's /v2 prefix
        write_paths_test.go:31: path = "/v2/repositories/{namespace}/{repository}/groups" retains raw OpenAPI parameter "{namespace}", want a {{ record.* }} template
    --- FAIL: TestDockerhubReverseETLWritePathsAreEngineRelativeAndInterpolated/create_access_token (0.00s)
        write_paths_test.go:28: path = "/v2/access-tokens", want engine-relative path without the base URL's /v2 prefix
    --- FAIL: TestDockerhubReverseETLWritePathsAreEngineRelativeAndInterpolated/update_access_token (0.00s)
        write_paths_test.go:28: path = "/v2/access-tokens/{uuid}", want engine-relative path without the base URL's /v2 prefix
        write_paths_test.go:31: path = "/v2/access-tokens/{uuid}" retains raw OpenAPI parameter "{uuid}", want a {{ record.* }} template
    --- FAIL: TestDockerhubReverseETLWritePathsAreEngineRelativeAndInterpolated/delete_access_token (0.00s)
        write_paths_test.go:28: path = "/v2/access-tokens/{uuid}", want engine-relative path without the base URL's /v2 prefix
        write_paths_test.go:31: path = "/v2/access-tokens/{uuid}" retains raw OpenAPI parameter "{uuid}", want a {{ record.* }} template
    --- FAIL: TestDockerhubReverseETLWritePathsAreEngineRelativeAndInterpolated/create_org_access_token (0.00s)
        write_paths_test.go:28: path = "/v2/orgs/{name}/access-tokens", want engine-relative path without the base URL's /v2 prefix
        write_paths_test.go:31: path = "/v2/orgs/{name}/access-tokens" retains raw OpenAPI parameter "{name}", want a {{ record.* }} template
    --- FAIL: TestDockerhubReverseETLWritePathsAreEngineRelativeAndInterpolated/update_org_access_token (0.00s)
        write_paths_test.go:28: path = "/v2/orgs/{org_name}/access-tokens/{access_token_id}", want engine-relative path without the base URL's /v2 prefix
        write_paths_test.go:31: path = "/v2/orgs/{org_name}/access-tokens/{access_token_id}" retains raw OpenAPI parameter "{org_name}", want a {{ record.* }} template
    --- FAIL: TestDockerhubReverseETLWritePathsAreEngineRelativeAndInterpolated/delete_org_access_token (0.00s)
        write_paths_test.go:28: path = "/v2/orgs/{org_name}/access-tokens/{access_token_id}", want engine-relative path without the base URL's /v2 prefix
        write_paths_test.go:31: path = "/v2/orgs/{org_name}/access-tokens/{access_token_id}" retains raw OpenAPI parameter "{org_name}", want a {{ record.* }} template
    --- FAIL: TestDockerhubReverseETLWritePathsAreEngineRelativeAndInterpolated/create_group (0.00s)
        write_paths_test.go:28: path = "/v2/orgs/{org_name}/groups", want engine-relative path without the base URL's /v2 prefix
        write_paths_test.go:31: path = "/v2/orgs/{org_name}/groups" retains raw OpenAPI parameter "{org_name}", want a {{ record.* }} template
    --- FAIL: TestDockerhubReverseETLWritePathsAreEngineRelativeAndInterpolated/replace_group (0.00s)
        write_paths_test.go:28: path = "/v2/orgs/{org_name}/groups/{group_name}", want engine-relative path without the base URL's /v2 prefix
        write_paths_test.go:31: path = "/v2/orgs/{org_name}/groups/{group_name}" retains raw OpenAPI parameter "{org_name}", want a {{ record.* }} template
    --- FAIL: TestDockerhubReverseETLWritePathsAreEngineRelativeAndInterpolated/update_group (0.00s)
        write_paths_test.go:28: path = "/v2/orgs/{org_name}/groups/{group_name}", want engine-relative path without the base URL's /v2 prefix
        write_paths_test.go:31: path = "/v2/orgs/{org_name}/groups/{group_name}" retains raw OpenAPI parameter "{org_name}", want a {{ record.* }} template
    --- FAIL: TestDockerhubReverseETLWritePathsAreEngineRelativeAndInterpolated/delete_group (0.00s)
        write_paths_test.go:28: path = "/v2/orgs/{org_name}/groups/{group_name}", want engine-relative path without the base URL's /v2 prefix
        write_paths_test.go:31: path = "/v2/orgs/{org_name}/groups/{group_name}" retains raw OpenAPI parameter "{org_name}", want a {{ record.* }} template
    --- FAIL: TestDockerhubReverseETLWritePathsAreEngineRelativeAndInterpolated/add_group_member (0.00s)
        write_paths_test.go:28: path = "/v2/orgs/{org_name}/groups/{group_name}/members", want engine-relative path without the base URL's /v2 prefix
        write_paths_test.go:31: path = "/v2/orgs/{org_name}/groups/{group_name}/members" retains raw OpenAPI parameter "{org_name}", want a {{ record.* }} template
    --- FAIL: TestDockerhubReverseETLWritePathsAreEngineRelativeAndInterpolated/remove_group_member (0.00s)
        write_paths_test.go:28: path = "/v2/orgs/{org_name}/groups/{group_name}/members/{username}", want engine-relative path without the base URL's /v2 prefix
        write_paths_test.go:31: path = "/v2/orgs/{org_name}/groups/{group_name}/members/{username}" retains raw OpenAPI parameter "{org_name}", want a {{ record.* }} template
    --- FAIL: TestDockerhubReverseETLWritePathsAreEngineRelativeAndInterpolated/bulk_create_invites (0.00s)
        write_paths_test.go:28: path = "/v2/invites/bulk", want engine-relative path without the base URL's /v2 prefix
    --- FAIL: TestDockerhubReverseETLWritePathsAreEngineRelativeAndInterpolated/cancel_invite (0.00s)
        write_paths_test.go:28: path = "/v2/invites/{id}", want engine-relative path without the base URL's /v2 prefix
        write_paths_test.go:31: path = "/v2/invites/{id}" retains raw OpenAPI parameter "{id}", want a {{ record.* }} template
    --- FAIL: TestDockerhubReverseETLWritePathsAreEngineRelativeAndInterpolated/resend_invite (0.00s)
        write_paths_test.go:28: path = "/v2/invites/{id}/resend", want engine-relative path without the base URL's /v2 prefix
        write_paths_test.go:31: path = "/v2/invites/{id}/resend" retains raw OpenAPI parameter "{id}", want a {{ record.* }} template
    --- FAIL: TestDockerhubReverseETLWritePathsAreEngineRelativeAndInterpolated/update_org_settings (0.00s)
        write_paths_test.go:28: path = "/v2/orgs/{name}/settings", want engine-relative path without the base URL's /v2 prefix
        write_paths_test.go:31: path = "/v2/orgs/{name}/settings" retains raw OpenAPI parameter "{name}", want a {{ record.* }} template
    --- FAIL: TestDockerhubReverseETLWritePathsAreEngineRelativeAndInterpolated/update_org_member (0.00s)
        write_paths_test.go:28: path = "/v2/orgs/{org_name}/members/{username}", want engine-relative path without the base URL's /v2 prefix
        write_paths_test.go:31: path = "/v2/orgs/{org_name}/members/{username}" retains raw OpenAPI parameter "{org_name}", want a {{ record.* }} template
    --- FAIL: TestDockerhubReverseETLWritePathsAreEngineRelativeAndInterpolated/remove_org_member (0.00s)
        write_paths_test.go:28: path = "/v2/orgs/{org_name}/members/{username}", want engine-relative path without the base URL's /v2 prefix
        write_paths_test.go:31: path = "/v2/orgs/{org_name}/members/{username}" retains raw OpenAPI parameter "{org_name}", want a {{ record.* }} template
    --- FAIL: TestDockerhubReverseETLWritePathsAreEngineRelativeAndInterpolated/create_scim_user (0.00s)
        write_paths_test.go:28: path = "/v2/scim/2.0/Users", want engine-relative path without the base URL's /v2 prefix
    --- FAIL: TestDockerhubReverseETLWritePathsAreEngineRelativeAndInterpolated/update_scim_user (0.00s)
        write_paths_test.go:28: path = "/v2/scim/2.0/Users/{id}", want engine-relative path without the base URL's /v2 prefix
        write_paths_test.go:31: path = "/v2/scim/2.0/Users/{id}" retains raw OpenAPI parameter "{id}", want a {{ record.* }} template
    write_paths_test.go:62: create_repository preview warnings = "create_repository executes a live mutation only after approval; dry run performs no external call\\nresolved request: POST https://hub.docker.com/v2/v2/namespaces/{namespace}/repositories", want resolved request "POST https://hub.docker.com/v2/namespaces/polymetrics/fixture-repository"
FAIL
FAIL	polymetrics.ai/internal/connectors/defs/dockerhub	0.746s
FAIL
```

### GREEN — write-path repair

All 26 actions now use engine-relative paths and `{{ record.* }}` path templates;
`create_repository.path_fields` includes `namespace`, so that path value cannot
leak into its JSON body. The exact corrected dry-run request is:

```text
POST https://hub.docker.com/v2/namespaces/polymetrics/repositories
```

Focused GREEN evidence:

```text
=== RUN   TestDockerhubReverseETLWritePathsAreEngineRelativeAndInterpolated
--- PASS: TestDockerhubReverseETLWritePathsAreEngineRelativeAndInterpolated (0.01s)
PASS
ok  	polymetrics.ai/internal/connectors/defs/dockerhub	0.694s
```

Follow-up static checks passed: `connectorgen validate` for Docker Hub reported 0
findings and `surface-sync --check` scanned 551 connectors with 0 corrections. The
first approved live create then exposed the distinct strict-write ALPN defect below;
after that defect was fixed, the corrected provider request reached Docker Hub and
returned HTTP 403 rather than a path or transport error. Full live accounting is in
`VERIFICATION.md`.

## Live E2E gap — strict write transport (2026-08-08)

After the path correction, the captain-authorized `repository create` plan and
preview resolved the expected Docker Hub URL, but the one approved execution failed
before a provider response. The error bytes were HTTP/2 SETTINGS frames received by
an HTTP/1 parser. The persisted plan had no destination `base_url` override and no
record string ending in a quote, so this was isolated as a runtime transport defect,
not a malformed live request. No repository was reported created.

### RED — `internal/connectors/connsdk/http_test.go`

Added `TestRequesterDisableRetriesUsesHTTP1WithHTTP2CapableServer`: an isolated TLS
server advertises HTTP/2, while a `DisableRetries` POST must remain one-shot and use
HTTP/1.1. This exactly models Docker Hub's provider transport shape without another
external mutation. Ran:

```text
go test ./internal/connectors/connsdk -run TestRequesterDisableRetriesUsesHTTP1WithHTTP2CapableServer -count=1 -v
```

Verbatim failure:

```text
=== RUN   TestRequesterDisableRetriesUsesHTTP1WithHTTP2CapableServer
2026/08/08 16:18:20 http2: server: error reading preface from client 127.0.0.1:62140: bogus greeting "POST /mutate HTTP/1.1\r\nH"
    http_test.go:641: Do: send request: Post "https://127.0.0.1:62139/mutate": net/http: HTTP/1.x transport connection broken: malformed HTTP response "\x00\x00\x1e\x04\x00\x00\x00\x00\x00\x00\x05\x00\x10\x00\x00\x00\x03\x00\x00\x00\xfa\x00\x06\x00\x10\x01@\x00\x01\x00\x00\x10\x00\x00\x04\x00\x10\x00\x00"
--- FAIL: TestRequesterDisableRetriesUsesHTTP1WithHTTP2CapableServer (0.00s)
FAIL
FAIL	polymetrics.ai/internal/connectors/connsdk	0.429s
FAIL
```

The root cause is `noReplayClient`: it sets `Transport.Protocols` to HTTP/1 but
retains a caller's TLS `NextProtos` advertisement of `h2`. A TLS server therefore
negotiates HTTP/2 and receives a raw HTTP/1 request. The production fix must clone
the TLS config and pin its ALPN list to `http/1.1`, preserving every other TLS
setting and the no-replay safeguards.

### GREEN

`noReplayClient` now clones any caller TLS config (or supplies a zero-value one)
and sets only `NextProtos` to `[]string{"http/1.1"}` alongside its existing
fresh-connection/no-redirect/no-replay policy. The caller config remains
unmodified. Re-ran the focused reproduction:

```text
=== RUN   TestRequesterDisableRetriesUsesHTTP1WithHTTP2CapableServer
--- PASS: TestRequesterDisableRetriesUsesHTTP1WithHTTP2CapableServer (0.00s)
PASS
ok  	polymetrics.ai/internal/connectors/connsdk	0.356s
```

Then ran `go test ./internal/connectors/connsdk -count=1` → PASS.

## Live read/auth evidence follow-up (2026-08-08)

No production code changed in this follow-up, so the RED/GREEN evidence above
remains the governing TDD record. Rebuilt the same committed `pm` binary and used
the live credential matrix to close the read/auth acceptance vector without logging
credentials, approval tokens, response bodies, or token-derived values.

- **Read + HEAD:** all 28 documented read operations were attempted. Four worked;
  24 failed with explicit provider/local errors recorded row-by-row in
  `VERIFICATION.md`.
- **Authentication:** all three actions completed normal plan → preview → approval
  before a live dispatch using deliberately invalid non-secret fixtures. Each
  provider response was an explicit HTTP 401; no real secret appeared in arguments
  or output.
- **SCIM-only empirical check:** with only `scim_bearer_token` configured and no
  `docker_pat`, every SCIM command path returned a visible failure rather than a
  silent success. Six normal requests returned HTTP 401. The canonical SCIM schema
  URN was rejected locally as `path variable id contains invalid character ':'`; a
  URL-safe sentinel for that same operation reached Docker Hub and returned HTTP
  401. This records observation only and does not change classification.
- **403 behavior:** a fresh `access-tokens list` request exited nonzero with the
  specific redacted error `http 403 for https://hub.docker.com/v2/access-tokens`.
  Permission failures are therefore observable/rejecting, not silent.

## Planned RED — Docker Registry rate-limit declaration (2026-08-08)

The captain's order is a behavior change, so it begins with two failing tests before
any production declaration or rate-limit runtime edit:

1. `internal/connectors/defs/dockerhub/rate_limits_test.go` loads the Docker Hub
   bundle declaration and requires the two registry-only pull policies, 100/200 request
   fixed windows, the documented 21,600-second window, exact `registry-1.docker.io`
   selection, non-secret scopes, and no invented Hub abuse budget.
2. `internal/connectors/connsdk/rate_limit_requester_test.go` requires the requester
   to parse Docker's standard parameterized `ratelimit-limit` and
   `ratelimit-remaining` forms (`200;w=21600`, `199;w=21600`) into typed numeric
   observations, without treating the inline window as a reset time.

The host selector is planned because the generic selector can currently distinguish
endpoint, tier, and auth type but not `registry-1.docker.io` from `hub.docker.com`.
Without it, applying a registry pull budget to any of the 54 Hub management operations
would be a false claim. The green slice must prove the explicit host gate and that
unbudgeted Hub abuse responses still use the existing `Retry-After` path.

### Verbatim RED failure

No production file had changed when the following commands ran:

```text
=== RUN   TestRateLimitObservationParsesDockerWindowedBudgetHeaders
    http_test.go:1237: rateLimitObservation: Docker's windowed headers must be observed
--- FAIL: TestRateLimitObservationParsesDockerWindowedBudgetHeaders (0.00s)
FAIL
FAIL	polymetrics.ai/internal/connectors/connsdk	0.474s
FAIL

=== RUN   TestDockerHubRegistryPullRateLimitsAreEmbedded
    rate_limits_test.go:16: Docker Hub bundle has no embedded provider-cited rate_limits.json
--- FAIL: TestDockerHubRegistryPullRateLimitsAreEmbedded (0.01s)
FAIL
FAIL	polymetrics.ai/internal/connectors/defs/dockerhub	0.728s
FAIL
```

Commands: `go test ./internal/connectors/connsdk -run
TestRateLimitObservationParsesDockerWindowedBudgetHeaders -count=1 -v` and
`go test ./internal/connectors/defs/dockerhub -run
TestDockerHubRegistryPullRateLimitsAreEmbedded -count=1 -v`.

### GREEN — registry-only declaration and pre-transport admission

Added `internal/connectors/defs/dockerhub/rate_limits.json` with the two provider-cited
Registry pull policies only: unauthenticated `100 / 21600s`, scoped by the
non-secret `registry_client_ip`; and authenticated free `200 / 21600s`, scoped by
the non-secret `docker_username`. Both select the exact `registry-1.docker.io`
hostname. Paid Registry access matches no fixed policy, and Docker Hub's separate
unnumbered abuse limiter has no invented budget; the existing requester's HTTP 429
path honors a provider `Retry-After`.

The shared declaration schema now has an exact `hosts` selector. This is not a new
limiter: `Runtime.requesterFor` still resolves every path through the existing
`coordination.RateLimitRegistry`. The Docker header parser retains the leading
numeric value in parameterized fields such as `200;w=21600`, without claiming the
inline window is a reset timestamp.

Focused green evidence:

```text
go test -timeout 20m ./internal/connectors/connsdk ./internal/connectors/engine ./internal/connectors/defs/dockerhub
ok   polymetrics.ai/internal/connectors/connsdk
ok   polymetrics.ai/internal/connectors/engine
ok   polymetrics.ai/internal/connectors/defs/dockerhub

go run ./cmd/connectorgen validate internal/connectors/defs/dockerhub
connectorgen validate: 1 connector(s) checked, 0 findings

go run ./cmd/connectorgen surface-sync --check
connectorgen surface-sync: 551 connector(s) scanned, 0 field(s) filled and 0 field(s) corrected across 0 connector(s)
```

`TestDockerHubRegistryPullPolicyBlocksBeforeTransport` loads the production Docker
Hub bundle, injects a two-request test budget, sends the two allowed requests to a
local transport, then observes the third blocked in admission for the documented
six-hour window and cancels it before any transport dispatch. It also proves the
same declared stream on `hub.docker.com` acquires no Registry pull admission.

Documentation was regenerated with `pm docs generate --dir docs/cli`: the generator
rewrote 1,029 unrelated connector outputs, all restored from the known clean base;
only Docker Hub's `MANUAL.md` and `SKILL.md` remain. `npm run gen:catalog` was run;
parsed-object comparison confirms both website catalog outputs changed only the
`dockerhub` object. No golden transcript changed.

### Binary proof — production budget admission (2026-08-08)

After the green commit, rebuilt `pm` ran `dockerhub tags list --limit 101` through
an isolated loopback HTTP proxy with `base_url=http://registry-1.docker.io/v2` and
the declared unauthenticated profile. Each response supplied a same-host next URL
and one valid record. The proxy received 100 requests; the 101st stayed blocked in
the existing rate-limit admission for five seconds and was then cancelled by the
test harness. The proxy count stayed exactly 100, proving admission happened before
transport dispatch.

Docker's free Registry quota HEAD immediately returned HTTP 200 with
`ratelimit-limit: 100;w=3600` and `ratelimit-remaining: 100;w=3600`; only those
non-secret headers were retained. Thus the provider still had full observed
headroom after local enforcement. The evidence records the header/documentation
window discrepancy and the honest same-PAT limitation in `VERIFICATION.md`.

## Planned RED — canonical Docker Hub SCIM schema URN (2026-08-08)

Captain review found that `scim-schemas get` cannot reach Docker Hub for its documented canonical
schema identifier: `urn:ietf:params:scim:schemas:core:2.0:User`. The existing direct-read path
substitution applies `safety.ValidateIdentifier` to every placeholder and therefore rejects `:`
before it can make a request. This is a correctness defect in our path-variable validation, not an
Enterprise-plan limitation or a missing SCIM credential.

**Red test contract:** `internal/connectors/defs/dockerhub/scim_schema_urn_test.go` will load the
embedded Docker Hub bundle, route it to an `httptest` server with authentication disabled only for
the isolated unit test, and invoke operation `dockerhub.get_scim_schema` with the canonical URN.
It requires exactly one server request to the correctly escaped SCIM path and no query or fragment
injection. Before production code changes it must fail locally with the colon-validation error and
zero server requests. The literal test output is appended below after it runs; this ledger and the
plan were committed before the test/production slice.

**Green contract:** add a dedicated safe URI path-segment validator used by typed operation endpoint
substitution. It must accept the canonical colon-bearing URI identifier, preserve `url.PathEscape`,
and reject separators, traversal, control/dangerous Unicode, query/fragment delimiters, and
ambiguous percent encodings. Engine and Docker Hub tests must prove those boundaries; the fleet
preflight sweep provides the cross-connector execution audit.

### Verbatim RED failure

No production source file had changed when the definition-owned test ran:

```text
=== RUN   TestDockerhubSCIMSchemaGetAcceptsCanonicalURNPathParameter
    scim_schema_urn_test.go:51: read canonical Docker Hub SCIM schema URN: path variable id contains invalid character ':'
--- FAIL: TestDockerhubSCIMSchemaGetAcceptsCanonicalURNPathParameter (0.01s)
FAIL
FAIL	polymetrics.ai/internal/connectors/defs/dockerhub	0.774s
FAIL
```

Command: `go test -timeout 20m ./internal/connectors/defs/dockerhub -run
TestDockerhubSCIMSchemaGetAcceptsCanonicalURNPathParameter -count=1 -v`.

### GREEN — canonical URI path segment reaches the declared Docker Hub route

Added `safety.ValidateURLPathSegment`, deliberately separate from
`ValidateIdentifier`. It permits the existing identifier alphabet plus `:` and rejects all other
characters before `url.PathEscape`: slash/backslash separators, query/fragment delimiters,
percent escapes (therefore no double-encoding ambiguity), whitespace, control/dangerous Unicode,
and traversal. `resolveSurfaceEndpointPath` now uses that new typed segment validator, so the same
safe behavior applies uniformly to declarative direct reads, binary reads, and direct writes that
substitute a one-segment endpoint variable; command/config/field identifiers retain their original
strict rule.

Focused green evidence:

```text
go test -timeout 20m ./internal/safety ./internal/connectors/engine ./internal/connectors/defs/dockerhub -run 'TestValidateURLPathSegment|TestDockerhubSCIMSchemaGetAcceptsCanonicalURNPathParameter|TestDirectRead' -count=1 -v
ok   polymetrics.ai/internal/safety
ok   polymetrics.ai/internal/connectors/engine
ok   polymetrics.ai/internal/connectors/defs/dockerhub
```

The original definition-owned test now observed one local request at
`/v2/scim/2.0/Schemas/urn:ietf:params:scim:schemas:core:2.0:User`; it did not use
a credential or contact Docker Hub. The shared change is therefore proven at the actual Docker
Hub operation boundary rather than merely by a helper test.

### Planned RED — other documented opaque path-segment values

The required cross-connector audit found that the same shared path-value guard is also used for
HubSpot's documented `emailAddress` path parameter (three declared communication-preferences
endpoints). A normal mailbox such as `person+coverage@example.test` is an opaque single segment,
but the first colon-only repair still rejected both `+` and `@`. This is the same class of our
runtime defect, not a HubSpot entitlement result. The red test is deliberately in the shared
safety package because it exercises the exact validator every typed direct read, binary read, and
direct write uses.

**Verbatim RED failure before production changes:**

```text
=== RUN   TestValidateURLPathSegment
=== RUN   TestValidateURLPathSegment/accept_person+coverage@example.test
    safety_test.go:58: ValidateURLPathSegment("person+coverage@example.test") error = test contains invalid path-segment character '+'
--- FAIL: TestValidateURLPathSegment (0.00s)
FAIL
FAIL	polymetrics.ai/internal/safety	0.443s
FAIL
```

Command: `go test -timeout 20m ./internal/safety -run TestValidateURLPathSegment -count=1 -v`.
This red checkpoint is committed before broadening the validator. The green contract permits only
the additional RFC 3986-safe opaque-segment characters required by documented values while
preserving rejection of raw separators, traversal, controls, query/fragment delimiters, and
pre-escaped percent sequences.

### GREEN — documented email segment reaches the engine safely

`ValidateURLPathSegment` now accepts `+` and `@` in addition to the prior documented colon. The
segment remains deliberately narrow: it does not accept raw slash/backslash, percent, query or
fragment delimiters, whitespace, controls, dangerous Unicode, or traversal. The new engine-level
test sends `person+coverage@example.test` through `DirectRead` and observes the complete opaque
segment at the local server, so the proof covers both the shared validator and the actual request
construction route used by declarative operations.

Focused GREEN evidence:

```text
go test -timeout 20m ./internal/safety ./internal/connectors/engine ./internal/connectors/defs/dockerhub -run 'TestValidateURLPathSegment|TestDirectReadAcceptsOpaqueEmailPathSegment|TestDockerhubSCIMSchemaGetAcceptsCanonicalURNPathParameter' -count=1 -v
ok   polymetrics.ai/internal/safety
ok   polymetrics.ai/internal/connectors/engine
ok   polymetrics.ai/internal/connectors/defs/dockerhub
```

Audit result: `encodeSurfacePathValue` is the only provider-value caller of the former strict
identifier validator. Its shared route covers 2,823 typed REST templates (1,303 reads and 1,520
writes) across 14 bundles (out of 25 bundles that carry `operations.json`). Docker Hub contains the only typed direct-operation SCIM/URN
placeholders; HubSpot declares three `emailAddress` placeholders. Other `ValidateIdentifier`
callers validate local connector, credential, command, flag, or configuration names, not provider
path values, and intentionally retain the stricter alphabet.

## Final GREEN verification — reconciled pilot (2026-08-08)

The final current-token ledger check matched all 54 implemented CLI surface commands exactly once:
14 `PROVEN`, 0 `PROVIDER-PLAN-LIMIT`, 31 `PROVIDER-PERMISSION`, and 9
`ENTERPRISE-ONLY`. The three never-dispatched writes carry literal `Named dependency:` reasons in
`VERIFICATION.md`; no row is silently called failed or certified.

Final automated evidence:

```text
go test -timeout 20m ./internal/safety ./internal/connectors/engine ./internal/connectors/defs/dockerhub
ok   polymetrics.ai/internal/safety
ok   polymetrics.ai/internal/connectors/engine
ok   polymetrics.ai/internal/connectors/defs/dockerhub

go test -timeout 20m ./internal/connectors/commandrunner -run TestEveryImplementedCommandPassesRuntimePreflight -count=1
PASS

go test -race -timeout 20m ./internal/safety ./internal/connectors/engine -run 'TestValidateURLPathSegment|TestDirectReadAcceptsOpaqueEmailPathSegment|TestDirectRead' -count=1
PASS

go run ./cmd/connectorgen validate internal/connectors/defs/dockerhub
connectorgen validate: 1 connector(s) checked, 0 findings

go run ./cmd/connectorgen surface-sync --check
connectorgen surface-sync: 551 connector(s) scanned, 0 field(s) filled and 0 field(s) corrected across 0 connector(s)

make agent-contract-check tidy-check lint docs-check smoke-no-build release-workflow-check
PASS

go vet ./... && go build ./cmd/pm && go test -timeout 20m ./internal/cli/...
PASS
```

The rebuilt binary completed 54/54 implemented Docker Hub command-help routes with zero failures;
`pm help dockerhub`, bare `pm dockerhub`, `pm dockerhub scim-schemas get --help`, and connector
docs validation also passed. Inline GSD verify-work and standard code-review completed with the
manual fallback recorded in `VERIFICATION.md` and `REVIEW.md`.

## Corrective slice — automated-review dispositions (2026-08-08)

This slice reworks delivered behaviour in response to automated review. The
historical RED/GREEN entries above are preserved verbatim as the record of what
was done at the time; the corrections below state where that record no longer
describes the delivered head.

### Correction 1 — the strict-write ALPN pin was DROPPED, not delivered

The "Strict-write HTTP/2 regression" RED/GREEN pair above (commits `d72ee21d1`,
`fefaa7251`) describes a `connsdk.noReplayClient` change that pins TLS
`NextProtos` to `http/1.1`, and reports
`TestRequesterDisableRetriesUsesHTTP1WithHTTP2CapableServer` passing.

**Neither is in the delivered head.** Both commits were resolved in favour of
`main` during the rebase onto current `origin/main` and now carry planning-doc
changes only. `internal/connectors/connsdk/http.go`'s `noReplayClient` clones
the transport and sets `DisableKeepAlives`; it never sets `Transport.Protocols`,
and no test by that name exists in the tree.

Disposition: the fix is **not needed against current `main`**. The defect it
repaired required `noReplayClient` to force HTTP/1 at the transport layer while
leaving an `h2` ALPN advertisement in the TLS config; main's `noReplayClient`
sets no `Transport.Protocols` at all, so the mismatch is unreachable. No
shared `connsdk` production change ships in this PR. `PLAN.md`'s reference to
"the shared strict-write ALPN fix and its isolated definition-owned regression
test" as delivered is corrected by this entry.

### Correction 2 — the old hook tests did not prove independent SCIM routing

The SCIM-only empirical check recorded under "Live read/auth evidence follow-up"
attributed six SCIM `HTTP 401`s to the provider. That reading was wrong about
the local cause, and the hook tests that existed then could not have caught it:
every SCIM fixture configured `docker_pat` **and** `scim_bearer_token`, so none
of them exercised a SCIM-only connection.

RED (reasoned, from the delivered head): `streams.json` gated the ONLY custom
auth spec on `when: "{{ secrets.docker_pat }}"`. A connection with just
`scim_bearer_token` matched no custom spec, fell through to `{ "mode": "none" }`,
and sent every `/v2/scim/2.0/**` request with no `Authorization` header at all —
contradicting `docs.md` and `spec.json`, both of which state the SCIM token
alone authenticates the SCIM commands. A second defect sat behind it: `dualAuth`
routed on a fixed `/v2/scim/2.0/` prefix, so a proxy `base_url` carrying its own
path prefix failed **open** and signed SCIM requests with the account session
JWT — the exact substitution the type's doc comment says cannot happen.

GREEN: `streams.json` declares a second custom auth spec gated on
`{{ secrets.scim_bearer_token }}` (the `when` grammar has no OR operator, so one
spec cannot express "either secret"). `Hooks.Authenticator` treats a spec with
no login fields as a SCIM-only connection and builds a `dualAuth` with a nil
session, which fails closed naming `docker_pat` on every non-SCIM path rather
than sending the SCIM token to a `bearerAuth` endpoint. SCIM prefixes are
derived from the resolved `base_url` path (both the base-relative write form and
the unstripped declared direct-read form), so proxy routing fails closed instead
of open, while an anchored match keeps a repository path that merely *contains*
`/scim/2.0/` on the session JWT.

```text
go test -timeout 20m ./internal/connectors/hooks/dockerhub -count=1
ok   polymetrics.ai/internal/connectors/hooks/dockerhub
```

New fixtures: `TestAuthenticator_SCIMOnlyConnectionAuthenticatesSCIMRequests`,
`TestAuthenticator_SCIMOnlyConnectionFailsClosedOnNonSCIMPath`,
`TestAuthenticator_NoCredentialConfiguredIsError`,
`TestAuthenticator_SCIMRoutingHonorsProxyBaseURLPathPrefix`,
`TestAuthenticator_ProxyBaseURLNonSCIMPathStillUsesSessionJWT`,
`TestAuthenticator_RepositoryPathNamedSCIMUsesSessionJWT`.

### Correction 3 — the three auth-exchange commands are NOT shipped

Captain decision [key=auth-secrets]. The entries above record `auth token
create`, `auth login create`, and `auth 2fa-login create` as implemented with
"redacted token output". That claim was not safe: each command's only input is a
live credential supplied as a plain CLI flag mapped to a record field, and the
reverse-ETL path it runs on persists that record to the project state file as
plaintext `connector_command_record` and echoes it in plan output. An action's
`redact_fields` names a RESPONSE field and never redacts the request credential,
so nothing in the delivered design made a password, PAT, or TOTP code safe in
argv or on disk.

Disposition: the three write actions are removed from `writes.json`; their
`cli_surface.json` commands stay in the 54-row surface as
`availability: "planned"` with no flags, examples, or executable target; and
their `api_surface.json` rows become blocked `sensitive_reverse_etl` rows whose
`notes` carry the machine-checkable dependency
`named_dependency=requires secure secret input (stdin/env/vault-reference),
encrypted or ephemeral plan storage, and a secure sink for the returned token`.
`cmd/connectorgen/dockerhub_api_surface_test.go` already asserted that
`named_dependency=` prefix for every blocked row; its target counts move from
54 covered / 0 blocked to **51 covered / 3 blocked**. The session-login exchange
itself is unaffected — the `dockerhub` AuthHook still performs it internally,
with no plan record, whenever `docker_pat` is configured.

### RED/GREEN — status-only HEAD checks must be able to report absence

Captain decision [key=head-semantics].

RED: `TestOperationDirectReadHEADNonSuccessStatusIsError` pinned a 404 HEAD
response as a transport error, so `pm dockerhub repository check` could only
ever return `status_code: 200`. The one question an existence check exists to
answer — the repository is not there — exited non-zero with an error string and
no status code.

GREEN: `statusOnlyAbsenceResponse` converts **only** a 404 and **only** for a
HEAD request into the documented `{"status_code": 404}` result. 401/403 (an auth
problem, not a fact about the resource), 429/5xx (the provider declining to
answer), and every non-HEAD direct read keep failing exactly as before. Global
HTTP behaviour is untouched: a GET 404 is still an error.

```text
go test -timeout 20m ./internal/connectors/engine -run 'TestOperationDirectReadHEAD|TestOperationDirectReadGETNotFound' -count=1
ok   polymetrics.ai/internal/connectors/engine
```

Replacing/added: `TestOperationDirectReadHEADAbsenceReturnsStructuredNotFound`,
`TestOperationDirectReadHEADNonAbsenceStatusesStayErrors`,
`TestOperationDirectReadGETNotFoundIsStillError`.

### RED/GREEN — add-time admission for incomplete credentials

Captain decision [key=namespace-compat]. `namespace` stays required and is
never silently defaulted to `docker_username`; the gap was that the requirement
was enforced only at read/check time.

RED: `internal/connectors/defs/dockerhub/credential_admission_test.go` —
`pm credentials add` accepted a Docker Hub credential with no `namespace` (and
with no config at all), because `engine.Schema.ValidateConfiguration`
deliberately skipped JSON Schema `required`. The operator learned what was
missing only later, from a connector-internal template error at read time.

```text
--- FAIL: TestDockerhubCredentialAdmissionRejectsIncompleteConfig
    credential_admission_test.go: ValidateConfiguration(map[]) = nil, want an
    incomplete-credential rejection before the credential can be saved
```

GREEN: `ValidateConfiguration` now enforces required-key presence for the keys
this boundary can actually see — declared required, non-secret (a required
secret lives in the separate secrets map), and carrying no `default` (which the
engine materializes for the caller). Supplied values are still checked first, so
a caller who types a rejected value is told what is wrong with what they typed
rather than being handed a different key's absence. The rule is derived from
each bundle's own `spec.json`, so it is declarative, not connector-specific
policy in shared code.

```text
go test -timeout 20m ./internal/connectors/engine ./internal/connectors/defs/dockerhub ./internal/app -count=1
ok   polymetrics.ai/internal/connectors/engine
ok   polymetrics.ai/internal/connectors/defs/dockerhub
ok   polymetrics.ai/internal/app
```

`TestSchemaWithoutConfigurationConstraintsIsNotAdvertised` pinned the previous
"required is not a configuration constraint" contract; its fixture is narrowed
to a genuinely constraint-free schema and the new contract gets its own
coverage in `TestSchemaRequiredConfigurationKeyIsAdvertisedAndEnforced` and
`TestSchemaRequiredSecretAndDefaultedKeysAreNotConfigurationConstraints`.

### Scope corrections

- `CLAUDE.md` is restored to its prior regular-file content; the change that
  replaced it with a symlink to `AGENTS.md` is reverted as out of scope for this
  phase (Captain Decision [key=doc-drift-scope]).
- The 13-line `AGENTS.md` HEAD-capability addition is reverted for the same
  reason. The capability itself remains documented where it is owned, in
  `internal/connectors/defs/dockerhub/docs.md`.
- Generated docs were regenerated into a temporary directory and byte-compared.
  Only Docker Hub's records differ: `docs/connectors/README.md`,
  `docs/connectors/catalog/all-connectors.{json,md}`,
  `docs/connectors/dockerhub/{MANUAL,SKILL}.md`,
  `website/data/connectors.generated.json`, and
  `website/lib/connectors.catalog.data.generated.json`. The pre-existing stale
  `warehouse` catalog description is deliberately left untouched. No generator
  or drift validator was added or broadened.

## Rebase and captain-auth correction (2026-08-09)

The prior no-mistakes run was cancelled before the captain-ordered rebase. Its
native repair worker was stopped, its changing isolated worktree was checked
for whitespace errors, and its complete tracked delta was preserved as an
ordinary checkpoint commit before cancellation. No shared daemon was stopped,
restarted, or modified. The checkpoint was restored on this branch before
rebasing; the pre-rebase restored head was `e4900724d0d65387047554f119a0a30958d035b2`,
fetched `origin/main` was `d453fbe256eb22d90ea77dbed634b245bd6e795b`, and the
rebased head has that fetched main as its merge base. The cancelled pipeline
result is therefore not used as validation for this head.

### Correction 4 — captain override restores the three auth exchanges

Correction 3 above remains historical evidence of the original security
finding, but Captain subsequently overruled its delivered disposition. The
three exchanges now remain implemented in the 54-operation surface with four
specific safeguards: a generated plaintext-retention warning in help and at
execution, `redact_fields` for `password`, `secret`, `login_2fa_token`, and
`code`, `--from-env` and `--value-stdin` inputs, and the typed direct-write
response body returned unchanged after approved execution.

### RED — restored command metadata did not reach runtime help

The checkpoint added `accepts_secret_input` to Docker Hub's JSON but omitted
the `engine.CLICommand` → `connectors.CommandSurfaceCommand` mapping. The
red test was added and run before repairing that mapping:

```text
--- FAIL: TestDockerHubAuthHelpGeneratesCredentialRetentionWarning (1.20s)
    cli_test.go:97: Docker Hub auth help missing "Warning: supplied credential values are written to plaintext local project state and retained for this command plan."
FAIL
FAIL	polymetrics.ai/internal/cli
```

Command: `go test -timeout 20m ./internal/cli -run
TestDockerHubAuthHelpGeneratesCredentialRetentionWarning -count=1`.

### GREEN — safe input, redacted plan sample, and unredacted response

`engine.commandSurface` now preserves `AcceptsSecretInput`; the CLI derives the
same warning string into both command help and point-of-use stderr, rather than
relying on a hand-authored Docker Hub note. `commandrunner` creates a genuinely
redacted plan sample only for declared secret-input commands while preserving
the complete execution record required by the explicitly warned local plan.
The copier also preserves an explicitly empty string array rather than changing
it to JSON `null`. A Docker Hub loopback test proves the login operation sends
no inherited connector authorization, redacts its password in the stored
operator-visible sample, retains the execution record after the warning, and
returns the provider token response unchanged. A CLI test proves environment
input never appears in JSON/stderr, and a helper-level test proves stdin input
is newline-trimmed and mapped to the declared secret field.

```text
go test -timeout 20m ./internal/cli -run 'TestApplySensitiveCommandInputsReadsValueStdin|TestDockerHubAuth' -count=1
ok   polymetrics.ai/internal/cli

go test -timeout 20m ./internal/app -run TestDockerHubAuthLoginPlanRedactsCredentialInputAndReturnsProviderToken -count=1
ok   polymetrics.ai/internal/app

go test -timeout 20m ./internal/connectors/engine ./internal/connectors/connsdk ./internal/connectors/hooks/dockerhub ./internal/connectors/defs/dockerhub ./internal/connectors/commandrunner ./internal/app
PASS

go run ./cmd/connectorgen validate internal/connectors/defs/dockerhub
connectorgen validate: 1 connector(s) checked, 0 findings

go run ./cmd/connectorgen surface-sync --check
connectorgen surface-sync: 551 connector(s) scanned, 0 field(s) filled and 0 field(s) corrected across 0 connector(s)
```

The generated documentation was rerun into a disposable directory. The
generated Docker Hub `MANUAL.md` and `SKILL.md` were copied byte-for-byte from
that output; parsed-object comparison found the Docker Hub row in both
`docs/connectors/catalog/all-connectors.json` and `.md` already identical, so
the unrelated stale warehouse row remains untouched. The website generators
changed exactly one object (`dockerhub`) in each generated JSON catalog.

### RED — Docker Hub generated security guidance overstated redaction

The Docker Hub manual claimed both that secret values must never be passed as
shell arguments and that provider-issued tokens were redacted from command
output. The first statement contradicted the intentionally retained
credential flags, and the second contradicted the runtime contract that
provider responses remain unchanged. This Docker Hub-only guide test was added
and run before changing the guide renderer or documentation sources:

```text
--- FAIL: TestDockerHubGuideExplainsSecretInputAndUnredactedTokenResponses
    registry_test.go:181: Docker Hub manual missing "For auth credential fields, prefer --from-env or --value-stdin over command-line flags; argv can be observed by other local processes and shell history."
FAIL
FAIL	polymetrics.ai/internal/connectors/bundleregistry
```

Command: `go test -timeout 20m ./internal/connectors/bundleregistry -run
TestDockerHubGuideExplainsSecretInputAndUnredactedTokenResponses -count=1`.

### GREEN — connector-scoped guidance states the real boundary

The guide renderer has a Docker Hub-only security branch, following the
existing connector-specific guide pattern. It tells callers to use
`--from-env` or `--value-stdin` rather than credential argv flags and says
that runtime provider responses remain unchanged. Docker Hub's authored
`docs.md`, `writes.json`, and `cli_surface.json` now make the same distinction:
plan samples redact declared secret *input* fields, while returned provider
tokens are output that operators must treat as secret material. No other
connector guide changes as a result.

```text
go test -timeout 20m ./internal/connectors/bundleregistry -run TestDockerHubGuideExplainsSecretInputAndUnredactedTokenResponses -count=1
ok  	polymetrics.ai/internal/connectors/bundleregistry
```

### RED/GREEN — restored pagination slice lint correction

The captain-rebase checkpoint also carried the Docker Hub operation-pagination
support. The first non-concurrent `make lint` run found a branch-local
staticcheck failure in its shared helper:

```text
internal/connectors/engine/direct_read_paginate.go:708:12: ST1023: should omit type any from declaration; it will be inferred from the right-hand side (staticcheck)
	var value any = decoded
	          ^
```

The declaration is now `value := decoded`, preserving its boundary type while
letting Go infer it. The green lint rerun reported zero issues:

```text
make lint
0 issues.
```

### Correction 5 — Docker Hub admission is connector-specific, not generic required-key enforcement

The historical GREEN narrative under “add-time admission for incomplete
credentials” says that `engine.Schema.ValidateConfiguration` began enforcing
JSON Schema `required` keys. That is **not** the delivered behavior and must
not be read as one. Required-only schema keys deliberately remain outside the
shared flat-string configuration validator; the current engine tests pin that
contract.

The delivered add-time safeguard is narrowly scoped in
`App.validateCredentialConfig`: when `connector == "dockerhub"`, a missing or
blank `namespace` rejects `credentials add` before the credential can be
saved. It does not substitute `docker_username`, and it does not change
admission for other connectors. `TestAddCredentialDockerhubNamespaceAdmissionIsConnectorSpecific`
proves rejection before persistence, successful namespace-only Docker Hub
admission, and unchanged non-Docker-Hub admission; the Docker Hub definition
and engine tests separately prove that required-only schema validation remains
non-global.

```text
go test -timeout 20m ./internal/app ./internal/connectors/engine ./internal/connectors/defs/dockerhub -count=1
ok   polymetrics.ai/internal/app
ok   polymetrics.ai/internal/connectors/engine
ok   polymetrics.ai/internal/connectors/defs/dockerhub
```

### GREEN — rebased final local validation

The cancelled pre-rebase pipeline result is not carried forward. After the
checkpoint restore and verified merge base, the repaired head passed its focused
Docker Hub/app/engine/CLI regressions, package vet/build, Docker Hub validation,
surface synchronization, implemented-command preflight, docs and website checks.
The freshly built binary resolved every declared implemented path in nine short
batches, avoiding the terminal wrapper's detached-long-command behavior:

```text
Docker Hub help batch 1-6/54: 6 routes reachable
Docker Hub help batch 7-12/54: 6 routes reachable
Docker Hub help batch 13-18/54: 6 routes reachable
Docker Hub help batch 19-24/54: 6 routes reachable
Docker Hub help batch 25-30/54: 6 routes reachable
Docker Hub help batch 31-36/54: 6 routes reachable
Docker Hub help batch 37-42/54: 6 routes reachable
Docker Hub help batch 43-48/54: 6 routes reachable
Docker Hub help batch 49-54/54: 6 routes reachable; help topic and bare namespace passed
```

Key gate evidence:

```text
go run ./cmd/connectorgen validate internal/connectors/defs/dockerhub
connectorgen validate: 1 connector(s) checked, 0 findings

go run ./cmd/connectorgen surface-sync --check
connectorgen surface-sync: 551 connector(s) scanned, 0 field(s) filled and 0 field(s) corrected across 0 connector(s)

go test -timeout 20m ./cmd/connectorgen -run TestDockerhubAPISurfaceOperationLedger -count=1
ok   polymetrics.ai/cmd/connectorgen

go test -timeout 20m ./internal/connectors/commandrunner -run TestEveryImplementedCommandPassesRuntimePreflight -count=1
ok   polymetrics.ai/internal/connectors/commandrunner

make lint
0 issues.
```
