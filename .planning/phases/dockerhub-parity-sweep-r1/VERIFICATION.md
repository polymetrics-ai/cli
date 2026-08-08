# Docker Hub parity — live E2E verification

## Result

Live E2E completed as far as the supplied Docker Hub account permits. The repaired
binary reached every one of the 54 documented command routes; 23 operations made a
real provider request: **4 worked and 19 failed**. The remaining **31 were
untestable** without repeating a proven account authorization failure, creating an
unrecoverable account object, or passing a secret as a raw CLI argument.

| Metric | Count |
| --- | ---: |
| Documented/implemented operations | 54 |
| Live-exercised (`worked + failed`) | 23 |
| Worked | 4 |
| Failed | 19 |
| Untestable, with concrete reason below | 31 |
| Binary routes reachable with `--help` | 54/54 |

The bare `pm dockerhub` namespace command also rendered contextual help and exited
successfully. The isolated project root and all E2E output remain ignored; no secret,
approval token, response body, or token-derived value was printed or committed.

## Mandatory private-repository/image chain

The live `repository create` plan and preview resolved exactly:

```text
POST https://hub.docker.com/v2/namespaces/polymetrics/repositories
```

The single approval-gated dispatch returned the following redacted provider error:

```text
pm dockerhub repository create --credential dockerhub-live --name <unique-name> \
  --namespace polymetrics --is-private=true --plan <plan-id> --approve <approval>
error: dockerhub action=create_repository record=0: http 403 for https://hub.docker.com/v2/namespaces/polymetrics/repositories: [redacted]
```

No repository was reported created, and the successful account repository listing
returned `count=0`. There was therefore no private repository to inspect, no image
to push, and no image/tag/manifest to read back. This is an account/PAT permission
limit after the connector reached the correct endpoint, not a connector transport or
path defect. No external cleanup is required because no external object was created.

## Worked live operations

All output bodies were discarded after success; only the command outcome was kept.

| Operation | Exact live command shape | Result |
| --- | --- | --- |
| `repositories list` | `pm dockerhub repositories list --credential dockerhub-live --limit 1` | success, zero account repositories |
| `repository check` | `pm dockerhub repository check --namespace library --repository alpine --credential dockerhub-live` | success (status-only HEAD) |
| `repository tags check` | `pm dockerhub repository tags check --namespace library --repository alpine --credential dockerhub-live` | success (status-only HEAD) |
| `repository tag check` | `pm dockerhub repository tag check --namespace library --repository alpine --tag latest --credential dockerhub-live` | success (status-only HEAD) |

## Live failures

`[redacted]` means the provider body was intentionally not retained. The HTTP
status is the exact observed error classification, and every command below used the
isolated project root and `--json` with normal output discarded.

| Operation | Exact command shape | Exact observed error | Disposition |
| --- | --- | --- | --- |
| `repository detail list` | `pm dockerhub repository detail list --config namespace=library --config repository=alpine --credential dockerhub-live --limit 1` | HTTP 404 | Account has no `polymetrics/alpine`; stream intentionally uses configured `docker_username` as its namespace. |
| `tags list` | `pm dockerhub tags list --config namespace=library --config repository=alpine --credential dockerhub-live --limit 1` | HTTP 404 | Same missing account repository. |
| `tag detail list` | `pm dockerhub tag detail list --config namespace=library --config repository=alpine --config tag=latest --credential dockerhub-live --limit 1` | HTTP 404 | Same missing account repository. |
| `repository immutable-tags verify` | `pm dockerhub repository immutable-tags verify --namespace library --repository alpine --regex '.*' --credential dockerhub-live` | HTTP 403 | PAT lacks this account-scoped permission. |
| `access-tokens list` | `pm dockerhub access-tokens list --credential dockerhub-live` | HTTP 403 | PAT/account scope limit. |
| `access-tokens get` | `pm dockerhub access-tokens get --uuid 00000000-0000-0000-0000-000000000000 --credential dockerhub-live` | HTTP 403 | PAT/account scope limit. |
| `audit-logs actions list` | `pm dockerhub audit-logs actions list --account polymetrics --credential dockerhub-live` | HTTP 403 | PAT/account scope limit. |
| `audit-logs list` | `pm dockerhub audit-logs list --account polymetrics --credential dockerhub-live` | HTTP 403 | PAT/account scope limit. |
| `groups list` | `pm dockerhub groups list --org-name polymetrics --credential dockerhub-live` | HTTP 403 | PAT/account scope limit. |
| `groups get` | `pm dockerhub groups get --org-name polymetrics --group-name pm-e2e-never-created --credential dockerhub-live` | HTTP 403 | PAT/account scope limit. |
| `groups members list` | `pm dockerhub groups members list --org-name polymetrics --group-name pm-e2e-never-created --credential dockerhub-live` | HTTP 403 | PAT/account scope limit. |
| `invites list` | `pm dockerhub invites list --org-name polymetrics --credential dockerhub-live` | HTTP 403 | PAT/account scope limit. |
| `org access-tokens list` | `pm dockerhub org access-tokens list --name polymetrics --credential dockerhub-live` | HTTP 403 | PAT/account scope limit. |
| `org access-tokens get` | `pm dockerhub org access-tokens get --org-name polymetrics --access-token-id 00000000-0000-0000-0000-000000000000 --credential dockerhub-live` | HTTP 403 | PAT/account scope limit. |
| `org members list` | `pm dockerhub org members list --org-name polymetrics --credential dockerhub-live` | HTTP 403 | PAT/account scope limit. |
| `org settings get` | `pm dockerhub org settings get --name polymetrics --credential dockerhub-live` | HTTP 403 | PAT/account scope limit. |
| `org members export` | `pm dockerhub org members export --org-name polymetrics --dest-root <isolated-download-root> --file-name members.csv --credential dockerhub-live` | HTTP 403 | PAT/account scope limit. |
| `scim-service-provider-config get` | `pm dockerhub scim-service-provider-config get --credential dockerhub-scim-only` | HTTP 401 | A credential containing only `scim_bearer_token` was used; the supplied PAT is not accepted as a SCIM bearer credential. |
| `repository create` | command shown in the mandatory chain above | HTTP 403 | PAT/account scope limit; no retry performed. |

