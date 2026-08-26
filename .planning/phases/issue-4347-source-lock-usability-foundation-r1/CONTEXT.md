# Issue 4347 — Source-lock usability foundation

## Task Delivery Header

- Issue: Refs #4347 — retain pinned source artifacts for hermetic verification.
- Base branch: main.
- Merges into: main.
- Delivery: A pull request from `fm/cli-source-lock-usability-foundation-r1` is open against `main`, with TDD, verification, and review evidence recorded.
- Working branch: fm/cli-source-lock-usability-foundation-r1.
- Task: Make `connectorgen source-retain` preserve bytes without import-time schema/inventory validation; support the existing operation and parity source-lock shapes; distinguish byte and canonical-JSON identity; reject bad source responses before declaring drift; preserve strict source-import identity in operation evidence; and prove the eight live exact locks retain from the built utility.
- Verification: Focused `cmd/connectorgen` red/green tests; v3 operation-evidence equivalence/impact tests; a built `connectorgen` binary retaining fastly, github, hubspot, pipedrive, shipstation, squarespace, woocommerce, and zendesk-support; `go vet`; targeted generator checks; generated-file/snapshot checks; and the applicable repository gates run separately with explicit timeouts.

## Evidence Table

| Acceptance criterion | Evidence | Observable assertion or fake reason |
| --- | --- | --- |
| Retention accepts an otherwise import-invalid lock with no form pin | live | The retain command writes a digest-addressed artifact and provenance manifest while `parseSourceImportLock` still rejects the same incomplete v3 inventory. |
| Existing parity locks can be retained | live | The retain loader identifies the connector-owned `*-parity-source-lock.json` and writes a retained record without changing the lock bytes. |
| Reordered JSON is verified by canonical identity | live | Two differently ordered JSON serialisations produce unequal byte digests but the same canonical digest; retention and retained-artifact verification accept the second only when the lock selects `canonical_json`. |
| Byte-pinned sources remain byte-verified | live | A changed byte stream under a byte identity still fails with an explicit drift error and no retained artifact is created. |
| Bad sources are distinct from drift | live | HTTP 403 and a response drastically smaller than its lock each produce `wrong source` output, never `source-lock refresh required`. |
| Plausible login/error bodies are distinct from drift | fake | Hermetic HTTP responses carry the actual MIME header and a plausible-size HTML login/error body; no provider request is authorized for this adversarial transport boundary. The retained command rejects it as `wrong source` before digest comparison. |
| Operation evidence consumes strict v3 document identity | fake | A hermetic v3 rewrite of the checked-in GitHub lock runs through `parseSourceImportLock`; operation evidence retains every row, all six classification lanes, and both a declared and a deferred row. Provider bytes are not needed to test decoder/evidence identity. |
| Rendered citations are operation-specific | fake | A hermetic rendered-reference lock rejects its generic publication URL and accepts only a fragment pointer or a capture/hash/location extraction binding. No citation URL is fetched by source import. |
| Eight exact live locks retain successfully | live | The built utility exits zero for each named connector and each resulting retained manifest/artifact proves the fetched bytes against the selected identity. |

## Delivery constraints and decisions

- This is shared generator tooling for #4347, not a connector lane. No connector definition, source lock, or generated declaration is changed by this PR.
- The Firstmate inbox closed the prior storage decision on 2026-08-24: tracked raw source bytes are the eventual retention policy, provenance/terms are data rather than an enforcement gate, and this worker executes GSD inline because role spawning is prohibited.
- The supplied local parity locks and `sources/` directories are ignored/untracked. Preserve their existing lock files; live retention is an evidence run only and any resulting ignored artifacts are not committed from this branch.
- Firstmate live evidence, 2026-08-24: exact (fastly 2,341,028; github 12,920,264; hubspot 4,132,728; pipedrive 1,782,400; shipstation 186,490; squarespace 345,839; woocommerce 4,400,931; zendesk-support 1,757,202); five landing-page sources; two unreachable sources; six real, modest drifts.
- A re-pin remains a separate explicit lock edit. Retention never mutates a lock; canonical verification must report both the lock's stored byte digest and the fetched artifact's byte digest when they differ.
- Independent audit R1 at PR head `f9d237840ed49737ef92667b7e2173465dcc000d` reopened the phase on 2026-08-26. The repair must keep strict import, public-URL limits, byte/canonical identity, no credential admission, and no live provider I/O. It must not broaden operation evidence from operation locks to parity locks or make source import fetch citation URLs.

## Required skills and lifecycle

Loaded: `golang-how-to`, `golang-cli`, `golang-testing`, `golang-error-handling`, `golang-security`, `golang-safety`, `golang-design-patterns`, `golang-structs-interfaces`, and `golang-documentation`.

GSD command path resolved with `scripts/gsd doctor`, `scripts/gsd sources` and `scripts/gsd prompt` for `discuss-phase`, `plan-phase --tdd`, `execute-phase`, `verify-work`, and `code-review`; the R1 gap closure additionally resolved `plan-phase --gaps` and `execute-phase --gaps-only`. `go run ./cmd/agentcontractgen check` is run from the clean nested worktree because this root intentionally preserves a nested diagnostic worktree that the command correctly sees as a duplicate project agent inventory. The Pi runtime cannot run the required compatible isolated roles and the repository contract prohibits role spawning, so this phase uses the documented inline/manual GSD fallback.
