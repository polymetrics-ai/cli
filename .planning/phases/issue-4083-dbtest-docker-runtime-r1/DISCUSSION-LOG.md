# Discussion log — Issue #4083

## Inputs considered

- The #3976 PostgreSQL catalog lane is blocked because the common dbtest
  harness is hard-wired to Podman.
- Docker and Podman expose compatible local Unix socket APIs, but the selected
  runtime must remain explicit and every CLI invocation must address the
  supplied socket directly.
- The host has Docker and Podman client binaries but, at planning time, no
  reachable direct Docker daemon and no discovered direct Podman socket. This
  is an evidence limitation, not permission to make enabled integration tests
  skip.

## Conclusions

1. Use a small runtime value in `dbtest.Config`, not endpoint guessing or a
   second harness.
2. Replace the Podman-specific MySQL environment input with an explicit
   runtime plus endpoint pair and make enabled incomplete configuration fatal.
3. Retain all endpoint refusals and identity/capacity checks; do not support
   external database endpoints in this issue.
4. Demonstrate Docker and Podman only where a direct local endpoint is
   observed. Record unavailable runtime evidence honestly.
