# GitHub Certification Matrix Drift — Verification Checklist

**Status:** Blocked by a non-reproducing task premise; no generated artifact or
runtime change is justified.

## Required evidence

- [x] Mandatory RED command was run at `origin/main` `51dd6d468e4a40ece70c36efb81df4fdede8a8b6`; it unexpectedly passed before any generated write.
- [x] Canonical GitHub-only generator ran twice; no hand-authored JSON and no artifact diff.
- [x] Generator audit confirms declaration-bundle and scoped runtime-endpoint-ledger inputs, deterministic ordering, and an unchanged GitHub certification shard.
- [x] Canonical matrix check is GREEN.
- [x] Second generation is byte-stable.
- [x] Existing focused generator happy, stale-artifact, and scope/determinism tests pass.
- [ ] Further connector/repository gates are not run: no code or generated-artifact correction exists to validate, and a no-op PR would violate the task's minimal-scope requirement.
- [ ] Inline verify-work and code review are not applicable until an exact failing base is supplied.
- [x] `git diff --check` passes; no GitHub certification artifact changed.

## Explicitly not applicable

No runtime command, help, manual, website documentation, external provider interaction, credential, or reverse-ETL operation changes.
