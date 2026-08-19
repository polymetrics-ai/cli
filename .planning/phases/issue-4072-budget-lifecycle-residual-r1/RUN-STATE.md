# #4072 budget lifecycle residual run state

- Lifecycle: inline/manual GSD fallback (named issue phase is not in the
  numeric roadmap; canonical lane also forbids role spawning).
- Discussion: complete — dispatch acceptance locks the only behavioral choice.
- Plan: complete before production edits.
- TDD: red reproduced and green counter proof passed.
- Execution: complete.
- Verification: complete; the full GitHub race regression has a disclosed
  pre-existing timing failure, while the focused lifecycle race proof passes.
- Code review: inline review complete; see `REVIEW.md`.
