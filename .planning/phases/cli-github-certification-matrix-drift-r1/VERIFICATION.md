# GitHub Certification Matrix Drift — Verification Checklist

**Status:** Planned

## Required evidence

- [ ] Mandatory RED drift command and inspected root cause recorded.
- [ ] Canonical GitHub-only regeneration recorded; no hand-authored JSON.
- [ ] Semantic audit confirms provenance, stable ordering, operation reachability, and unchanged certification truth.
- [ ] Canonical matrix check is GREEN.
- [ ] Second generation is byte-stable.
- [ ] Focused generator happy, stale-artifact, and deterministic tests pass (or a focused regression supplies missing coverage).
- [ ] Focused GitHub connector checks pass.
- [ ] Connector validate, surface sync, boundary, generated docs/artifact checks, and required scoped repository verification pass.
- [ ] Inline manual-GSD verify-work and code review record no unresolved finding.
- [ ] `git diff --check` passes and final changed paths stay within the scope fence.

## Explicitly not applicable

No runtime command, help, manual, website documentation, external provider interaction, credential, or reverse-ETL operation changes.
