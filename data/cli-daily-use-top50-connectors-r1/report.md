# Daily-use local connector certification shortlist (r1)

Prepared 2026-08-10 from the two requested Git refs. This is a certification
selection report, not a claim that the listed providers are equally popular or
equally ready.

## Decision in one paragraph

Choose for local usability first, not raw operation count. The best early
certification work is Docker Hub, DocuSeal, Amazon SQS, then Google Calendar
after an OAuth ownership decision. They are the only high-value candidates in
this list with a fully implemented mapped command surface (4/4, 9/9, 23/23,
and 38/38 respectively). The artifact branch has 33 of the recommended 50
already surface-mapped, so those are the cheap path. However, the requested
17-file baseline is a filename match, not 17 valid live certifications:
six of its matches are Employment Hero fixtures/schema files, the 11 remaining
bundle files are harness contracts, and there are zero accepted live
certification artifacts in internal/connectors/certifications. Under the
written definition of certified, the honest starting count is therefore zero,
not 17.

That distinction matters: do not certify by adding a file. Certify only after
the real built binary has reached every command, no unsafe/disallowed command
remains, rate policy is declared, parity/docs are complete, and a fresh
accepted live artifact exists.

## Measurement and count reconciliation

Snapshot:

| Ref | Commit | CLI surfaces |
|---|---|---:|
| origin/main | f96a47e801b89f25386c33951a53a93d1a4c7c8d (2026-08-10) | 36 |
| origin/fm/cli-mass-artifact-materialize-r1 | 89eb661a0d08add2eba2acb696186a5a27bab56f (2026-08-10) | 234 |

The exact requested commands both returned 17:

    git ls-tree -r --name-only origin/main internal/connectors/defs/ | grep -c 'certification\.json$'
    17
    git ls-tree -r --name-only origin/fm/cli-mass-artifact-materialize-r1 internal/connectors/defs/ | grep -c 'certification\.json$'
    17

The path list explains the discrepancy. Eleven are exactly
internal/connectors/defs/<connector>/certification.json. The other six are
Employment Hero fixture writes named archive/create/delete/update
certification, plus schemas/certification.json. They cannot certify Employment
Hero, much less six separate connectors.

| Interpretation | Count | Consequence |
|---|---:|---|
| Requested suffix-match result | 17 | Reproducible, but contains six false positives. |
| Root bundle certification.json contracts | 11 | Harness configuration only; not proof of a live run. |
| Accepted live artifacts in internal/connectors/certifications | 0 | This is the repository's actual enablement evidence. |
| Recommended shortlist | 50 | 0 already certified; 33 mapped; 17 not mapped. |

The repository design calls certification.json an optional definition-owned
contract, not an accepted result: see
[connector-certification-design.md:27-31](../../../.treehouse/cli-83d592/22/cli/docs/architecture/connector-certification-design.md)
and
[conventions.md:326-330](../../../.treehouse/cli-83d592/22/cli/docs/migration/conventions.md).
It says enablement requires an accepted artifact under
internal/connectors/certifications with passed=true, mode=live, matching
manifest hash, current schema, and age under 90 days
([design:177-186](../../../.treehouse/cli-83d592/22/cli/docs/architecture/connector-certification-design.md)).
Both requested refs contain zero accepted-artifact JSON files:

    git ls-tree -r --name-only origin/main internal/connectors/certifications | rg '/[^/]+\.json$' | wc -l
    0
    git ls-tree -r --name-only origin/fm/cli-mass-artifact-materialize-r1 internal/connectors/certifications | rg '/[^/]+\.json$' | wc -l
    0

The 33-more math is usable only as a planning convention if the captain elects
to count the supplied 17 suffix matches. It does not satisfy the stated
live-binary definition. Even treating only the 11 real bundle contracts as
certified leaves 39, not 33, to reach 50.

## How I ranked

Score is a transparent judgement score out of 14:

