# DISCUSSION LOG — issue #3853

Mode: `discuss-phase` inline/manual fallback using parent issue #3853 and the inspected source
paths.

No unresolved gray area remains. The issue fixes the boundary: engine preview and engine
direct-read/binary-download error text must preserve connector content, while legacy
`redact_fields` declarations stay load-compatible. The explicit exclusions keep #3771's
command-runner ownership, #3852's declarable-policy ownership, successful-response policy work,
and generic source-table plan samples out of this branch.

The required delivery topology is one branch and eventual PR. The worker does not spawn roles,
does not call providers, does not handle credentials, and never executes reverse ETL.
