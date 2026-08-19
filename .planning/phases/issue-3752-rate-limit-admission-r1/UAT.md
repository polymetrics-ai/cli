# UAT — issue-3752-rate-limit-admission-r1

Manual-GSD verification fallback: the Pi runtime cannot provide the official verifier role and the
delivery contract forbids spawning it. This is an automated, fixture-only UAT; no live provider,
credential, CLI, or browser behavior is in scope.

| ID | Deliverable | Automated evidence | Verdict |
| --- | --- | --- | --- |
| D1 | Provider-cited optional declaration contract | `TestBundleLoadParsesProviderCitedRateLimits`, malformed/state tables, and `go run ./cmd/connectorgen validate internal/connectors/defs` (550, zero findings) | passed |
| D2 | Future production declarations cannot be silently omitted from `defs.FS` | `TestProductionDefinitionsEmbedEveryRateLimitDeclaration` | passed |
| D3 | Exact provider reset, typed terminal 429, bounded fallback jitter | requester unit tests plus focused `-race` run | passed |
| D4 | Context-aware admission covers logical requester attempts; replayable reads may replay internally and strict writes cannot | JSON/form/multipart/stream no-send and cancellation, redirect, read-replay, and strict-write tests | passed |
| D5 | Deferred seams remain unactivated | changed-path audit: no `commandrunner`, engine read/write/direct-operation, connector migration, or CLI changes | passed |

Verdict: **passed** for the foundation’s automated deliverables. Full repository CI and the
no-mistakes validation/ship flow are deliberately firstmate-gated. The full connsdk `-race` suite
has an unchanged multipart test-data race; focused rate-limit and loader race suites pass.
