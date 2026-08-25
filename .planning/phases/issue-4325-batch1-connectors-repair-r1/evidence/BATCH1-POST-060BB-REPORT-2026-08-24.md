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

## 2026-08-25 — declaration-admission inventory and certification design

### Captain rule

Provider-backed coverage is admitted before runtime capability. Every provider
operation—including destructive writes, reverse-ETL-shaped actions, direct
operations, and binary operations—must remain in the JSON ledger and be
discoverable. Runtime safety remains strict: only an `implemented` command
with closed metadata and successful real preflight is runnable. Deferred rows
must name their exact source and missing foundation; they are never omitted or
promoted.

### Exact JSON artifacts by connector lane

All ten lanes require these connector-owned artifacts:

- `sources/<connector>-operation-source-lock.json` (pinned source identity),
  `sources/<connector>-operation-crosswalk.json`,
  `sources/<connector>-declaration-disposition.json`, and
  `sources/<connector>-retained-artifacts.json`;
- `api_surface.json` (discoverable endpoint ledger), `operations.json`,
  `writes.json`, `cli_surface.json` containing the declared state of every
  exposed source operation,
  `certification-sweep.json`, `metadata.json`, `spec.json`, `streams.json`,
  and `sync_transport.json`.

`sources/<connector>-operation-descriptor.json` is additionally required for
the importable lanes **Bitbucket, GitLab, CircleCI, Sentry, Vercel, Asana, and
Jira**. Docker Hub, Notion, and Stripe require it as the output of the
source-dialect-tolerant importer; its current absence is an explicit
source-bound importer gap, not evidence that their source operations vanish.
`sources/<connector>-mutation-dispositions.json` is required where a source
mutation is non-executable: present today for Bitbucket, GitLab, Sentry, Asana,
and Jira. Docker Hub additionally owns its reverse-ETL action audit. Vercel
must add `cli_surface.json` with explicit deferred states for every applicable
source operation; the absence of a runnable command cannot remove its 400
source rows or make its command namespace undiscoverable.

### Counts (measured from the retained source and declaration artifacts)

The ten retained provider locks contain **4,341** exact REST method/path
operations. The previously re-derived operational split was **767 runnable,
1,666 immediately declarable, and 1,908 source-bound blocked**. Under the
captain rule, all 1,908 blocked rows remain required declarations with explicit
deferred/foundation state; the split is an execution plan, never a reason to
shrink the source denominator.

| Lane | Locked source operations | Existing source-ledger rows | This-branch converted and credential-proven | Immediately declarable / source-bound deferred |
| --- | ---: | ---: | ---: | --- |
| Docker Hub | 54 | 54 | 0 | 0 / 50 |
| Notion | 49 | 49 | 0 | 0 / 7 |
| Stripe | 589 | 589 | 0 | 0 / 581 |
| Bitbucket | 297 | 331 (34 stale rows to refresh) | 0 | 136 / 111 |
| GitLab | 1,752 | 1,755 (3 stale rows; base-path-normalized) | 0 | 917 / 835 |
| CircleCI | 111 | 111 | 0 (43 existing commands re-proven) | 75 / 20 |
| Sentry | 223 | 223 | 32 | 144 / 76 |
| Vercel | 400 | 400 | 0 | 261 / 139 |
| Asana | 249 | 249 | 0 | 130 / 37 |
| Jira | 617 | 617 | 0 | 3 / 52 |
| **Total** | **4,341** | **4,378** | **32** | **1,666 / 1,908** |

The ledger total is 37 larger than the current source denominator only because
Bitbucket and GitLab retain the 34 and 3 known stale endpoint rows. Every
current Bitbucket descriptor endpoint is present in its ledger; GitLab's
descriptor uses the documented `/api/v4` base-path normalization and must be
compared after that normalization. No source lock or provider bytes were
rewritten for this report.

### Current restrictions and their correct scope

| Tool/restriction | Current behavior | Required boundary |
| --- | --- | --- |
| Source projection (`sourceprojection.go:187-290`) | Source import errors when a mutation lacks a complete executable action, unless its narrow non-executable mutation disposition applies. | It must preserve a source-cited deferred declaration for every non-runnable operation, not demand an invented action. |
| Source executable coverage validation (`sourceprojection.go:2430-2524`) | Rejects reads without reachable operations and mutations without a complete action/implemented command unless a blocking gap is already materialized. | This is a runtime-suitability result, not the source-admission denominator. |
| Batch gate (`batch.go:939-956`) | Drops a bundle with zero `implemented` commands. | Keep it as runtime certification; do not use it as provider-declaration admission. |
| Surface reconcile (`surfacereconcile.go:18-22`, `:276-315`) | Grants `covered_by.direct_read` only after real preflight and otherwise records a blocked reason. | Keep it runtime-only; its output cannot delete a declared source row. |
| Surface sync | Synchronizes metadata for existing operation-owned commands; it does not invent a command. | No change: it must not be asked to prove or fabricate deferred coverage. |
| Certification sweep (`certificationsweep.go:810-814`) | Marks a non-implemented command `not_applicable`. | Keep this accounting, but do not interpret N/A as absent source coverage. |
| Operation evidence (`operationevidence.go:743-752`, `:1200`) | Records CLI/runtime gaps and excludes non-enabled rows from the fixed-100 eligible set. | Retain these gaps as runtime evidence; the artifact must still retain the source row and deferred foundation. |

### Smallest safe certification change and tests

Add a source-declaration-completeness certificate, separate from runtime
preflight. It joins the lock/descriptor identity set, crosswalk,
declaration-disposition ledger, and API surface; each source operation must
have exactly one canonical declaration or an exact deferred foundation record.
It must reject omitted, duplicate, stale, citation-free, class-changing,
destructive-metadata-free, or falsely implemented rows. It must not alter
`commandrunner.Preflight`, credential checks, `surface-sync`, or the
implemented-command sweep.

The test matrix is: (1) runnable REST read passes admission and real preflight;
(2) deferred mutation/delete and binary rows pass admission only with exact
source and foundation facts, while execution remains blocked; (3) an importer
dialect failure retains an operation-level source gap instead of dropping a
descriptor; (4) missing/duplicate/stale/base-path-mismatched rows fail with
the source identity; (5) a deferred row falsely marked implemented fails both
admission and the existing runtime sweep; and (6) a complete zero-runnable
connector is admitted as deferred coverage while batch/runtime certification
reports it as non-runnable. This is design and test planning only: no
certification code has started.
