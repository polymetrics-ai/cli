# VERIFICATION — issue #3995 shared connector-certification Shepherd gate

## Status

Inline/manual GSD verification and code review are complete. The final `no-mistakes` gate and
child-PR publication remain pending.

## Required acceptance evidence

- [x] R1 RED fails before evaluator implementation and records the exact expected GitHub ID.
- [x] R1/R2 all-green fixture returns `PROCEED`; each isolated criterion defect names its exact ID.
- [x] Invalid/missing/unknown schema, unknown field, missing evidence pointer/sidecar, and omitted
      adapter gate field fail closed.
- [x] Current #3984 GitHub generated baseline returns `RETRY`, while generic contract check passes.
- [x] Evaluation is read-only: no evidence creation, credential access, provider call, or
      `cmd/connectorgen/certification*.go` edit.
- [x] Four generated harness projections contain equivalent canonical gate I/O schema; sync/check
      pass and deliberately missing/drifted projection tests fail.
- [x] Required transition boundaries reject non-`PROCEED` verdicts without discarding exact IDs.
- [x] Focused Go, generator, formatting, static/build, individual repository, and workflow evidence
      gates pass: `go test -timeout 20m ./internal/agentcontract -count=1`,
      `go test -timeout 20m ./cmd/agentcontractgen -count=1`,
      `go test -timeout 20m ./internal/cli -count=1`, `go vet ./...`, `go build ./cmd/pm`,
      `make lint`, `go run ./cmd/agentcontractgen sync`,
      `go run ./cmd/agentcontractgen check`, `make agent-contract-check`,
      `scripts/verify-gsd-workflow origin/feat/3988-github-certification`, and `git diff --check`.
      Earlier individual repository gates also passed: `tidy-check`, `docs-check-no-build`,
      `smoke-no-build`, `connectorgen-validate`, `connectorgen-surface-sync`,
      `connectorgen-certification-matrix`, `connector-canon-check`, `connector-boundary`, and
      `release-workflow-check`.
- [x] Inline GSD verify/gap loop and code review record every finding/disposition. `REVIEW.md`
      records two findings, both fixed in correction round 1 of 5; no actionable review findings
      remain.
- [ ] `no-mistakes` runs without `--yes`; final branch/PR target and parent references are correct.

## Correction round 2 focused evidence

- [x] Empty command/target/credential fingerprint sequences halt, while reordered proof JSON object
      members compare semantically and a valid fixture proceeds.
- [x] Missing sync-mode/primitive or flow-pair topology, connector-roster disagreement, symlinked
      artifact/evidence paths, and non-regular evidence records halt before certification.
- [x] A producer-valid false delivery guarantee with a named limitation proceeds; generic or
      unmatched limitations still halt.
- [x] The canonical `agentcontractgen certification-gate` argv is embedded in each Claude, Codex,
      Pi, and OpenCode projection and the checked-in GitHub baseline exits nonzero with its
      deterministic `RETRY` JSON.
- [x] Focused review verification passed:

      ```sh
      go test -timeout 20m ./internal/agentcontract ./cmd/agentcontractgen -count=1
      ```

      Both packages reported `ok`. Full suite, lint, CI, PR, and outer pipeline validation were not
      run in this review phase.
- [x] CLI help/manual/website parity is not applicable: this adds a repository-internal
      `agentcontractgen` transition command, not a `pm` command surface.

## Correction round 3 focused evidence

- [x] The shared producer/consumer flow-kind catalog rejects omitted, added, and remapped kinds.
- [x] Capability/workflow/sync/flow completion reports, status artifacts, and baseline aggregates
      halt when they disagree with matched evidence.
- [x] Malformed pointers retain a trusted cell coordinate and only expose a canonical safe record.
- [x] Every protected transition rejects missing, relative, traversing, symlinked, and
      symlink-contract roots; `certification-gate --help` emits `HALT` and exits nonzero.
- [x] Focused package tests, certification-generator checks, projection synchronization/check, and
      the explicit four-transition command tests pass without provider, credential, evidence
      creation, network, production mutation, full-suite, lint, CI, PR, or outer-pipeline work.

      ```sh
      go test -timeout 20m ./internal/agentcontract ./cmd/agentcontractgen -count=1
      go test -timeout 20m ./cmd/connectorgen -run '^(TestCertification|TestRunCertification)' -count=1
      go run ./cmd/agentcontractgen sync --root "$PWD"
      go run ./cmd/agentcontractgen check --root "$PWD"
      go run ./cmd/connectorgen certification-matrix --check
      go test -timeout 20m ./cmd/agentcontractgen -run '^TestRunCertificationGate(BlocksEveryProtectedTransition|RejectsUntrustedRootsForEveryTransition|HelpBlocksWithoutProceedVerdict)$' -count=1
      ```

## Correction round 4 focused evidence

- [x] No `cmd/connectorgen/certification*.go` path differs from the base; the unchanged producer
      remains checked against `flow-matrix.json` by `connectorgen certification-matrix --check`.
- [x] `agentcontractgen sync/check` generates and verifies the importable consumer flow-kind
      catalog directly from that matrix, with no hand-maintained duplicate inventory.
- [x] Overrides cannot promote immutable facts, every raw and override pointer is bound before
      report derivation, and exact large JSON-number mismatches halt.
- [x] Focused Green verification passed without full-suite, lint, CI, PR, provider, credential,
      network, evidence-creation, or outer-pipeline work:

      ```sh
      go run ./cmd/agentcontractgen sync --root "$PWD"
      go test -timeout 20m ./internal/agentcontract ./cmd/agentcontractgen -count=1
      go test -timeout 20m ./cmd/connectorgen -run '^(TestCertification|TestRunCertification)' -count=1
      go run ./cmd/connectorgen certification-matrix --check
      go run ./cmd/agentcontractgen check --root "$PWD"
      git diff --check da7747a796049601a179a97c025bfb05f011f1e8
      ```

## Correction round 5 focused evidence

- [x] Each exact raw flow pair coordinate now halts when its endpoint-role applicability,
      declared/implemented conjunctions, or derived not-applicable code/reason disagree with its
      pair facts; fixture/live evidence remains separately validated and overrides retain their
      correction-round-four immutable-base rule.
- [x] With `flow_gen.go` intentionally absent, stable catalog code compiled and
      `agentcontractgen sync` recreated one data-only catalog from `flow-matrix.json`; missing,
      empty, invalid, duplicate, or externally mutable catalog data fails closed.
- [x] The intentional canonical root-instruction change had left both deterministic render-hash
      expectations stale; they were refreshed from the current base and connector renderings
      before the final focused package run.
- [x] Focused Green verification passed without full-suite, lint, CI, PR, provider, credential,
      network, evidence-creation, or outer-pipeline work:

      ```sh
      go run ./cmd/agentcontractgen sync --root "$PWD"
      go test -timeout 20m ./internal/certificationcatalog ./internal/agentcontract ./cmd/agentcontractgen -count=1
      go test -timeout 20m ./cmd/connectorgen -run '^(TestCertification|TestRunCertification)' -count=1
      go run ./cmd/connectorgen certification-matrix --check
      go run ./cmd/agentcontractgen check --root "$PWD"
      git diff --check da7747a796049601a179a97c025bfb05f011f1e8
      git diff --name-only da7747a796049601a179a97c025bfb05f011f1e8 -- 'cmd/connectorgen/certification*.go'
      ```
