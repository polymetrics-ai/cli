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

## Final review

No critical, warning, or informational finding remains. The renderer derives every worker field
from the checked-in canonical policy; sync is root-contained and only creates the two required
Claude target files. Whole-file exact comparison prevents a locally edited frontmatter allowlist
from bypassing the canonical no-`Agent` rule.
