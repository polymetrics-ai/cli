# Verification checklist — greenhouse documented-operation parity

Phase `greenhouse-parity-sweep-r1`. GSD lifecycle and fallback recorded in
`.planning/traces/gsd-top50-sweep-continue-r1.md`.

## Derivation

- [x] Artifact re-fetched: `developers.greenhouse.io/harvest.html`, HTTP 200, **1,636,662 bytes** —
      byte-identical to the recorded sweep derivation, so it is the same artifact, not a lookalike.
- [x] Extraction **reproduced independently** in this lane: 138 `HTTP Request` declarations, 0
      duplicates, `GET 69 / POST 28 / PUT 8 / PATCH 19 / DELETE 14`.
- [x] Cross-check: 135 `<h2>` sections, 3 of which declare two versioned operations each →
      132 + 3×2 = 138. Two independent methods agree.
- [x] Ledger reconciles exactly (138 = 138); no number adopted without derivation.

## Red before green

- [x] `cmd/connectorgen/greenhouse_api_surface_test.go` written **first** and **run** against the
      unmodified bundle. Verbatim failure captured in `TDD-LEDGER.md` and `RUN-STATE.json`.
- [x] Red state committed on its own (`test(connectors): add red greenhouse surface test…`) before
      any production edit.
- [x] Finding F5 check: no pre-existing `cmd/connectorgen` test referenced greenhouse.
- [x] No test weakened, skipped, narrowed, or deleted.

## Gates

- [x] `go run ./cmd/connectorgen validate internal/connectors/defs/greenhouse` → **0 findings**.
- [x] `go test ./cmd/connectorgen` — the **whole package**, not a targeted `-run` → `ok … 11.372s`.
- [x] `go test ./internal/connectors/commandrunner -run
      TestEveryImplementedCommandPassesRuntimePreflight` → **PASS**.
- [x] `go run ./cmd/connectorgen surface-sync --check` → 551 scanned, **0 filled / 0 corrected**.
- [x] `gofmt -l cmd internal` clean (an inherited lever-hiring violation was fixed in its own commit).
- [x] `go build ./cmd/pm` succeeds.

## Reachability — run, not assumed

- [x] `pm greenhouse` renders a command surface. Before this phase it returned
      `error: unknown command "greenhouse"`.
- [x] **All 127 commands invoked through the built binary**; 0 unreachable. Authoring a command is
      not evidence that it routes — this is the check that caught gmail's 79 phantom operations.
- [x] Spot-checked shapes: a singleton stream exposes `--candidate-id` (not `--config k=v`); the new
      `destroy-candidate-tag plan` renders its destructive risk text and typed-confirmation approval.

## Confinement — the delta is greenhouse only

- [x] `internal/connectors/defs/operation_endpoint_ledger.json` **unchanged**. greenhouse declares no
      `direct_read`/`direct_write`/`binary_download` operation, so there is nothing to ledger.
- [x] No other connector's bundle touched.
- [x] Counts recomputed **from the files themselves**, not read off the generation report: 138 rows,
      138 unique keys, 127 covered, 11 blocked, 0 excluded, 0 blank, 0 dangling `covered_by`, every
      stream and write referenced exactly once.

## Standing constraints

- [x] **No hand-authored paging flags.** Checked with the standing regex → empty.
- [x] Every blocked row carries a machine-checkable `Named dependency:` marker → 11/11.
- [x] No webhook EVENT rows (Greenhouse documents them on a separate page; Harvest exposes no webhook
      management endpoints).

## Known-unmet — recorded, not skipped

- [ ] **CLI help/docs/website parity overlay.** `docs/cli/**`, `website/**`, generated help/manual
      artifacts and golden transcripts are **not** regenerated in this phase. The consolidated-sweep
      model regenerates shared artifacts **once at the end**, and doing it per connector would churn
      ~1,034 files of pre-existing `main` drift (finding F6) on every commit.
- [ ] `TestGoldenTranscripts/root_bare_manual` currently fails on this branch. **Verified
      pre-existing**: it fails with greenhouse stashed out, because chatwoot and gmail already added
      command surfaces the committed transcript predates. It resolves with the end-of-sweep
      regeneration.

Both are sweep-level obligations that must be discharged before the PR merges. They are listed here
so the gap is checkable rather than invisible.
