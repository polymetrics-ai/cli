# Summary — PR #4308 status-check result preservation

`runConnectorCommand` now delegates to a shared result shaper that preserves a typed `StatusCheck` before the legacy ETL fallback. JSON emits the additive `ConnectorCommandStatusCheck` envelope and human mode emits one deterministic, non-empty status line. Binary-download and direct-read shaping were extracted without semantic changes.

The focused tests prove happy, non-200, zero-byte HEAD, fallback-classification, and binary manifest behavior. The ordinary loader rejects malformed status/text-export declarations before I/O, and the temporary source-locked public GitHub Pages proof reached the provider through the installed binary, then was removed without committing any declaration, icon alias, download, or credential carrier.
