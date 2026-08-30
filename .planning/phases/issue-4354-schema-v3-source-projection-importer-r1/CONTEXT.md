# Schema-v3 source projection/importer foundation — observed context

## Exact current incompatibility

- `origin/main` is `b33983927d863032dac8220949990506e812937d`.
- The Outreach candidate at
  `18248d233e6abd9d7ec03075a225cf35ee2f5399` (also retained by PR #4355)
  contains `outreach-operation-source-lock.json` at `schema_version: 2`, not
  schema v3. Its `rest` object adds `source_kind`, `operation_counts`,
  `supplements`, and per-operation `source_url` fields.
- `parseSourceImportLock` deliberately calls `decodeSourceStrictJSON`. The
  schema-v2 `sourceImportREST` / `sourceImportRESTOperation` contract does not
  own those fields, so canonical source import refuses the retained candidate
  before it can make a descriptor. Treating those fields as arbitrary ignored
  extension data would hide a multi-document provenance model: 253 OpenAPI
  operations and 6 custom-object operations cite two different source hashes.
- A schema-v3 lock can represent that fact safely through ordered
  `rest.source_documents[]`: a complete OpenAPI document and a separate
  rendered-reference document, each with its own artifact/published identity
  and operation citations. The v3 importer, source projection, validator, and
  operation-evidence reader already consume that model when exact retained
  artifacts are available.

## Byte-backed import constraint

The candidate lock records but does not retain the provider bodies:

| Provider document | Locked SHA-256 | Locked bytes | Candidate-tree retention |
| --- | --- | ---: | --- |
| Outreach OpenAPI v2 | `d1f697f6558fda68cd6d8059044e323c20849aeebf303e15c43e0eb9875e2ef6` | 1,384,297 | absent |
| Outreach Custom Objects documentation | `2e74714a933b74cb9a83ddbdb18eeb0b9d045115102ed7465021a45db19e3dda` | 422,602 | absent |

Read-only evidence:

1. `git ls-tree -r --name-only 18248… -- internal/connectors/defs/outreach/sources`
   lists only the lock, disposition, and documentation-audit files. It has no
   `sources/artifacts/` directory or retained-artifacts manifest.
2. The current #4355 head has the same source-only tree. Its documented red
   state is `connectorgen validate …outreach` rejecting the `source_kind`.
3. `git cat-file --batch-all-objects` finds no local Git blob with either
   locked byte count, so the raw source cannot be recovered from local history.

`source-import` must receive the exact provider bytes through the
connector-owned retained-artifact reader. It is prohibited from accepting a
derived `api_surface.json`, streams, fixtures, a hash-only assertion, a
replacement provider response, or a network call. That strict byte-backed
contract remains unchanged.

## Source-reference redesign

The source lock's canonical provider document URLs and operation identities are
still closed declaration evidence. The foundation can add a separate explicit
source-reference/declaration path that maps those citations into projection and
operation evidence without raw bytes. The path must not read the network or
claim that provider bytes have been verified. It must preserve operation/source
identity, classify absent operation-contract detail as
`source_contract_unavailable`, and classify a known but non-executable shape
as `missing_foundation`. It remains invalid for execution-capable
materialization; acquiring the original bytes later can use the existing strict
byte-backed import path without re-pinning.
