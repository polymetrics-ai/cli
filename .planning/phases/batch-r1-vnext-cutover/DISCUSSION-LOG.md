# Batch R1 vNext cutover discussion log

## 2026-09-01 restart decision record

The captain has already made every product and architecture decision needed for this continuation:

- Finish legacy cleanup before migrating any additional connector.
- Use only `source.lock.json -> canonical descriptor -> deterministic execution JSON -> existing PM-CLI runtime`.
- Delete alternate execution/admission paths; do not preserve them as fixtures, compatibility adapters, feature flags, importers, certification, or retention gates.
- Do not introduce a shared runtime foundation, recover excluded local work, create credentials, use provider I/O, merge, force push, rebase, or open another PR.
- Migrate the named remaining connectors in fixed order and push every independently green cohort normally to the established Batch R1 remote branch.

No interactive product question remains. The only operational constraint discovered at restart is the missing GSD ROADMAP/prompt prerequisite described in `CONTEXT.md`; it is handled through the documented inline/manual-GSD evidence fallback, not by changing scope or weakening TDD/review requirements.
