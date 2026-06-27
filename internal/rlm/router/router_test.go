package router

import (
	"context"
	"errors"
	"testing"
)

func TestClassifyHeuristic(t *testing.T) {
	cases := []struct {
		name    string
		request string
		want    Task
	}{
		{"simple top-n", "show top 10 customers by revenue", TaskSimpleQuery},
		{"simple count", "count all orders from last week", TaskSimpleQuery},
		{"simple list", "list every contact in the acme table", TaskSimpleQuery},
		{"analysis correlation", "find the correlation between spend and region", TaskDataAnalysis},
		{"analysis trend", "what is the monthly revenue trend by product", TaskDataAnalysis},
		{"analysis join", "join orders and customers and break down revenue", TaskDataAnalysis},
		{"analysis bus factor", "what is the bus factor for the rails/rails repository", TaskDataAnalysis},
		{"ml predict", "predict which customers will churn next month", TaskML},
		{"ml train", "train a model to classify leads by quality", TaskML},
		{"ml cluster", "cluster customers into segments", TaskML},
	}
	r := &Router{} // no LLM → heuristic only
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := r.Classify(context.Background(), tc.request, "")
			if got.Task != tc.want {
				t.Fatalf("Classify(%q).Task = %q, want %q (indicators=%v)",
					tc.request, got.Task, tc.want, got.KeyIndicators)
			}
		})
	}
}

type mockClassifier struct {
	resp string
	err  error
}

func (m mockClassifier) Complete(_ context.Context, _ string) (string, error) {
	return m.resp, m.err
}

func TestClassifyLLM_Tier2(t *testing.T) {
	// LLM returns a valid (fenced) JSON decision that differs from the heuristic
	// ("export everything" would be simple by heuristic; LLM says ml).
	llm := mockClassifier{resp: "```json\n{\"task\":\"ml\",\"confidence\":0.91,\"reasoning\":\"needs a model\",\"key_indicators\":[\"propensity\"]}\n```"}
	r := &Router{LLM: llm}
	got := r.Classify(context.Background(), "rank leads", "")
	if got.Task != TaskML {
		t.Fatalf("Task = %q, want ml", got.Task)
	}
	if got.Confidence != 0.91 {
		t.Fatalf("Confidence = %v, want 0.91", got.Confidence)
	}
}

func TestClassifyLLM_FallsBackOnError(t *testing.T) {
	llm := mockClassifier{err: errors.New("boom")}
	r := &Router{LLM: llm}
	got := r.Classify(context.Background(), "predict churn", "")
	if got.Task != TaskML { // heuristic catches "predict"/"churn"
		t.Fatalf("fallback Task = %q, want ml", got.Task)
	}
}

func TestClassifyLLM_FallsBackOnBadJSON(t *testing.T) {
	llm := mockClassifier{resp: "I think this is a simple query, friend."}
	r := &Router{LLM: llm}
	got := r.Classify(context.Background(), "show top 10 customers", "")
	if got.Task != TaskSimpleQuery {
		t.Fatalf("fallback Task = %q, want simple_query", got.Task)
	}
}

func TestClassifyLLM_RejectsInvalidTask(t *testing.T) {
	llm := mockClassifier{resp: `{"task":"banana","confidence":0.9}`}
	r := &Router{LLM: llm}
	got := r.Classify(context.Background(), "count orders", "")
	if got.Task != TaskSimpleQuery { // invalid task → fall back to heuristic
		t.Fatalf("Task = %q, want simple_query", got.Task)
	}
}

func TestExtractJSON(t *testing.T) {
	in := "Sure! ```json\n{\"task\":\"ml\",\"nested\":{\"a\":1}}\n``` done"
	got := extractJSON(in)
	want := `{"task":"ml","nested":{"a":1}}`
	if got != want {
		t.Fatalf("extractJSON = %q, want %q", got, want)
	}
}

func TestTaskIsRLM(t *testing.T) {
	if TaskSimpleQuery.IsRLM() {
		t.Error("simple_query should not be RLM")
	}
	if !TaskDataAnalysis.IsRLM() || !TaskML.IsRLM() {
		t.Error("data_analysis and ml should be RLM")
	}
}
