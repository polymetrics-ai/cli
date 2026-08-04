# AWS CloudTrail parity resume r1 summary

## Status

Current-main scoped validation is green; commit and delivery gate pending.

## Intended result

Restore runtime reachability for the 57 safe, documented AWS CloudTrail operations already represented by the connector bundle, while retaining the three unrestricted-SQL operations as policy-disallowed.

AWS's official [CloudTrail Actions](https://docs.aws.amazon.com/awscloudtrail/latest/APIReference/API_Operations.html)
reference enumerates 60 actions. This connector makes 57 genuinely reachable: 19 ETL streams, 9
typed direct reads, and 29 typed reverse-ETL writes. `StartQuery`, `CreateDashboard`, and
`UpdateDashboard` deliberately remain `unsafe_or_disallowed` because their request models admit
unrestricted CloudTrail Lake SQL `QueryStatement` text; they are policy refusals, not planned
operations.

Provider request-field research is recorded in
`traces/aws-cloudtrail-request-field-citation-research.json`: all 187 request fields have an
operation-specific AWS API Reference URL, `Request Parameters` source section, provider-reference
evidence type, high confidence, and `Required: Yes/No` rationale (57 required, 130 optional).
