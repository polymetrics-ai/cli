package cli

import (
	"encoding/json"
	"io"
	"strconv"
	"strings"

	"polymetrics.ai/internal/app"
	"polymetrics.ai/internal/safety"
)

var errUsage = usageErrorf("invalid usage")

type parsedFlags struct {
	values map[string][]string
}

func (p parsedFlags) first(name string) string {
	values := p.values[name]
	if len(values) == 0 {
		return ""
	}
	return values[len(values)-1]
}

func parseGlobal(args []string) (root string, jsonOut bool, clean []string) {
	root = "."
	for i := 0; i < len(args); i++ {
		arg := args[i]
		switch {
		case arg == "--json":
			jsonOut = true
		case arg == "--root" && i+1 < len(args):
			root = args[i+1]
			i++
		case strings.HasPrefix(arg, "--root="):
			root = strings.TrimPrefix(arg, "--root=")
		default:
			clean = append(clean, arg)
		}
	}
	return root, jsonOut, clean
}

func parseFlags(args []string) parsedFlags {
	out := parsedFlags{values: map[string][]string{}}
	for i := 0; i < len(args); i++ {
		arg := args[i]
		if !strings.HasPrefix(arg, "--") {
			out.values["_"] = append(out.values["_"], arg)
			continue
		}
		keyval := strings.TrimPrefix(arg, "--")
		key, value, ok := strings.Cut(keyval, "=")
		if !ok {
			value = "true"
			if i+1 < len(args) && !strings.HasPrefix(args[i+1], "--") {
				value = args[i+1]
				i++
			}
		}
		out.values[key] = append(out.values[key], value)
	}
	return out
}

func parseEndpoint(raw string) (app.EndpointConfig, error) {
	connector, credential, ok := strings.Cut(raw, ":")
	if !ok || connector == "" || credential == "" {
		return app.EndpointConfig{}, validationErrorf("invalid endpoint %q, want connector:credential", raw)
	}
	if err := safety.ValidateIdentifier(connector, "connector"); err != nil {
		return app.EndpointConfig{}, validationErrorf("%v", err)
	}
	if err := safety.ValidateIdentifier(credential, "credential"); err != nil {
		return app.EndpointConfig{}, validationErrorf("%v", err)
	}
	return app.EndpointConfig{Connector: connector, Credential: credential}, nil
}

func keyValues(values []string) (map[string]string, error) {
	out := map[string]string{}
	for _, item := range values {
		key, value, ok := strings.Cut(item, "=")
		if !ok || key == "" {
			return nil, validationErrorf("invalid key-value %q, want key=value", item)
		}
		out[key] = value
	}
	return out, nil
}

func colonValues(values []string) (map[string]string, error) {
	out := map[string]string{}
	for _, item := range values {
		key, value, ok := strings.Cut(item, ":")
		if !ok || key == "" || value == "" {
			return nil, validationErrorf("invalid mapping %q, want source:destination", item)
		}
		out[key] = value
	}
	return out, nil
}

func valueOr(value, fallback string) string {
	if value == "" {
		return fallback
	}
	return value
}

func parseIntFlag(name, value string, fallback int) (int, error) {
	if value == "" {
		return fallback, nil
	}
	n, err := strconv.Atoi(value)
	if err != nil {
		return 0, validationErrorf("invalid --%s %q, want integer", name, value)
	}
	return n, nil
}

func writeJSON(w io.Writer, v any) error {
	// Stamp every JSON envelope with the API version so agents get a consistent
	// contract (api_version + kind) on every response, not just on errors.
	if env, ok := v.(envelope); ok {
		if _, exists := env["api_version"]; !exists {
			env["api_version"] = apiVersion
		}
	}
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(v)
}
