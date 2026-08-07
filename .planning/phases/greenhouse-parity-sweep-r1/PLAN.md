# greenhouse — documented-operation parity (sweep slice)

Program: `cli-top50-fixed-schema-sweep-r1` · consolidated branch `fm/cli-top50-sweep-consolidated`
→ continued on `fm/cli-top50-sweep-continue-r1`. Landing order **#4** (smallest first).

> **Banner (standing):** if this connector surprises you, **STOP and record it here** rather than
> forcing it into the batch shape.

## Artifact

| | |
| --- | --- |
| URL | `https://developers.greenhouse.io/harvest.html` |
| Kind | Slate-generated static HTML reference (no OpenAPI/Swagger exists for Harvest) |
| Retrieved | 2026-08-07, HTTP 200, **1,636,662 bytes** — byte-identical to the sweep derivation |
| Extraction | the canonical `<h3>HTTP Request</h3><p><code>METHOD https://harvest.greenhouse.io/vN/path</code></p>` declaration that follows every endpoint section |

## Count — 138, reconciles exactly

| Ledger | Re-derived | Delta |
| ---: | ---: | ---: |
| 138 | **138** | +0 |

Method split `GET 69 · POST 28 · PUT 8 · PATCH 19 · DELETE 14`, zero duplicates. Independently
reproduced in this lane against a freshly fetched copy of the same page; byte count matched the
recorded derivation exactly, so this is the same artifact, not a lookalike.

Webhook events: **0 in scope.** Greenhouse documents "Recruiting Webhooks" on a separate page
(`developers.greenhouse.io/webhooks.html`) and exposes **no webhook management endpoints** on
Harvest, so there is nothing to count in either direction.

## Baseline gap — the count was right in aggregate but the bundle was not

The shipped bundle carries **129** rows: 126 `covered_by`, **3 legacy `excluded`**, 0 blocked,
`operation_ledger_version` unset. `cli_surface.json` and `operations.json` are **absent**, so
`pm greenhouse` returns `unknown command` and not one of its 126 covered operations is reachable —
the same defect gmail had.

Nine rows are missing, and they are missing for two distinct reasons:

1. **One markup-damaged declaration.** `DELETE /v1/tags/candidate/{tag id}` — Greenhouse's own docs
   emit a stray unescaped `&#39;` before the URL and a placeholder containing a literal space. A
   naive regex drops it; so did the shipped bundle.
2. **Eight Harvest v2 operations.** Three `<h2>` sections document *two* versioned operations under
   one heading (a deprecated v1 and its v2 replacement); the other five v2 operations
   (`job_posts` ×2, `users` ×3) have no v1 sibling at all.

## Hazards, and the judgement each forces

### 1. Out-of-base v2 rows — blocked, and the block is structural

The bundle binds exactly one HTTPBase, `https://harvest.greenhouse.io/v1`. The runtime operation
endpoint ledger (`deriveOperationDirectReadEndpointLedger`) only emits an entry when
`operation.rest.path` equals an `api_surface` row **verbatim**, which closes both doors for a `/v2`
path: a base-relative form builds `…/v1/v2/…`, and an absolute form is not in the surface so
preflight refuses it. This is finding 3 on the sweep, settled on chatwoot and not re-litigated here.

**Disposition: all 8 blocked**, recorded at their documented host-root path (`/v2/…`) so they cannot
collide with a base-relative v1 row, each carrying a `Named dependency:` marker naming the
per-operation base URL override.

### 2. The 3 deprecated v1 mutations are counted, not excluded

`DELETE /v1/jobs/{job_id}/openings`, `POST /v1/scheduled_interviews`,
`PATCH /v1/scheduled_interviews/{id}` are struck through in the docs as DEPRECATED in favour of v2.
The counting policy counts deprecated operations. The shipped bundle parked them under a legacy
`excluded` stub, which is not one of the three dispositions this sweep accepts.

**Disposition: blocked with `model: deprecated`**, each naming its v2 replacement — which is itself
blocked by hazard 1. That is a *named, checkable* reason, not a shrug.

### 3. `PUT /v1/candidates/{id}/anonymize?fields={field_names}`

The documented path embeds a query-string template. It collides with nothing, so it does not change
the count. It is kept verbatim because the shipped bundle already covers it that way and the write
executes today; rewriting it would change a working operation for cosmetics.

### 4. Read vs write

Harvest is conventional: `GET` is a read, everything else mutates. No POST is a disguised read —
checked; there is no search/query POST in the surface. Nothing here needs the semantic POST
classification that bit other connectors.

### 5. Stream vs direct read

`GET` collection endpoints are streams (the bundle already models 69 GETs, mostly as streams).
Detail/singleton GETs that take a required path id are **direct reads**, since a stream over a
single fixed record is not a sync. Binary: **none** — Harvest returns JSON everywhere; attachments
are referenced by URL, not served by the API. So there is no `binary_download` operation here.

## Work order

1. ✅ Red test `cmd/connectorgen/greenhouse_api_surface_test.go`, run, failure captured verbatim in
   `TDD-LEDGER.md` and `RUN-STATE.json`. Committed red.
2. `api_surface.json` → 138 rows, `operation_ledger_version: 1`, 0 excluded, every row exactly one
   disposition, every blocked row carrying `Named dependency:`.
3. `writes.json` → add the recovered `DELETE /tags/candidate/{tag_id}`.
4. `cli_surface.json` + `operations.json` authored from scratch so every covered row is reachable.
5. Gates: `connectorgen validate` · **whole** `cmd/connectorgen` package (finding F5) ·
   `TestEveryImplementedCommandPassesRuntimePreflight` · `connectorgen surface-sync --check`.
6. **Run the binary** over every generated command; authoring it is not evidence it works.
7. One commit, push, tick `PROGRESS.md`.

## Constraints that bind this slice

- ⛔ **No hand-authored paging flags** (`page`, `per_page`, `limit`, `offset`, `cursor`, …). A
  foundation lane derives them. If one seems required, record and escalate.
- Shared generated artifacts (endpoint ledger, website catalogs, golden transcripts, docs) are
  regenerated **once at the end of the sweep**, not per connector.
- Never weaken, skip, or delete a test.
