# Verification — CI toolchain and integration branch guard

## Automated checks

- [ ] Go 1.25.12 red scan captured.
- [ ] Go 1.25.13 clean scan captured.
- [ ] All active explicit workflow pins and `go.mod` match.
- [ ] Workflow branch-name expression accepts the required integration name and rejects malformed names.
- [ ] Workflow YAML parses and repository CI configuration gates pass.
- [ ] `git diff --check` passes.

## Manual review focus

- [ ] No `govulncheck` suppression, exemption, skip, or non-blocking conversion.
- [ ] No special-case branch name and no relaxation of existing branch classes.
- [ ] Historical evidence remains unchanged.
