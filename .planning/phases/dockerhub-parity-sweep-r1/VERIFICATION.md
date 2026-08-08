# Docker Hub full documented-operation parity — reconciled verification

## Certification status

**Not certified.** Docker Hub models and implements all **54** documented operations, and every command route is reachable in the rebuilt binary. The live account used for this run does not authorize every Docker Hub API family, so this ledger is an honest current-token accounting, not a claim that every implemented command has been provider-accepted.

**Superseding correction — Captain auth-command override.** The earlier withdrawal recorded below is retained as historical evidence, but it is no longer the delivered state. `auth token create`, `auth login create`, and `auth 2fa-login create` are implemented `direct_write` operations. Their credential fields are declared in `redact_fields` (`secret`, `password`, `login_2fa_token`, `code`); generated command help and the point of use warn that the execution record is retained in plaintext local project state; and `--from-env`/`--value-stdin` avoid argv exposure. The persisted operator-visible plan sample and JSON output redact those fields, while the provider response body remains unmodified on execution. The recorded fake-fixture HTTP 401s followed successful plan → preview → approval lifecycles and prove the routes, so the three auth operations are counted as PROVEN.

Every row below is mutually exclusive and records the outcome from the current write-scoped PAT replay. Superseded pre-upgrade observations are preserved only in git history; they are not mixed into this final ledger.

| Exclusive bucket | Count |
| --- | ---: |
| PROVEN | 14 |
| PROVIDER-PLAN-LIMIT | 0 |
| PROVIDER-PERMISSION | 31 |
| ENTERPRISE-ONLY | 9 |
| **Total implemented operations** | **54** |
| **Total documented operations modelled** | **54** |

No open OUR-DEFECT remains in this slice. The Docker Hub canonical SCIM schema URN and the cross-connector opaque-email path-segment defect were both fixed red-first; their tests and audit are in TDD-LEDGER.md.

## Method and safety boundary

- Rebuilt the branch binary and loaded the captain-upgraded PAT only from an environment variable through --from-env; no credential, approval token, or token-derived value was printed or kept in this record.
- Used a fresh isolated project root. Every dispatched non-destructive mutation used plan → preview → approval → execute; delete/cancel/remove operations also used the typed destructive confirmation.
- Current nonzero outcomes record the exact HTTP status only. This verification record omits Docker
  Hub response bodies and request authentication material; the runtime itself returns declared
  provider response bodies unchanged.
- The current binary completed 54/54 Docker Hub --help routes with zero failures, plus pm help dockerhub and bare pm dockerhub successfully. All 54 routes are executable at the runtime preflight boundary.

The retained private fixture is polymetrics/pm-e2e-20260808-173040. It was created and proven private during the earlier captain-authorized E2E and is intentionally retained. Its historical creation is **not** used as the current-token result for repository create; that command's one current outcome is the HTTP 403 shown below.

## PROVEN (14)

PROVEN includes a provider's deliberate safe-fixture rejection when the command reached Docker Hub through the complete lifecycle and the response clearly proves the route.

| Operation | Current-token evidence |
| --- | --- |
| repositories list | Completed successfully against the retained namespace. |
| repository detail list | Completed successfully; a second current replay with namespace=library, repository=alpine also succeeded, proving namespace is honored rather than discarded. |
| tags list | Completed successfully; current library/alpine override replay also succeeded. |
| tag detail list | Completed successfully; current library/alpine:latest override replay also succeeded. |
| repository check | Status-only HEAD command completed successfully. |
| repository tags check | Status-only HEAD command completed successfully. |
| repository tag check | Status-only HEAD command completed successfully. |
| repository immutable-tags verify | Completed successfully against the retained repository. |
| repository immutable-tags update | Plan, preview, approval, and live execution completed successfully. |
| repository group assign | Plan, preview, approval, and execution reached Docker Hub; deliberate nonexistent group_id=0 received explicit HTTP 400 validation. |
| invites resend | Plan, preview, approval, and execution reached Docker Hub; deliberate nonexistent id=0 received explicit HTTP 405. |
| auth token create | Plan, preview, approval, and execution reached Docker Hub; deliberately fake non-secret exchange material received explicit HTTP 401. |
| auth login create | Plan, preview, approval, and execution reached Docker Hub; deliberately fake non-secret exchange material received explicit HTTP 401. |
| auth 2fa-login create | Plan, preview, approval, and execution reached Docker Hub; deliberately fake non-secret exchange material received explicit HTTP 401. |

