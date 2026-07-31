# SOURCES — Crisp parity wave02 r1

## Official sources

- REST API reference: https://docs.crisp.chat/references/rest-api/v1/
- Authentication guide: https://docs.crisp.chat/guides/rest-api/authentication/
- Website token guide: https://docs.crisp.chat/guides/rest-api/authentication/website-token/
- Plugin token guide: https://docs.crisp.chat/guides/rest-api/authentication/plugin-token/
- Rate limits guide: https://docs.crisp.chat/guides/rest-api/rate-limits/

## Generated inventory

Generated operation rows: 234.

Method counts:

- DELETE: 26
- GET: 91
- HEAD: 14
- PATCH: 44
- POST: 47
- PUT: 12

Lane counts:

- ETL/read: 82
- direct/provider search/query: 12
- binary/file/import/export: 8
- changefeed/events: 4
- reverse/write: 114
- HEAD/existence checks: 14

The official documentation parse found 14 HEAD existence-check rows in addition to the parent r2 audit's 220 non-HEAD operation allocation. GitHub issue count tables were preserved; the connector-local ledger records the HEAD rows as planned/blocked non-data operations.

`cli_surface.json` keeps the same lane grouping but marks all non-read HTTP methods as planned typed write metadata (`reverse_etl` for write lane, `direct_write` for direct/binary/changefeed lanes) with approval requirements. DELETE/destructive rows include typed destructive confirmation.

## Auth evidence

Crisp REST API authentication uses Basic auth token keypairs plus `X-Crisp-Tier` (`website` or `plugin` depending on token type). This bundle models `token_id` and `token_key` as secret fields only; no secret values, provider calls, writes, certification, VPS/Thaalam work, push, PR, or merge were used.

## Parser/evidence artifacts

- Deterministic parser/generator: `generate_crisp_bundle.py`
- Parsed source inventory: `crisp-source-inventory.json`
- Generated connector bundle: `internal/connectors/defs/crisp/**`
- Generated docs/catalog: `docs/connectors/crisp/**`, `docs/connectors/README.md`, `docs/connectors/catalog/all-connectors.*`
- Generated website connector data: `website/data/connectors.generated.json`, `website/lib/connectors*.generated*`
