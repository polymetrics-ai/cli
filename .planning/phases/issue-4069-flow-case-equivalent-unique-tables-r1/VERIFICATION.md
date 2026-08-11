# #4069 Verification checklist

**Status:** Planning complete; strict RED pending

## Acceptance evidence

- [x] The RED fixture creates `acme/records` and `globex/RECORDS` with one
      distinct row per exact unique spelling.
- [x] Scoped acme and globex reads return only their own records.
- [ ] Generic `SELECT 1` succeeds after GREEN.
- [ ] Omitted flow bare-name reads contain
      `*warehouse.AmbiguousTableError` and the manifest `connection` remedy.
- [ ] Rejected flows return no rows, write no successful checkpoint, and never
      reveal the other owner's row.
- [ ] Existing #4066 generated-owner alias, three-table case variant, exact
      ambiguity, generic, action, reverse/read, and schedule behavior stays
      covered by focused selectors.
- [ ] No SQL-text filtering, provider credential, connector certification,
      transport registration, production wiring, or unrelated evidence cleanup
      appears in the child diff.

## GSD/manual lifecycle

- [x] `discuss-phase` prompt resolved and decisions captured.
- [x] `plan-phase --tdd` prompt resolved; PRD/context, plan, ledger, and this
      checklist exist before production edits.
- [ ] `execute-phase` RED/GREEN/REFACTOR evidence recorded.
- [ ] `verify-work` completed after the CPU gate.
- [ ] `code-review` completed after the CPU gate.

## Targeted pre-pause checks

- [x] Exact #4069 RED command, recorded as expected failure.
- [ ] Exact #4069 GREEN command.
- [ ] Short #4066 regression selector.
- [ ] `gofmt` and `git diff --check` for the child range.

## Resume-only checks

- [ ] Proportionate `internal/app`, `internal/cli`, `internal/flow`, and
      `internal/warehouse` package tests with `-timeout 20m`.
- [ ] Focused race coverage for #3897/#4066/#4069 selectors.
- [ ] GSD verify-work and code-review, including gaps if found.
- [ ] `go vet`/configured lint and canonical generator, surface-sync, and
      certification matrix checks.
- [ ] Applicable docs/help/website checks and PR issue guard.
- [ ] `no-mistakes` pipeline without `--yes`, exact-head CI, and draft stacked
      PR verification after Firstmate authorizes the heavy CPU lane.

## Inherited baseline

The audited parent base has seven trailing-whitespace findings in older #3897
planning records. The #4069 child range must itself pass `git diff --check`;
the inherited files are not edited or normalized here.
