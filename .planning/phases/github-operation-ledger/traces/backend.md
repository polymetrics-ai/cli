# Backend Trace

- Added operation-ledger fields to the engine bundle model.
- Extended `api_surface.schema.json` with a versioned `operation` classifier.
- Updated connectorgen and conformance static checks with matching ledger-mode rules.
- Converted GitHub legacy `excluded` rows to blocked operation metadata while preserving covered rows.
