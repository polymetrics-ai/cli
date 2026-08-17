# Run state — GitHub certification suite r1

- Status: planning complete; implementation has not started.
- Base verified: `a64c4be58156d30bd35632e5c32cfeef33a7cd1f` (`origin/integration/4015-mvp-flat-r1`).
- Dependency: PR #4198 is open (15 checks passed, 1 failed, 4 skipped) and is the hard dependency for accepted `http_exchanges` evidence.
- Scope: one GitHub connector; shared generator code must be connector-neutral and definition-driven.
- Safety: no credential value, ambient GitHub login, arbitrary external target, or provider mutation has been used.
