# Local review best practices

Use local automated review to catch defects before handoff without depending on remote PR-bot
review.

## Practices

- Bind every review to an exact head SHA or diff range.
- Use an independent role from the implementing worker: reviewer for quality/scope,
  security-auditor for risky inputs/effects, verifier for command evidence, debugger for failures.
- Keep reviewer passes read-only unless a separate fix worker is explicitly authorized.
- Review output is evidence, not authority. Disposition every actionable finding as accepted,
  accepted with modification, declined, deferred, or needs human.
- Rerun focused tests after accepted fixes.
- Rerun local review when fixes materially change the reviewed code.
- Record the review route, reviewed files/scope, findings, dispositions, and remaining human gates
  in the phase artifact, worker handoff, or PR body.

## Non-goals

- Do not require remote PR-bot review for default delivery.
- Do not treat optional remote PR-bot comments as required approval.
- Do not use review to bypass TDD, verification, issue scope, or final human merge gates.
