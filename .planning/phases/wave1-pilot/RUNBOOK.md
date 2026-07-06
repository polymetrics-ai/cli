# RUNBOOK — wave1-pilot (dispatch / gate / rollback)

## Dispatch

1. Preconditions: wave0 closed (SUMMARY.md status completed; REVIEW.md GO), working tree clean on
   `connector-architecture-v2`, `make verify` green at the starting HEAD.
2. DW-0: dispatch P-0 (single sonnet backend+tester pair, TDD). Do not start DW-1 until P-0 is
   merged into the working tree and `make verify` is green — the conformance mirror must be in
   place before the github bundle can be honestly conformance-tested.
3. DW-1: dispatch P-1..P-10 simultaneously (10 sonnet gsd-loop-backend agents, one connector
   each; prompts rendered from the migration-executor template with PLAN.md's per-task row
   inlined). All share the single worktree; safety = disjoint writable dirs (SPEC §6) — do NOT
   run `go run ./cmd/connectorgen gen`, do NOT commit, from any agent.
4. Collect each agent's result JSON + trace. Spot-run the path guard after DW-1
   (`git status --porcelain` vs the union of assigned dirs) BEFORE review, so a stray write is
   caught while agents' context is fresh.
5. DW-2: dispatch P-11 Fable reviews (read-only). Blocker findings → repair agent per connector
   (repair template; max 1 retry, then quarantine: revert that connector's dirs, record in
   `docs/migration/quarantine.json`, connector keeps legacy implementation).
6. DW-3: P-12 (docs patch) and P-13 (cost report) in parallel. DW-4: P-14 wave close.

## Gate (P-14 exact sequence, single-writer)

```
git status --porcelain                                  # path guard vs assigned dirs
go run ./cmd/connectorgen gen                           # hookset regen; registryset must be byte-identical
git diff --stat internal/connectors/registryset/        # MUST be empty (no registration this phase)
go build ./... && go vet ./...
go test ./...                                           # includes paritytest/... and hooks/...
go test ./internal/connectors/conformance -run TestConformance
go test -race ./internal/connectors/hooks/... ./internal/connectors/paritytest/...
make verify                                             # lint + connectorgen-validate + smoke
```

Record outputs in VERIFICATION.md with the HEAD hash (wave0 B3 lesson: run AND record). Then
refresh SUMMARY.md / TDD-GATE.json / RUN-STATE.json, commit once
(`wave 1: pilot — <k>/10 connectors migrated`).

## Rollback

- Everything this phase is additive under `internal/connectors/{defs,paritytest,hooks}/<name>/`
  plus P-0's conformance fix and docs. Rollback unit = one connector: `git checkout -- <its three
  dirs>` (pre-commit) or a single revert of the wave commit path-limited to those dirs
  (post-commit). No registry, catalog, schema-migration, or production-path changes exist to roll
  back; legacy connectors remain registered and untouched throughout (conventions §7 FORBIDDEN
  list) — production behavior is IDENTICAL before/during/after this phase.
- P-0 rollback (worst case): revert the conformance commit; only the new self-test bundle and the
  `github_date_range` assertion branch are affected. Wave0 goldens do not use
  `github_date_range` (verified in REVIEW.md N1), so goldens stay green either way.
- Quarantine (per-connector failure): revert dirs, log to quarantine.json, roster proceeds with
  9/10 — acceptance still measurable (EVAL-PLAN metric 1 counts typed-blocked honestly).

## Stop / escalate

- Same gate failure twice without new evidence → halt wave, Fable inspects for systemic defect
  (orchestration-plan: >30% review failure = template defect, halt).
- Any agent requests go.mod / shared file / other dir → deny, typed blocker, coordinator decides.
- ENGINE_GAP count ≥3 on the same missing feature → pause DW-1 remainder, consider mini engine
  increment (orchestration-plan failure handling) — candidate already on watch: POST-body reads
  (monday) if sentry/others also end up hook-bound for expressibility.
