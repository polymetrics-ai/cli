# Verification checklist — issue 4361

- [ ] Focused red test result recorded.
- [ ] Focused engine tests green with `-timeout 20m`.
- [ ] Focused commandrunner tests green with `-timeout 20m`.
- [ ] Full changed `cmd/connectorgen` package tests green with `-timeout 20m`.
- [ ] `go build ./cmd/pm` green.
- [ ] Source/definition, generated docs/help/catalog, connector-boundary, and relevant repository checks recorded.
- [ ] Current `origin/main` merged normally, with post-merge tests rerun.
- [ ] `git diff --check` green.
- [ ] Credential-free current-main command census records zero usable-surface delta rather than fabricating a command.
- [ ] Fresh-context independent Codex exact-head audit complete; every blocker fixed or explicitly absent.
- [ ] PR base verified from the GitHub API-equivalent response after opening.

