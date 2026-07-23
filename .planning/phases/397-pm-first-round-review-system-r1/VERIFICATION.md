# Issue #397 PM First-Round Review System Verification

**Status:** planned; `verificationPassed: false`

## Identity and scope

- [x] Isolated disposable worktree is not the primary clone (`SETUP-EVIDENCE.md`).
- [x] Status/log/untracked/stash/diff inventory captured clean before production edits (`SETUP-EVIDENCE.md`).
- [x] Normal fetch verified remote parent contains and equals PR #495 squash `0f8c964ba9cfbe1b1eec8e7998eacf4158ef0e20` (`SETUP-EVIDENCE.md`).
- [x] Clean detached remote-base transition and branch creation recorded without reset/rebase/amend/force (`SETUP-EVIDENCE.md`).
- [x] Branch is `chore/pm-first-round-review-system-r1`.
- [x] Parent PR #438 exists and remains draft/human-only.
- [ ] Diff remains disjoint from all PR #493-owned paths.
- [ ] No #408/TUI, Gong/#497, write-URL, dependency, credential, connector, CLI help, or reverse-ETL behavior change.
- [ ] Historical PR #495 evidence remains truthful and is not rewritten as clean.
- [ ] Diff is within the exact positive path allowlist in `PLAN.md`.
- [ ] No new CLI/subcommand, orchestration owner, article/#498, `go.mod`, or `go.sum` change.

## RED/GREEN requirements

- [x] RED: canonical `final_parent_readiness` incorrectly classifies `human_ready`.
- [x] RED: valid identifier `SEC1` with noncanonical disposition is not rejected by current parser.
- [x] RED: prose-only dependency mutation escapes baseline.
- [x] RED: transitive prohibited-template mutation escapes direct-file baseline.
- [x] RED: replacement-head lineage reset/fragmentation escapes shape-only baseline.
- [x] RED: unsafe absolute/parent/control/option-like references are rejected at the semantic corpus layer; real symlink/identity integration RED remains pending.
- [x] RED: cross-format graph, cycle, and missing-target semantics are exercised in the frozen corpus; real file-parser integration RED remains pending.
- [x] RED: stale packet identity, incomplete coverage, overflow, threshold boundaries, replacement/resume, and cap boundary fail semantically; one-way legacy migration/append-only integration RED remains pending.
- [x] RED: opaque corpus and separate oracle are frozen and hashed before treatment implementation.
- [x] GREEN: all five concrete cases and pre-frozen mutation cases are detected for the intended semantic reason.
- [x] GREEN: unknown schema/kind, stale evidence, cap exceeded, arbitrary IDs, and missing active targets block.
- [x] GREEN: clean/metamorphic controls do not produce findings.

## Review compiler

- [x] Active required-reference closure records source, target, and edge reason.
- [x] Missing active targets and prohibited reachable targets fail.
- [x] Authority registry records authoritative state plus writers/readers/mirrors.
- [x] Dispatch/readiness checks parse relationships rather than trust prose.
- [x] Exact base/head are verified and embedded in each packet.
- [x] Small coherent changes stay one packet only at ≤20 files, ≤600 lines, one domain.
- [x] 21–25 files, 601–800 lines, or exactly two domains split conservatively; >25 files, >800 lines, or >2 domains split mandatorily.
- [x] Any partition that cannot meet ≤20 changed, ≤10 closure/authority, and declared 30K-token packet caps blocks rather than truncates.
- [x] Every changed file is assigned; each response declares reviewed, closure, invariants, unreviewed, findings, and overflow/truncation.
- [x] Missing response/coverage, stale identity, overflow, or silent truncation cannot synthesize clean.
- [x] Findings are unlimited and synthesize to one PM-owned local-Codex disposition.
- [x] Shepherd remains independent and runs only after clean synthesis.

## Measurement

