---
phase: workday-rest-parity-sweep-r1
program: cli-top50-fixed-schema-sweep-r1
connector: workday-rest
branch: fm/cli-top50-sweep-resume2-r1
coverage:
  documented_operations: 907
  covered: 907
  blocked: 0
  legacy_carried: 4
  commands_before: 0
  commands_after: 911
  write_actions_before: 0
  write_actions_after: 252
  reachable_verified: 911
---

# workday-rest — documented-operation parity

Landing order **#2** under the captain's largest-first reversal, behind github.

| | Before | After |
| --- | ---: | ---: |
| `api_surface` rows | 4 | **911** (907 documented + 4 legacy) |
| CLI commands | 0 (no `cli_surface.json`) | **911** |
| Write actions | 0 (no `writes.json`) | **252** |
| Operations | 0 (no `operations.json`) | **12** |
| `capabilities.write` | `false` | `true` |
| `operation_ledger_version` | unset | `1` |

**907 documented operations, 907 covered, 0 blocked.** Every documented operation is reachable and
executable; **all 911 commands were verified by running the binary**, asserting the rendered `NAME`
line rather than the exit code (sweep finding 30).

## The count: 920 → 916 → **907**

The artifact was **re-fetched, not trusted**. The manifest returned HTTP 200 at **617,538 bytes** —
byte-identical to slice 1 and to the sweep derivation — and all 52 production service specs were
re-fetched and the derivation reproduced independently.

| | Count | |
| --- | ---: | --- |
| Raw method rows across 52 specs | 920 | `GET 655 · POST 154 · PATCH 58 · DELETE 33 · PUT 20` — matches the master plan exactly |
| − published by two service modules | −4 | Custom Object Data single- and multi-instance v2 declare the **identical** `servers` URL |
| − query-string variants of an endpoint already counted | −9 | **found in this slice; slice 1 shipped all nine** |
| **Documented operations** | **907** | `GET 648 · POST 152 · PATCH 56 · DELETE 32 · PUT 19` |

Slice 1 deduped on the resolved `(method, base+path)` pair. That is what caught the four
custom-object rows, and it is also why the nine query-string rows survived: **passing the first dedup
made the second look unnecessary.** Sweep finding 22 already required a `?` guard in every red test —
the test had it; the derivation feeding it did not. Recorded as sweep finding 34.

