# TDD ledger — issue #3714 parent readiness

| ID | Contract | RED evidence | GREEN evidence | Refactor / verification |
|---|---|---|---|---|
| R1 | Parent PR head contains current `origin/main` | `git merge-base --is-ancestor origin/main HEAD` exited nonzero before integration | `git merge --no-edit origin/main` created `fc99e1836`; the ancestry assertion then exited zero | `git log` shows #3730, #3731, and #3726 beneath the parent merge commit |
| R2 | Every required Codex, Claude, and Pi projection matches the canonical source | No drift was found to repair; projections were not edited by hand | `go run ./cmd/agentcontractgen sync` synchronized `0` projections; `go run ./cmd/agentcontractgen check` passed | `go test ./internal/agentcontract ./cmd/agentcontractgen` passed after a serial retry |
| R3 | Mainline destructive-write confirmation behavior survives the integration | Integration used Git's clean `ort` merge; no #3730 path or test was altered by conflict resolution | `go test -p 1 ./internal/app`, the focused CLI tests, `./internal/connectors`, and `./internal/connectors/engine` passed | `go vet` on affected packages, connector gates, lint, docs validation, and smoke passed; remote full CI remains pending |
| R4 | Parent head continues to contain the current `origin/main` after it advances | At `554235545`, `git merge-base --is-ancestor origin/main HEAD` returned nonzero for `4871df2f8` | `git merge --no-edit origin/main` created clean merge `08377a5ae`; the ancestry assertion then passed | `agentcontractgen sync` changed 0 projections; its check, clean Pi-agent check, focused Recurly/engine tests, vet, and scoped gates passed; a fresh no-mistakes run remains |

The initial parallel focused-test attempt ran out of temporary disk space. After the volume was
restored, fresh serial retries passed; the capacity error is not recorded as a product failure.
Remote parent CI is still required before readiness is claimed.
