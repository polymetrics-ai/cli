# Discussion log — Issue #3974 typed database foundation recovery

**Mode:** `discuss-phase --auto` with repository-approved inline fallback\
**Date:** 2026-08-11

The supplied recovery brief fixed the material choices. This log records them rather than
reopening them through interactive prompts.

| Area | Selected decision | Reason/evidence |
| --- | --- | --- |
| Recovery unit | Recover the complete twelve-commit aggregate | The topology report says the type/admission hardening is one cohesive F1 contract. |
| Base and PR topology | Start from the current remote PostgreSQL parent and target it, never `main` | Parent PR #4017 and the brief establish the stacked topology. |
| Documentation conflicts | Keep #4003 canon; port only database-specific clauses | The topology report identifies stale #4014 doc overlap. |
| PostgreSQL seam | Retain current TLS and fail-closed CDC state; no capability promotion | The F1 boundary is non-executing. |
| Shared dependency | Do not absorb #3864 | It is only a later transport dependency. |
| Certification gate | Use bounded #3995-compatible evidence if its branch is not in ancestry | The brief forbids claiming automatic approval. |

No deferred ideas were introduced.
