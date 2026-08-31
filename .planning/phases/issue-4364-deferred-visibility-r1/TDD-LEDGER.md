# TDD ledger — issue 4364

| Slice | Red evidence | Green evidence | Edge/refactor evidence |
| --- | --- | --- | --- |
| Actual cohort visibility | `GOCACHE=/private/tmp/gocache-4364-deferred-visibility-red go test -count=1 ./cmd/connectorgen -run '^TestDeferredVisibilityBatchR1Cohort$'` failed red: `connectorgen: unknown subcommand "deferred-visibility"`. | `GOCACHE=/private/tmp/gocache-4364-deferred-visibility-green go test -count=1 ./cmd/connectorgen -run '^TestDeferredVisibility'` passed. It checks all ten matrices, the frozen 4,341 primary operations, 4,343 rows, 30,401 cells, all M-U/M-F entries, source facts/citations, and zero executable declarations. | JSON output is deterministically sorted by connector, matrix source ID, and fixed lane order. |
| Explicit source-to-lock identity | Before the generic resolver, the real green attempt failed because GitLab’s matrix identity `gitlab.rest.deleteApiV4AdminActiveContextDeadQueue` differs from the lock key `deleteApiV4AdminActiveContextDeadQueue`. | Unit tests prove direct lock identity and explicit provider-operation-ID + exact method/path identity; the real cohort test proves the GitLab pair is preserved in output. | Missing opaque identities, route mismatch, and zero-or-multiple exact provider-ID matches fail; no name-prefix or method heuristic is used. |
| Reason/capability and execution boundary | The CLI usage test first had no command; the red cohort test made the absent behavior observable. | `TestDeferredVisibilityCLIUsage` checks `--help` and rejects a missing `--check`; the cohort test checks stable M-U/M-F typed reasons and known Atlas capabilities. | The report boundary rejects executable root/entry fields while allowing source facts to retain provider semantic keys such as `write`; no runtime import/materialize/project/transport call exists. |

## Planned commands

```sh
GOCACHE=/private/tmp/gocache-4364-deferred-visibility \
  go test -count=1 ./cmd/connectorgen -run 'TestDeferredVisibility'
go run ./cmd/connectorgen deferred-visibility \
  data/connector-canon/batch1-source-operation-mapping-cohort.json --check --json
go run ./cmd/connectorgen source-operation-mapping-cohort \
  data/connector-canon/batch1-source-operation-mapping-cohort.json --check
go run ./cmd/agentcontractgen check
git diff --check
```

No credentialed or provider-I/O test is permitted. A focused race check is
pending the shared heavy-test slot and disk floor.
