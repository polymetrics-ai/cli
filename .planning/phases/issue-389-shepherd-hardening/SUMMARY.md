# Issue #389 Slice C Summary

Status: GREEN candidate on `fix/389-shepherd-proof-recovery` from base `1a050692f9e47b5b4d3d74cfb38e56c67d461399`.

Implemented only crash-safe promotion:
- protected SQLite staging intent and eight-state promotion journal;
- immutable proof and full attestation identity binding;
- bounded deterministic `.gsd` manifests and root-confined copies;
- WAL-safe SQLite online backup plus integrity verification;
- exact attempt/Git/resource revalidation before promotion;
- same-filesystem backup/install renames with both parent directories fsynced;
- forward recovery after candidate Git, fail-closed moved/dirty/ambiguous state;
- rooted cleanup tombstones and universal blocked-journal access gates.

TDD: RED compile failures were recorded before production edits. Focused promotion suites, nine
journal/Git/swap failpoints, full nested tests, race coverage, vet, build, gofmt, and diff checks pass.
Two read-only reviewer/security cycles were run and actionable Slice C findings were fixed. The
complete staged SQLite snapshot is protected by the journal after worker quiescence; changing Slice
A's independent-validator evidence schema was rejected as outside this slice.

Slice D onward, PR creation, canaries, GitHub mutation, and merge remain blocked/human-gated.
