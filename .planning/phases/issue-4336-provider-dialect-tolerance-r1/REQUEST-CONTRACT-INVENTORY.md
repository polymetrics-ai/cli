# Batch-1 request-contract inventory — issue 4336

## Scope and method

This is the follow-on census requested after the seven-case provider-dialect
foundation was committed. It deliberately does **not** widen that implementation.
It identifies the next foundation's request-contract work so a future lane can
make an explicit decision once, rather than discovering one provider at a time.

On 2026-08-23 the local-only analysis harness downloaded the current official
source document for every Batch-1 provider, parsed it through the production
source-import parser and resolver, then checked **every operation parameter and
every request-body media schema independently** with the production request
contract validators. It de-duplicated exact `(provider, operation, subject,
error)` records and grouped them by the exact refusing construct. This yields
10,051 currently observable request-unit refusals.

The normal importer returns the first refusal in sorted operation order. The
independent census is necessary so a failure in an early operation cannot hide
the same contract category elsewhere. A schema which first fails at a parent
may have further invalid children; this inventory reports every current
request-unit blocker the existing validator can observe, not speculative
post-fix errors behind that blocker.

The harness and downloaded artifacts were removed after this report was made;
they are neither product code nor a pinned source lock. The command and its
passing result are recorded in `RUN-STATE.md`.

### Documents measured

| Provider | Official document | SHA-256 |
| --- | --- | --- |
| Asana | `https://raw.githubusercontent.com/Asana/openapi/master/defs/asana_oas.yaml` | `2f7ffc2ad8efc7ccead50377aaebf3ba8d53f74d0c54fe509e411b5ca5290e51` |
| Bitbucket | `https://developer.atlassian.com/cloud/bitbucket/swagger.v3.json` | `3dbfe6a80143511a287e58c21a193d3551ab5d41e8b60e65c1ae121b7000dec3` |
| CircleCI | `https://circleci.com/api/v2/openapi.json` | `61c6ce11e8de509948aa3d53dcd0169913f52de20920b130b6a85dea41d66d07` |
| Docker Hub | `https://docs.docker.com/reference/api/hub/latest.yaml` | `99d9d53c2d93656a3c66d604885abd153dc5df285abc0ecb13802a3bc53d0756` |
| GitLab | `https://gitlab.com/gitlab-org/gitlab/-/raw/master/doc/api/openapi/openapi_v3.yaml` | `6b6ad591ff1b54ab429d0502812a2b2955501f1f6bebdae1888ba0bea086cf82` |
| Jira | `https://developer.atlassian.com/cloud/jira/platform/swagger-v3.v3.json` | `511d0b97390cc47aa0e1367189210a41f32088d9c869e7bb01f43698bdf7e5e8` |
| Notion | `https://developers.notion.com/openapi.json` | `dee5763763b0b9fbad2aa8d5adb173ca350ec26dda557e658c5dbe9d2ea2f258` |
| Sentry | `https://raw.githubusercontent.com/getsentry/sentry-api-schema/main/openapi-derefed.json` | `b71216654e44cc18f5e262fbb5075df67f1504a123d4bcb51cc8e8cc74ebd435` |
| Stripe | `https://raw.githubusercontent.com/stripe/openapi/master/openapi/spec3.json` | `3653ad45bbec54fcbe461c541c908355b715018bdf455a0e11b27bedb2cbdee5` |
| Vercel | `https://openapi.vercel.sh` | `74cb7ff3dc0b89cc344b13ac9c6d5f1d9b7d7a9356cfd6b5a779da51fd43da28` |

### Classifications and refusing clauses

`legitimate` means a valid OpenAPI/JSON Schema declaration which the importer
cannot safely turn into a finite generated input contract today. It is not a
reason to weaken validation. `malformed` means the provider violates the
OpenAPI path-parameter contract. `finite-bound` means a valid document exceeds
an intentional finite retention guard and requires a separately justified
sizing/retention decision.

| Construct | Refusing production clause | Classification |
| --- | --- | --- |
| `allOf`, `anyOf`, or `oneOf` request composition | `cmd/connectorgen/sourceimport.go:6620` | legitimate |
| no schema type | `cmd/connectorgen/sourceimport.go:6650` | legitimate |
| string without `maxLength` | `cmd/connectorgen/sourceimport.go:6912` | legitimate |
| number/integer without lower or upper bound | `cmd/connectorgen/sourceimport.go:6957`, `:6960` | legitimate |
| array without `maxItems` | `cmd/connectorgen/sourceimport.go:6936` | legitimate |
| object with dynamic `additionalProperties` | `cmd/connectorgen/sourceimport.go:6720` | legitimate |
| object without fixed `properties` | `cmd/connectorgen/sourceimport.go:6725` | legitimate |
| template placeholder lacks a required path parameter | `cmd/connectorgen/sourceimport.go:6134` | malformed |
| request media exceeds retained-descriptor byte quota | `cmd/connectorgen/sourceimport.go:1561` | finite-bound |

