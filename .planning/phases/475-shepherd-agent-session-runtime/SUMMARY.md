# Summary — Issue #475

Status: exact-head correction Cycle 4 GREEN; complete revalidation pending and Cycle 3 evidence is
superseded.

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

The focused suite passes 24/24 and the complete Shepherd suite passes 161/161. Strict owned plus
all-Shepherd-production TypeScript passes against the explicit Pi 0.80.6 installation, as do the
pinned offline RPC smoke and diff/scope checks. The healthy repo-local GSD adapter does not expose
`programming-loop`, so the phase completed under the recorded `manual_gsd_fallback`.

No dependency, external state, credential, GitHub state, or parent artifact has been changed.

PR #486 review at `4e41c2ec1175a109c10f125203dc54d381b982bd` identified two P1 corrections:
late session creation can escape lifecycle ownership after the cleanup bound, and quoted
JSON/YAML/Bearer secret values can escape redaction. Both are corrected: a claimed/abandoned
creation owner retains and cleans every eventual session without extending bounded run completion,
and line-bounded quoted-value patterns redact prompt, tool, handoff summary, and finding fields
without changing ordinary prose. The strict correction cycle and every declared lane gate pass.

Re-review at `526dfec4282b442c4b32138ab036d4cc7e97b475` found that multiline YAML/quoted credential forms
still escape the line-limited patterns and that ambiguous assignment prose can be modified. It also
found that abandoned-session cleanup can wait forever on abort or idle, preventing forced disposal
and quarantine. The Cycle 3 focused suite now passes 27/27: structured multiline forms are redacted,
ambiguous multiword prose is byte-identical, and independently hung abort/wait hooks reach one
forced disposal plus quarantine within the shared cleanup bound without unhandled rejection. The
complete Shepherd suite passes 164/164; focused and all-production strict TypeScript pass against
the explicit Pi 0.80.6 installation; pinned offline RPC registers `pm-shepherd`; and diff,
immutable-base, and issue-owned scope checks pass. Parent orchestration owns fresh exact-head review
and integration; this lane did not invoke Go/connectors, `make verify`, live GitHub, merge, or review
bots.

Review at `b4061d4e1a1545b0c8810b14b510cf048385a567` found that the foreground/main cleanup path can
still skip disposal when abort or idle never settles, both for a session obtained during cleanup
grace and for an ordinary claimed session. It also found unquoted YAML flow-map and spaced
line-start `client_secret` gaps. Cycle 4 is active under a fresh strict test-only RED gate and the
same narrow Shepherd-only verification boundary. RED is now captured with 23 passes and 8 expected
failures; production remained unchanged. The focused suite now passes 31/31 after independently
bounded abort/idle phases with unconditional exactly-once disposal and a linear flow-aware scanner
for spaced structured `client_secret` values. Complete Shepherd, all-production strict TypeScript,
pinned offline RPC, and diff/scope gates remain pending.
