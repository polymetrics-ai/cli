# #4069 Verification checklist

**Status:** Correction 1 local/GSD verification green; the existing draft PR
is deliberately untouched, and no-mistakes/push/exact-head CI remain pending
an explicit Firstmate instruction.

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

## Completed local command records

The historical cross-owner record remains
`traces/resume-broad-verification.txt`. The active correction's non-secret
command/result record is `traces/correction-1-broad-verification.txt`; it
covers affected package and focused-race results, candidate formatting/vet/
lint/tidy/build/smoke, canonical checks, docs/help/website parity, the GSD
workflow check, and the local issue-guard result.

## Inherited baseline

The audited parent base has seven trailing-whitespace findings in older #3897
planning records. `git diff --check
659efd8a0d69f26b55fcbd3c02150e995c159519..HEAD` passes for this child.
The comparison from `5c92888c996319b41eec6e86ca99fcda4cb365f9` reports only
the documented seven inherited lines in three older #3897 planning files; the
inherited files are not edited or normalized here.

## Correction 1 / 5 gate — same-owner case-equivalent inventory

**Status:** local/GSD verification green; the remaining delivery gates are
no-mistakes without `--yes`, exact-head CI, and a fresh independent Sol audit.

**Canonical finish-plan SHA-256:**
`939f14f61defd993f8ad0335a5aeb617d97083c9f73a6a75259d0e312ae8f408`

- [x] GSD plan, ledger, run state, and this checklist record correction 1 / 5
      and policy 1 before a production edit.
- [x] A committed RED rejects new `records`/`RECORDS` destinations under one
      local-warehouse connection after defaults and before persisted mutation.
- [x] A committed RED loads legacy state unchanged and proves sync refusal
      before `beginRun`, run/checkpoint/stream state, owner, directory, WAL,
      temporary, or Parquet mutation.
- [x] A committed RED covers generic/selected bare and quoted SQL collision
      references with the new typed error, never raw DuckDB catalog text or a
      misleading one-owner ambiguity; generic `SELECT 1` remains executable.
- [x] A committed RED covers unscoped/selected flow and schedule re-entry with
      no success checkpoint, plus exact query/action/reverse refusal for a
      missing physical case variant on the host filesystem.
- [x] A committed RED retains exact query/action/reverse resolution for only
      host-resolver-visible physical spellings, and rejects a missing spelling
      without case aliasing it.
- [x] Targeted GREEN uses one deterministic ASCII key helper, creation and sync
      inventory validation, declared-inventory plus resolver-snapshot SQL
      policy, and a typed same-owner error without migration, SQL rewriting,
      Unicode folding, flat alias reservation, provider work, or certification
      claim.
- [x] Cross-owner, `_unattributed`, real-table/generated-alias, action,
      reverse, selector, and schedule controls remain green. The same-owner
      real `RECORDS__<owner-id>` control also confirms the real table wins.
- [x] Connection/query docs and website guidance, runtime-help applicability,
      docs generator/checks, and PR-body parity evidence are recorded.
- [x] Inline GSD verify-work and code-review completed through the documented
      manual fallback with no candidate gap or finding.
- [ ] No-mistakes without `--yes`, exact-head CI, and a fresh Sol audit
      complete before any integration.

### Targeted GREEN command record

`traces/correction-1-same-owner-green.txt` records the focused C1–C5 suite and
the existing #4066/#3897 app/flow selector fence. The independent broad record
is `traces/correction-1-broad-verification.txt`; it covers affected packages,
focused race, quality/generator/release, docs/help/website, local issue-guard,
and manual GSD verification. It also separates the inherited website lint
warnings from the candidate result. No no-mistakes, push, PR mutation,
exact-head CI, or Sol-audit claim is inferred from either local record.

## Authorized recovery gap — Website aggregate parity

- [x] Captain authorization, exact #4071 topology, clean registered worktree,
      remote `9a5b23fe` custody fast-forward, and ancestry preconditions
      verified before the recovery mutation.
- [x] Exact-head Website CI RED recorded from runs `31579519649` and
      `31579519744`; both regenerate the stale aggregate and reject it as a
      dirty generated file.
- [x] The no-install generator changes only `website/lib/docs.generated.ts` to
      the audited SHA-256.
