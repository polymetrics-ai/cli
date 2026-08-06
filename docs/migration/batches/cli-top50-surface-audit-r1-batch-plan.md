# Top-50 provider surface batch map — R1

Status: planning only. No Top-50 connector is authored or promoted by this
map. It is the reviewable input to later `connectorgen batch plan` manifests.

## Evidence baseline

This map is derived from the externally maintained audit at
`/Users/karthiksivadas/karthik-agent-workspace/data/cli-top50-surface-audit-r1/report.md`
(read on 2026-08-06; audit date 2026-08-05). Its companion JSON records 50
fixed provider slugs, 46 defensibly counted surfaces, and **13,761 documented
callable operations** (6,106 reads / 7,655 writes). Four providers have no
finite count: Bahmni, Workday REST, QuickBooks, and PostgreSQL.

The counts are provider-inventory planning volume, not implemented-command
claims. A later manifest must still be made from the live provider-artifact
ledger and must carry a public artifact URL, version, retrieval date, and
measured read/write totals. The audit is not an alternate connector schema and
does not replace `connectorgen batch plan`.

Batch 001 (DocuSeal, DefiLlama, Docker Hub, Flexmail, and Alpaca Broker API)
is a five-connector pipeline proof, not a member of this Top-50 cohort. Its
post-main re-gate is complete: five included, zero drops, 203 freshly declared
artifact operations, and a 39 executable / 27 provider-blocked / 137 excluded
split.

## Batching rule

`batch plan` has a 40-connector mechanical limit, but a connector count is not
a safe measure of review or provenance volume. A 1,913-operation provider must
not be treated as a 23-operation provider merely because each has one bundle.

| Provider surface size | Proposed mergeable unit | Authoring approach |
| --- | --- | --- |
| 800+ operations | One provider, one branch, one PR | Work by provider module/domain inside the branch, but materialize and gate the entire cited provider artifact before the PR. |
| 500–799 | One provider, one branch, one PR | Dedicated provider task; do not combine with another connector. |
| 250–499 | One provider, one branch, one PR | Full provider artifact, individual gate, then the batch gate of one. |
| 100–249 | Two or three providers, normally 350–500 aggregate operations | Group only after every member has its own eligible ledger record and can pass an individual gate. |
| Under 100 | Four to six providers, normally at most 350 aggregate operations | This is the only Top-50 stratum suitable for a conventional multi-connector batch. |

The provider-module/domain sequencing for very large connectors is an internal
authoring checklist, not separate partial connector PRs. The current pipeline
correctly materializes and classifies the whole cited artifact; inventing a
partial omission reason to make an early slice mergeable would violate the
operation-completeness rule. No new `connectorgen` command is needed for this
map: a single-candidate manifest can raise `--max-operations` explicitly, then
the existing materializer and real-runtime batch gate cover the full bundle.

## Surface tiers

| Tier | Connectors | Provider operations | Batch shape |
| --- | ---: | ---: | --- |
| XL (800+) | 4 | 5,085 | Four dedicated provider batches |
| L (500–799) | 4 | 2,341 | Four dedicated provider batches |
| M (250–499) | 10 | 3,154 | Ten dedicated provider batches |
| S (100–249) | 15 | 2,555 | Six conditional 2–3 connector batches |
| XS (under 100) | 13 | 626 | Two ready-after-refresh batches plus one three-provider repair batch |
| Count unknown / not applicable | 4 | unknown | Research/inventory work only |
| **Measured cohort total** | **46** | **13,761** | **27 proposed authoring batches after evidence gates** |

## Pre-authoring evidence gates

These are provider-inventory tasks, not connector-authoring batches. They must
finish before the named provider can enter a batch manifest.

| Gate | Providers | Why it is not authoring-ready |
| --- | --- | --- |
| Replace invalid foundations | `shopify` (805), `crisp` (234), `aws-cloudtrail` (60), `google-calendar` (38), `google-analytics-data-api` (24) | The audit marks Shopify and Crisp `absent`, and the three Google/AWS surfaces `self_referential`; each needs a provider-derived artifact ledger before any percentage or executable surface can be claimed. |
| Establish a finite inventory | `bahmni`, `workday-rest`, `quickbooks` | The audit explicitly reports the total as `unknown`; do not manufacture a batch size. |
| Use a different parity unit | `postgres` | A fixed HTTP-operation inventory is not applicable to the PostgreSQL wire/SQL surface. |
| Refresh the six largest gaps | `zoom`, `gitlab`, `shopify`, `github`, `jira`, `linear` | The audit directs inventory work first. GitLab already has work on current `main`; refresh its provider total rather than duplicating that lane. |

