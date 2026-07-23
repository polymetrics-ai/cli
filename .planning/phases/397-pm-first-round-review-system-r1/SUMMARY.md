# Issue #397 PM First-Round Review System Summary

Status: complete RED baseline and corpus freeze captured; treatment GREEN next.

- Base verified at parent squash `0f8c964ba9cfbe1b1eec8e7998eacf4158ef0e20`.
- Separate branch: `chore/pm-first-round-review-system-r1`.
- Scope: deterministic semantic preflight, bounded exact-version packets, one PM synthesis,
  independent downstream Shepherd, and fixture/replay measurement.
- Required baseline: two accepted PR #495 findings plus three original preventable misses.
- GSD adapter: healthy, but `programming-loop` absent; active `/pm-orchestrate` owns the GSD/TDD lifecycle.
- PR #493-owned paths, #408/TUI, Gong product behavior, dependencies, credentials, and reverse ETL are excluded.
- Setup/fetch/branch evidence is captured in `SETUP-EVIDENCE.md`.
- Semantic RED reproduced both accepted findings, all three original misses, ten opaque mutations, and six threshold decisions. Ten paired clean controls stayed unflagged.
- Opaque inputs and separate oracle are frozen and hashed before GREEN; the detector receives no oracle argument.
- Verification, review, no-mistakes delivery, PR, and CI remain pending.
