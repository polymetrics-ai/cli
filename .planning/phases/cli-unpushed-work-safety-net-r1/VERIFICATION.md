# Verification — unpushed-work safety net r1

## Local evidence

- [x] Red: `python3 scripts/tests/unpushed-work-safety-net_test.py` failed before production code
  existed (seven failures; all expected to be missing-observer failures).
- [x] Green: `python3 -m unittest -q` split into short groups, including the multi-operation
  rebase/merge/cherry-pick/revert/bisect/squash test, a real remote change between snapshot and
  final recheck, the no-force wrapper test, and the status/launchd test. All 14 acceptance tests
  passed. The split is deliberate: this worker's
  per-command runner cuts off a full ~40-second suite; the Make target runs the complete suite in
  CI.
- [x] `python3 -m py_compile scripts/unpushed-work-safety-net.py
  scripts/tests/unpushed-work-safety-net_test.py`
- [x] `ruff check scripts/unpushed-work-safety-net.py
  scripts/tests/unpushed-work-safety-net_test.py`
- [x] `ruff format --check scripts/unpushed-work-safety-net.py
  scripts/tests/unpushed-work-safety-net_test.py`
- [x] `python3 scripts/unpushed-work-safety-net.py --help`
- [x] `git diff --check`
- [x] `scripts/gsd prompt verify-work cli-unpushed-work-safety-net-r1 --auto` generated and
  executed inline: all deliverables have automated real-Git evidence; no human-only product
  judgment is required.
- [x] `scripts/gsd prompt code-review cli-unpushed-work-safety-net-r1` generated and executed
  inline; all findings are recorded in `REVIEW.md` and fixed.
- [ ] `make unpushed-work-safety-net-check` is intentionally left to the no-mistakes/CI stage
  because it invokes the same complete ~40-second test process and this worker's command ceiling
  would truncate it. `make -n unpushed-work-safety-net-check` confirmed the exact complete-suite
  command; it is wired into `make verify` for remote enforcement.

## Scope and parity

- No `cmd/` or `internal/` change is planned, so CLI help/manual/website parity is not applicable.
- The existing `.githooks/pre-commit` and its unwired `core.hooksPath` documentation remain
  untouched; docs will state that this observer is separate and opt-in.
