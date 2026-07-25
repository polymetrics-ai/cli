# Issue #397 PM First-Round Review System Verification

**Status:** captain-authorized advisory F1–F6 correction full GREEN on the uncommitted candidate; exact clean commit + self-review pending; correction budget remains 2/5

> Historical round-0/1/2 checklist detail is condensed below; the current F1–F6 correction and delivery gates are authoritative. Earlier focused-fixture claims are pre-review evidence, not final acceptance.

## Historical rounds (condensed; full detail in git history and REVIEW-R1/R2 artifacts)

- [x] Setup: isolated worktree, clean status inventory, verified remote base `0f8c964ba9cfbe1b1eec8e7998eacf4158ef0e20`, clean branch creation, no reset/rebase/amend/force (`SETUP-EVIDENCE.md`).
- [x] Scope: diff disjoint from all PR #493-owned paths; within the `PLAN.md` positive allowlist; no CLI/subcommand/dependency/connector/reverse-ETL behavior change; PR #495 history preserved truthfully.
- [x] Round-0 RED→GREEN: readiness-gate, disposition-parser, prose-dependency, transitive-prohibited-target, lineage-reset, unsafe-reference, cross-format-closure, packet-coverage/overflow/threshold, and frozen opaque-corpus cases pass; clean controls stay clean.
- [x] Compiler: active closure with edge reasons, authority writers/readers/mirrors, parsed dispatch, exact base/head in every packet, conservative/mandatory split thresholds, block-not-truncate, one PM synthesis, downstream Shepherd.
- [x] Algorithm research: `ALGORITHM-RESEARCH.md` selected typed multigraph + forward/reverse adjacency + relation-policy BFS with measured tradeoffs and no new dependency.
- [x] Bidirectional impact graph: seeds = changed + canonical roots; typed edges with three-valued certainty; Markdown/JSON/YAML/shell/Python/Go coverage; cycle-safe bounded BFS; complete impact coverage or block; practical-impact-only disclosure.
- [x] Counterfactual lab: detached disposable exact-head copies; scrubbed env; canonical/outside/symlink/network/shell/commit/install/live/deploy denial; proven OS sandbox or block; time/process/disk/output bounds; before/after identity proof; discriminating hypotheses; versioned contracts.
- [x] Codex round 1 at `b1d869732d230575ab7c8295b15cef42cc0078ef`: 17/17 packets, 89 findings, dispositioned in `REVIEW-R1-DISPOSITION.md`; R1-A..R1-O RED→GREEN; correction round 1/5; Shepherd not run.
- [x] Round 2 at `92ce5e6a1`: 41/41 responses, 141 findings, `REVIEW-R2-DISPOSITION.md`; R2-A..R2-O/N1..N6 RED→GREEN; lineage advanced to 2/5; operational provider retries excluded from rounds.
- [x] Round-2 route/docs/state parity and fresh-runtime lab-fixture recovery (aggregate `changed_paths`, `__CF_USER_TEXT_ENCODING`, `internal/lib` path, numeric loopback service, `local_only`) all closed; full repository verification passed on the uncommitted candidate.
- [x] Bounded-efficiency replacement reached 63 packets with every rendered prompt ≤30,000 bytes via exact partitioning, cross-role best-fit, and assignment-preserving coalescing (no truncation, no bound reduction); committed `c28126bb3` and reconfirmed clean exact head/tree with full verification.

## Captain-authorized advisory F1–F6 correction

