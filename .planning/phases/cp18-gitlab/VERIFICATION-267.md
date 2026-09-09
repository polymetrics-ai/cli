# CP18 GitLab verification checklist — 267

- [ ] Baseline merged/reconciled against current R1 without discarding 84f6d456 receipts or preserved untracked material.
- [ ] Every G1–G8 ledger row has a fresh source-bound Red/Green (or an explicit source-backed limitation), exact selected membership, and input/toolchain/candidate binding.
- [ ] Renderer output came solely from the GitLab source lock; no generated execution/index/proof artifact was hand-authored.
- [ ] `lock-render gitlab --check`, definitions validation, source-to-canonical-to-execution joins, and commandrunner preflight pass.
- [ ] Direct read/write/binary tests assert exact fake-provider wire shape, byte/count bounds, approval and credential boundaries.
- [ ] ETL/reverse/sync tests prove warehouse mediation, mode constraints, checkpoints, acknowledgements, and durable failure state.
- [ ] All seven lanes and applicable source/destination/mode roles are honestly dispositioned; 1752 primary plus two labelled supplements remain exact.
- [ ] CLI help/manual/website parity is updated or explicitly not applicable for each changed runtime surface.
- [ ] Frozen candidate has Firstmate-routed independent Terra/xhigh review, complete finding disposition, and no-mistakes gate as instructed.
- [ ] Any R1 increment was validated from the clean integration worktree after the normal merge/cherry-pick, pushed without force, and remote SHA was read back.

## Verified slices

- [x] G1 source scalar declarations: renderer check, targeted one-connector validation, 203 source-to-canonical-to-CLI joins, generated command preflight, and three exact fake-provider wires. See `TDD-LEDGER-267.md` and `RUN-267.md`.
