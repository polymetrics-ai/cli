# Cycle 12 Pi Lifecycle Mapper

## Assignment

Read-only map of the explicitly installed `@earendil-works/pi-coding-agent` 0.80.6 contract for
issue #475, frozen start `7882cd70c25971e889ec04f63b98c936d605003e`.

Determine, without edits, network, credentials, services, or model calls:

1. exact no-tool and one-tool `AgentSessionEvent` order, including user/tool-result messages,
   intermediate assistants, tool execution, `agent_end.messages`, retries, and `agent_settled`;
2. the authoritative terminal/freeze boundary;
3. a supported public seam that lets the entire actual `createAgentSession` result and actual
   session prompt offline through a custom inert stream/provider;
4. exact factory-result/extension fields and installed assistant diagnostic shapes; and
5. the smallest relevant source/type locations.

Conclude once those facts are established. Return concise evidence and implementation guidance;
do not edit or expand into unrelated source.

## Authority

- Read-only filesystem/source inspection only.
- No writes, commits, tests that mutate state, network, credentials, live auth/model calls,
  services, GitHub, Go/connectors, or `make`.
- The issue worker owns every artifact, test, production, commit, and verification action.
