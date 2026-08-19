# TDD Ledger: Issue #4307

## Planned Red/Green Evidence

| Slice | Red | Green | Edge/regression proof |
| --- | --- | --- | --- |
| Declaration-owned headers | Synthetic operation declarations cannot represent a typed request header or command mapping. | One exact declared `header.<name>` map becomes a validated request header. | Two identities prove header/path/query/body and operation boundaries remain isolated. |
| Header hardening | Forbidden header variants can flow through request construction or invalid input reaches a provider double. | Canonical protected-header/name/value validation refuses before I/O. | Case/normalization, duplicate, unknown, CR/LF, malformed, byte, enum/pattern/length, requiredness, and runtime-owned failures leave a request counter at zero. |
| F4 kind admission | Valid synthetic status/text/binary/multipart declarations fail at an unregistered loader, validator, generator, or runner mirror. | Every named kind reaches exactly its existing typed executor through the installed command path. | Invalid/bare/unbounded declarations fail before I/O and cannot fall through to generic REST handling. |
| Download and text output | Existing primitive cannot prove exact media/redirect/cap/output atomicity through a declaration. | A bounded declared response produces only the expected final output. | Over-cap, wrong media/charset, redirect-policy, stream, unsafe-path, and replacement failure yield no partial-success file. |
| Upload and multipart | Caller data can select arbitrary parts/content types or evade declaration limits. | Declared parts, media, file/count/byte caps, approval, and destructive confirmation construct the exact request. | Unknown parts/raw bytes/unbounded file/changed approval digest fail before execute I/O. |
| Declared result preservation | Fixed-operation results drop unusual ordinary provider fields because a policy guesses they are irrelevant. | Every status/header/body/field element admitted by the exact declared response contract reaches the result unchanged. | Credential and transport-secret canaries retain their field name with the established explicit masking marker; the test proves no other scope/tier/risk filter exists. |
| Existing surfaces | New common paths can change current typed command semantics. | Existing operation families retain their present output and preflight behaviour. | GraphQL, scalar/form/SCIM, #4305 structured bodies, credential/auth redaction, and no-credential preflight tests stay green. |

## Actual Evidence

Pending implementation. Each entry will record the exact failing test command (Red), passing command (Green), files changed, and observable no-I/O or provider-double assertion.
