---
name: go-implementation
description: Production Go standards for pm CLI and connector-engine work — errors, interfaces, concurrency/context, naming, testing, JSON serialization, CLI ergonomics. Load BEFORE writing or reviewing any Go in cmd/** or internal/**; every rule is cited to Uber/Google style guides, Go team sources, Dave Cheney, Ardan Labs, or clig.dev.
---

# Go implementation standards (pm CLI + connector engine)

Load this skill before implementing or reviewing Go. It distills the sources this repo treats as
authoritative into rules you apply directly; the full cited rule list lives in
[references/go-rules.md](references/go-rules.md) — consult it when a decision is contested.

## How to apply

1. Read the repo overlay below first — it binds the general rules to this codebase's gates.
2. Skim the themed sections of `references/go-rules.md` relevant to your change (errors,
   interfaces, concurrency, naming, testing, JSON, CLI).
3. Cite the rule number in review findings and dispositions (e.g. "go-rules #17: goroutine has no
   stop condition").

## Repo overlay (binds rules to this codebase)

- **Gates are the floor, not the bar.** `make verify` enforces gofmt (mutate-then-diff), go vet,
  golangci-lint (govet, staticcheck, errcheck, ineffassign, unused, misspell — scoped to
  connector-architecture packages), tests, build, docs-check, smoke, connectorgen-validate. This
  skill covers what the linters cannot: design, lifetimes, API shape.
- **JSON rules #37–39 are load-bearing here**: the connector engine round-trips JSON bundles;
  nil-vs-empty slices and `omitempty`-on-struct mistakes silently corrupt `--json` output and
  generated website data. Initialize slices explicitly where `null` would break a consumer;
  optional sub-objects are pointers.
- **CLI rules #40–46 restate repo contracts**: stdout machine / stderr human, `--json` parity with
  human output, bare namespaces exit 0 with help, destructive ops behind plan→preview→approval→
  execute. A change violating these also violates
  `.agents/agentic-delivery/references/cli-help-docs-website-parity.md`.
- **No mutable package-level registries** (#31): bundle definitions load via `bundleregistry.New()`
  — never add init-time global state to the engine; it breaks reload and test isolation.
- **Interfaces stay consumer-side and small** (#10–12): the engine's extension points
  (hooks, native overrides) follow this; do not add producer-side interface layers "for mocking".
- **Every goroutine needs an owner and a stop condition** (#17–19): engine code runs inside
  long-lived `pm` processes and worker loops; a leaked goroutine per connector read is a fleet
  problem at 548 bundles.

## Verification before handoff

`gofmt -l cmd internal` clean · `go vet ./...` · focused package tests · `go build ./cmd/pm` ·
`go run ./cmd/connectorgen validate internal/connectors/defs --json` when defs touched ·
`make verify` when feasible. Record results in the TDD ledger / handoff.
