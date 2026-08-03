# TDD ledger — engine-request-shaping-foundations-r1

GSD programming loop, manual/inline fallback (see `PLAN.md` for why). Every behaviour-adding task
below starts **Red** — a test written before the production code, run, and observed failing for the
stated reason — then **Green**.

Commands used throughout:

```
go test ./internal/connectors/engine/ -run <TestName> -count=1
go test ./... -count=1
```

---

## Task 1 — `minItems` / `maxItems` in the engine schema dialect

### Red

Test: `TestCompileSchemaArrayCardinalityKeywords`, `TestSchemaValidateMinItems`,
`TestSchemaValidateMaxItems`, `TestCompileSchemaRejectsInvalidArrayCardinality` in
`internal/connectors/engine/schema_test.go`.

Observed failure before any production change (the dialect rejects the keyword outright):

```
--- FAIL: TestSchemaValidateMinItems
    schema_test.go: CompileSchema: compile schema: unknown keyword "minItems"
--- FAIL: TestSchemaValidateMaxItems
    schema_test.go: CompileSchema: compile schema: unknown keyword "maxItems"
--- FAIL: TestCompileSchemaArrayCardinalityKeywords
    schema_test.go: CompileSchema: compile schema: unknown keyword "minItems"
--- FAIL: TestCompileSchemaRejectsInvalidArrayCardinality
    schema_test.go: negative minItems: want compile error mentioning "minItems", got unknown keyword
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

Test: `TestNewPaginatorStartIndex`, `TestStartIndexPaginatorWalk`,
`TestStartIndexPaginatorStopsOnTotalResults`, `TestStartIndexPaginatorStopsOnEmptyPage`,
`TestStartIndexPaginatorNonAdvancingIndexIsError`, `TestStartIndexPaginatorIgnoresLyingItemsPerPage`
in `internal/connectors/engine/paginate_test.go`.

Observed failure before any production change:

```
--- FAIL: TestNewPaginatorStartIndex
    paginate_test.go: newPaginator: new paginator: unknown pagination type "start_index"
--- FAIL: TestStartIndexPaginatorWalk
    paginate_test.go: newPaginator: new paginator: unknown pagination type "start_index"
```

### Green

`paginate.go`: `startIndexPaginator` added, wired into `newPaginator`'s switch; `bundle.go`:
`PaginationSpec` fields plus load-time validation; `schema/streams.schema.json`: new properties in
both the `base` and per-stream `pagination` blocks.

---

## Task 3 — `required_query` any-of groups

### Red

Test: `TestOperationDirectReadRequiredQueryAnyOf`,
`TestOperationDirectReadRequiredQuerySatisfiedByDeclaredQuery`,
`TestOperationDirectReadRequiredQueryRejectsBlankValue`,
`TestBundleRejectsEmptyRequiredQueryGroup` in `internal/connectors/engine/direct_read_test.go` and
`bundle_test.go`.

Observed failure before any production change (no constraint exists, so the unfiltered request
succeeds where it must fail):

```
--- FAIL: TestOperationDirectReadRequiredQueryAnyOf
    direct_read_test.go: want error requiring one of [email id], got nil (request was issued)
--- FAIL: TestBundleRejectsEmptyRequiredQueryGroup
    bundle_test.go: want load error for empty any_of, got nil
```

### Green

`bundle.go`: `RequiredQuery []RequiredQueryGroup` on `RESTOperationSpec` + load-time validation;
`direct_read.go`: `requireOperationQueryGroups` enforced on the merged query before the request;
`schema/operations.schema.json`: `required_query` declared.

---

## Task 4 — `base64_upload` write body

### Red

Test: `TestWriteBase64UploadFromPath`, `TestWriteBase64UploadRejectsOversizeFile`,
`TestWriteBase64UploadRejectsPathOutsideProject`, `TestWriteBase64UploadStrictBase64Source`,
`TestWriteBase64UploadRejectsNonStrictBase64`, `TestWriteBase64UploadOmitsSourceFieldFromBody`,
`TestWriteBase64UploadRequiresApprovedDigest`, `TestBundleValidatesBase64UploadSpec` in
`internal/connectors/engine/write_test.go` and `bundle_test.go`.

Observed failure before any production change (the body type is unknown, so the action silently
falls through to the default JSON body and transmits the raw local path):

```
--- FAIL: TestWriteBase64UploadFromPath
    write_test.go: body["file"]: want base64 content, got <nil>
--- FAIL: TestWriteBase64UploadOmitsSourceFieldFromBody
    write_test.go: body["file_path"]: want absent, got "payload.txt"   <-- local path on the wire
--- FAIL: TestBundleValidatesBase64UploadSpec
    bundle_test.go: want load error for missing base64_upload block, got nil
```

The second failure is the important one: without this task the "obvious" workaround leaks a local
filesystem path to the provider.

### Green

`bundle.go`: `Base64UploadSpec` on `WriteAction` + `validateWriteBodies` rules;
`write.go`: `buildBase64UploadPayload` with `os.Root` containment, bounded read, strict decode,
digest verification, and source-field removal; `schema/writes.schema.json`: `base64_upload`
declared and `base64_upload` added to the `body_type` enum.

---

## Final gate

```
gofmt -l cmd internal        -> (empty)
go vet ./...                 -> ok
go test ./...                -> ok
go build ./cmd/pm            -> ok
connectorgen validate --json -> 0 findings
connectorgen boundary --json -> clean
```

Executed-not-read evidence for the built binary is recorded in `VERIFICATION.md`.
