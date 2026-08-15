# Correction round 2 of 5 — issue #3995

## Tracking

- Child [#4024](https://github.com/polymetrics-ai/cli/issues/4024) remains limited to rejecting
  empty proof fingerprints and comparing proof `json.RawMessage` values semantically.
- Child #4028 owns invalid-artifact hardening: matrix topology and connector identity, supplied-root
  confinement, and producer-equivalent delivery limitations.
- Child [#4030](https://github.com/polymetrics-ai/cli/issues/4030) owns the executable, read-only
  `agentcontractgen certification-gate` transition boundary and its canonical argv in every harness.
- All three remain under #3995 with `Refs #3988`; this round does not alter correction-round-1
  history or merge a child PR.

## RED-to-GREEN record

The six reported correction cases were authored before their production changes. This review phase
permits one focused verification only, so those RED cases were not separately executed: they specify
empty fingerprint rejection, reordered JSON proof equality, omitted matrix topology,
escaped/non-regular inputs, producer-valid named delivery limitations, canonical projection argv,
and the real command baseline. The final focused command passed:

```sh
go test -timeout 20m ./internal/agentcontract ./cmd/agentcontractgen -count=1
```

It reported `ok` for both packages.

## Disposition

The evaluator now rejects empty fingerprint sequences, compares proof JSON values independent of
object-member order, requires complete connector/matrix topology, and reads artifacts/evidence only
through a root-bound reader that rejects symlink ancestors and non-regular records. A false delivery
guarantee is accepted only with the producer-valid named limitation. The canonical command emits the
deterministic JSON verdict and blocks every protected transition unless it receives `PROCEED`.

The checked-in GitHub baseline remains `RETRY` and retains
`capability/github/capability:check/live_evidence`; a complete generated fixture may return
`PROCEED`. Evaluation remains read-only and does not access providers, credentials, or production
state.
