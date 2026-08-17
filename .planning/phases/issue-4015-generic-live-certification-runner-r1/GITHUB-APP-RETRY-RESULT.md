# GitHub App retry result

Source selection: `GITHUB-APP-200-RETRY.json`, generated only from the captain-supplied 2026-08-18 probe rows whose literal `new_status` was `200`. It selected 16 declaration-owned commands; no `/enterprises/` path was selected.

| HTTP status | Before (fine-grained run) | After (App run) | Delta |
| --- | ---: | ---: | ---: |
| 200 | 0 | 15 | +15 |
| 403 | 15 | 0 | -15 |
| 400 | 1 | 1 | 0 |

The App run executed all 16 selected commands, wrote 15 accepted schema-v2 records immediately, verified each with `connectorgen certification-matrix --check`, and removed its named credential/project afterward.

## Product defect

`dependabot list-alerts-for-org` was the one unchanged `400`.

- The captain-supplied direct App probe records `400 -> 200` for this exact command.
- The declaration-owned `pm` invocation returned HTTP 400 with the provider diagnostic reduced to `[redacted]`.
- This is both an execution discrepancy to investigate and a proof-quality defect: a non-pass receipt cannot carry GitHub's own response text when the transport replaces it wholesale with `[redacted]`.

The supplied probe narrative says 17 pure credential successes, but its TSV has 16 rows with `new_status=200` (15 `403 -> 200`, one `400 -> 200`). This run deliberately selected the literal TSV evidence rather than inventing a seventeenth command.
