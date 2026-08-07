# github parity — TDD ledger

Red-green-refactor evidence for `.planning/phases/github-parity-sweep-r1`.
github is delivered in **five slices** because 1220 documented operations do not fit one context.
This ledger grows one section per slice; nothing already recorded is rewritten.

## 1. Baseline before any production edit

| Check | Result |
| --- | --- |
| Artifact | GitHub's own `rest-api-description` → `descriptions/api.github.com/api.github.com.json` |
| Artifact bytes | **12,920,264** — byte-identical to the sweep derivation, so the extraction is reproduced rather than trusted |
| Spec | `openapi: 3.0.3`, `info.version: 1.1.4` |
| Documented operations | **1220** over 808 paths, all unique (GET 636 · POST 193 · PUT 134 · DELETE 187 · PATCH 70), 37 `deprecated: true` |
| Webhook events | **270**, under the vendor extension `x-webhooks` — excluded from the operation surface by policy |
| `api_surface.json` rows | **509** = 505 REST + 4 GRAPHQL; of the 505, **501 are real** and **4 are synthetic path rows** |
| Missing documented operations | **719**, missing by whole *scope* (org / user / enterprise / app / admin), not scattered |
| Pre-existing github tests (finding F5 check) | **TWO** — `github_api_surface_test.go` and `github_documented_surface_test.go` |

## 2. RED — committed failing, before production edits (slice 1)

`cmd/connectorgen/github_documented_surface_test.go` against the real bundle, commit `6848cbb2d`:

```
=== RUN   TestGitHubDocumentedRESTSurfaceIsComplete
    github_documented_surface_test.go:138: POST /app/installations/{installation_id}/access_tokens: blocked row must carry a 'Named dependency:' marker
    github_documented_surface_test.go:156: 4 synthetic path row(s) are not documented endpoints: PATCH /repos/{owner}/{repo}/issues/{issue_number} (close), PATCH /repos/{owner}/{repo}/issues/{issue_number} (reopen), PATCH /repos/{owner}/{repo}/pulls/{pull_number} (close), PATCH /repos/{owner}/{repo}/pulls/{pull_number} (reopen)
    github_documented_surface_test.go:167: REST endpoints = 505, want 1220 documented operations
    github_documented_surface_test.go:176: restByMethod = map[DELETE:72 GET:259 PATCH:36 POST:91 PUT:47], want map[DELETE:187 GET:636 PATCH:70 POST:193 PUT:134]
    github_documented_surface_test.go:190: expected "GET /orgs/{org}" — the shipped bundle enumerated only /repos/{owner}/{repo}/…
    github_documented_surface_test.go:190: expected "GET /user" — the shipped bundle enumerated only /repos/{owner}/{repo}/…
    github_documented_surface_test.go:190: expected "GET /enterprises/{enterprise}/copilot/billing/seats" — the shipped bundle enumerated only /repos/{owner}/{repo}/…
    github_documented_surface_test.go:190: expected "GET /app/hook/config" — the shipped bundle enumerated only /repos/{owner}/{repo}/…
    github_documented_surface_test.go:190: expected "POST /markdown" — the shipped bundle enumerated only /repos/{owner}/{repo}/…
    github_documented_surface_test.go:190: expected "GET /teams/{team_id}" — the shipped bundle enumerated only /repos/{owner}/{repo}/…
--- FAIL: TestGitHubDocumentedRESTSurfaceIsComplete (0.00s)
FAIL	polymetrics.ai/cmd/connectorgen	0.525s
```

Slice 1 was committed at this red state. Nothing was authored before it was observed.

---

## Slice 2 — the documented GET surface (636 of 1220)

### 2a. RED still red, and correctly narrower

After slice 2 the same test still fails, and the GET assertions are gone from the failure:

```
    github_documented_surface_test.go:168: REST endpoints = 882, want 1220 documented operations
    github_documented_surface_test.go:177: restByMethod = map[DELETE:72 GET:636 PATCH:36 POST:91 PUT:47], want map[DELETE:187 GET:636 PATCH:70 POST:193 PUT:134]
    github_documented_surface_test.go:157: 4 synthetic path row(s) are not documented endpoints: …
    github_documented_surface_test.go:198: expected "POST /markdown" — …
```

