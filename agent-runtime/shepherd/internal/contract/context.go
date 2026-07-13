package contract

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strings"
)

const MaxIssueContextBytes = 1024 * 1024

type IssueContext struct {
	Issue              int      `json:"issue"`
	ParentIssue        int      `json:"parent_issue"`
	Objective          string   `json:"objective"`
	Scope              []string `json:"scope"`
	NonGoals           []string `json:"non_goals"`
	AcceptanceCriteria []string `json:"acceptance_criteria"`
	Dependencies       []int    `json:"dependencies"`
	WriteScope         []string `json:"write_scope"`
	RequiredReading    []string `json:"required_reading"`
	RequiredSkills     []string `json:"required_skills"`
	TDD                struct {
		Red      string `json:"red"`
		Green    string `json:"green"`
		Refactor string `json:"refactor"`
	} `json:"tdd"`
	Verification []string `json:"verification"`
	Safety       []string `json:"safety"`
	HumanGates   []string `json:"human_gates"`
	Branch       string   `json:"branch"`
	PRBase       string   `json:"pr_base"`
	ReviewRoute  string   `json:"review_route"`
	Sources      []string `json:"sources"`
}

func DecodeIssueContext(reader io.Reader, expectedIssue int) (IssueContext, []byte, error) {
	if expectedIssue <= 0 {
		return IssueContext{}, nil, errors.New("expected issue is required")
	}
	raw, err := io.ReadAll(io.LimitReader(reader, MaxIssueContextBytes+1))
	if err != nil {
		return IssueContext{}, nil, fmt.Errorf("read issue context: %w", err)
	}
	if len(raw) == 0 || len(raw) > MaxIssueContextBytes {
		return IssueContext{}, nil, errors.New("issue context is empty or oversized")
	}
	var context IssueContext
	decoder := json.NewDecoder(bytes.NewReader(raw))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&context); err != nil {
		return IssueContext{}, nil, fmt.Errorf("decode issue context: %w", err)
	}
	if decoder.Decode(&struct{}{}) != io.EOF {
		return IssueContext{}, nil, errors.New("issue context must contain one JSON value")
	}
	if context.Issue != expectedIssue || context.ParentIssue <= 0 {
		return IssueContext{}, nil, errors.New("issue context does not match requested issue")
	}
	if strings.TrimSpace(context.Objective) == "" || strings.TrimSpace(context.Branch) == "" ||
		strings.TrimSpace(context.PRBase) == "" || strings.TrimSpace(context.ReviewRoute) == "" {
		return IssueContext{}, nil, errors.New("objective, branch, PR base, and review route are required")
	}
	required := map[string][]string{
		"scope": context.Scope, "acceptance criteria": context.AcceptanceCriteria,
		"write scope": context.WriteScope, "required reading": context.RequiredReading,
		"required skills": context.RequiredSkills, "verification": context.Verification,
		"safety": context.Safety, "human gates": context.HumanGates, "sources": context.Sources,
	}
	for name, values := range required {
		if len(values) == 0 {
			return IssueContext{}, nil, fmt.Errorf("%s is required", name)
		}
		for _, value := range values {
			if strings.TrimSpace(value) == "" || strings.ContainsAny(value, "\r\x00") {
				return IssueContext{}, nil, fmt.Errorf("%s contains an unsafe value", name)
			}
		}
	}
	for _, scope := range context.WriteScope {
		if strings.HasPrefix(scope, "/") || strings.Contains(scope, "..") {
			return IssueContext{}, nil, fmt.Errorf("unsafe write scope %q", scope)
		}
	}
	if strings.TrimSpace(context.TDD.Red) == "" || strings.TrimSpace(context.TDD.Green) == "" || strings.TrimSpace(context.TDD.Refactor) == "" {
		return IssueContext{}, nil, errors.New("red, green, and refactor TDD evidence are required")
	}
	return context, raw, nil
}
