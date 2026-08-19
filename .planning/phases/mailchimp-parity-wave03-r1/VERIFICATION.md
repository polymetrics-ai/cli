# Mailchimp parity wave03-r1 verification checklist

Required fixture-only gates:

```bash
go run ./cmd/connectorgen validate internal/connectors/defs/mailchimp
go test ./internal/connectors/conformance -run 'TestConformance/mailchimp' -count=1
go test ./internal/cli -run 'Connector|Dynamic|Golden' -count=1
go build ./cmd/pm
make connector-boundary
make verify
git diff --check
```

Additional docs/help checks:

```bash
pm help connectors
pm connectors inspect mailchimp --json
rg -n "Mailchimp|mailchimp" docs/connectors website internal/connectors/defs/mailchimp
```

## Results (re-run after rebase onto `origin/main` @ `d8082031e`, which includes #3890-#3895)

The pre-rebase results were discarded rather than carried forward: main moved a long way (direct-read
runtime preflight #3890, rate-limit declaration/admission, API-surface provenance, typed multipart
writes), so the earlier trace logs no longer described this tree.

- PASS `go run ./cmd/connectorgen validate internal/connectors/defs/mailchimp`: `1 connector(s) checked, 0 findings`.
- PASS `go test ./internal/connectors/conformance -run 'TestConformance/mailchimp' -count=1`: `ok ... 7.422s`.
- PASS `go test ./internal/cli -run 'Connector|Dynamic|Golden' -count=1`: `ok ... 302.007s`.
- PASS `go test ./internal/connectors/... -count=1`: 132 packages ok, no failures.
- PASS `go test ./internal/connectors/commandrunner -run 'TestEveryImplementedCommandPassesRuntimePreflight' -count=1`.
- PASS `go build ./cmd/pm`; `go vet ./...` clean; `gofmt -l cmd internal` empty.
- PASS `make connector-boundary`: `outcome: clean`, `checked_files: 182`, `findings: []`.
- PASS `make verify`: exit 0.
- PASS `git diff --check`: no whitespace errors.

All 298 operations re-confirmed dispositioned after the rebase: 295 `covered_by` (79 stream / 148
write / 68 direct_read), 3 blocked each carrying both a reason and an official `source_url`, and
**0 undispositioned**. The audit trace and `api_surface.json` remain an exact 1:1 set match at 298.

### Runtime preflight regression found and fixed this round

`TestEveryImplementedCommandPassesRuntimePreflight` initially reported *68 of 1617 commands marked
"implemented" fail runtime Preflight* — every Mailchimp direct read, all with the same cause: the
runtime operation endpoint ledger still held `"mailchimp": []`, so no `GET <path>` / `rest_read` /
`max_bytes 1048576` entry existed for them. `operations.json` was already correct (68 `rest_read` ops,
all with `max_bytes`); only the derived ledger was stale.

Fixed by running the generator that owns that file — `go run ./cmd/connectorgen surface-sync` — not by
hand-editing it, per the AGENTS.md rule against hand-editing derivable command metadata. Result:
`runtime operation endpoint ledger: updated N endpoint(s)`, adding exactly Mailchimp's 68 entries,
and the diff is a single hunk replacing `"mailchimp": []`. `surface-sync --check` is now clean.

### Rebase onto `d8082031e` (#3892 dynamic schema discovery, #3894 notion, #3895 gong)

All three rebase conflicts were in generated artifacts — `golden_transcripts.json`,
`connectors.catalog.data.generated.json`, `connectors.catalog.generated.ts`. No `.go` file
conflicted, so nothing was resolved by picking a side in real code. Each was resolved by taking
main's side and then regenerating from the generator, and each regenerated diff was inspected:

- Golden transcripts: 9 lines, all additive, all the `pm mailchimp` namespace line.
- Website catalog: semantic diff shows exactly one changed key, `mailchimp`, none added or removed.
- Endpoint ledger: auto-merged; one changed key (`mailchimp`, 68 entries), 1866 → 1934 total,
  retaining main's `notion` (18) and `gong` (13).

Two pieces of main-side drift surfaced and were handled differently, deliberately:

