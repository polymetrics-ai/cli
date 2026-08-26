# Issue #4354 source-projection/importer foundation — discussion log

## Fixed decisions

- Delivery is an ordinary direct PR to `main` that **references** #4354. It is
  an incremental shared foundation, not the Outreach connector-parity delivery
  described by #4354, so the PR must use `Refs #4354`, not `Closes #4354`.
- The retained provider bytes and their lock identity are immutable provenance.
  Their SHA-256 verifies the retained bytes; it is never a credential,
  executor, certification, or admission substitute.
- Outreach is the vertical proof. A second small schema-v3 case (or an
  existing retained schema-v3 lock) will prove the reader/projector is not
  Outreach-specific.
- Import must preserve every source operation and its citations. Unsupported
  operation shapes stay in all six capability lanes with a precise
  `missing_foundation` explanation; none can be promoted to `implemented` to
  satisfy a structural sweep.
- No provider network calls, credentials, source re-pins, generic HTTP/write,
  shell, or SQL execution are in scope. If no new command becomes executable,
  the usable-surface delta is explicitly zero.
- Byte-backed source importing remains a strict retained-artifact contract.
  The unavailable Outreach bytes do not prevent source-cited declaration
  mapping: a separate explicit source-reference/declaration path may preserve
  canonical provider URLs and operation identities, but must label unavailable
  contract detail `source_contract_unavailable` and cannot certify a command.

## Inline GSD fallback

The canonical delivery contract prohibits GSD-role spawning and this runner
does not provide the project Pi runtime. The worker therefore reviewed the
adapter-generated prompts and executes `discuss-phase` → `plan-phase --tdd` →
`execute-phase` → `verify-work` → `code-review` inline, recording equivalent
artifacts in this phase directory.
