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
