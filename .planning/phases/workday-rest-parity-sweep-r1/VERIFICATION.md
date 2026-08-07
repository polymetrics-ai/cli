# workday-rest — verification

Branch `fm/cli-top50-sweep-resume2-r1`. Every gate below was **run**, and its real output is quoted.

## TDD

| | |
| --- | --- |
| Red written and RUN before authoring | ✅ `TDD-LEDGER.md` cycle 1 (slice 1) and **cycle 2** (this slice, at the corrected 907) |
| Red captured verbatim | ✅ `RUN-STATE.json.tdd.red_failure` |
| Assertion tightened, never weakened | ✅ 916 → 907 **and** a new requirement that each collapsed variant's base endpoint is *present*, so a double-count cannot be "fixed" into a missing operation |
| Green | ✅ `ok polymetrics.ai/cmd/connectorgen 0.514s` |

## Gates

| Gate | Result |
| --- | --- |
| `gofmt -l cmd internal` | clean |
| `go vet ./cmd/connectorgen/ ./internal/connectors/commandrunner/` | clean |
| `connectorgen validate` | `551 connector(s) checked, 0 findings` |
| `connectorgen surface-sync --check` | `551 scanned, 0 filled, 0 corrected` |
| **whole** `cmd/connectorgen` package (finding F5) | `ok … 10.609s` |
| `TestEveryImplementedCommandPassesRuntimePreflight` | `ok … 7.679s` |
| `pm docs validate --connectors-dir docs/connectors` | `Validated connector docs in docs/connectors` |
| `make tidy-check` | clean |
| `make connector-boundary` | `"findings": []` |
| `make agent-contract-check` | `canonical contract and registered projections are current` |
| `make release-workflow-check` | `homebrew release notification assertions passed` |

## Reachability — 911 of 911, by running the binary

```
commands: 911  failures: 0
```

The probe asserts the rendered `NAME` line, **not** the exit code: `pm <connector> <nonsense> --help`
renders group help and **exits 0**, so an exit-code-only probe reports passes it never verified
(finding 30). Spot-checks:

```
pm workday-rest wql read-data - Returns the data from a WQL query.
pm workday-rest staffing update-workers-by-id-check-ins - Partially updates an existing Check-In instance.
pm workday-rest common create-workers-by-id-business-title-changes
  --type (enum): … values=me maps_to=record.type
```

## Endpoint-ledger delta — inspected BY OBJECT, not by line (finding 8)

```
connectors before 551  after 551
connectors added []    removed []
connectors CHANGED: ['workday-rest']   entries 0 -> 7
```

551 connectors before and after, none added, none removed, **exactly one changed** — from 0 to the 7
`rest_read` entries the POST-shaped reads require. `git diff --stat`: 44 insertions, 1 deletion.

## Artifact reproduction

| | |
| --- | --- |
| Manifest | HTTP 200, **617,538 bytes** — byte-identical to slice 1 and to the sweep derivation |
| Service specs | all **52** re-fetched (5.8 MB) |
| Derivation reproduced | 920 raw · `GET 655 · POST 154 · PATCH 58 · DELETE 33 · PUT 20` · 916 after cross-service dedup |
| New this slice | −9 query-string variants → **907** |

The byte-identical manifest is what proves it is the same artifact rather than a lookalike, and it
is why slice 1's derivation could be **reproduced** rather than trusted — and why the divergence at
907 is a finding rather than a disagreement.

## Rules held

- ✅ **No hand-authored paging flags.** The generator raises rather than authoring one; `PAGING_PARAMS`
  is a blocklist covering the path-variable sweep too. `streams.json` already declares this
  connector's `page_number` pagination for the foundation lane to derive from.
- ✅ **Webhook events excluded** — the directory documents none (`webhook_events: 0`).
- ✅ **No wildcard or query-string rows.** The red test rejects `" ?*"` in any path.
- ✅ **Every row carries exactly one disposition**; the one blocked row carries a
  `Named dependency:` marker and `model: deprecated`.
- ✅ **Generated files were never hand-merged.** Every fix went into
  `tools/gen_workday_rest.py` and the bundle was regenerated from
  `git checkout -- internal/connectors/defs/` — three times (schema keys, the misleading `create-*`
  naming, and the collapsed `?type=` hardcoding).
- ✅ **No test weakened, skipped, or deleted.**

## Known unmet — carried, not hidden

1. **`TestGoldenTranscripts` — 11 subtests fail, all pre-existing.** Proven by stashing this slice and
   re-running against the branch tip: the failing set is **identical** (`root_bare_manual`,
   `root_long_help`, `root_short_help`, `root_help_command`, `root_man_command`, `root_json_help`,
   `root_late_json_help`, `root_equals_form`, `root_space_form`, `connectors_inspect_github_json`,
   `dynamic_connector_bare_json`). They are github manual/root-help drift and are discharged by the
   end-of-sweep regeneration, which must happen before the PR merges. **The handoff recorded this as
   one subtest; it is eleven.**
2. **Website catalogs not regenerated** — a shared artifact the program regenerates once at the end.
   No `make verify` gate covers them.
3. **`internal/cli` was not run to completion at this tree.** It takes ~13 minutes and its only
   failures are the pre-existing golden transcripts above, established against both trees. CI carries
   the full suite at `-timeout 20m`; a bare `go test` uses the 10m default and dies mid-run, which
   looks like a hang and is not one.
