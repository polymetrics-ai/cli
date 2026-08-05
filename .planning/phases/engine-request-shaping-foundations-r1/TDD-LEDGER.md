# TDD ledger — engine-request-shaping-foundations-r1

GSD programming loop, manual/inline fallback (see `PLAN.md` for why). Every behaviour-adding task
below starts **Red** — a test written before the production code, run, and observed failing for the
stated reason — then **Green**.

Commands used throughout:

```
go test ./internal/connectors/engine/ -run <TestName> -count=1
```

Package-group runs (never a bare `go test ./...`) are listed in the Final gate below.

---

## Task 1 — `minItems` / `maxItems` in the engine schema dialect

### Red

Tests in `internal/connectors/engine/schema_test.go`:
`TestCompileSchemaArrayCardinalityKeywords`, `TestSchemaValidateMinItems`,
`TestSchemaValidateMaxItems`, `TestSchemaArrayCardinalityIgnoresNonArrays`,
`TestSchemaMinItemsZeroIsHonored`, `TestCompileSchemaRejectsInvalidArrayCardinality`.

Observed failure before any production change (the dialect rejects the keyword outright):

```
--- FAIL: TestCompileSchemaArrayCardinalityKeywords (0.00s)
    schema_test.go:364: CompileSchema({"type":"array","minItems":1}): unexpected error: compile schema: unknown keyword "minItems"
--- FAIL: TestSchemaValidateMinItems (0.00s)
    schema_test.go:372: CompileSchema: compile schema: properties.ids: compile schema: unknown keyword "minItems"
--- FAIL: TestSchemaValidateMaxItems (0.00s)
    schema_test.go:401: CompileSchema: compile schema: unknown keyword "maxItems"
--- FAIL: TestSchemaArrayCardinalityIgnoresNonArrays (0.00s)
    schema_test.go:418: CompileSchema: compile schema: unknown keyword "minItems"
--- FAIL: TestSchemaMinItemsZeroIsHonored (0.00s)
    schema_test.go:430: CompileSchema: compile schema: unknown keyword "minItems"
--- FAIL: TestCompileSchemaRejectsInvalidArrayCardinality/maxItems_below_minItems (0.00s)
    schema_test.go:456: error should mention "maxItems", got compile schema: unknown keyword "minItems"
```

This is the exact failure the Airtable ledger predicts: a bundle cannot even *load* with `minItems`
declared, which is why 25 operations could not be declared executable.

### Green

`schema.go`: `minItems`/`maxItems` added to `structuralKeywords`, compiled into `schemaNode` with
`hasMinItems`/`hasMaxItems`, enforced in `validate` only against array instances, with compile-time
rejection of negative bounds and of `maxItems < minItems`.

### Reach test

`TestWriteRecordSchemaMinItemsRejectsEmptyArray` (write `record_schema`),
`TestWriteJSONArrayBodySchemaMinItems` (`json_array` `body_schema`) and
`TestOperationDirectReadBodySchemaMinItems` (operation `rest.body_schema`) prove the one dialect
change reaches all three request-building paths without any per-site edit — the property that turns
one rule into 25 unblocked operations.

---

## Task 2 — `start_index` pagination strategy

### Red

Tests in `internal/connectors/engine/paginate_test.go`:
`TestNewPaginatorStartIndexDefaultsToSCIMNames`, `TestStartIndexPaginatorWalksEveryPageOnce`,
`TestStartIndexPaginatorStopsAtTotalResults`, `TestStartIndexPaginatorStopsOnEmptyPage`,
`TestStartIndexPaginatorIgnoresLyingItemsPerPage`,
`TestStartIndexPaginatorNonAdvancingIndexIsStickyError`,
`TestStartIndexPaginatorHonorsDeclaredParamNames`, `TestNewPaginatorStartIndexRejectsNegativeBase`,
`TestStartIndexPaginatorWithoutPageSizeIssuesOnePage`.

Red arrived in two stages. First a compile failure — the spec fields did not exist:

```
paginate_test.go:1257:75: unknown field StartIndexBase in struct literal of type PaginationSpec
paginate_test.go:1280:3:  unknown field StartIndexParam in struct literal of type PaginationSpec
paginate_test.go:1281:3:  unknown field CountParam in struct literal of type PaginationSpec
```

Then, once `PaginationSpec` carried the fields, the behavioural red:

```
--- FAIL: TestNewPaginatorStartIndexDefaultsToSCIMNames (0.00s)
    paginate_test.go:1106: newPaginator() error = new paginator: unknown pagination type "start_index"
--- FAIL: TestStartIndexPaginatorWalksEveryPageOnce (0.00s)
    paginate_test.go:1142: newPaginator() error = new paginator: unknown pagination type "start_index"
    (… and five more, all the same cause)
```

### Green

`paginate.go`: `startIndexPaginator` added, wired into `newPaginator`'s switch; `bundle.go`:
`PaginationSpec` fields plus load-time validation; `schema/streams.schema.json`: new properties in
both the `base` and per-stream `pagination` blocks.

---

## Task 3 — `required_query` any-of groups

### Red

Tests in `internal/connectors/engine/direct_read_test.go`:
`TestOperationDirectReadRequiredQueryAnyOf`,
`TestOperationDirectReadRequiredQueryRejectsBlankValue`,
`TestOperationDirectReadRequiredQuerySatisfiedByDeclaredQuery`,
`TestOperationDirectReadRequiredQueryEveryGroupMustBeSatisfied`; in `bundle_test.go`:
`TestBundleLoadAcceptsRequiredQueryGroups`, `TestBundleLoadRejectsUnenforceableRequiredQuery`.

