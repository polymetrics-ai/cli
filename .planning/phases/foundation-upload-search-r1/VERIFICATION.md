# Verification — foundation-upload-search-r1

Issue tree: parent #3694; sub-issues #3695 (containment), #3696 (media-type bound),
#3697 (bounded provider_search). Created under `karthik-sivadas` per the captain's bounded
identity exception (issue creation only; approvals and merges remain the captain's, and firstmate
owns every merge). No `alfred-polymetrics-ai` credential exists in this environment.

## What shipped, against what each sub-issue asked for

### #3695 — confine multipart upload file access with `os.Root`

- `MultipartFile` carries `Root *os.Root` + `RelPath`; `stat()` and `open()` route **every** access
  through the root. The three previously unconfined sites — the pre-read stat, the snapshot copy, and
  the wire write — now all re-resolve under containment instead of trusting one earlier check.
- `engine/write.go` opens the root at the project dir (`openMultipartRoot`), closes it when the
  request finishes, and hands connsdk a root-relative path
  (`multipartRootRelativePath`) instead of a pre-resolved absolute path.
- The lexical `safety.ValidateLocalWritePath` check is kept ahead of it as a cheap first filter with
  a clearer message, but it is no longer load-bearing for containment.
- Absolute paths inside the project directory remain accepted, as before; the project dir is compared
  both as given and with symlinks resolved, because a configured `/var/...` root and an absolute
  `/private/var/...` record value denote the same directory on macOS.
- After a snapshot exists the root handle is cleared, so later opens target the verified copy — which
  is what makes the digest and media-type checks binding rather than advisory.

Verified: symlink-swap-after-validation refused; root-relative traversal refused with zero HTTP
calls; nested happy path still uploads; a mutant reverting the wire-write site fails the suite.

### #3696 — declared media-type bound

- `MultipartPartSpec.AllowedMediaTypes` (bundle-declared) → `MultipartFile.AllowedMediaTypes`.
- Sniffed with `http.DetectContentType` over the first 512 bytes, captured by `prefixWriter` inside
  the **existing** snapshot `io.MultiWriter` — no extra read, and no second pass that could race the
  first.
- The snapshot gate widened from "has an approved digest" to "has a digest **or** a media bound", so
  a part can be type-bounded without also requiring digest approval. Parts with neither take the
  existing no-snapshot path unchanged.
- Upload **fails closed** on mismatch, deliberately unlike the download direction: there the provider
  makes the Content-Type claim and providers misreport it, so a mismatch is surfaced rather than
  rejected; here we are the party making the claim, so an unsatisfiable claim is our bug.
