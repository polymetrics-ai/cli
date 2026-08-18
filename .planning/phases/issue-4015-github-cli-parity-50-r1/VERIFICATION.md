# Verification — issue #4015 GitHub declared-command parity

Status: **verified** on 2026-08-18.

The supplied inventory joins exactly to the current bundle: **50 unique paths, 50 matched, 25
implemented, 25 retained, zero missing**. No declaration was removed. All 25 promotions have one
fixed `api_surface` and pass the real commandrunner preflight; all 25 retained rows carry concrete
provider/runtime evidence.

## GSD and TDD evidence

- `scripts/gsd doctor` — pass; official GSD and the project-local Pi adapter are healthy.
- `scripts/gsd sources discuss-phase|plan-phase|execute-phase|verify-work|code-review` — each resolved
  to the pinned official source before use.
- `scripts/gsd prompt discuss-phase ...`, `plan-phase ... --tdd`, `execute-phase ...`,
  `verify-work ...`, and `code-review ...` — executed through the documented inline/manual fallback
  because the canonical single-worker contract forbids role spawning in this lane.
- Red/green/refactor evidence is recorded in `TDD-LEDGER.md`. The final live red caught GitHub
  returning JSON metadata instead of release bytes and then refusing an undeclared asset redirect;
  the repeated live green matched the exact 51-byte fixture.
- `scripts/verify-gsd-workflow` — pass; implementation changes have GSD/TDD evidence.
- `go run ./cmd/agentcontractgen check` — pass; canonical contract projections are current.

## Automated verification

| Command | Result |
| --- | --- |
| `go test -timeout 20m ./internal/connectors/connsdk -count=1` | pass |
| `go test -timeout 20m ./internal/connectors/engine -count=1` | pass |
| `go test -timeout 20m ./internal/connectors/commandrunner -count=1` | pass |
| `go test -timeout 20m ./cmd/connectorgen -count=1` | pass (148.665s) |
| `go test -timeout 20m ./internal/cli -count=1` | pass (544.001s) |
| `node scripts/github-command-reachability.mjs --pm ./pm --root .planning/phases/issue-4015-github-cli-parity-50-r1/.reachability-workers --report .planning/phases/issue-4015-github-cli-parity-50-r1/COMMAND-REACHABILITY.json --workers 16` | pass: 1,571 reachable, 0 unreachable; 1,546 implemented |
| `go vet ./...` | pass |
| `go build ./cmd/pm` | pass |
| `make tidy-check lint docs-check-no-build smoke-no-build agent-contract-check` | pass; owned smoke project deleted afterward |
| `make connectorgen-validate connectorgen-surface-sync github-parity-artifacts-check connectorgen-certification-matrix connectorgen-certification-candidates connectorgen-certification-sweep connector-boundary connector-runtime-preflight connector-canon-check release-workflow-check` | pass; 552 bundles valid, boundary clean, 1,571-row sweep current |
| `npm --prefix website run test:scripts` | pass (29 tests) |
| `npm --prefix website run lint` | pass with 13 pre-existing warnings and 0 errors |
| `npm --prefix website run typecheck` | pass |
| `npm --prefix website run build` | pass; expected unset local BetterAuth secret notices, exit 0 |
| `git diff --check` | pass |
| changed-file credential-pattern scan | pass; no token/private-key pattern or Keychain service identifier |
| sealed-project scan | pass; no `.polymetrics` project remains in the phase directory |

The repository explicitly says not to run `go test ./...` or `make verify` as one command under a
per-command timeout because the 550+ connector suite is routinely cut off. The affected packages,
the long `internal/cli` package, and every non-monolithic `make verify` gate were therefore run
separately as required; CI carries the full-suite fanout.

## CLI help, manual, skill, and website parity

- `./pm help github` — pass.
- bare `./pm github` — contextual GitHub command help at exit 0.
- `./pm github pr diff --help` and `./pm github run download --help` — exact command help and required
  binary destination flags rendered.
- `./pm docs generate --dir docs/cli --connectors-dir docs/connectors` and
  `./pm skills generate --dir docs/skills --json` — regenerated tracked artifacts.
- `make docs-check-no-build`, website generator checks, and docs/website grep assertions — pass.

## Live GitHub certification and cleanup

The certification credential was read inline from the macOS Keychain and passed only through an
environment variable into a sealed temporary pm project. Its value never entered argv, output,
fixtures, planning artifacts, or git.

- Safe read commands returned the declared shapes for repositories, licenses, gitignore templates,
  caches, secret metadata, variables, organizations, gists, codespaces, GPG keys, SSH keys, and agent
  tasks.
- Variable create/get/delete, autolink create/list/delete, workflow disable/enable, issue pin/unpin,
  PR draft/ready, and PR diff were directly observed against the authorized private fixture.
- Release download retrieved the exact 51-byte asset with `application/octet-stream` through the
  credential-stripped provider redirect and matched the source byte-for-byte.
- Independent cleanup reads returned 404 for the variable, autolink, workflow file, temporary PR
  branch, release, and release tag. The issue has no match, the existing PR is restored to
  closed/non-draft, and the codespace list is empty.
- GitHub refused codespace creation for the authorized repository. The repository has no workflow
  artifact, so `run download` reached its fixed provider path and returned 404 without a partial
  file. Gist creation was not executed because the launch authorization allows writes only inside
  the named organization/repository, while gists are user-scoped.
- Every sealed certification project, download root, reachability worker root, and smoke project was
  deleted after verification.

## Review

The inline GSD code review is recorded in `REVIEW.md`; no local finding remains. The direct PR uses
the repository's primary `claude_auto` route on open. Automated review records and any dispositions
are PR-hosted because they can only exist after the verified commit is pushed and the PR is opened.
