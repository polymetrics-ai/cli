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
