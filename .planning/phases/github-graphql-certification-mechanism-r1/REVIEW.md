# Inline code review — GraphQL certification mechanism

Review route: inline/manual. The direct-PR brief and repository contract forbid spawning the canonical reviewer role, so the required code-review stage was completed against the changed files after all local gates passed.

Reviewed areas:

- definition model and JSON schema validation;
- embedded definition-source ownership and connector-boundary behavior;
- source-pinned GraphQL root/schema compiler and its bounded document reader;
- live assertion execution, secret scan, `not_live`, and `unexecutable` report semantics;
- deterministic sweep classification and generated artifact;
- regression tests, including the post-compilation false assertion and declared-unexecutable path.

Findings: none. The only review-adjacent issue was two staticcheck style findings from `make lint`; both were applied mechanically, and the lint gate then passed with zero issues.

Residual boundary: schema conformance proves only the fixed operation's declared root field and source argument binding against the source-pinned lock. It deliberately does not claim provider output values, authorization, or mutation effects; those retain live or fixture-bound non-pass results.
