# Cache policy199 verification

The explicit cache-only setup and affected physical witness passed. Independent review remains pending; no checkpoint acceptance or hosted-green claim.

| Original run | Exit | Pass/run events | Raw SHA256 |
|---|---:|---:|---|
| dbtest-cache-red-199-01 | 1 | 0/0 | af6e20da1e0aa8bca6bcffa2d2300be365d23c7aa4b3a08a78760d0583479bf0 |
| dbtest-cache-red-199-02 | 1 | 1/11 | 86931ca5c8891870574e570eee8c3080434ed413ce8a571ea69d4c88a935a9ba |
| dbtest-cache-red-199-03 | 1 | 1/11 | 3a8e5bbcd6cf01c788d9cca0fd16ac02a415ed479e6e9770b63a7bdac561aeec |
| dbtest-cache-green-199-01 | 0 | 93/93 | 906e5aa4fefa96723520f2af8d8b1a216f03de900ee4f3ff691b7cdfd2ca7652 |
| dbtest-cache-override-red-199-01 | 1 | 0/1 | 7e95e9c855520705274582ce08a2ffbcc8e01798be19d6b8b408d2aebaa9bc9c |
| dbtest-cache-green-199-02 | 0 | 94/94 | 0b05e889d6a61b456635479d50e5ec5df37ec40d47efc9ebba127f951b73ff99 |
| dbtest-cache-race-199-01 | 0 | 94/94 | f8c2297eebeb0feac19be1f9089718a31c5d502650b4200151e5e5373288be00 |
| shared-checkpoint-physical-199-01 | 0 | 1/1 | 506e68c3302bfca29a7b428389d9a96ec35e02d6cb6fb3e5e627a72bee9593bb |
| dbtest-cache-vet-199-01 | 0 | 0/0 | e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855 |
| dbtest-cache-lint-199-01 | 0 | 0/0 | e92606b0bf483111dff0a120c315ea165821348f31365020e2468a0059095c47 |

The initial RED01 was a new-test compile error (missing existing Close context argument), not behavioral RED. RED02/03 reached actual Start and unwanted pulls/admission/identity-evidence assertions. Fields added only to make the test configuration compile did not implement policy. The override RED reached actual New accepting conflicting --pull arguments. Default pull was retained; all existing harness tests also passed. Parent/child event counts overlap and are not added. Original capture runner and every failed attempt remain unchanged.

The physical invocation ran fresh count1 against the explicitly selected local Docker socket with cache-only. Actual log names immutable image sha256:21f6013073bc6b92830a2129570e2f5ec42a6c734b5a985a41e83aa58f54c3c1. It passed the existing1001-row WAL/Parquet/manifest, approval/readback/ack, CDC monotonic oracle, interrupted lease and resumed-row assertions in187.860496625s. Actual restart counts before/interruption/after are1/1/2 with one resumed key. All assertions unchanged by199 setup plumbing. Original194 pre-container pull/helper failure remains separate. Post-terminal explicit owner-label listing returned empty; shared-db-cleanup-199.json preserves it.

Every listed terminal receipt has changed_inputs_after empty. The compact private shared-cache-handoff-199-receipts.json binds command/source/tree/receipt/raw identities. Pre-edit dbtest and PostgreSQL source snapshots remain in cache-199/pre-edit-inputs.json for196 read-only architect drift reconciliation. No production connector/authoring changes, no new dependency or provider access. Atlas has no dbtest entry and no production foundation contract changed. Test setup README updated; pm help/manual/site are not affected.

Verification executed inline through resolved GSD sources/prompts per single shared-owner contract. Scoped vet and new-only tagged lint pass, gofmt and diff checks pass. The physical test builds/runs real pm binaries; no unaffected full suite rerun. Published R1 Verify and CodeQL are still failing separately, with full log and annotations retained in shared-verify-observation-199.json and cp16-codeql-observation-194.json. They are not fixed by this helper change.

Introduced states/edges for review: Config.ImagePolicy empty normalization and closed validation; refusal of conflicting container args; Start cached-absence refusal vs existing pull path; captured immutable ID/report evidence; explicit no-pull start; tagged PostgreSQL opt-in/logging. Existing target/capacity/ownership/readiness/cleanup remains shared. Fake runtime tests assert original-ID launch after mutable-tag substitution, mismatched generated-ID refusal and caller-readiness cleanup.
