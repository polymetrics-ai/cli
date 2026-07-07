# Next Connector Rollout Batches

Sequenced candidate connectors for rolling out the GitHub pilot's CLI parity shape beyond GitHub.
Ordered by provider API maturity, parity value, and isolation (each batch is a set of disjoint
connector dirs so workers can run in parallel without shared-file collisions).

## Selection criteria

- A documented REST or GraphQL API with stable auth (token / OAuth / app).
- A provider CLI to model `gh`-like commands against (preferred, not required).
- Disjoint `internal/connectors/defs/<name>/` dir per connector (no shared-file collisions).
- No new runtime dependencies without a human gate.

## Batch 1 — high parity value, mature REST APIs

| Connector | Provider | Why | Notes |
| --- | --- | --- | --- |
| gitlab | GitLab | Closest parity analog to GitHub; REST + GraphQL. | Large surface; split read vs write sub-issues. |
| slack | Slack | High demand; web API + Conversations API. | Token + OAuth; `chat.postMessage` is sensitive write. |
| stripe | Stripe | Already a declarative-HTTP template; expand CLI surface. | Idempotency keys on writes. |
| jira | Jira (Atlassian) | REST v3 + Agile API; issue tracking parity. | Cloud auth (API token / OAuth). |
| linear | Linear | GraphQL API; modern issue tracker. | Fixed GraphQL queries fit the engine. |

## Batch 2 — platform / CRM / collab

| Connector | Provider | Why | Notes |
| --- | --- | --- | --- |
| hubspot | HubSpot | CRM REST + GraphQL; high reverse-ETL demand. | Private apps (token); OAuth for more scopes. |
| notion | Notion | REST API; block/page model. | Sensitive writes (page mutations). |
| salesforce | Salesforce | REST + SOQL; enterprise CRM. | OAuth + bulk API; large surface, shard. |
| pagerduty | PagerDuty | REST v2; incidents/oncall. | Incident write actions are sensitive. |
| datadog | Datadog | REST APIs (metrics/logs/events). | Multiple sub-APIs; split by product. |

## Batch 3 — cloud / ads / google

| Connector | Provider | Why | Notes |
| --- | --- | --- | --- |
| google | Google (multiple) | Admin/Workspace/Ads; many disjoint APIs. | Shard by product (Admin, Drive, Ads, Calendar). |
| aws | AWS | Service APIs; IAM/STS/S3. | Credential chain; shard by service. |

## Sequencing rules

- One parent issue per batch (mirrors #44), with per-connector sub-issues.
- Batch 1 connectors are disjoint from each other and from GitHub — safe to fan out up to the
  runtime concurrency cap (Pi: 8 tasks / 4 concurrent).
- Start each connector from `templates/connector-rollout-prompt.md` and follow
  `rollout-checklist.md` + `validation-gates.md`.
- Do not start a batch until the previous batch's shared-schema changes (if any) are merged, so
  in-flight workers do not collide on `internal/connectors/engine/**`.

## Out of scope until human-gated

- New runtime dependencies (e.g. AWS SDK, Google SDK) — each needs a dependency human gate.
- Auth-scope refresh for any provider OAuth flow.
- Production deploys or live reverse-ETL execution outside plan/preview/approval/execute.
