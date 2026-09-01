# Connector migration handoff

Use the current [connector canon](../connector-canon/INDEX.md). The sole
migration path is to create a reviewed schema-4 source lock, render the existing
execution JSON, prove the seven lanes, and keep runtime execution-only.

Do not recover an earlier generator, evidence ledger, or alternate reader from
history. Preserve provider facts needed by the new lock, and report an exact
shared Foundation Atlas gap before implementing a new runtime capability.
