# Threat model

Untrusted inputs include issue text, sub-issue text, planner output, repository files, test output, and
GitHub timestamps. Mitigations are bounded exact schemas, control-character rejection, canonical paths,
non-symlink/no-clobber publication, host-owned branch/actor/expiry/issue fields, marker idempotency,
immutable command-ID selection, closed Node/Go/Make quality-gate recipes, shell-free POSIX process-group
spawn, descendant termination, hard post-kill settlement, output/time bounds, cancellation/join,
inline-safe GitHub fields, shared dependency/scope validation before journaling, and exact-head GitHub
checks. Residual accepted risk: trusted local tests may execute arbitrary repository
test code as the current user. Secrets, generic Git commands, and default-branch merge authority are never
delegated.
