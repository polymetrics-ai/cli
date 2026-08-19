# TDD Ledger: Issue #4305

## Planned red/green evidence

| Slice | Red | Green | Edge and regression |
| --- | --- | --- | --- |
| Declared nested body | An installed synthetic command refuses a declaration-backed nested body. | The exact declared route, query, headers, content type, and nested JSON reach the provider double. | Body values cannot populate route/query mappings. |
| Recursive validation | Input is either accepted as opaque raw JSON or rejected only after a request is formed. | The materializer recursively accepts only declared fields and types before request creation. | Unknown, required, malformed, over-depth, over-field, over-item, and over-byte cases have zero I/O. |
| Action reuse and approval | Typed actions have no shared structured-body construction or payload identity. | CLI and typed-action paths share the materializer and confirmation identity. | Separate actions cannot borrow fields; post-preview nested mutation is rejected before I/O. |
| Surface and contract | Generated command surface has no typed representation for declared nested inputs. | Help/schema present declared typed inputs and lack any raw-body bypass. | Scalar/form/SCIM/binary/specialized GitHub focused tests retain outcome. |

## Actual evidence

Pending execution. Each red command, resulting failure, green command, and outcome will be recorded here as work proceeds.
