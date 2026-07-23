# Issue #397 PM First-Round Review System Summary

Status: deterministic treatment and packet route focused GREEN; exact-head full verification/review pending.

- Base verified at parent squash `0f8c964ba9cfbe1b1eec8e7998eacf4158ef0e20`.
- Separate branch: `chore/pm-first-round-review-system-r1`.
- Scope: deterministic semantic preflight, bounded exact-version packets, one PM synthesis,
  independent downstream Shepherd, and fixture/replay measurement.
- Required baseline: two accepted PR #495 findings plus three original preventable misses.
- GSD adapter: healthy, but `programming-loop` absent; active `/pm-orchestrate` owns the GSD/TDD lifecycle.
- PR #493-owned paths, #408/TUI, Gong product behavior, dependencies, credentials, and reverse ETL are excluded.
- Setup/fetch/branch evidence is captured in `SETUP-EVIDENCE.md`.
- Semantic RED reproduced both accepted findings, all three original misses, ten opaque mutations, and six threshold decisions. Ten paired clean controls stayed unflagged.
- Opaque inputs and separate oracle were frozen and hashed before GREEN; the detector receives no oracle argument.
- Dependency-free treatment result on the frozen fixture corpus: 15/15 defect detections, 0/10 clean-control false positives, seven of seven threshold decisions. Baseline escaped 15/15. This is deterministic preflight evidence only; model tokens/cost, correction rounds, and prospective production evidence are unavailable.
- Active PM route now compiles closure/authority/semantic gates, creates bounded packets, requires complete responses, synthesizes one PM verdict, and keeps Shepherd independent/downstream.
- Full verification, exact-head packet review, Shepherd, no-mistakes delivery, PR, and CI remain pending.
