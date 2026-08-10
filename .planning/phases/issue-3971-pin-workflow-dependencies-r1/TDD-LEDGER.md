# TDD ledger: pin build dependencies

## Manual GSD fallback

`scripts/gsd doctor`, all required `scripts/gsd sources` commands, and `go run ./cmd/agentcontractgen check` passed. The official phase commands were generated, but `gsd-sdk query init.phase-op 3971` returned `phase_found: false`; this GitHub issue is not a roadmap phase. The documented inline/manual GSD fallback is active for this isolated, YAML-only security repair. Red and Green evidence is recorded below before the production workflow edits.

## Immutable-resolution capture

The following values were resolved from the original refs on 2026-08-10 with `git ls-remote`. Annotated tags use their peeled commit (`^{}`). `actions/dependency-review-action@v5` is an upstream `v5` branch; its current branch commit is recorded exactly rather than replacing it with a different tag.

| Original action ref | Current immutable commit |
| --- | --- |
| `actions/checkout@v7` | `3d3c42e5aac5ba805825da76410c181273ba90b1` |
| `actions/setup-go@v6` | `924ae3a1cded613372ab5595356fb5720e22ba16` |
| `ossf/scorecard-action@v2.4.3` | `4eaacf0543bb3f2c246792bd56e8cdeffafb205a` |
| `github/codeql-action@v4` | `5595ccaf912efad79be6eef63a5619ff05969be3` |
| `actions/dependency-review-action@v5` | `a1d282b36b6f3519aa1f3fc636f609c47dddb294` |
| `pnpm/action-setup@v6` | `0977fd99725f1db4007ccb2928dbb4e90d06cc86` |
| `actions/setup-node@v6` | `249970729cb0ef3589644e2896645e5dc5ba9c38` |
| `googleapis/release-please-action@v5` | `45996ed1f6d02564a971a2fa1b5860e934307cf7` |
| `actions/upload-artifact@v4` | `ea165f8d65b6e75b540449e92b4886f43607fa02` |
| `actions/download-artifact@v4` | `d3f86a106a0bac45b974a628896c90dbdf5c8093` |
| `docker/setup-qemu-action@v3` | `c7c53464625b32c7a7e944ae62b3e17d2b600130` |
| `sigstore/cosign-installer@v3` | `398d4b0eeef1380460a10c8013a76f728fb906ac` |
| `actions/attest@v4` | `1e69f48acb82d1966a394da916b4c1698aa569d6` |
| `anthropics/claude-code-action@v1` | `6b082c41935b4c8a3b8b0ef85ba4ba4d9eeb8975` |

The image manifest-list digests were resolved on the same date with `docker buildx imagetools inspect`:

| Original image tag | Current immutable manifest-list digest |
| --- | --- |
| `node:26-alpine` | `sha256:aadf416b2cdce311a8811ba3f0608a61b77dbf997500e2eafe781b51f6a0b019` |
| `postgres:17-alpine` | `sha256:742f40ea20b9ff2ff31db5458d127452988a2164df9e17441e191f3b72252193` |

## Red

- [ ] Red: `./scripts/tests/pinned-build-dependencies.sh` must fail against the baseline because all external `uses:` lines carry mutable version refs and the Node/PostgreSQL images carry mutable tags.
- [ ] Red: the failure must name at least one actionable mutable reference rather than merely accepting YAML syntax.

## Green

- [ ] Green: after pinning, `./scripts/tests/pinned-build-dependencies.sh` passes and confirms a full SHA plus version comment on every external action and a digest on every literal build image.
- [ ] Green: `make release-workflow-check` passes with the new gate wired into the existing release workflow verification target.

## Refactor / review checks

- [ ] Preserve the existing action/image tags in comments or tag-before-digest form; no action/image version upgrade is permitted in this slice.
- [ ] Inspect the final diff for only expected workflow, image, gate, and evidence changes.
- [ ] Run the explicit-file manual code review and record its result in `REVIEW.md`.
