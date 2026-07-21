# Summary — Issue #475

Status: exact-head correction Cycle 6 complete; every declared Shepherd-only gate passes and Cycle
5 evidence is superseded.

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

The focused suite passes 40/40 and the complete Shepherd suite passes 177/177. Strict owned plus
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
line-start `client_secret` gaps. Cycle 4 captured a fresh strict test-only RED with 23 passes and 8
expected failures while production remained unchanged. At implementation head
`01b42ae168176956d864ff10f40d1c981f37ac04`, the focused suite passes 31/31 after independently
bounded abort/idle phases with unconditional exactly-once disposal and a linear flow-aware scanner
for flow-map and spaced structured `client_secret` values. Two adversarial refactor probes each
captured a targeted 0/1 RED before closing apostrophe quote-state and nested-flow hiding gaps. The
complete Shepherd suite passes 168/168; both strict TypeScript scopes, pinned offline RPC, diff,
immutable-base, and issue-owned path checks pass. Parent orchestration owns fresh independent
exact-head review and integration; this lane did not invoke Go/connectors, `make verify`, live
GitHub, merge, or review bots.

Review at `e41f075a9b3bfb01d410296712740b54f943ba71` found a referenced deadline timer can survive an
immediate duplicate/concurrency rejection because `CancellationScope` is constructed before
reservation. It also found that accumulated redactor traversal state mishandles nested flow values,
leading apostrophe prose, and ordinary braces/comments. Cycle 5 replaces that traversal with an
explicit deterministic line/flow lexical state machine and moves reservation ahead of scope
creation. Production remains unchanged until the timer, all four redaction consumers, and
byte-identical prose controls produce the expected committed RED.

The Cycle 5 focused RED now exits 1 with 29 passes and 7 expected failures, while focused strict
TypeScript passes and production remains unchanged. The failures independently expose timer
ownership plus prompt, handoff, direct nested-flow, direct apostrophe, brace/comment control, and
typed-tool redaction boundaries.

At implementation head `8ff2d9631809d09db26811b4cd1335b92a9c457c`, Cycle 5 passes 36/36
focused and 173/173 complete Shepherd tests. Admission checks precede scope construction, and a
typed monotonic lexical state machine owns newline quote reset, comment/prose discrimination, flow
nesting, and balanced nested-value consumption. Both strict TypeScript scopes, explicit Pi 0.80.6
offline RPC, diff, immutable-base, and issue-owned path gates pass. Parent orchestration owns fresh
exact-head review and integration; this lane did not invoke Go/connectors, `make verify`,
runtime-backed services, live GitHub, merge, or review bots.

Review at `d918617a19749cd16d6bfcf3d2fee3e5146e7380` found three narrower transformer invariants:
nested value-local delimiter ownership stops at a newline, punctuation-adjacent apostrophes can
open false quote state, and per-assignment line-end searches rescan large single-line suffixes.
Cycle 6 keeps lifecycle code unchanged and corrects the existing typed scanner. Production remains
locked until all five consumer boundaries plus a deterministic 25/50/100 KiB scanner-work guard
produce the expected committed RED.

The Cycle 6 focused RED exits 1 with 33 passes and 7 expected failures, while the safe apostrophe
control and focused strict TypeScript pass and production remains unchanged. The failures expose
prompt, handoff, `workspace_read`, typed-capability, direct multiline-nested, direct punctuation-
apostrophe, and deterministic scan-metric boundaries.

Cycle 6 GREEN now passes 40/40 focused tests and focused strict Pi 0.80.6 TypeScript. The same typed
scanner carries a value-local closer stack across lines, distinguishes a YAML sequence marker from
a word-internal hyphen, and reuses the current line end for assignment decisions. Deterministic
line-boundary visits equal the 25,618 / 51,218 / 102,418-byte input sizes. Full declared
verification now passes at implementation head `93314a54302e84e053ad0d6ff44371fbf1a167e0`:
177/177 complete Shepherd tests, both strict TypeScript scopes, explicit Pi 0.80.6 offline RPC,
diff, immutable-base, and issue-owned scope checks. Parent orchestration owns fresh independent
exact-head review and integration; this lane did not invoke Go/connectors, `make verify`,
runtime-backed services, live GitHub, merge, or review bots.
