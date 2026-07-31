Part of #277 (S7). Write scope: `fixtures/**`, `docs.md`. Deps: #279, #280, #281, #282, #283.

- `fixtures/**` for streams + writes; `docs.md`; `pm connectors certify twenty`; parity-deviation ledger entries for any documented gap (target: none).

Acceptance: `make connectorgen-validate`, `make verify`, focused tests, `pm connectors certify twenty` all green.
Refs #277

<!-- captain-policy-twenty-destructive-confirmation-v1 -->
## Captain policy addendum — destructive/admin parity safety

This issue's existing scope and operation counts are preserved. Documented Twenty CRM DELETE/destructive/admin operations remain in scope for the connector ledger instead of being blanket-excluded as unsafe. They may be executable only when represented by connector-owned typed schemas, bounded fixtures, `confirm: "destructive"` / typed destructive confirmation, and the existing reverse-ETL plan -> preview -> explicit approval -> execute path.

This addendum authorizes no live provider calls, no credentials, no generic raw write tools, no unsafe execution, and no count changes; it only records the captain policy that destructive operations are included with typed confirmation and safety gates.
