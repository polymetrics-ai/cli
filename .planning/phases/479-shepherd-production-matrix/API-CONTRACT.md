# API contract

Internal boundaries added by this correction are `ProductionPlanningIssueSource`,
`ProductionPlanAuthoritySource`, `ProductionPlanSession`, `ProductionPlanGitHubPort`, the
`ProductionPlanBootstrapper`, and an AgentSession-backed `ProductionVerificationPort`. All inputs are
closed, bounded records. Plan proposals omit numeric child issues and host authority. `host_verify`
accepts exactly `{id}` in planning order for the independent verification role and permits repeat runs
for implementation/correction RED→GREEN work; raw executable, argv, cwd, environment, and shell text are
not capability input fields. The final verification stage always reruns the complete immutable sequence.
