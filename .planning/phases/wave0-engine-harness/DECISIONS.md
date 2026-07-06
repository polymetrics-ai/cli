# Coordinator decisions — wave0-engine-harness

Answers to planner open questions (2026-07-02):

1. **golangci-lint acquisition**: local binary via Homebrew (installed: 2.12.2). No go.mod change,
   no `go run` build latency. `.golangci.yml` (B-18) targets this version; Makefile target guards
   with a "golangci-lint not found — brew install golangci-lint" hint.
2. **B-12 (certify report/cliharness)**: APPROVED to float to Wave B — zero engine deps, shortens
   the critical path.
3. **inventory.json loc convention**: all `.go` including tests (matches orchestration-plan
   calibration).
4. **Golden api_surface depth**: minimal honest surfaces in wave0 — every implemented stream/write
   listed, remaining documented endpoints as `excluded: {category: out_of_scope, reason: "Pass B
   capability expansion"}`. Full surfaces are wave5 (Pass B). Reviewer checklist aligned.

Standing wave0 rules (from SPEC):
- Goldens' legacy packages keep compiling and stay registered; engine-backed versions are built in
  tests from defs.FS only. Registration flip is wave6.
- Sole legacy edit allowed: cmd/registrygen skip map + byte-identical registryset regen guard.
- No new Go module dependencies (validator is internal minimal draft-07 subset).