- The part's **wire `Content-Type` is set from the sniffed type** when an allowlist is present
  (captain's decision, option A). The sniffed type has just been proven to be one the bundle accepts,
  so the claim we make is one we verified rather than one we merely declared; asserting the declared
  `content_type` over disagreeing bytes was a claim we had already falsified. `allowed_media_types`
  remains the only restriction, so **a single-entry list is how a bundle demands exactly one type**.
  With no allowlist the declared `content_type` is sent untouched, because nothing verified the sniff
  and `http.DetectContentType` is coarse enough (every CSV sniffs as `text/plain`) that overriding
  would lose author intent.
- Unclassifiable content gets its own message rather than being reported as an ordinary mismatch.
- Load-time validation: entries must parse; a present-but-empty list is refused so "bounded" and
  "unbounded" can never be confused; a declared `content_type` must be a member of its own allowlist;
  the field is refused on non-file parts.

### #3697 — bounded `provider_search`

- `maxItems`/`minItems` added to the engine's schema dialect and enforced. Before this, the bound
  Freshchat needs (`ids[]` ≤ 100) could not be *declared* — unknown keywords are a compile error, so
  the bundle failed to load outright.
- New `provider_search` operation kind, validated at load: POST only; connector-relative path;
  `application/json`; positive `max_bytes`; non-mutating; `body_schema` required with
  `additionalProperties: false` at the root; and **every array property must declare `maxItems`**,
  walked recursively. An unbounded list cannot ship.
- Executes through the existing bounded read path, reusing its response bounding, clamping,
  redaction, and output-policy handling rather than adding a parallel back half.
- `max_items`/`min_items` on `string_array` flags as a second, independent bound that can name the
  flag the user typed.

## Commands run, with actual results

```bash
$ go vet ./internal/connectors/... ./cmd/connectorgen/
# exit 0

$ go test ./internal/connectors/connsdk/ ./internal/connectors/engine/ \
         ./internal/connectors/commandrunner/ ./internal/connectors/ -count=1
ok  	polymetrics.ai/internal/connectors/connsdk	0.519s
ok  	polymetrics.ai/internal/connectors/engine	2.349s
ok  	polymetrics.ai/internal/connectors/commandrunner	4.458s
ok  	polymetrics.ai/internal/connectors	0.615s

$ go test ./internal/connectors/bundleregistry/ ./cmd/connectorgen/ ./internal/cli/... -count=1
ok  	polymetrics.ai/internal/connectors/bundleregistry	3.893s
ok  	polymetrics.ai/cmd/connectorgen	8.914s
ok  	polymetrics.ai/internal/cli	406.153s

$ go build ./cmd/pm
# exit 0

$ ./pm gong calls upload-media --help
AVAILABILITY  implemented
NOTES         Uses typed multipart write support; no generic upload command is exposed.

$ go run ./cmd/connectorgen boundary . --json
# exit 0  (this is the gate .github/workflows/connector-boundary.yml runs)
```

Local suite scoping: a single `go test ./...` does not fit inside one 10-minute tool call across
550+ connectors, so runs are scoped to changed packages plus `internal/cli`; CI carries the full
suite. `internal/connectors/certify` is run separately with an extended Go timeout for the same
reason.

## Path-ownership guardrail

**Zero files under `internal/connectors/defs/**` changed.** Confirmed:

```bash
$ { git diff --name-only origin/main...HEAD; git status --porcelain | awk '{print $2}'; } \
    | grep -c 'internal/connectors/defs/'
0
```

`go run ./cmd/connectorgen ownership .` reports `ownership_scope_missing` — "changed paths require
exactly one connector target, but none was declared or inferred", whose own remediation is "move
shared changes to a foundation PR". That is the expected result for a foundation PR: the command is
the changed-path guard for connector *implementation* lanes, and this branch has no connector target
by design. The gate CI actually runs for this workflow is `connectorgen boundary`, which passes.

## Invariants held

- **No raw request escape hatch.** The method is fixed by the operation kind or the write action, the
  path by the bundle, and the body by a closed schema whose every list is bounded. Every field added
  here is bundle-authored; none is caller-supplied. `TestOperationDirectReadRejectsUndeclaredProviderSearchBodyKey`
  pins that a caller cannot smuggle in an undeclared key.
- **No file path or credential via argv.** The multipart path stays record-driven, exactly as before.
- **No existing bundle stops loading.** Every new declaration field is optional and every new
  constraint fires only for the new kind or when the new field is present — verified by loading all
  550 bundles.
- **No block gate, Connector Guard rule, or ownership guardrail weakened.**

## Deliberately not done

- **No second upload executor.** Bounded multipart upload already ships and is CLI-reachable
  (`pm gong calls upload-media`, `availability: implemented`), established by execution before any
  code was written. Building a parallel one is the duplication this phase was told to avoid.
- **The 32 `file_upload` operation-kind declarations remain dead** (xero 22, zendesk-support 5,
  bitbucket 4, asana 1). They are load-validated but have no executor, while the executable contract
  lives in `writes.json`. Reconciling them requires connector-bundle edits, which the ownership
  guardrail forbids here. **Follow-up for the connector lanes.**
- **The 169 `rest_read` POST operations without `additionalProperties: false` are untouched** — an
  unrelated migration that must not ride a foundation PR.

## Rebase onto #3700

The branch was rebased from `504a7c07a` onto `48e928398`, dropping a merge commit this repo's
`required_linear_history` setting would reject. Verified after the rebase:

```bash
$ git merge-base --is-ancestor origin/main HEAD && echo linear
linear

$ git log --merges origin/main..HEAD --oneline | wc -l
0
```

Conflict resolutions, on the merits rather than by picking a side:

- **`minItems`/`maxItems` (schema.go)** — both PRs added the same dialect keywords. Main's factored
  `compileArrayCardinality`/`validateArrayCardinality` kept, this phase's duplicate dropped, the
  provider-search motivation folded into the surviving comment. `requireBoundedArrays` is *not*
  redundant with it and stays: the keyword makes a bound expressible, that rule makes it mandatory
  for `provider_search`.
- **`bundle.go`** — pure adjacency; `validateBase64UploadSpec` (#3700) and
  `validateProviderSearchSemantics`/`requireBoundedArrays`/`validateMultipartMediaTypes` (this phase)
  are unrelated and all survive.
- **`write.go`** — #3700's `base64_upload` builders kept whole; `buildMultipartPayload` takes this
  phase's `root *os.Root` parameter. The two upload paths are complementary, not duplicates:
  `base64_upload` buffers a payload to inline it in JSON, `multipart` streams under a held-open root.
- **Two test assertions** updated from this phase's error wording to main's. The behaviour they pin —
  rejection before any HTTP call — is unchanged.

Re-verified by execution after the rebase, not by reading:

```bash
$ go build -o pm ./cmd/pm && ./pm connectors list --json | jq length
554

$ ./pm gong calls upload-media --help
NAME
  pm gong calls upload-media - Add call media (/v2/calls/{id}/media)
# exit 0

$ go run ./cmd/connectorgen boundary . --json
# exit 0

$ { git diff --name-only origin/main...HEAD; git status --porcelain | awk '{print $2}'; } \
    | grep -c 'internal/connectors/defs/'
0
```

## Website catalog

`.github/workflows/website.yml` triggers only on `website/**`,
`internal/connectors/icon_data.json`, `docs/connectors/icons/**`, `.github/workflows/website.yml`,
and `.gitlab-ci.yml`. This branch touches none of them, so `Website checks` does not run; the
connector catalog is generated from `internal/connectors/defs/**`, which is unchanged.

Confirmed by regenerating rather than by assuming — the generator reproduces the committed
catalog byte-for-byte, so there is nothing to commit:

```bash
$ node website/scripts/gen-connector-catalog.mjs
Wrote 550 connectors to lib/connectors.catalog.generated.ts and
lib/connectors.catalog.data.generated.json (11413 KB).

$ git status --porcelain website/
# (no output)
```

## What Freshchat can now do, in its own lane

| Freshchat endpoint | was blocked on | now |
|---|---|---|
| `POST /files/upload` | "no connector-local binary/multipart execution contract" | adoptable as a typed `writes.json` multipart action; file access is now soundly confined |
| `POST /images/upload` | same, for image input | same, plus `allowed_media_types` to make the image contract enforceable |
| `POST /users/fetch` | "executable provider_search/provider_query support is blocked on shared foundation #2985" | declarable as a `provider_search` operation with `ids[] maxItems: 100` |
