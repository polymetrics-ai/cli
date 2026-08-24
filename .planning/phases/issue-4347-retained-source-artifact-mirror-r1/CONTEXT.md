# Issue #4347 — retained source artifact mirror context

## Delivery header

- Issue: Refs #4347 — retain pinned source artifacts for hermetic verification.
- Base branch: main.
- Merges into: main.
- Delivery: A direct main-targeted PR that uses retained artifact bytes without a provider request, contains GSD/TDD evidence, and reports exact recovery limits for historic pins.
- Working branch: fm/cli-retained-source-artifact-mirror-r1.
- Task: Add a connector-owned, tracked retained-artifact reader for source locks. Preserve a valid pin; only Firstmate's narrowly authorized, per-artifact re-pin amendment may replace a real provider document's lock identity.
- Verification: Focused retained-artifact happy/bad/bundle/Elasticsearch/Zoom tests, source-import command tests, format checks, and applicable repository gates.

## Decisions and constraints

- Captain decision (inbox 001, 2026-08-24): retain provider bytes in tracked material; record URL, retrieval time, and known license/terms as data; do not add a license enforcement gate. Scope is all 28 nonzero locked artifacts (GitHub plus 27 Batch 8–10 inputs), after measuring their real footprint.
- Measurement completed before any artifact write: current live responses total 454,667,903 bytes; locked corpus totals 453,725,820 bytes; `df -k .` reported 33,781,928 KiB free. The footprint is safe.
- The source lock is immutable evidence unless the **Firstmate-authorized** amendment below applies. A retained file is accepted only when its observed byte count and SHA-256 equal the lock values it is retained alongside; a re-pin must first record both old and new values in the visible report.
- Retained data belongs at `internal/connectors/defs/<connector>/sources/artifacts/<lowercase-lock-sha256>.artifact`, beside the lock. A tracked provenance manifest records source URL, retrieval time, and terms/license text separately from the immutable lock.
- Builds and verification make no provider request. A missing or corrupt retained file is a terminal error naming the lock and retained path.
- Manual GSD fallback: the task contract forbids role spawning, so the resolved `discuss-phase`, `plan-phase --tdd`, `execute-phase`, `verify-work`, and `code-review` prompts are executed inline and recorded here rather than delegated.

## Historic recovery evidence

- Read-only current-response measurement finds 351 of the 917 Batch 8–10 source documents byte-mismatched against their locks; Auth0 is a single-document example: locked `5,655,921` bytes vs current `5,655,204` bytes. These live bodies must not be stored as replacements.
- GitHub REST still matches at `12,920,264` bytes; the locked GraphQL schema does not: locked `1,546,421` bytes vs current `1,551,372` bytes.
- A read-only scan of every reachable Git object whose byte length matched a locked input recovered only the DocuSeal artifact (`192,929` bytes, SHA-256 `7ac10d1c…e0e4b`), in the older PersistIQ validation artifact. It found no stored historic copy for the mismatching source inputs.
- Zoom's historic accounts source lock on `fm/cli-zoom-full-definition-mapping-r1` names a dated `_next/data/…/accounts.json` document. A fresh read-only request on 2026-08-24 returned HTTP `404` and `8,329` bytes, not its locked `805,789` bytes. A reachable-Git-object exact digest scan found no copy of that accounts body. The task’s known 404 therefore cannot be repaired by a refetch or a lock rewrite; tests record that a retained copy would recover a 404, while the delivery report states the actual historic bytes are unavailable.
- Elasticsearch's historic lock on `fm/cli-map-batch23-r1` is `6,458,869` bytes/SHA-256 `9b2ad824…dde9`; a read-only current request returned HTTP `200` but `6,458,837` bytes. The same reachable-Git-object exact digest scan found no retained matching copy. **Firstmate** authorized an explicit re-pin only if the fresh body passes the real-provider-document classification; record old/new identity and retrieval date.

## Firstmate re-pin amendment

- Firstmate (inbox 002, 2026-08-24) — not the captain — authorized per-artifact re-pins solely for a retrievable real provider document. The report must name Firstmate as the authorizer and record connector/artifact, old and new byte/SHA-256 values, and retrieval date.
- HTTP errors, redirects, login walls, and other non-document responses are never re-pin material. Zoom accounts is currently an HTTP 404/8,329-byte error response and remains irrecoverable; its historic 805,789-byte raw input is unavailable.
- Firstmate inbox 004 confines this foundation PR to the GitHub lock on current `main`. It must not import Batch 8–10, Elasticsearch, Zoom, or any other unmerged lane's connector files. Those lanes adopt the mechanism and retain their own source bytes when their connector PRs merge.
