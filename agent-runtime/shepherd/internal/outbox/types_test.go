package outbox

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/polymetrics-ai/cli/agent-runtime/shepherd/internal/domain"
)

func TestStrictPayloadRoundTripAndRejection(t *testing.T) {
	t.Parallel()
	target := Target{Repository: "polymetrics-ai/cli", Issue: 389, PullRequest: 390}
	summary := "## Shepherd decisions\n\n| Decision |\n|---|\n| continue |"
	intent, err := NewSummaryIntent(target, "issue-389", 7, strings.Repeat("a", 40), 3,
		summaryHash(summary), summary)
	if err != nil {
		t.Fatal(err)
	}
	canonical := intent.Payload()
	decoded, err := DecodeIntent(KindDecisionSummary, canonical)
	if err != nil {
		t.Fatal(err)
	}
	if decoded.PayloadHash() != intent.PayloadHash() || string(decoded.Payload()) != string(canonical) {
		t.Fatalf("round trip changed immutable payload: before=%s after=%s", intent.PayloadHash(), decoded.PayloadHash())
	}

	invalid := []string{
		`{}`,
		`{"schema_version":1,"delivery_id":"issue-389","delivery_id":"issue-389"}`,
		`{"schema_version":1,"delivery_id":"issue-389","Delivery_ID":"issue-389"}`,
		`{"schema_version":1,"delivery_id":"issue-389","revision":1,"ledger_hash":"sha256:` + strings.Repeat("b", 64) + `","summary":"ok","unknown":true}`,
		`{"schema_version":1,"delivery_id":"issue-389","revision":1,"ledger_hash":"sha256:` + strings.Repeat("b", 64) + `","summary":"token=do-not-store"}`,
		`{"schema_version":1,"delivery_id":"issue-389","revision":1,"ledger_hash":"sha256:` + strings.Repeat("b", 64) + `","summary":"ok"} trailing`,
	}
	for _, raw := range invalid {
		if _, err := DecodeIntent(KindDecisionSummary, []byte(raw)); err == nil {
			t.Fatalf("unsafe payload accepted: %q", raw)
		}
	}
	if _, err := NewSummaryIntent(target, "issue-389", 7, strings.Repeat("a", 40), 0,
		summaryHash(summary), summary); err == nil {
		t.Fatal("revision-zero summary was accepted")
	}
	if _, err := DecodeIntent(Kind("github.pr.merge.v1"), canonical); err == nil {
		t.Fatal("unsupported merge effect kind accepted")
	}
	if _, err := NewSummaryIntent(target, "issue-389", 7, strings.Repeat("a", 40), 3,
		summaryHash("token=do-not-store"), "token=do-not-store"); err == nil {
		t.Fatal("secret-bearing summary was accepted")
	}
	if _, err := NewSummaryIntent(Target{Repository: "../escape", Issue: 389, PullRequest: 390},
		"issue-389", 7, strings.Repeat("a", 40), 3, summaryHash("summary"), "summary"); err == nil {
		t.Fatal("path-traversal repository target was accepted")
	}
}

func TestQuestionPayloadBindsRequestGenerationUnitHeadAndExpiry(t *testing.T) {
	t.Parallel()
	now := time.Unix(1_700_000_000, 0).UTC()
	target := Target{Repository: "polymetrics-ai/cli", Issue: 389, PullRequest: 390}
	question := Question{
		RequestID: "decision-1", DeliveryID: "issue-389", UnitID: "execute-task/M001/S01/T01",
		Generation: 7, HeadSHA: strings.Repeat("a", 40), Evidence: "retry budget exhausted",
		Options: []string{"retry", "stop"}, RecommendedOption: "retry", SafeDefault: "stop",
		ExpiresAt: now.Add(time.Hour), Mention: "karthik-sivadas",
	}
	intent, err := NewQuestionIntent(target, question)
	if err != nil {
		t.Fatal(err)
	}
	decoded, err := DecodeIntent(KindDecisionQuestion, intent.Payload())
	if err != nil || decoded.SourceID() != question.RequestID || decoded.Generation() != question.Generation ||
		decoded.HeadSHA() != question.HeadSHA || decoded.Target() != target {
		t.Fatalf("decoded question=%+v err=%v", decoded, err)
	}
	controller, _ := newTestController(t, ControllerFacts{
		DeliveryID: question.DeliveryID, Repository: target.Repository, Issue: target.Issue,
		PullRequest: target.PullRequest, Generation: question.Generation, HeadSHA: question.HeadSHA,
		Owner: "controller-1",
	}, now)
	if _, err := controller.Authorize(context.Background(), intent, question.ExpiresAt); err == nil {
		t.Fatal("expired question effect was authorized")
	}
	question.Options = []string{"retry", "RETRY"}
	if _, err := NewQuestionIntent(target, question); err == nil {
		t.Fatal("case-duplicate question options were accepted")
	}
	question.Options = []string{"retry", "stop"}
	question.Evidence = "authorization: secret"
	if _, err := NewQuestionIntent(target, question); err == nil {
		t.Fatal("secret-bearing question evidence was accepted")
	}
	question.Evidence = "retry budget exhausted"
	question.UnitID = "execute-task/M001; gh api"
	if _, err := NewQuestionIntent(target, question); err == nil {
		t.Fatal("command-shaped question unit identity was accepted")
	}
}

