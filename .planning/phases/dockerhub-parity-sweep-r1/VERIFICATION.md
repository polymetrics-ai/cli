# Docker Hub parity — live E2E verification

## Result

The write-scoped continuation supersedes the initial read-only account result below.
The repaired binary reached every documented Docker Hub command route; each count
here is **one final classification per one of the 54 artifact operations**, so repeat
probes are not double-counted. **50/54 operations made a real provider request: 10
worked and 40 returned a specific, surfaced error. Four are deliberately untestable
with machine-checkable named dependencies.**

| Metric | Count |
| --- | ---: |
| Documented/implemented operations | 54 |
| Live-exercised (`worked + failed`) | 50 |
| Worked | 10 |
| Failed with an explicit nonzero result | 40 |
| Untestable, with concrete reason below | 4 |
| Binary routes reachable with `--help` | 54/54 |

The bare `pm dockerhub` namespace command also rendered contextual help and exited
successfully. The isolated project root and all E2E output remain ignored; no secret,
approval token, response body, or token-derived value was printed or committed.

### Final per-operation live ledger

`[redacted]` is the intentionally discarded provider body. It is part of the exact
observed `pm` error string, not a rewritten success or a hidden retry.

| Final result | Operations | Count |
| --- | --- | ---: |
| Worked | `repositories list`; `repository detail list` (live `--config namespace=library` proof); `tags list`; `tag detail list`; `repository check`; `repository tags check`; `repository tag check`; `repository immutable-tags verify`; `repository create`; `repository immutable-tags update` | 10 |
| `http 403 for <resolved Docker Hub URL>: [redacted]` | `access-tokens list/get/create/update/delete`; `audit-logs actions list/list`; `groups list/get/members list/create/replace/update/delete/members add/members remove`; `invites list/bulk-create/cancel`; `org access-tokens list/get/update/delete`; `org members list/update/remove/export`; `org settings get`; `repository group assign` was not in this group (see HTTP 400); all listed 403s exited nonzero with the concrete endpoint in the `pm` error | 28 |
| `http 401 for <resolved Docker Hub URL>: [redacted]` | `scim-resource-types list/get`; `scim-schemas list/get`; `scim-service-provider-config get`; `scim-users list/get`; `auth token create`; `auth login create`; `auth 2fa-login create` | 10 |
| `http 400 for https://hub.docker.com/v2/repositories/polymetrics/pm-e2e-20260808-173040/groups: [redacted]` | `repository group assign` with deliberately invalid `group_id=0`, after plan/preview/approval | 1 |
| `http 405 for https://hub.docker.com/v2/invites/0/resend: [redacted]` | `invites resend` with deliberately nonexistent `id=0`, after plan/preview/approval | 1 |
| Untestable | `org access-tokens create` — `Named dependency: least-privilege typed reverse-ETL source record with resources[] plus Docker Hub organization-token permission`; `org settings update` — `Named dependency: Docker Hub organization settings read/write permission to preserve the current values`; `scim-users create/update` — `Named dependency: valid Docker Hub Enterprise organization SCIM bearer token` | 4 |

The prior SCIM-only probe remains valid: only `scim_bearer_token` was configured;
every dispatched SCIM request was nonzero (HTTP 401). The canonical colon-bearing
schema URN first hit local path validation, then a URL-safe sentinel reached Docker
Hub and returned the recorded 401. This does not prove the bearer is present on the
wire, so the classification was deliberately not changed.

### Mandatory private-repository/image chain — completed

Created and intentionally retained private repository:

```text
polymetrics/pm-e2e-20260808-173040
```

The approval-gated `repository create` run completed (`records_succeeded=1`), and
`pm dockerhub repository detail list` returned `is_private: true` **before** any
image transfer. `docker` then pushed the zero-layer `e2e` image through Docker's
Registry v2 protocol. `docker manifest inspect` confirmed a schema-2 image manifest;
`pm dockerhub tags list`, `tag detail list`, `repository detail list`, and both tag
HEAD checks read the resulting repository/tag state successfully. The repository was
not deleted.

