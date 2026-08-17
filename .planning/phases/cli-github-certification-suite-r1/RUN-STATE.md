# Run state — GitHub certification suite r1

- Status: implementation in progress; source-derived generator and generated GitHub sweep are complete, verification remains.
- Base verified: `a64c4be58156d30bd35632e5c32cfeef33a7cd1f` (`origin/integration/4015-mvp-flat-r1`).
- Dependency: PR #4198 is open (15 checks passed, 1 failed, 4 skipped) and is the hard dependency for accepted `http_exchanges` evidence.
- Scope: one GitHub connector; shared generator code must be connector-neutral and definition-driven.
- Live assertion proof: after schema validation, a scratch-impossible declaration made `direct_read_sweep_repo_read_file` red on its own produced-value assertion; exact restoration made all 23 declaration-owned direct-read stages pass. It used the named disposable fine-grained token through `--from-env`, serial read-only execution, and `Polymetrics-Cert/pm-cert-3993-20260810-wz0fru` only. No ambient GitHub login or provider mutation was used.
- Evidence dependency: PR #4198 remains open, so no accepted `http_exchanges` evidence record is emitted and no generated command is promoted to `pass`.
