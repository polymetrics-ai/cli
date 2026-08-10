# Verification — #3855 parent topology scaffold

## Required evidence

| Check | Evidence / command | Status |
| --- | --- | --- |
| Supplied isolated worktree | `pwd -P` equals `git rev-parse --show-toplevel`; worktree began detached and clean | passed |
| Current source ref | local `origin/feat/3862-any-to-any-transport` and `git ls-remote` both resolved to `30b2fb4aeb121641b6158903fe1d3b54668599a6` | passed |
| Combined branch ref | local `origin/docs/4015-connector-release-certification` and `git ls-remote` both resolved to `5996a8a2a5e99c8aa8eb5a8603ecb1f6bba21f12` | passed |
| Requested branch | `git switch --create feat/3855-polling-apply-foundations refs/remotes/origin/feat/3862-any-to-any-transport` | passed locally |
| GSD adapter | `scripts/gsd doctor`; required `sources` resolutions; generated lifecycle prompts | passed |
| Canonical delivery projection | `go run ./cmd/agentcontractgen check` | passed |
| Parent/child GitHub hierarchy | `gh-axi issue view` / `gh-axi issue subissue list` | live check found five open children: #3856–#3860; #3860 follows #3856–#3859 |
| #3880, #4016, #4019 live inspection | `gh-axi pr view` | pending — independent live PR inspection remains required |
| Markdown and diff hygiene | `git diff --check`; untracked/staged changed-path assertion | passed — only this phase directory is present |
| Repository checks | `make tidy-check`; `make docs-check`; `make lint` | passed |
| no-mistakes | contract argv with `--skip=push,pr,ci`, no `--yes` | pending after commit |
| Draft PR shape | `gh-axi pr create` then `gh-axi pr view` | pending after local gates |

## GSD verification result

The adapter and command provenance are valid. The inline fallback is required because the canonical
single-worker delivery contract forbids spawning a GSD role. This phase has no executable product
acceptance test: verification is the exact branch, artifact, PR-shape, and scope-fence evidence
listed above. A passing documentation check never certifies product behavior.

## Explicit non-applicable checks

- Go formatting, package tests, build, runtime services, credentials, provider calls, and database
  containers: no Go or runtime path changes.
- CLI help/manual/website parity: no command, flag, output, connector surface, or end-user
  documentation changes.
- Automated review: this PR must remain draft; review coverage is recorded as pending rather than
  requested while draft.
