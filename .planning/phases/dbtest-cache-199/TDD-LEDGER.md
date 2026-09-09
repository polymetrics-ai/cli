# Cache policy199 TDD

Red: pending behavioral harness command/side-effect controls. Green: pending.

Initial dbtest-cache-red-199-01 failed to compile because new test omitted existing Close context argument. Zero behavioral events; not Red. Corrected test invocation before actual behavioral RED, helper behavior unchanged.

Behavioral Red: dbtest-cache-red-199-02 executed11 events/1pass with actual unwanted pull, missing/invalid cache admitted, absent policy/image evidence and unknown policy accepted. Final pre-edit counterfactual dbtest-cache-red-199-03 additionally requires explicit --pull=never;11/1, raw3a8e5bbcd6cf01c788d9cca0fd16ac02a415ed479e6e9770b63a7bdac561aeec. Representation-only Config/Report fields preceded RED; no cache behavior existed then.

Green: full dbtest-cache-green-199-01 passed93/93. Discovered competing ContainerArgs --pull override: real New refusal RED dbtest-cache-override-red-199-01 failed1/0, raw7e95e9c855520705274582ce08a2ffbcc8e01798be19d6b8b408d2aebaa9bc9c. Narrow config refusal preserves default mode. Final full normal dbtest-cache-green-199-02 passed94/94 raw0b05e889d6a61b456635479d50e5ec5df37ec40d47efc9ebba127f951b73ff99; full race dbtest-cache-race-199-01 passed94/94 rawf8c2297eebeb0feac19be1f9089718a31c5d502650b4200151e5e5373288be00. Counts overlap and are not summed. All captures have no input drift.

Affected physical witness shared-checkpoint-physical-199-01 now running fresh count1 with explicit local Docker endpoint and cache-only. Original194 setup failure preserved. No terminal physical claim until receipt and cleanup.

Physical Green: shared-checkpoint-physical-199-01 exit0,1/1,187.860496625s, raw506e68c3302bfca29a7b428389d9a96ec35e02d6cb6fb3e5e627a72bee9593bb. Existing actual rows/WAL/ack/lease/resume plus corrected monotonic checkpoint assertions pass; owned container listing empty. Scoped vet/lint pass. No acceptance or hosted-green claim.
