# Threat Model

## Risks

- User-supplied command flags enter GraphQL variables.
- Project data can require elevated GitHub scopes.
- Mutations could affect visible repository/project state if prematurely exposed.

## Controls

- Fixed GraphQL documents only; flags never alter query text.
- Existing command argument validation and interpolation CR/LF guards remain active.
- New read streams are ETL only.
- Project/discussion mutations stay planned/direct-write until reverse-ETL schema and approval
  policy exist.
- No secret values are requested, printed, stored, summarized, or invented.
