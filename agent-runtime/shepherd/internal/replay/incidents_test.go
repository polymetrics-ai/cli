package replay

import (
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/polymetrics-ai/cli/agent-runtime/shepherd/internal/authority"
	"github.com/polymetrics-ai/cli/agent-runtime/shepherd/internal/contract"
	"github.com/polymetrics-ai/cli/agent-runtime/shepherd/internal/domain"
	"github.com/polymetrics-ai/cli/agent-runtime/shepherd/internal/gsd"
	"github.com/polymetrics-ai/cli/agent-runtime/shepherd/internal/outbox"
)

func TestTwentyIncidentGuards(t *testing.T) {
	t.Parallel()

	now := time.Unix(1_700_000_000, 0).UTC()
	validRatification := authority.RatificationRequest{
		Repository: "polymetrics-ai/cli", PR: 285, BaseBranch: "main",
		BaseSHA:       "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
		CandidateHead: "bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb",
		ObservedHead:  "bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb",
		RunID:         "twenty", Generation: 4, UnitID: "S7/validate", Attempt: 1, StateVersion: 48,
		ContractHash: "sha256:" + strings.Repeat("c", 64), EvidenceHash: "sha256:" + strings.Repeat("e", 64),
		Validator: authority.RequiredValidator, Thinking: "high", ValidatorSessionID: "11111111-1111-1111-1111-111111111111", Verdict: "PROCEED",
		LocalGates: true, UAT: true, MilestoneValid: true,
		RequiredLocalGates: true, RequiredUAT: true, RequiredMilestoneValid: true,
		IssuedAt: now.Add(-time.Minute), ExpiresAt: now.Add(time.Minute),
	}

	tests := []struct {
		name  string
		guard func() error
	}{
		{"dead-worker", func() error { _, err := gsd.DecodeQuery([]byte(`{}`)); return err }},
		{"false-green-validation", func() error {
			request := validRatification
			request.LocalGates = false
			_, err := authority.Ratify(request, now)
			return err
		}},
		{"loader-scope-gap", func() error {
			return contract.Dispatch{Objective: "load", OutputFormat: "handoff", ToolGuidance: []string{"edit"}, Tools: []contract.Tool{contract.ToolEdit}, Boundaries: []string{"scope"}, WriteScope: []string{"../docs"}}.Validate()
		}},
		{"hard-gate-breach", func() error {
			return rejectExternalCapability(outbox.Capability("merge.main"))
		}},
		{"scope-reduction-halt", func() error {
			return contract.Dispatch{Objective: "reduce", OutputFormat: "handoff", ToolGuidance: []string{"edit"}, Tools: []contract.Tool{contract.ToolEdit}, Boundaries: nil, WriteScope: []string{"internal/**"}}.Validate()
		}},
		{"unsupervised-mega-turn", func() error {
			_, err := gsd.NewRunner(gsd.Config{Command: []string{"gsd"}, WorkDir: "/tmp", GSDHome: "/tmp", Model: "openai-codex/gpt-5.6-sol", Thinking: "high", HeartbeatInterval: 16 * time.Second})
			return err
		}},
		{"engine-scope-drift", func() error {
			return contract.Dispatch{Objective: "engine", OutputFormat: "handoff", ToolGuidance: []string{"edit"}, Tools: []contract.Tool{contract.ToolEdit}, Boundaries: []string{"scope"}, WriteScope: []string{"/internal/connectors/engine"}}.Validate()
		}},
		{"merge-before-ratification", func() error {
			return rejectExternalCapability(outbox.Capability("pr.merge"))
		}},
		{"stale-verify-head", func() error {
			request := validRatification
			request.ObservedHead = "cccccccccccccccccccccccccccccccccccccccc"
			_, err := authority.Ratify(request, now)
			return err
		}},
		{"restart-without-human-decision", func() error {
			_, err := domain.ResumeBlocked("twenty", 4, domain.HumanDecision{RunID: "twenty", Generation: 4, ActorKind: domain.ActorAgent, Approved: true})
			return err
		}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if err := test.guard(); err == nil {
				t.Fatalf("incident guard did not block %s", test.name)
			}
		})
	}
}

func rejectExternalCapability(capability outbox.Capability) error {
	if outbox.IsGrantableCapability(capability) {
		return nil
	}
	return errors.New("external capability is not grantable")
}
