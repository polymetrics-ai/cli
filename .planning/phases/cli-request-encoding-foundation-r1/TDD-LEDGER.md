# TDD ledger — source-backed request encoding foundation

| ID | Guarantee | Required red assertion | Green proof |
| --- | --- | --- | --- |
| E1 | The source-derived cohort is exact | Current importer labels each valid form request with the generic encoding foundation and has no 51-row reconciliation. | Reimported fixture reconciles 51 cited identities: 50 multipart plus one urlencoded, with provider 47/2/1/1 split. |
| E2 | Encoder selection is closed and source-owned | Valid selected form/multipart source cannot become an admissible typed encoder; malformed/mismatch checks are absent. | Only JSON, urlencoded, and multipart source declarations are admitted; unknown/ambiguous/malformed declarations fail with source location. |
| E3 | Schema gaps do not disappear | Removing the generic encoding gap would make a source operation with `oneOf`/dynamic or unsupported part shape look executable. | Encoding disposition clears independently while the exact unrelated source-cited schema foundation remains. |
| E4 | Multipart preserves text/file wire semantics | A direct operation cannot prove exact part names, requiredness, content metadata, file semantics, and byte cap through a transport spy. | Spy observes exact multipart fields/files; missing/unknown/oversize/mismatch calls never hit transport. |
| E5 | Form serialization is provider-declared | Form emission accepts a broad map or silently converts a complex value. | Required/missing/duplicate/unknown checks and declared scalar/repeated serialization produce exact URL-encoded bytes. |
| E6 | Reconciliation is honest | A test could replace citations/identity or count a method-only mutation as a supported lane. | Citation/identity mismatch fails; the report divides credential-bound runnable, residual missing_foundation, and reason totals without method-based classification. |

## Red evidence

Pending implementation. Before production edits, capture focused failures for
`cmd/connectorgen` source import/projection and engine direct-operation request
tests in `traces/`; do not replace these with an exit-status-only test.

## Green evidence

Pending implementation. Record exact focused commands and their observed
transport-spy assertions after each smallest green slice.

