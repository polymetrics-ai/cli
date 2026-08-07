# github — documented-operation parity (sweep slice)

Program: `cli-top50-fixed-schema-sweep-r1` · branch `fm/cli-top50-sweep-continue-r1`.
**Landing order #1 under the captain's largest-first reversal — this is the frog.**

> **Banner (standing):** if this connector surprises you, **STOP and record it here** rather than
> forcing it into the batch shape.

## Sliced delivery — this connector will not fit one context

Per the captain's largest-first order: commit and push **each slice**, never one giant uncommitted
change. Each slice must leave the tree green on shared gates; github's own red test staying red
between slices is expected and honest.

| # | Slice | State |
| ---: | --- | --- |
| 1 | red test + `DERIVED-OPERATIONS.json` + this plan | ✅ **done, committed, pushed** |
| 2 | `api_surface.json` → 1220 REST + 4 GraphQL rows, all dispositioned | ☐ |
| 3 | streams / writes / operations | ☐ |
| 4 | `cli_surface.json` + full binary reachability sweep | ☐ |
| 5 | `SUMMARY.md` + `VERIFICATION.md` | ☐ |

## Artifact

| | |
| --- | --- |
| URL | GitHub's own `rest-api-description` → `descriptions/api.github.com/api.github.com.json` |
| Kind | OpenAPI **3.0.3**, `info.version` **1.1.4** |
| Retrieved | 2026-08-07, HTTP 200, **12,920,264 bytes** — byte-identical to the sweep derivation |

**Derivation reproduced, not trusted:** 808 paths → **1220** method entries, all unique, method split
`GET 636 · POST 193 · PUT 134 · DELETE 187 · PATCH 70`, 37 `deprecated: true`. Reconciles exactly
with the ledger's 1220. Full inventory committed as `DERIVED-OPERATIONS.json`.

## The gap: the bundle enumerated ONE scope out of many

The shipped `api_surface.json` has **509 rows** — and its own `scope` says why: it enumerated only
`/repos/{owner}/{repo}/…`, declaring "org-level, user-level, enterprise-level, and Enterprise
Cloud/Server admin surfaces remain out of scope". Against the documented surface that is **501 real
operations out of 1220**; **719 are missing**, and they are missing by whole scope, not scattered.

The red test spot-pins one row per absent scope — `GET /orgs/{org}`, `GET /user`,
`GET /enterprises/{enterprise}/copilot/billing/seats`, `GET /app/hook/config` — so a partial
re-expansion cannot pass by filling the repository surface again.

## Hazards, and the judgement each forces

### 1. ⚠️ 270 webhook events hide under `x-webhooks`, not `webhooks`

The spec declares `openapi: 3.0.3`, which has **no** native top-level `webhooks` object. GitHub puts
all 270 event payloads under the **vendor extension** `x-webhooks`, a sibling of `paths`. **Any
inventory checking only a literal `webhooks` key records zero events for the largest connector in
the sweep.** Confirmed here: `[k for k in spec if 'webhook' in k]` → `['x-webhooks']`.

Events stay **excluded** by policy. The **28 webhook MANAGEMENT operations** (`/app/hook/*`,
`/orgs/{org}/hooks/*`, `/repos/{owner}/{repo}/hooks/*`) are ordinary `paths` entries and **are**
part of the 1220.

### 2. Four synthetic path rows that are not endpoints

The bundle carries `PATCH /repos/{owner}/{repo}/issues/{issue_number} (close)` and three siblings —
a **behaviour variant encoded into the path**. No such path is documented. This is the same defect
class as help-scout's `?async=true` and lever-hiring's `?include=` rows: model the variant as a flag
or a `duplicate` disposition, never as a second path. The red test rejects any REST path containing
a space, `?` or `*`.

### 3. The four GraphQL rows are counted separately, and the gap is named

`ListProjects`, `ListProjectItems`, `ListDiscussions`, `ViewDiscussion` back real shipped streams.
They are **not** among the 1220 REST operations and are asserted as their own count, never folded in.

**GitHub's GraphQL schema exposes far more than four top-level fields, and this sweep does not
enumerate it.** That is a **named scope gap**, recorded so it cannot be mistaken for completeness —
not a claim that github's GraphQL surface is covered.

### 4. Read vs write — five POSTs are semantically reads

`POST /markdown`, `POST /markdown/raw`, `POST /applications/{client_id}/token` (check-token),
`POST /orgs/{org}/attestations/bulk-list`, `POST /users/{username}/attestations/bulk-list`.
Each is a query that uses POST for body size, not a mutation. They must be classified as reads.

### 5. 37 deprecated operations are counted, not excluded

Concentrated in the legacy non-org-scoped `/teams/{team_id}/*` surface and the shut-down
`/repos/{owner}/{repo}/import` Importer. Deprecated operations count; they get `model: deprecated`.

### 6. The pre-existing test will need its numbers updated — that is not weakening it

`TestGitHubAPISurfaceOperationLedgerMetrics` pins 509 rows / 440 covered / 69 blocked with
`t.Fatalf`. Once the surface reaches 1224 rows it must fail. **Update its expected counts and keep
every structural assertion** (`blocked_by_default`, reason present, `source_url`-or-`notes`,
`duplicate_of` on duplicates). Changing an expected value because the underlying truth changed is
not the same as relaxing a check — and no check may be removed.

### 7. Binary detection — to be settled in slice 3

Candidates: the archive/tarball/zipball redirect endpoints
(`GET /repos/{owner}/{repo}/{archive_format}/{ref}`) and raw content reads. These return redirects or
raw bytes rather than JSON, and must be adjudicated individually rather than pattern-matched.

## Work order per slice

Standard bar: `connectorgen validate` · **whole** `cmd/connectorgen` package (finding F5 — github has
**two** surface tests) · `TestEveryImplementedCommandPassesRuntimePreflight` ·
`connectorgen surface-sync --check` · **run the binary over every generated command** · no
hand-authored paging flags · every blocked row carrying `Named dependency:`.
