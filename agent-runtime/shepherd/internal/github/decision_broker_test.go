package github

import (
	"context"
	"strings"
	"testing"
	"time"
)

func TestPollDecisionRepliesAcceptsOnlyAuthorizedExactUneditedReply(t *testing.T) {
	t.Parallel()
	comments := `[[
{"id":101,"body":"/shepherd decide decision-1 retry","created_at":"2026-01-01T00:00:00Z","updated_at":"2026-01-01T00:00:00Z","user":{"login":"karthik-sivadas","id":6113982,"type":"User"}},
{"id":102,"body":"/shepherd decide decision-1 stop","created_at":"2026-01-01T00:00:00Z","updated_at":"2026-01-01T00:00:00Z","user":{"login":"mallory","id":1,"type":"User"}},
{"id":103,"body":"/shepherd decide decision-1 retry","created_at":"2026-01-01T00:00:00Z","updated_at":"2026-01-01T00:00:00Z","user":{"login":"karthik-sivadas","id":6113982,"type":"Bot"}},
{"id":104,"body":"/shepherd decide decision-1 retry","created_at":"2026-01-01T00:00:00Z","updated_at":"2026-01-01T00:00:01Z","user":{"login":"karthik-sivadas","id":6113982,"type":"User"}},
{"id":105,"body":"/shepherd decide stale retry","created_at":"2026-01-01T00:00:00Z","updated_at":"2026-01-01T00:00:00Z","user":{"login":"karthik-sivadas","id":6113982,"type":"User"}},
{"id":106,"body":"/shepherd decide decision-1 unsafe","created_at":"2026-01-01T00:00:00Z","updated_at":"2026-01-01T00:00:00Z","user":{"login":"karthik-sivadas","id":6113982,"type":"User"}},
{"id":107,"body":"/shepherd decide decision-1 stop","created_at":"2028-01-01T00:00:00Z","updated_at":"2028-01-01T00:00:00Z","user":{"login":"karthik-sivadas","id":6113982,"type":"User"}}
]]`
	runner := &fakeRunner{responses: [][]byte{[]byte(comments)}}
	replies, err := newReplyPoller(cliReplyReader{runner: runner}).PollDecisionReplies(context.Background(), testQuestionRequest(), "karthik-sivadas")
	if err != nil {
		t.Fatal(err)
	}
	if len(replies) != 1 || replies[0].CommentID != 101 || replies[0].AuthorID != 6113982 ||
		replies[0].CreatedAt.IsZero() || replies[0].Option != "retry" {
		t.Fatalf("replies=%+v", replies)
	}
}

func TestPollDecisionRepliesBlocksMultipleValidAnswers(t *testing.T) {
	t.Parallel()
	comments := `[[
{"id":101,"body":"/shepherd decide decision-1 retry","created_at":"2026-01-01T00:00:00Z","updated_at":"2026-01-01T00:00:00Z","user":{"login":"karthik-sivadas","id":6113982,"type":"User"}},
{"id":102,"body":"/shepherd decide decision-1 stop","created_at":"2026-01-01T00:00:00Z","updated_at":"2026-01-01T00:00:00Z","user":{"login":"karthik-sivadas","id":6113982,"type":"User"}}
]]`
	runner := &fakeRunner{responses: [][]byte{[]byte(comments)}}
	if _, err := newReplyPoller(cliReplyReader{runner: runner}).PollDecisionReplies(
		context.Background(), testQuestionRequest(), "karthik-sivadas"); err == nil {
		t.Fatal("multiple valid decision replies were accepted")
	}
}

func testQuestionRequest() QuestionRequest {
	return QuestionRequest{
		RequestID: "decision-1", Repository: "polymetrics-ai/cli", Issue: 389, PullRequest: 391,
		DeliveryID: "issue-389", UnitID: "execute-task/M001/S01/T01", Generation: 2,
		HeadSHA: strings.Repeat("a", 40), Evidence: "retry budget exhausted", Options: []string{"retry", "stop"},
		RecommendedOption: "retry", SafeDefault: "stop", ExpiresAt: time.Unix(1_800_000_000, 0).UTC(),
		Mention: "karthik-sivadas", QuestionCommentID: 77,
	}
}
