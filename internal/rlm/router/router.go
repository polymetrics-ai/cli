// Package router classifies a natural-language data request into one of three
// task kinds so `pm extract` can choose a backend: a simple SQL query, a
// data-analysis RLM run, or a machine-learning RLM run.
//
// It mirrors the two-tier classifier from the previous polymetrics RLM
// (rlm_ruby's QueryClassifier): an optional LLM classifier (Tier 2) with a
// deterministic keyword heuristic (Tier 1) as the always-available fallback.
package router

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
)

// Task is the routing decision.
type Task string

const (
	// TaskSimpleQuery: a direct retrieval answerable by a single SELECT.
	TaskSimpleQuery Task = "simple_query"
	// TaskDataAnalysis: a multi-step analytical question (joins, aggregates,
	// correlations, trends) better served by the RLM agent.
	TaskDataAnalysis Task = "data_analysis"
	// TaskML: a machine-learning task (train/predict/cluster/score).
	TaskML Task = "ml"
)

// IsRLM reports whether the task requires the RLM agent rather than a query.
func (t Task) IsRLM() bool { return t == TaskDataAnalysis || t == TaskML }

// Decision is the classifier output.
type Decision struct {
	Task          Task     `json:"task"`
	Confidence    float64  `json:"confidence"`
	Reasoning     string   `json:"reasoning"`
	KeyIndicators []string `json:"key_indicators,omitempty"`
	// SuggestedSQL is an optional LLM-proposed SELECT for simple_query routes.
	// It MUST be validated (validateSelectOnly) before execution.
	SuggestedSQL string `json:"suggested_sql,omitempty"`
}

// LLMClassifier is the optional Tier-2 hook. It is satisfied structurally by
// rlm.LLMClient, so the router does not import internal/rlm.
type LLMClassifier interface {
	Complete(ctx context.Context, prompt string) (string, error)
}

// Router classifies requests. LLM may be nil (heuristic-only).
type Router struct {
	LLM LLMClassifier
}

// Classify routes a request. When an LLM is configured it is tried first; on any
// error (transport, bad JSON, invalid task) it falls back to the heuristic.
// schema is an optional DESCRIBE-style summary of available tables/columns to
// ground the LLM (ignored by the heuristic).
func (r *Router) Classify(ctx context.Context, request, schema string) Decision {
	if r.LLM != nil {
		if d, err := r.classifyLLM(ctx, request, schema); err == nil {
			return d
		}
	}
	return classifyHeuristic(request)
}

var mlKeywords = []string{
	"train", "predict", "prediction", "forecast", "classify", "classification",
	"regression", "cluster", "clustering", "segment customers", "anomaly",
	"accuracy", "precision", "recall", "f1", "cross-validation", "feature importance",
	"logistic", "random forest", "naive bayes", "svm", "decision tree",
	"neural network", "churn", "recommend", "recommendation", "propensity",
}

var analysisKeywords = []string{
	"correlat", "trend", "cohort", "distribution", "group by", "aggregate",
	"average over", "compare", "comparison", "percentile", "growth", "retention",
	"funnel", "join", "across tables", "month over month", "year over year",
	"breakdown", "rolling", "moving average",
}

func classifyHeuristic(request string) Decision {
	q := strings.ToLower(request)

	var ml []string
	for _, kw := range mlKeywords {
		if strings.Contains(q, kw) {
			ml = append(ml, kw)
		}
	}
	if len(ml) > 0 {
		return Decision{
			Task: TaskML, Confidence: 0.6,
			Reasoning:     "heuristic: machine-learning keywords detected",
			KeyIndicators: ml,
		}
	}

	var an []string
	for _, kw := range analysisKeywords {
		if strings.Contains(q, kw) {
			an = append(an, kw)
		}
	}
	if len(an) > 0 {
		return Decision{
			Task: TaskDataAnalysis, Confidence: 0.55,
			Reasoning:     "heuristic: multi-step analysis keywords detected",
			KeyIndicators: an,
		}
	}

	return Decision{
		Task: TaskSimpleQuery, Confidence: 0.5,
		Reasoning: "heuristic: no ML/analysis signals; treat as simple retrieval",
	}
}

func (r *Router) classifyLLM(ctx context.Context, request, schema string) (Decision, error) {
	raw, err := r.LLM.Complete(ctx, buildClassifyPrompt(request, schema))
	if err != nil {
		return Decision{}, err
	}
	var out struct {
		Task          string   `json:"task"`
		Confidence    float64  `json:"confidence"`
		Reasoning     string   `json:"reasoning"`
		KeyIndicators []string `json:"key_indicators"`
		SuggestedSQL  string   `json:"suggested_sql"`
	}
	if err := json.Unmarshal([]byte(extractJSON(raw)), &out); err != nil {
		return Decision{}, fmt.Errorf("router: parse classifier output: %w", err)
	}
	task := Task(strings.TrimSpace(out.Task))
	if task != TaskSimpleQuery && task != TaskDataAnalysis && task != TaskML {
		return Decision{}, fmt.Errorf("router: invalid task %q", out.Task)
	}
	return Decision{
		Task:          task,
		Confidence:    out.Confidence,
		Reasoning:     out.Reasoning,
		KeyIndicators: out.KeyIndicators,
		SuggestedSQL:  strings.TrimSpace(out.SuggestedSQL),
	}, nil
}

func buildClassifyPrompt(request, schema string) string {
	var b strings.Builder
	b.WriteString("You are a query router for a local data warehouse. Classify the user's request into exactly one task.\n")
	b.WriteString("Tasks:\n")
	b.WriteString("- simple_query: a single SELECT can answer it (lookups, counts, top-N, filters, simple sorts).\n")
	b.WriteString("- data_analysis: needs multi-step analysis (joins across tables, aggregations with derived metrics, correlations, trends, cohorts).\n")
	b.WriteString("- ml: needs machine learning (train/predict/classify/cluster/score/forecast).\n")
	if schema != "" {
		b.WriteString("\nAvailable schema:\n")
		b.WriteString(schema)
		b.WriteString("\n")
	}
	b.WriteString("\nUser request:\n")
	b.WriteString(request)
	b.WriteString("\n\nRespond with ONLY a JSON object: ")
	b.WriteString(`{"task":"simple_query|data_analysis|ml","confidence":0.0-1.0,"reasoning":"...","key_indicators":["..."],"suggested_sql":"<a single SELECT, only for simple_query, else empty>"}`)
	return b.String()
}

// extractJSON returns the first balanced top-level JSON object in s, tolerating
// markdown code fences and surrounding prose from an LLM.
func extractJSON(s string) string {
	s = strings.TrimSpace(s)
	if i := strings.Index(s, "```"); i >= 0 {
		rest := s[i+3:]
		if nl := strings.IndexByte(rest, '\n'); nl >= 0 {
			rest = rest[nl+1:]
		}
		if j := strings.Index(rest, "```"); j >= 0 {
			s = strings.TrimSpace(rest[:j])
		}
	}
	start := strings.IndexByte(s, '{')
	if start < 0 {
		return s
	}
	depth := 0
	inStr := false
	esc := false
	for i := start; i < len(s); i++ {
		c := s[i]
		switch {
		case esc:
			esc = false
		case c == '\\' && inStr:
			esc = true
		case c == '"':
			inStr = !inStr
		case inStr:
		case c == '{':
			depth++
		case c == '}':
			depth--
			if depth == 0 {
				return s[start : i+1]
			}
		}
	}
	return s[start:]
}
