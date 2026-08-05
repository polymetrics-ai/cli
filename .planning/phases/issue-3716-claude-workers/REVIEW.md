# REVIEW — issue #3716 clean project-local Claude workers

Mode: standard inline review from the generated `code-review` prompt. The Pi adapter does not
provide the prompt's reviewer subagent, and the canonical worker contract forbids delegation, so
the review was performed inline against the parent-branch diff.

## Dispositions

### R1 — staticcheck error-string capitalization

- Severity: warning
- Disposition: fixed
- Finding: the first local `make lint` pass reported capitalized errors in the new Claude
  frontmatter parser/validator.
- Action: use idiomatic lowercase error strings in `internal/agentcontract/claude.go`.
- Evidence: rerun `make lint` reported `0 issues`.

### R2 — required skill access without delegation

- Severity: error
- Disposition: fixed
- Finding: the worker omitted `Skill`, so it could not load repository-required Go/design guidance.
- Root cause: bare `Skill` would expose every discoverable skill, including unrelated skills able
  to use `context: fork`; adding the unscoped tool would violate the single-worker boundary.
- Action: render only the required `Skill(name)` rules, deny `Agent` and `Task`, document the exact
  reachable names, and validate the policy and generated full files.

### R3 — clean-home selection evidence

- Severity: error
- Disposition: captain-authorized deferral
- Action: record the smoke as **NOT PERFORMED** until a real authenticated Claude session runs in a
  clean trusted home with unrelated global definitions. Static generation, precedence, and drift
  proof remain separate and do not satisfy the runtime criterion.

### R4 — stacked sub-PR ordering

- Severity: error
- Disposition: fixed
- Action: local verification, code review, no-mistakes child gates, commit, and push now precede PR
  creation; CI and GitHub automated review remain after the PR exists.

### R5 — Wave 1 review coverage

- Severity: error
- Disposition: captain-approved known gap
- Action: record PR #3724 as accepted after 23 successful checks, 6 skipped checks, and no failures,
  with no separate automated code-review pass. Do not retrofit coverage or block Wave 2.

## Final review

No actionable source finding remains. The renderer derives every worker field from the checked-in
canonical policy; sync is root-contained and only creates the two required Claude target files.
Whole-file exact comparison prevents a locally edited tool, skill, or denylist rule from bypassing
the canonical boundary. The clean-home smoke and Wave 1 coverage gap remain explicitly recorded
under the captain's decisions above.
