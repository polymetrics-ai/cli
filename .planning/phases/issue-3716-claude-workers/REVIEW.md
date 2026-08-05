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

### R9 — nested Claude project scopes

- Severity: error
- Disposition: fixed
- Root cause: the first inventory correction walked only the repository-root `.claude/agents`
  tree, while Claude discovers every nested `.claude/agents` scope between the launch directory
  and repository root and gives the closest same-name definition precedence.
- Action: walk the repository once at the shared validation boundary, skip Git metadata, recognize
  every nested Claude agent scope, and reject scope symlinks, unexpected definitions, and duplicate
  canonical names before checking registered projections.

### R10 — CRLF projection drift

- Severity: warning
- Disposition: fixed
- Root cause: the parser and whole-file comparator required raw LF bytes, so Git's platform EOL
  conversion could make a canonical Windows checkout fail before semantic validation.
- Action: normalize CRLF to LF for Claude frontmatter parsing, checking, and no-op sync comparison;
  retain byte-exact canonical comparison after that line-ending normalization.

### R11 — case-folded Git metadata exclusion

- Severity: error
- Disposition: fixed
- Root cause: repository inventory skipped every directory whose name case-folded to `.git`, even
  though `.GIT` and `.Git` are ordinary discoverable project directories on case-sensitive hosts.
- Action: prune only the exact root `.git` metadata directory. Inventory all case variants and
  nested paths normally, then reject any unexpected or duplicate Claude definition they contain.

### R12 — CRLF in canonical expected output

- Severity: warning
- Disposition: fixed
- Root cause: the first EOL correction normalized only checked-out bytes. Canonical JSON strings
  can decode escaped CRLF into renderer output, leaving expected CRLF unequal to normalized actual
  LF and causing every sync to rewrite the same projection.
- Action: make the Claude renderer emit LF-canonical whole files and normalize both expected and
  actual operands again at the shared check/sync boundary before exact comparison or writing.

### R13 — non-idempotent repeated carriage-return normalization

- Severity: warning
- Disposition: fixed
- Root cause: pairwise CRLF replacement reduced only the final carriage return in a repeated run,
  so `\r\r\r\n` required three calls to reach LF and renderer/check/sync boundaries could compare
  different intermediate forms.
- Action: collapse each complete carriage-return run immediately before LF to one LF in a single
  linear pass while preserving bare carriage returns. Add direct fixed-point cases and inject a
  repeated run through canonical render/check/sync coverage.

## Final review

No actionable source finding remains. The renderer derives every worker field from the checked-in
canonical policy; sync is root-contained and only creates the two required Claude target files.
Repository-wide nested-scope inventory plus EOL-normalized whole-file comparison prevents an extra
definition or a locally edited tool, preload, or denylist from bypassing the canonical boundary.
Only exact root `.git` metadata is pruned; case-variant and nested directories remain inventoried.
The renderer and validator share idempotent LF-canonical expected/actual semantics, including
repeated carriage-return runs. Official documentation demonstrates the plugin-namespace collision
rule; generated-file tests demonstrate configured qualified identifiers and tool denial. They do
not demonstrate authenticated clean-home runtime selection, plugin source/version pinning, or
immunity to managed/CLI overrides. The clean-home smoke and Wave 1 coverage gap remain explicitly
recorded under the captain's decisions above.
