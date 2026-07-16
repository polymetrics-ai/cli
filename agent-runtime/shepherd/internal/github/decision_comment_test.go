package github

import (
	"context"
	"encoding/json"
	"strings"
	"testing"

	"github.com/polymetrics-ai/cli/agent-runtime/shepherd/internal/outbox"
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

func TestOutboxCommentTransportKeepsBodyOnStdin(t *testing.T) {
	t.Parallel()
	target := outbox.Target{Repository: "polymetrics-ai/cli", Issue: 389, PullRequest: 390}
	runner := &fakeRunner{responses: [][]byte{
		[]byte(`[[{"id":1,"body":"unrelated","user":{"login":"shepherd-test","type":"User"}}]]`),
		[]byte(`{"id":42,"body":"<!-- shepherd-effect:test -->\n\nbounded body\n","user":{"login":"shepherd-test","type":"User"}}`),
	}}
	port := cliCommentPort{runner: runner}
	comments, err := port.ListComments(context.Background(), target)
	if err != nil {
		t.Fatal(err)
	}
	if len(comments) != 1 || comments[0].ID != 1 {
		t.Fatalf("comments=%+v", comments)
	}
	body := "<!-- shepherd-effect:test -->\n\nbounded body\n"
	created, err := port.CreateComment(context.Background(), target, body)
	if err != nil {
		t.Fatal(err)
	}
	if created.ID != 42 || created.Author != "shepherd-test" || len(runner.calls) != 2 {
		t.Fatalf("comment=%+v calls=%d", created, len(runner.calls))
	}
	write := runner.calls[1]
	if joined := strings.Join(write.args, " "); strings.Contains(joined, "bounded body") ||
		!strings.Contains(joined, "--method POST") || !strings.Contains(joined, "--input -") {
		t.Fatalf("unsafe write args=%q", joined)
	}
	var payload struct {
		Body string `json:"body"`
	}
	if err := json.Unmarshal([]byte(write.stdin), &payload); err != nil {
		t.Fatal(err)
	}
	if payload.Body != body {
		t.Fatalf("stdin body=%q", payload.Body)
	}
}

func TestOutboxCommentTransportRejectsUnsafeTargets(t *testing.T) {
	t.Parallel()
	port := cliCommentPort{runner: &fakeRunner{}}
	if _, err := port.ListComments(context.Background(), outbox.Target{Repository: "../escape", Issue: 389, PullRequest: 390}); err == nil {
		t.Fatal("unsafe repository accepted")
	}
	if _, err := port.UpdateComment(context.Background(), outbox.Target{Repository: "polymetrics-ai/cli", Issue: 389, PullRequest: 390}, 0, "body"); err == nil {
		t.Fatal("non-positive comment identity accepted")
	}
}
