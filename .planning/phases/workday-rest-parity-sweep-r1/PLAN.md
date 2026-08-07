# workday-rest — documented-operation parity (sweep slice)

Program: `cli-top50-fixed-schema-sweep-r1` · branch `fm/cli-top50-sweep-resume2-r1`.
**Landing order #2 under the captain's largest-first reversal**, behind github (done).

> **Banner (standing):** if this connector surprises you, **STOP and record it here** rather than
> forcing it into the batch shape. It already has, twice — see Hazards 1 and 2.

## Sliced delivery

| # | Slice | State |
| ---: | --- | --- |
| 1 | red test + `DERIVED-OPERATIONS.json` + this plan + `RUN-STATE.json` | ✅ **done, red observed** |
| 2 | `api_surface.json` → 916 dispositioned rows | ☐ |
| 3 | reads (`direct_read` commands) | ☐ |
| 4 | mutations (write actions) + full binary reachability sweep | ☐ |
| 5 | `SUMMARY.md` + `VERIFICATION.md` | ☐ |

## Artifact — a DIRECTORY, not a spec

| | |
| --- | --- |
| Manifest | `https://community.workday.com/sites/default/files/file-hosting/restapi/services2026.30.json` |
| Retrieved | 2026-08-07, HTTP 200, **617,538 bytes** — byte-identical to the sweep derivation |
| Contents | **52** production service entries + 38 archived (older versions of the same services) |
| Specs | each entry names its own spec file, fetched individually from the same host |

The `index.html` recorded on the connector is a React SPA shell with no server-rendered content; the
manifest above is the machine-readable directory behind it.

**There is no single API version.** Each of the 52 services is independently versioned (v1 … v7) and
mounted at its own base. An operation's identity is therefore the **resolved `(method, base+path)`
pair** — a service-relative path is ambiguous across 52 modules, and the red test rejects any row
that is not resolved.

## Derivation: 920 raw rows → **916** documented operations

Reproduced, not trusted: 52 specs → **920** raw method entries,
`GET 655 · POST 154 · PATCH 58 · DELETE 33 · PUT 20`, matching the sweep's recorded 920 and its
655/265 read-write split exactly.

**Then it does not hold up.** Deduped on the resolved path the provider itself publishes, the count
is **916** — `GET 654 · POST 153 · PATCH 58 · DELETE 32 · PUT 19`.

## Hazards, and the judgement each forces

### 1. ⚠️ Two service modules publish the SAME four endpoints — the count is 916, not 920

`Custom Object Data (multi-instance) v2` and `Custom Object Data (single-instance) v2` are two
separate directory entries that declare the **identical** `servers` URL
(`https://<tenantHostname>/customObject/v2`) and publish the same paths. Single- versus
multi-instance is a property of the custom **object**, not of the URL.

| Method | Resolved path |
| --- | --- |
| POST | `/customObject/v2/customObjects/{customObjectAlias}` |
| GET | `/customObject/v2/customObjects/{customObjectAlias}/{customObjectID}` |
| PUT | `/customObject/v2/customObjects/{customObjectAlias}/{customObjectID}` |
| DELETE | `/customObject/v2/customObjects/{customObjectAlias}/{customObjectID}` |

This is the **fourth** appearance of the sweep's recurring defect class — after notion's
`(body=markdown)`, lever-hiring's `?include=`, help-scout's `?async=true` and github's `(close)` —
and the first where the duplication is across **two provider service modules** rather than within one
document. The red test pins all four by name so a re-derivation cannot quietly reintroduce them.

**Finding 23 applies when these are authored:** the surviving row must not silently take only the
winning page's behaviour. Both descriptions are preserved in
`DERIVED-OPERATIONS.json.duplicates_collapsed`.

### 2. ⚠️ Three specs are OpenAPI 3.0.1, and that is what CAUSED hazard 1

49 of the 52 specs are **Swagger 2.0** and carry `basePath`. Three are **OpenAPI 3.0.1** and carry
`servers` instead — OAS3 has no `basePath` at all. A reader that looks only at `basePath` records an
**empty** base for those three, which is exactly what makes the two Custom Object services look like
they collide with each other and with Prism Analytics.

The three are also the only specs whose filenames are **not** stamped `20260727`: two are `20230712`
and one is `20231120`. The sweep's recorded derivation note asserts that *every* spec filename is
stamped 20260727. **That is wrong**, and it is wrong in exactly the place that matters.

### 3. Prism Analytics is mounted differently from all 51 others

Its server is `/api/prismAnalytics/v3/{tenant}` — an `/api` prefix *and* a server-level template
variable (`tenant`, default `super`). Every other service is `/<serviceName>/<version>`. Whoever
authors these rows must decide whether `{tenant}` is config-supplied or a flag, and record it. It is
the connector's only out-of-shape base.

### 4. The bundle is nearly empty: 4 rows against 916

`api_surface.json` holds **4** rows (3 GET + 1 POST) for the three legacy HCM read streams, with
`operation_ledger_version` **unset** and **one legacy `excluded` row**. There is **no**
`cli_surface.json`, **no** `writes.json`, **no** `operations.json`. `capabilities.write` is false and
will have to become true. This is the largest before/after gap in the sweep.

### 5. Read vs write is 654 / 262 — decide per operation, not per method

Workday is read-heavy. Do not assume every POST is a mutation without reading its summary; the sweep
has already found POST-shaped reads in github.

### 6. Auth is `oauth2_client_credentials`

Unlike github's token/app split. Confirm what the existing `spec.json` declares before authoring
anything that depends on it, and never accept a credential as a command flag.

## Work order per slice

Standard bar: `connectorgen validate` · **whole** `cmd/connectorgen` package ·
`TestEveryImplementedCommandPassesRuntimePreflight` · `connectorgen surface-sync --check` ·
**run the binary over every generated command**, asserting the rendered `NAME` line rather than the
exit code · no hand-authored paging flags · every blocked row carrying `Named dependency:`.

**Inherited from github, use it:** `covered_by.writes` now exists. If two write actions land on one
endpoint, list them — do not invent a variant path.

### 7. ⚠️ The four shipped rows are NOT among the 916

The bundle's three legacy streams point at `/ccx/api/hcm/v1/{tenant}/workers|organizations|jobs`
(plus one excluded POST). **The current service directory publishes no `hcm` service and no `/ccx/`
path anywhere** — worker resources live under `staffing/v7`, `absenceManagement/v5`,
`compensation/v3`, `performanceEnablement/v5`, `timeTracking/v5` and `api/common/v1`. They are not in
the archived list either, which holds only older versions of the 52 listed services. This is an
older Workday HCM REST shape that the published directory has superseded.

They therefore cannot be counted as documented — and they must not simply vanish, because deleting
them deletes three shipped, schema- and fixture-backed streams inside a parity commit (the
greenhouse finding-21 reasoning). The red test **pins them by name and counts them apart**, and
requires each to carry its own disposition, so the decision is made deliberately rather than by
arithmetic.

**Two defensible options, both requiring evidence:** re-point the streams at the documented service
endpoints, or disposition the legacy rows as superseded with the replacing service named. Either way,
say which and why in `SUMMARY.md`.