## Complete observed categories

Examples are the first three stable operation/subject records in each category;
the `count` is the full number of distinct records in the measured provider
document, not a declaration count.

| Provider | Construct / count | Examples | Refusing line | Classification |
| --- | --- | --- | --- | --- |
| Asana | array has no `maxItems` — 189 | `GET /access_requests` query `opt_fields`; `GET /agents/{agent_gid}` query `opt_fields`; `GET /allocations/{allocation_gid}` query `opt_fields` | `:6936` | legitimate |
| Asana | number has no minimum — 1 | `GET /workspaces/{workspace_gid}/typeahead` query `count` | `:6957` | legitimate |
| Asana | object dynamic `additionalProperties` — 105 | `POST /access_requests` JSON; `POST /allocations` JSON; `POST /attachments` multipart | `:6720` | legitimate |
| Asana | string has no `maxLength` — 438 | `DELETE /allocations/{allocation_gid}` path `allocation_gid`; `DELETE /attachments/{attachment_gid}` path `attachment_gid`; `DELETE /budgets/{budget_gid}` path `budget_gid` | `:6912` | legitimate |
| Bitbucket | `allOf` — 41 | `POST /repositories/{workspace}/{repo_slug}/branch-restrictions` JSON; `POST /repositories/{workspace}/{repo_slug}/commit/{commit}/comments` JSON; `POST /repositories/{workspace}/{repo_slug}/commit/{commit}/reports/{reportId}/annotations` JSON | `:6620` | legitimate |
| Bitbucket | no type — 5 | `POST /repositories/{workspace}/{repo_slug}/forks` JSON; `POST /repositories/{workspace}/{repo_slug}/pullrequests/{pull_request_id}/tasks` JSON; `POST /repositories/{workspace}/{repo_slug}` JSON | `:6650` | legitimate |
| Bitbucket | number has no maximum — 1 | `GET /repositories/{workspace}/{repo_slug}/pipelines` query `page` | `:6960` | legitimate |
| Bitbucket | number has no minimum — 53 | `DELETE /repositories/{workspace}/{repo_slug}/commit/{commit}/comments/{comment_id}` path `comment_id`; `DELETE /repositories/{workspace}/{repo_slug}/pullrequests/{pull_request_id}/approve` path `pull_request_id`; `DELETE /repositories/{workspace}/{repo_slug}/pullrequests/{pull_request_id}/comments/{comment_id}/resolve` path `comment_id` | `:6957` | legitimate |
| Bitbucket | object dynamic `additionalProperties` — 5 | `POST /repositories/{workspace}/{repo_slug}/pullrequests/{pull_request_id}/merge` JSON; `PUT /repositories/{workspace}/{repo_slug}/commit/{commit}/properties/{app_key}/{property_name}` JSON; `PUT /repositories/{workspace}/{repo_slug}/properties/{app_key}/{property_name}` JSON | `:6720` | legitimate |
| Bitbucket | string has no `maxLength` — 758 | `DELETE /repositories/{workspace}/{repo_slug}/branch-restrictions/{id}` paths `id`, `repo_slug`, `workspace` | `:6912` | legitimate |
| CircleCI | array has no `maxItems` — 2 | `POST /webhook` JSON; `PUT /webhook/{webhook_id}` JSON | `:6936` | legitimate |
| CircleCI | no type — 2 | `GET /project/{project-slug}/pipeline/{pipeline-number}` path `pipeline-number`; `POST /project/{project-slug}/job/{job-number}/cancel` path `job-number` | `:6650` | legitimate |
| CircleCI | number has no minimum — 4 | `GET /deploy/components` query `page-size`; `GET /deploy/environments` query `page-size`; `GET /organizations/{org_id}/groups` query `limit` | `:6957` | legitimate |
| CircleCI | object dynamic `additionalProperties` — 18 | `GET /insights/pages/{project-slug}/summary` queries `branches`, `workflow-names`; `GET /insights/{org-slug}/summary` query `project-names` | `:6720` | legitimate |
| CircleCI | string has no `maxLength` — 197 | `DELETE /context/{context_id}/environment-variable/{env_var_name}` paths `context_id`, `env_var_name`; `DELETE /context/{context_id}/restrictions/{restriction_id}` path `context_id` | `:6912` | legitimate |
| Docker Hub | missing required `name` path parameter — 2 | `GET /v2/orgs/{name}/access-tokens`; `POST /v2/orgs/{name}/access-tokens` | `:6134` | malformed |
| Docker Hub | no type — 5 | `PATCH /v2/orgs/{org_name}/groups/{group_name}` JSON; `POST /v2/orgs/{org_name}/groups` JSON; `PUT /v2/orgs/{name}/settings` JSON | `:6650` | legitimate |
| Docker Hub | number has no maximum — 2 | `GET /v2/namespaces/{namespace}/repositories` query `page`; `GET /v2/scim/2.0/Users` query `startIndex` | `:6960` | legitimate |
| Docker Hub | number has no minimum — 15 | `GET /v2/access-tokens` queries `page_size`, `page`; `GET /v2/auditlogs/{account}` query `page_size` | `:6957` | legitimate |
| Docker Hub | object dynamic `additionalProperties` — 15 | `PATCH /v2/access-tokens/{uuid}` JSON; `PATCH /v2/namespaces/{namespace}/repositories/{repository}/immutabletags` JSON; `PATCH /v2/orgs/{org_name}/access-tokens/{access_token_id}` JSON | `:6720` | legitimate |
| Docker Hub | string has no `maxLength` — 78 | `DELETE /v2/access-tokens/{uuid}` path `uuid`; `DELETE /v2/invites/{id}` path `id`; `DELETE /v2/orgs/{org_name}/access-tokens/{access_token_id}` path `access_token_id` | `:6912` | legitimate |
| GitLab | `oneOf` — 55 | `GET /api/v4/admin/data_management/{model_name}` query `identifiers`; `GET /api/v4/groups/{id}/issues_statistics` queries `assignee_id`, `epic_id` | `:6620` | legitimate |
| GitLab | missing required `epic_issue_id` path parameter — 1 | `POST /api/v4/groups/{id}/(-/)epics/{epic_iid}/issues/{epic_issue_id}` | `:6134` | malformed |
| GitLab | missing required `sbom_digest` path parameter — 1 | `GET /api/v4/jobs/{id}/sbom_scans/{sbom_digest}` | `:6134` | malformed |
| GitLab | number has no maximum — 4 | `GET /api/v4/projects/{id}/jobs/{job_id}/trace` query `byte_offset`; `GET /api/v4/projects/{id}/repository/diverging_commits` query `max_count`; `GET /api/v4/projects/{id}/repository/files/{file_path}/blame` query `range[end]` | `:6960` | legitimate |
| GitLab | number has no minimum — 1,366 | `DELETE /api/v4/admin/clusters/{cluster_id}` path `cluster_id`; `DELETE /api/v4/admin/zoekt/shards/{node_id}/indexed_namespaces/{namespace_id}` paths `namespace_id`, `node_id` | `:6957` | legitimate |
| GitLab | object dynamic `additionalProperties` — 688 | `DELETE /api/v4/experiments/{experiment_name}/assignments` query `context`; `DELETE /api/v4/projects/{id}/variables/{key}` query `filter`; `GET /api/v4/analytics/code_review` query `not` | `:6720` | legitimate |
| GitLab | string has no `maxLength` — 1,602 | `DELETE /api/v4/admin/ci/variables/{key}` path `key`; `DELETE /api/v4/admin/sidekiq/queues/{queue_name}` path `queue_name`; same route query `ai_resource` | `:6912` | legitimate |
| Jira | `allOf` — 4 | `POST /rest/api/3/bulk/issues/fields` JSON; `POST /rest/api/3/expression/evaluate` JSON; `POST /rest/api/3/expression/eval` JSON | `:6620` | legitimate |
| Jira | `oneOf` — 1 | `PUT /rest/api/3/field/{fieldId}/context/defaultValue` JSON | `:6620` | legitimate |
| Jira | array has no `maxItems` — 10 | `GET /rest/api/3/classification-levels` query `status`; `GET /rest/api/3/field/search` query `type`; `GET /rest/api/3/project/search` query `status` | `:6936` | legitimate |
| Jira | no type — 14 | `DELETE /rest/api/3/issue/properties/{propertyKey}` JSON; `POST /rest/api/3/issuetype/{id}/avatar2` any media; `POST /rest/api/3/project/{projectIdOrKey}/avatar2` any media | `:6650` | legitimate |
| Jira | number has no maximum — 1 | `GET /rest/api/3/projects/fields` query `startAt` | `:6960` | legitimate |
| Jira | number has no minimum — 416 | `DELETE /rest/api/3/config/fieldschemes/{id}` path `id`; `DELETE /rest/api/3/dashboard/{dashboardId}/gadget/{gadgetId}` paths `dashboardId`, `gadgetId` | `:6957` | legitimate |
| Jira | object dynamic `additionalProperties` — 56 | `DELETE /rest/api/3/config/fieldschemes/fields/parameters` JSON; `DELETE /rest/api/3/config/fieldschemes/fields` JSON; `DELETE /rest/api/3/field/association` JSON | `:6720` | legitimate |
| Jira | object has no fixed properties — 2 | `GET /rest/api/3/project/recent` query `properties`; `GET /rest/api/3/project/search` query `properties` | `:6725` | legitimate |
| Jira | string has no `maxLength` — 777 | `DELETE /rest/api/3/attachment/{id}` path `id`; `DELETE /rest/api/3/comment/{commentId}/properties/{propertyKey}` paths `commentId`, `propertyKey` | `:6912` | legitimate |
| Notion | `allOf` — 3 | `PATCH /v1/pages/{page_id}/markdown` JSON; `POST /v1/blocks/meeting_notes` JSON; `POST /v1/comments` JSON | `:6620` | legitimate |
| Notion | `anyOf` — 5 | `PATCH /v1/blocks/{block_id}` JSON; `PATCH /v1/pages/{page_id}` JSON; `POST /v1/data_sources/{data_source_id}/query` JSON | `:6620` | legitimate |
| Notion | `oneOf` — 12 | `DELETE /v1/agents/{agent_id}` path `agent_id`; `GET /v1/agents/{agent_id}/insights` path `agent_id`; `GET /v1/agents/{agent_id}` path `agent_id` | `:6620` | legitimate |
| Notion | retained request-media byte limit — 2 | `PATCH /v1/blocks/{block_id}/children`; `POST /v1/pages` | `:1561` | finite-bound |
| Notion | number has no maximum — 2 | `GET /v1/agents/{agent_id}/insights` queries `end_time`, `start_time` | `:6960` | legitimate |
| Notion | number has no minimum — 3 | `GET /v1/blocks/{block_id}/children` query `page_size`; `GET /v1/pages/{page_id}/properties/{property_id}` query `page_size`; `GET /v1/users` query `page_size` | `:6957` | legitimate |
| Notion | object dynamic `additionalProperties` — 13 | `PATCH /v1/data_sources/{data_source_id}` JSON; `PATCH /v1/databases/{database_id}` JSON; `PATCH /v1/views/{view_id}` JSON | `:6720` | legitimate |
| Notion | string has no `maxLength` — 56 | `DELETE /v1/blocks/{block_id}` path `block_id`; `DELETE /v1/comments/{comment_id}` path `comment_id`; `DELETE /v1/views/{view_id}/queries/{query_id}` path `query_id` | `:6912` | legitimate |
| Sentry | `anyOf` — 23 | `DELETE /api/0/organizations/{organization_id_or_slug}/detectors/` query `project`; `DELETE /api/0/organizations/{organization_id_or_slug}/issues/` query `project`; `DELETE /api/0/organizations/{organization_id_or_slug}/workflows/` query `project` | `:6620` | legitimate |
| Sentry | array has no `maxItems` — 11 | `GET /api/0/organizations/{organization_id_or_slug}/issues/{issue_id}/` queries `collapse`, `expand`; `GET /api/0/organizations/{organization_id_or_slug}/issues/` query `collapse` | `:6936` | legitimate |
| Sentry | no type — 2 | `GET /api/0/organizations/{organization_id_or_slug}/stats-summary/` query `project`; `GET /api/0/organizations/{organization_id_or_slug}/stats_v2/` query `project` | `:6650` | legitimate |
| Sentry | number has no maximum — 4 | `GET /api/0/organizations/{organization_id_or_slug}/scim/v2/Groups` queries `count`, `startIndex`; `GET /api/0/organizations/{organization_id_or_slug}/scim/v2/Users` query `count` | `:6960` | legitimate |
| Sentry | number has no minimum — 62 | `DELETE /api/0/organizations/{organization_id_or_slug}/dashboards/{dashboard_id}/` path `dashboard_id`; `DELETE /api/0/organizations/{organization_id_or_slug}/detectors/{detector_id}/` path `detector_id`; same detector collection query `id` | `:6957` | legitimate |
| Sentry | object dynamic `additionalProperties` — 66 | `DELETE /api/0/organizations/{organization_id_or_slug}/spike-protections/` JSON; `PATCH /api/0/organizations/{organization_id_or_slug}/scim/v2/Groups/{team_id_or_slug}` JSON; `PATCH /api/0/organizations/{organization_id_or_slug}/scim/v2/Users/{member_id}` JSON | `:6720` | legitimate |
| Sentry | string has no `maxLength` — 659 | `DELETE /api/0/organizations/{organization_id_or_slug}/dashboards/{dashboard_id}/` path `organization_id_or_slug`; `DELETE /api/0/organizations/{organization_id_or_slug}/detectors/{detector_id}/` path `organization_id_or_slug`; same detector collection path `organization_id_or_slug` | `:6912` | legitimate |
| Stripe | `anyOf` — 84 | `GET /v1/accounts` query `created`; `GET /v1/application_fees` query `created`; `GET /v1/balance/history` query `created` | `:6620` | legitimate |
| Stripe | array has no `maxItems` — 437 | `DELETE /v1/customers/{customer}/bank_accounts/{id}` form; `DELETE /v1/customers/{customer}/cards/{id}` form; `DELETE /v1/customers/{customer}/sources/{id}` form | `:6936` | legitimate |
| Stripe | number has no minimum — 190 | `DELETE /v1/subscription_items/{item}` form; `GET /v1/accounts/{account}/external_accounts` query `limit`; `GET /v1/accounts/{account}/people` query `limit` | `:6957` | legitimate |
| Stripe | object dynamic `additionalProperties` — 81 | `DELETE /v1/subscriptions/{subscription_exposed_id}` form; `GET /v1/accounts/{account}/people` query `relationship`; `GET /v1/accounts/{account}/persons` query `relationship` | `:6720` | legitimate |
| Stripe | string has no `maxLength` — 75 | `DELETE /v1/accounts/{account}/bank_accounts/{id}` path `id`; `DELETE /v1/accounts/{account}/external_accounts/{id}` path `id`; `DELETE /v1/customers/{customer}/bank_accounts/{id}` path `id` | `:6912` | legitimate |
| Vercel | `anyOf` — 7 | `DELETE /v1/global-config/{edgeConfigId}/tokens` JSON; `GET /v1/query/web-analytics/events/aggregate` query `by`; `GET /v1/query/web-analytics/visits/aggregate` query `by` | `:6620` | legitimate |
| Vercel | `oneOf` — 29 | `DELETE /v1/security/firewall/bypass` JSON; `DELETE /v2/aliases/{aliasId}` path `aliasId`; `GET /v1/bulk-redirects` query `diff` | `:6620` | legitimate |
| Vercel | array has no `maxItems` — 2 | `POST /v1/log-drains` JSON; `POST /v1/webhooks` JSON | `:6936` | legitimate |
| Vercel | no type — 1 | `POST /v1/global-config/{edgeConfigId}/schema` JSON | `:6650` | legitimate |
| Vercel | number has no maximum — 6 | `GET /v1/ai-gateway/virtual-model-configs/list` query `limit`; `GET /v1/bulk-redirects` query `page`; `GET /v1/security/firewall/attack-status` query `since` | `:6960` | legitimate |
| Vercel | number has no minimum — 49 | `GET /v1/domains/{domain}/project-domains` queries `limit`, `since`, `until` | `:6957` | legitimate |
| Vercel | object dynamic `additionalProperties` — 78 | `DELETE /v1/bulk-redirects` JSON; `DELETE /v1/env` JSON; `DELETE /v1/projects/{projectId}/routes` JSON | `:6720` | legitimate |
| Vercel | string has no `maxLength` — 1,155 | `DELETE /storage/stores/blob/{id}` path `id`; `DELETE /v1/access-groups/{accessGroupIdOrName}/projects/{projectId}` paths `accessGroupIdOrName`, `projectId` | `:6912` | legitimate |

## Deliberate next-foundation boundary

The seven implementation cases in this PR remain prerequisites only. This
inventory shows that full Batch-1 import is still blocked by request-contract
semantics: composition selection, finite-value policy, dynamic objects, and
four malformed path declarations. A next foundation must choose, per category,
whether the construct can be represented correctly or must be retained with a
source-traced merge-blocking gap. It must keep finite resource bounds and must
not treat the large legitimate populations above as malformed provider input.
