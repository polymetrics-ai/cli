# Summary — CLI PM Broker profile/context foundation

Status: planning checkpoint created; implementation pending.

This phase targets CLI #566 on the `integration/pm-broker-production-program` parent branch. It will add safe PM Broker Organization/Workspace/Environment/BrokerProfile context foundations, runtime-mode validation, and contract-version refusal seams without enabling live broker/provider operations or touching credentials.

Manual-GSD fallback is active because `scripts/gsd prompt programming-loop ...` is not exposed by the current repo adapter, though `scripts/gsd doctor`, `scripts/gsd list`, and `scripts/gsd prompt plan-phase ...` work.
