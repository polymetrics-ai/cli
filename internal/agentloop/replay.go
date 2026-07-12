package agentloop

// ReplayResult is the deterministic machine-readable oracle output.
type ReplayResult struct {
	SchemaVersion           string   `json:"schema_version"`
	IncidentID              string   `json:"incident_id"`
	ViolationCode           string   `json:"violation_code"`
	ReasonCodes             []string `json:"reason_codes"`
	ObservedDecision        string   `json:"observed_decision"`
	ObservedOutcome         string   `json:"observed_outcome"`
	ObservedDecisionCorrect bool     `json:"observed_decision_correct"`
	ObservedOutcomeCorrect  bool     `json:"observed_outcome_correct"`
	RequiredDecision        string   `json:"required_decision"`
	RequiredOutcome         string   `json:"required_outcome"`
	RequiredExitClass       string   `json:"required_exit_class"`
	MatchedExpectation      bool     `json:"matched_expectation"`
}

type policyResult struct {
	violationCode     string
	reasonCodes       []string
	requiredDecision  string
	requiredOutcome   string
	requiredExitClass string
}

type factMatch struct {
	kind  string
	after string
}

// Replay derives policy from fact relationships and ordering, then verifies
// that the declared expectation agrees with that independent derivation.
func Replay(fixture Fixture) (ReplayResult, error) {
	if err := ValidateFixture(fixture); err != nil {
		return ReplayResult{}, err
	}
	policies := derivePolicies(fixture.Events)
	if len(policies) == 0 {
		return ReplayResult{}, validationError("FIXTURE_INCIDENT_UNCLASSIFIED", "fixture facts do not match a known incident")
	}
	if len(policies) > 1 {
		return ReplayResult{}, validationError("FIXTURE_INCIDENT_AMBIGUOUS", "fixture facts match more than one incident")
	}
	policy := policies[0]
	canonicalViolation, _ := canonicalViolationForIncident(fixture.IncidentID)
	if canonicalViolation != policy.violationCode {
		return ReplayResult{}, validationError("FIXTURE_INCIDENT_ID_MISMATCH", "fixture incident id does not match derived policy")
	}
	if fixture.Expected.ViolationCode != policy.violationCode ||
		fixture.Expected.RequiredDecision != policy.requiredDecision ||
		fixture.Expected.RequiredOutcome != policy.requiredOutcome ||
		fixture.Expected.RequiredExitClass != policy.requiredExitClass {
		return ReplayResult{}, validationError("FIXTURE_EXPECTATION_MISMATCH", "fixture expectation differs from derived policy")
	}
	return ReplayResult{
		SchemaVersion:           FixtureSchemaVersion,
		IncidentID:              fixture.IncidentID,
		ViolationCode:           policy.violationCode,
		ReasonCodes:             append([]string(nil), policy.reasonCodes...),
		ObservedDecision:        fixture.Expected.ObservedDecision,
		ObservedOutcome:         fixture.Expected.ObservedOutcome,
		ObservedDecisionCorrect: *fixture.Expected.ObservedDecisionCorrect,
		ObservedOutcomeCorrect:  *fixture.Expected.ObservedOutcomeCorrect,
		RequiredDecision:        policy.requiredDecision,
		RequiredOutcome:         policy.requiredOutcome,
		RequiredExitClass:       policy.requiredExitClass,
		MatchedExpectation:      true,
	}, nil
}

// ReplayAll preserves caller order and stops at the first invalid incident.
func ReplayAll(fixtures []Fixture) ([]ReplayResult, error) {
	results := make([]ReplayResult, 0, len(fixtures))
	for _, fixture := range fixtures {
		result, err := Replay(fixture)
		if err != nil {
			return nil, err
		}
		results = append(results, result)
	}
	return results, nil
}

