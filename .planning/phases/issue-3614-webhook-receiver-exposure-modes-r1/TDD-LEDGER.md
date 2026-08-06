# TDD LEDGER — issue #3614 webhook receiver exposure modes

| ID | Contract | RED evidence | GREEN evidence |
| --- | --- | --- | --- |
| R1 | The three exposure modes are closed and change behavior | Pending: mode tests compile against missing receiver API | Pending |
| R2 | Tunnel mode binds loopback only and accepts only the named external Tailscale Funnel callback | Pending: loopback/address and named-tool tests | Pending |
| R3 | Raw request bytes are verified before parse or durable write; failures persist nothing | Pending: verifier/store sequencing tests | Pending |
| R4 | A receipt is durable before success; duplicate/out-of-order inputs are safe | Pending: injected durable-store transaction tests | Pending |
| R5 | Size, timeout, and in-flight limits reject excess work instead of buffering | Pending: HTTP handler boundary tests | Pending |
| R6 | URL rotation or missing heartbeat degrades and yields #3810 recovery work until explicit re-registration/reconciliation | Pending: lifecycle/state tests | Pending |
| R7 | CLI/help/docs expose the active mode without leaking callback URLs or secret material | Pending: CLI/manual/website parity tests | Pending |