For every later candidate, `connectorgen batch plan` remains the entry gate.
`partial` audit health is a signal to refresh the provider inventory; `real`
health is not permission to skip the current public-artifact retrieval and
individual runtime gate.

## Proposed batches

### XL — dedicated provider batches

| ID | Provider | Operations (read / write) | Readiness and branch shape |
| --- | --- | ---: | --- |
| `top50-xl-01` | `zoom` | 1,913 (881 / 1,032) | Provider inventory refresh first. Author by its 35 documented modules on one branch; final materialization/gate covers all modules. |
| `top50-xl-02` | `github` | 1,220 (636 / 584) | Refresh the REST-only lower-bound inventory; do not mix in authenticated GraphQL schema to create a fake exact total. |
| `top50-xl-03` | `gitlab` | 1,147 (507 / 640) | Refresh after the landed GitLab stream-read lane; do not duplicate existing work. |
| `top50-xl-04` | `shopify` | 805 (287 / 518) | Repair the absent provider-derived GraphQL inventory first; preserve the audited Admin GraphQL-only unit. |

### L — dedicated provider batches

| ID | Provider | Operations (read / write) | Readiness |
| --- | --- | ---: | --- |
| `top50-l-01` | `zendesk-support` | 625 (325 / 300) | Refresh the current OAS, then dedicated batch. |
| `top50-l-02` | `jira` | 616 (275 / 341) | Rebuild the provider inventory before materialization. |
| `top50-l-03` | `stripe` | 589 (263 / 326) | Current public OAS verification, then dedicated batch. |
| `top50-l-04` | `linear` | 511 (153 / 358) | Reproduce the official-SDK-based inventory and preserve its documented exclusions before materialization. |

### M — one provider per batch

| ID | Provider | Operations | Audit health |
| --- | --- | ---: | --- |
| `top50-m-01` | `chargebee` | 438 | real / high |
| `top50-m-02` | `square` | 334 | partial / high |
| `top50-m-03` | `bitbucket` | 331 | real / high |
| `top50-m-04` | `intercom` | 324 | partial / high |
| `top50-m-05` | `marketo` | 322 | partial / high |
| `top50-m-06` | `bamboo-hr` | 311 | real / medium |
| `top50-m-07` | `mailchimp` | 298 | partial / high |
| `top50-m-08` | `monday` | 280 | partial / medium |
| `top50-m-09` | `trello` | 261 | partial / high |
| `top50-m-10` | `front` | 255 | partial / high |

### S — two or three complete provider bundles per batch

| ID | Providers | Operations | Conditions |
| --- | --- | ---: | --- |
| `top50-s-01` | `asana` (249), `xero` (235) | 484 | Both have real, high-confidence artifact inventories; still generate fresh manifests. |
| `top50-s-02` | `recurly` (197), `segment` (197) | 394 | Refresh Recurly's partial inventory before admission. |
| `top50-s-03` | `twilio` (197), `ashby` (186) | 383 | Verify the current artifacts; Ashby's callable/event split remains explicit. |
| `top50-s-04` | `google-ads` (163), `greenhouse` (138), `gorgias` (114) | 415 | Preserve Google Ads semantic RPC counting; refresh Gorgias before admission. |
| `top50-s-05` | `help-scout` (146), `chatwoot` (146), `reddit` (145) | 437 | All three need refreshed provider inventories before a manifest is generated. |
| `top50-s-06` | `crisp` (234), `mixpanel` (105), `airtable` (103) | 442 | Conditional on a new provider-derived Crisp inventory; do not use the currently absent API surface as evidence. |

### XS — conventional multi-connector batches only after evidence refresh

| ID | Providers | Operations | Conditions |
| --- | --- | ---: | --- |
| `top50-xs-01` | `hubplanner` (98), `gmail` (79), `gong` (69), `lever-hiring` (67) | 313 | Preserve Hubplanner's callable-vs-webhook distinction; refresh Lever's partial inventory. |
| `top50-xs-02` | `dynamodb` (57), `notion` (50), `freshchat` (34), `amazon-sqs` (23), `youtube-analytics` (16), `google-search-console` (11) | 191 | Reproduce semantic AWS/Google method splits before admission. |
| `top50-xs-03` | `aws-cloudtrail` (60), `google-calendar` (38), `google-analytics-data-api` (24) | 122 | Conditional repair batch only: replace the three self-referential ledgers before generating a manifest. |

Every listed batch stays one branch and one PR, produces individual candidate
reports before its final batch report, and uses the existing explicit drop
protocol. A candidate that fails its gate is removed with its named report;
the other candidates are re-gated rather than being silently shipped.
