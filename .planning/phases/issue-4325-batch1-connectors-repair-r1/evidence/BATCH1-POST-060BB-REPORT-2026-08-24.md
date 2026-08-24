# Batch 1 post-`060bb7864` remeasurement

Date: 2026-08-24

Branch: `fm/cli-batch1-repair-r1` at `b298686df` (merge of `origin/main`
`060bb7864`, #4343)
Scope: Docker Hub, Notion, Stripe, Bitbucket, GitLab, CircleCI, Sentry,
Vercel, Asana, and Jira.

## Method

The retained-source identities and provider-document checks are recorded in
[`SOURCE-RETENTION-REPORT-2026-08-24.md`](SOURCE-RETENTION-REPORT-2026-08-24.md)
and [`REPIN-REPORT-2026-08-24.md`](REPIN-REPORT-2026-08-24.md). No provider
source was classified unavailable from a command-line HTTP failure. In
particular, Docker Hub's retained source is the byte-identical 148,322-byte
artifact already documented there.

This measurement ran sequentially with `GOMAXPROCS=2`:

```text
go run ./cmd/connectorgen source-import <connector> --check
go run ./cmd/connectorgen validate internal/connectors/defs/<connector> --json
go build -o /tmp/cli-batch1-inbox013-20260824/pm ./cmd/pm
```

For every declared `availability: implemented` command in each connector that
has `cli_surface.json`, an isolated `pm init --root <temporary-project>` was
created and the built binary invoked with no credential. A command is counted
as runtime-reachable only when its complete output is exactly
`error: missing --credential`. Commands were run one at a time; no test suites
or command sweeps ran in parallel. Vercel has no CLI surface, so the built
binary was instead exercised as `pm vercel`, which returned exit 2 and
`error: unknown command "vercel"`.

## Results

| Connector | Source-import check | Validator | Built-binary credential-boundary evidence | Honest current state |
| --- | --- | --- | --- | --- |
| Docker Hub | blocked: `components.schemas["team_repo"]` resolves dangling `#/components/responses/team_repo` | 7 findings | 24/45 reach boundary; 19 direct reads lack their `rest_read` runtime endpoint-ledger entries and 2 SCIM writes lack closed object body schemas | Importer-blocked on the byte-identical provider document; also has present runtime metadata defects. No local workaround applied. |
| Notion | blocked: `/v1/blocks/{block_id}/children` PATCH exceeds the resolved descriptor byte limit while retaining request media | 1 finding: canonical source descriptor missing | 45/45 | Runtime-reachable, but cannot pass until the importer retains this descriptor. |
| Stripe | blocked: `/v1/account` GET 200 response exceeds schema-depth limit | 1 finding: canonical source descriptor missing | 8/8 | Runtime-reachable, but cannot pass until the importer retains this descriptor. |
| Bitbucket | pass: 297 operations, 0 inbound events | 0 findings | 50/50 | Local source/validation/runtime evidence complete; awaits independent gate. |
| GitLab | pass: 1,752 operations, 0 inbound events | 0 findings | 4/4 | Local source/validation/runtime evidence complete; awaits independent gate. |
| CircleCI | pass: 111 operations, 0 inbound events | 0 findings | 43/43 | Local source/validation/runtime evidence complete; awaits independent gate. |
| Sentry | pass: 223 operations, 0 inbound events | 32 source-projection findings: actions are absent for source mutations | 4/4 | Runtime-reachable surface exists; source-action disposition coverage still blocks certification. |
| Vercel | pass: 400 operations, 0 inbound events | 112 source-projection findings: 104 missing actions, 4 missing request-field unions, 4 unresolved source-bound gaps | no `cli_surface.json`; `pm vercel` is `unknown command` | No executable CLI surface; source-action coverage also blocks certification. |
| Asana | blocked: 25 source operations have no complete executable action | 90 source-projection findings: 25 missing actions, 65 unresolved source-bound gaps | 82/82 | Runtime-reachable surface exists; per-operation source-cited mutation disposition is still required. |
| Jira | blocked: 14 source operations have no complete executable action | 319 findings: 16 missing actions, 210 request-field gaps, 86 unresolved gaps, 1 descriptor-provenance drift, and 6 binary-response metadata findings | 581/584; the 3 `universal-avatar` image commands are blocked because their operations have no response `content_types` (the validator also reports their missing success statuses) | Runtime surface is mostly reachable, but source disposition and binary response metadata still block certification. |

The measured split is therefore **5/10 source-import passes** and **3/10
validator passes**. The local proof is not an independent-GO declaration:
Bitbucket, GitLab, and CircleCI are the only three that clear all local checks
in this run, and all ten still require the independent gate to declare GO.

## Foundation boundary

The Docker Hub input is not source drift: its retained 148,322-byte provider
artifact contains the unresolved `team_repo` reference. The source importer
currently refuses the whole document during grammar preflight. This must be
fixed in the importer/foundation path by retaining the source-traced defect;
this connector branch does not mask or rewrite the provider reference.

Likewise, Notion and Stripe need importer tolerance that retains a source-traced
gap instead of losing the descriptor; Asana, Sentry, Vercel, and Jira need
source-cited per-operation action/disposition coverage. These are recorded as
facts from the commands above, not converted into declaration-count reductions.
