# Correction round 1 of 5 — issue #3995

## Tracking

- Child issue: [#4024](https://github.com/polymetrics-ai/cli/issues/4024),
  `fix(agentcontract): fail closed proof-consumer fingerprints and semantic JSON`
- Parent relationship: #4024 is a GitHub sub-issue of #3995 and contains `Refs #3988`.
- Review findings: R-1 strict fingerprint-only proof consumption; R-2 canonical semantic JSON
  comparison.

## Commit linkage

- Remediation: `842f1c271 fix(3995): harden certification proof validation`
- Initial verification record: `d511186bc docs(3995): record certification gate verification`
- This record is committed after #4024 was created and attached, before resuming the same
  delivery validation with a fresh pipeline run.

## Disposition

The evaluator accepts only declared, repository-salted fingerprint values (or explicit `null`) in
all consumed proof locations and rejects a raw JSON scalar fail closed. Typed proof comparison uses
canonical JSON so semantically identical whitespace cannot create false evidence drift. The scope
remains read-only and provider-free. #3989 remains the future proof-schema integration owner;
unknown versions continue to halt until an explicit change consumes them.

## Workflow recovery

No-mistakes run `01KZPTX65SYF6JJJK8QH7BWTBA` was cancelled by operator before it reached a
code-gate result. It is workflow recovery only, not a failed validation and not an additional
correction round. Do not sync or resume that run; start one fresh full run without `--yes` after
this provenance commit.
