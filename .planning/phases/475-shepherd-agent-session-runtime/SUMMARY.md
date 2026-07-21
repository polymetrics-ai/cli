# Summary — Issue #475

Status: exact-head correction in progress; prior verification is superseded.

The scoped in-process runtime, least-authority tool policy, trusted role prompt envelopes, and
bounded redacted handoffs are implemented behind injected ports. Implementation/correction route
only to `openai-codex/gpt-5.6-sol`/`high`; every planning/research/review/validation/verification/
orchestration role routes to the same model at `xhigh`. Caller, session, and terminal route drift
fail closed.

Opaque workspace read/edit/write tools enforce relative allowlisted paths, sensitive-path denial,
bounded/redacted output, and read-only mutation denial. Typed host capabilities are closed,
bounded, allowlisted, and reject generic shell, HTTP/SQL write, secret export, and recursive agent
authority. Runtime setup/session teardown is deadline-bound; abort, timeout, close, and parent
shutdown coalesce; child setup/session settlement joins once; failed cleanup quarantines further
dispatch. Handoffs accept one closed bounded JSON schema bound to run/generation/lane/head/nonce and
redact secret-like material before return.

The focused suite passes 22/22 and the complete Shepherd suite passes 159/159. Strict owned plus
all-Shepherd-production TypeScript passes against the explicit Pi 0.80.6 installation, as do the
pinned offline RPC smoke and diff/scope checks. The healthy repo-local GSD adapter does not expose
`programming-loop`, so the phase completed under the recorded `manual_gsd_fallback`.

No dependency, external state, credential, GitHub state, or parent artifact has been changed.

PR #486 review at `4e41c2ec1175a109c10f125203dc54d381b982bd` identified two P1 corrections:
late session creation can escape lifecycle ownership after the cleanup bound, and quoted
JSON/YAML/Bearer secret values can escape redaction. A new strict TDD cycle is active; terminal
results below must not be treated as current until the correction gates are rerun.
