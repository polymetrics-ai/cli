# Eval Plan

## Functional Eval

- Sample 10 source slugs across runtime families and run direct `check/catalog/read`.
- Sample 5 destination slugs and run reverse ETL into local receipts.
- Confirm database-family slugs expose query capability and reject non-SELECT statements.

## Safety Eval

- Search generated docs and JSON outputs for known secret-like strings.
- Confirm no code path invokes connector images or non-Go runtimes.
- Confirm reverse ETL still fails without approval.
