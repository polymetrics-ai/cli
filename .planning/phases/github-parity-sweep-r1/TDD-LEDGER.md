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
