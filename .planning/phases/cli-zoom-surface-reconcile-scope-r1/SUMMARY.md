# Summary — scoped operation-surface reconciliation, R1

`connectorgen surface-reconcile` now accepts `--notes-contains <text>`. It selects only ledger
operations whose durable provenance note contains the supplied text and combines that selector with
the existing reason selector when both are present. This foundation lets the Zoom 35-category run
reconcile one `provider_module=<name>` at a time without rewriting unrelated blocked reasons.
