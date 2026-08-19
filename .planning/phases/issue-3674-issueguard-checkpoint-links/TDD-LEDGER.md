# TDD Ledger — Issue 3674 Issueguard Unvalidated-Cloud-Checkpoint Issue Links

Manual-GSD fallback: `scripts/gsd prompt programming-loop` is absent from the repo-local command
registry (`unknown GSD command: programming-loop`), so this ledger records the manual GSD/TDD loop
per `PLAN.md`.

## Red/green slices

| Slice | Red evidence | Green evidence | Status |
|---|---|---|---|
| Checkpoint issue-link recognition | Red: `TestValidatePRAcceptsUnvalidatedCheckpointCanonicalIssueLinks` (LF, CRLF, GitHub-indented, GitHub-indented CRLF) run against the base-commit `guard.go` (`5d61794f7`) fails all four subtests with `ValidatePR() OK = false, violations = [PR body must reference an issue ...]`. | Green: `addCheckpointIssueRefs` extracts issues 3070-3077 as non-closing `checkpoint` refs from the canonical-issue-links section; all four subtests pass. | Green |
| Wording gate | Red: `TestValidatePRRejectsCanonicalIssueSectionWithoutCompletedTaskWording` fails with `ValidatePR() OK = true, want false` when the `hasCompletedTaskWording(...)` condition is removed from `addCheckpointIssueRefs` (mutant run in a scratch copy of the package). | Green: with the wording gate in place, a checkpoint-shaped body lacking `unvalidated cloud checkpoint for the completed ... task` is rejected for a missing issue reference. | Green |
| Checkpoint-heading gate | Red: `TestValidatePRRejectsCanonicalIssueSectionWithoutCheckpointHeading` fails with `ValidatePR() OK = true, want false` when the `unvalidatedCheckpointHeadingPattern` lookup is removed from `addCheckpointIssueRefs` (mutant run in a scratch copy of the package). | Green: with the heading gate in place, a body carrying the canonical-issue-links section plus the completed-task wording but no `## Unvalidated cloud checkpoint — do not merge yet` heading is rejected. | Green |
| Canonical-section bound | Red: `TestValidatePRIgnoresIssueLinksAfterCanonicalCheckpointSection` (LF, CRLF, GitHub-indented trailing heading) fails with `ValidatePR() harvested issue 9001 from the section after the canonical issue links` when the `markdownH2StartPattern` bounding branch is removed from `addCheckpointIssueRefs` (mutant run in a scratch copy of the package). | Green: with the bound in place, a checkpoint body whose canonical-issue-links section is followed by a `## Later notes outside the task record` H2 carrying a bare issue URL still yields exactly issues 3070-3077. | Green |
| Shared negation filter | Red: `hasCompletedTaskWording` and `hasExplicitIssueWording` were byte-identical scan/skip-negated/return loops differing only in the pattern, so a fix to the negation window had to be applied twice. | Green: `hasNonNegatedMatch(re *regexp.Regexp, text string) bool` is the single implementation; both callers delegate to it and the full package suite stays green (behavior-neutral). | Green |
| General gate not loosened | Red: `TestValidatePRRejectsAmbiguousIssueRelationship` includes a bare `https://github.com/polymetrics-ai/cli/issues/123` body, which must stay rejected outside the checkpoint pattern. | Green: bare issue URLs are only harvested inside the bounded canonical-issue-links section of a doubly gated checkpoint body; the ambiguous-relationship suite still passes. | Green |

## Actual evidence

```bash
gofmt -l internal/coordination/issueguard
# no output

go vet ./internal/coordination/issueguard/...
# exit 0

go test ./internal/coordination/issueguard/... -count=1
# ok polymetrics.ai/internal/coordination/issueguard 0.420s

go build ./cmd/pm
# exit 0

# base-commit red evidence (scratch copy: guard.go from 5d61794f7 + current guard_test.go)
go test ./...
# --- FAIL: TestValidatePRAcceptsUnvalidatedCheckpointCanonicalIssueLinks (LF, CRLF, GitHub-indented, GitHub-indented CRLF)

# gate mutants (scratch copies with one gate removed each)
go test -run TestValidatePRRejectsCanonicalIssueSectionWithoutCheckpointHeading ./...
# --- FAIL: guard_test.go:164: ValidatePR() OK = true, want false
go test -run TestValidatePRRejectsCanonicalIssueSectionWithoutCompletedTaskWording ./...
# --- FAIL: guard_test.go:144: ValidatePR() OK = true, want false
go test -run TestValidatePRIgnoresIssueLinksAfterCanonicalCheckpointSection ./...
# --- FAIL: ValidatePR() harvested issue 9001 from the section after the canonical issue links
```

## Notes

- Red evidence for the two gate tests was produced by mutation runs in throwaway copies of the
  package outside the worktree; no gate was ever weakened on the branch.
- `completedTaskPattern` still requires the trailing noun `task` in the checkpoint sentence, so the
  marketo checkpoint body (PR #3578), whose sentence ends in `implementation`, remains unrecognized.
  That is a deliberate open decision for firstmate, not covered by this ledger.
- No secrets requested, printed, summarized, or stored. No live provider or credentialed calls.