func derivePolicies(events []Event) []policyResult {
	policies := make([]policyResult, 0, 1)
	if phantom, missing := workerCompletionFacts(events); phantom && missing {
		policies = append(policies, retryPolicy("WORKER_COMPLETION_UNPROVEN", "PHANTOM_DISPATCH", "HANDOFF_MISSING"))
	}
	if falseGreenFacts(events) {
		policies = append(policies, retryPolicy("VALIDATION_FALSE_GREEN", "REQUIRED_ARTIFACT_MISSING", "REPO_GATE_FAILED_AFTER_PROCEED"))
	}
	if unauthorizedOwnerTransition(events, "human_gate", "cleared") {
		policies = append(policies, haltPolicy("AUTHORITY_FABRICATED", "AGENT_CROSSED_HUMAN_OWNER"))
	}
	if repeatedHaltFacts(events) {
		policies = append(policies, policyResult{
			violationCode:     "HALT_REVOCATION_MISSING",
			reasonCodes:       []string{"CHILD_ACTIVE_AFTER_HALT", "UNAUTHORIZED_RESUME", "HALT_REPEATED"},
			requiredDecision:  "halt",
			requiredOutcome:   "halt_persisted_children_revoked",
			requiredExitClass: "halt_required",
		})
	}
	if megaTurnFacts(events) {
		policies = append(policies, haltPolicy("TURN_SUPERVISION_EXCEEDED", "TURN_BUDGET_EXCEEDED", "SUPERVISION_DETACHED"))
	}
	if dualWriterFacts(events) {
		policies = append(policies, haltPolicy("WORKTREE_DUAL_WRITER", "OVERLAPPING_WRITER_OWNERS"))
	}
	if mergeBeforeRatificationFacts(events) {
		policies = append(policies, haltPolicy("MERGE_BEFORE_RATIFICATION", "RATIFICATION_RECORDED_AFTER_MERGE"))
	}
	if mergeStateChangedAfterAttestationFacts(events) {
		policies = append(policies, haltPolicy("MERGE_ATTESTATION_STALE", "MERGE_STATE_CHANGED_DURING_VALIDATION"))
	}
	if mergeAuthorityFacts(events) {
		policies = append(policies, haltPolicy("MERGE_AUTHORITY_DENIED", "AGENT_CROSSED_MERGE_OWNER"))
	}
	if staleBindingFacts(events, "verification_binding", "validator_decision", "retry") {
		policies = append(policies, retryPolicy("VERIFY_HEAD_STALE", "HEAD_MOVED_AFTER_VERIFICATION"))
	}
	if dirtyWorktreeFacts(events) {
		policies = append(policies, retryPolicy("WORKTREE_DIRTY", "DIRTY_STATE_REQUIRES_RECONCILE"))
	}
	if humanWaitProjectionFacts(events) {
		policies = append(policies, policyResult{
			violationCode:     "HUMAN_WAIT_PROJECTED_FINAL",
			reasonCodes:       []string{"INTERIM_WAIT_PROJECTED_DONE"},
			requiredDecision:  "wait",
			requiredOutcome:   "human_wait_preserved",
			requiredExitClass: "human_wait_required",
		})
	}
	if terminalMismatchFacts(events) {
		policies = append(policies, haltPolicy("TERMINAL_PROJECTION_MISMATCH", "DURABLE_AND_PROJECTED_TERMINALS_DIFFER"))
	}
	return policies
}

func retryPolicy(code string, reasons ...string) policyResult {
	return policyResult{
		violationCode:     code,
		reasonCodes:       reasons,
		requiredDecision:  "retry",
		requiredOutcome:   "correction_preserved_once",
		requiredExitClass: "retry_required",
	}
}

func haltPolicy(code string, reasons ...string) policyResult {
	return policyResult{
		violationCode:     code,
		reasonCodes:       reasons,
		requiredDecision:  "halt",
		requiredOutcome:   "halt_persisted",
		requiredExitClass: "halt_required",
	}
}

func workerCompletionFacts(events []Event) (bool, bool) {
	phantom := false
	missingHandoff := false
	for dispatchIndex, dispatch := range events {
		if dispatch.Fact.Kind != "worker_dispatch" || dispatch.Fact.After != "requested" {
			continue
		}
		for executionIndex := dispatchIndex + 1; executionIndex < len(events); executionIndex++ {
			execution := events[executionIndex]
			if execution.Fact.Kind != "worker_execution" ||
				execution.Fact.ResourceID != dispatch.Fact.ResourceID ||
				!sameExactSnapshot(dispatch, execution) {
				continue
			}
			if execution.Fact.Before == "not_started" && execution.Fact.After == "not_started" {
				phantom = true
			}
			if execution.Fact.After != "completed" {
				continue
			}
			for handoffIndex := executionIndex + 1; handoffIndex < len(events); handoffIndex++ {
				handoff := events[handoffIndex]
				if handoff.Fact.Kind == "worker_handoff" && handoff.Fact.After == "missing" &&
					handoff.Fact.ResourceID == execution.Fact.ResourceID && sameExactSnapshot(execution, handoff) {
					missingHandoff = true
				}
			}
		}
	}
	return phantom, missingHandoff
}