func TestControllerRevalidationRejectsGenerationRollover(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	now := time.Now().UTC()
	controller, _ := newTestController(t, ControllerFacts{
		DeliveryID: "issue-389", Repository: "polymetrics-ai/cli", Issue: 389, PullRequest: 390,
		Generation: 1, HeadSHA: strings.Repeat("a", 40), Owner: "controller", Epoch: 1,
	}, now)
	if _, err := controller.authority.BeginAttempt(ctx, "issue-389", "rollover-attempt"); err != nil {
		t.Fatal(err)
	}
	if err := controller.authority.FinishAttempt(ctx, "issue-389", "rollover-attempt", domain.RunFailed); err != nil {
		t.Fatal(err)
	}
	if err := controller.authority.ResumeDelivery(ctx, domain.HumanDecision{
		RunID: "issue-389", Generation: 1, ActorKind: domain.ActorHuman, Approved: true,
	}); err != nil {
		t.Fatal(err)
	}
	if err := controller.Revalidate(ctx, now.Add(time.Second)); err == nil {
		t.Fatal("old generation controller survived durable generation rollover")
	}
}

func TestControllerAuthorizationIsNarrowImmutableAndFenced(t *testing.T) {
	t.Parallel()
	facts := ControllerFacts{
		DeliveryID: "issue-389", Repository: "polymetrics-ai/cli", Issue: 389, PullRequest: 390,
		Generation: 7, HeadSHA: strings.Repeat("a", 40), Owner: "controller-1",
	}
	controller, facts := newTestController(t, facts, time.Unix(1_700_000_000, 0).UTC())
	intent, err := NewSummaryIntent(Target{Repository: facts.Repository, Issue: facts.Issue, PullRequest: facts.PullRequest},
		facts.DeliveryID, facts.Generation, facts.HeadSHA, 4, summaryHash("summary"), "summary")
	if err != nil {
		t.Fatal(err)
	}
	authorized, err := controller.Authorize(context.Background(), intent, time.Unix(1_700_000_000, 0).UTC())
	if err != nil {
		t.Fatal(err)
	}
	if authorized.Capability() != CapabilityGitHubCommentWrite || authorized.Epoch() != facts.Epoch ||
		authorized.Owner() != facts.Owner || authorized.GrantID() == "" || authorized.EffectID() == "" {
		t.Fatalf("incomplete authorization: capability=%q epoch=%d owner=%q grant=%q effect=%q",
			authorized.Capability(), authorized.Epoch(), authorized.Owner(), authorized.GrantID(), authorized.EffectID())
	}

	mismatches := []ControllerFacts{
		{DeliveryID: "issue-390", Repository: facts.Repository, Issue: facts.Issue, PullRequest: facts.PullRequest, Generation: facts.Generation, HeadSHA: facts.HeadSHA, Owner: facts.Owner, Epoch: facts.Epoch},
		{DeliveryID: facts.DeliveryID, Repository: facts.Repository, Issue: facts.Issue, PullRequest: 391, Generation: facts.Generation, HeadSHA: facts.HeadSHA, Owner: facts.Owner, Epoch: facts.Epoch},
		{DeliveryID: facts.DeliveryID, Repository: facts.Repository, Issue: facts.Issue, PullRequest: facts.PullRequest, Generation: facts.Generation + 1, HeadSHA: facts.HeadSHA, Owner: facts.Owner, Epoch: facts.Epoch},
		{DeliveryID: facts.DeliveryID, Repository: facts.Repository, Issue: facts.Issue, PullRequest: facts.PullRequest, Generation: facts.Generation, HeadSHA: strings.Repeat("c", 40), Owner: facts.Owner, Epoch: facts.Epoch},
	}
	for _, mismatch := range mismatches {
		other, _ := newTestController(t, mismatch, time.Unix(1_700_000_000, 0).UTC())
		if _, err := other.Authorize(context.Background(), intent, time.Unix(1_700_000_000, 0).UTC()); err == nil {
			t.Fatalf("mismatched controller facts authorized effect: %+v", mismatch)
		}
	}

	for _, forbidden := range []Capability{"forbidden_main_merge", "merge.main", "pr.merge"} {
		if IsGrantableCapability(forbidden) {
			t.Fatalf("forbidden capability %q became grantable", forbidden)
		}
	}
}
