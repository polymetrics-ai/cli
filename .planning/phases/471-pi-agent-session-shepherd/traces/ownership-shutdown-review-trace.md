# Ownership And Shutdown Review Trace

The independent concurrency audit blocked checkpoint `c2c4447c` with executable reproductions:

- two controllers sharing one FileStateStore both completed and dispatched four lanes;
- stop during target capture returned no-state and later allowed both lanes;
- abort during delayed AgentSession creation allowed a later prompt;
- hung abort/wait-for-idle prevented close and dispose;
- graceful shutdown could be recorded as halted rather than interrupted.

Each reproduction became RED before correction. The corrected design has a single global
mode-0600 O_EXCL lease with strict PID/token/inode fencing, explicit-resume-only dead-owner
recovery, controller lifecycle ownership before its first await, unowned-stop rejection,
pre-registration cancellation tombstones, AbortSignal prompt cancellation, bounded evidence
commands/cleanup/extension shutdown, and unconditional dispose. The expanded focused suite is
49/49 green and strict TypeScript passes. A final review must re-run against the correction commit.
