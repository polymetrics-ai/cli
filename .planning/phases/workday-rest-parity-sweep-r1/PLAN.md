# workday-rest — documented-operation parity (sweep slice)

Program: `cli-top50-fixed-schema-sweep-r1` · branch `fm/cli-top50-sweep-resume2-r1`.
**Landing order #2 under the captain's largest-first reversal**, behind github (done).

> **Banner (standing):** if this connector surprises you, **STOP and record it here** rather than
> forcing it into the batch shape. It already has, twice — see Hazards 1 and 2.

## Sliced delivery

| # | Slice | State |
| ---: | --- | --- |
| 1 | red test + `DERIVED-OPERATIONS.json` + this plan + `RUN-STATE.json` | ✅ **done, red observed — then RE-observed at 907, see Hazard 8** |
| 2 | `api_surface.json` → **907** dispositioned rows | ✅ |
| 3 | reads (`direct_read` commands) | ✅ 654 |
| 4 | mutations (write actions) + full binary reachability sweep | ✅ 252 writes, 911/911 reachable |
| 5 | `SUMMARY.md` + `VERIFICATION.md` | ✅ |

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

## Derivation: 920 raw rows → 916 → **907** documented operations

Reproduced, not trusted, **twice** — the second time by re-fetching the manifest (HTTP 200,
**617,538 bytes**, byte-identical) and all 52 specs: **920** raw method entries,
`GET 655 · POST 154 · PATCH 58 · DELETE 33 · PUT 20`, matching the sweep's recorded 920 and its
655/265 read-write split exactly.

**Then it collapses twice.** Deduped on the resolved path the provider itself publishes, the count is
916. Deduped again on the **templated** path — dropping nine query-string variants of endpoints
already counted — it is **907**: `GET 648 · POST 152 · PATCH 56 · DELETE 32 · PUT 19`.

See Hazard 8. The `916` in slice 1 was a real derivation defect, not a rounding argument.

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

### 8. ⚠️ The count collapses a SECOND time: 916 → 907, and that reshapes six commands

Slice 1 deduped on the resolved `(method, base+path)` pair. That caught Hazard 1's four custom-object
rows and **never looked for a `?`**. Nine of the 916 are query-string variants of endpoints already
counted — the provider publishes each as its own Swagger path key with the query string baked in,
and **every one has its base-path sibling documented separately**, so collapsing them loses no
endpoint.

| Method | Variant path | Variant summary |
| --- | --- | --- |
| GET | `/accountsPayable/v1/supplierInvoiceRequests/{ID}/attachments?type=viewContent` | *(empty)* |
| GET | `/accountsPayable/v1/supplierInvoiceRequests/{ID}/attachments/{subresourceID}?type=viewContent` | *(empty)* |
| GET | `/procurement/v5/requisitions/{ID}/attachments?type=getFileContent` | *(empty)* |
| GET | `/procurement/v5/requisitions/{ID}/attachments/{subresourceID}?type=getFileContent` | *(empty)* |
| GET | `/recruiting/v4/prospects/{ID}/resumeAttachments?type=viewFile` | *(empty)* |
| GET | `/recruiting/v4/prospects/{ID}/resumeAttachments/{subresourceID}?type=viewFile` | *(empty)* |
| PATCH | `/staffing/v7/workers/{ID}/checkInTopics/{subresourceID}?type=archive` | "…to archived or un-archived" |
| PATCH | `/staffing/v7/workers/{ID}/checkIns/{subresourceID}?type=archive` | "…to archived or un-archived" |
| POST | `/api/common/v1/workers/{ID}/businessTitleChanges?type=me` | *(empty)* |

**Seven carry an empty summary** — documented as an addendum to the base row, not as an operation.
Procurement's base row says it outright: *"Retrieves the metadata **or the attachment content** of
the specified requisition."* One endpoint, two modes. **The two staffing PATCHes** carry a summary
describing a **behaviour** of the base endpoint, which is finding 23's shape exactly.

#### The judgement this forces, and it is not the count

> **Corrected before authoring.** My first pass wrote "six of the nine are the binary mode",
> inferring binary from the `type=viewContent` / `getFileContent` / `viewFile` names. **That is
> guessing from the path, which this sweep has explicitly rejected** (github's generator: "read out
> of the artifact, never guessed from the path"). Read from `produces`, only **two** of the nine
> declare `application/octet-stream` — both `procurement ?type=getFileContent`. The accountsPayable
> `?type=viewContent` and recruiting `?type=viewFile` variants declare `application/json` **only**,
> so they collapse into a plain metadata read and nothing binary is lost.

**Two** of the nine (`procurement ?type=getFileContent`) are the **binary** mode of an attachment
endpoint whose default mode returns JSON metadata. **This is the connector's binary-detection
judgement**, and it must not be resolved by picking one mode:

- Modelling the row as `direct_read` alone silently drops the ability to fetch the file.
- Modelling it as `binary_download` alone silently drops the metadata read.
- Inventing a `?type=` path row is the synthetic-variant defect the sweep has now rejected five times.

**Use `covered_by.direct_reads` (plural)** — the array github's foundation fix generalised — to hang
**both** a metadata read and a binary download off **one** documented endpoint row. That is precisely
what the plural arrays exist for.

The remaining three (`?type=archive` ×2, `?type=me`) re-express as a **flag** on the surviving write
action (help-scout's `--async` pattern), never as a second path.

#### The full binary surface, read from `produces` across all 52 specs

Five endpoints, and every one is evidenced by a declared `application/octet-stream`:

| Endpoint | `produces` | Shape |
| --- | --- | --- |
| `GET /api/prismAnalytics/v3/{tenant}/buckets/{id}/errorFile` | `octet-stream` only | pure `binary_download` |
| `GET /attachments/v1/graphql/{ID}` | `json` + `octet-stream` | dual: metadata read **and** download |
| `GET /customerAccounts/v1/invoicePDFs/{ID}` | `json` + `octet-stream` | dual |
| `GET /procurement/v5/requisitions/{ID}/attachments` | base `json`; collapsed variant `octet-stream` + `json` | dual |
| `GET /procurement/v5/requisitions/{ID}/attachments/{subresourceID}` | base `json`; collapsed variant `octet-stream` + `json` | dual |

The four dual endpoints are exactly what `covered_by.direct_reads` (plural) is for: **one** documented
row, **two** commands, no synthetic variant path.

### 9. Two Prism Analytics GETs declare `*/*`, not `application/json`

`GET /api/prismAnalytics/v3/{tenant}/dataChanges/{dataChangeID}` and `…/validate` declare their 200
content as `*/*`. Read literally that is not `application/json`, which is github's blocking rule.

**They are modelled as reads anyway, and this is a judgement, not an oversight.** OAS3's `*/*` is a
wildcard *media type* carrying a real schema, and both point at a named component object schema
(`dataChangeResponse`, `apiObject`) — a JSON object, not a file. Blocking them would claim the
provider documents no response body when it documents a typed one. Contrast the errorFile endpoint
above, which declares `octet-stream` and genuinely is a file.
