# TDD Ledger: Issue #4305

## Planned red/green evidence

| Slice | Red | Green | Edge and regression |
| --- | --- | --- | --- |
| Declared nested body | An installed synthetic command refuses a declaration-backed nested body. | The exact declared route, query, headers, content type, and nested JSON reach the provider double. | Body values cannot populate route/query mappings. |
| Recursive validation | Input is either accepted as opaque raw JSON or rejected only after a request is formed. | The materializer recursively accepts only declared fields and types before request creation. | Unknown, required, malformed, over-depth, over-field, over-item, and over-byte cases have zero I/O. |
| Action reuse and approval | Typed actions have no shared structured-body construction or payload identity. | CLI and typed-action paths share the materializer and confirmation identity. | Separate actions cannot borrow fields; post-preview nested mutation is rejected before I/O. |
| Surface and contract | Generated command surface has no typed representation for declared nested inputs. | Help/schema present declared typed inputs and lack any raw-body bypass. | Scalar/form/SCIM/binary/specialized GitHub focused tests retain outcome. |

## Actual evidence

### 2026-08-20 — declaration-backed REST structured-body red checkpoint

- Red: `go test -timeout 20m ./internal/connectors/commandrunner -run TestBuildOperationDirectWriteCommandSupportsDeclaredStructuredRESTBody -count=1` failed. The production-shaped synthetic command carries a declared path field, query field, scalar body field, closed nested object, and bounded nested array. Runtime refused the first `json` body flag with: `structured JSON variables require a fixed GraphQL operation, got "rest_write"`.
- Observable contract: no caller-provided method, route, content type, or raw body exists in the fixture; the only missing capability is admitting the source-owned nested fields of the declared REST body.
- Next green slice: add an operation-owned structured-body preflight, use it before parsing the JSON flag, then materialize the value only into the declared body path. Engine tests will prove recursive schema and limit failures occur before transport I/O.