`pm` cannot drive the image push or OCI-manifest read itself. This is an honest,
out-of-surface gap: neither appears among the Docker Hub OpenAPI artifact's 54
management operations or the connector's command surface; image transfer is Docker
Registry v2 protocol work and needs its own declared connector/protocol foundation.

The namespace repair is also live-proven: a fresh
`repository detail list --config namespace=library --config repository=alpine`
returned the public `library/alpine` record (rather than the prior misleading
`polymetrics/alpine` 404). The regression test covers the remaining three streams
and the check route with distinct auth identity and target namespace values.

## Historical initial read-only private-repository attempt (superseded)

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
isolated project root and `--json` with normal output discarded. A fresh
`access-tokens list` diagnostic returned exit 1 and the specific redacted error
`http 403 for https://hub.docker.com/v2/access-tokens: [redacted]`; the 403 rows are
therefore clean provider permission rejections, not silent failures.

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
| `scim-resource-types list` | `pm dockerhub scim-resource-types list --credential dockerhub-scim-only` | HTTP 401 | Explicit SCIM provider rejection using the SCIM-only credential. |
| `scim-resource-types get` | `pm dockerhub scim-resource-types get --name User --credential dockerhub-scim-only` | HTTP 401 | Explicit SCIM provider rejection using the SCIM-only credential. |
| `scim-schemas list` | `pm dockerhub scim-schemas list --credential dockerhub-scim-only` | HTTP 401 | Explicit SCIM provider rejection using the SCIM-only credential. |
| `scim-schemas get` | Canonical `--id urn:ietf:params:scim:schemas:core:2.0:User`, then a safe `--id User` live sentinel, both with `dockerhub-scim-only` | Canonical ID: `error: path variable id contains invalid character ':'`; sentinel: HTTP 401 | The normal SCIM URN is rejected locally before dispatch; the sentinel proved the route reaches Docker Hub and receives an explicit SCIM auth rejection. Classification unchanged; record the input-validation gap for a later TDD slice. |
| `scim-service-provider-config get` | `pm dockerhub scim-service-provider-config get --credential dockerhub-scim-only` | HTTP 401 | Re-run as part of the complete SCIM-only probe; explicit provider rejection. |
| `scim-users list` | `pm dockerhub scim-users list --credential dockerhub-scim-only` | HTTP 401 | Explicit SCIM provider rejection using the SCIM-only credential. |
| `scim-users get` | `pm dockerhub scim-users get --id 00000000-0000-0000-0000-000000000000 --credential dockerhub-scim-only` | HTTP 401 | Explicit SCIM provider rejection using the SCIM-only credential. |
| `auth token create` | `pm dockerhub auth token create` with deliberately invalid non-secret identifier/secret fixture, then plan → preview → approve | HTTP 401 | Plan and preview succeeded; live exchange cleanly rejected the non-secret fixture. No real credential was passed as a raw argument. |
| `auth login create` | `pm dockerhub auth login create` with deliberately invalid non-secret username/password fixture, then plan → preview → approve | HTTP 401 | Plan and preview succeeded; live exchange cleanly rejected the non-secret fixture. No real credential was passed as a raw argument. |
| `auth 2fa-login create` | `pm dockerhub auth 2fa-login create` with deliberately invalid non-secret intermediate-token/code fixture, then plan → preview → approve | HTTP 401 | Plan and preview succeeded; live exchange cleanly rejected the non-secret fixture. No real credential was passed as a raw argument. |
| `repository create` | command shown in the mandatory chain above | HTTP 403 | PAT/account scope limit; no retry performed. |

## Untestable live operations (22)

| Operations | Count | Machine-checkable reason for not dispatching |
| --- | ---: | --- |
| `access-tokens create/delete/update`; `groups create/delete/members add/members remove/replace/update`; `invites bulk-create/cancel/resend`; `org access-tokens create/delete/update`; `org members remove/update`; `org settings update` | 18 | `Named dependency: Docker Hub PAT with account/organization write scope in this worker`; matching reads return explicit HTTP 403, and no disposable target can be safely created/cleaned with the current credential. |
| `repository group assign`, `repository immutable-tags update` | 2 | `Named dependency: writable Docker Hub repository in polymetrics`; repository creation returned HTTP 403 and account listing is empty. |
| `scim-users create/update` | 2 | `Named dependency: valid Docker Hub organization SCIM bearer token`; the SCIM-only reads return explicit HTTP 401. |

