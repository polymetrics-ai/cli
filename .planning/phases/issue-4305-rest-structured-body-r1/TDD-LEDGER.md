# TDD Ledger: Issue #4305

## Planned red/green evidence

| Slice | Red | Green | Edge and regression |
| --- | --- | --- | --- |
| Declared nested body | An installed synthetic command refuses a declaration-backed nested body. | The exact declared route, query, headers, content type, and nested JSON reach the provider double. | Body values cannot populate route/query mappings. |
| Recursive validation | Input is either accepted as opaque raw JSON or rejected only after a request is formed. | The materializer recursively accepts only declared fields and types before request creation. | Unknown, required, malformed, over-depth, over-field, over-item, and over-byte cases have zero I/O. |
| Action reuse and approval | Typed actions have no shared structured-body construction or payload identity. | CLI and typed-action paths share the materializer and confirmation identity. | Separate actions cannot borrow fields; post-preview nested mutation is rejected before I/O. |
| Surface and contract | Generated command surface has no typed representation for declared nested inputs. | Help/schema present declared typed inputs and lack any raw-body bypass. | Scalar/form/SCIM/binary/specialized GitHub focused tests retain outcome. |
| Write query absence | `omit_when_absent` does not tolerate an absent `record.*` query value. | Only an object-form, source-locked query declaration can omit its own missing `record.*` value. | Required/undeclared/wrong-source/malformed values fail before I/O; config, secret, incremental, and explicit record values retain their established behavior. |

## Actual evidence

### 2026-08-20 — declaration-backed REST structured-body red checkpoint

- Red: `go test -timeout 20m ./internal/connectors/commandrunner -run TestBuildOperationDirectWriteCommandSupportsDeclaredStructuredRESTBody -count=1` failed. The production-shaped synthetic command carries a declared path field, query field, scalar body field, closed nested object, and bounded nested array. Runtime refused the first `json` body flag with: `structured JSON variables require a fixed GraphQL operation, got "rest_write"`.
- Observable contract: no caller-provided method, route, content type, or raw body exists in the fixture; the only missing capability is admitting the source-owned nested fields of the declared REST body.
- Next green slice: add an operation-owned structured-body preflight, use it before parsing the JSON flag, then materialize the value only into the declared body path. Engine tests will prove recursive schema and limit failures occur before transport I/O.

### 2026-08-20 — write-query record-absence red checkpoint

- Red: `go test -timeout 20m ./internal/connectors/engine -run 'Test(BuildWriteQueryOmitWhenAbsentScopesMissingRecordValuesToTheirDeclaredQuery|WriteActionRecordQueryRejectionsHappenBeforeProviderIO)' -count=1` failed in `missing optional record value is omitted` with `interpolate: unresolved key "optional" in record`.
- Observable contract: the existing shared query resolver deliberately omits only absent config, secret, or incremental values. A write query's exact object-form declaration must additionally, and only additionally, omit an absent `record.*` reference.
- Next green slice: keep the stream/check resolver unchanged; add a write-only resolver that recognizes typed unresolved record errors only under `omit_when_absent`, retaining all other pre-I/O failures.

### 2026-08-20 — declaration-backed REST structured-body green checkpoint

- Green: `go test -timeout 20m ./internal/connectors/engine ./internal/connectors/commandrunner ./cmd/connectorgen ./internal/cli -run 'Test(OperationDirectWriteStructuredRESTBody|OperationStructuredJSONBodyPreflight|BuildOperationDirectWriteCommandSupportsDeclaredStructuredRESTBody|GitHubUserDraftCommandBuildsFixedGraphQLMutation|Validate_CLISurfaceDirectWriteStructuredRESTBody|StructuredRESTBodyCommandHelpAndManual|ConnectorsManualDocumentsConnectorArchitectureAndGithubExamples)' -count=1` passed.
- The runtime admits only a declared top-level object/array REST body field after validating its bounded closed schema, and `operationWriteBody` repeats that exact declaration validation for typed actions. The provider fixture proves exact route/query/header/content-type/body, action isolation, approval digest binding, and zero I/O for rejected inputs.
- Existing GraphQL preflight remains the direct-read behavior, and the focused specialized GitHub assertion remains green.

### 2026-08-20 — write-query record-absence green checkpoint

- Green: `go test -timeout 20m ./internal/connectors/engine -run 'Test(BuildWriteQueryOmitWhenAbsentScopesMissingRecordValuesToTheirDeclaredQuery|WriteActionRecordQueryRejectionsHappenBeforeProviderIO|WriteActionOptionalRecordQueryIsOmittedOrPreservedAtProvider|WriteActionQuery)' -count=1` passed.
- `buildWriteQuery` now uses a write-only resolver. It omits only an unresolved `record.*` reference whose exact object-form `QueryParam` sets `omit_when_absent`; plain/required entries, an undeclared record key attempting to populate a declared query, wrong sources, malformed values, and all other errors reject before provider I/O.
- Config, secrets, and incremental omission/default behavior stays in the shared resolver and is regression-tested unchanged. An explicit `record.*` value is transmitted through the declared parameter unchanged.
