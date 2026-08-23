# TDD ledger — issue 4323

| Slice | Red evidence | Green evidence | Refactor / scope evidence |
| --- | --- | --- | --- |
| Recursive source schemas | **Red (2026-08-23):** `go test -count=1 -timeout 20m ./cmd/connectorgen -run '^TestSourceImportRetainsRecursiveSchemaReferencesAsSourceBoundGaps$'` failed for direct, mutual, and deeply nested fixtures. Each stopped at `preflight source grammar … response "200": reference cycle at "#/components/schemas/Folder"`; no operation descriptor was retained. | Pending: the same named test passes with explicit existing-path missing-foundation evidence, source pointer, and retained operation. | Confirm no descriptor flattening/truncation and no rewrite of v3 source-lock/provenance structures. |
| Non-recursive regression | Pending: include a non-cyclic fixture in the red suite to distinguish a valid import from cycle handling. | Pending: prove no spurious cycle gap is emitted for that fixture. | Keep the assertion at output behavior, not helper implementation. |
| Real affected connector | Pending: identify a checked-in affected connector using the importer and capture its pre-change failure. | Pending: its import accepts source-derived output without credentials. | Keep generated GitHub lock and descriptor bytes/checksums identical. |
