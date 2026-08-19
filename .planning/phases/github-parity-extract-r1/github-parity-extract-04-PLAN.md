---
phase: github-parity-extract-r1
plan: "04"
type: tdd
wave: 4
depends_on: ["01", "02", "03"]
files_modified:
  - internal/connectors/connectors.go
  - internal/connectors/connsdk/http.go
  - internal/connectors/connsdk/http_test.go
  - internal/connectors/engine/direct_read.go
  - internal/connectors/engine/direct_read_paginate.go
  - internal/connectors/engine/direct_read_test.go
  - internal/connectors/commandrunner/runner.go
  - internal/connectors/commandrunner/runner_test.go
  - cmd/connectorgen/validate.go
  - cmd/connectorgen/main_test.go
  - cmd/connectorgen/surfacesync.go
  - cmd/connectorgen/surfacesync_test.go
  - scripts/gen-github-parity.py
  - internal/connectors/defs/github/api_surface.json
  - internal/connectors/defs/github/operations.json
  - internal/connectors/defs/github/cli_surface.json
  - .planning/phases/github-parity-extract-r1/TDD-LEDGER.md
  - .planning/phases/github-parity-extract-r1/VERIFICATION.md
autonomous: true
requirements: []
must_haves:
  truths:
    - "D-14: GitHub's nine documented 204 status checks, text-response reads, and Markdown raw-text request enter only through bounded declared operation contracts; no final parity claim is made while other classified gaps remain."
    - "A status-only operation succeeds only on an empty successful response and returns an explicit nil body rather than attempting JSON decoding."
    - "A text-response operation returns only valid UTF-8 within its declared byte cap; JSON response policies retain their existing redaction behavior."
    - "The sole raw-body shape is an operation-declared POST text/plain request with a root string schema and one exact CLI body mapping; arbitrary raw HTTP bodies remain impossible."
    - "Every promoted GitHub endpoint is regenerated from its source bundle generator, reaches commandrunner's real preflight, and maps to exactly one documented api_surface row."
    - "D-15: This response-contract slice does not promote OAuth application endpoints; no ordinary GitHub bearer credential is reused for their documented Basic-auth contract, and their later promotion must remain secret-withholding and fail-closed."
    - "D-13: Inline execution is a documented GSD fallback because isolated role delegation is unavailable and unauthorised; tests record real RED then GREEN evidence."
  artifacts:
    - path: "internal/connectors/defs/github/api_surface.json"
      provides: "GitHub-only ledger promotion for status/text direct reads."
    - path: "internal/connectors/defs/github/operations.json"
      provides: "Bounded declared REST read contracts, including the raw Markdown input contract."
  key_links:
    - from: "internal/connectors/commandrunner/runner.go"
      to: "internal/connectors/engine/direct_read.go"
      via: "operation-specific path/query/JSON-or-declared-plain-text body mapping"
      pattern: "OperationDirectReadRequest"
    - from: "scripts/gen-github-parity.py"
      to: "internal/connectors/defs/github/api_surface.json"
      via: "generated covered_by direct-read declarations"
      pattern: "direct_reads"
---

<objective>
Close the first bounded response/body foundation slice that prevents documented GitHub reads from
being executable: nine empty 204 membership/status reads, four text-response reads, and Markdown
raw's exact plain-text request body.

Purpose: remove genuine executor limitations rather than relabeling documented operations.
Output: red/green-tested closed operation contracts and mechanically generated GitHub declarations.
</objective>

<must_haves>
<truths>
- D-14: GitHub's nine documented 204 status checks, text-response reads, and Markdown raw-text request enter only through bounded declared operation contracts; no final parity claim is made while other classified gaps remain.
- A status-only operation succeeds only on an empty successful response and returns an explicit nil body rather than attempting JSON decoding.
- A text-response operation returns only valid UTF-8 within its declared byte cap; JSON response policies retain their existing redaction behavior.
- The sole raw-body shape is an operation-declared POST `text/plain` request with a root string schema and one exact CLI body mapping; arbitrary raw HTTP bodies remain impossible.
- Every promoted GitHub endpoint is regenerated from its source bundle generator, reaches commandrunner's real preflight, and maps to exactly one documented api_surface row.
- D-15: This response-contract slice does not promote OAuth application endpoints; no ordinary GitHub bearer credential is reused for their documented Basic-auth contract, and their later promotion must remain secret-withholding and fail-closed.
- D-13: Inline execution is a documented GSD fallback because isolated role delegation is unavailable and unauthorised; tests record real RED then GREEN evidence.
</truths>
</must_haves>

<threat_model>
| Threat | Mitigation | Verification |
|---|---|---|
| Empty provider success is mistaken for malformed JSON | A distinct `none` read response policy requires a bounded empty response and produces `nil`. | A 204 fixture passes; a nonempty `none` response fails closed. |
| Text mode turns direct read into binary/unbounded transport | Permit valid UTF-8 only, enforce the operation byte cap before decoding, and keep binary downloads in their existing executor. | Fixtures reject oversized and invalid UTF-8 text. |
| A raw-body feature becomes generic HTTP | Accept only operation-declared POST `text/plain`, a root string schema, and exactly `maps_to: body`; reject mixed/dotted body mappings. | Engine, runner, and validator tests reject JSON/raw mixing and unsupported media types. |
| A CLI row claims execution without runtime support | Keep commandrunner preflight as the source of truth; generator output must pass it. | `TestEveryImplementedCommandPassesRuntimePreflight` and GitHub surface tests pass. |
| Hand-edited generated artifacts drift | Modify the generator source, run it, then run `surface-sync`; never hand-edit its emitted GitHub declaration output. | Generator output and `surface-sync --check` are clean. |
</threat_model>

