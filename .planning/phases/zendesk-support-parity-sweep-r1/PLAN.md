# zendesk-support — documented-operation parity (GSD plan, TDD)

Program `cli-top50-fixed-schema-sweep-r1`, landing order **#3** under the captain's largest-first
reversal, behind github (1220) and workday-rest (907). Branch `fm/cli-top50-sweep-resume2-r1`.

Lifecycle resolved with `./scripts/gsd sources <command>` (a **Node** script — `bash scripts/gsd`
dies with a syntax error). Provenance for `plan-phase`, `execute-phase` and `verify-work` all
resolve to `.gsd/commands.json`, `.gsd/upstream.lock.json`, `.gsd/official-docs/COMMANDS.md`.

## Goal

Every operation Zendesk documents as executable REST is reachable from `pm zendesk-support`, and
every row that is not carries a **named** runtime dependency instead of a blanket sentence.

## The count: 625, and the first ledger in this sweep that reconciles

`https://developer.zendesk.com/zendesk/oas.yaml` re-fetched 2026-08-07 → HTTP 200,
**1,701,930 bytes**, byte-identical to the count `MASTER-PLAN.json` recorded. The derivation was
re-run from that artifact by `tools/derive_zendesk_support.py` rather than carried forward:

| | |
| --- | ---: |
| paths | 434 |
| operations | **625** |
| GET / POST / PUT / DELETE / PATCH | 325 / 111 / 89 / 86 / 14 |
| reads (GET + 8 read-shaped POSTs) | 333 |
| writes | 292 |
| binary | 1 |
| deprecated | 0 |
| union (oneOf/anyOf) request bodies | 3 |

**The provider ledger reconciles exactly, with a zero delta.** That is the first time in this sweep;
it has been wrong six times (notion +1, intercom −93, lever-hiring ~−40, help-scout's 146
query-string double-count, workday-rest 920 → 916 → 907). It reconciles because Zendesk publishes a
machine-readable spec keyed one operation per (method, path), not documentation pages. **A
reconciling total is evidence about this artifact, not a reason to trust the next ledger.**

Finding 34 applied rather than restated — the derivation was run through the red test's own rules
**before** the number was adopted: zero rows contain `?`, `*` or a space, and zero (method, path)
pairs repeat. Neither collapse that shrank workday-rest exists here.

**PyYAML cannot parse this artifact as shipped.** It contains bare `=` scalars (e.g. `change: =`
around line 36047), which YAML 1.1 resolves to `tag:yaml.org,2002:value`; SafeLoader has no
constructor for that tag and raises `ConstructorError`. The derivation registers one that treats it
as a plain string. Anything else in this programme calling plain `yaml.safe_load()` on this file
hits the same error and must apply the same workaround rather than silently guessing.

## Baseline — what the bundle already ships, and why this is not a from-nothing connector

