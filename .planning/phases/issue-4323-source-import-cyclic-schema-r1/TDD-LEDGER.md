# TDD ledger — issue 4323

| Slice | Red evidence | Green evidence | Refactor / scope evidence |
| --- | --- | --- | --- |
| Recursive source schemas | Pending: add the real importer-path behavioral test before importer implementation, then record its command and grammar-preflight cycle failure. | Pending: the same named test passes with explicit existing-path missing-foundation evidence, source pointer, and retained operation. | Confirm no descriptor flattening/truncation and no rewrite of v3 source-lock/provenance structures. |
| Non-recursive regression | Pending: include a non-cyclic fixture in the red suite to distinguish a valid import from cycle handling. | Pending: prove no spurious cycle gap is emitted for that fixture. | Keep the assertion at output behavior, not helper implementation. |
| Real affected connector | Pending: identify a checked-in affected connector using the importer and capture its pre-change failure. | Pending: its import accepts source-derived output without credentials. | Keep generated GitHub lock and descriptor bytes/checksums identical. |
