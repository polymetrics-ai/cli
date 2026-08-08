# Docker Hub full documented-operation parity — reconciled verification

## Certification status

**Not certified.** Docker Hub models all 54 documented operations; **51 are implemented** and every one of those command routes is reachable in the rebuilt binary, but the live account used for this run does not authorize every Docker Hub API family. This ledger is deliberately an honest current-token accounting, not a claim that every implemented command has been provider-accepted.

**Corrected after automated review (Captain Decision [key=auth-secrets]).** The three credential-for-token authentication exchanges — `auth token create`, `auth login create`, `auth 2fa-login create` — are **no longer shipped as executable commands**. Each one's only input is a live credential passed as a plain CLI flag mapped to a record field, and the reverse-ETL path it ran on persists that record to the project state file as plaintext and echoes it in plan output; an action's `redact_fields` names a RESPONSE field and never redacts the request credential. They remain in the 54-row surface as `availability: "planned"` with the machine-checkable dependency `named_dependency=requires secure secret input (stdin/env/vault-reference), encrypted or ephemeral plan storage, and a secure sink for the returned token`. Their previously recorded HTTP 401 lifecycle observations are preserved below under NOT-IMPLEMENTED as history, and are **not** counted as PROVEN. The bucket split therefore moves from 14/0/31/9 over 54 to **11 PROVEN / 0 PROVIDER-PLAN-LIMIT / 31 PROVIDER-PERMISSION / 9 ENTERPRISE-ONLY over 51 implemented, plus 3 NOT-IMPLEMENTED**; any PR body must state the corrected split.

Every row below is mutually exclusive and records the outcome from the current write-scoped PAT replay. Superseded pre-upgrade observations are preserved only in git history; they are not mixed into this final ledger.

| Exclusive bucket | Count |
| --- | ---: |
| PROVEN | 11 |
| PROVIDER-PLAN-LIMIT | 0 |
| PROVIDER-PERMISSION | 31 |
| ENTERPRISE-ONLY | 9 |
| **Total implemented operations** | **51** |
| NOT-IMPLEMENTED (blocked on a named dependency) | 3 |
| **Total documented operations modelled** | **54** |

No open OUR-DEFECT remains in this slice. The Docker Hub canonical SCIM schema URN and the cross-connector opaque-email path-segment defect were both fixed red-first; their tests and audit are in TDD-LEDGER.md.

## Method and safety boundary

- Rebuilt the branch binary and loaded the captain-upgraded PAT only from an environment variable through --from-env; no credential, approval token, or token-derived value was printed or kept in this record.
- Used a fresh isolated project root. Every dispatched non-destructive mutation used plan → preview → approval → execute; delete/cancel/remove operations also used the typed destructive confirmation.
- Current nonzero outcomes record the exact HTTP status only. Docker Hub response bodies and request authentication material were intentionally redacted.
- The current binary completed 54/54 Docker Hub --help routes with zero failures, plus pm help dockerhub and bare pm dockerhub successfully. Three of those routes now render as declared-but-not-implemented rather than executable.

The retained private fixture is polymetrics/pm-e2e-20260808-173040. It was created and proven private during the earlier captain-authorized E2E and is intentionally retained. Its historical creation is **not** used as the current-token result for repository create; that command's one current outcome is the HTTP 403 shown below.

## PROVEN (11)

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

## NOT-IMPLEMENTED (3)

Withdrawn from the executable surface by Captain Decision [key=auth-secrets]. The live observations are retained as history of what was attempted; they are not a capability claim, and these operations cannot return a token through generic reverse ETL.

