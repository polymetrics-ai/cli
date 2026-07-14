# Automated review routing best practices

Use local review routing to select the smallest independent review pass that covers the change risk.

## Source-backed practices

- Independent review catches different defects than implementation self-checks.
- Security-sensitive changes need a separate security pass focused on trust boundaries, secrets,
  filesystem paths, external effects, and untrusted input.
- Verification-sensitive changes need a reproducibility pass that checks exact commands, exit codes,
  and candidate head.
- Review evidence should be bounded and durable: exact head/diff range, reviewed files, findings,
  dispositions, and follow-up gates.

## Policy

- Default route: local `reviewer` pass after local verification.
- Add `security-auditor` for security or external-effect risk.
- Add `verifier` for command/gate evidence risk.
- Escalate to human review for human-gated findings or unavailable local review.
- Remote PR-bot review is optional and not part of the default gate.
