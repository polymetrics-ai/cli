# Verification — zendesk-support parity sweep

Every command below was run in this worktree on 2026-08-07 and its real output is quoted. Nothing
here is a claim that was not measured.

## Shared gates — all green

```
$ gofmt -l cmd internal
(no output)

$ go vet ./cmd/connectorgen/ ./internal/connectors/...
(no output)

$ go run ./cmd/connectorgen validate
connectorgen validate: 551 connector(s) checked, 0 findings

$ go run ./cmd/connectorgen surface-sync --check
connectorgen surface-sync: 551 connector(s) scanned, 0 field(s) filled and 0 field(s) corrected across 0 connector(s)

$ go test ./cmd/connectorgen/
ok  	polymetrics.ai/cmd/connectorgen	11.712s

$ go test ./internal/connectors/commandrunner/ -run TestEveryImplementedCommandPassesRuntimePreflight
ok  	polymetrics.ai/internal/connectors/commandrunner	2.770s
```

The whole `cmd/connectorgen` package was run, never a targeted `-run` (finding F5: gong had a
second surface test a targeted run missed).

`surface-sync --check` is clean **on generation**, not after a correcting run. Its first pass
corrected an `output_policy` the generator had emitted on the binary-download command; the generator
stopped emitting it rather than the correction being accepted, because a field the runtime strips is
the hand-maintained drift `AGENTS.md` forbids.

## Reachability — measured by running the binary

```
$ go build -o /tmp/pm-zd ./cmd/pm
$ python3 -c "... implemented+partial command paths ..." > /tmp/zd-cmds.txt
511 commands
$ PM_BIN=/tmp/pm-zd PM_CONNECTOR=zendesk-support \
    xargs -P 12 -I{} tools/probe_reachability.sh "{}" < /tmp/zd-cmds.txt
failures: 0
```

**511 of 511.** Exit status proves nothing here — a namespace miss renders group help and exits 0
(finding 30) — so the probe asserts the rendered `NAME` line. It was checked against a bogus path
first to prove it discriminates rather than passing everything:

```
$ probe_reachability.sh "operations definitely_not_a_real_command"
FAIL(unrouted): operations definitely_not_a_real_command ::   pm zendesk-support - Zendesk Support command surface

$ probe_reachability.sh "operations get_sources_by_target"
(silent — routed)
```

Bare namespace behaviour, per the CLI parity reference:

```
$ pm zendesk-support ; echo $?
0        # renders the contextual command surface, does not error
```

## Operation endpoint ledger — delta confirmed BY OBJECT, not by line

```
connectors before 551  after 551
added    []
removed  []
changed  []
```

`internal/connectors/defs/operation_endpoint_ledger.json` is byte-identical after
`connectorgen surface-sync` (`git status --short` reports it unmodified), even though the run
recomputed 1,998 endpoints. Promoting an already-declared operation to a reachable command changes
no `rest.path`, so it adds no ledger entry — which is also why this phase cannot have perturbed
another connector.

## internal/cli — one pre-existing failure set, MEASURED as unchanged

```
$ go test -timeout 20m ./internal/cli/
FAIL	polymetrics.ai/internal/cli	606.895s
```

`-timeout 20m` is mandatory: the bare form inherits Go's 600s default and dies mid-run (finding 36).

The failure is `TestGoldenTranscripts`, which has failed since before github and is discharged by
the **end-of-sweep** regeneration, never per connector (finding F6: a per-connector
`pm docs generate` rewrites ~1,031 files of pre-existing `main` drift).

**Proved pre-existing rather than assumed.** The failing set was captured on this tree, then the
phase's changes were reverted in place (`git checkout HEAD~3 -- internal/connectors/defs/
cmd/connectorgen/`), the same test re-run, and the two sets diffed:

```
$ diff /tmp/golden-before.txt /tmp/golden-after.txt
(no output)
IDENTICAL failing sets — this phase adds zero new golden-transcript failures

baseline count: 11    current count: 11
```

Both sets are exactly:

```
TestGoldenTranscripts/connectors_inspect_github_json
TestGoldenTranscripts/dynamic_connector_bare_json
TestGoldenTranscripts/root_bare_manual
TestGoldenTranscripts/root_equals_form
TestGoldenTranscripts/root_help_command
TestGoldenTranscripts/root_json_help
TestGoldenTranscripts/root_late_json_help
TestGoldenTranscripts/root_long_help
TestGoldenTranscripts/root_man_command
TestGoldenTranscripts/root_short_help
TestGoldenTranscripts/root_space_form
```

Eleven, matching the handoff's corrected list exactly. The quoted diffs are github command-group
names, from github's own delivery. The tree was restored to `HEAD` afterwards and verified clean.

## GSD gate

```
$ GSD_BASE_REF=origin/main ./scripts/verify-gsd-workflow
exit 0
```

`scripts/gsd` is a **Node** script — `bash scripts/gsd` dies with a syntax error. `sources` for
`plan-phase`, `execute-phase` and `verify-work` all resolve to `.gsd/commands.json`,
`.gsd/upstream.lock.json`, `.gsd/official-docs/COMMANDS.md`.

## Artifact reproduction

```
$ curl -sS -w 'HTTP %{http_code} bytes=%{size_download}\n' https://developer.zendesk.com/zendesk/oas.yaml
HTTP 200 bytes=1701930
```

Byte-identical to `MASTER-PLAN.json`, so the derivation was reproduced rather than trusted:
434 paths, 625 operations, GET 325 / POST 111 / PUT 89 / DELETE 86 / PATCH 14, zero rows containing
`?`, `*` or a space, zero repeated (method, path) pairs, zero deprecated.

## Known-unmet, carried deliberately

- `TestGoldenTranscripts` — 11 subtests, pre-existing, measured above. Discharged by the
  end-of-sweep artifact regeneration before the sweep PR merges.
- `docs/cli/**` and `website/**` parity for this connector's new commands — regenerated once at the
  end of the sweep with the other shared artifacts, per the handoff, never per connector.
