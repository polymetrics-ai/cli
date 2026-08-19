# Code review — issue 4264

## Scope reviewed

- `cmd/connectorgen/certificationsweep.go`
- `cmd/connectorgen/certificationsweep_test.go`
- generated GitHub certification sweep and this phase evidence

## Findings

No actionable findings.

The classifier retains `writes.json` precedence, derives only from shared
declaration fields and standard REST operation semantics, makes ambiguous REST
actions a contextual error that becomes a sweep `product_defect`, and contains
no connector-specific condition. The GitHub artifact is generator-produced and
passes its check. The full `make verify` connector-boundary report is clean.

## Automated review routing

After the direct PR is opened as a trusted author, use `claude_auto`; do not
post a manual request unless the automatic workflow fails or is skipped.