- [x] Captain authorization, advisory report, and correction/self-review instructions read in full.
- [x] Advisory path is recorded as non-R3 and non-budget-consuming; correction usage stays 2/5.
- [x] Partial prospective R3 at `c28126bb3` was halted without synthesis; Shepherd was not run.
- [x] `scripts/gsd doctor` and `scripts/gsd list` pass; absent `programming-loop` manual fallback recorded.
- [x] Before baseline captured: ready, 63 packets, max 29,994 rendered bytes, 1.018025x duplication, derived headroom 1, 14.24s, 285,360,128-byte max RSS.
- [x] RED F1: accepted `combined` domain assignment stayed ready instead of named nonzero `packet_coverage` block.
- [x] RED F1 byte proof: permanent assertion requires all changed files' packed diff bytes equal exact Git diff bytes.
- [x] RED F2/F3: renderer lacked honest held-contract assumption and actionable-vs-residual-risk criterion.
- [x] RED F4: packets/responses lacked per-impact-file provided/total bytes and exact synthesis echo.
- [x] RED F5: manifest access raised `KeyError: packet_headroom`; comment still overclaimed two slots.
- [x] RED F6: replay duplicated deterministic occurrence ids across four roots.
- [x] GREEN F1–F6 focused regressions pass with unchanged 30,000-byte/64-packet hard limits.
- [x] Bounded perf pass (prefilter, edge-slice cache, account early-stop) proven byte-identical output on a frozen tree; clawed back a ~15.8s pre-opt regression to ~14.6s; three compiles share one semantic-manifest digest.
- [x] Packet-budget honesty: with the required honest F2/F3 prompt the accumulated candidate compiles ready within the unchanged 64-packet hard cap; the max rendered packet is ≤30,000 (hard cap) with 0 overflow. The baseline's single packet of headroom was consumed by required corrections plus legitimate content growth; recorded, not hidden, and no limit/gate was weakened.
- [x] Complete static/focused/adversarial/repository verification passes on the corrected candidate (gofmt clean, vet, full `go test`, build, module checks, `make verify` with 0 lint/547 connectors/PM+Shepherd suites/safe reverse ETL).
- [x] One clean exact correction commit `0933611bdc3da5045ef4370fcc38519f7c7276df` is pushed; clean-tree focused/full identity reconfirmed.
- [x] Fresh trust-bound self-review at the new head uses no `c28126bb3` artifact and deliberately proves reintroduced F1 blocks: fix present blocks with 44 named `packet_coverage` findings; invariant removed silently unpacks 22 changed files (`data/cli-first-round-review-audit-r1/self-review-0933611bd/`).
- [x] PM disposition reported without silent iteration: deterministic self-review clean (compile ready, F1 caught); fresh-context model packet synthesis blocked by Codex subscription quota exhaustion (recorded operational blocker, not a correction round); Shepherd not run because model synthesis is not clean.
- [x] Stacked PR #501 targets `feat/cli-architecture-v2`; `verify` and required CI checks are green (snyk pending is a deferrable third-party integration); PR remains open/unmerged. A shallow-checkout fix deepens the test clone so CI has the base commit for the full-range compile.

## Measurement

- [x] Detector execution does not receive the separate oracle; opaque held-out mutations and clean/metamorphic controls run; corpus provenance/hash and fixture-blinding limitation are explicit.
- [x] Machine report captures recall, precision, escapes, false positives, exact invalidations, rounds, overflows, wall time, and available token/cost fields; deterministic fixtures are labeled regression evidence, not model-review or prospective production results.
- [x] Packet artifacts contain paths/metadata only; an environment-sentinel regression proves no environment-value copy.
- [x] R1/R2 dedup diagnostics are labeled retrospective development because R2 labels were inspected during tuning; `deterministic-partial-keys/v1` is frozen before the fresh prospective R3.
- [x] Correction measurement (`data/cli-first-round-review-audit-r1/f1-correction-{baseline-c28126bb33,after-uncommitted}/measurement.json`) separates before/after wall time, memory, packet count, max rendered bytes, duplication, and headroom.

## Focused commands

```bash
bash scripts/tests/pm-review-system.sh
bash scripts/tests/pm-orchestrator-contract.sh
bash scripts/tests/pi-model-routing.sh
python3 -m py_compile scripts/pm-review-system.py scripts/pm-review-lab.py
```

## Full local gates

```bash
gofmt -w cmd internal && git diff --exit-code -- cmd internal
git diff --check
go vet ./...
go test -timeout 20m ./...
go build ./cmd/pm
go mod verify
go mod tidy -diff
make verify
```

Pre-correction passes at `7f1b2d8fe`/`5601be8d0`/`0e210d18`/`528dcc68`/`e4ca19ce`/`1e640f9a4`/`c28126bb3` are historical. Current-correction full verification is recorded in `TDD-LEDGER.md`.

## Review and delivery

- [x] Captain's 2026-07-24 conditional Firstmate merge authorization is recorded; this agent remains no-merge, parent PR #438 remains draft/human-only, and the deliverable is a green open stacked PR.
- [x] Every round-1/round-2 finding has a canonical disposition; future-round findings remain pending by definition.
- [x] Stable five-round `rounds_by_range` usage is 2/5 with append-only head history and no lineage reset.
- [x] Complete correction-candidate verification and clean-tree identity reconfirmation; committed and pushed on the branch.
- [ ] Fresh trust-bound self-review synthesizes one result before Shepherd; raw responses live outside the tracked worktree and are hashed/summarized in delivery evidence.
- [ ] Independent Shepherd exact-head trajectory validation recorded only after a clean self-review.
- [x] Branch pushed normally without force; PR #501 has a Conventional Commit title, targets `feat/cli-architecture-v2`, uses `Refs #397`, and reports URL/head/risk/metrics/limitations/round usage.
- [x] Required CI checks green (snyk deferrable); PR remains open and unmerged for captain approval.