- [x] A second generator invocation is idempotent and candidate diff scope/check
      pass. Lint (13 inherited warnings), typecheck, unit, script, and build
      checks pass without dependency installation; browser E2E remains for
      natural CI because its workflow-owned browser install is out of scope.
- [x] Inline/manual GSD verify-work and code-review find no candidate gap or
      finding: deterministic aggregate content is proven by the expected hash,
      idempotence, source-generator ownership, and affected automated checks.
- [ ] One fresh no-mistakes run records PR/CI as explicitly skipped and pushes
      only the existing #4069 branch.
- [ ] Existing #4071 remains the sole open, draft, stacked PR and all natural
      exact-head required checks are green before a fresh independent Sol audit.

## Correction 2 / 5 — flow manual generator parity

**Status:** planned before source edits. The existing #4071 Verify failure is
the causal RED; it must not be rerun remotely.

- [x] Exact local `TestGoldenDocsGenerateMatchesTrackedCLIManuals` RED exits 1
      with `flow.md` drift and is recorded before an embedded-help edit.
- [x] `internal/cli/docs.go` is the only hand-authored manual source changed;
      the three approved flow lines and same-commit `connectionsHelp` wording
      are generated into their tracked CLI manuals.
- [x] Golden transcript update path regenerates flow help/bare/JSON artifacts;
      focused transcript and manual tests pass.
- [x] The checked-in website data generator is run twice; its exact scoped
      output or no-change result is recorded without altering website sources.
- [x] Focused CLI/manual, relevant Go, docs/generator/lint, and candidate
      diff checks pass; inherited findings are separate from candidate output.
- [x] Inline/manual GSD verify-work and code-review record no remaining gap.
- [ ] One new no-mistakes run without `--yes` records PR/CI as skipped and
      pushes only the existing #4069 branch through supported custody.
- [ ] Existing draft #4071 naturally reaches green exact-head checks before a
      fresh independent Sol audit; no workflow is rerun manually.

### Correction 2 local/GSD verification

The focused golden/manual selector and full `internal/cli` package passed with
`-timeout 20m`. `go vet ./...`, direct and configured Go lint, `gofmt`,
`go build ./cmd/pm`, `make docs-check-no-build`, the agent-contract check, and
the canonical connector validation/surface-sync checks passed. The checked-in
website generator was idempotent; Website lint exited 0 with its 13 inherited
warnings, while typecheck, unit, script, and build checks passed. The candidate
range from `678e294568a8a010a460ecb05fe11a42e1eb40f2` through the minimum
GREEN commit changes only the source-owned help string, its golden transcripts,
and #4069 planning evidence; `git diff --check` is clean.

`scripts/verify-gsd-workflow` passed against `origin/main`, confirming the
implementation files have #4069 GSD/TDD evidence. The resolved `verify-work`
and `code-review` prompts were executed inline under the recorded single-worker
fallback and found no acceptance gap or candidate finding. Exact commands,
results, generator idempotence, and the separate inherited-warning disposition
are in `traces/correction-2-flow-manual-broad-verification.txt`.

## Correction 3 / 5 — destination-scoped legacy collision admission

**Status:** local/GSD verification green at evidence checkpoint
`5a4f3d23645d3046d02055273fedd9a8b9dd67d1` (its only addition after the
verified production commit is this correction's GSD evidence); child-PR
delivery remains pending and #4071 is unchanged.

- [x] Restored non-local ETL regression fails against exact source head
      `3b75f4a62fd8d743ec883a5b824164374f661857` with the current typed
      same-owner local-warehouse error before the unrelated source or
      destination executes.
- [x] The production preflight is narrowed only through the existing
      `connectionMaterializesLocalWarehouse` abstraction; no connector name or
      warehouse literal is introduced in shared code.
- [x] Restored non-local ETL regression passes with one loaded record, one
      destination acknowledgement, and one source request while the legacy
      local collision remains present.
- [x] Existing local true-positive regression remains a typed
      `*warehouse.SameOwnerCaseEquivalentTableError` refusal before `beginRun`
      and state mutation.
- [x] Full `internal/app` tests and proportionate quality/build/GSD-workflow
      gates pass.
- [x] Inline/manual `verify-work` and `code-review` record no acceptance gap
      or unresolved finding.
- [ ] The delivery owner must create the required child PR with base
      `fix/4069-flow-case-equivalent-unique-tables-r1`, then record its actual
      head SHA in `RUN-STATE.json` and the published child PR body without
      changing #4071.
