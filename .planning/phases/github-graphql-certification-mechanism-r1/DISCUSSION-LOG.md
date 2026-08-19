# GSD discuss-phase — GitHub GraphQL certification mechanism

Command resolved and executed inline: `scripts/gsd prompt discuss-phase github-graphql-certification-mechanism-r1`.

The non-interactive brief resolved the material decisions:

1. **Scope:** exactly one target connector, GitHub. The generic harness change is allowed because it is definition-driven and can apply to every connector; GitHub schema/query data stays in its bundle.
2. **Truthfulness:** schema compilation validates document shape only. A pass requires both a compiled schema and a connector-owned assertion evaluated by the stage. Provider values, entitlements, and mutation effects require a real run, and absent runs must remain non-pass.
3. **Bound:** no general GraphQL execution framework, query language, or connector branch. If definition metadata cannot express the next requirement, stop and record `needs-decision`.
4. **Proof:** write a red test that keeps schema compilation valid while breaking a certified value assertion; capture the red result before restoring the declaration and implementation.
5. **Safety:** use no ambient GitHub login; a credential is loaded only into an exported environment variable by command substitution, never emitted or stored. Live work is serial and resumes from definition-owned classification.