`GET: 636` now equals the derived truth on both sides of the comparison. The three scope spot-pins
that name GETs (`GET /orgs/{org}`, `GET /user`, `GET /app/hook/config`, `GET /teams/{team_id}`) pass.
Everything still red is non-GET work, which is slice 3.

### 2b. A red-test assertion that could never have passed

The spot-pin `GET /enterprises/{enterprise}/copilot/billing/seats` **is not in the artifact at all.**
Copilot billing is org-scoped: the artifact documents `GET /orgs/{org}/copilot/billing/seats` and
publishes **no** `/enterprises/{enterprise}/copilot/billing/…` path. With the GET surface complete at
636/636 that pin can never pass, so it is not a gap in the bundle — it is a wrong assertion.

It was **replaced, not deleted**, by a genuine enterprise-scope GET from the same artifact
(`GET /enterprises/{enterprise}/code-security/configurations`), with the substitution and its reason
written into the test. The scope the pin existed to guard is still guarded; the pin count is
unchanged. **This is the only assertion touched, and it was widened by nothing.**

### 2c. `TestGitHubAPISurfaceOperationLedgerMetrics` — counts updated, structure untouched

Its `t.Fatalf` snapshot moved 509→886 rows, 440→806 covered, 69→80 blocked, GET 259→636. Every
structural assertion is byte-for-byte unchanged: `blocked_by_default`, reason-present,
source_url-or-notes, `duplicate_of`-on-duplicates, and all six per-method maps still compared with
`reflect.DeepEqual`. Changing an expected value because the truth changed is not relaxing a check.

### 2d. GREEN evidence for slice 2

| Gate | Result |
| --- | --- |
| `connectorgen validate` | **551 connectors, 0 findings** |
| `connectorgen surface-sync --check` | 551 scanned, **0 filled / 0 corrected** |
| `TestEveryImplementedCommandPassesRuntimePreflight` | **PASS** (`ok … 15.256s`) |
| `TestGitHubAPISurfaceOperationLedgerMetrics` | **PASS** |
| `TestGitHubDocumentedRESTSurfaceIsComplete` | **still red on non-GET rows — expected, honest** |
| Commands reachable by running the binary | **749/749**, routing asserted on the NAME line |
| Per-command paging flags authored | **zero** — the generator raises `SystemExit` rather than emit one |
| Endpoint-ledger delta | **none** — plain direct reads are not operation-backed, so `operation_endpoint_ledger.json` is untouched |

### 2e. The reachability probe was wrong the first time, and that is a finding

The first probe checked only the exit code. **`pm github <nonsense> --help` exits 0** — a namespace
miss renders the `pm github` group help and succeeds, which is the documented namespace behaviour and
exactly what makes exit status worthless as proof. The probe now asserts the rendered `NAME` line
reads `pm github <path> - …`; against the fixed probe an unroutable path is reported. Every one of
the 749 implemented commands was re-swept under the fixed probe.

### 2f. Four judgements, none of them mechanical

1. **read vs write** — every documented GET is a read; no GET was modelled as a write and no non-GET
   was touched by this slice.
2. **stream vs direct read** — the 364 new GET commands are plain `direct_read`, not streams. A
   stream needs a hand-authored record schema, primary key and fixture; inventing 364 of those would
   invent data contracts GitHub never published, and greenhouse finding 21 already records that
   parity holds either way. Plain direct reads also add **zero** entries to the shared
   `operation_endpoint_ledger.json`, keeping this slice's blast radius inside github. Precedent:
   help-scout's 49 direct reads and gmail.
3. **binary detection** — read out of the artifact, never guessed: an operation is binary iff its
   documented success response is a **302 redirect**, and its operationId verb is not `check…`
   (`orgs/check-membership-for-user` documents a 302 as a *status*, not a download). That yields
   exactly 2 new binary downloads — the org and user migration archives — alongside the 8 the bundle
   already shipped.
4. **named-dependency blocking** — 11 GETs are blocked, each naming the component that refuses it:
   9 boolean `204 No Content` status checks and `/zen` + `/octocat`, which return `text/plain` and
   `application/octocat-stream`. `engine.decodeDirectReadBody` json-decodes every direct-read body
   and `commandrunner.supportedDirectReadOutputPolicies` declares no status-only or text policy, so
   shipping them as `implemented` would ship commands that fail on every invocation. Both halves of
   that dependency are grep-checkable.

