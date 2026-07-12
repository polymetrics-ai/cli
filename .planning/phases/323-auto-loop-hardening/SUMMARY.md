# Summary: Autonomous Delivery Control-Plane Hardening

Status: parent PR open; existing Shepherd implementation aligned; dependent hardening remains.

Issue #323 is the remediation roadmap for the Twenty autonomous-delivery post-mortem. PR #324 now
targets `feat/pi-shepherd-loop`, and the hardening branch normally merges that existing implementation
instead of recreating it. Fifteen native child issues carry dependency-ordered test-first contracts;
their designs must integrate through the existing launcher and prompts rather than become a parallel
delivery system.

Phase 0 now guards all three existing launchers before side effects. The Shepherd validator alone
defaults to GPT-5.6 Sol with high reasoning and fails before mutable work when its runtime cannot
provide that model; orchestrator and worker models are unchanged. No parent-to-main merge is
authorized, and the migration fuse remains closed while dependent controls are incomplete.
