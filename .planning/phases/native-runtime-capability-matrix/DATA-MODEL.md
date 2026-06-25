# Data Model

## RuntimeCapabilities

- `metadata`: connector catalog metadata is available.
- `check`: native credential/reachability check is enabled.
- `catalog`: native stream/action catalog discovery is enabled.
- `read`: native ETL extraction is enabled.
- `write`: native destination or reverse ETL mutation is enabled.
- `query`: native query execution is enabled.
- `etl`: source-to-destination ETL can use this connector in an enabled path.
- `reverse_etl`: reverse ETL can use this connector in an enabled path.
- `unsupported_reason`: human and agent readable reason when a runtime operation is disabled.

The matrix is generated into `internal/connectors/catalog_data.json` and exported in `docs/connectors/catalog/all-connectors.json`.
