# Native Go Connector Catalog Threat Model

- Secret values must never be embedded, logged, rendered, or returned.
- Upstream image references are metadata only and must not become executable runtime inputs.
- Catalog-only connectors must not pass credential validation as runnable connectors.
- Agent output must expose docs and schema metadata, not generic HTTP or shell mutation tools.