<tasks>
  <task type="tdd">
    <name>RED — express bounded empty and text direct-read response contracts</name>
    <read_first>
      - internal/connectors/engine/direct_read.go
      - internal/connectors/engine/direct_read_paginate.go
      - internal/connectors/connsdk/http.go
      - internal/connectors/engine/direct_read_test.go
    </read_first>
    <action>
      Add focused local-server tests before implementation: a 204 operation read using `none`, a
      nonempty `none` response rejection, a bounded valid UTF-8 text response, and an invalid or
      oversized text rejection. Run the focused test command while the policies do not exist and
      record the actual failure in the TDD ledger.
    </action>
    <acceptance_criteria>
      - The RED test fails because current direct reads insist on JSON decoding or reject the new policy.
      - Test fixtures use no provider credential and assert status, nil/text body, and byte bounds.
    </acceptance_criteria>
  </task>
  <task type="tdd">
    <name>GREEN — implement closed status/text response handling</name>
    <read_first>
      - internal/connectors/connectors.go
      - internal/connectors/commandrunner/runner.go
      - cmd/connectorgen/validate.go
      - cmd/connectorgen/surfacesync.go
    </read_first>
    <action>
      Thread the declared output policy into the bounded direct-read walk. Add only `none` and
      `text` response policies, update commandrunner/connectorgen support in lockstep, and retain
      existing JSON decoding/redaction for every existing policy. A response governed by `none`
      must contain no body; `text` must be valid UTF-8 and remain bounded by `rest.max_bytes`.
    </action>
    <acceptance_criteria>
      - The focused engine and commandrunner tests pass.
      - Existing JSON direct-read regression tests still pass unchanged.
      - `connectorgen validate` rejects unsupported direct-read output policies.
    </acceptance_criteria>
  </task>
  <task type="tdd">
    <name>RED/GREEN — add the single declared plain-text POST input shape</name>
    <read_first>
      - internal/connectors/connsdk/http.go
      - internal/connectors/engine/direct_read.go
      - internal/connectors/commandrunner/runner.go
      - cmd/connectorgen/validate.go
    </read_first>
    <action>
      First add a local-server test for a `POST /markdown/raw` operation with `content_type:
      text/plain`, a root string body schema, and one `maps_to: body` CLI flag. It must fail before
      a request is sent on current code. Then add the narrow requester method and mapping path:
      only a declared plain-text POST can send raw text, it shares the existing requester (and rate
      admission/observation) path, and it rejects missing/mixed/oversized bodies before dispatch.
    </action>
    <acceptance_criteria>
      - Test observes literal plain-text bytes and `Content-Type: text/plain`, never a JSON string.
      - A JSON operation cannot use `maps_to: body`; a text operation cannot use `body.field`.
      - The request is still routed through `RequesterFor`/`Do*Limited`, not a bypass transport.
    </acceptance_criteria>
  </task>
  <task type="tdd">
    <name>GREEN — regenerate GitHub status/text commands from the generator</name>
    <read_first>
      - scripts/gen-github-parity.py
      - internal/connectors/defs/github/api_surface.json
      - internal/connectors/defs/github/operations.json
      - internal/connectors/defs/github/cli_surface.json
      - cmd/connectorgen/github_api_surface_test.go
    </read_first>
    <action>
      Extend the GitHub generator's explicit classified-endpoint table, not a broad default, for the
      nine 204 checks and four text/Markdown operations. Generate their operation, command, flags,
      response policy, request body contract, and `covered_by` linkage; run the generator and
      `connectorgen surface-sync` rather than editing emitted JSON. Update surface inventory
      expectations only by the generated delta and assert each named endpoint's executable
      preflight contract.
    </action>
    <acceptance_criteria>
      - Ledger blocked count decreases by exactly 13 from this slice; every promoted row is GitHub.
      - `go run ./cmd/connectorgen validate internal/connectors/defs` is clean.
      - `go run ./cmd/connectorgen surface-sync --check` is clean after regeneration.
      - `TestEveryImplementedCommandPassesRuntimePreflight` passes.
    </acceptance_criteria>
  </task>
</tasks>

<verification>
- focused red/green tests in `internal/connectors/engine`, `connsdk`, `commandrunner`, and `cmd/connectorgen`
- `go test -timeout 20m ./internal/connectors/engine/ ./internal/connectors/connsdk/ ./internal/connectors/commandrunner/ ./cmd/connectorgen/`
- `go run ./cmd/connectorgen validate internal/connectors/defs`
- `go run ./cmd/connectorgen surface-sync --check`
- `go build ./cmd/pm`
</verification>

<success_criteria>
- Every status/text/raw-Markdown operation has a narrow declared contract, not a generic transport exception.
- GitHub's ledger and command surface are generated from the changed source generator and survive real preflight.
- The phase ledger records the actual failing test before each implementation change and its green result afterwards.
</success_criteria>