## PROVIDER-PERMISSION (31)

Each HTTP 403 was surfaced by pm as a nonzero, redacted provider error; none was a silent success. Docker's response did not identify a paid-plan/entitlement condition, so none is inferred to be a plan-limit result. These rows stay implemented; their named dependency is provider access for the account actually tested.

| Operation | Current outcome | Named dependency |
| --- | --- | --- |
| access-tokens create | Full write lifecycle → HTTP 403 | Named dependency: Docker Hub provider permission for the current polymetrics principal to manage personal access tokens. |
| access-tokens delete | Full destructive lifecycle → HTTP 403 | Named dependency: Docker Hub provider permission for the current polymetrics principal to manage personal access tokens. |
| access-tokens get | Direct read → HTTP 403 | Named dependency: Docker Hub provider permission for the current polymetrics principal to manage personal access tokens. |
| access-tokens list | Direct read → HTTP 403 | Named dependency: Docker Hub provider permission for the current polymetrics principal to manage personal access tokens. |
| access-tokens update | Full write lifecycle → HTTP 403 | Named dependency: Docker Hub provider permission for the current polymetrics principal to manage personal access tokens. |
| audit-logs actions list | Direct read → HTTP 403 | Named dependency: Docker Hub audit-log API permission for the current polymetrics principal. |
| audit-logs list | Direct read → HTTP 403 | Named dependency: Docker Hub audit-log API permission for the current polymetrics principal. |
| groups create | Full write lifecycle → HTTP 403 | Named dependency: Docker Hub organization API permission on polymetrics to manage groups/teams. |
| groups delete | Full destructive lifecycle → HTTP 403 | Named dependency: Docker Hub organization API permission on polymetrics to manage groups/teams. |
| groups get | Direct read → HTTP 403 | Named dependency: Docker Hub organization API permission on polymetrics to manage groups/teams. |
| groups list | Direct read → HTTP 403 | Named dependency: Docker Hub organization API permission on polymetrics to manage groups/teams. |
| groups members add | Full write lifecycle → HTTP 403 | Named dependency: Docker Hub organization API permission on polymetrics to manage groups/teams. |
| groups members list | Direct read → HTTP 403 | Named dependency: Docker Hub organization API permission on polymetrics to manage groups/teams. |
| groups members remove | Full destructive lifecycle → HTTP 403 | Named dependency: Docker Hub organization API permission on polymetrics to manage groups/teams. |
| groups replace | Full write lifecycle → HTTP 403 | Named dependency: Docker Hub organization API permission on polymetrics to manage groups/teams. |
| groups update | Full write lifecycle → HTTP 403 | Named dependency: Docker Hub organization API permission on polymetrics to manage groups/teams. |
| invites bulk-create | Full write lifecycle with provider dry-run fixture → HTTP 403 | Named dependency: Docker Hub organization API permission on polymetrics to manage invitations. |
| invites cancel | Full destructive lifecycle → HTTP 403 | Named dependency: Docker Hub organization API permission on polymetrics to manage invitations. |
| invites list | Direct read → HTTP 403 | Named dependency: Docker Hub organization API permission on polymetrics to manage invitations. |
| org access-tokens create | Full write lifecycle → HTTP 403 | Named dependency: Docker Hub organization API permission on polymetrics to manage organization access tokens. |
| org access-tokens delete | Full destructive lifecycle → HTTP 403 | Named dependency: Docker Hub organization API permission on polymetrics to manage organization access tokens. |
| org access-tokens get | Direct read → HTTP 403 | Named dependency: Docker Hub organization API permission on polymetrics to manage organization access tokens. |
| org access-tokens list | Direct read → HTTP 403 | Named dependency: Docker Hub organization API permission on polymetrics to manage organization access tokens. |
| org access-tokens update | Full write lifecycle → HTTP 403 | Named dependency: Docker Hub organization API permission on polymetrics to manage organization access tokens. |
| org members export | Binary download → HTTP 403 | Named dependency: Docker Hub organization API permission on polymetrics to read/export members. |
| org members list | Direct read → HTTP 403 | Named dependency: Docker Hub organization API permission on polymetrics to read/export members. |
| org members remove | Full destructive lifecycle → HTTP 403 | Named dependency: Docker Hub organization API permission on polymetrics to manage members. |
| org members update | Full write lifecycle → HTTP 403 | Named dependency: Docker Hub organization API permission on polymetrics to manage members. |
| org settings get | Direct read → HTTP 403 | Named dependency: Docker Hub organization API permission on polymetrics to read settings. |
| org settings update | Plan and preview succeeded; execution was not sent because the current setting values cannot be read after the HTTP 403 above, so a safe no-op cannot be formed. | Named dependency: Docker Hub organization API permission on polymetrics to read and update settings. |
| repository create | Full write lifecycle → HTTP 403 | Named dependency: Docker Hub provider permission for the current polymetrics principal to create repositories in namespace polymetrics. |

