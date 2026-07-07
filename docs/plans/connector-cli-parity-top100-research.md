# Connector CLI Parity — Top 100 Research & Top 5 Selection

Status: research + selection. Produced from the Polymetrics catalog (547 native connectors),
the SaaS/API landscape, and the GitHub CLI parity pilot (#44) as the rework template.

## Marketing thesis

Polymetrics `pm` is a CLI. The differentiator for each connector is **"the CLI they never built"**:
providers with strong REST/GraphQL APIs and high ETL + reverse-ETL demand, but **no first-party
CLI**. For those, `pm <connector>` becomes the canonical command surface — read, write, inspect,
help, and local download — the way `gh` is for GitHub. Providers that already ship a strong CLI
(`gh`, `glab`, `stripe`, `slack`, `aws`, `gcloud`, `vercel`, `wrangler`, `firebase`, `supabase`,
`twilio`, `contentful`, `sanity`) are **excluded** or down-ranked because the marketing hook is
weak; Polymetrics can still add reverse-ETL, but it is not "the first CLI."

## Selection criteria (for the rework top 5)

Each candidate is scored 0–3 on:
- **API depth**: REST + fixed GraphQL support (the declarative engine supports both).
- **ETL demand**: read-side pull (lists, items, records, metrics).
- **Reverse-ETL demand**: write-side push (create/update/delete issues, records, secrets, files).
- **CLI gap**: no first-party CLI (3 = none exists; 2 = a limited/third-party CLI exists; 0 = strong CLI).
- **Local download**: binary/file/git artifact download (release assets, attachments, exports).
- **Marketing pull**: developer mindshare + "no CLI" surprise factor.

A connector qualifies for the deep rework (top 5) only if it has **REST + GraphQL + local
download** and **CLI gap ≥ 2**.

## Catalog cross-check (current Polymetrics state)

The 547-connector catalog already covers most candidates as **read-only native connectors**. The
"rework" is the GitHub-parity elevation: add `cli_surface`, help renderer, stream runner,
operation ledger, direct read, GraphQL engine, and sensitive/admin policy — **not** a new
connector. Candidates not in the catalog (Bitbucket, Figma, …) are greenfield + parity.

---

## Top 100 (ranked, marketing-ranked, no first-party CLI unless noted)

Legend: API = R(rest)/G(graphql)/X(xml); ETL = read; Rev = reverse-ETL write; DL = local download;
CLIgap = 0(strong CLI)/1(partial)/2(limited/3rd-party)/3(none); Parity = rework fit (★/★★/★★★).

### A. Project / Work / Dev management (25)

| # | Provider | API | ETL | Rev | DL | CLIgap | Notes |
|---|---|---|---|---|---|---|---|
| 1 | Linear | G | y | y | att | 3 | GraphQL-first; no CLI. ★★★ |
| 2 | Jira (Atlassian) | R+G | y | y | att | 2 | REST v3 + Agile + Cloud GraphQL; no first-party CLI. ★★★ |
| 3 | Monday | R+G | y | y | file | 3 | REST + GraphQL; no CLI. ★★★ |
| 4 | Bitbucket | R+G | y | y | git | 3 | No CLI; greenfield in catalog. ★★★ |
| 5 | GitLab | R+G | y | y | git | 1 | `glab` exists but limited; reverse-ETL + parity still strong. ★★★ |
| 6 | Asana | R | y | y | att | 3 | No CLI. ★★ |
| 7 | ClickUp | R | y | y | att | 3 | No CLI. ★★ |
| 8 | Airtable | R | y | y | att | 3 | No CLI. ★★ |
| 9 | Notion | R | y | y | file | 3 | REST only; no GraphQL; no CLI. ★★ |
| 10 | Shortcut (Clubhouse) | R | y | y | - | 3 | No CLI. ★ |
| 11 | Height | R | y | y | - | 3 | No CLI. ★ |
| 12 | Wrike | R | y | y | exp | 3 | No CLI. ★ |
| 13 | Smartsheet | R | y | y | exp | 3 | Greenfield; no CLI. ★ |
| 14 | Basecamp | R | y | n | att | 3 | No CLI. ★ |
| 15 | YouTrack | R | y | y | att | 3 | Greenfield; no CLI. ★ |
| 16 | Trello | R | y | y | att | 3 | No CLI. ★ |
| 17 | Teamwork | R | y | y | - | 3 | No CLI. ★ |
| 18 | AceProject | R | y | n | - | 3 | No CLI. ★ |
| 19 | GanttPRO | R | y | n | - | 3 | No CLI. |
| 20 | Backlog | R | y | y | att | 3 | No CLI. |
| 21 | Leankit (Planview) | R | y | n | - | 3 | No CLI. |
| 22 | Miro | R | y | n | file | 3 | No CLI; board export. |
| 23 | FigJam | R | y | n | file | 3 | No CLI. |
| 24 | OpenProject | R | y | y | - | 3 | Self-hosted; no CLI. |
| 25 | Taiga | R | y | y | - | 3 | Open source; no CLI. |

### B. CRM / Sales / Marketing (15)

| # | Provider | API | ETL | Rev | DL | CLIgap | Notes |
|---|---|---|---|---|---|---|---|
| 26 | HubSpot | R+G | y | y | exp | 3 | REST + GraphQL; no CLI. ★★ |
| 27 | Salesforce | R+G | y | y | exp | 2 | SOQL + REST + GraphQL; no first-party CLI (3rd-party sfdx). ★★ |
| 28 | Pipedrive | R | y | y | exp | 3 | No CLI. ★ |
| 29 | Attio | R+G | y | y | - | 3 | New CRM; no CLI. ★ |
| 30 | Close | R | y | y | - | 3 | No CLI. |
| 31 | Insightly | R | y | y | - | 3 | No CLI. |
| 32 | Keap (Infusionsoft) | R | y | y | - | 3 | No CLI. |
| 33 | Copper | R | y | y | - | 3 | No CLI. |
| 34 | Capsule | R | y | y | - | 3 | No CLI. |
| 35 | Zoho CRM | R | y | y | exp | 3 | No CLI. |
| 36 | Brevo (Sendinblue) | R | y | y | - | 3 | No CLI. |
| 37 | ActiveCampaign | R | y | y | - | 3 | No CLI. |
| 38 | Mailchimp | R | y | y | - | 3 | No CLI. |
| 39 | Klaviyo | R | y | y | exp | 3 | No CLI. |
| 40 | Marketo (Adobe) | R | y | y | exp | 3 | No CLI. |

### C. Comms / Support / CX (15)

| # | Provider | API | ETL | Rev | DL | CLIgap | Notes |
|---|---|---|---|---|---|---|---|
| 41 | Zendesk | R | y | y | att | 3 | REST; no CLI. ★ |
| 42 | Intercom | R | y | y | - | 3 | No CLI. ★ |
| 43 | Front | R | y | y | att | 3 | No CLI. |
| 44 | Help Scout | R | y | y | att | 3 | No CLI. |
| 45 | Missive | R | y | y | - | 3 | No CLI. |
| 46 | Drift | R | y | n | - | 3 | No CLI. |
| 47 | LiveChat | R | y | y | - | 3 | No CLI. |
| 48 | Olark | R | y | n | - | 3 | No CLI. |
| 49 | SnapEngage | R | y | n | - | 3 | No CLI. |
| 50 | Zoho Desk | R | y | y | att | 3 | No CLI. |
| 51 | Freshdesk | R | y | y | att | 3 | No CLI. |
| 52 | Kayako | R | y | y | - | 3 | No CLI. |
| 53 | Userlike | R | y | n | - | 3 | No CLI. |
| 54 | Tawk.to | R | y | n | - | 3 | No CLI. |
| 55 | Crisp | R | y | y | - | 3 | No CLI. |

### D. DevOps / Observability / Incident (15)

| # | Provider | API | ETL | Rev | DL | CLIgap | Notes |
|---|---|---|---|---|---|---|---|
| 56 | PagerDuty | R | y | y | - | 3 | No CLI. ★ |
| 57 | Datadog | R | y | y | - | 3 | No CLI (dogshell retired). ★ |
| 58 | New Relic | R+G | y | y | - | 3 | NerdGraph GraphQL; no CLI. ★ |
| 59 | Grafana | R | y | y | - | 3 | No first-party CLI. ★ |
| 60 | Splunk | R | y | y | exp | 3 | No CLI. |
| 61 | Buildkite | R+G | y | y | art | 3 | REST + GraphQL; artifacts; no CLI. ★ |
| 62 | CircleCI | R+G | y | y | art | 3 | REST + GraphQL; artifacts; no CLI. ★ |
| 63 | Travis CI | R | y | n | log | 3 | No CLI. |
| 64 | Sentry | R | y | y | - | 3 | No CLI. |
| 65 | LaunchDarkly | R | y | y | - | 3 | No CLI. |
| 66 | Statuspage (Atlassian) | R | y | y | - | 3 | No CLI. |
| 67 | incident.io | R | y | y | - | 3 | No CLI. |
| 68 | FireHydrant | R | y | y | - | 3 | No CLI. |
| 69 | Better Uptime | R | y | y | - | 3 | No CLI. |
| 70 | UptimeRobot | R | y | n | - | 3 | No CLI. |

### E. Payments / Billing / Fintech (10)

| # | Provider | API | ETL | Rev | DL | CLIgap | Notes |
|---|---|---|---|---|---|---|---|
| 71 | Square | R | y | y | - | 3 | No CLI. |
| 72 | Adyen | R | y | y | - | 3 | No CLI. |
| 73 | Plaid | R | y | n | - | 3 | No CLI. |
| 74 | Lemon Squeezy | R | y | y | - | 3 | No CLI. |
| 75 | Paddle | R | y | y | - | 3 | No CLI. |
| 76 | PayPal | R+G | y | y | - | 3 | REST + GraphQL; no CLI. |
| 77 | Chargebee | R | y | y | exp | 3 | No CLI. |
| 78 | Recurly | R | y | y | - | 3 | No CLI. |
| 79 | Stripe (excluded) | R+G | y | y | - | 0 | `stripe` CLI exists — not a parity target. |
| 80 | Ramp | R | y | y | - | 3 | No CLI. |

### F. Analytics / Product (5)

| # | Provider | API | ETL | Rev | DL | CLIgap | Notes |
|---|---|---|---|---|---|---|---|
| 81 | Amplitude | R+G | y | y | - | 3 | REST + GraphQL; no CLI. |
| 82 | Mixpanel | R | y | y | - | 3 | No CLI. |
| 83 | Heap | R | y | n | - | 3 | No CLI. |
| 84 | PostHog | R+G | y | y | - | 3 | REST + GraphQL; no CLI. |
| 85 | Pendo | R | y | n | - | 3 | No CLI. |

### G. Storage / Files / Design (5)

| # | Provider | API | ETL | Rev | DL | CLIgap | Notes |
|---|---|---|---|---|---|---|---|
| 86 | Figma | R | y | n | file | 3 | Greenfield; file download; no CLI. ★★ |
| 87 | Dropbox | R | y | y | file | 3 | No CLI. |
| 88 | Box | R | y | y | file | 3 | No CLI. |
| 89 | WeTransfer | R | y | n | file | 3 | No CLI. |
| 90 | Frame.io | R | y | y | file | 3 | No CLI. |

### H. CMS / Content / Docs (5)

| # | Provider | API | ETL | Rev | DL | CLIgap | Notes |
|---|---|---|---|---|---|---|---|
| 91 | Storyblok | R+G | y | y | - | 3 | REST + GraphQL; no CLI. |
| 92 | Strapi | R+G | y | y | - | 3 | REST + GraphQL; no CLI. |
| 93 | Directus | R+G | y | y | file | 3 | REST + GraphQL; no CLI. |
| 94 | Ghost | R | y | y | - | 3 | No CLI. |
| 95 | Webflow | R | y | y | - | 3 | No CLI. |

### I. HR / IT / Other (5)

| # | Provider | API | ETL | Rev | DL | CLIgap | Notes |
|---|---|---|---|---|---|---|---|
| 96 | Greenhouse | R | y | y | - | 3 | No CLI. |
| 97 | Lever | R | y | y | - | 3 | No CLI. |
| 98 | BambooHR | R | y | y | - | 3 | No CLI. |
| 99 | Workable | R | y | y | - | 3 | No CLI. |
| 100 | Notion (dup #9 kept for HR wiki) | R | y | y | file | 3 | — |

> Note: #100 is a deliberate placeholder noting Notion's HR-wiki use case; replace with the next
> ranked no-CLI provider (e.g. Calendly, Eventbrite, or Cal.com) when the list is consumed.

---

## Top 5 selection (deep rework, GitHub-parity style)

Required: **REST + GraphQL + local download + CLI gap ≥ 2 + high ETL/reverse-ETL demand**.
GitLab and Bitbucket are mandated; the remaining 3 are chosen on merit.

### 1. GitLab — `feat/gitlab-cli-parity` (★★★)
- **API**: REST v4 + GraphQL API + git clone/artifact/release download.
- **Catalog state**: native connector exists, read-only, **no write, no CLI surface**.
- **CLI gap**: `glab` exists but is limited (no reverse-ETL plan/preview/execute, no unified
  operation ledger); Polymetrics adds write parity + sensitive/admin policy.
- **Marketing**: "the write side `glab` never had" — CI/CD variables/secrets as a sensitive surface.

### 2. Bitbucket — `feat/bitbucket-cli-parity` (★★★)
- **API**: REST 2.0 + GraphQL (Bitbucket Cloud GraphQL) + git clone/downloads.
- **Catalog state**: **greenfield** (not in catalog).
- **CLI gap**: **no first-party CLI** — strongest marketing hook of the five.
- **Marketing**: "the CLI Atlassian never shipped for Bitbucket."

### 3. Linear — `feat/linear-cli-parity` (★★★)
- **API**: **GraphQL-first** (no REST) + attachments download.
- **Catalog state**: native connector exists, read-only.
- **CLI gap**: **no CLI**; GraphQL-first is the perfect showcase for the declarative GraphQL engine.
- **Marketing**: "the Linear CLI — GraphQL-native, plan/preview/approve writes."

### 4. Jira (Atlassian) — `feat/jira-cli-parity` (★★★)
- **API**: REST v3 + Agile API + Atlassian Cloud GraphQL + issue attachments.
- **Catalog state**: native connector exists, read-only.
- **CLI gap**: no first-party CLI (third-party `jira-cli` exists, gaps in writes).
- **Marketing**: "the Jira CLI for read + write + attachments, JSON-first."

### 5. Monday — `feat/monday-cli-parity` (★★★)
- **API**: REST + **GraphQL** + file-column download.
- **Catalog state**: native connector exists, read-only.
- **CLI gap**: **no CLI**; GraphQL-first.
- **Marketing**: "the Monday CLI — boards, items, and file columns from one command."

### Why not Stripe/Slack/Vercel/etc.?
They ship strong first-party CLIs (`stripe`, `slack`, `vercel`), so "the CLI they never built" is
false. Polymetrics can still add reverse-ETL later, but they are **not** top-5 parity targets.

---

## Issue & sub-issue plan (mirrors #44)

For each of the 5 connectors, create one **parent issue** (roadmap table, like #44) and **7
sub-issues** (the GitHub-parity milestones, trimmed from #44's 9 to the core 7):

1. CLI surface metadata (`<name> cli_surface` validated).
2. Help renderer (`pm <name> --help` from metadata).
3. Stream runner (execute stream-backed commands).
4. Operation ledger (reclassify REST/GraphQL rows into execution models).
5. Direct read (constrained safe reads + redaction).
6. GraphQL engine (fixed GraphQL queries/mutations).
7. Sensitive/admin reverse-ETL policy (secrets/variables/admin, blocked by default).

Cross-connector rollout (#42-equivalent) is **shared** across all 5 and tracked once, not per
connector. Binary/local-download support folds into each connector's operation ledger + direct
read (kind `binary_download` / `local_git`), mirroring GitHub's `github.repo.clone`/release
download.

Branch policy mirrors #44: parent branch from `main`, sub-issue branches from the parent branch,
sub-PRs target the parent branch with `Refs #<sub>` and `Refs #<parent>`, closing keywords
reserved for the parent PR into `main`.
