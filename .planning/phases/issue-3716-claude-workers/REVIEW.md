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
- Action: the first scoped-`Skill(name)` correction was incomplete because it retained unqualified
  names that personal skills could shadow. The final correction uses Claude's documented `skills`
  frontmatter, preloads only plugin-qualified `cc-skills-golang:*` and
  `frontend-design:frontend-design` sources, and omits and denies runtime `Skill` together with
  `Agent` and `Task`. The generator records the three unqualified design skills that become
  unavailable and the handoff cost for website/docs UI work.

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

### R6 — recursive Claude agent inventory

- Severity: error
- Disposition: fixed
- Root cause: whole-file checks covered only registered target paths even though Claude recursively
  discovers every Markdown definition under `.claude/agents` and identifies agents by frontmatter
  name.
- Action: inventory the complete project tree before drift comparison and reject symlinks,
  unexpected definitions, duplicate names, canonical name/path mismatches, and missing targets.

### R7 — canonical path portability

- Severity: warning
- Disposition: fixed
- Root cause: native `filepath.Clean` was applied to slash-separated JSON contract paths, so valid
  targets normalize differently on Windows.
- Action: validate contract paths with `io/fs.ValidPath` and `path.Clean`; convert to native paths
  only at the filesystem boundary.

### R8 — Windows GSD adapter execution

- Severity: warning
- Disposition: fixed
- Root cause: the checker executed the extensionless JavaScript shebang file directly and its test
  fixture required `/bin/sh`.
- Action: resolve Node and pass `scripts/gsd` as its script argument; use a portable JavaScript test
  fixture that also checks the selected working directory and argv.

## Final review

No actionable source finding remains. The renderer derives every worker field from the checked-in
canonical policy; sync is root-contained and only creates the two required Claude target files.
Recursive inventory plus whole-file exact comparison prevents an extra definition or a locally
edited tool, preload, or denylist from bypassing the canonical boundary. Official documentation
demonstrates the plugin-namespace collision rule; generated-file tests demonstrate configured
qualified identifiers and tool denial. They do not demonstrate authenticated clean-home runtime
selection, plugin source/version pinning, or immunity to managed/CLI overrides. The clean-home
smoke and Wave 1 coverage gap remain explicitly recorded under the captain's decisions above.