## ENTERPRISE-ONLY (9)

The provider artifact declares the SCIM family under its separate bearerSCIMAuth scheme. A credential configured with **only** scim_bearer_token (the current PAT placed solely in that field for this probe) made all read requests explicit HTTP 401s: Docker Hub did not silently fall back to unauthenticated requests. The canonical schema URN now reaches the live service and gets that HTTP 401, proving the local colon defect is repaired.

| Operation | Current outcome | Named dependency |
| --- | --- | --- |
| scim-resource-types get | Live direct read → HTTP 401 | Named dependency: Docker Hub Enterprise organization SCIM entitlement and a valid scim_bearer_token authorized for that organization. |
| scim-resource-types list | Live direct read → HTTP 401 | Named dependency: Docker Hub Enterprise organization SCIM entitlement and a valid scim_bearer_token authorized for that organization. |
| scim-schemas get | Canonical urn:ietf:params:scim:schemas:core:2.0:User reached live Docker Hub → HTTP 401 | Named dependency: Docker Hub Enterprise organization SCIM entitlement and a valid scim_bearer_token authorized for that organization. |
| scim-schemas list | Live direct read → HTTP 401 | Named dependency: Docker Hub Enterprise organization SCIM entitlement and a valid scim_bearer_token authorized for that organization. |
| scim-service-provider-config get | Live direct read → HTTP 401 | Named dependency: Docker Hub Enterprise organization SCIM entitlement and a valid scim_bearer_token authorized for that organization. |
| scim-users create | Never dispatched: no valid Enterprise SCIM bearer and a create could mutate an account. | Named dependency: Docker Hub Enterprise organization SCIM entitlement and a valid scim_bearer_token authorized for that organization. |
| scim-users get | Live direct read → HTTP 401 | Named dependency: Docker Hub Enterprise organization SCIM entitlement and a valid scim_bearer_token authorized for that organization. |
| scim-users list | Live direct read → HTTP 401 | Named dependency: Docker Hub Enterprise organization SCIM entitlement and a valid scim_bearer_token authorized for that organization. |
| scim-users update | Never dispatched: no valid Enterprise SCIM bearer and an update could mutate an account. | Named dependency: Docker Hub Enterprise organization SCIM entitlement and a valid scim_bearer_token authorized for that organization. |

## Operations never dispatched to the live service

Exactly three implemented operations were not dispatched in this reconciliation:

