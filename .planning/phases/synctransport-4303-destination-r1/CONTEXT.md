# Refs #4303 — Discussion context

Inline/manual `discuss-phase` fallback: the issue and repository contracts resolve every material design choice. This runner cannot host compatible isolated GSD workers, and the canonical single-worker delivery contract forbids role spawning.

Locked decisions:

- Production destination composition is connector-neutral. Declarations select their own conformance evidence; no connector-name branch is permitted.
- A destination remains a typed adapter for a named connector-owned action. A generic HTTP, SQL, shell, or arbitrary-operation writer is prohibited.
- The new generic adapter must accept the action, source bindings, acknowledgement, and one closed apply strategy for each executable destination mode declared by the bundle.
- Invalid, unknown, and wrong-role declarations fail during composition/preflight, before I/O. `change_capture` remains source-only.
- GitHub’s issue-label route must run unchanged through the neutral adapter before its closed composition is deleted.
- Connector definitions are owned by other workers and are out of scope for this foundation PR.
