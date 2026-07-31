Part of #277 (S6). Write scope: `cli_surface.json`, `docs/cli/**`, `website/**`. Deps: #280, #281, #282.

- `cli_surface.json`: gh-like commands with `intent` (etl/direct_read/reverse_etl), `availability`, `record.*` flag mappings; namespace command renders contextual help. Update `docs/cli/**`, `website/**`, generated help/manual artifacts, tests per cli-help-docs-website-parity.md.

Acceptance: `pm help twenty`, `pm connectors inspect twenty --json` (no creds read), `pm connectors` bare-namespace help; website data regen idempotent.
Refs #277