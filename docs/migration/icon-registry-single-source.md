# Connector icon registry single-source policy

Decision key: `icon-registry-single-source-bare-identifiers-20260802`.

`internal/connectors/icon_data.json` is the only authored connector-to-icon registry. Registry keys are canonical bare connector identifiers such as `apify-dataset` and `apple-search-ads`; `source-*` and `destination-*` keys are invalid. Direction is represented by connector capabilities, not duplicate icon identity.

Canonical SVG assets live under `docs/connectors/icons/**`. Website assets under `website/public/connectors/**` are generated copies from that docs tree. Simple Icons metadata needed for reviewed fetching (`simple_icon_slug`, `simple_icon_hex`, license, match fields, and review URL) is stored on the canonical registry entry, not in a website-local override file.

## Consumer contract

- Go runtime lookup is exact bare-key lookup through `ConnectorIconFor`; it does not synthesize or strip `source-*`/`destination-*` fallbacks.
- Website bundle generation reads only `internal/connectors/icon_data.json` and fails if an implemented connector lacks a canonical registry entry.
- `website/scripts/fetch-simple-icons.mjs` reads Simple Icons fetch metadata from the canonical registry and writes refreshed SVGs into `docs/connectors/icons/simple-icons/**`; website generation then copies those assets to `website/public/connectors/icons/simple-icons/**`.
- Connector icon ownership helpers map both `docs/connectors/icons/**` source assets and `website/public/connectors/icons/**` generated copies back to the same canonical bare connector. Ambiguous shared fallback paths, orphan paths, duplicate changed paths, and undeclared icon paths are rejected instead of authorizing connector ownership.
- Registry rows with `implemented: false` are retained runtime/builtin dispositions only; they do not authorize connector definition ownership. Their declared asset paths are still declared, so ownership lookup reports them as `ErrConnectorIconPathRuntimeBuiltin` rather than as an undeclared orphan path.
- `cmd/iconregistrygen` treats the curated registry (`--curated`, defaulting to `--out`) as authored state: an empty or `source-*`/`destination-*`-prefixed curated connector key is a hard error naming the key and file, not a silently dropped/backfilled row. A curated key that survives decoding but owns neither a connector definition nor a runtime builtin is likewise a hard error, so curated review status, review URL, and attribution can never be silently reverted to fallback or built-in defaults. Prefix collapse is applied only to raw upstream registry records, never to curated rows.
- Website script invariants (bare registry keys, canonical-vs-generated asset parity, icon sync path containment) are gated by `pnpm run test:scripts` in the website workflow. That workflow runs for both canonical inputs of the generated website icon tree — `internal/connectors/icon_data.json` and `docs/connectors/icons/**` — so a refreshed or replaced canonical SVG cannot merge while `website/public/connectors/icons/**` stays stale.

## Source/destination collapse audit

The legacy registry had 22 base identifiers with more than one `source-*`/`destination-*` or bare entry. The canonical registry keeps only bare identifiers for implemented definitions and explicit runtime builtins.

