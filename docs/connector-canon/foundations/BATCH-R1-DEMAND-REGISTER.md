# Batch R1 Foundation Atlas demand register

This is the issue-scoped reconciliation record for [#4400].  It records what
the frozen Batch R1 source evidence says about deferred cells and what the
Foundation Atlas already owns.  It is not a runtime registry, source lock,
mapping admission rule, certification rule, connector definition, or execution
claim.

## Frozen evidence boundary

At `fm/cli-top100-declaration-batch-r1@0a708dea5e0024a173b19959d2c43f2bf5a6e0f2`,
the check-only command below reports:

```sh
go run ./cmd/connectorgen deferred-visibility \
  data/connector-canon/batch1-source-operation-mapping-cohort.json --check --json
```

| Fact | Value |
| --- | ---: |
| Primary retained source operations | 4,341 |
| Source rows, including declared supplements | 4,343 |
| Seven-lane matrix cells | 30,401 |
| Deferred cells | 6,790 |
| `mapped_unproven` cells | 6,778 |
| `missing_foundation` cells | 12 |
| Executable declarations reported | 0 |

The 6,778 `mapped_unproven` cells all resolve to the existing
[`source.projection-admission.v1`](catalog.json) authoring prerequisite.  That
is retained source mapping, not evidence that a runtime foundation is missing.
Neither importer/retention status nor certification status may erase such a
cell or turn it into a runtime-foundation demand.

## Atlas reconciliation rules

1. [`transport.sync-contract.v1`](catalog.json) already owns closed
   source-to-warehouse-to-destination composition, selected executor and
   conformance references, and acknowledgement/checkpoint order.  It does not
   receive an inbound provider webhook or infer a provider delivery contract
   from a webhook-registration mutation.
2. [`runtime.provider-extension-seams.v1`](catalog.json) already owns the
   closed connector-specific extension mechanism.  A future adapter must select
   that seam and the closed transport contract; it must not add arbitrary
   webhook ingestion or a connector-name branch to shared runtime code.
3. A provider operation that creates or updates a webhook configures provider
   delivery to a URL.  It is not, by itself, an ingress endpoint, receiver,
   worker, executor, command, or executable sync transport.
4. `missing_foundation` below therefore means a **connector-specific inbound
   adapter is absent**, while the generic transport/extension foundations are
   reused.  It does not mean that Batch R1 needs a generic webhook runtime.

## Exact `missing_foundation` register

Every row is `sync_transport`, `applicable`, and `missing_foundation` in the
connector-owned source-lane matrix.  All runtime claims remain `none` in the
deferred-visibility report.

| Connector | Exact retained source ID | Method and path | Citation | Existing gap ID | Atlas disposition |
| --- | --- | --- | --- | --- | --- |
| Bitbucket | `bitbucket.rest.post_/repositories/{workspace}/{repo_slug}/hooks` | `POST /repositories/{workspace}/{repo_slug}/hooks` | [`bitbucket-operation-source-lock.json`](../../../internal/connectors/defs/bitbucket/sources/bitbucket-operation-source-lock.json), `paths["/repositories/{workspace}/{repo_slug}/hooks"].post` | `cli-webhook-event-surface-foundation-r1` | New planned `bitbucket.inbound-webhook-receiver.v1`; reuse generic sync + extension seam. |
| Bitbucket | `bitbucket.rest.post_/workspaces/{workspace}/hooks` | `POST /workspaces/{workspace}/hooks` | same source, `paths["/workspaces/{workspace}/hooks"].post` | `cli-webhook-event-surface-foundation-r1` | Same planned connector-specific adapter. |
| Bitbucket | `bitbucket.rest.put_/repositories/{workspace}/{repo_slug}/hooks/{uid}` | `PUT /repositories/{workspace}/{repo_slug}/hooks/{uid}` | same source, `paths["/repositories/{workspace}/{repo_slug}/hooks/{uid}"].put` | `cli-webhook-event-surface-foundation-r1` | Same planned connector-specific adapter. |
| Bitbucket | `bitbucket.rest.put_/workspaces/{workspace}/hooks/{uid}` | `PUT /workspaces/{workspace}/hooks/{uid}` | same source, `paths["/workspaces/{workspace}/hooks/{uid}"].put` | `cli-webhook-event-surface-foundation-r1` | Same planned connector-specific adapter. |
| CircleCI | `circleci.rest.createWebhook` | `POST /webhook` | [`circleci-operation-source-lock.json`](../../../internal/connectors/defs/circleci/sources/circleci-operation-source-lock.json), `paths["/webhook"].post` | `circleci-inbound-webhook-receiver-r1` | New planned `circleci.inbound-webhook-receiver.v1`; reuse generic sync + extension seam. |
| CircleCI | `circleci.rest.updateWebhook` | `PUT /webhook/{webhook_id}` | same source, `paths["/webhook/{webhook_id}"].put` | `circleci-inbound-webhook-receiver-r1` | Same planned connector-specific adapter. |
| GitLab | `gitlab.rest.postApiV4GroupsIdHooks` | `POST /api/v4/groups/{id}/hooks` | [`gitlab-operation-source-lock.json`](../../../internal/connectors/defs/gitlab/sources/gitlab-operation-source-lock.json), `paths["/api/v4/groups/{id}/hooks"].post` | `gitlab-inbound-webhook-source-executor-r1` | Existing planned `gitlab.inbound-webhook-receiver.v1`; no change in this issue. |
| GitLab | `gitlab.rest.postApiV4Hooks` | `POST /api/v4/hooks` | same source, `paths["/api/v4/hooks"].post` | `gitlab-inbound-webhook-source-executor-r1` | Existing planned entry; no change in this issue. |
| GitLab | `gitlab.rest.postApiV4ProjectsIdHooks` | `POST /api/v4/projects/{id}/hooks` | same source, `paths["/api/v4/projects/{id}/hooks"].post` | `gitlab-inbound-webhook-source-executor-r1` | Existing planned entry; no change in this issue. |
| Jira | `jira.rest.registerDynamicWebhooks` | `POST /rest/api/3/webhook` | [`jira-operation-source-lock.json`](../../../internal/connectors/defs/jira/sources/jira-operation-source-lock.json), `paths["/rest/api/3/webhook"].post` | `cli-webhook-event-surface-foundation-r1` | New planned `jira.inbound-webhook-receiver.v1`; reuse generic sync + extension seam. |
| Stripe | `stripe.rest.PostWebhookEndpoints` | `POST /v1/webhook_endpoints` | [`stripe-operation-source-lock.json`](../../../internal/connectors/defs/stripe/sources/stripe-operation-source-lock.json), `paths["/v1/webhook_endpoints"].post` | `stripe-inbound-webhook-receiver-r1` | Existing planned `stripe.inbound-webhook-receiver.v1`; no change in this issue. |
| Vercel | `vercel.rest.createWebhook` | `POST /v1/webhooks` | [`vercel-operation-source-lock.json`](../../../internal/connectors/defs/vercel/sources/vercel-operation-source-lock.json), `paths["/v1/webhooks"].post` | `cli-webhook-event-surface-foundation-r1` | New planned `vercel.inbound-webhook-receiver.v1`; reuse generic sync + extension seam. |

The source URLs and current lock digests remain in the linked locks.  This
register intentionally does not copy provider schemas, docs, credentials,
delivery secrets, or provider policy into the Atlas.

## Planned adapter boundary

The four new catalog records are `planned` and
`connector_specific_reference`, not generic foundations:

| Planned Atlas ID | Affected source rows | Existing reusable foundation | Future connector-local gap |
| --- | --- | --- | --- |
| `bitbucket.inbound-webhook-receiver.v1` | the four Bitbucket rows above | `transport.sync-contract.v1`, `runtime.provider-extension-seams.v1` | Bitbucket inbound receiver/source executor selected through a closed definition reference. |
| `circleci.inbound-webhook-receiver.v1` | `createWebhook`, `updateWebhook` | same | CircleCI inbound receiver/source executor selected through a closed definition reference. |
| `jira.inbound-webhook-receiver.v1` | `registerDynamicWebhooks` | same | Jira inbound receiver/source executor selected through a closed definition reference. |
| `vercel.inbound-webhook-receiver.v1` | `createWebhook` | same | Vercel inbound receiver/source executor selected through a closed definition reference. |

None of those records creates `sync_transport.json`, an HTTP listener, a
receiver, a worker, an executor, a CLI command, credential handling, provider
I/O, or a runtime registration.

## Required future proof and approval gate

Implementation of any planned adapter requires **separate captain approval**.
Before that approval, it stays non-executing.  A proposed implementation must
first retain the provider-specific delivery facts and then prove, where those
facts exist:

1. invalid inbound authentication is rejected before parsing;
2. event scope is bound to the exact registered provider source;
3. replay/duplicate handling uses an explicit source-cited identity, never an
   inferred one;
4. accepted work is durably staged before acknowledgement;
5. a connector-owned selected executor/worker and conformance pair are
   registered through the existing closed seams; and
6. checkpoint advancement occurs only after durable downstream acknowledgement.

If the retained provider source does not document authentication, replay, event
scope, retry, or delivery semantics, that remains a precise source-evidence
gap.  It is not permission to invent a generic policy or mark the transport
executable.

## Baseline separation

- Existing GitLab and Stripe planned records already represent their exact
  connector-local missing adapters.  This issue leaves their source mappings
  and their non-execution boundary unchanged.
- GitLab's documented regex and typed-alias restrictions remain separately
  classified by their existing Atlas records and connector-local source
  evidence.  They are not an inbound-webhook or new shared-runtime demand.
- Source-retention/import and certification/conformance systems remain
  supporting evidence/overlays.  They cannot be used to hide a source row,
  declare a receiver, or replace the runtime proof described above.

[#4400]: https://github.com/polymetrics-ai/cli/issues/4400
