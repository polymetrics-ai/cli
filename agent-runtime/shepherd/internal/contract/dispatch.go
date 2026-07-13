package contract

import (
	"errors"
	"fmt"
	"strings"
)

const MaxHandoffLines = 40

type Dispatch struct {
	Objective    string   `json:"objective"`
	OutputFormat string   `json:"output_format"`
	ToolGuidance []string `json:"tool_guidance"`
	Tools        []Tool   `json:"tools"`
	Boundaries   []string `json:"boundaries"`
	WriteScope   []string `json:"write_scope"`
}

type Tool string

const (
	ToolRead      Tool = "read"
	ToolEdit      Tool = "edit"
	ToolTest      Tool = "test"
	ToolGitStatus Tool = "git.status"
)

var allowedTools = map[Tool]struct{}{
	ToolRead: {}, ToolEdit: {}, ToolTest: {}, ToolGitStatus: {},
}

func (d Dispatch) Validate() error {
	if strings.TrimSpace(d.Objective) == "" {
		return errors.New("objective is required")
	}
	if strings.TrimSpace(d.OutputFormat) == "" {
		return errors.New("output format is required")
	}
	if len(d.ToolGuidance) == 0 {
		return errors.New("tool guidance is required")
	}
	if len(d.Tools) == 0 {
		return errors.New("typed tools are required")
	}
	if len(d.Boundaries) == 0 {
		return errors.New("boundaries are required")
	}
	if len(d.WriteScope) == 0 {
		return errors.New("write scope is required")
	}
	for _, tool := range d.Tools {
		if _, ok := allowedTools[tool]; !ok {
			return fmt.Errorf("tool %q is not allowed", tool)
		}
	}
	for _, scope := range d.WriteScope {
		if scope == "" || strings.Contains(scope, "..") || strings.HasPrefix(scope, "/") {
			return fmt.Errorf("unsafe write scope %q", scope)
		}
	}
	return nil
}

func ValidateHandoff(handoff string) error {
	trimmed := strings.TrimSuffix(handoff, "\n")
	if trimmed == "" {
		return errors.New("handoff is required")
	}
	if strings.Count(trimmed, "\n")+1 > MaxHandoffLines {
		return fmt.Errorf("handoff exceeds %d lines", MaxHandoffLines)
	}
	return nil
}
