# Context — issue-3752-rate-limit-admission-r1

## Contract read

- Parent delivery issue: [#3752](https://github.com/polymetrics-ai/cli/issues/3752), read in full on 2026-08-06.
- Dependency sub-issue: [#3751](https://github.com/polymetrics-ai/cli/issues/3751), read in full on 2026-08-06 and implemented first.
- Foundation brief: [#3750](https://github.com/polymetrics-ai/cli/issues/3750), read in full on 2026-08-06.
- Deferred seams read for compatibility: [#3753](https://github.com/polymetrics-ai/cli/issues/3753), [#3754](https://github.com/polymetrics-ai/cli/issues/3754), and [#3755](https://github.com/polymetrics-ai/cli/issues/3755).

## Evidence established before planning

- The confirmed defect is in `internal/connectors/connsdk/http.go:853-862`: a valid `Retry-After` duration is clamped by `Requester.maxBackoff()` (default 30 seconds). A provider reset is deterministic and must be honoured exactly.
- `Requester.doWithBody` makes the JSON, form, and multipart HTTP call at `internal/connectors/connsdk/http.go:739`; its `Requester` shallow clones at `internal/connectors/engine/write.go:496` preserve added interface fields. `Requester.DoStream` has its own send loop at `internal/connectors/connsdk/stream.go:106` and must receive the same admission/observation contract.
- Existing read-only pacing is page-loop local: `internal/connectors/engine/read.go:275-276` creates the old limiter and `:309-313` waits only after page one. It is intentionally not changed in this foundation; #3753 replaces it by attaching the requester contract across engine paths.
- `internal/connectors/engine/bundle.go:260-263` and `engine/schema/{metadata,streams}.schema.json` contain the old informational / page-only declarations. The audit's exact count is reproducible: `rg -l '"rate_limit"' internal/connectors/defs/*/metadata.json | wc -l` returned `11` on 2026-08-06. No legacy bundle becomes mandatory or is migrated here.
- Bundle files are typed and meta-schema-validated by `engine.Bundle`, `Load`, and `metaSchemas` in `internal/connectors/engine/bundle.go:27-42, 978-1176`; production definitions are embedded by `internal/connectors/defs/defs.go:17`. This is the #3751 extension point.

## Fixed decisions

1. `rate_limits.json` is an optional, closed bundle file. It has explicit `declared`, `unknown`, and `not_applicable` states, so absence is not retroactively made invalid.
2. Every `declared` policy requires a source URL and ISO date-only retrieval date; an optional provider version adds precision but cannot replace the date.
3. A policy declaration can express selectors for endpoint, tier, and auth type; separate burst/sustained budgets; fixed-window, token-bucket, leaky-bucket, and cost-unit models. This records provider truth without claiming every model is already enforced.
4. The policy scope names only a non-secret subject kind: account, installation, application, endpoint, or IP. #3754 must construct any key from credential binding + policy ID + this declared non-secret subject; it must never derive or persist a key from credential material.
5. `connsdk.Requester` receives small context-aware admission and observation interfaces. This foundation does not instantiate a registry, resolve policies, or emit CLI events.
6. `Retry-After` is authoritative and uncapped. Exponential fallback is capped then receives bounded full jitter; deterministic provider reset waits never receive jitter.
7. A terminal HTTP 429 is a typed `RateLimitError` that unwraps the existing safe `HTTPError` and carries the reset timestamp when the provider supplied one. No new redaction or masking path is introduced.

## Non-goals and handoff seams

- #3753 owns the policy resolver and all engine attachment: read, check, direct operations, write forms, retry behavior, and removal of page-only pacing. This branch must not edit `engine/read.go`'s request wiring, `engine/write.go`, `engine/direct_read.go`, `engine/direct_write.go`, or `commandrunner` for rate-limit activation.
- #3754 owns in-process/shared registries, the opaque scope-key implementation, shared availability policy, and `require-shared` refusal. The `RateLimitAdmission` interface is its seam.
- #3755 owns human/JSON event rendering, wait/deadline messaging, CLI help, manual, website parity, and inspect/operator surfaces. The `RateLimitObserver` / `RateLimitObservation` types are its seam.
- No provider call, credential, connector declaration migration, generic HTTP write capability, output-redaction declaration, or `availability: implemented` command change is in scope.