- `pm docs generate` rewrote MANUAL.md/SKILL.md for ~515 connectors because main's doc generator now
  renders field types (`gid(string)` rather than `gid()`) and main's committed docs predate it. All
  1030 unrelated files were reverted; only mailchimp's two were kept, where the new rendering is
  legitimately part of this connector's docs.
- `docs/connectors/catalog/all-connectors.{json,md}` regenerated with one extra line beyond
  mailchimp: `gong` gains the `upload_target_assignments` write action (26 → 27) that #3895 added to
  the bundle without refreshing the committed catalog. This one was **kept**, because the catalog is
  a single generated file — emitting a correct mailchimp entry while preserving a known-stale gong
  entry would require hand-editing generated output, which AGENTS.md forbids. It is a one-line
  mechanical correction of main's own staleness, not connector work.

### Recorded conflict: `connectorgen ownership` vs `connectorgen surface-sync`

These two checks disagree about `internal/connectors/defs/operation_endpoint_ledger.json`, and the
disagreement is recorded here rather than silenced. Neither check was weakened, skipped, or edited.

- `connectorgen surface-sync --check` (a gate inside `make verify`) **fails** if the ledger is not
  updated: reverting it to main's version yields `runtime endpoint ledger drift=true`, exit 1.
- `connectorgen ownership --base origin/main` (advisory; invoked by no Makefile target and no CI
  workflow) **reports** the same file as a shared path outside the `mailchimp` target.

A connector that adds operations is required to update the ledger, so the required gate is the
authoritative one and the advisory allowlist (`isNarrowSharedOwnershipOutput` in
`internal/connectors/boundary/ownership.go`) is what is out of date — it already lists the directly
analogous `docs/connectors/catalog/all-connectors.json` and `website/data/connectors.generated.json`.

Precedent: PR #3891 merged to `main` on 2026-08-07, and its entire ledger change was the single line
`+  "crisp": [],` — a connector's alphabetic entry added after `surface-sync` reported drift. The
CI-enforced guard, `connectorgen boundary`, is clean on this branch.

Ledger diff verified mechanical before every commit: against the rebase base, zero connectors added,
zero removed, exactly one key changed (`mailchimp: [] → 68 entries`). On the current base
`d8082031e` the ledger carries 551 connectors / 1934 entries (1866 + 68), retaining main's `crisp`,
`notion` (18) and `gong` (13) alongside mailchimp's 68.

### Executability verified by running the built binary, not by reading files

- `./pm connectors inspect mailchimp --json` → `streams: 79`, `write_actions: 148`.
- `./pm mailchimp` renders the namespace help and exits 0.
- `./pm mailchimp lists --help` lists the ETL stream command *and* all 13 `lists ...` direct reads in
  one merged group — confirming the duplicate group label in the index is cosmetic and shadows nothing.
- Inside a scratch `pm init` project, `lists get` (direct read), `lists tag-search get` (search), and
  `lists list` (stream) each reach `missing --credential`, proving they resolve and clear preflight
  rather than failing as unknown or unimplemented. No credential was supplied and no provider call made.
- The destructive gate was verified through the fixture layer rather than by fabricating a credential
  to drive `delete-lists`, which would have risked a live mutation.

### Fixture coverage

Verified programmatically, both directions (no missing, no orphans): streams **79/79**, writes
**148/148**. Direct reads have no fixture slot in the conformance harness and no connector in the repo
ships one, so they are covered by `connectorgen validate` plus the runtime preflight guard.

Docs/help parity checks performed:

- `./pm docs generate --dir docs/cli` regenerated connector docs. It also rewrote unrelated
  connectors' `{MANUAL,SKILL}.md` from pre-existing main-side generator drift; that drift was
  reverted every round and is not in this commit. See the rebase section above for the current
  round's scale (~515 connectors / 1030 files).
- `pnpm run gen:website-data` under `website/` regenerated the catalog. A semantic diff of
  `connectors.catalog.data.generated.json` confirms exactly one changed key — `mailchimp` — with no
  connectors added or removed.
- `POLYMETRICS_UPDATE_GOLDEN_TRANSCRIPTS=1 go test ./internal/cli -run 'Golden' -count=1` refreshed the
  CLI golden transcripts; the diff is 9 lines, all additive, all the new `pm mailchimp` namespace line.
