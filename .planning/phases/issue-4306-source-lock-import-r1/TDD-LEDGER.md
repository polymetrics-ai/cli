# TDD ledger — source-lock operation import

## Cycle 1 — closed retrieval and integrity

- **Red (planned):** fixture tests demand connector-owned lock resolution, exact fixed URL retrieval, byte count and SHA-256 equality, bounds, and zero output after failure.
- **Green:** pending.
- **Refactor:** pending.

## Cycle 2 — canonical provider descriptors

- **Red (planned):** two synthetic connector fixtures require stable output and distinct path/query/header/body schemas, pagination/auth/media/output metadata, all declared response status shapes and fields (including rare/sensitive-looking fields), output classifications, and empty provider ID preservation.
- **Green:** pending.
- **Refactor:** pending.

## Cycle 3 — bounded reference and request contracts

- **Red (planned):** each rejection fixture must fail with its named reason before a descriptor is generated: external/cycle/ambiguity/missing references; duplicate/callback routes; unbounded/dynamic/unsupported request contract; and all configured limits.
- **Green:** pending.
- **Refactor:** pending.

## Cycle 4 — closed adoption command and documentation

- **Red (planned):** command/help tests require a connector plus lock-derived output only and reject URL/method/path/header/body/credential flags; docs tests require the migration contract.
- **Green:** pending.
- **Refactor:** pending.