func falseGreenFacts(events []Event) bool {
	return hasOrderedCorrelatedFacts(events, sameStageScope,
		factMatch{kind: "required_artifact", after: "missing"},
		factMatch{kind: "validator_decision", after: "proceed"},
		factMatch{kind: "repo_gate", after: "failed"},
	)
}

func unauthorizedOwnerTransition(events []Event, kind, after string) bool {
	for _, event := range events {
		if event.Fact.Kind == kind && event.Fact.After == after && event.ActorID != event.Fact.OwnerID {
			return true
		}
	}
	return false
}

func repeatedHaltFacts(events []Event) bool {
	for firstIndex, first := range events {
		if first.Fact.Kind != "halt_latch" || first.Fact.After != "halted" {
			continue
		}
		for secondIndex := firstIndex + 1; secondIndex < len(events); secondIndex++ {
			second := events[secondIndex]
			if second.Fact.Kind != "halt_latch" || second.Fact.After != "halted" ||
				second.Binding.TurnID == first.Binding.TurnID {
				continue
			}
			workerFound := false
			resumeFound := false
			for between := firstIndex + 1; between < secondIndex; between++ {
				candidate := events[between]
				if candidate.Fact.Kind == "worker_liveness" && candidate.Fact.After == "active" &&
					sameExactSnapshot(first, candidate) {
					workerFound = true
				}
				if candidate.Fact.Kind == "resume_authorization" && candidate.Fact.After == "requested" &&
					candidate.ActorID != candidate.Fact.OwnerID && sameExactSnapshot(candidate, second) {
					resumeFound = true
				}
			}
			if workerFound && resumeFound {
				return true
			}
		}
	}
	return false
}

func megaTurnFacts(events []Event) bool {
	return hasOrderedCorrelatedFacts(events, sameExactSnapshot,
		factMatch{kind: "turn_budget", after: "exceeded"},
		factMatch{kind: "turn_supervision", after: "detached"},
		factMatch{kind: "stage_decision", after: "proceed"},
	)
}

func dualWriterFacts(events []Event) bool {
	for firstIndex, first := range events {
		if first.Fact.Kind != "writer_lease" || first.Fact.After != "active" || first.ActorID != first.Fact.OwnerID {
			continue
		}
		for secondIndex := firstIndex + 1; secondIndex < len(events); secondIndex++ {
			second := events[secondIndex]
			if second.Fact.Kind != "writer_lease" || second.Fact.After != "active" ||
				second.ActorID != second.Fact.OwnerID || second.Fact.ResourceID != first.Fact.ResourceID ||
				second.Fact.OwnerID == first.Fact.OwnerID || second.ActorID == first.ActorID ||
				!sameWorkScope(first, second) {
				continue
			}
			for mutationIndex := secondIndex + 1; mutationIndex < len(events); mutationIndex++ {
				mutation := events[mutationIndex]
				if mutation.Fact.Kind == "worktree_state" && mutation.Fact.After == "mutated" &&
					mutation.Fact.ResourceID == second.Fact.ResourceID && sameExactSnapshot(second, mutation) {
					return true
				}
			}
		}
	}
	return false
}

func mergeBeforeRatificationFacts(events []Event) bool {
	return hasOrderedCorrelatedFacts(events, func(selected ...Event) bool {
		return sameStageScope(selected...) && selected[0].Fact.ResourceID == selected[1].Fact.ResourceID
	},
		factMatch{kind: "merge_state", after: "merged"},
		factMatch{kind: "ratification_state", after: "recorded"},
	)
}

func mergeStateChangedAfterAttestationFacts(events []Event) bool {
	for attestationIndex, attestation := range events {
		if attestation.Fact.Kind != "attestation_binding" || attestation.Fact.After != "open" {
			continue
		}
		for mergeIndex := attestationIndex + 1; mergeIndex < len(events); mergeIndex++ {
			merge := events[mergeIndex]
			if merge.Fact.Kind != "merge_state" || merge.Fact.ResourceID != attestation.Fact.ResourceID ||
				merge.Fact.Before != attestation.Fact.After || merge.Fact.After != "merged" {
				continue
			}
			for verdictIndex := mergeIndex + 1; verdictIndex < len(events); verdictIndex++ {
				verdict := events[verdictIndex]
				if verdict.Fact.Kind == "validator_decision" && verdict.Fact.After == "proceed" &&
					sameStageScope(attestation, merge, verdict) {
					return true
				}
			}
		}
	}
	return false
}

