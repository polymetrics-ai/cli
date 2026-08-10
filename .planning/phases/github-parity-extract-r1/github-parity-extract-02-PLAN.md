---
phase: github-parity-extract-r1
plan: "02"
type: tdd
wave: 2
depends_on: ["01"]
files_modified:
  - scripts/github-live-proof-sweep.mjs
  - scripts/tests/github-live-proof-sweep.test.mjs
  - internal/connectors/defs/github/cli_surface.json
  - internal/connectors/defs/github/docs.md
  - internal/connectors/engine/binary_download*.go
  - internal/connectors/engine/*binary*_test.go
  - .planning/phases/github-parity-extract-r1/TDD-LEDGER.md
autonomous: true
requirements: []
must_haves:
  truths:
    - "D-01: A GitHub-specific committed runner enumerates every implemented command and writes one redacted terminal record with an assertion or concrete reason."
    - "D-02: The runner allows only proven, untestable-with-reason, or failed; fixture replay and dispatch reachability are never terminal proof states."
    - "D-03: A reproduced codeload redirect or live-read failure is fixed in this PR with a focused test; unreproduced defects are documented rather than asserted by sample."
    - "D-04: The runner operates only against the dedicated private test repository and never persists or prints credentials, grants, or token-derived values."
  artifacts:
    - path: "scripts/github-live-proof-sweep.mjs"
      provides: "GitHub-only full live-proof enumeration and reporting harness."
    - path: "scripts/tests/github-live-proof-sweep.test.mjs"
      provides: "Deterministic redaction and complete-accounting regression suite."
  key_links:
    - from: "scripts/github-live-proof-sweep.mjs"
      to: "internal/connectors/defs/github/cli_surface.json"
      via: "implemented command enumeration"
      pattern: "implemented"
---

<objective>
Build a committed GitHub-only live-proof harness whose self-test proves full implemented-surface
enumeration, redacted result handling, explicit untestable reasons, and returned-data assertions;
fix GitHub binary-download redirect handling if the harness's codeload probe exposes it.

Purpose: make a complete live proof repeatable and prevent sample-based coverage claims.
Output: a tested script plus any smallest safe GitHub redirect fix discovered by its tests.
</objective>

<must_haves>
<truths>
- D-01: A GitHub-specific committed runner enumerates every implemented command and writes one redacted terminal record with an assertion or concrete reason.
- D-02: The runner allows only proven, untestable-with-reason, or failed; fixture replay and dispatch reachability are never terminal proof states.
- D-03: A reproduced codeload redirect or live-read failure is fixed in this PR with a focused test; unreproduced defects are documented rather than asserted by sample.
- D-04: The runner operates only against the dedicated private test repository and never persists or prints credentials, grants, or token-derived values.
</truths>
</must_haves>

<threat_model>
| Threat | Mitigation | Verification |
|---|---|---|
| Token/grant or provider body leaks into a report | Keep raw subprocess output in memory; persist only redacted command identity, HTTP status, and assertion summary. | Fixture test injects token-like output and proves it is absent from records/logs. |
| Partial sweep passes as full proof | Derive command list from GitHub `cli_surface.json` and fail if any `implemented` row lacks a terminal result. | Self-test uses a deliberately omitted command and expects failure. |
| Redirect forwards Authorization to codeload | Trace engine binary request policy and restrict redirect follow behavior to bounded, safe targets. | Unit test asserts redirect succeeds only under safe policy and sensitive headers are not forwarded. |
| A case overrides the dedicated repository target before a live write | Reject a write-case `--owner` or `--repo` value unless it exactly equals the runner's supplied dedicated test repository identity. | Deterministic Node test supplies a mismatched owner and expects case validation to fail before any subprocess is started. |
| Generic cross-provider harness expands scope | Place runner under `scripts/github-*`; it may read only GitHub definitions. | Test/scope check rejects a non-GitHub connector argument. |
</threat_model>

<feature>
  <name>GitHub live-operation proof harness</name>
  <files>scripts/github-live-proof-sweep.mjs, scripts/tests/github-live-proof-sweep.test.mjs</files>
  <behavior>
    The runner enumerates every `availability: implemented` GitHub command, invokes the built `pm`
    binary through a supplied credential name, and records exactly one terminal result per command.
    A result is valid only when it includes a returned-data assertion or a concrete verifiable
    untestable reason. It never writes credentials, approval tokens, raw output, or response bodies.
  </behavior>
  <implementation>
    Use Node built-ins already present in the repository; add no dependency. Provide `--self-test`
    for deterministic fixtures and a live mode that requires an explicitly named dedicated private
    test repository. Treat 429 as failure, measure `/rate_limit` without exposing authorization,
    and leave destructive confirmation semantics intact.
  </implementation>
</feature>

<tasks>
  <task type="tdd">
    <name>RED — define complete, secret-safe proof accounting</name>
    <read_first>
      - internal/connectors/defs/github/cli_surface.json
      - internal/connectors/defs/github/operations.json
      - internal/connectors/commandrunner/runner_test.go
      - .planning/phases/github-parity-extract-r1/github-parity-extract-RESEARCH.md
    </read_first>
    <action>
      Add a deterministic Node test fixture and test file before the runner. Specify a known
      implemented command set, one omitted command, fake raw output containing a token-shaped
      value, a data assertion, and an untestable reason. Run the test before the runner exists and
      record the RED failure in `TDD-LEDGER.md`.
    </action>
    <acceptance_criteria>
      - `node --test scripts/tests/github-live-proof-sweep.test.mjs` fails because the harness is absent.
      - The test demands all implemented commands and rejects raw token-shaped data in persisted output.
      - The RED result is recorded before implementation.
    </acceptance_criteria>
  </task>
  <task type="tdd">
    <name>GREEN — implement GitHub-only enumeration and proof records</name>
    <read_first>
      - scripts/package.json
      - internal/connectors/defs/github/cli_surface.json
      - internal/cli/cli.go
      - docs/cli/reverse.md
    </read_first>
    <action>
      Implement `scripts/github-live-proof-sweep.mjs` with `--self-test` and live modes. It must
      enumerate only GitHub's `implemented` rows; reject an unknown command or an incomplete result
      set; capture subprocess output in memory; write only redacted records with command path,
      terminal state, HTTP status where known, and assertion/reason. Live mode must require a built
      `pm` path, a credential name, and a dedicated private test repository identity. Use no raw
      token environment or file; do not print captured approval values.
    </action>
    <acceptance_criteria>
      - `node --test scripts/tests/github-live-proof-sweep.test.mjs` exits 0.
      - `node scripts/github-live-proof-sweep.mjs --self-test` exits 0 and reports no raw response or secret.
      - The self-test fails if any implemented row lacks `proven`, `untestable` with reason, or `failed`.
      - The runner never accepts another connector name.
    </acceptance_criteria>
  </task>
  <task type="tdd">
    <name>RED/GREEN — repair a codeload redirect only if reproduced</name>
    <read_first>
      - internal/connectors/engine/binary_download.go
      - internal/connectors/connsdk/rate_limit_requester.go
      - internal/connectors/defs/github/operations.json
      - internal/connectors/defs/github/api_surface.json
    </read_first>
    <action>
      First reproduce the prior GitHub binary-download failure with a local redirect fixture or the
      live safe binary candidate. If the existing path does not follow a legitimate GitHub codeload
      redirect, add a focused failing engine test, then make the smallest redirect-policy fix that
      preserves byte bounds, destination checks, and strips Authorization outside the original host.
      If it is already correct, record the proof and do not add a duplicate implementation.
    </action>
    <acceptance_criteria>
      - A test exists only when a real redirect defect is reproduced; it fails before and passes after the fix.
      - Redirect handling does not relax destination, byte limit, or header safety.
      - GitHub's binary candidate returns a bounded manifest/assertion rather than a redirect error.
    </acceptance_criteria>
  </task>
  <task type="tdd">
    <name>RED/GREEN — contain every executable write case to the dedicated repository</name>
    <read_first>
      - scripts/github-live-proof-sweep.mjs
      - scripts/tests/github-live-proof-sweep.test.mjs
      - internal/connectors/defs/github/cli_surface.json
    </read_first>
    <action>
      Add a deterministic failing case-validation test before changing the live runner: an
      executable GitHub write case that supplies a different `--owner` or `--repo` must be
      rejected before credential inspection, planning, preview, or execution. Then pass the
      production command metadata into case validation and accept only matching values (including
      `--flag=value` syntax). Reads remain eligible to use non-test public data; this containment
      rule is only for dispatchable writes. Record both commands and the behavioral RED/GREEN
      result in the TDD ledger.
    </action>
    <acceptance_criteria>
      - A mismatched write `--owner` or `--repo` fails case validation with no `pm` subprocess.
      - Exact matching values and writes that use the credential's configured target remain valid.
      - The Node regression suite and self-test pass after the change.
    </acceptance_criteria>
  </task>
</tasks>

<verification>
- `node --test scripts/tests/github-live-proof-sweep.test.mjs`
- `node scripts/github-live-proof-sweep.mjs --self-test`
- focused binary-download test when applicable
- `go test -timeout 20m ./internal/cli/ ./internal/connectors/engine/`
- `go build ./cmd/pm`
</verification>

<success_criteria>
- The committed runner cannot silently report a sample as full coverage.
- The runner's persisted artifacts contain neither credential material nor raw provider output.
- The known codeload problem is either fixed by a red/green test or proven not reproducible.
</success_criteria>
