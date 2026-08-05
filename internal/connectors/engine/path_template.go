package engine

import (
	"fmt"
	"strings"
)

type pathTemplateVariable struct {
	Name  string
	Start int
	End   int
}

func parsePathTemplate(template string) ([]pathTemplateVariable, error) {
	variables := make([]pathTemplateVariable, 0, strings.Count(template, "{"))
	for i := 0; i < len(template); {
		switch template[i] {
		case '{':
			closing := strings.IndexByte(template[i+1:], '}')
			if closing < 0 {
				return nil, fmt.Errorf("unclosed path variable at byte %d", i)
			}
			closing += i + 1
			name := template[i+1 : closing]
			if !validPathTemplateVariableName(name) {
				return nil, fmt.Errorf("invalid path variable %q", name)
			}
			variables = append(variables, pathTemplateVariable{Name: name, Start: i, End: closing + 1})
			i = closing + 1
		case '}':
			return nil, fmt.Errorf("unexpected closing brace at byte %d", i)
		default:
			i++
		}
	}
	return variables, nil
}

func validPathTemplateVariableName(name string) bool {
	if name == "" || !isPathTemplateVariableStart(name[0]) {
		return false
	}
	for i := 1; i < len(name); i++ {
		if !isPathTemplateVariableContinue(name[i]) {
			return false
		}
	}
	return true
}

func isPathTemplateVariableStart(ch byte) bool {
	return ch == '_' || ch >= 'A' && ch <= 'Z' || ch >= 'a' && ch <= 'z'
}

func isPathTemplateVariableContinue(ch byte) bool {
	return isPathTemplateVariableStart(ch) || ch >= '0' && ch <= '9' || ch == '.' || ch == '-'
}