- Daily workflow frequency: 0-5.
- A person can obtain the credential locally without vendor review: 0-4.
- Free, freemium, self-hosted, test, or local path: 0-3.
- Useful daily action shape: 0-2, where B (read and safe write) scores above
  R (read/inspect only).

M means a cli_surface.json exists on the artifact ref; its x/y is implemented
commands/total commands. It is not a certification claim. U means an existing
bundle has no CLI surface. N means no definition bundle was found. KB means a
current typed quarantine entry blocks correct expansion. All M candidates below
still need rate-limit work: the sampled rate_limits.json files report unknown,
and Crisp has no rate_limits.json at all. Estimates are focused certification
work after a sandbox and credential are available; they include parity and
live-binary proof, not just adding JSON.

Popularity claims are deliberately restrained. GitHub's 180M-plus developer
community and Docker's 20M-plus developer claim are the only usage-scale facts
used here. The rest of the order is labelled judgement: it reflects common
developer/operator daily workflows plus the local-runnability constraint, not
invented market-share statistics.

Action legend: R = inspect/read; W = provider-state write that must have a
sandbox/cleanup pairing; B = both. “Free” means a plausible no-paid-vendor path,
not that every provider account, quota, or organization policy is free forever.

## Ranked top 50

| # | Connector | Score | Bucket and current evidence | Local credential and free path | Daily action | Work to certify / generator verdict |
|---:|---|---:|---|---|---|---|
| 1 | GitHub | 14 | M 383/461; 12 unsafe, 78 not runnable | Personal access token is self-serve, although an organization can restrict it; free personal account. | B | High value but not quick: resolve unsafe policy, close reachable-surface gap, declare rate policy, prove. 5-10d. [GitHub PAT](https://docs.github.com/en/authentication/keeping-your-account-and-data-secure/managing-your-personal-access-tokens) |
| 2 | PostgreSQL | 14 | U; existing native bundle, no CLI surface | Local/self-managed database login; no vendor or paid tier. | R | Map native safe operations and settle local-DB certification/rate policy. 3-5d; no REST generator. |
| 3 | Docker Hub | 14 | M 4/4 | Anonymous reads or self-created scoped personal token; free path. | R | Best first proof job: declare cited rate policy and run real binary against a personal namespace. 1-2d. [token](https://docs.docker.com/security/access-tokens/) / [pull limits](https://docs.docker.com/docker-hub/usage/pulls/) |
| 4 | MySQL | 14 | U; existing native bundle, no CLI surface | Local/self-managed login; no vendor or paid tier. | R | Same native decision as PostgreSQL, then surface + proof. 3-5d; no REST generator. |
| 5 | GitLab | 13 | M 4/1,745 | Self-created personal token; GitLab.com and self-managed have a free path. | R | Great local usability, poor current surface readiness. Narrow safe scope or close large gap before certification. 8-15d. [PAT](https://docs.gitlab.com/user/profile/personal_access_tokens/) |
| 6 | Gitea | 13 | N; no bundle | Personal token on a self-hosted Gitea instance; free/self-hosted. | B | New but attractive: public REST/OpenAPI-style API should be generator-friendly after artifact audit. 3-6d. |
| 7 | Jira | 12 | M 584/617 | User-generated Atlassian API token; Jira Free site is a usable sandbox. | B | Close 33 commands, rate policy, then live proof. 3-6d. [token](https://support.atlassian.com/atlassian-account/docs/manage-api-tokens-for-your-atlassian-account/) |
| 8 | Vercel | 12 | M 25/400 | Personal account access token; Hobby/free path. | B | Value is high, but 375 commands are unavailable. Certify a deliberately reduced safe surface or remediate first. 7-12d. [token](https://vercel.com/kb/guide/how-do-i-use-a-vercel-api-access-token) |
| 9 | Notion | 12 | M 45/49; 3 partial | Internal integration token is self-created, but the user must share pages/databases to it; personal/free workspace path. | B | Close four commands and prove a personal workspace sandbox. 2-4d. [authorization](https://developers.notion.com/guides/get-started/authorization) |
| 10 | Stripe | 12 | M 8/589 | Restricted key and test mode are self-serve; no merchant production account needed for proof. | B | Excellent sandbox, but current surface is far from complete. Do not call it fast. 8-15d. [keys](https://docs.stripe.com/keys) / [testing](https://docs.stripe.com/testing) |
| 11 | Bitbucket | 12 | M 180/331 | App password or access token in a personal/free workspace. | B | Close 151 commands, declare rate policy, prove. 5-8d. |
| 12 | CircleCI | 11 | M 15/111 | Personal API token; free-tier/project path. | B | Good developer daily use, but 96 command gap. 4-7d. |
| 13 | Sentry | 11 | U; existing bundle; KB | Self-created token on self-hosted Sentry or a personal/org account. | R | Not a proving job: hook contract only knows four legacy streams. Engine/hook decision first. [quarantine:192-205](../../../.treehouse/cli-83d592/22/cli/docs/migration/quarantine.json) |
| 14 | Grafana | 11 | U; existing REST bundle | API/service-account key on self-hosted Grafana; free/self-hosted. | R | Map safe reads; likely generator candidate only after an official API artifact audit. 2-4d. |
| 15 | PostHog | 11 | U; existing REST bundle | Personal project API key; self-host/free path. | R | Map safe analytics reads after provider artifact audit. 2-4d. |
| 16 | Airtable | 11 | U; existing REST bundle | Self-created personal token scoped to a base; free workspace path. | B | Map safe base/table operations; generator likely after source-artifact audit. 3-5d. |
| 17 | Linear | 11 | U; existing bundle; KB | Personal API key/access token; free workspace path. | B | High user value, but GraphQL POST-body reads/mutations are a real engine gap. Not a quick target. [quarantine:39-53](../../../.treehouse/cli-83d592/22/cli/docs/migration/quarantine.json) |
| 18 | Asana | 11 | M 82/249 | Self-created personal access token; free personal workspace path. | B | Close 167 commands or certify a properly scoped safe contract. 5-9d. [PAT](https://developers.asana.com/docs/personal-access-token) |
| 19 | Trello | 10 | M 3/264 | Own API key plus user token; free path, but requires creating an integration record. | R | Poor current coverage; reserve for after more-ready work. 5-9d. [authorization](https://developer.atlassian.com/cloud/trello/guides/rest-api/authorization/) / [limits](https://developer.atlassian.com/cloud/trello/guides/rest-api/rate-limits/) |
| 20 | Slack | 10 | U; existing REST bundle | A personal test workspace can install a custom app; company workspaces may require owner approval. | B | Map only a least-privilege sandbox scope after captain decides policy for shared/custom app ownership. [OAuth](https://docs.slack.dev/authentication/installing-with-oauth/) / [workspace approvals](https://slack.com/help/articles/202035138-Add-apps-to-your-Slack-workspace) |
| 21 | ConfigCat | 10 | M 39/123 | Project credential is self-service for an owner; free tier path. | B | Close 84 commands and declare provider rate policy. 3-6d. |
| 22 | Cloudflare | 10 | N; no bundle | User-created scoped API token; generous free plan path. | B | Strong new candidate: REST source appears generator-friendly after current artifact audit. 3-5d. [token](https://developers.cloudflare.com/fundamentals/api/get-started/create-token/) |
| 23 | Auth0 | 10 | M 15/543 | Create a development tenant and M2M application; free developer path, no vendor review. | B | Application setup is self-serve but gap is 528 commands. 8-14d. |
| 24 | n8n | 10 | M 42/176 | API key on a self-hosted n8n instance; free/self-hosted. | B | Certify only expressible reads/non-destructive writes; remaining root-array/query writes are quarantined. 4-7d. [quarantine:73-85](../../../.treehouse/cli-83d592/22/cli/docs/migration/quarantine.json) |
| 25 | Todoist | 10 | U; existing REST bundle | Personal API token; free personal path. | B | Map a compact high-value API subset; likely generator candidate after source audit. 2-4d. |
| 26 | Google Calendar | 10 | M 38/38 | Own OAuth desktop client plus user consent/refresh credential; free personal calendar. | B | Very good proof candidate after deciding whether certification uses per-run personal OAuth clients or an owned shared client. 2-3d. [installed-app OAuth](https://developers.google.com/identity/protocols/oauth2/native-app) |
| 27 | Google Sheets | 10 | N; no bundle | Same self-created OAuth client/user consent path; free personal sheet. | B | New mapping; Google discovery metadata needs an importer/adaptor rather than assuming direct OpenAPI extraction. 3-6d, gated with Calendar/Gmail OAuth policy. |
| 28 | Gmail | 10 | M 63/79; 3 partial | Own OAuth desktop client/user consent; free personal mailbox. | B | Close 16 commands; shared OAuth-client ownership is a captain decision. 3-6d. [installed-app OAuth](https://developers.google.com/identity/protocols/oauth2/native-app) |
| 29 | Shopify | 9 | N; no bundle | Free Partner development store with generated test data, but app/store ownership and cleanup policy are required. | B | Do not treat as easy: GraphQL and developer-store/app ownership need a decision. [development store](https://help.shopify.com/partners/getting-started/development-stores) |
| 30 | Square | 9 | M 4/336 | Developer sandbox application credential; self-serve test path. | R | Good local sandbox, large 332-command gap. 6-10d. |
| 31 | Twilio | 9 | M 163/394 | Account SID/auth token with a self-serve trial; trial constraints must be respected. | B | Close 231 commands and use an isolated test account/cleanup. 5-10d. |
| 32 | Mailgun | 9 | M 4/274 | Self-created API key; use sandbox/trial only unless current plan permits more. | R | Large gap and account-plan validation needed. 5-9d. |
| 33 | Mailchimp | 9 | M 295/295 | Self-created API key; a free account can provide a small test audience. | B | Surface is complete, but proving hundreds of commands safely is not a quick count. 5-8d. |
| 34 | Brevo | 9 | M 23/568 | Self-created API key; free quota/test account path. | B | Large 545-command gap. 7-12d. |
| 35 | Plausible | 9 | M 4/12 | Self-hosted account/token; free/self-hosted. | R | Small safe proof after eight-command gap is closed. 2-3d. |
| 36 | Metabase | 9 | U; existing REST bundle | Self-hosted account/session path; free/open-source. | R | Map safe metadata/query reads and decide write exposure separately. 2-4d. |
| 37 | Datadog | 9 | M 36/235 | API/application keys require an account; assume trial, not a durable free tier. | B | Valuable ops tool, but 199-command gap and non-free risk make it later. 6-10d. |
| 38 | PagerDuty | 8 | M 4/466 | API key from an account; trial/team-admin constraints likely. | R | Not a fast locally-run target; 462-command gap. 7-12d. |
| 39 | Chatwoot | 8 | M 100/148; 1 partial | API access token on a self-hosted instance; free/self-hosted. | B | Strong self-hosted candidate after 48 commands close. 3-5d. |
| 40 | Formbricks | 8 | M 26/66 | API key on a self-hosted instance; free/self-hosted. | B | Compact self-hosted proof after 40-command gap. 2-4d. |
| 41 | DocuSeal | 8 | M 9/9 | API key on a self-hosted instance; free/self-hosted. | B | Best quick self-hosted proof: rate policy, docs parity, live binary, cleanup. 1-2d. |
| 42 | Crisp | 8 | M 21/21 | Identifier/key for a user account; free inbox path. | R | Complete command surface but no rate_limits.json: add authoritative declaration, then proof. 2-3d. |
| 43 | GitBook | 8 | M 298/367 | Personal access token; personal/free documentation path. | B | Good daily documentation workflow but 69 commands remain. 3-6d. |
| 44 | Coda | 8 | M 13/124 | Personal API token; free document path. | B | Map/close 111 commands before proof. 4-7d. |
| 45 | ClickUp API | 8 | M 26/185 | Personal API token; free workspace path. | B | Map/close 159 commands before proof. 4-7d. |
| 46 | Amazon SQS | 8 | M 23/23 | AWS credentials in an isolated account or local emulator; free-tier/local test path. | B | High readiness but niche: declare policy and prove safe queue create/send/read/delete cleanup. 2-3d. |
| 47 | Elasticsearch | 8 | U; existing REST bundle | Self-hosted API key/basic auth; free/self-hosted. | R | Map safe cluster/index reads. Official REST spec is a likely generator input after audit. 2-4d. |
| 48 | MongoDB | 8 | N; no bundle | Local MongoDB or free development cluster credential. | R | Native/non-REST connector build, not generator extraction. 4-7d. |
| 49 | Redis | 8 | N; no bundle | Local Redis ACL/password; free/self-hosted. | R | Native protocol build, not generator extraction. 3-6d. |
| 50 | SearxNG | 7 | M 2/4 | Local self-hosted instance; no credential or optional key. | R | Low operational risk, small proof after two-command gap. 1-2d. |

## Bucket split

| Bucket | Count | Connectors | Meaning |
|---|---:|---|---|
| Already certified | 0 | None | Zero accepted fresh live artifacts were found. Do not count bundle contracts as completed certification. |
| Surface mapped, needs proving | 33 | GitHub, Docker Hub, GitLab, Jira, Vercel, Notion, Stripe, Bitbucket, CircleCI, Asana, Trello, ConfigCat, Auth0, n8n, Google Calendar, Gmail, Square, Twilio, Mailgun, Mailchimp, Brevo, Plausible, Datadog, PagerDuty, Chatwoot, Formbricks, DocuSeal, Crisp, GitBook, Coda, ClickUp API, Amazon SQS, SearxNG | Cheap relative to a new connector because command surfaces exist. They still require availability/unsafe remediation, rate policy, parity docs, and a real binary proof. |
| Not mapped | 17 | PostgreSQL, MySQL, Gitea, Sentry, Grafana, PostHog, Airtable, Linear, Slack, Cloudflare, Todoist, Google Sheets, Shopify, Metabase, Elasticsearch, MongoDB, Redis | Existing native/REST bundles without a CLI surface (11), six genuinely absent definitions, and two typed engine blockers within the existing-bundle set. |

The task mentions roughly 195 known-blocked providers. I could not substantiate
that number in this checkout. docs/migration/quarantine.json currently contains
15 entries total and 13 marked ENGINE_GAP. This report labels only current,
typed blockers as blocked rather than inventing a 195-provider status.

### Not-mapped disposition: generator versus real blocker

| Connector | Starting point | Generator disposition |
|---|---|---|
| PostgreSQL, MySQL | Existing native definitions | No: native database work, not REST extraction. Shared local-live/rate-policy captain call first. |
| Gitea | Absent definition | Yes, likely: self-hosted REST/OpenAPI-style source; audit and pin the provider artifact first. |
| Sentry | Existing definition | Known blocked: StreamHook/Link-pagination contract is insufficient; adding definition JSON alone makes reads fail. |
| Grafana, PostHog, Airtable, Todoist, Metabase, Elasticsearch | Existing REST definitions | Likely, but not yet proven: obtain and pin an official provider artifact, then run the generator/import audit. None is presently typed as a blocker. |
| Linear | Existing definition | Known blocked: GraphQL fixed-body reads/mutations need an engine capability, not generic extraction. |
| Slack | Existing REST definition | Artifact/import may map safe calls, but it does not solve custom-app installation and workspace-approval policy. |
| Cloudflare | Absent definition | Yes, likely: its scoped-token REST API is a good artifact-first generator candidate. |
| Google Sheets | Absent definition | Not directly: Google Discovery metadata needs an importer/adaptor, plus the OAuth ownership choice. |
| Shopify | Absent definition | No generic path today: GraphQL plus development-store/app ownership policy. |
| MongoDB, Redis | Absent definitions | No: native wire-protocol connectors, not REST generator targets. |

## Recommended attack order for the next 33

Rank and delivery order are deliberately different. The table ranks end-user
value; this queue maximizes credible completions without filling the first wave
with vendor-gated or engine-blocked work. A connector enters the count only
after its accepted live artifact is committed.

1. Docker Hub — 4/4, personal token/anonymous read, cited pull policy.
2. DocuSeal — 9/9, self-hosted sandbox, safe write cleanup.
3. Amazon SQS — 23/23, isolated account or local emulator, safe cleanup.
4. Google Calendar — 38/38 after OAuth ownership choice.
5. Crisp — 21/21 after authoritative rate declaration.
6. Mailchimp — 295/295; split the proof plan into safe command families, do not fake a one-command sweep.
7. Notion — close four commands; personal workspace sandbox.
8. Jira — close 33 commands; free site/test project.
9. Chatwoot — close 48; self-hosted.
10. GitBook — close 69; personal document sandbox.
11. Gmail — close 16 after OAuth ownership choice.
12. Asana — close 167; personal workspace.
13. Bitbucket — close 151; personal workspace.
14. Twilio — close 231; isolated trial account.
15. ConfigCat — close 84; project-owned test configuration.
16. n8n — certify only declared expressible actions; leave quarantined root-array/query actions non-implemented.
17. Formbricks — close 40; self-hosted.
18. Plausible — close eight; self-hosted.
19. Coda — close 111.
20. ClickUp API — close 159.
21. CircleCI — close 96.
22. Vercel — decide a meaningful safe subset versus full parity; do not create a ceremonial certificate.
23. Trello — close 261 after a self-owned integration/key setup.
24. SearxNG — close two; low-risk completion.
25. PostgreSQL — captain decision on local-native live proof and rate-policy form, then map.
26. MySQL — same shared native decision; one shared policy change can unblock both.
27. Airtable — map compact safe surface.
28. Todoist — map compact safe surface.
29. Grafana — map safe self-hosted reads.
30. PostHog — map safe self-hosted reads.
31. Metabase — map safe self-hosted reads.
32. Elasticsearch — map safe self-hosted reads.
33. Gitea — new, but an unusually good self-hosted generator-first build.

Do not use Linear or Sentry as count commitments in this 33: both are
high-value, but today require an engine/hook change rather than a certification
proof. Do not use Shopify until its GraphQL and developer-store/app ownership
choices are settled. Cloudflare is a good next reserve candidate once the
initial 33 are on track.

## Past effort that is low value for a daily developer/operator target

Strictly, there are no currently accepted certifications to criticize. If
“currently-certified” means the 11 root certification.json contracts, past
effort was poorly aligned with this particular daily-use goal for:

- Ashby — recruiting-only workflow.
- Freshchat — customer-support-only workflow.
- HubPlanner — staffing/resource-management-only workflow.
- Recurly — specialized billing-operations workflow.
- Xero — specialized accounting workflow.

Amazon SQS and Google Search Console are useful but role-specific rather than
general daily developer/operator connectors. GitHub, Bitbucket, Asana, and
Google Calendar are defensible daily-use choices among the existing contracts.
This is not an argument to delete the specialist connectors; it is a plain
priority correction for a “50 locally runnable daily” certification program.

## Captain's calls before the expensive work starts

These are real choices, not routine implementation details:

1. Google Calendar, Google Sheets, and Gmail OAuth ownership. Choose either
   per-run personal installed-app clients (recommended; no shared client
   secret) or a maintained shared OAuth client with owner, verification,
   rotation, and incident responsibility. Google documents that installed apps
   still need API-console credentials and potentially verification for public
   sensitive scopes.
2. Slack scope. Certify only a personal/sandbox workspace app, or own a shared
   app and accept that workspace app-approval settings can require an owner.
   A broad shared app is not a locally runnable default.
3. Shopify. Decide whether a Partner-owned development store plus a test-only
   app is an acceptable certification sandbox, who owns it, and the cleanup
   rules. Shopify says development stores are free and can use generated test
   data, but app/store access has ownership boundaries.
4. GitHub unsafe surface. The 12 unsafe commands must be removed from the
   certification contract or redesigned into reviewed narrow operations. Zero
   unsafe/disallowed and a certificate cannot both be true while they remain
   in scope.
5. Native database policy. Decide whether a locally started PostgreSQL/MySQL
   instance is valid “live” proof for a native connector and standardize
   rate_limits as not-applicable with a documented rationale. The existing
   PostgreSQL convention explicitly says its parity proof is fixture mode, so
   live local certification needs a deliberate policy rather than implication
   ([conventions:1279-1280](../../../.treehouse/cli-83d592/22/cli/docs/migration/conventions.md)).

## Evidence trail and repeatable commands

Repository evidence:

- Requested suffix counts and file paths:

      git ls-tree -r --name-only origin/main internal/connectors/defs/ | grep -c 'certification\.json$'
      git ls-tree -r --name-only origin/fm/cli-mass-artifact-materialize-r1 internal/connectors/defs/ | grep -c 'certification\.json$'
      git ls-tree -r --name-only origin/main internal/connectors/defs/ | rg 'certification\.json$'

- Surface counts:

      git ls-tree -r --name-only origin/main internal/connectors/defs | rg '/cli_surface\.json$' | wc -l
      git ls-tree -r --name-only origin/fm/cli-mass-artifact-materialize-r1 internal/connectors/defs | rg '/cli_surface\.json$' | wc -l

- Root source of the runtime availability guard:
  [runner_test.go:2547-2563](../../../.treehouse/cli-83d592/22/cli/internal/connectors/commandrunner/runner_test.go).
  It sweeps commands through the real preflight; availability metadata alone
  is not enough.

- Typed blockers:
  [Linear quarantine:39-53](../../../.treehouse/cli-83d592/22/cli/docs/migration/quarantine.json),
  [n8n quarantine:73-85](../../../.treehouse/cli-83d592/22/cli/docs/migration/quarantine.json),
  and
  [Sentry quarantine:192-205](../../../.treehouse/cli-83d592/22/cli/docs/migration/quarantine.json).

External sources used only for defensible reach and credential claims:

- [GitHub Octoverse 2025](https://github.blog/news-insights/octoverse/octoverse-a-new-developer-joins-github-every-second-as-ai-leads-typescript-to-1/)
  reports 180M-plus developers on GitHub; [Docker's 2024 report](https://www.docker.com/press-release/unveils-2024-state-of-application-development-report/)
  reports more than 20M developers. Those facts support the top-two
  ecosystem-scale claims, not every rank.
- [Stack Overflow Developer Survey 2025: Technology](https://survey.stackoverflow.co/2025/technology/)
  is a useful directional cross-check for developer infrastructure choices,
  not a source for any uncited percentage in this report.
- Provider credential/rate documents are linked in the relevant table rows.

## Recommended success metric

Track three independent numbers in the captain dashboard:

1. Root bundle certification contracts.
2. Mapped surfaces with no unsafe/disallowed and a declared rate policy.
3. Fresh accepted live certification artifacts.

Only number 3 is “certified.” This prevents a fast-looking file count from
masking unproven commands or a vendor-gated sandbox.
