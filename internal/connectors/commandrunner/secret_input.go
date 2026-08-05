package commandrunner

import (
	"errors"
	"io"
	"os"
	"regexp"
	"strings"
)

const maxSecretInputBytes = 1 << 20

var environmentNamePattern = regexp.MustCompile(`^[A-Za-z_][A-Za-z0-9_]*$`)

// SecretInputField is one declared, typed secret-bearing request field. Path
// is a command body's dotted field path; values may never be supplied through
// a similarly named CLI flag. Only string inputs are currently supported,
// which matches the nine Zendesk Support secret-input fields.
type SecretInputField struct {
	Path     string
	Type     string
	Required bool
}

// SecretInputRequest carries only command-line source references. FromEnv
// entries have the existing `field=ENV` form; StdinFields name the one field
// whose value is read from stdin. Neither form carries a secret value in argv.
//
// LookupEnv and Stdin make the boundary testable. Nil values use os.LookupEnv
// and os.Stdin. Apply is the in-process request-body sink; errors from it are
// intentionally not wrapped because a provider or downstream implementation
// may include a secret in its error text.
type SecretInputRequest struct {
	InputMode   string
	Fields      []SecretInputField
	FromEnv     []string
	StdinFields []string
	LookupEnv   func(string) (string, bool)
	Stdin       io.Reader
	Apply       func(path, value string) error
}

type resolvedSecretInput struct {
	path  string
	value string
}

// ResolveSecretInputs materializes declared secret values from environment
// variables or stdin and sends each value directly to Apply. It returns only
// fixed, value-free errors: callers must never receive a wrapped error that
// can reveal a secret through logs, CLI stderr, or JSON output.
func ResolveSecretInputs(req SecretInputRequest) error {
	if !supportsSecretInputMode(req.InputMode) {
		return errors.New("secret input mode is unsupported")
	}
	fields, err := secretInputFields(req.Fields)
	if err != nil {
		return err
	}
	if len(req.FromEnv) == 0 && len(req.StdinFields) == 0 {
		if hasRequiredSecretInput(fields) {
			return errors.New("required secret input is missing")
		}
		return nil
	}
	if req.Apply == nil {
		return errors.New("secret input sink is unavailable")
	}

	resolved := make([]resolvedSecretInput, 0, len(req.FromEnv)+len(req.StdinFields))
	used := make(map[string]struct{}, len(req.FromEnv)+len(req.StdinFields))
	lookupEnv := req.LookupEnv
	if lookupEnv == nil {
		lookupEnv = os.LookupEnv
	}
	for _, raw := range req.FromEnv {
		if !secretInputModeAllowsEnv(req.InputMode) {
			return errors.New("secret input mode does not allow environment sources")
		}
		path, environment, ok := strings.Cut(raw, "=")
		if !ok || path == "" || environment == "" || !environmentNamePattern.MatchString(environment) {
			return errors.New("invalid secret environment source")
		}
		if _, ok := fields[path]; !ok {
			return errors.New("secret input target is not declared")
		}
		if _, exists := used[path]; exists {
			return errors.New("secret input target has multiple sources")
		}
		value, ok := lookupEnv(environment)
		if !ok || value == "" {
			return errors.New("secret input environment value is unavailable")
		}
		used[path] = struct{}{}
		resolved = append(resolved, resolvedSecretInput{path: path, value: value})
	}

	if len(req.StdinFields) > 1 {
		return errors.New("only one secret input may use stdin")
	}
	if len(req.StdinFields) == 1 {
		if !secretInputModeAllowsStdin(req.InputMode) {
			return errors.New("secret input mode does not allow stdin")
		}
		path := req.StdinFields[0]
		if _, ok := fields[path]; !ok {
			return errors.New("secret input target is not declared")
		}
		if _, exists := used[path]; exists {
			return errors.New("secret input target has multiple sources")
		}
		reader := req.Stdin
		if reader == nil {
			reader = os.Stdin
		}
		value, err := readSecretInput(reader)
		if err != nil {
			return err
		}
		used[path] = struct{}{}
		resolved = append(resolved, resolvedSecretInput{path: path, value: value})
	}

	for path, field := range fields {
		if field.Required {
			if _, ok := used[path]; !ok {
				return errors.New("required secret input is missing")
			}
		}
	}
	for i := range resolved {
		input := &resolved[i]
		err := req.Apply(input.path, input.value)
		// The request body owns its copy after Apply. Clear this short-lived
		// parser buffer on both success and failure to narrow its lifetime.
		input.value = ""
		if err != nil {
			return errors.New("apply secret input failed")
		}
	}
	return nil
}

func secretInputFields(fields []SecretInputField) (map[string]SecretInputField, error) {
	declared := make(map[string]SecretInputField, len(fields))
	for _, field := range fields {
		if !strings.HasPrefix(field.Path, "body.") || strings.TrimPrefix(field.Path, "body.") == "" || field.Type != "string" {
			return nil, errors.New("invalid declared secret input field")
		}
		if _, err := validateDottedTargetPath(strings.TrimPrefix(field.Path, "body."), "secret input field"); err != nil {
			return nil, errors.New("invalid declared secret input field")
		}
		if _, exists := declared[field.Path]; exists {
			return nil, errors.New("duplicate declared secret input field")
		}
		declared[field.Path] = field
	}
	return declared, nil
}

func hasRequiredSecretInput(fields map[string]SecretInputField) bool {
	for _, field := range fields {
		if field.Required {
			return true
		}
	}
	return false
}

func supportsSecretInputMode(inputMode string) bool {
	switch inputMode {
	case "env", "stdin", "env_or_stdin":
		return true
	default:
		return false
	}
}

func secretInputModeAllowsEnv(inputMode string) bool {
	return inputMode == "env" || inputMode == "env_or_stdin"
}

func secretInputModeAllowsStdin(inputMode string) bool {
	return inputMode == "stdin" || inputMode == "env_or_stdin"
}

func readSecretInput(reader io.Reader) (string, error) {
	data, err := io.ReadAll(io.LimitReader(reader, maxSecretInputBytes+1))
	if err != nil {
		return "", errors.New("read secret input from stdin failed")
	}
	if len(data) > maxSecretInputBytes {
		return "", errors.New("secret input from stdin exceeds the maximum size")
	}
	value := strings.TrimRight(string(data), "\r\n")
	if value == "" {
		return "", errors.New("secret input is empty")
	}
	return value, nil
}
