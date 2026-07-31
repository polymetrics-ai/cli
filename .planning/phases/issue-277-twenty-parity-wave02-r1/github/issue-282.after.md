Part of #277 (S5). Write scope: `writes.json` (destructive rows). Deps: #281 (serializes shared writes.json).

- 28 `destructive_admin` delete actions (`DELETE /rest/{objects}/{id}`), `risk: destructive`, typed-confirmation required; blocked by default outside plan/approval/execute.

Acceptance: destructive rows gated; `covered_by` mapping complete.
Refs #277

<!-- captain-policy-twenty-destructive-confirmation-v1 -->
## Captain policy addendum — destructive/admin parity safety

This issue's existing scope and operation counts are preserved. Documented Twenty CRM DELETE/destructive/admin operations remain in scope for the connector ledger instead of being blanket-excluded as unsafe. They may be executable only when represented by connector-owned typed schemas, bounded fixtures, `confirm: "destructive"` / typed destructive confirmation, and the existing reverse-ETL plan -> preview -> explicit approval -> execute path.

This addendum authorizes no live provider calls, no credentials, no generic raw write tools, no unsafe execution, and no count changes; it only records the captain policy that destructive operations are included with typed confirmation and safety gates.
