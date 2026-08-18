# TDD ledger — GitHub mutation certification slice 4 writes-d

## Red

- Before recording a `certified` outcome, each command's acceptance predicate begins false: the provider read-back does not yet show the new `pm-cert-` tagged state.
- Cleanup is independently red until a direct provider DELETE has completed and a separate read returns 404 or omits the tagged object. A successful `pm` delete is explicitly insufficient because #4221 reproduces a false success.

## Green

- A command reaches green only after `pm ... plan`, `preview`, and approved `run` produce a provider-visible change that satisfies the record's declared or `agent_derived` predicate.
- Cleanup reaches green only when the independent direct-provider read-back proves absence.
- Evidence is retained only after `go run ./cmd/connectorgen certification-matrix --check` succeeds; invalid evidence is deleted rather than repaired into a fabricated exchange.

## Resolved surface finding — 2026-08-18

- `branches apps create` was planned successfully as `rplan_46dedc2add0a381b` with the contained `cert-classic` credential and a fixture-repository branch path. No provider write was dispatched.
- The stdin-only attempt was made twice and rejected before dispatch because connector commands use `--approve <token>` rather than the ETL transport's stdin-token path.
- Fleet ruling confirms that the locally minted, single-use, time-bounded approval token is not a GitHub credential and is authorized on argv. This is therefore a one-line surface finding, withdrawn as a product defect; mint a fresh plan/token rather than reuse either prior plan.
- Raw provider control: authenticated `GET /repos/Polymetrics-Cert/pm-cert-3993-20260810-wz0fru` returned HTTP 200.

## Refactor

- No product code, generated shared certification artifacts, command metadata, or CLI/help/docs surface is changed in this lane. The only committed changes are planning/verification records and valid proof-bearing evidence.

## Resumed 50-command batch — 2026-08-18

- Red: every new `pm-cert-` fixture began without the commanded field/state; delete cases began with an independently read-back object. Package, enterprise, self-fork, unattached-configuration, and self-reviewer cases were classified only after their parent collection or raw provider response proved why no usable object existed.
- Green: paths 57–62, 78, 80, 83, 85, 88–89, and 92–94 produced the exact expected value or absence. Plausible wrong values were the prior field value, empty array, opposite boolean, missing tagged name, or still-present object as appropriate.
- Raw controls: all eleven product defects have a direct `api.github.com` control independent of `pm`; raw controls returned 200/201/202/204 and their effects were read back before cleanup.
- Cleanup: custom patterns used bulk provider DELETE plus zero-by-ID collection reads; configurations and gists used direct DELETE plus 404; the shared private PR repository used direct DELETE 204 plus GET 404. No `pm` delete result was treated as cleanup proof.