The SCIM probe used a credential configured with **only** `scim_bearer_token` (and
non-secret username configuration), with no `docker_pat`. Every one of the seven
SCIM command paths produced an explicit nonzero result rather than a silent success:
six normal requests returned HTTP 401, and `scim-schemas get` returned HTTP 401 when
its URL-safe sentinel was used. The canonical documented URN first hit the local
colon-validation error recorded above. This proves the command outcome is not
silently unauthenticated/successful; without wire-header visibility, HTTP 401 alone
still cannot distinguish an invalid bearer from an omitted bearer header. The SCIM
classification therefore remains unchanged, as directed.

## Docker Registry rate-limit live proof (2026-08-08)

The rate policy was exercised through the rebuilt `pm` binary, not inferred from
source inspection. An isolated temporary project held a non-secret Docker Hub
configuration with `base_url=http://registry-1.docker.io/v2`,
`auth_type=unauthenticated`, and a documentation-only public-IP scope value. A
loopback HTTP proxy returned one valid `tags` record and a same-host next URL per
page. Its requested authority remained `registry-1.docker.io`, so the production
host selector and requester admission path were used while no test page reached
Docker.

`pm dockerhub tags list --limit 101 --json` made 101 logical requests. The proxy
received exactly **100** requests, then remained at **100** after the CLI had been
blocked for five seconds on the 101st admission. The process was interrupted
deliberately rather than waiting out the documented six-hour window. Therefore the
local limiter stopped one request before transport dispatch; **zero** Registry pull
GETs were consumed for this saturation proof.

Immediately afterwards, Docker's free `HEAD
https://registry-1.docker.io/v2/ratelimitpreview/test/manifests/latest` probe
returned:

```text
HTTP/2 200
ratelimit-limit: 100;w=3600
ratelimit-remaining: 100;w=3600
```

This proves Docker still reported 100 pulls of headroom after local admission
stopped the run. The live header's `w=3600` conflicts with the captain-supplied
documentation value of 21,600 seconds; the declaration retains the documented,
conservative six-hour window and records the observed contradiction rather than
silently substituting it. The HEAD itself does not consume a pull.

Rate-proof accounting: **101 exercised / 2 worked checks (local pre-transport
stop and provider HEAD) / 0 failed / 1 not literally testable condition**. The
unmet literal condition is “the same PAT-backed Registry request”: **Named
dependency: Docker Registry v2 bearer-token acquisition and pull-operation support
in `pm`**. Docker Hub's 54-operation management surface has neither an OCI/Registry
pull operation nor a Registry bearer-token acquisition action; its PAT AuthHook is
intentionally Hub-session based. The proof therefore honestly uses the declared
unauthenticated Registry profile and an anonymous bearer solely for Docker's free
quota HEAD, without claiming a PAT was sent to Registry.

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

### Rate-limit final verification

- `go test -timeout 20m ./internal/connectors/connsdk ./internal/connectors/engine
  ./internal/connectors/defs/dockerhub` → PASS.
- `go test -race -timeout 20m ./internal/connectors/engine -run
  'Test(DockerHubRegistryPullPolicyBlocksBeforeTransport|RateLimitSelectorMatchesExactHost)'` → PASS.
- `go test -timeout 20m ./internal/connectors/commandrunner -run
  TestEveryImplementedCommandPassesRuntimePreflight -count=1` → PASS for all 551 bundles.
- `go vet` on changed packages and `go build ./cmd/pm` → PASS.
- `make tidy-check`, `make lint`, `make docs-check`, `make smoke-no-build`,
  `make agent-contract-check`, `make connectorgen-validate`,
  `make connectorgen-surface-sync`, `make connector-boundary`, and
  `make release-workflow-check` → PASS.
