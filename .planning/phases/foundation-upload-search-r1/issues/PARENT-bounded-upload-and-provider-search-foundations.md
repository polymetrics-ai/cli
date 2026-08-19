# feat(connectors): bounded multipart upload and bounded provider-search runtime foundations

> **Draft — not created on GitHub.** Issue/PR creation for this program is supposed to use the
> `alfred-polymetrics-ai` account. No Alfred credential exists in this environment (`gh auth status`
> shows only `karthik-sivadas`, the captain's own account), and GitHub authorship cannot be
> reattributed after the fact. This tree is drafted to files so the execution record exists; a
> maintainer with the Alfred credential should create it verbatim.

## Objective

Close the shared-runtime gaps that connector lanes are blocked on, all typed and inspectable, with no
raw method/path/body escape hatch:

1. **Sound containment for multipart upload file access** — `os.Root` instead of a lexical
   check-then-open.
2. **A declared, enforced media-type bound on uploads** — what makes "image upload" a contract rather
   than an unverified assertion.
3. **Bounded provider-search** — a typed, list-bounded POST-body read, distinct from warehouse
   `pm query`.

> **The upload *transport* is not one of these gaps.** It already ships and is reachable from the
> CLI — see "Verified state" below. Building a second executor was the original framing of this work
> and is explicitly rejected.

Scope is `internal/connectors/engine/`, `internal/connectors/connsdk/`, and the command runner glue
plus tests. **No connector bundle under `internal/connectors/defs/**` is edited** — the connector
path-ownership guardrail is live on `main` and connectors adopt these capabilities in their own lanes
afterwards.

## Verified state of the repository (read directly, not assumed)

### Bounded multipart upload already ships — proven by execution, not by reading

```
$ go build -o pm ./cmd/pm && ./pm gong calls upload-media --help
NAME          pm gong calls upload-media - Add call media (/v2/calls/{id}/media)
INTENT        reverse_etl
AVAILABILITY  implemented
WRITE         upload_call_media
NOTES         Uses typed multipart write support; no generic upload command is exposed.
FLAGS
  --id (string): ... maps_to=record.id
  --media-file-path (string): Project-relative media file path to upload. maps_to=record.media_file_path
```

`cli_surface.json` command → `writes.json` action with `body_type: multipart` →
`engine/write.go:430-436` → `buildMultipartPayload` (`write.go:506-556`) → `connsdk.DoMultipart`
(`connsdk/http.go:244-256`). Gong is the only adopter — a search of every
`internal/connectors/defs/*/*.json` for `"multipart"` returns exactly `gong/writes.json`
(`upload_call_media` 1.5 GiB cap, `upload_crm_entities` 200 MiB cap). Bounded, digest-approval-bound,
record-driven (never argv), behind plan → preview → approval → execute.

Freshchat's two upload rows say *"the current **Freshchat bundle** has no connector-local
binary/multipart execution contract"* — a statement about the bundle, not the runtime. Gong proves
the runtime contract exists.

### The `file_upload` operation kind is a dead parallel declaration

| Surface | Status |
| --- | --- |
| Kind in schema enum | present — `internal/connectors/engine/schema/operations.schema.json:30` |
| Kind in block map | present — `internal/connectors/engine/bundle.go:1315` (`file_upload` → `file`) |
| Semantic validation | present — `bundle.go:1368-1377` (direction must be `upload`, positive `max_bytes`, approval required) |
| Spec struct | `FileOperationSpec{Direction, Path, MaxBytes}` — `bundle.go:551-555` |
| Executor | **absent** |
| Runtime behaviour | hard-blocked at `internal/connectors/commandrunner/runner.go:239-247` — any command with a non-empty `operation` that is not an implemented `direct_read` returns `operation %s executor is not implemented in this slice` |

**32 `file_upload` operations are already declared**, counted from
`internal/connectors/defs/*/operations.json`:

| connector | count | declared `max_bytes` |
| --- | ---: | --- |
| xero | 22 | 26 214 400 (25 MiB) |
| zendesk-support | 5 | 10 485 760 (10 MiB) |
| bitbucket | 4 | 104 857 600 (100 MiB) |
| asana | 1 | 26 214 400 (25 MiB) |

Unlike `binary_download`, **no certification stage asserts uploads stay blocked** — there is no
upload analogue of `internal/connectors/certify/stages_binary.go:30-41`.

Reconciling these 32 declarations against the live `writes.json` contract requires editing connector
bundles, which the path-ownership guardrail forbids on a foundation branch. **Recorded as a follow-up
for the connector lanes; not built here.**

### The multipart machinery must be reused, not re-invented

`connsdk` already does bounded multipart correctly:

- `Requester.DoMultipart` — `connsdk/http.go:244-256`
- `validateMultipartForm` — `http.go:258-289` (regular-file check, per-file and aggregate size caps)
- `snapshotApprovedMultipartFiles` — `http.go:291-343` (digest-approval binding, cleanup on failure)
- `snapshotMultipartFile` — `http.go:345-388` — the template the download research names:
  `LimitReader` → `io.Copy` into `os.CreateTemp` → SHA-256 via `io.MultiWriter` → **read one byte past
  the limit** to detect overflow → cleanup defer

The only caller is the reverse-ETL write path: `engine/write.go:430-436` (`bodyTypeOf == "multipart"`)
→ `buildMultipartPayload` (`write.go:506-556`) → `resolveMultipartFilePath` (`write.go:558-596`).
That path is driven by a **record field**, never by argv.

**So the genuinely missing pieces are narrow**, and this program must not duplicate the above:

1. **Containment is lexical and TOCTOU-racy.** `resolveMultipartFilePath` calls
   `safety.ValidateLocalWritePath`, which does no symlink resolution at all
   (`internal/safety/safety.go:128-158` — verified: `Clean` + `Rel` + prefix compare), then bolts
   `filepath.EvalSymlinks` + `requireInsideRoot` on separately. After that one check the file is
   re-opened **three more times by path with no containment** — `os.Stat` at `connsdk/http.go:273`,
   `os.Open` at `:346`, `os.Open` at `:475`. The download research's conclusion applies unchanged:
   use `os.Root`.
2. **Nothing validates that the bytes sent match the media type the part claims.**
   `MultipartPartSpec.ContentType` (`engine/bundle.go:414-421`) goes straight into the part header
   (`connsdk/http.go:462-472`); a ZIP uploads happily under a declared `image/png`.

### Bounded provider-search is *closer* than expected — and one bound is impossible to express

`OperationDirectRead` (`engine/direct_read.go:32-129`) **already** executes a POST with a JSON body:

- POST is accepted (`direct_read.go:44-46`)
- `content_type` must be `application/json` (`:50-52`)
- `body_schema` is **required** for POST (`:53-55`)
- the body is built from the operation's declared `rest.body` plus caller overrides and then
  **validated against `body_schema`** (`operationReadBody`, `:229-247`)
- the endpoint must be connector-relative (`:47-49`) and declared in `api_surface` (`:214-227`)

The CLI input path is already typed, not a raw body: flags declare `maps_to: body.<field>` and are
coerced per declared flag type (`commandrunner/runner.go:825-877`, `coerceFlagValue` at `:1081-1124`).
A `string_array` flag already produces a `[]string` — the `ids[]` shape Freshchat needs.

**The blocking gap is the bound itself.** The engine's schema dialect
(`internal/connectors/engine/schema.go`) supports only these structural keywords (`schema.go:60-72`):

```
type, required, properties, items, enum, pattern, minProperties,
additionalProperties, x-secret, x-primary-key, x-cursor-field
```

and **unknown keywords are a hard compile error** (`schema.go:104-110`). So a bundle declaring

```json
"ids": { "type": "array", "items": {"type":"string"}, "maxItems": 100 }
```

**fails to load**. There is today no way to express — let alone enforce — "bounded `ids[]` list".
The CLI flag schema has the same hole: `cli_surface.schema.json` flag objects allow only
`name, type, summary, values, maps_to, format, allow_empty, required` — no item bound — so a caller
may pass 10 000 ids to a `string_array` flag and nothing rejects it.

Second gap: a `rest_read` POST body_schema **defaults to `additionalProperties: true`**
(`schema.go:108`). Of the 217 declared `rest_read` POST operations, **169 do not set
`additionalProperties: false`**. For a general read that is the status quo; for a capability whose
entire justification is "bounded and typed", it is exactly the escape hatch the captain banned, and
must be closed by construction for the new kind rather than retrofitted onto 169 existing rows.

Related design authority: **#2985** already carries the captain's decision — provider search/query is
a **separate typed capability**, `pm query` stays warehouse-focused, validation must reject
metadata-only enablement and any raw SQL/GraphQL/HTTP/body escape hatch, and fixture/conformance/docs
tests land before connector fan-out. This program implements that decision; it does not re-open it.

## What this unblocks

**Freshchat (immediate, 3 operations)** — quoted from
`internal/connectors/defs/freshchat/api_surface.json` on `origin/fm/cli-freshchat-parity-wave02-r2`:

| endpoint | model / status | blocking reason (verbatim, abbreviated) |
| --- | --- | --- |
| `POST /files/upload` | `sensitive_reverse_etl` / blocked, risk high | "multipart/form-data with local file input and a documented 25 MB single-file cap; the current Freshchat bundle has no connector-local binary/multipart execution contract without shared binary/file foundation or an approved hook" |
| `POST /images/upload` | `sensitive_reverse_etl` / blocked, risk high | "multipart/form-data with local image input; … no connector-local binary/multipart execution contract without shared binary/file foundation or an approved hook" |
| `POST /users/fetch` | `direct_read` / blocked, risk medium | "a fixed POST-body provider-search operation with ids[] bounded by the provider to 100 users, but executable provider_search/provider_query support is blocked on shared foundation #2985; no raw request body/query escape hatch is exposed" |

Both upload rows state explicitly: *"Represented as a blocked binary/file operation, not an exclusion.
No filesystem path or binary payload is accepted by this connector today."*

**Beyond Freshchat:** 32 declared `file_upload` operations across 4 connectors (table above), and the
provider-search gap the parity audit measured at 310 operations across 3 connectors (#2985).

## Sub-issues

1. **Confine multipart upload file access with `os.Root`** — draft:
   `.planning/phases/foundation-upload-search-r1/issues/SUB-1-bounded-multipart-file-upload.md`
2. **Declared media-type bound for uploads (image upload)** — draft:
   `.planning/phases/foundation-upload-search-r1/issues/SUB-2-upload-media-type-bound.md`
3. **Bounded provider-search operation kind and list bounds** — draft:
   `.planning/phases/foundation-upload-search-r1/issues/SUB-3-bounded-provider-search.md`

## Acceptance criteria

- Every capability stays typed and inspectable. **No caller may supply an arbitrary method, path, or
  body** — not through a flag, not through a config key, not through a spec field.
- No local file path or credential is accepted via argv.
- Local file input is confined with `os.Root`, size-checked before reading, and bounded during the
  read by reading one byte past the limit.
- Bounds are declarable in the bundle and enforced at load time, so an unbounded declaration cannot
  ship.
- No connector bundle under `internal/connectors/defs/**` is edited on this branch.
- `make verify` green; `./cmd/pm` built and the affected surfaces **executed**, not merely read —
  the audit that found 174 commands declaring `availability: implemented` while failing at runtime
  (validator exemption at `cmd/connectorgen/validate.go:1427-1430` vs. runtime enforcement at
  `commandrunner/runner.go:419-421`) is the standing reason this is not optional.

## Safety boundaries

- No provider credentials, no live provider calls, no certification claims.
- No weakening of Connector Guard, the ownership guardrail, or existing block gates.
- `pm query` remains warehouse/materialized-data focused; provider search is a distinct capability.
- PR stays draft; firstmate merges.
