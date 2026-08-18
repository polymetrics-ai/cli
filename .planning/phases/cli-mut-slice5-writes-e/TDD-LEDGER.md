# TDD Ledger — GitHub mutation certification slice 5 writes-e

## Manual GSD fallback

The repository-local Pi adapter prompts were generated and reviewed. This isolated terminal runtime cannot run Pi subagents, so the lifecycle is completed inline.

## Cycle 1 — certification evidence representability

**Red:** Inspect the validator before recording a mutation. The current `acceptedEvidence` schema only accepts connector capability/workflow/sync-mode/flow scopes and has no command path, mutation classification, output assertion, or cleanup proof field.

**Expected failure:** A schema-v2 command-level certification receipt cannot be represented or accepted without production-schema work; writing a capability record for an unexecuted command would be fabricated evidence.

**Green criterion:** Do not persist a record until an existing supported schema and a real provider proof are both available. Record the finding and classify the command truthfully.

## Cycle 2 — live mutation containment

**Red criterion:** A successful CLI exit alone is insufficient because issue #4221 demonstrates delete success can leave the object present.

**Green criterion:** A passed command must include an independent provider read-back of its state change and a direct-provider deletion followed by a 404 or empty collection.
