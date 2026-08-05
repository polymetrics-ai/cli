# TDD Ledger — foundation-upload-search-r1

Manual-GSD fallback: `scripts/gsd prompt programming-loop init --phase foundation-upload-search-r1
--dry-run` returned `scripts/gsd: unknown GSD command: programming-loop`, so this ledger records the
manual GSD/TDD loop per `PLAN.md`. The planning prompt fallback
(`scripts/gsd prompt plan-phase foundation-upload-search-r1 --skip-research`, 142 lines) was applied
inline.

## Red/green slices

| Slice | Red evidence | Green evidence | Status |
|---|---|---|---|
| Multipart file access is check-then-open | Red: a scratch probe on unmodified `connsdk` uploaded content from outside the intended root. Output below. | `MultipartFile.Root`/`RelPath`; `stat()`/`open()` route every access through `os.Root`. `TestRequesterDoMultipartRefusesEscapingSymlinkSwappedAfterValidation` now refuses it. | Green |
| Containment must survive at the wire-write site | Red: mutant reverting `f, err := file.open()` to `os.Open(file.Path)` in `writeMultipartFile` → `--- FAIL ... DoMultipart error = nil, want refusal of the escaping symlink`. | Restored `file.open()`; test passes. | Green |
| Traversal refused before any request | Red: with an absolute `Path` the earlier code would have opened the escaped file. Test asserts `HTTP calls = 0`. | `TestRequesterDoMultipartRefusesRootRelativeTraversal` passes. | Green |
| Declared media type is unverified | Red: a ZIP uploaded successfully under a declared `image/png`; nothing read the bytes. | `AllowedMediaTypes` + `http.DetectContentType` over the first 512 bytes, sniffed on the existing snapshot copy via `prefixWriter`; `checkAllowedMediaType` refuses before the request. | Green |
| Unclassifiable content needs its own message | Red: would otherwise be reported as an ordinary type mismatch. | `TestRequesterDoMultipartRejectsUnclassifiableContent` pins the distinct "could not be classified" message. | Green |
| Unconstrained upload must stay valid | Red risk: adding the field could have made every part implicitly bounded. | `TestRequesterDoMultipartUnconstrainedMediaTypeStillUploads` keeps absent == unconstrained. | Green |
| Media declaration must be load-validated | Red: an empty/unparseable/non-member declaration loaded fine. | `validateMultipartMediaTypes`; `TestValidateMultipartMediaTypes` covers empty, unparseable, non-member, and non-file-part cases. | Green |
| `maxItems` cannot be expressed | Red: `CompileSchema` on `{"ids":{"type":"array","maxItems":100}}` failed with `compile schema: unknown keyword "maxItems"`. | `maxItems`/`minItems` added to `structuralKeywords` and enforced in array validation. `TestCompileSchemaAcceptsMaxItems`, `TestSchemaEnforcesMaxItems`, `TestSchemaEnforcesMinItems`. | Green |
| Contradictory bounds must not compile | Red: `minItems: 5, maxItems: 2` would have compiled. | `TestCompileSchemaRejectsContradictoryItemBounds`. | Green |
| `provider_search` must be bounded by construction | Red: no such kind; an unbounded/open-bodied declaration had nothing to reject it. | `validateProviderSearchSemantics` + `requireBoundedArrays`. `TestProviderSearchLoadContract` covers 9 refusal cases plus the valid one. | Green |
| `provider_search` must execute | Red: `OperationDirectRead` errored `requires rest_read operation`. | Kind accepted on the bounded read path. `TestOperationDirectReadExecutesProviderSearch` asserts `POST /users/fetch` and the `ids` body against a real `httptest` server. | Green |
| The bound must fire before the wire | Red: a 101-item list would have been sent. | `TestOperationDirectReadRejectsOverlongProviderSearchList` asserts rejection **and** `HTTP calls = 0`. | Green |
| No body escape hatch | Red: an undeclared key could ride along if the root were open. | `additionalProperties: false` is required at the root; `TestOperationDirectReadRejectsUndeclaredProviderSearchBodyKey` pins that a `sql` key is refused. | Green |
| Flag-level list bound | Red: a `string_array` flag accepted unbounded item counts. | `max_items`/`min_items` on `CLIFlag`/`CommandSurfaceFlag`, enforced in `coerceFlagValue`. `TestCoerceFlagValueBoundsStringArrayItems` asserts the error names the flag. | Green |
| Meta-schema regression caught and fixed | Red: adding `"minimum"` to `cli_surface.schema.json` broke **all 550 bundles** — `meta-schemas failed to compile: ... unknown keyword "minimum"`. The meta-schemas are themselves compiled by the engine's restricted dialect. | `"minimum"` removed; the bound is enforced in `cmd/connectorgen/validate.go` instead. `TestNewLoadsDeclarativeBundlesWithHooksAndNativeOverrides` green. | Green |

