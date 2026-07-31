Part of #277 (S2). Write scope: `schemas/**`. Deps: #278.

- 28 per-object JSON schemas, 546 fields total; each with `x-primary-key: id` and `x-cursor-field: updatedAt`; common `id`/`createdAt`/`updatedAt` typed.

Acceptance: `jq .` clean; conformance `path_fields ⊆ record_schema` prerequisite satisfied.
Refs #277

<!-- captain-policy-twenty-destructive-confirmation-v1 -->
## Captain policy addendum — destructive/admin parity safety

This issue's existing scope and operation counts are preserved. Documented Twenty CRM DELETE/destructive/admin operations remain in scope for the connector ledger instead of being blanket-excluded as unsafe. They may be executable only when represented by connector-owned typed schemas, bounded fixtures, `confirm: "destructive"` / typed destructive confirmation, and the existing reverse-ETL plan -> preview -> explicit approval -> execute path.

This addendum authorizes no live provider calls, no credentials, no generic raw write tools, no unsafe execution, and no count changes; it only records the captain policy that destructive operations are included with typed confirmation and safety gates.