| Bare identifier | Legacy entries | Canonical disposition | Audit note |
| --- | --- | --- | --- |
| `azure-blob-storage` | `destination-azure-blob-storage` → `icons/azureblobstorage.svg` (upstream_registry/upstream_seeded)<br>`source-azure-blob-storage` → `icons/azureblobstorage.svg` (upstream_registry/upstream_seeded) | not retained | No implemented definition/runtime builtin in this branch. |
| `bigquery` | `destination-bigquery` → `icons/bigquery.svg` (upstream_registry/upstream_seeded)<br>`source-bigquery` → `icons/bigquery.svg` (upstream_registry/upstream_seeded) | not retained | No implemented definition/runtime builtin in this branch. |
| `clickhouse` | `destination-clickhouse` → `icons/clickhouse.svg` (upstream_registry/upstream_seeded)<br>`source-clickhouse` → `icons/clickhouse.svg` (upstream_registry/upstream_seeded) | not retained | No implemented definition/runtime builtin in this branch. |
| `convex` | `destination-convex` → `icons/convex.svg` (official/official_verified)<br>`source-convex` → `icons/convex.svg` (upstream_registry/upstream_seeded) | `convex` → `icons/convex.svg` | Same SVG asset; selected official verified provenance. |
| `customer-io` | `destination-customer-io` → `icons/customer-io.svg` (upstream_registry/upstream_seeded)<br>`source-customer-io` → `icons/customer-io.svg` (upstream_registry/upstream_seeded) | `customer-io` → `icons/customer-io.svg` | Same SVG asset; collapsed to one bare connector entry. |
| `dynamodb` | `destination-dynamodb` → `icons/dynamodb.svg` (upstream_registry/upstream_seeded)<br>`source-dynamodb` → `icons/dynamodb.svg` (upstream_registry/upstream_seeded) | `dynamodb` → `icons/dynamodb.svg` | Same SVG asset; retained review URL-bearing upstream provenance. |
| `elasticsearch` | `destination-elasticsearch` → `icons/elasticsearch.svg` (official/official_verified)<br>`source-elasticsearch` → `icons/elasticsearch.svg` (upstream_registry/upstream_seeded) | `elasticsearch` → `icons/elasticsearch.svg` | Same SVG asset; selected official verified provenance. |
| `file` | `file` → `icons/pm-file.svg` (polymetrics/polymetrics)<br>`source-file` → `icons/file.svg` (upstream_registry/upstream_seeded) | `file` → `icons/pm-file.svg` with `implemented: false` | Existing bare Polymetrics runtime builtin disposition wins; it is not definition-owned. |
| `firebolt` | `destination-firebolt` → `icons/firebolt.svg` (official/official_verified)<br>`source-firebolt` → `icons/firebolt.svg` (upstream_registry/upstream_seeded) | not retained | No implemented definition/runtime builtin in this branch. |
| `gcs` | `destination-gcs` → `icons/googlecloudstorage.svg` (upstream_registry/upstream_seeded)<br>`source-gcs` → `icons/gcs.svg` (upstream_registry/upstream_seeded) | not retained | Conflicting source/destination assets and no implemented definition/runtime builtin; omitted rather than silently choosing. |
| `google-sheets` | `destination-google-sheets` → `icons/google-sheets.svg` (upstream_registry/upstream_seeded)<br>`source-google-sheets` → `icons/google-sheets.svg` (upstream_registry/upstream_seeded) | not retained | No implemented definition/runtime builtin in this branch. |
| `hubspot` | `destination-hubspot` → `icons/hubspot.svg` (upstream_registry/upstream_seeded)<br>`source-hubspot` → `icons/hubspot.svg` (upstream_registry/upstream_seeded) | `hubspot` → `icons/hubspot.svg` | Same SVG asset; collapsed to one bare connector entry. |
| `kafka` | `destination-kafka` → `icons/kafka.svg` (upstream_registry/upstream_seeded)<br>`source-kafka` → `icons/kafka.svg` (upstream_registry/upstream_seeded) | not retained | No implemented definition/runtime builtin in this branch. |
| `mssql` | `destination-mssql` → `icons/pm-warehouse.svg` (polymetrics/polymetrics)<br>`source-mssql` → `icons/pm-sample.svg` (polymetrics/polymetrics) | not retained | Conflicting fallback assets and no implemented definition/runtime builtin; omitted rather than silently choosing. |
| `mysql` | `destination-mysql` → `icons/mysql.svg` (upstream_registry/upstream_seeded)<br>`source-mysql` → `icons/mysql.svg` (upstream_registry/upstream_seeded) | not retained | No implemented definition/runtime builtin in this branch. |
| `oracle` | `destination-oracle` → `icons/oracle.svg` (upstream_registry/upstream_seeded)<br>`source-oracle` → `icons/oracle.svg` (upstream_registry/upstream_seeded) | not retained | No implemented definition/runtime builtin in this branch. |
| `postgres` | `destination-postgres` → `icons/postgresql.svg` (upstream_registry/upstream_seeded)<br>`source-postgres` → `icons/postgresql.svg` (upstream_registry/upstream_seeded) | `postgres` → `icons/postgresql.svg` | Same SVG asset; retained current-version upstream docs URL. |
| `redshift` | `destination-redshift` → `icons/redshift.svg` (upstream_registry/upstream_seeded)<br>`source-redshift` → `icons/redshift.svg` (upstream_registry/upstream_seeded) | not retained | No implemented definition/runtime builtin in this branch. |
| `s3` | `destination-s3` → `icons/s3.svg` (upstream_registry/upstream_seeded)<br>`source-s3` → `icons/s3.svg` (upstream_registry/upstream_seeded) | not retained | No implemented definition/runtime builtin in this branch. |
| `salesforce` | `destination-salesforce` → `icons/salesforce.svg` (official/official_verified)<br>`source-salesforce` → `icons/salesforce.svg` (upstream_registry/upstream_seeded) | `salesforce` → `icons/salesforce.svg` | Same SVG asset; selected official verified provenance. |
| `snowflake` | `destination-snowflake` → `icons/snowflake.svg` (upstream_registry/upstream_seeded)<br>`source-snowflake` → `icons/snowflake.svg` (upstream_registry/upstream_seeded) | not retained | No implemented definition/runtime builtin in this branch. |
| `teradata` | `destination-teradata` → `icons/teradata.svg` (upstream_registry/upstream_seeded)<br>`source-teradata` → `icons/teradata.svg` (upstream_registry/upstream_seeded) | not retained | No implemented definition/runtime builtin in this branch. |

## Catalog handoff

This foundation intentionally does not reconcile PR #3590's generated catalog allowances. The later ownership PR must preserve the exact generated catalog allowances for:

- `docs/connectors/catalog/all-connectors.json`
- `docs/connectors/catalog/all-connectors.md`
