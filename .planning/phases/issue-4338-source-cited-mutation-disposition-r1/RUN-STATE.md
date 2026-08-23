# Run state — issue 4338

- Base verified: `e338cd301` is the detached-start HEAD and an ancestor before
  branch creation.
- Branch: `fm/cli-mutation-disposition-foundation-r1`.
- Issue: #4338 opened before implementation.
- Red: focused source-projection behavior test failed before implementation with
  the expected undefined disposition model/application symbols.
- Green: the focused source-projection and executable-coverage suite passes for
  separate Asana, Jira, Sentry, and Vercel-sized fixtures, closed safety
  controls, and the byte-identical GitHub projection control.
- Verification: affected-package tests, `internal/cli`, vet, build, generated
  checks, certification/boundary/canon/runtime preflight, release, docs, and
  smoke gates pass serially; see `VERIFICATION.md`.
- Review: inline GSD standard review found no actionable finding; see
  `REVIEW.md`.
- Merge-before-push: fetched `origin/main` on 2026-08-24; it remains
  `e338cd301`, and `git merge origin/main` reported `Already up to date.`
- Status: ready to push/open the direct PR.
