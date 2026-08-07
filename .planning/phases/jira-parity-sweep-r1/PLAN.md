# jira — connector parity sweep plan

> **If jira surprises you, STOP and record it** rather than forcing it into the batch shape. This
> plan is generated from `MASTER-PLAN.json` plus a fresh derivation; the per-connector findings are
> deliberately NOT pre-planned.

Program: `cli-top50-fixed-schema-sweep-r1`. Branch: `fm/cli-top50-sweep-resume2-r1`.
Landing order: **4 of 21** under the captain's largest-first reversal (github 1220 → workday-rest
907 → zendesk-support 625 → **jira 617** → stripe 589 → …).

## Goal

Bring `pm jira` to parity with Jira Cloud platform REST API v3's **documented** operation surface,
where parity means **reachability**, not inventory (finding 40).

## Starting state — this is a RESTRUCTURE, not an extension

The shipped bundle carries **15** `api_surface` rows and no `cli_surface.json`, `writes.json` or
`operations.json` at all. Twelve of the fifteen are comma-joined or wildcard "and similar" families:

```
GET  /rest/api/3/issue/{issueIdOrKey}/comment, /worklog, /watchers, /votes, /remotelink,
     /properties and similar issue subresources
GET  /rest/api/3/auditing/record, /events, /field, /issuetype, /priority, /resolution, /status,
     /workflow*, /screen*, /permission*, /configuration and other platform/admin metadata
DELETE /rest/api/3/issue*, /comment*, /worklog*, /attachment*, /project*, /filter*, /workflow*
     and related delete actions
```

A wildcard row is not an operation (finding 24) and a comma-joined row is mixpanel's collapse defect
(PROGRESS.md "mixpanel STOPPED"). Twelve rows stand for **602** endpoints. The three real rows are
the ETL streams. So the count moves **15 → 617** and every wildcard family is replaced, not extended.

## Derived surface — 617 operations

Derived by `tools/derive_jira.py` from Atlassian's own OpenAPI description. 421 path keys, one
operation per `(method, path)` under `paths`.

| | count |
| --- | ---: |
| GET | 276 |
| POST | 134 |
| PUT | 118 |
| DELETE | 89 |
| **total** | **617** |
| reads (GET + 24 read-shaped POSTs) | 300 |
| writes | 317 |
| deprecated (a disposition, not an exclusion — finding 18) | 29 |

**The derivation was run through the red test's own rules before its number was adopted**
(finding 34): zero paths contain `?`, `*` or a space; zero `(method, path)` pairs repeat templated;
zero repeat once path variables are normalised to `{}`. Neither collapse that took workday-rest from
920 to 907 applies — Atlassian publishes no query-string variant path keys and no endpoint under two
modules.

## Hazards

### 1. ⚠️ THE BYTE-COUNT CHECK IS NOT AVAILABLE FOR THIS CONNECTOR

Every earlier connector in this sweep proved its artifact by re-fetching **identical bytes**. jira
cannot: the version-pinned URL the ledger records (`?_v=1.8516.72`) returns **404**, the unpinned URL
serves a rolling snapshot, and `info.version` is `1001.0.0-SNAPSHOT-<git sha>`. Between the master
plan's derivation and this one — **the same calendar day** — the document went

```
2,445,625 bytes → 2,449,760 bytes     420 path keys → 421     616 operations → 617   (+1 GET)
```

with a different sha in `info.version`. **A byte match proves nothing here and its absence disproves
nothing.** What replaces it is the artifact's sha256, recorded in `DERIVED-OPERATIONS.json`, in
`RUN-STATE.json` and in `api_surface.json`'s own scope prose:
`5a51740d7ab3c77c521fc8895a7a58b4ff684bc0d2ebeb830135e8320b063ced`.

The +1 GET **cannot be named**: naming it needs the prior artifact, and the prior artifact was never
cached — only its aggregate counts survive in `MASTER-PLAN.json`. Saying which endpoint appeared
would be a guess, so it is not said. A future worker who re-fetches and gets a third sha has learned
that the snapshot moved again, which is a finding about Atlassian, not about this bundle.

### 2. Binary detection runs in BOTH directions here, on the same resource family

Exactly three operations declare an image media type on a success response, all GETs, all avatar
image fetches whose own summary says "Returns … avatar image":

```
GET /rest/api/3/universal_avatar/view/type/{type}
GET /rest/api/3/universal_avatar/view/type/{type}/avatar/{id}
GET /rest/api/3/universal_avatar/view/type/{type}/owner/{entityId}
```

**The trap sits next to them.** `POST /rest/api/3/universal_avatar/type/{type}/owner/{entityId}`
*uploads* an avatar. A rule keyed on "avatar", or on "declares a non-JSON media type anywhere", ships
it as a download and silently drops the mutation — finding 45's shape, on the same resource. Binary
is **GET-only**, which is what makes the judgement checkable rather than a matter of taste.

