# Verification Checklist — engine-shared-capabilities-r1

Filled in as the phase executes. `pending` until evidence exists.

## Local gates (AGENTS.md "Verification")

| Gate | Command | Result |
| --- | --- | --- |
| Format | `gofmt -l cmd internal` | pending |
| Vet | `go vet ./...` | pending |
| Tests | `go test ./...` | pending |
| Build | `go build ./cmd/pm` | pending |
| Repo verify | `make verify` | pending |

## Phase-specific gates

| Gate | Command / check | Baseline | Result |
| --- | --- | --- | --- |
| Connector boundary | `go run ./cmd/connectorgen boundary . --json` | `clean`, 0 findings | pending |
| GSD evidence | `scripts/verify-gsd-workflow` | n/a | pending |
| No connector bundle touched | `git diff --name-only main...HEAD -- internal/connectors/defs` is empty | n/a | pending |
| Binary cert gate still passes unmodified | `internal/connectors/certify` tests green, file unchanged | passes today | pending |

## Additive-and-opt-in assertions

| Invariant | How verified | Result |
| --- | --- | --- |
| Bounded-JSON direct-read path unchanged | `direct_read_test.go` green, `direct_read.go` diff limited to shared helpers | pending |
| Current write behaviour unchanged | `write_test.go` green; regression guards 3.5 and 4.6 | pending |
| Closed record schemas stay closed | regression guard 4.6 | pending |
| Every new bundle field optional | existing bundles validate unchanged; boundary + bundle tests green | pending |

## Security assertions (`golang-security`, `golang-safety`)

| Assertion | Evidence | Result |
| --- | --- | --- |
| Credentials never sent to a non-owned host | tests 1.2, 1.3, 1.4, 1.5 | pending |
| Custom auth headers stripped cross-origin (Go does not do this) | test 1.3 | pending |
| No unbounded read into memory | tests 2.1, 2.2 | pending |
| Path traversal and symlink escape contained by `os.Root` | tests 2.7, 2.8 | pending |
| No archive extraction path exists | test 2.4 | pending |
| No raw request escape hatch introduced | tests 4.2, 4.3, 4.4 | pending |
| Restrictive file permissions | test 2.12 | pending |
| Shared HTTP client not mutated | test 1.7 | pending |

## CLI help / docs / website parity

Per `.agents/agentic-delivery/references/cli-help-docs-website-parity.md`.

| Surface | Applicable? | Result |
| --- | --- | --- |
| Runtime `pm help <topic>` | pending assessment | pending |
| Bare namespace command behavior | pending assessment | pending |
| `docs/cli/**` | pending assessment | pending |
| `website/**` | pending assessment | pending |
| Generated help/manual artifacts | pending assessment | pending |

Expected outcome: **not applicable** — this phase adds engine/connsdk capabilities with no new CLI
command, flag, or output surface. No connector wires them yet, so no user-visible CLI behaviour
changes. To be confirmed, not assumed, before the PR.

## Open items carried out of the phase

- Alfred identity blocker: the issue tree could not be created (no `alfred-polymetrics-ai`
  credential in this environment; escalated to the captain, tracked as
  `cli-pipeline-alfred-identity-r1`). Draft bodies are held at `ISSUE-TREE-DRAFT.md` in this phase
  directory for creation the moment identity is resolved.
- Binary certification gate flip: deliberately deferred to first lane adoption. Rationale and both
  options recorded in `ISSUE-TREE-DRAFT.md`.
- Five captain decisions on download policy: built configurable, surfaced in `PLAN.md`, not silently
  picked.
