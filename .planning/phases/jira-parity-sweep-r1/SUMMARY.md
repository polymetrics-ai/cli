---
phase: jira-parity-sweep-r1
program: cli-top50-fixed-schema-sweep-r1
connector: jira
state: green-at-pause
coverage:
  documented_operations: 617
  covered: 590
  blocked: 27
  api_surface_rows_before: 15
  api_surface_rows_after: 617
  commands_before: 0
  commands_after: 590
  write_actions_before: 0
  write_actions_after: 292
  operations_json_entries: 25
  reachable_verified: 590
  endpoint_ledger_entries_before: 0
  endpoint_ledger_entries_after: 22
---

# jira — 617 documented operations, 590 reachable

**Paused mid-phase by captain order on 2026-08-07**, after the implementation slices landed green.
Everything below is committed and pushed; nothing lives only in the worktree.

## What changed

The shipped bundle carried **15** `api_surface` rows and no `cli_surface.json`, `writes.json` or
`operations.json` at all. Twelve of the fifteen were comma-joined or wildcard "and similar" families
standing for **602** endpoints — one `DELETE /rest/api/3/issue*, /comment*, /worklog*, /attachment*,
/project*, /filter*, /workflow* and related delete actions` row for all 89 documented deletes. A
wildcard row is not an operation (finding 24), so jira was a **restructure**, not an extension.

| | before | after |
| --- | ---: | ---: |
| `api_surface` rows | 15 | **617** |
| commands | 0 | **590** |
| write actions | 0 | **292** |
| `operations.json` entries | 0 | 25 |
| operation endpoint ledger entries | 0 | 22 |
| verified reachable by running the binary | — | **590 / 590** |

Disposition of the 617: 3 ETL streams (kept) · 270 plain direct reads · 3 binary downloads ·
22 read-shaped POSTs as operation-backed `rest_read` · 292 write actions (286 implemented,
6 partial) = **590 covered**; **27 blocked**, each naming a runtime component a reader can go and
check in `engine/`.

## Findings this connector produced

### 48. ⚠️ A ROLLING-SNAPSHOT ARTIFACT DEFEATS THE BYTE-COUNT CHECK, AND THAT IS A FINDING

Every earlier connector proved its artifact by re-fetching identical bytes. jira cannot. The ledger's
version-pinned URL (`?_v=1.8516.72`) returns **404**; the unpinned URL serves whatever is current;
`info.version` is `1001.0.0-SNAPSHOT-<git sha>`. Between `MASTER-PLAN.json`'s derivation and this
one — **the same calendar day** — the document went 2,445,625 → 2,449,760 bytes, 420 → 421 path keys
and **616 → 617** operations, gaining exactly one GET.

**The +1 GET cannot be named.** Naming it needs the prior artifact, which was never cached — only its
aggregate counts survive. Guessing would be an invention, so it is not guessed.

The remedy, and it generalises: record the artifact's **sha256**
(`5a51740d7ab3c77c521fc8895a7a58b4ff684bc0d2ebeb830135e8320b063ced`) in `DERIVED-OPERATIONS.json`,
`RUN-STATE.json` and `api_surface.json`'s scope prose. **Check `info.version` for a snapshot marker
before trusting any recorded byte count** — the sweep has 17 connectors left and Atlassian will not
be the only vendor serving a moving document from a stable URL.

### 49. "No schema declared" and "declared as a string" must stay distinguishable

Jira spells "the body is unconstrained" two ways: an absent `schema` key (the `*/*` avatar uploads)
and a literal `"schema": {}` (the entity-property PUTs). A dereferencer that falls through to
`type: string` collapses them into "documented scalar body". The first classification run put **14**
rows in `scalar_body` and **0** in `raw_binary_body`; the truth is 2 and 3, with 12 in
`unbounded_body`. Same shape as finding 44: the guard must test what was actually built.

### 50. A WRITE may be `partial` and still cover its row; a READ may not

`connectorgen validate` requires an implemented operation-backed direct read to bind every required
request-body path, and `covered_by.direct_read` accepts only an **implemented** command. So a
read-shaped POST whose required body field has no scalar leaf cannot be downgraded to `partial` the
way a write can — it must be **blocked**. Two are: `/workflows/create/validation` and
`/workflows/update/validation`, whose required `payload` is an object, and no `cli_surface` flag type
(boolean, string, integer, number, enum, string_array) carries one.

The corollary that saved a third: `commandBodyFlagCoveringRequiredPath` binds the **scalar leaf**,
not the container. `POST /rest/api/3/jql/sanitize` requires `queries.0.query`, and a flag
`maps_to: body.queries.0.query` covers it — google-calendar already ships `body.items.0.id`, so an
array-element leaf is a supported target rather than a guess. Binding the container would have
blocked a reachable operation.

### 51. The binary trap and its inverse sit on the SAME resource family here

Three GETs declare `image/png`; all three are avatar reads whose own summary says so.
`POST /rest/api/3/universal_avatar/type/{type}/owner/{entityId}` **uploads** an avatar. A rule keyed
on the path word "avatar", or on "declares a non-JSON media type anywhere in the operation", ships
the upload as a download and silently drops the mutation. Binary is **GET-only** (finding 45), and
success media must be read from **2xx responses only** — Atlassian attaches the same content map to
every response code, so the avatar reads declare `image/png` on their 401, 403 and 404 too.

### 52. `check_red_observed.py` did not exist

`PROGRESS.md` has said since github that this tool "enforces this and rejects placeholder text", and
`gen_planning_slice.py`'s contract table names it as the enforcement for red-first. Running it
raised `No such file or directory`. **Three workers were told an enforcement existed that did not.**
It is now committed and passes all ten phases. Its placeholder matching is anchored, because a bare
substring search flags `n/a` inside `expression/analyse`, a real Jira endpoint — a check that cries
wolf gets disabled.

### 53. Two pre-existing failing packages the handoff does not list

`internal/connectors/certify` and `internal/connectors/defs/zendesk-support` fail on this branch
**before** any jira change. The known-red list in `PROGRESS.md` names only the current connector's
own surface test and `TestGoldenTranscripts`. Measured, not assumed: the failing set is byte-identical
with and without the jira slice (`git stash push -u`, re-run, diff — only durations differ). Carried
into the handoff so the next worker does not spend a context deciding whether they caused it.

## Non-mechanical judgements, recorded

- **read-vs-write** — 24 POSTs read. Three break a keyword pass in both directions:
  `bulkSetIssuesPropertiesList` matches "list" and **sets** properties (write), while
  `analyseExpression` and `evaluateJiraExpression` match no read keyword and mutate nothing (reads).
  Both directions are pinned in the surface test.
- **stream-vs-direct-read** — the three shipped streams (issue/project/user search) stay streams;
  they carry hand-authored schemas and fixtures, and converting them inside a parity commit would
  delete shipped functionality (finding 21). The test asserts they are still streams.
- **binary detection** — see finding 51.
- **named-dependency blocking** — 27 rows, six classes, every note naming a component in `engine/`.
