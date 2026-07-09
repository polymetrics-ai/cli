# Tester Trace

- Added red loader test for `operation_ledger_version` and `operation` fields.
- Added red connectorgen tests for valid operation rows, legacy exclusions in ledger mode, dual
  classification, unblocked rows, missing reason, missing duplicate target, and missing source/notes.
- Added GitHub metrics test asserting 503 total rows, 100 covered rows, 403 operation rows, zero
  legacy exclusions, and model/risk/status counts.
- Verified targeted engine, connectorgen, conformance, and GitHub CLI command-surface packages.
