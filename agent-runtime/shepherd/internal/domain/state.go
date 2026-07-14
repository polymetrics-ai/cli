package domain

import (
	"errors"
	"fmt"
)

type RunState string

const (
	RunPlanned          RunState = "planned"
	RunReady            RunState = "ready"
	RunRunning          RunState = "running"
	RunBlocked          RunState = "blocked"
	RunAwaitingDecision RunState = "awaiting_decision"
	RunFailed           RunState = "failed"
	RunHumanGate        RunState = "human_gate"
	RunComplete         RunState = "complete"
)

var allowedRunTransitions = map[RunState]map[RunState]struct{}{
	RunPlanned:          {RunRunning: {}, RunBlocked: {}, RunFailed: {}},
	RunReady:            {RunRunning: {}, RunFailed: {}},
	RunRunning:          {RunReady: {}, RunBlocked: {}, RunAwaitingDecision: {}, RunFailed: {}, RunHumanGate: {}},
	RunBlocked:          {RunFailed: {}},
	RunAwaitingDecision: {RunFailed: {}},
	RunHumanGate:        {RunFailed: {}},
	RunFailed:           {},
	RunComplete:         {},
}

type ActorKind string

const (
	ActorHuman ActorKind = "human"
	ActorAgent ActorKind = "agent"
)

type HumanDecision struct {
	RunID      string
	Generation int64
	ActorKind  ActorKind
	Approved   bool
	Consumed   bool
}

func ResumeBlocked(runID string, generation int64, decision HumanDecision) (int64, error) {
	return ResumeStopped(runID, generation, RunBlocked, decision)
}

// ResumeStopped validates the exceptional human-authorized recovery path. A
// generic failed transition remains forbidden; only a generation-bound,
// unconsumed human decision can return a blocked, awaiting-decision, or failed
// delivery to ready.
func ResumeStopped(runID string, generation int64, state RunState, decision HumanDecision) (int64, error) {
	if state != RunBlocked && state != RunAwaitingDecision && state != RunFailed {
		return 0, errors.New("only a blocked, awaiting-decision, or failed run can be resumed")
	}
	if runID == "" || generation <= 0 || decision.RunID != runID || decision.Generation != generation {
		return 0, errors.New("human decision does not match run generation")
	}
	if decision.ActorKind != ActorHuman || !decision.Approved || decision.Consumed {
		return 0, errors.New("an unconsumed explicit human approval is required")
	}
	return generation + 1, nil
}

func ValidateRunTransition(from, to RunState) error {
	allowed, exists := allowedRunTransitions[from]
	if !exists {
		return fmt.Errorf("unknown run state %q", from)
	}
	if _, ok := allowed[to]; !ok {
		return fmt.Errorf("run transition %s -> %s is not permitted", from, to)
	}
	return nil
}

type Capability string

const (
	CapabilityIssueCreate Capability = "issue.create"
	CapabilityIssueUpdate Capability = "issue.update"
	CapabilityIssueClose  Capability = "issue.close"
	CapabilityPRCreate    Capability = "pr.create"
	CapabilityPRUpdate    Capability = "pr.update"
	CapabilityPRClose     Capability = "pr.close"
	CapabilityGitPush     Capability = "git.push"
)

var allowedCapabilities = map[Capability]struct{}{
	CapabilityIssueCreate: {},
	CapabilityIssueUpdate: {},
	CapabilityIssueClose:  {},
	CapabilityPRCreate:    {},
	CapabilityPRUpdate:    {},
	CapabilityPRClose:     {},
	CapabilityGitPush:     {},
}

func IsGrantableCapability(capability Capability) bool {
	_, ok := allowedCapabilities[capability]
	return ok
}

type Grant struct {
	RunID      string
	Repository string
	Issue      int
	Capability Capability
	Epoch      int64
}

func NewGrant(runID, repository string, issue int, capability Capability, epoch int64) (Grant, error) {
	if runID == "" || repository == "" || issue <= 0 || epoch <= 0 {
		return Grant{}, errors.New("grant identity, issue, and epoch are required")
	}
	if _, ok := allowedCapabilities[capability]; !ok {
		return Grant{}, fmt.Errorf("capability %q is not grantable", capability)
	}
	return Grant{RunID: runID, Repository: repository, Issue: issue, Capability: capability, Epoch: epoch}, nil
}