- `go test -timeout 20m ./internal/cli/...` → PASS in 566.239s.
- Rechecked `pm dockerhub`, `pm help dockerhub`, `pm dockerhub tags list --help`,
  and `pm docs validate --connectors-dir docs/connectors` → PASS.
- Regenerated website data was compared as parsed objects against `HEAD`: only the
  `dockerhub` object changed in `website/data/connectors.generated.json` and
  `website/lib/connectors.catalog.data.generated.json`; no golden transcript changed.

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
- Docker Hub SCIM schema identifiers are documented as colon-bearing URNs, while the
  current direct-read path guard rejects `:` before dispatch. A URL-safe sentinel
  reached the service and returned the expected explicit 401, so the live auth
  observation is complete; canonical-URN support needs its own later TDD slice.
- Authentication exchanges take sensitive fields as action flags. Deliberately
  invalid non-secret fixtures can verify plan/preview/approval and provider error
  handling, but a successful exchange needs a future secret-reference/from-env
  action-input foundation rather than raw credential flags.
- Full route-help verification must be chunked in this execution environment; one
  54-process sweep exceeds a single command window even though all routes pass.
- Docker's live Registry quota headers reported a one-hour parameter while the
  provided published documentation says six hours. The declaration intentionally
  retains the longer documented window, but this discrepancy makes provider-policy
  evidence less mechanically stable than the JSON declaration model assumes.
- A Docker Hub PAT/session credential is not a Docker Registry bearer credential.
  The 54 management commands have no Registry v2 pull or bearer-token action, so
  the literal same-credential quota proof cannot be expressed without a future
  typed protocol foundation; the pilot used the documented unauthenticated profile
  and stated that limitation rather than faking authenticated coverage.

## GSD inline verify-work and code-review

Resolved the adapter sources for `verify-work` and `code-review` with
`scripts/gsd sources <command>`, then read and followed the generated prompts. The
canonical parent-worker contract forbids spawning the requested GSD verifier/reviewer
roles, so both phases use the documented inline/manual fallback.

- **verify-work verdict:** code and command-route verification pass; the full
  read/auth vector is now live-accounted. Live acceptance remains blocked only on
  the captain-held repository-write/account/organization scope and a valid SCIM
  organization bearer credential, documented above with exact operation accounting.
- **code-review scope:** `internal/connectors/connsdk/http.go`, its local TLS/HTTP2
  regression test, Docker Hub `writes.json`, the Docker Hub write-path regression
  test, and phase evidence.
- **code-review findings:** no Critical, Warning, or Info findings in the committed
  source scope. The strict transport clones rather than mutates caller TLS
  configuration; all provider paths are tested as engine-relative and interpolated;
  no secret appears in source, output, or evidence. The canonical SCIM-URN input
  rejection above is recorded as a follow-up correctness gap, not reclassified or
  changed in this evidence-only slice.
- **additional review evidence:** `go test -race ./internal/connectors/connsdk -run
  TestRequesterDisableRetriesUsesHTTP1WithHTTP2CapableServer -count=1` → PASS;
  `git diff --check` → clean.

### Rate-limit inline verify-work and code-review completion

Sources for `verify-work` and `code-review` were re-resolved with
`scripts/gsd sources <command>` and their generated prompts were followed. The
canonical parent-worker contract still forbids GSD role spawning, so this is the
documented inline/manual fallback.

- **verify-work verdict:** PASS. The production bundle has the two cited
  Registry-only policies; the shared requester stops an over-budget binary run
  before transport, and the free provider HEAD shows remaining headroom.
- **code-review scope:** `connsdk/rate_limits.go`, parameterized header parsing,
  engine selector/validation/resolution, Docker Hub rate declaration/spec/docs,
  generated Docker Hub docs/catalog objects, tests, and phase evidence.
- **code-review findings:** no Critical, Warning, or Info findings. Exact host
  matching prevents the Registry policy from throttling Hub management calls;
  scope material stays non-secret; the response parser retains only typed numeric
  fields; and the test verifies the over-budget request never reaches transport.