Every one of the nine has its base-path sibling documented separately, so collapsing them loses no
endpoint. Seven carry an **empty summary** (the provider documenting an addendum, not an operation —
procurement's base row says "Retrieves the metadata **or the attachment content**"); the two that
carry text describe a *behaviour* of the base endpoint, which is finding 23's shape.

## The four judgements

### 1. Read vs write — 654 reads, 252 writes, decided per operation

654 direct reads against 252 write actions. **Seven POSTs are documented reads, not mutations**, and
they were read out of the provider's own summaries rather than inferred from the method:
`/wql/v1/data` is decisive — Workday itself calls it *"the read-only POST request"*. The others
validate, calculate against a hypothetical, check a permission, or look an ID up. None persists.

They ship as **operation-backed** reads (`kind: rest_read`, `content_type: application/json`,
`body_schema` — both fields required, and they were caught one run apart on github, finding 33) and
are the connector's only seven `operation_endpoint_ledger` entries.

**They are named `read-*`, not `create-*`.** The verb is the one thing a user reads before running a
command, and `pm workday-rest wql create-data` on a read-only endpoint would be a lie the surface
test cannot catch.

### 2. Stream vs direct read — 648 GETs became direct reads, and the 3 legacy streams stayed streams

New GETs are plain `direct_read` commands. A stream needs a hand-authored record schema, primary key
and fixture; inventing 648 of those would be inventing data contracts Workday never published. A
plain `direct_read` also adds **nothing** to the endpoint ledger (finding 26), which is why this
connector's ledger delta is 7 entries and not 654.

### 3. Binary detection — five endpoints, every one evidenced by `produces`

> **I got this wrong first and caught it before authoring.** `?type=viewContent`,
> `?type=getFileContent` and `?type=viewFile` all *sound* like file fetches, and I recorded six of
> the nine collapsed variants as binary on that basis. That is guessing from the path, which github's
> generator explicitly rejects. Read from `produces`, **only two** declare
> `application/octet-stream`. Recorded as sweep finding 35.

| Endpoint | `produces` | Shape |
| --- | --- | --- |
| `GET /api/prismAnalytics/v3/{tenant}/buckets/{id}/errorFile` | `octet-stream` only | pure `binary_download` |
| `GET /attachments/v1/graphql/{ID}` | `json` + `octet-stream` | dual |
| `GET /customerAccounts/v1/invoicePDFs/{ID}` | `json` + `octet-stream` | dual |
| `GET /procurement/v5/requisitions/{ID}/attachments` | collapsed variant declares `octet-stream` | dual |
| `GET /procurement/v5/requisitions/{ID}/attachments/{subresourceID}` | collapsed variant declares `octet-stream` | dual |

The four dual endpoints carry **one** documented row and **two** commands via
`covered_by.direct_reads` (plural). Modelling them as `direct_read` alone drops the file fetch;
as `binary_download` alone drops the metadata read; a `?type=` path row is the synthetic-variant
defect the sweep has now rejected five times. **workday-rest is the second connector to need
github's plural-array foundation fix, which confirms it generalised** rather than solving one
connector's problem.

Two Prism Analytics GETs declare `*/*` rather than `application/json` and are still modelled as
reads: OAS3's `*/*` is a wildcard *media type* carrying a real schema, and both point at a named
component object (`dataChangeResponse`, `apiObject`). Blocking them would claim Workday documents no
response body when it documents a typed one. Contrast `errorFile`, which declares `octet-stream` and
genuinely is a file.

### 4. Named-dependency blocking — one blocked row, and it is a legacy row

**Zero of the 907 documented operations are blocked.** Workday's specs are uniformly JSON with
declared response schemas, so no documented endpoint hits a runtime refusal. The single blocked row
is the legacy `POST /ccx/api/hcm/v1/{tenant}/workers`, carrying `model: deprecated` and naming
`staffing/v7 /workers` as its successor — greenhouse finding 18's shape.

## The legacy `/ccx/` rows: dispositioned, NOT re-pointed, NOT deleted

The four shipped rows point at `/ccx/api/hcm/v1/{tenant}/…`. The current directory publishes **no
`hcm` service and no `/ccx/` path**, and they are not in the archived list either, so they are not
among the 907. The red test pins them by name and **counts them apart**.

**They were not re-pointed at `staffing/v7`.** Re-pointing `workers` would keep the stream name while
silently changing the response shape out from under `schemas/workers.json` and its fixture — shipping
a stream whose declared contract no longer matches what the endpoint returns. That is worse than
carrying a superseded path. The three GET rows keep their `covered_by.stream` disposition; the POST
row is blocked as `deprecated` with its successor named. Tenants with the legacy HCM API enabled
still reach the three streams, and nothing shipped was deleted inside a parity commit (finding 21).

## Two things worth carrying to the next connector

- **Workday declares almost no required request-body fields.** Only **2 of 226** mutation body
  schemas declare a top-level `required` array, though 105 component schemas declare one somewhere
  nested. So nearly every write command's flags are its path variables and the body arrives via
  `--record`. `availability: partial` never triggered here — verified against the specs rather than
  assumed from the generator's zero count.
- **No `oneOf`/`anyOf`-rooted `record_schema`.** The `AGENTS.md` hazard about union-rooted reverse
  ETL contracts does not bite this connector; checked across all 52 specs, zero roots.

## Known unmet, deliberately

- **`TestGoldenTranscripts` — 11 subtests fail, and every one is pre-existing.** Verified by stashing
  this slice and re-running against the branch tip: the failing set is **identical**, so this work
  adds zero new failures. They are github manual/root-help drift, discharged by the end-of-sweep
  regeneration. **The handoff recorded this as one subtest (`root_bare_manual`); it is eleven.**
- **Website catalogs** (`website/data/connectors.generated.json`,
  `website/lib/connectors.catalog.data.generated.json`) still describe the old 3-stream connector.
  They are a shared artifact the program regenerates **once at the end**, not per connector, and no
  `make verify` gate covers them.
- **`docs/connectors/workday-rest/` regenerated; 1,029 other doc files reverted.** A bare
  `pm docs generate` rewrites 1,031 files of pre-existing `main` drift (finding F6).
