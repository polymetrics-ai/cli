# Issue #397 PM Orchestrator Extension Summary

Status: REVIEW CORRECTION ROUND 1 — re-verification/re-review, Shepherd, and PR checks pending.

Captain authorized additive canonical orchestration corrections on draft PR #495. The extension starts from reviewed synchronization head `3fd63fbe0f526873fa3adb8a75fa5f20342d52a6`, keeps PR #493 head `e21e56339390c5e1946eb4cfaf276eb80a889f29` path-disjoint, and adds no product behavior or dependency.

The canonical route now has one owner: `/pm-orchestrate` runs the universal lifecycle when registry discovery shows `programming-loop` is absent. It preserves durable reconciliation, isolated workers, bounded correction, machine/credential contracts, and human merge authority. Exact-head verification is followed by a fresh-context read-only local Codex review with written dispositions, then independent Shepherd validation before integration. Every changed head invalidates prior exact-head evidence.

Current/future PM instructions no longer require or route fallback coverage through Claude or GitHub Copilot. Legacy bot workflow documents and the old disposition role remain only with explicit migration/deprecation pointers; historical phase records were not globally rewritten.

Focused RED failed against the old route. Focused GREEN passes for the PM contract, Pi model routing, Shepherd guards, JSON/YAML parsing, dependency invariance, and PR #493 path disjointness. Gofmt/diff, vet, full tests, build, module checks, and `make verify` passed at implementation head `d72a93018933541d390884f96b285856e269a1ab` and evidence head `3c88fc78062ba0a3437f79bc88c395286c228c65`.

Fresh-context local Codex review at `3c88fc78062ba0a3437f79bc88c395286c228c65` produced five findings. Correction round 1 tightens unavailable-command examples and the PR #493 migration gate, makes the review schema conditional/backward-readable with focused fixtures, and persists a default four-round correction budget. The cited Gong direct-read defect is byte-identical to current main and deferred under the explicit no-product-change boundary. Re-verification and exact-head re-review follow.