func staleBindingFacts(events []Event, bindingKind, terminalKind, terminalAfter string) bool {
	for bindingIndex, binding := range events {
		if binding.Fact.Kind != bindingKind || binding.Fact.After == "none" {
			continue
		}
		for headIndex := bindingIndex + 1; headIndex < len(events); headIndex++ {
			head := events[headIndex]
			if head.Fact.Kind != "head_state" || head.Fact.ResourceID != binding.Fact.ResourceID ||
				head.Fact.Before != binding.Fact.After ||
				head.Fact.After == binding.Fact.After {
				continue
			}
			for terminalIndex := headIndex + 1; terminalIndex < len(events); terminalIndex++ {
				terminal := events[terminalIndex]
				if terminal.Fact.Kind == terminalKind && terminal.Fact.After == terminalAfter &&
					sameStageScope(binding, head, terminal) {
					return true
				}
			}
		}
	}
	return false
}

func mergeAuthorityFacts(events []Event) bool {
	for authorityIndex, authority := range events {
		if authority.Fact.Kind != "merge_authority" || authority.Fact.After != "agent_requested" ||
			authority.ActorID == authority.Fact.OwnerID {
			continue
		}
		for mergeIndex := authorityIndex + 1; mergeIndex < len(events); mergeIndex++ {
			merge := events[mergeIndex]
			if merge.Fact.Kind == "merge_state" && merge.Fact.ResourceID == authority.Fact.ResourceID &&
				merge.Fact.After == "merged" && merge.ActorID != merge.Fact.OwnerID && sameStageScope(authority, merge) {
				return true
			}
		}
	}
	return false
}

func dirtyWorktreeFacts(events []Event) bool {
	return hasOrderedCorrelatedFacts(events, sameExactSnapshot,
		factMatch{kind: "worktree_state", after: "dirty"},
		factMatch{kind: "validator_decision", after: "retry"},
	)
}

func humanWaitProjectionFacts(events []Event) bool {
	return hasOrderedCorrelatedFacts(events, sameStageScope,
		factMatch{kind: "stage_state", after: "blocked"},
		factMatch{kind: "human_wait", after: "active"},
		factMatch{kind: "terminal_state", after: "human_gate"},
		factMatch{kind: "terminal_projection", after: "done"},
	)
}

func terminalMismatchFacts(events []Event) bool {
	return hasOrderedCorrelatedFacts(events, func(selected ...Event) bool {
		if !sameStageScope(selected...) {
			return false
		}
		for _, event := range events {
			if event.Fact.Kind == "human_wait" && event.Fact.After == "active" &&
				sameStageScope(append(selected, event)...) {
				return false
			}
		}
		return true
	},
		factMatch{kind: "stage_state", after: "running"},
		factMatch{kind: "terminal_state", after: "none"},
		factMatch{kind: "terminal_projection", after: "human_gate"},
		factMatch{kind: "validator_decision", after: "proceed"},
	)
}

func hasOrderedCorrelatedFacts(events []Event, correlated func(...Event) bool, matches ...factMatch) bool {
	selected := make([]Event, 0, len(matches))
	var search func(matchIndex, eventIndex int) bool
	search = func(matchIndex, eventIndex int) bool {
		if matchIndex == len(matches) {
			return correlated(selected...)
		}
		match := matches[matchIndex]
		for i := eventIndex; i < len(events); i++ {
			event := events[i]
			if event.Fact.Kind != match.kind || event.Fact.After != match.after {
				continue
			}
			selected = append(selected, event)
			if search(matchIndex+1, i+1) {
				return true
			}
			selected = selected[:len(selected)-1]
		}
		return false
	}
	return search(0, 0)
}

func sameExactSnapshot(events ...Event) bool {
	if len(events) < 2 {
		return true
	}
	binding := events[0].Binding
	for _, event := range events[1:] {
		if event.Binding != binding {
			return false
		}
	}
	return true
}

func sameStageScope(events ...Event) bool {
	if len(events) < 2 {
		return true
	}
	binding := events[0].Binding
	for _, event := range events[1:] {
		other := event.Binding
		if binding.RunID != other.RunID || binding.GenerationID != other.GenerationID ||
			binding.ControllerID != other.ControllerID || binding.StageID != other.StageID ||
			binding.TicketID != other.TicketID || binding.HeadSHA != other.HeadSHA {
			return false
		}
	}
	return true
}

func sameWorkScope(events ...Event) bool {
	if len(events) < 2 {
		return true
	}
	binding := events[0].Binding
	for _, event := range events[1:] {
		other := event.Binding
		if binding.RunID != other.RunID || binding.StageID != other.StageID ||
			binding.TicketID != other.TicketID || binding.HeadSHA != other.HeadSHA {
			return false
		}
	}
	return true
}
