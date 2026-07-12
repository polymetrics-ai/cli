# Verification: Phase 0

No implementation gate is claimed yet. `verificationPassed` remains false until the real
`make verify` exits 0.

| Check | Status | Command | Notes |
| --- | --- | --- | --- |
| Toolchain/module integrity | pending | `go version && go mod verify` | Existing dependency set only. |
| GSD health | passed | `PATH="$HOME/.nvm/versions/node/v24.13.1/bin:$PATH" scripts/gsd doctor` | Node 24; 69 registry commands. |
| Replay package | pending | `go test ./internal/agentloop/... -count=1` | Must include all thirteen fixtures. |
| Replay race | pending | `go test -race ./internal/agentloop/... -count=1` | Required despite pure core. |
| loopctl | pending | `go test ./cmd/loopctl/... -count=1` | Help/status/replay/exit behavior. |
| Driver safety | pending | `bash scripts/tests/auto-loop-control.sh` | Fake local binaries only. |
| Aggregated phase | pending | `make agent-loop-test` | New non-weakening target. |
| Full repository | pending | `make verify` | Authoritative broad gate. |
| Patch integrity | pending | `git diff --check` | Also inspect scoped paths. |

## CLI help/docs/website parity

- Applies to runtime help: yes, for the internal `loopctl` tool.
- Bare command and `--help`: pending tests.
- `pm help`, `docs/cli/**`, generated `pm` manual, and `website/**`: not applicable because no
  `pm` command or public CLI surface changes and issue #325 explicitly excludes those paths.
- This exemption must be repeated in the stacked PR body and worker handoff.

## Optional checks

- Runtime services, provider credentials, network APIs, accessibility, load, and visual checks are
  not applicable to this local dependency-free control-plane slice.
