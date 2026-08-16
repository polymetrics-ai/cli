---
phase: issue-4072-budget-lifecycle-residual-r1
status: in_progress
---

# #4072 residual verification checklist

| Must-have | Result | Evidence |
| --- | --- | --- |
| Grant calls `Decide` and `Finish` once each | pending | Counter-based focused test. |
| Refusal has one `Decide`, zero `Finish`, and zero sends | pending | Counter-based focused test plus typed error assertion. |
| Required shared failure remains typed and pre-I/O | pending | Existing missing/unreachable regressions. |
| Failed mint still makes exactly one POST | pending | Existing 500 regression. |
| No lifecycle secret retention | pending | Fake only persists call counts and safe batch/observation fields. |
| Required package, consumer, generator, lint, docs, and boundary gates pass | pending | Exact commands/results recorded after implementation. |
