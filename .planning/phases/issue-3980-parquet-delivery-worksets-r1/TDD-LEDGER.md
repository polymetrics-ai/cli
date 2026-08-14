# TDD ledger — Issue 3980: immutable Parquet delivery worksets

Manual inline GSD TDD execution. Red and green command output is retained in
`traces/` before and after production changes.

| ID | Guarantee | Red assertion | Green proof |
| --- | --- | --- | --- |
| R1 | StreamID-keyed immutable destination identity | No workset can derive the #3981 ledger key plus schema/key/source/baseline bindings. | Two derivations from identical Parquet inputs produce byte-identical identity; changed destination, schema, or ordered key binding changes it. |
| R2 | Immutable output snapshot | A returned value can retain a mutable source path or caller data. | After source Parquet replacement, reopening the original workset yields the original projection/delta and unchanged content hash. |
| R3 | Keyed real-Parquet delta | New/changed/unchanged source rows cannot be distinguished against the destination baseline. | DuckDB fixture asserts exactly insert/update rows in delta and no unchanged rows across composite keys, null payloads, and type differences. |
| R4 | Explicit deletes only | Snapshot absence can become a delete artifact. | A physically absent baseline row produces zero tombstones; supplied validated tombstones are preserved exactly and deterministically. |
| R5 | No baseline advancement before receipt | Derivation can overwrite the input baseline or a target-delivery artifact. | Input baseline SHA-256 is identical before/after every success/refusal; candidate baseline is a separate immutable output and no write-session/ledger fake call occurs. |
| R6 | Bounded cleanup/refusal | Cancellation, duplicate/null keys, or a corrupt reuse candidate can leave a publishable partial workset. | Refusal/cancellation asserts no final workset path, no temp directory, and zero baseline-byte changes. |

## Red command

```sh
go test -timeout 20m ./internal/connectors/database -run 'TestDeriveChangeDeliveryWorksetImmutableIdentity' -count=1
```

Before production code, the new live-Parquet test must fail because
`DeriveChangeDeliveryWorkset` and its sealed request/workset contract do not
exist. Preserve the compiler/test failure in `traces/workset-identity-red.txt`.

## Green commands

```sh
go test -timeout 20m ./internal/connectors/database -run 'TestDeriveChangeDeliveryWorkset' -count=1
go test -timeout 20m ./internal/warehouse/... ./internal/connectors/database/... ./internal/synctransport/... -count=1
go test -race -timeout 20m ./internal/connectors/database -run 'TestDeriveChangeDeliveryWorkset' -count=1
```

The green proof must report the observable output rows, tombstones, workset
identity/content hash, immutable source-replacement result, and zero-publish
refusal paths.