## Untestable operations (31)

| Operations | Count | Machine-checkable reason for not dispatching |
| --- | ---: | --- |
| `scim-resource-types get/list`, `scim-schemas get/list`, `scim-users get/list` | 6 | `Named dependency: valid Docker Hub organization SCIM bearer token`; the SCIM-only empirical request returned HTTP 401, so repeating the same auth failure would add no coverage. |
| `auth token create`, `auth login create`, `auth 2fa-login create` | 3 | `Named dependency: connector-command secret-reference/from-env input`; their required secret values are raw action flags (`--secret`, `--password`, or a one-time 2FA value), which cannot be supplied without violating the no-secret-in-arguments rule. Credential validation already proved the connector's protected login hook itself succeeds. |
| `access-tokens create/delete/update`; `groups create/delete/members add/members remove/replace/update`; `invites bulk-create/cancel/resend`; `org access-tokens create/delete/update`; `org members remove/update`; `org settings update` | 18 | `Named dependency: Docker Hub PAT with account/organization write scope`; matching read endpoints consistently returned HTTP 403, and no disposable target can be safely created/cleaned with this credential. |
| `repository group assign`, `repository immutable-tags update` | 2 | `Named dependency: writable Docker Hub repository in polymetrics`; repository creation returned HTTP 403 and account listing is empty. |
| `scim-users create/update` | 2 | `Named dependency: valid Docker Hub organization SCIM bearer token`; same SCIM-only account limit as the six SCIM reads. |

The SCIM-only request establishes that the connector did not silently succeed through
the normal session credential: that credential contained no `docker_pat`. The 401
alone cannot distinguish an invalid bearer token from an omitted bearer header, so
this evidence does **not** change the existing SCIM classification or claim a new
runtime defect.

## TDD and local verification

- Captured RED before production changes for the provider-rooted `/v2` paths and
  raw OpenAPI placeholders in every Docker Hub write action.
- `TestDockerhubReverseETLWritePathsAreEngineRelativeAndInterpolated` → PASS for
  all 26 actions after converting them to engine-relative `{{ record.* }}` paths.
- Captured a second RED from an HTTP/2-capable local TLS server, then fixed
  strict-write ALPN pinning in `connsdk`; focused test and full connsdk package → PASS.
- `go run ./cmd/connectorgen validate internal/connectors/defs/dockerhub` → 0 findings.
- `go run ./cmd/connectorgen surface-sync --check` → 551 connectors scanned, 0 corrections.
- Rebuilt `./pm`, then ran all 54 `pm dockerhub <command> --help` routes: 54/54 exit 0.

## Pilot friction points

- The derived write surface copied provider-rooted `/v2` paths and raw OpenAPI
  placeholders into an engine whose base URL already ends in `/v2`; previewing a
  real command was the first test that exposed both defects.
- The one-shot write transport had an untested HTTP/2 ALPN edge case. Local
  `httptest` reproduction converted a provider-only symptom into a narrow,
  provider-neutral runtime fix.
- The supplied PAT authenticates and supports public/HEAD paths but not repository
  creation or account/organization APIs, preventing the required image-push chain
  and all safe mutation cleanup.
- Full route-help verification must be chunked in this execution environment; one
  54-process sweep exceeds a single command window even though all routes pass.

## GSD inline verify-work and code-review

Resolved the adapter sources for `verify-work` and `code-review` with
`scripts/gsd sources <command>`, then read and followed the generated prompts. The
canonical parent-worker contract forbids spawning the requested GSD verifier/reviewer
roles, so both phases use the documented inline/manual fallback.

- **verify-work verdict:** code and command-route verification pass; live acceptance
  remains blocked only on the supplied account's missing repository-write,
  organization, and SCIM scopes, documented above with exact operation accounting.
- **code-review scope:** `internal/connectors/connsdk/http.go`, its local TLS/HTTP2
  regression test, Docker Hub `writes.json`, the Docker Hub write-path regression
  test, and phase evidence.
- **code-review findings:** no Critical, Warning, or Info findings. The strict
  transport clones rather than mutates caller TLS configuration; all provider paths
  are tested as engine-relative and interpolated; no secret appears in source,
  output, or evidence.
- **additional review evidence:** `go test -race ./internal/connectors/connsdk -run
  TestRequesterDisableRetriesUsesHTTP1WithHTTP2CapableServer -count=1` → PASS;
  `git diff --check` → clean.