Observed failure before any production change — no constraint exists, so the unfiltered request
reaches the provider where it must not:

```
--- FAIL: TestOperationDirectReadRequiredQueryAnyOf (0.09s)
    direct_read_test.go:750: unfiltered request: want error, got nil
--- FAIL: TestOperationDirectReadRequiredQueryRejectsBlankValue (0.00s)
    direct_read_test.go:785: blank value: want error, got nil
--- FAIL: TestOperationDirectReadRequiredQueryEveryGroupMustBeSatisfied (0.00s)
    direct_read_test.go:839: second group unsatisfied: want error, got nil
```

Then, after the runtime check landed, the load-side red:

```
--- FAIL: TestBundleLoadAcceptsRequiredQueryGroups (0.02s)
    bundle_test.go:2118: Load: … /operations/0/rest/required_query: additional property not allowed
```

### Green

`bundle.go`: `RequiredQuery []RequiredQueryGroup` on `RESTOperationSpec` + load-time validation;
`direct_read.go`: `requireOperationQueryGroups` enforced on the merged query before the request;
`schema/operations.schema.json`: `required_query` declared.

---

## Task 4 — `base64_upload` write body

### Red

Tests in `internal/connectors/engine/write_test.go`:
`TestWriteBase64UploadEncodesFileAndOmitsSourceField`, `TestWriteBase64UploadRejectsOversizePayload`,
`TestWriteBase64UploadRejectsPathEscape`, `TestWriteBase64UploadStrictSourceRejectsSloppyEncoding`,
`TestWriteBase64UploadEnforcesEncodedBound`, `TestWriteBase64UploadHonorsApprovedPayloadDigest`,
`TestDryRunBase64UploadDoesNotReadTheFile`; in `bundle_test.go`:
`TestBundleLoadAcceptsBase64UploadAction`, `TestBundleLoadRejectsInvalidBase64UploadAction`.

Observed failure before any production change — the body type is unknown, so the action falls
through to the default JSON body:

```
--- FAIL: TestWriteBase64UploadEncodesFileAndOmitsSourceField (0.02s)
    write_test.go:1227: body[file] = "", want "aGVsbG8gYXR0YWNobWVudA=="
--- FAIL: TestWriteBase64UploadRejectsOversizePayload (0.01s)
    write_test.go:1258: oversize payload: want error, got nil
--- FAIL: TestWriteBase64UploadRejectsPathEscape (0.01s)
    write_test.go:1296: "../escape.txt": want containment error, got nil
--- FAIL: TestWriteBase64UploadStrictSourceRejectsSloppyEncoding (0.01s)
    write_test.go:1332: "aGVsbG8": want strict-base64 rejection, got nil
--- FAIL: TestWriteBase64UploadEnforcesEncodedBound (0.02s)
    write_test.go:1372: over encoded bound: want error, got nil
--- FAIL: TestWriteBase64UploadHonorsApprovedPayloadDigest (0.04s)
    write_test.go:1415: substituted payload: want digest mismatch error, got nil
```

The path-escape and digest failures are the important ones. Without this task the "obvious"
workaround — carrying the payload as an ordinary JSON string field — would transmit a local
filesystem path unchecked and bind nothing to the approved batch.

### Green

`bundle.go`: `Base64UploadSpec` on `WriteAction` + `validateWriteBodies` rules;
`write.go`: `buildBase64UploadPayload` with `os.Root` containment, bounded read, strict decode,
digest verification, and source-field removal; `schema/writes.schema.json`: `base64_upload`
declared and `base64_upload` added to the `body_type` enum.

---

## Final gate

```
gofmt -l cmd internal        -> (empty)
go vet ./...                 -> clean
go build ./cmd/pm            -> ok
make connectorgen-validate   -> 550 connectors checked, 0 findings
make connector-boundary      -> outcome clean, 0 findings, 550 connectors loaded
pnpm run gen:website-data    -> regenerated, zero drift
```

Tests were run per package group rather than as one `go test ./...`: `internal/cli` (444s) and
`internal/connectors/certify` (557s) each exceed a single tool call, and a bare `go test` also hits
Go's own 10-minute per-package default. All groups green — engine, conformance, commandrunner,
connsdk, safety, cmd/..., connectors, app, cli, certify. Full matrix in `VERIFICATION.md` §1.

Executed-not-read evidence — a scratch bundle validated by the **built** `connectorgen`, the built
`pm` still loading 554 connectors, and twelve runtime checks driven against live HTTP servers — is
recorded in `VERIFICATION.md` §2 and §3.

## One correction worth recording

`TestWriteBase64UploadStrictSourceRejectsSloppyEncoding` failed green-side on the newline case:

```
--- FAIL: TestWriteBase64UploadStrictSourceRejectsSloppyEncoding
    write_test.go:1332: "aGVs\nbG8=": want strict-base64 rejection, got nil
```

Go's base64 decoder skips `\r` and `\n` **unconditionally** — `Strict()` only enforces canonical
trailing padding bits, not a canonical alphabet. Relying on `Strict()` alone would have accepted a
MIME-wrapped payload and silently re-encoded it into something the operator never wrote. Fixed by
checking the alphabet explicitly before decoding (`requireCanonicalBase64Alphabet`). The test caught
this; reading the stdlib documentation would not have.