### 2g. Deliberate scope limits, stated rather than discovered later

- **Flags cover path variables and REQUIRED query parameters only.** Optional query filters are not
  operations, and authoring several thousand of them would bury the parity signal. `{owner}`/`{repo}`
  stay config-supplied, matching the 162 direct reads the bundle already shipped.
- **`search prs` stays `planned`.** gh models it as a preset over the *same* endpoint as
  `search issues` (`GET /search/issues`). Its note now names the covering command instead of claiming
  a missing capability. Four sibling `search *` rows that were also `planned` were **promoted in
  place** rather than duplicated, so no endpoint gained a second command name.
- **GitHub's GraphQL schema is still enumerated at 4 fixed operations.** That is a named scope gap
  carried from slice 1, not a completeness claim.

---

## Slice 3a — `covered_by.writes`, the change the synthetic rows actually needed

The four `(close)`/`(reopen)` rows could not simply be deleted: `covered_by.write` is a single
string, and validate requires **every** declared write action to be referenced by some row. Six
actions sat on two endpoints.

Deleting the extra actions looked like the cheaper fix and is wrong.
`internal/connectors/hooks/github/hooks.go` assembles the close/reopen bodies by switching on the
action **name**, and `internal/connectors/certify/pairing_test.go` binds `create_issue` to
`close_issue` as its cleanup pair. Removing them would delete shipped, hook-backed,
certification-bound behaviour inside a parity commit — the same mistake greenhouse finding 21
declined to make.

### RED

```
cmd/connectorgen/validate_surface_test.go:66:40: unknown field Writes in struct literal of type engine.SurfaceCoverage
cmd/connectorgen/validate_surface_test.go:92:40: unknown field Writes in struct literal of type engine.SurfaceCoverage
FAIL	polymetrics.ai/cmd/connectorgen [build failed]
```

Two tests, written before the field existed: one asserting an endpoint may back several write
actions, one asserting a plural entry naming an **undeclared** action is still a finding. Widening
the shape must not widen what goes unchecked.

### GREEN

| Gate | Result |
| --- | --- |
| Both new tests | **PASS** |
| `internal/connectors/engine` | **ok** 5.361s |
| `internal/connectors/conformance` | **ok** 15.851s |
| `internal/connectors/commandrunner` | **ok** 8.335s |
| `connectorgen validate` | 551 connectors, **0 findings** |
| `connectorgen surface-sync --check` | 0 filled / 0 corrected |
| `pm github issue close` / `pm github pr reopen` | still route |

`SurfaceCoverage.WriteTargets()` returns singular and plural together, so no reader has to remember
that both spellings exist; connectorgen validate, `conformance/static`, `batch` and
`batch_materialize` all go through it. github's api_surface drops 886 → **882 real rows**.

**Not done here, deliberately:** notion ships the same defect (`PATCH /v1/comments/{comment_id}
(body=markdown)` and `POST /v1/pages/{page_id}/move (parent=data_source_id)`) and is already merged
as #3894. It can now adopt `covered_by.writes`, but converting a merged connector is not this
connector's parity commit. **Recorded, not folded in.**

---

## Slice 3b — the documented mutation surface (584 of 1220), and the surface closes

342 documented mutations were missing (POST 102 · PUT 87 · DELETE 115 · PATCH 38). All 342 are now
enumerated: **322 new write actions**, **2 POST-shaped reads**, **18 blocked**.

### 3b.1 GREEN — the red test from slice 1 finally passes

```
ok  	polymetrics.ai/cmd/connectorgen	11.290s
```

`TestGitHubDocumentedRESTSurfaceIsComplete` and `TestGitHubAPISurfaceOperationLedgerMetrics` both
pass. The surface is **1224 rows = 1220 REST + 4 GraphQL**, reconciling exactly with the artifact:
`GET 636 · POST 193 · PUT 134 · DELETE 187 · PATCH 70`. 1126 covered · 98 blocked · 0 excluded ·
0 blank · 0 duplicates · no query-string rows · no wildcard rows.

### 3b.2 read vs write, decided per operation and not per method

Five POSTs are semantically reads. They are **not** all treated the same way, because the reason each
one is a read differs:

| Operation | Disposition | Why |
| --- | --- | --- |
| `POST /orgs/{org}/attestations/bulk-list` | **implemented read** | returns JSON; uses a body only because the subject list is too long for a query string |
| `POST /users/{username}/attestations/bulk-list` | **implemented read** | same |
| `POST /markdown` | **blocked** | returns `text/html` |
| `POST /markdown/raw` | **blocked** | returns `text/html`, and its request body is `text/plain` |
| `POST /applications/{client_id}/token` | **blocked** | its request body carries an OAuth `access_token` |

The two bulk-lists are the only new **operation-backed** direct reads in the whole github delivery,
and they are the reason the shared `operation_endpoint_ledger.json` moves at all: **162 → 164**,
under the `github` key and nowhere else. Verified by diffing the ledger object-by-object against
`HEAD`: exactly one connector's list changed.

`POST /applications/{client_id}/token` was recorded in slice 1 as "classified as read". That
classification is **corrected here**: a read whose input is a live credential cannot be a command,
because AGENTS.md forbids taking credentials as command arguments. Its three siblings
(`PATCH`/`DELETE .../token`, `DELETE .../grant`) are blocked for the same reason.

### 3b.3 The 18 blocked mutations, each naming a checkable blocker

| Count | Blocker |
| ---: | --- |
| 12 | request body rooted at `oneOf`/`anyOf`. AGENTS.md: "A declarative reverse-ETL `record_schema` rooted at `oneOf` or `anyOf` is not one executable command contract… model each reachable arm as a separate named action, or leave it non-implemented." This takes the second option and names it. |
| 4 | request body carries an OAuth `access_token` |
| 2 | documented success response is `text/html` |

Not counted above: the pre-existing `POST /app/installations/{installation_id}/access_tokens`
(github_app AuthHook), which also gained a named dependency in slice 2.

**One of the 12 is worth calling out as tractable follow-up:** `POST`/`DELETE /user/emails` are
`oneOf` only because GitHub accepts three spellings of the same input — `{emails: [...]}`, a bare
array, and a bare string. One arm is the documented canonical shape and the others are conveniences,
so a single action on the object arm would cover the operation. `POST .../projectsV2/{n}/items`
(`{id}` vs `{owner, repo, number}`) is a genuine two-action split. The
`secret-scanning/custom-patterns` `anyOf` is an at-least-one-of constraint rather than alternative
contracts. **Recorded, not done**, because splitting arms is per-operation judgement and belongs in
its own change with its own red test.

### 3b.4 One out-of-base operation

`POST /repos/{owner}/{repo}/releases/{release_id}/assets` is the **only** operation in the artifact
carrying an operation-level `servers` override (`https://uploads.github.com`). Blocked, naming
`engine.normalizeDirectReadPathForBaseURL` and the single configured `base_url` — the same
out-of-base rule that produced chatwoot's 47 blocks and help-scout's 5.

### 3b.5 31 commands are `partial`, and that is the rule working

`checkCLISurfaceWriteFlags` requires a flag for **every** required record field, recursing into
nested required objects. Where a required field has no scalar leaf — `gists create` requires `files`,
a map of file objects — no flag can carry it from a command line. Those commands are `partial`, not
`implemented`, with a note naming the field. The write action is still declared and still reachable
through reverse ETL with a record payload; what is honest is the availability label.

### 3b.6 GREEN evidence for slice 3

| Gate | Result |
| --- | --- |
| **Whole** `cmd/connectorgen` package (finding F5 — github has two surface tests) | **ok** 11.290s |
| `connectorgen validate` | 551 connectors, **0 findings** |
| `connectorgen surface-sync` | ledger synced; `--check` clean afterwards |
| Endpoint-ledger delta | **github only**, 162 → 164 |
| `internal/connectors/commandrunner` (incl. the runtime preflight sweep) | **ok** 8.461s |
| Blocked rows missing a `Named dependency:` marker | **0 of 98** |
| Per-command paging flags authored | **zero** |

### 3b.7 Two schema rejections the gates caught, recorded because both were mine

1. **`rest_read` POST must declare `body_schema`.** The first generator emitted a POST read with a
   body and no schema; `connectorgen validate` refused it at bundle load. The engine compiles and
   validates that schema before the request leaves, so an absent one is a real gap, not ceremony.
2. **`rest_read` POST must declare `content_type: application/json`.** Caught by
   `checkCLISurfaceOperationSafety` on the next run. Both were fixed in the generator and the bundle
   regenerated from scratch, never patched by hand.