A second trap: Atlassian attaches the SAME content map to every response code, so the avatar reads
declare `image/png` on their 401, 403 and 404 too. Read from **2xx** responses only.

### 3. Read-vs-write cannot be decided by keyword, and this connector proves it

24 POSTs are read-shaped. Three of them break a keyword pass in both directions:

- `POST /rest/api/3/issue/properties` — operationId `bulkSetIssuesPropertiesList` matches "list", and
  its description reads "**Sets or updates** … on issues". It is a **write**.
- `POST /rest/api/3/expression/analyse` and `/expression/eval` match no read keyword at all and
  persist nothing. They are **reads**.

Each of the 24 was checked against its own `description`, and both directions are pinned in the
surface test so the judgement cannot drift silently.

### 4. Paging is NEVER hand-authored

`streams.json` already declares this connector's pagination (`offset_limit` over `startAt` /
`maxResults`); the foundation lane derives flags from it. `startAt`, `maxResults`, `startIndex` and
friends are on the generator's refusal list and it raises rather than emitting one.

### 5. Webhook EVENTS are excluded; webhook MANAGEMENT is in scope

The artifact declares **zero** webhook events (no OAS 3.1 `webhooks` block — jira does not block
`connectorgen batch`, unlike notion/chargebee/bamboo-hr). Its five webhook management endpoints are
ordinary operations and are counted.

## Coverage plan — 592 covered / 25 blocked

| disposition | count |
| --- | ---: |
| ETL streams (kept) | 3 |
| direct reads (plain GET, no operations.json entry — finding 26) | 270 |
| binary downloads | 3 |
| read-shaped POSTs (`rest_read`, operation-backed) | 24 |
| write actions (286 implemented + 6 partial) | 292 |
| **covered** | **592** |
| blocked, each naming the runtime component that refuses it | 25 |

Blocked classes, all derived from what Atlassian's spec does and does not declare, and all checked
on the **built** record contract rather than on its inputs (finding 44):

| class | n | named dependency |
| --- | ---: | --- |
| `unbounded_body` | 12 | body declared as arbitrary JSON (`"schema": {}` on nine entity-property PUTs; an object with no properties on three JSON-Patch plan updates). No bounded record contract to derive; inventing one is the generic HTTP write `AGENTS.md` forbids. |
| `dynamic_key_map` | 5 | body declared `additionalProperties: <object>` keyed by custom field / scheme id. `engine`'s `dynamic_fields` region is the one declared capability for a dynamic-key payload and `validateDynamicFields` accepts **scalar** `value_types` only. |
| `raw_binary_body` | 3 | avatar uploads declared `*/*` with no schema. `engine`'s write body types are json, form, none, graphql, json_array, multipart, base64_upload; none emits a raw byte stream, and inline raw bytes are banned outright. |
| `empty_contract` | 3 | no body, no path variable, no required query parameter. `engine.PreflightWriteAction` refuses a `record_schema` admitting only `{}`. |
| `scalar_body` | 2 | body is a bare JSON string. `buildJSONBody` assembles an object from record fields and `json_array` covers a top-level array; nothing emits a top-level scalar, and `body_type: none` would send the request with the documented value silently dropped. |

**A DELETE is deliberately NOT blocked for "no request body"** — it is addressed by its path, so
that is its normal shape. Same judgement zendesk-support made, for the same reason.

## Slices — each leaves the tree green on every SHARED gate

1. **red** — derivation + `DERIVED-OPERATIONS.json` + surface test written, RUN, failure captured
   verbatim + this plan + `RUN-STATE.json`. Commit, push.
2. **surface + reads + writes** — `api_surface.json` rewritten to 617 rows, `cli_surface.json`,
   `operations.json`, `writes.json` generated. Rows and commands land together: a `covered_by`
   disposition must name a command that already exists, so slices 2 and 3 cannot be separated
   (github's lesson). Commit, push.
3. **reachability** — build the binary and run **every** implemented command, asserting the rendered
   `NAME` line (finding 30: a namespace miss exits 0). Commit, push.
4. **`SUMMARY.md` + `VERIFICATION.md`**. Commit, push.

Each slice may leave jira's own surface test red — that is correct and honest — but must never leave
a different connector or a shared gate broken.

## Gates

`go run ./cmd/connectorgen validate` (551/0) · `go run ./cmd/connectorgen surface-sync --check` ·
the **whole** `cmd/connectorgen` package (finding 5: a connector can have more than one surface
test) · runtime preflight via `TestEveryImplementedCommandPassesRuntimePreflight` ·
`go test -timeout 20m ./internal/cli/` (finding 36: the bare form inherits Go's 600s default and
dies mid-run) · endpoint-ledger delta inspected **by object, not by line**.

## Required skills

`golang-how-to`, `golang-testing`, `golang-cli`, `golang-security`, `golang-safety`,
`golang-error-handling`.
