package github

import (
	"context"
	"encoding/json"
	"strings"
	"testing"
)

type fakeRunner struct {
	responses [][]byte
	calls     []call
}

type call struct {
	args  []string
	stdin string
}

func (f *fakeRunner) Run(_ context.Context, args []string, stdin []byte) ([]byte, error) {
	f.calls = append(f.calls, call{args: append([]string(nil), args...), stdin: string(stdin)})
	if len(f.responses) == 0 {
		return nil, nil
	}
	response := f.responses[0]
	f.responses = f.responses[1:]
	return response, nil
}

func TestDecisionCommentSyncCreatesThenUpdatesMarkerOwnedComment(t *testing.T) {
	t.Parallel()

	marker := "<!-- shepherd-decisions:issue-380 -->"
	tests := []struct {
		name       string
		comments   string
		wantMethod string
		wantPath   string
	}{
		{name: "create", comments: `[[{"id":1,"body":"unrelated"}]]`, wantMethod: "POST", wantPath: "repos/polymetrics-ai/cli/issues/388/comments"},
		{name: "update", comments: `[[{"id":42,"body":"` + marker + `\n\nold"}]]`, wantMethod: "PATCH", wantPath: "repos/polymetrics-ai/cli/issues/comments/42"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			runner := &fakeRunner{responses: [][]byte{[]byte(test.comments)}}
			client := NewClient(runner)
			if err := client.SyncDecisionComment(context.Background(), Target{Repository: "polymetrics-ai/cli", PullRequest: 388, DeliveryID: "issue-380"}, "## Shepherd decisions\n\n| Actor |\n|---|\n| shepherd |"); err != nil {
				t.Fatal(err)
			}
			if len(runner.calls) != 2 {
				t.Fatalf("calls=%d, want 2", len(runner.calls))
			}
			write := runner.calls[1]
			joined := strings.Join(write.args, " ")
			if !strings.Contains(joined, test.wantMethod) || !strings.Contains(joined, test.wantPath) {
				t.Fatalf("write args=%q", joined)
			}
			if strings.Contains(joined, "Shepherd decisions") {
				t.Fatal("comment body leaked into process arguments")
			}
			var payload struct {
				Body string `json:"body"`
			}
			if err := json.Unmarshal([]byte(write.stdin), &payload); err != nil {
				t.Fatal(err)
			}
			if !strings.Contains(payload.Body, marker) || !strings.Contains(payload.Body, "shepherd") {
				t.Fatalf("write stdin=%q", write.stdin)
			}
		})
	}
}

func TestDecisionCommentSyncRejectsAmbiguousOwnershipAndUnsafeTarget(t *testing.T) {
	t.Parallel()
	marker := "<!-- shepherd-decisions:issue-380 -->"
	runner := &fakeRunner{responses: [][]byte{[]byte(`[[{"id":1,"body":"` + marker + `"},{"id":2,"body":"` + marker + `"}]]`)}}
	client := NewClient(runner)
	if err := client.SyncDecisionComment(context.Background(), Target{Repository: "polymetrics-ai/cli", PullRequest: 388, DeliveryID: "issue-380"}, "summary"); err == nil {
		t.Fatal("ambiguous owned comments accepted")
	}
	if err := client.SyncDecisionComment(context.Background(), Target{Repository: "../escape", PullRequest: 388, DeliveryID: "issue-380"}, "summary"); err == nil {
		t.Fatal("unsafe repository accepted")
	}
}
