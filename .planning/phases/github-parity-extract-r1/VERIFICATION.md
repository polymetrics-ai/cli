# VERIFICATION — github-parity-extract-r1

## 1. The shared artifacts were regenerated, not hand-merged

The cherry-pick of `6fe60991d` auto-merged `operation_endpoint_ledger.json`. That merged blob
was **discarded**: the file was reset to `main`'s content (`08cc41c87`) and regenerated.

```
$ git checkout 08cc41c87 -- internal/connectors/defs/operation_endpoint_ledger.json
$ go run ./cmd/connectorgen surface-sync
runtime operation endpoint ledger: updated 1978 endpoint(s)
connectorgen surface-sync: 551 connector(s) scanned, 0 field(s) filled and 0 field(s) corrected
```

The regenerated ledger is **byte-identical** to the one the branch carried (`git diff HEAD`
on that path is empty), which is the strongest available evidence that the branch's ledger was
itself correctly derived and that nothing was lost in the extraction.

Other generators, all run from their own entry points:

| artifact | generator |
| --- | --- |
| `docs/connectors/github/{MANUAL,SKILL}.md` | `./pm docs generate --dir docs/cli` |
| `docs/skills/pm-github/SKILL.md` | `./pm skills generate --dir docs/skills` |
| `website/data/connectors.generated.json`, `website/lib/connectors.catalog.data.generated.json` | `node scripts/gen-connector-bundles.mjs`, `gen-connector-catalog.mjs`, `gen-connectors.mjs` |
| `internal/cli/testdata/golden_transcripts.json` | `POLYMETRICS_UPDATE_GOLDEN_TRANSCRIPTS=1 go test ./internal/cli/ -run GoldenTranscript` |

## 2. The delta is confined to github

Every generated artifact was diffed **per connector / per transcript**, not by eyeballing the
patch:

```
website/data/connectors.generated.json:                 ['GitHub']
website/lib/connectors.catalog.data.generated.json:     ['GitHub']
internal/connectors/defs/operation_endpoint_ledger.json:['github']   (162 -> 164 rows)
golden transcripts:  ['connectors_inspect_github_json', 'dynamic_connector_bare_json']
```

`dynamic_connector_bare_json` is not a false positive — its `args` are `["github", "--json"]`.

The doc generators additionally wanted to rewrite all 551 connectors' `MANUAL.md`/`SKILL.md`
plus four `pm-*` skill files. That is **pre-existing drift on main** (field type annotations
the committed docs render as empty parens, e.g. `created_at()` -> `created_at(string)`, and a
missing `## Icon` section), unrelated to github. It was reverted and left for its own PR rather
than smuggled into a github diff.

## 3. Every reachable command was verified by running the binary

Not inherited from the branch. A prior worker found gmail answering `unknown command` for all
79 operations while the records claimed success, so the figure was re-derived by invoking the
built `pm` binary once per command.

Method: for each of the 1086 github commands with `availability` in {`implemented`, `partial`},
run `pm github <path>` inside an initialised project with no credential configured, and check
the output for `unknown command`. That string is the exact discriminator — a dispatchable
command falls through to `error: missing --credential`, an undispatchable one does not:

```
$ pm github issue list          ->  error: missing --credential      (reachable)
$ pm github nosuchgroup cmd     ->  error: unknown command "nosuchgroup cmd"
```

Each of the 8 workers ran in its own project directory, because a shared one produces
state-lock races that are not reachability failures.

**Result (parity extraction, before the unblock):** `probed=1079 unreachable=0`.
**Result (final binary, after the unblock):** recorded in `RUN-STATE.json`.

No network call is made: with no credential configured every command stops at the credential
gate, ahead of any request.

## 4. Runtime honours every `implemented` claim

`TestEveryImplementedCommandPassesRuntimePreflight` sweeps every bundle in `defs.FS` through
the real `commandrunner.Preflight`, so the 7 restored commands are covered by the repo's own
guard and not only by the test added here. It passes.

## 5. Local gates

```
gofmt -l cmd internal                                   clean
go vet ./...                                            clean
go test ./cmd/connectorgen/                             ok
go test ./internal/connectors/engine/                   ok
go test ./internal/connectors/conformance/              ok
go test ./internal/connectors/commandrunner/            ok
go test ./internal/connectors/hooks/github/             ok
go test ./internal/cli/                                 see RUN-STATE.json
go run ./cmd/connectorgen validate internal/connectors/defs   551 checked, 0 findings
go run ./cmd/connectorgen surface-sync --check          no drift
make lint                                               0 issues
make connector-boundary                                 ok
make agent-contract-check                               contract and projections current
make docs-check                                         ok
go build ./cmd/pm                                       ok
```

`go test ./...` and `make verify` were not run as single commands: per `AGENTS.md`, the full
suite spans 551 connectors and routinely exceeds an agent's per-command timeout, where a cutoff
is indistinguishable from a hang. Scoped package runs plus the individual `make verify` gates
were run instead, and CI carries the whole suite.

## 6. Surface totals

| | main (`08cc41c87`) | this branch |
| --- | --- | --- |
| github commands | 461 | 1147 |
| `implemented` | — | 1049 |
| `partial` | — | 37 |
| reachable (`implemented` + `partial`) | — | 1086 |
| `unsafe_or_disallowed` | — | 5 |
| REST endpoint rows | 505 | 1224 (1220 documented + 4 synthetic close/reopen) |
| covered endpoints | — | 1126 |
| operation-blocked endpoints | — | 98 |
| ledger rows for github | 162 | 164 |