1. org settings update — its read is current HTTP 403, so the worker cannot derive a safe no-op setting payload; dispatching guessed values could weaken organization restrictions. Named dependency: Docker Hub organization settings read/write permission on polymetrics.
2. scim-users create — no Enterprise SCIM bearer was available and creating a user is not a safe negative fixture. Named dependency: Docker Hub Enterprise organization SCIM entitlement and a valid scim_bearer_token.
3. scim-users update — no Enterprise SCIM bearer was available and modifying a user is not a safe negative fixture. Named dependency: Docker Hub Enterprise organization SCIM entitlement and a valid scim_bearer_token.

All other 51 implemented operations were dispatched to Docker Hub during the current-token reconciliation.

## Rate-limit and registry facts retained from the pilot

The declarative rate policy applies only to registry-1.docker.io, not the mostly hub.docker.com API surface. With a configured budget of 100, the rebuilt binary issued 101 logical requests through a local proxy: the proxy observed exactly 100 and the 101st was stopped before transport. Docker's free ratelimitpreview HEAD probe then reported 100 pulls remaining, so local admission stopped before Docker would have. Docker documentation says a 21,600-second window; the observed header used w=3600. The policy deliberately retains the documented 21,600-second window, and the contradiction remains an open provider question.

pm has no OCI/Registry v2 image pull/push operation or Registry bearer-acquisition flow. The retained test image was transferred with Docker's Registry protocol and then read through Hub API tags/repository operations. This is a product-surface gap for the captain, not a Docker Hub connector defect.

## Pilot friction points

- Docker Hub returns useful HTTP status codes but does not identify whether a 403 is a role, resource, or subscription-tier denial. The ledger therefore classifies only documented SCIM as Enterprise-only and treats the remaining concrete 403s as provider permissions rather than inventing plan claims.
- Connector-command JSON deliberately omits approval tokens. The safe live harness must capture the bounded token from human-readable plan/preview output in memory; it must not log it.
- The Registry pull service and Hub API are distinct hosts with different limiter semantics, and the documented/observed window mismatch needs an explicit provider follow-up.
- Shared path validation was originally modeled as local identifier validation. The audit exposed two documented opaque-segment classes, SCIM URNs and email addresses, so path values now have their own narrowly tested validator while command and credential names remain strict.

## Rebased local validation (2026-08-09)

The interrupted no-mistakes run was cancelled and deliberately not reused. Its
pipeline-owned repair was preserved, restored, and the branch was rebased with
`origin/main` at `d453fbe256eb22d90ea77dbed634b245bd6e795b` as the verified merge
base. Focused app, engine, command-runner, bundle-registry, Docker Hub definition,
hook, and safety suites passed; the full `internal/cli` regression suite was run
separately with its required 20-minute timeout. `go vet` on changed packages,
`go build ./cmd/pm`, Docker Hub `connectorgen validate`, `surface-sync --check`,
the Docker Hub ledger test, and the fleet implemented-command preflight passed.

The rebuilt binary was then exercised in nine bounded batches: all **54/54**
implemented `pm dockerhub <command> --help` routes succeeded, along with `pm help
dockerhub` and bare `pm dockerhub`. Docs validation, tidy, lint, docs, smoke,
agent-contract, generator, website script tests, and website typecheck also passed
locally. A fresh no-mistakes pipeline is required for this newly committed head;
the cancelled pre-rebase result remains non-authoritative.

## GSD/TDD evidence

The required inline lifecycle evidence is recorded in PLAN.md, TDD-LEDGER.md, RUN-STATE.json, and REVIEW.md. The canonical parent-worker contract prohibits role spawning in this worktree, so the lifecycle uses the documented inline/manual fallback. Final focused/package gates, 54/54 rebuilt-binary help reachability, and standard inline review are green. The automated-review corrective slice (SCIM-only auth selection, SCIM proxy routing, status-only HEAD 404 semantics, credential add-time admission, the captain-authorized restoration of the three authentication exchanges, and the CLAUDE.md/AGENTS.md scope reverts) is recorded in TDD-LEDGER.md under "Corrective slice — automated-review dispositions"; the next gate is fresh no-mistakes validation of this rebased head.