## Actual evidence

```bash
# RED — scratch probe against unmodified connsdk (TOCTOU)
$ go test ./internal/connectors/connsdk/ -run TestProbeMultipartFollowsEscapingSymlinkSwappedAfterValidation -v
    zz_probe_toctou_test.go:64: DoMultipart err = <nil>
    zz_probe_toctou_test.go:65: bytes that reached the wire = "OUTSIDE-ROOT-SECRET"
    zz_probe_toctou_test.go:67: DEFECT CONFIRMED: content from outside the root reached the wire
--- FAIL: TestProbeMultipartFollowsEscapingSymlinkSwappedAfterValidation (0.12s)

# RED — mutant: revert containment at the wire-write site
$ perl -0pi -e 's/\tf, err := file\.open\(\)/\tf, err := os.Open(file.Path)/' internal/connectors/connsdk/http.go
$ go test ./internal/connectors/connsdk/ -run 'RefusesEscapingSymlink|RefusesRootRelativeTraversal'
--- FAIL: TestRequesterDoMultipartRefusesEscapingSymlinkSwappedAfterValidation (0.00s)
    multipart_bounds_test.go:94: DoMultipart error = nil, want refusal of the escaping symlink

# RED — the 550-bundle regression this phase introduced and fixed
$ go test ./internal/connectors/bundleregistry/ -run TestNewLoadsDeclarativeBundlesWithHooksAndNativeOverrides
load bundle zoom: meta-schemas failed to compile: compile schema: properties.global_flags:
  compile schema: items: compile schema: properties.max_items: compile schema: unknown keyword "minimum"
FAIL

# GREEN — after removing "minimum" from the meta-schema
$ go test ./internal/connectors/bundleregistry/ -run TestNewLoadsDeclarativeBundlesWithHooksAndNativeOverrides -count=1
ok  	polymetrics.ai/internal/connectors/bundleregistry	2.355s

# GREEN — changed packages
$ go test ./internal/connectors/connsdk/ ./internal/connectors/engine/ ./internal/connectors/commandrunner/ ./internal/connectors/ -count=1
ok  	polymetrics.ai/internal/connectors/connsdk	0.519s
ok  	polymetrics.ai/internal/connectors/engine	2.349s
ok  	polymetrics.ai/internal/connectors/commandrunner	4.458s
ok  	polymetrics.ai/internal/connectors	0.615s

$ go test ./internal/connectors/bundleregistry/ ./cmd/connectorgen/ ./internal/cli/... -count=1
ok  	polymetrics.ai/internal/connectors/bundleregistry	3.893s
ok  	polymetrics.ai/cmd/connectorgen	8.914s
ok  	polymetrics.ai/internal/cli	406.153s

$ go vet ./internal/connectors/... ./cmd/connectorgen/
# exit 0
```

## Review-gate round (run 01KZ5G009DG5DXZH0QAF4FTG70)

| Slice | Red evidence | Green evidence | Status |
|---|---|---|---|
| `provider_search` boundedness bypass via multi-form array types | Red: review finding F1 — `requireBoundedArrays` matched only the single string `"type": "array"`, but `compileTypes` (`schema.go:270-288`) also accepts `["array","null"]`, and an items-bearing node with no string type was skipped entirely. An unbounded list could therefore be declared, load, and reach a provider — contradicting this phase's own hard constraint. Confirmed independently by reading `compileTypes` before escalating. | Captain decided FIX. Pipeline commit `55891bed0` adds `isArrayType`, matching the single string form, any type list containing `"array"`, and items-bearing nodes with no string type. Four new `TestProviderSearchLoadContract` cases: multi-form list, items-bearing node, nested multi-form, and deterministic ordering with several unbounded lists. Re-review returned `findings: []` and lowered risk from medium to low. | Green |
| Upload-refusal test was timing-flaky | Red: `TestRequesterDoMultipartRefusesEscapingSymlinkSwappedAfterValidation` failed with `ParseMultipartForm: read tcp ...: use of closed network connection`. A refused upload aborts mid-stream, so the handler legitimately sees a half-written body; the helper called `t.Errorf` on that, turning the behaviour under test into a failure. Whether the handler reached `ParseMultipartForm` was timing-dependent. | `uploadEcho` now returns quietly on an incomplete request and leaves the captured body empty, which is what the assertions check. `-count=30` passes. Verified the relaxation did **not** neuter the test: the wire-write mutant still fails it with `DoMultipart error = nil, want refusal of the escaping symlink`. | Green |

