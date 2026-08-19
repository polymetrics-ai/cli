# Verification checklist — issue 4276

## Before push

- [x] Read `docs/architecture/connector-certification-design.md`; it requires
  accepted proof to certify, but defines no precondition prohibiting Zoom from
  entering certification scheduler scope.
- [x] Pre-change protected transition gate returns `HALT` with
  `capability/zoom/missing`.
- [x] Pre-change scoped matrix check rejects Zoom as non-allowlisted.
- [x] Regenerated and checked Zoom's scoped certification shard after rebasing
  onto `origin/main` at `31bfe62eb`.
- [x] Regenerated and checked the complete certification status projection.
- [x] Post-change gate is `RETRY` for unproven cells and has no
  `capability/zoom/missing` failure.
- [x] Generated Zoom status remains `COMMUNITY BUILD, UNCERTIFIED`.
- [x] Confirmed the changed-file count is below 10 and no Zoom definition source or
  unrelated connector artifact changed.
- [x] Ran post-rebase focused Go tests, generator checks, `gofmt`, `go vet`,
  and `go build ./cmd/pm`.
- [x] Final `make verify` passed (2026-08-19) after the fixed-count test
  expectations were replaced with exact generated-allowlist consistency checks.
- [x] `git diff --check` and inline code review found no actionable issue.

## PR read-back

- [ ] API reports base branch exactly `main`.
- [ ] PR body contains `Refs #4276` and records the GSD/TDD evidence, skills,
  generator verification, full `make verify`, before/after gate verdicts, and
  docs parity decision.
- [ ] Required CI checks and mergeable state are read from GitHub; no merge is
  performed.