| Operation | Retained historical observation | Named dependency |
| --- | --- | --- |
| auth token create | Earlier run: plan, preview, and approval succeeded; deliberately fake non-secret exchange material received explicit HTTP 401. Not dispatchable in the delivered head. | requires secure secret input (stdin/env/vault-reference), encrypted or ephemeral plan storage, and a secure sink for the returned token. |
| auth login create | Earlier run: plan, preview, and approval succeeded; deliberately fake non-secret exchange material received explicit HTTP 401. Not dispatchable in the delivered head. | requires secure secret input (stdin/env/vault-reference), encrypted or ephemeral plan storage, and a secure sink for the returned token. |
| auth 2fa-login create | Earlier run: plan, preview, and approval succeeded; deliberately fake non-secret exchange material received explicit HTTP 401. Not dispatchable in the delivered head. | requires secure secret input (stdin/env/vault-reference), encrypted or ephemeral plan storage, and a secure sink for the returned token. |

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

Exactly three implemented operations were not dispatched in this reconciliation (the three NOT-IMPLEMENTED authentication exchanges above are a separate, deliberate withdrawal, not a dispatch gap):

1. org settings update — its read is current HTTP 403, so the worker cannot derive a safe no-op setting payload; dispatching guessed values could weaken organization restrictions. Named dependency: Docker Hub organization settings read/write permission on polymetrics.
2. scim-users create — no Enterprise SCIM bearer was available and creating a user is not a safe negative fixture. Named dependency: Docker Hub Enterprise organization SCIM entitlement and a valid scim_bearer_token.
3. scim-users update — no Enterprise SCIM bearer was available and modifying a user is not a safe negative fixture. Named dependency: Docker Hub Enterprise organization SCIM entitlement and a valid scim_bearer_token.

All other 48 implemented operations were dispatched to Docker Hub during the current-token reconciliation.

## Rate-limit and registry facts retained from the pilot

The declarative rate policy applies only to registry-1.docker.io, not the mostly hub.docker.com API surface. With a configured budget of 100, the rebuilt binary issued 101 logical requests through a local proxy: the proxy observed exactly 100 and the 101st was stopped before transport. Docker's free ratelimitpreview HEAD probe then reported 100 pulls remaining, so local admission stopped before Docker would have. Docker documentation says a 21,600-second window; the observed header used w=3600. The policy deliberately retains the documented 21,600-second window, and the contradiction remains an open provider question.

pm has no OCI/Registry v2 image pull/push operation or Registry bearer-acquisition flow. The retained test image was transferred with Docker's Registry protocol and then read through Hub API tags/repository operations. This is a product-surface gap for the captain, not a Docker Hub connector defect.

## Pilot friction points

- Docker Hub returns useful HTTP status codes but does not identify whether a 403 is a role, resource, or subscription-tier denial. The ledger therefore classifies only documented SCIM as Enterprise-only and treats the remaining concrete 403s as provider permissions rather than inventing plan claims.
- Connector-command JSON deliberately omits approval tokens. The safe live harness must capture the bounded token from human-readable plan/preview output in memory; it must not log it.
- The Registry pull service and Hub API are distinct hosts with different limiter semantics, and the documented/observed window mismatch needs an explicit provider follow-up.
- Shared path validation was originally modeled as local identifier validation. The audit exposed two documented opaque-segment classes, SCIM URNs and email addresses, so path values now have their own narrowly tested validator while command and credential names remain strict.

## GSD/TDD evidence

The required inline lifecycle evidence is recorded in PLAN.md, TDD-LEDGER.md, RUN-STATE.json, and REVIEW.md. The canonical parent-worker contract prohibits role spawning in this worktree, so the lifecycle uses the documented inline/manual fallback. Final focused/package gates, 54/54 rebuilt-binary help reachability, and standard inline review are green. The automated-review corrective slice (SCIM-only auth selection, SCIM proxy routing, status-only HEAD 404 semantics, credential add-time admission, withdrawal of the three authentication exchanges, and the CLAUDE.md/AGENTS.md scope reverts) is recorded in TDD-LEDGER.md under "Corrective slice — automated-review dispositions"; the next gate is the requested no-mistakes delivery pipeline.
