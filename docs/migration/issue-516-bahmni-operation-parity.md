# Issue #516 Bahmni operation parity

This migration note no longer duplicates the Bahmni operation matrix.

Current Bahmni connector facts are owned by the bundle:

- `internal/connectors/defs/bahmni/api_surface.json` owns endpoint coverage and blocked-operation evidence.
- `internal/connectors/defs/bahmni/writes.json` owns retained reverse-ETL actions.
- `internal/connectors/defs/bahmni/cli_surface.json` owns typed CLI command metadata.
- `internal/connectors/defs/bahmni/docs.md` owns the generated connector manual prose.

Generated connector docs, catalog data, and website connector data must be regenerated from those
sources instead of hand-synchronizing a second parity table here.
