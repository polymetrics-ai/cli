Part of #277 (S3). Write scope: `streams.json`. Deps: #278, #279.

- 28 list streams (`stream_read`, cursor pagination on `pageInfo`/`endCursor`, `starting_after`+`limit`, `order_by`, `filter`, `depth`) + 28 `direct_read` get-by-id ops. Each references its S2 schema.

Acceptance: pagination/cursor fields set; focused stream tests.
Refs #277

<!-- captain-policy-twenty-destructive-confirmation-v1 -->
## Captain policy addendum — destructive/admin parity safety

This issue's existing scope and operation counts are preserved. Documented Twenty CRM DELETE/destructive/admin operations remain in scope for the connector ledger instead of being blanket-excluded as unsafe. They may be executable only when represented by connector-owned typed schemas, bounded fixtures, `confirm: "destructive"` / typed destructive confirmation, and the existing reverse-ETL plan -> preview -> explicit approval -> execute path.

This addendum authorizes no live provider calls, no credentials, no generic raw write tools, no unsafe execution, and no count changes; it only records the captain policy that destructive operations are included with typed confirmation and safety gates.
