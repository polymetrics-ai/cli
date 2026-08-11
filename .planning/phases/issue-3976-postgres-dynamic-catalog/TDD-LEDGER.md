# TDD ledger — Issue #3976: PostgreSQL dynamic typed catalog discovery

| ID | Guarantee | RED assertion | GREEN proof |
| --- | --- | --- | --- |
| R1 | Dynamic variation | Two materially different live PostgreSQL schema fixtures can be represented by the same static catalog/table/field model. | Each fixture produces a distinct typed catalog/fingerprint without code or connector-bundle schema edits. |
| R2 | Structured source identity | Database/schema/table identity is flattened into a string or same-named relations across schemas merge. | Typed catalog values retain configured database, schema, and relation identity separately. |
| R3 | Ordered metadata | Column order, nullable state, primary-key membership/order, or deterministic relation order can be lost. | Live discovery and independent `pg_catalog` oracle agree on ordinal order and metadata. |
| R4 | Supported native/logical mapping | PostgreSQL native identity/modifiers collapse into the old coarse field type. | Supported native types retain native identity/modifiers and map through #4034's logical types. |
| R5 | Fail closed | Enum/domain/composite/unsupported shapes silently become string/object or return a guessed catalog. | A typed, named, secret-safe unsupported-shape error is returned before success. |
| R6 | Runtime connection | #4034's typed catalog foundation remains descriptor-only while production PostgreSQL `Catalog()` uses a disconnected static/coarse model. | The native runtime calls the typed adapter and any legacy projection derives from its result. |
| R7 | Scope / safety | System schemas, views, arbitrary SQL, target DDL/write, Parquet materialization, or CDC execution leak into source discovery. | Configured-schema base-table reads only; ownership tests/changed-path audit prove downstream boundaries stayed untouched. |
| R8 | Resource/cancellation | Queries ignore cancellation, leak rows/pools, or concatenate configured identity into SQL. | `QueryContext`, parameter binding, close/error paths, cancellation tests, and race/vet/lint proof pass. |

## RED command target

The exact focused command will be captured before implementation in
`traces/dynamic-catalog-red.txt`. It will execute only the new regression
against the PostgreSQL catalog adapter; the fixture will use the repository's
explicit local database-test harness and never print a connection string or
credential.

## GREEN command targets

```sh
go test -timeout 20m -count=1 ./internal/connectors/native/postgres
go test -race -timeout 20m -count=1 ./internal/connectors/native/postgres
```

The exact live-harness command and independent-oracle assertions will be added
with the test. Broader checks and their outcomes belong in `VERIFICATION.md`.
