Part of #277 (S4). Write scope: `writes.json` (non-destructive rows). Deps: #278, #279.

- 84 `reverse_etl` actions: create `POST /rest/{objects}`, update `PATCH /rest/{objects}/{id}`, batch `POST /rest/batch/{objects}` (≤60) with `record_schema`, `path_fields`, `body_type`, `risk: normal`. Plan → preview → approval → execute only.

Acceptance: every `reverse_etl` row maps to an `api_surface` write via `covered_by`.
Refs #277

<!-- captain-policy-twenty-destructive-confirmation-v1 -->
## Captain policy addendum — destructive/admin parity safety

This issue's existing scope and operation counts are preserved. Documented Twenty CRM DELETE/destructive/admin operations remain in scope for the connector ledger instead of being blanket-excluded as unsafe. They may be executable only when represented by connector-owned typed schemas, bounded fixtures, `confirm: "destructive"` / typed destructive confirmation, and the existing reverse-ETL plan -> preview -> explicit approval -> execute path.

This addendum authorizes no live provider calls, no credentials, no generic raw write tools, no unsafe execution, and no count changes; it only records the captain policy that destructive operations are included with typed confirmation and safety gates.
