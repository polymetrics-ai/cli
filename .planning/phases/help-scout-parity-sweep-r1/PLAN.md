# help-scout — documented-operation parity (sweep slice)

Program: `cli-top50-fixed-schema-sweep-r1` · branch `fm/cli-top50-sweep-continue-r1`.
Landing order **#5** (smallest first). **State: red test authored, red NOT yet observed — blocked on
a corrupted shared Go build cache. Do not author until red has been run and captured.**

> **Banner (standing):** if this connector surprises you, **STOP and record it here** rather than
> forcing it into the batch shape.

## Artifact

The ledger's recorded `artifact_url` (`…/mailbox-api/endpoints/`) **404s** — it is a section prefix,
not a page. Help Scout publishes **no machine-readable spec** (no `.json`/`.yaml`/openapi/swagger
linked from the docs, and no spec repo in the `helpscout` GitHub org).

The real inventory is the shared left-nav on `https://developer.helpscout.com/mailbox-api/`:
**146 unique `/mailbox-api/endpoints/**` page URLs** (183 raw hrefs before dedup — the nav repeats the
current-page entry, a rendering artifact, not a count signal). All 146 were fetched individually and
each renders **exactly one** `METHOD path` request line — no page with zero, none with two.

The full derived inventory is committed beside this file as `DERIVED-OPERATIONS.json`, with each
operation's templated path, doc page, example request and embedded-collection key.

## Count — 144, and BOTH deltas are dedup, not missing endpoints

| Source | Count | What it counts |
| --- | ---: | --- |
| Provider-artifact ledger | 146 | documentation **pages** |
| Sweep derivation | 145 | deduped by (method, **literal example** path) |
| **This slice** | **144** | deduped by (method, **templated** path) — the stated policy |

`GET 79 · POST 21 · PUT 20 · PATCH 6 · DELETE 18`.

**146 → 145.** `GET /v2/conversations/{conversationId}/threads/{threadId}/original-source` is
documented on two pages, one per `Accept` header (`application/json` vs `message/rfc822`). Content
negotiation on one endpoint. Their literal request lines are byte-identical, so the sweep derivation
already caught this.

**145 → 144.** `DELETE /v2/customers/{customerId}` is documented on two pages, *Delete Customer* and
*Delete Customer Asynchronously*. Their literal request lines **differ** —
`/v2/customers/100` vs `/v2/customers/100?async=true` — so deduping on the *example* path misses it.
Both pages declare the **same** templated path in their own `Path Parameters` block. `async=true`
switches the response from `204 No Content` to `202 Accepted`; it is a query parameter, not a second
endpoint.

**That is the exact defect class the captain flagged on lever-hiring: double-counting query-string
variants of one endpoint.** The counting is therefore done on the templated path Help Scout itself
publishes, never on the example URL, and the red test forbids a bare `?` in any row so it cannot
creep back in. The async behaviour is not lost — it becomes a flag on the one delete command.

## Baseline gap — this connector is 8 rows standing in for 144

| Check | Result |
| --- | --- |
| `api_surface.json` rows | **8** (6 GET, 2 POST) |
| Dispositions | 4 `covered_by` · **4 legacy `excluded`** · 0 blocked |
| Streams | **4** (conversations, customers, mailboxes, users) |
| `writes.json` | **ABSENT** — `capabilities.write` is false |
| `cli_surface.json` / `operations.json` | **ABSENT** |
| `operation_ledger_version` | unset |

One of the four excluded rows is **`GET /v2/reports/*`** — a **wildcard standing for 33 distinct
report endpoints**. A wildcard is a family of operations, not an operation, and the red test rejects
`*` in any row for that reason.

## Hazards, and the judgement each forces

### 1. Binary detection — exactly ONE download, and the trap is its sibling

- `GET /v2/conversations/{conversationId}/attachments/{attachmentId}/file` → **binary_download**.
  The page shows `Content-Disposition: attachment; filename="file.txt"`, `Content-Type: image/png`
  and a body of `binary_data`.
- `GET …/attachments/{attachmentId}/data` → **ordinary direct read**. It returns
  `{"data": "ZmlsZQ=="}` as `application/hal+json` — base64 *inside* JSON, not a file.

Treating `/data` as binary (or `/file` as JSON) is the mistake available here. They are one path
segment apart.

`…/threads/{threadId}/original-source` is **not** a download either: it is content-negotiated and
returns a body, with no `Content-Disposition`. Direct read.

### 2. Read vs write

All 79 GETs are reads. All 65 mutations are writes: **no POST is a disguised read** — Help Scout puts
every filter and search on GET query parameters (`GET /v2/conversations?query=…`), so there is no
search-shaped POST anywhere in the surface.

### 3. Stream vs direct read

Classified from each page's own title and response shape:

- **Streams** — pages titled `List …`, plus `Get Organization Conversations` / `Get Organization
  Customers`, whose responses are `_embedded` collections despite the "Get" wording.
- **Direct reads** — every singleton `Get …` page, **and all 33 report endpoints**. A report is an
  aggregate over a time window, not a record collection; syncing it as a stream would model a
  bounded fixed-target read as an ETL. `Get Conversation` embeds `_embedded.threads` and would be
  misclassified by an `_embedded`-only rule — the title is the reliable signal, not the response.

### 4. The five Mailbox API v3 rows are out of base

`base_url` is `https://api.helpscout.net/v2`, so `/v3/…` is outside it and hits the settled
endpoint-ledger constraint (finding 3). **Disposition: blocked with a named dependency.**

The named remedy is concrete: rebase `base_url` to the host root (`https://api.helpscout.net`) and
rewrite every stream and operation path to carry its own version prefix. That is a small change here
— only 4 streams exist — but it **changes a shipped `spec.json` default**, so any stored connection
whose `base_url` is `https://api.helpscout.net/v2` would start building `…/v2/v2/…`. That is a config
migration, and smuggling a user-visible breaking default change into a parity commit is not
acceptable. Recorded as the dependency, not performed here.

## Work order

1. ⛔ **Run the red test and capture the failure verbatim.** Currently blocked — see below.
2. `api_surface.json` → 144 rows, `operation_ledger_version: 1`, 0 excluded, 0 wildcards.
3. `streams.json` → expand from 4 to the full collection set.
4. `writes.json` → author from scratch (65 mutations); flip `capabilities.write`.
5. `operations.json` → direct reads + the one binary download.
6. `cli_surface.json` → every covered row reachable.
7. Gates, then **run the binary over every command**, then commit + push + tick.

## Blocker

The shared Go build cache (`~/Library/Caches/go-build`, 55 GB, shared by every lane on this host) was
corrupted during a host-wide disk-full window: the index still points at output files that were
truncated or never written, so `go test` fails with
`could not import <pkg> (open …/go-build/XX/<hash>-d: no such file or directory)` — **a different
missing entry on each retry**, so it is not self-healing.

Repairing it means `go clean -cache`, which is destructive to a resource every other lane is using.
That is not this lane's call to make. Escalated.
