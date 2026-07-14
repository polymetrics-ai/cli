package github

import (
	"context"
	"encoding/json"
	"strings"
	"testing"
	"time"
)

func TestQuestionCommentSyncCreatesMarkerOwnedMentionedQuestion(t *testing.T) {
	t.Parallel()
	runner := &fakeRunner{responses: [][]byte{[]byte(`[[{"id":1,"body":"unrelated"}]]`), []byte(`{"id":77}`)}}
	client := NewClient(runner)
	request := testQuestionRequest()
	id, err := client.SyncQuestionComment(context.Background(), request)
	if err != nil {
		t.Fatal(err)
	}
	if id != 77 {
		t.Fatalf("comment id=%d", id)
	}
	if len(runner.calls) != 2 {
		t.Fatalf("calls=%d, want 2", len(runner.calls))
	}
	var payload struct {
		Body string `json:"body"`
	}
	if err := json.Unmarshal([]byte(runner.calls[1].stdin), &payload); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(payload.Body, "<!-- shepherd-question:issue-389:decision-1 -->") ||
		!strings.Contains(payload.Body, "@karthik-sivadas") ||
		!strings.Contains(payload.Body, "/shepherd decide decision-1 <option>") {
		t.Fatalf("question body missing contract: %s", payload.Body)
	}
	if strings.Contains(strings.Join(runner.calls[1].args, " "), "karthik-sivadas") {
		t.Fatal("question body leaked into process args")
	}
}

func TestPollDecisionRepliesAcceptsOnlyAuthorizedExactUneditedReply(t *testing.T) {
	t.Parallel()
	comments := `[[
{"id":1,"body":"/shepherd decide decision-1 retry","created_at":"2026-01-01T00:00:00Z","updated_at":"2026-01-01T00:00:00Z","user":{"login":"karthik-sivadas","type":"User"}},
{"id":2,"body":"/shepherd decide decision-1 stop","created_at":"2026-01-01T00:00:00Z","updated_at":"2026-01-01T00:00:00Z","user":{"login":"mallory","type":"User"}},
{"id":3,"body":"/shepherd decide decision-1 retry","created_at":"2026-01-01T00:00:00Z","updated_at":"2026-01-01T00:00:00Z","user":{"login":"karthik-sivadas","type":"Bot"}},
{"id":4,"body":"/shepherd decide decision-1 retry","created_at":"2026-01-01T00:00:00Z","updated_at":"2026-01-01T00:00:01Z","user":{"login":"karthik-sivadas","type":"User"}},
{"id":5,"body":"/shepherd decide stale retry","created_at":"2026-01-01T00:00:00Z","updated_at":"2026-01-01T00:00:00Z","user":{"login":"karthik-sivadas","type":"User"}},
{"id":6,"body":"/shepherd decide decision-1 unsafe","created_at":"2026-01-01T00:00:00Z","updated_at":"2026-01-01T00:00:00Z","user":{"login":"karthik-sivadas","type":"User"}}
]]`
	runner := &fakeRunner{responses: [][]byte{[]byte(comments)}}
	replies, err := NewClient(runner).PollDecisionReplies(context.Background(), testQuestionRequest(), "karthik-sivadas")
	if err != nil {
		t.Fatal(err)
	}
	if len(replies) != 1 || replies[0].CommentID != 1 || replies[0].Option != "retry" {
		t.Fatalf("replies=%+v", replies)
	}
}

func TestQuestionCommentRejectsAmbiguousOwnership(t *testing.T) {
	t.Parallel()
	marker := "<!-- shepherd-question:issue-389:decision-1 -->"
	runner := &fakeRunner{responses: [][]byte{[]byte(`[[{"id":1,"body":"` + marker + `"},{"id":2,"body":"` + marker + `"}]]`)}}
	if _, err := NewClient(runner).SyncQuestionComment(context.Background(), testQuestionRequest()); err == nil {
		t.Fatal("ambiguous question ownership accepted")
	}
}

func testQuestionRequest() QuestionRequest {
	return QuestionRequest{
		RequestID: "decision-1", Repository: "polymetrics-ai/cli", Issue: 389, PullRequest: 391,
		DeliveryID: "issue-389", UnitID: "execute-task/M001/S01/T01", Generation: 2,
		HeadSHA: strings.Repeat("a", 40), Evidence: "retry budget exhausted", Options: []string{"retry", "stop"},
		RecommendedOption: "retry", SafeDefault: "stop", ExpiresAt: time.Unix(1_800_000_000, 0).UTC(), Mention: "karthik-sivadas",
	}
}
