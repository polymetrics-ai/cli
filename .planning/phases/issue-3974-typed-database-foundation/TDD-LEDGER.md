# TDD ledger — Issue #3974: typed database connector foundation

Manual inline GSD TDD execution. The first failing focused run is retained in
`traces/red-run.txt` before production implementation is added.

| ID | Guarantee | Red assertion | Green proof |
| --- | --- | --- | --- |
| R1 | Strict definition | `database.json` accepts unknown fields, invalid schema version, unknown logical kind, or a secret-looking unknown value. | Loader and schema reject all cases with safe errors that contain no submitted value. |
| R2 | Immutable projections | A caller can mutate mappings, modes, catalog relations, keys, or nested array types returned from a public projection. | Definitions, catalog values, read plans, and nested logical types return independent copies. |
| R3 | Bounded resources | A zero, unlimited, negative, or above-hard-limit page/batch/pool/timeout/parameter policy is accepted. | Every resource is finite, positive, defaulted within its maximum, and explicit overrides fail outside that range. |
| R4 | Exact types | Unknown types, precision loss, timezone loss, stringification, or opaque-native mapping can be admitted. | Only exact/lossless plans compile; text/binary/JSON/array/timezone changes and opaque types refuse. |
| R5 | Structured identity/catalog | Source/target refs drop the owner triple or unsafe relation/key identifiers become a stringly SQL surface. | Typed refs retain workspace/connector/connection, catalog keys/columns are relation-bound, and fingerprints are stable. |
| R6 | Stable read plan | A plan can use no order, a non-unique suffix, an unbounded page size, or mutable catalog data. | Read plan construction requires a declared unique-key order suffix, finite page limit, fingerprint, and defensive copies. |
| R7 | Admission is real | A valid declaration passes without a registered driver/admission, or pairs with a different protocol/API version. | Registry requires exact registered driver and shared native admission/evidence; source execution additionally requires a runner. |
| R8 | Cancellation | Load, admission, or plan construction ignores an already-cancelled context. | Each boundary returns the context error before work / registry exposure. |
| R9 | PostgreSQL seam | The native PostgreSQL package can drift from the foundation's driver contract unnoticed. | Compile-time assertion and a definition-load test establish the reference-driver seam without registering an operation. |
| R10 | Truthful metadata | Foundation work accidentally promotes PostgreSQL write or CDC. | Focused test reads the loaded PostgreSQL bundle and asserts both flags remain false. |

## Red command

```sh
go test ./internal/connectors/database ./internal/synccontract ./internal/connectors/native/postgres -run 'Test(Database|NativeAdmission|PostgresDatabase)' -count=1
```

**Red:** captured before production implementation. `traces/red-run.txt` shows
that the database package did not exist, a declaration-only admission could not
register because source execution was still required, and the dedicated runner
error did not yet exist.

## Green commands

```sh
go test ./internal/connectors/database ./internal/synccontract ./internal/connectors/engine ./internal/connectors/native/postgres -count=1
```

**Green:** passed after the implementation and after the contract-refinement
lint fixes. The recorded focused run covered `database`, `synccontract`,
`engine`, and `native/postgres`; all four packages passed. The same run also
proves `engine.Load(defs.FS, "postgres")` parses the embedded declaration
without changing the public metadata capabilities.

The initial green result is preserved in `traces/green-run.txt`; subsequent
focused reruns, direct lint, and repository gates are recorded in
`VERIFICATION.md`.
