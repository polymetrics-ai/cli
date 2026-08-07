# Discussion log — #3771 command-runner runtime content

## Mode

Auto-mode discussion. The parent issue and all four child issues provide the product decisions,
acceptance criteria, dependency order, file ownership, and safety boundaries. No unresolved
product choice remains to ask a human.

## Reversal being implemented

Earlier runner tests encoded `***` masking of connector-command ETL records, reverse previews,
and errors. The captain's standing no-redaction policy overrules that behavior. The tests will be
intentionally changed from masking assertions to exact content-preservation assertions; this is a
policy correction, not a weakened test suite.

## Non-goals confirmed

- Do not remove `redact_fields` declarations from connector bundles.
- Do not alter generic source-table reverse-plan redaction/output behavior.
- Do not alter approval-token serialization or destructive-write lifecycle.
- Do not add API-surface or capability declarations, so this work cannot mint an implemented
  command without an executable endpoint.

## Execution fallback

The issue is not represented as a GSD roadmap phase, so the Pi adapter cannot initialize the
normal phase directory. The local plan and ledger are the explicit manual fallback required by
the delivery contract; their checks will be kept current through implementation and review.