- [x] Historical source identities are retained for PR #495 replays.
- [x] Detector execution does not receive the separate oracle.
- [x] Opaque held-out mutations and clean/metamorphic controls run.
- [x] Machine report captures recall, precision, escapes, false positives, exact invalidations, rounds, overflows, wall time, and available token/cost fields.
- [x] Deterministic fixture results are not described as model-review or prospective production results.
- [x] Corpus provenance/hash and fixture-level blinding limitation are explicit.
- [x] Unavailable token/cost/prospective evidence is explicit.
- [x] Packet artifacts contain paths/metadata only; environment-sentinel regression proves no environment-value copy.

## Focused commands

```bash
bash scripts/tests/pm-review-system.sh
bash scripts/tests/pm-orchestrator-contract.sh
bash scripts/tests/pi-model-routing.sh
bash -n scripts/pm-terminal-classifier.sh scripts/tests/pm-review-system.sh scripts/tests/pm-orchestrator-contract.sh
# focused test additionally exercises classifier usage/malformed JSON/legacy stdout+stderr+exit compatibility,
# JSON envelope fields, non-TTY execution, unsafe paths and symlinks, closure formats/cycles,
# threshold boundaries, state transitions, and stale evidence
ruby -e 'require "psych"; Psych.parse_file(ARGV.fetch(0))' .agents/agentic-delivery/schemas/orchestration-state.schema.yaml
ruby -e 'require "psych"; Psych.parse_file(ARGV.fetch(0))' .planning/traces/cli-architecture-v2-orchestration-state.yaml
python3 -m py_compile scripts/pm-review-system.py
python3 -m json.tool .agents/agentic-delivery/contracts/pm-review-system.json >/dev/null
python3 -m json.tool .planning/phases/397-pm-first-round-review-system-r1/MEASUREMENT.json >/dev/null
```

## Full local gates

```bash
gofmt -w cmd internal
git diff --exit-code -- cmd internal
git diff --check
go vet ./...
go test -timeout 20m ./...
go build ./cmd/pm
go mod verify
go mod tidy -diff
make verify
```

- [ ] Focused gates pass.
- [ ] Shellcheck runs when available; unavailability is recorded without installing dependencies.
- [ ] JSON/YAML parsing passes.
- [ ] Go formatting produces no product diff.
- [ ] `go vet ./...` passes.
- [ ] `go test -timeout 20m ./...` passes.
- [ ] `go build ./cmd/pm` passes.
- [ ] `go mod verify` passes.
- [ ] `go mod tidy -diff` passes.
- [ ] `make verify` passes.

## Review and delivery

- [ ] Exact implementation commit exists and coherent plan/RED/GREEN/measurement checkpoints were pushed additively.
- [ ] All tracked evidence was committed before final exact-head packet/Shepherd gates; no tracked write followed them.
- [ ] Fresh local-Codex packet system reviews exact base/head and synthesizes one result before Shepherd; raw responses live outside the tracked worktree and are hashed/summarized in delivery evidence.
- [ ] Every finding has a canonical disposition.
- [ ] Fresh five-round `rounds_by_range` usage and append-only head history recorded without lineage reset.
- [ ] Independent Shepherd exact-head trajectory validation recorded after clean Codex review.
- [ ] `no-mistakes axi` returns `checks-passed`; `passed` (merged/closed) is treated as a violation/escalation, not success for this task.
- [ ] Any AXI-created commit/base/head change invalidated prior evidence and triggered applicable full verification, fresh packet synthesis, and fresh Shepherd at final identities.
- [ ] No parallel/manual reviewer ran outside the specified PM packet system.
- [ ] Branch pushed normally without force.
- [ ] PR has Conventional Commit title, targets `feat/cli-architecture-v2`, uses `Refs #397`, and reports full URL, exact source/head, risk, metrics, limitations, and round usage.
- [ ] Published branch history remained additive; any proposed post-publication non-additive pipeline rewrite stopped for human direction.
- [ ] CI green.
- [ ] PR remains open and unmerged for captain approval.
