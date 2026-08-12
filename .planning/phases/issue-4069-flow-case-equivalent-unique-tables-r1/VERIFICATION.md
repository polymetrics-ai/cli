# #4069 Verification checklist

**Status:** Local verification green; no-mistakes, draft PR, and CI are
deliberately pending Firstmate coordination.

## Acceptance evidence

- [x] The RED fixture creates `acme/records` and `globex/RECORDS` with one
      distinct row per exact unique spelling.
- [x] Scoped acme and globex reads return only their own records.
- [x] Generic `SELECT 1` succeeds after GREEN.
- [x] Omitted flow bare-name reads contain
      `*warehouse.AmbiguousTableError` and the manifest `connection` remedy.
- [x] Rejected flows return no rows, write no successful checkpoint, and never
      reveal the other owner's row.
- [x] Existing #4066 generated-owner alias, three-table case variant, exact
      ambiguity, generic, action, reverse/read, and schedule behavior stays
      covered by focused selectors.
- [x] The targeted diff adds no SQL-text filtering, provider credential,
      transport registration, production wiring, or unrelated evidence cleanup
      appears in the child diff.

## GSD/manual lifecycle

- [x] `discuss-phase` prompt resolved and decisions captured.
- [x] `plan-phase --tdd` prompt resolved; PRD/context, plan, ledger, and this
      checklist exist before production edits.
- [x] `execute-phase` RED/GREEN/REFACTOR evidence recorded inline.
- [x] `verify-work` completed after the CPU gate through the documented
      inline/manual fallback; no acceptance gap was found.
- [x] `code-review` completed after the CPU gate through the documented
      inline/manual fallback; no unresolved candidate finding remains.

## Targeted pre-pause checks

- [x] Exact #4069 RED command, recorded as expected failure.
- [x] Exact #4069 GREEN command.
- [x] Short #4066 flow/action/selector/schedule regression selectors.
- [x] `gofmt` and `git diff --check` for the targeted child range.

## Resume-only checks

- [x] Proportionate `internal/app`, `internal/cli`, `internal/flow`, and
      `internal/warehouse` package tests with `-timeout 20m`.
- [x] Full `internal/schedule` package test for the preserved schedule
      re-entry boundary.
- [x] Focused race coverage for #3897/#4066/#4069 selectors.
- [x] GSD verify-work and code-review; no gaps or review fixes were needed.
- [x] `go vet`/configured lint and canonical generator, surface-sync, and
      certification matrix checks.
- [x] Applicable docs/help/website checks and prospective PR issue guard.
- [ ] `no-mistakes` pipeline without `--yes`, exact-head CI, and draft stacked
      PR verification after Firstmate authorizes the heavy CPU lane.

## Completed local command record

The non-secret command/result record is
`traces/resume-broad-verification.txt`. It includes all affected package and
focused race results, direct candidate lint, repository vet/lint, canonical
checks, docs/help applicability, the GSD workflow check, and the local
prospective-body PR issue-guard result.

## Inherited baseline

The audited parent base has seven trailing-whitespace findings in older #3897
planning records. `git diff --check
659efd8a0d69f26b55fcbd3c02150e995c159519..HEAD` passes for this child.
The comparison from `5c92888c996319b41eec6e86ca99fcda4cb365f9` reports only
the documented seven inherited lines in three older #3897 planning files; the
inherited files are not edited or normalized here.

## Correction 1 / 5 gate — same-owner case-equivalent inventory

**Status:** RED required; prior local-green record covers only the cross-owner
matrix and is not acceptance evidence for this correction.

- [ ] GSD plan, ledger, run state, and this checklist record correction 1 / 5
      and policy 1 before a production edit.
- [ ] A committed RED rejects new `records`/`RECORDS` destinations under one
      local-warehouse connection after defaults and before persisted mutation.
- [ ] A committed RED loads legacy state unchanged and proves sync refusal
      before `beginRun`, run/checkpoint/stream state, owner, directory, WAL,
      temporary, or Parquet mutation.
- [ ] A committed RED covers generic/selected bare and quoted SQL collision
      references with the new typed error, never raw DuckDB catalog text or a
      misleading one-owner ambiguity; generic `SELECT 1` remains executable.
- [ ] A committed RED covers unscoped/selected flow and schedule re-entry with
      no success checkpoint, plus exact query/action/reverse refusal for a
      missing physical case variant on the host filesystem.
- [ ] GREEN uses one deterministic ASCII key helper, creation and sync
      inventory validation, declared-inventory plus resolver-snapshot SQL
      policy, and a typed same-owner error without migration, SQL rewriting,
      Unicode folding, flat alias reservation, provider work, or certification
      claim.
- [ ] Cross-owner, `_unattributed`, real-table/generated-alias, action,
      reverse, selector, and schedule controls remain green.
- [ ] Connection/query docs and website guidance, runtime-help applicability,
      docs generator/checks, and PR-body parity evidence are recorded.
- [ ] Inline GSD verify-work and code-review, no-mistakes without `--yes`,
      exact-head CI, and a fresh Sol audit complete before any integration.