F2 (`http.DetectContentType` recognises only a fixed signature set, so an
unrecognised-but-genuine format sniffs as `application/octet-stream` and is
rejected fail-closed) was classified `no-op`. The captain directed no action: it
is the deliberate tradeoff, and bundle authors must account for Go's sniff table
when declaring `allowed_media_types`.

## Rebase round onto #3700, and the content-type decision

This phase was rebased from `504a7c07a` onto `48e928398` (`origin/main`), which had
advanced by two merged foundation PRs. The rebase dropped a merge commit that a
`required_linear_history` repo would not accept; the branch is now linear on top of
`origin/main` with no merge commits.

`#3700` independently added `minItems`/`maxItems` to the same schema dialect this phase
needed them in. The two implementations were semantically equivalent — both reject
negative bounds and a `maxItems` below `minItems` — so main's factored
`compileArrayCardinality`/`validateArrayCardinality` was kept and this phase's duplicate
removed, with the provider-search motivation folded into the surviving comment. The
`provider_search`-specific `requireBoundedArrays` rule is *not* redundant with it and
stays: the dialect keyword makes a bound *expressible*, while `requireBoundedArrays`
makes it *mandatory* for this kind. Two tests asserted this phase's error wording and
were updated to main's; the behaviour they pin (rejection before any HTTP call) is
unchanged.

`#3700`'s `base64_upload` and this phase's streaming `multipart` upload both read local
files under `os.Root`, but are not duplicates: `base64_upload` buffers a whole payload to
inline it in JSON, while `multipart` holds the root open across a streamed snapshot so
containment is re-checked at every open. Both survived intact.

| Slice | Red evidence | Green evidence | Status |
|---|---|---|---|
| Part header asserted a type we had just disproved | Red: `TestRequesterDoMultipartSendsSniffedContentTypeWhenBounded` — a JPEG file declared `content_type: image/png` under an allowlist admitting both was uploaded with `Content-Type: image/png`. Confirmed as a real red by mutating the fix to `if false`: `part Content-Type = "image/png", want the sniffed image/jpeg`. | Captain decided option A: **the wire header is set from the sniffed type; the allowlist stays the restriction, and a single-entry allowlist is the documented way to demand exactly one type.** `snapshotApprovedMultipartFiles` overwrites the prepared part's `ContentType` with the sniffed value when — and only when — an allowlist made that sniff binding. | Green |
| Overriding must not damage the unbounded case | Red risk: `http.DetectContentType` is coarse (every CSV sniffs as `text/plain`), so overriding unconditionally would have downgraded a deliberate `content_type`. | `TestRequesterDoMultipartKeepsDeclaredContentTypeWhenUnbounded` pins that a declared `text/csv` survives untouched with no allowlist present. | Green |

## Mutant that survived, and why

Reverting `file.stat()` to `os.Stat(file.Path)` in `validateMultipartForm` alone does **not** fail the
traversal test. That is not a coverage gap: `snapshotApprovedMultipartFiles` performs a second
confined `stat()`, so the traversal is still refused before any request and the asserted behaviour
(`HTTP calls = 0`) holds. The two confined stats are defense in depth; which one catches a given
escape is an implementation detail the test deliberately does not pin.

## Verification by execution, not by reading

The standing reason is the audit that found 174 commands declaring `availability: implemented` while
failing at runtime — the validator exempts operation-backed direct reads
(`cmd/connectorgen/validate.go:1427-1430`) from a check the runtime enforces
(`commandrunner/runner.go:419-421`). So every capability claim here is backed by a command that ran:

```bash
$ go build -o pm ./cmd/pm && ./pm gong calls upload-media --help
NAME          pm gong calls upload-media - Add call media (/v2/calls/{id}/media)
AVAILABILITY  implemented
NOTES         Uses typed multipart write support; no generic upload command is exposed.
FLAGS
  --id (string): ... maps_to=record.id
  --media-file-path (string): Project-relative media file path to upload. maps_to=record.media_file_path
```

That command is what established the finding in `PLAN.md`: bounded multipart upload already ships, so
this phase hardens it rather than rebuilding it. The `provider_search` executor and both containment
paths are exercised against real `httptest` servers and real temp files, not mocks of the transport.
