# Threat Model

## Threats

- Agent assumes a planned connector can run ETL or reverse ETL.
- CDC setup guidance causes unsafe database configuration.
- Reverse ETL actions are enabled for a connector without action validation.
- Secret fields are copied into plans or docs.

## Mitigations

- Native port plan is distinct from runtime capabilities.
- Planned connectors keep data-plane booleans false.
- CDC sections describe prerequisites but do not execute database commands.
- Reverse ETL operations remain empty until write conformance exists.
- Tests verify secret names only; secret values are not present.
