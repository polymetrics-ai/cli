# Native Go Connector Catalog PLAN

1. Add failing catalog model tests for count, field completeness, enabled mapping, filters, and deterministic slug lookup.
2. Add failing CLI/docs tests for `--all`, `catalog`, catalog-only `inspect`, generated catalog docs, and docs validation.
3. Implement an embedded generated catalog and helper APIs in `internal/connectors`.
4. Add a native Go catalog generator command that reads the public registry JSON and writes embedded JSON plus Markdown/JSON docs.
5. Extend connector CLI and docs generation/validation for catalog entries.
6. Run targeted tests, regenerate docs/catalog, then run full verification.