Unlike github and workday-rest, zendesk-support already carries a **complete 631-row inventory**
(#3532, "Zendesk Support operation ledger parity"). The rows are all there. What is missing is that
**509 of the 631 are blocked behind one shared-foundation sentence** and their commands ship as
`availability: planned` placeholders.

| | count |
| --- | ---: |
| api_surface rows | 631 (625 documented + 6 carried) |
| covered | 122 — 33 `stream` + 89 `write` |
| blocked | 509 — none carrying a `Named dependency:` marker |
| cli_surface commands | 631 — 95 implemented, 27 partial, **509 planned** |
| operations.json | 514 — 306 `rest_read`, 202 `rest_write`, 5 `file_upload`, 1 `binary_download` |
| writes.json actions | 89 |
| streams | 33 |

**This connector's parity job is promotion, not enumeration.** The inventory is honest; the
availability claim is not. `AGENTS.md` is explicit that `availability: implemented` is a claim the
runtime has to honour — the inverse failure, a permanently-`planned` inventory, makes the surface
truthful and useless.

### The 6 carried rows

`GET /api/v2/business_hours/schedules.json`, `/community/posts`, `/community/topics`,
`/help_center/categories`, `/help_center/incremental/articles`, `/help_center/sections`. These are
**not** in the Support OAS — Zendesk publishes Guide and community endpoints in separate specs — and
they back shipped fixture-backed streams. They are counted **apart** from the 625, exactly as
workday-rest's legacy `/ccx/` rows are, so carrying them cannot inflate the parity number and a new
one cannot be smuggled in as a documented operation.

## The four non-mechanical judgements

### 1. Read vs write — METHOD decides, with eight documented exceptions

Every GET is a read. Every POST/PUT/PATCH/DELETE is a write **except** eight POSTs that query,
validate or preview and store nothing:

```
POST /api/v2/any_channel/validate_token
POST /api/v2/autocomplete/tags
POST /api/v2/custom_objects/{custom_object_key}/records/search
POST /api/v2/it_asset_management/assets/search
POST /api/v2/problems/autocomplete
POST /api/v2/users/autocomplete
POST /api/v2/views/preview
POST /api/v2/views/preview/count
```

Modelling these as writes puts a plan → preview → approval gate in front of a lookup. They are
pinned by name in the red test and re-checked against the artifact, per finding 32.

The inverse trap is real and the sweep's earlier derivation hit it: three POSTs match a naive
read keyword while literally saying "Creates a …" — `POST /api/v2/saved_searches` (matched via
`saved_searches`), `POST /api/v2/task_list_templates` and
`POST /api/v2/tickets/{ticket_id}/task_lists` (both via `task_list`). **The resource is named after
a read; the operation is a create.** All three are writes.

Two POSTs enqueue an asynchronous export job (`/audit_logs/export`, `/suspended_tickets/export`).
They return a job to poll rather than data, and creating a job resource is a write.

### 2. Stream vs direct read — the 33 shipped streams stay streams

Greenhouse finding 21: converting fixture- and schema-backed streams inside a parity commit deletes
shipped data contracts. The 33 existing streams keep their rows. Every other read becomes a bounded
direct read against the operation already declared in `operations.json`.

### 3. Binary detection — read `content`, and this connector's trap runs backwards

Exactly **one** operation declares `application/octet-stream`:
`GET /api/v2/custom_objects/{custom_object_key}/records/{record_id}/attachments/{id}/download`.

The trap is `PUT /api/v2/brands/{brand_id}`, the only *other* non-JSON success response in the whole
spec — it declares `image/jpg` **and** `image/png` because updating a brand echoes back its logo. A
rule that looked only for "declares a non-JSON media type" would ship it as a download and silently
drop the mutation. **Binary is GET-only** (finding 5), which is what makes the trap checkable rather
than a matter of judgement. workday-rest's trap ran the other way — paths that *sounded* binary and
declared JSON — so the pair together is the rule: read the artifact's media types, and require GET.

### 4. Named-dependency blocking

A row stays blocked only when a named runtime component refuses it, and the note names it. The three
classes expected here, each checkable:

- a GET whose documented success response declares no JSON body — `engine.decodeDirectReadBody`
  json-decodes the response and `commandrunner.supportedDirectReadOutputPolicies` declares no
  status-only policy;
- a write whose required record field has no scalar leaf — `availability: partial`, not
  `implemented`, mirroring `validate.go`'s `checkCLISurfaceWriteFlags` recursion;
- `PUT /api/v2/users/update_many`, whose request body is rooted at `oneOf`. `AGENTS.md`: a
  reverse-ETL `record_schema` rooted at `oneOf`/`anyOf` is not one executable command contract —
  runtime preflight expands its arms and rejects promotion. Model each reachable arm as a separately
  named action, or leave it non-implemented naming the missing capability. The other two union
  bodies are read-shaped POSTs, so their union sits in a search filter rather than a record
  contract.

## Paging — one judgement recorded, because it nearly went the other way

`start_time` is **not** treated as a paging parameter, and that is deliberate. Every stream in this
bundle declares pagination `type: next_url` over `links.next`, `next_page` or `before_url`, so the
foundation lane derives Zendesk paging from **response links**, never from a request parameter.
`start_time` on `/api/v2/incremental/*` is the export's required opening watermark — the endpoint
returns 400 without it — so it is an input to the operation, not a page control. The first
derivation run blocked it and would have made every incremental read unreachable while naming paging
as the reason. `page`, `per_page`, `limit`, `offset`, `cursor`, `page[...]` and the rest remain hard
blocked, and the generator raises rather than authoring one.

## Slices — each leaves shared gates green

Per the handoff's largest-first addendum: commit and push per slice; an intermediate commit may
leave **this** connector's own red test red, never a different connector or a shared gate.

1. **red** — derivation + `DERIVED-OPERATIONS.json` + this plan + `RUN-STATE.json` +
   `TDD-LEDGER.md` + the red surface test, run and captured verbatim. ← *this commit*
2. **reads** — 306 `rest_read` promotions + 8 read-shaped POSTs (`body_schema` +
   `content_type: application/json`, finding 33) + the 1 binary download; rows flip
   `operation` → `covered_by`.
3. **writes** — the remaining mutations, `oneOf` arms handled explicitly.
4. **reachability** — build the binary and assert the rendered `NAME` line for every command
   (finding 30: a namespace miss exits 0).
5. **SUMMARY.md** + **VERIFICATION.md**.

## Hazards carried in

- **F5** — a connector can have more than one surface test; run the **whole** `cmd/connectorgen`
  package before pushing, never a targeted `-run`.
- **F6/F4** — do **not** run `pm docs generate` per connector; it rewrites ~1,031 files of
  pre-existing `main` drift. Shared artifacts regenerate once at the end of the sweep.
- **Finding 36** — `go test ./internal/cli/` inherits Go's 600s default and dies mid-run. Always
  pass `-timeout 20m`.
- **Finding 3** — the operation endpoint ledger only emits an entry when `operation.rest.path`
  equals an `api_surface` row **verbatim**. Every zendesk row and every `operations.json` path is the
  documented `/api/v2/...` form, and the base is `<account>/api/v2`, so
  `engine.normalizeDirectReadPathForBaseURL` resolves them. Nothing here is out of base.
- **Finding 8** — inspect the regenerated ledger diff **by object**: 551 connectors before and
  after, none added, none removed, exactly one changed.
- **Finding 37** — a collapsed query-string behaviour must be `omit_when_absent`, never a fixed
  `query` value. No test catches the difference.
- **Missing tool** — the handoff points at `tools/check_red_observed.py`; it is not in the committed
  `tools/`. The red state below is therefore recorded with the verbatim `go test` output instead.
