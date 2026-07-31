Part of #277 (S6). Write scope: `cli_surface.json`, `docs/cli/**`, `website/**`. Deps: #280, #281, #282.

- `cli_surface.json`: gh-like commands with `intent` (etl/direct_read/reverse_etl), `availability`, `record.*` flag mappings; namespace command renders contextual help. Update `docs/cli/**`, `website/**`, generated help/manual artifacts, tests per cli-help-docs-website-parity.md.

Acceptance: `pm help twenty`, `pm connectors inspect twenty --json` (no creds read), `pm connectors` bare-namespace help; website data regen idempotent.
Refs #277

<!-- captain-policy-twenty-destructive-confirmation-v1 -->
## Captain policy addendum — destructive/admin parity safety

This issue's existing scope and operation counts are preserved. Documented Twenty CRM DELETE/destructive/admin operations remain in scope for the connector ledger instead of being blanket-excluded as unsafe. They may be executable only when represented by connector-owned typed schemas, bounded fixtures, `confirm: "destructive"` / typed destructive confirmation, and the existing reverse-ETL plan -> preview -> explicit approval -> execute path.

This addendum authorizes no live provider calls, no credentials, no generic raw write tools, no unsafe execution, and no count changes; it only records the captain policy that destructive operations are included with typed confirmation and safety gates.
