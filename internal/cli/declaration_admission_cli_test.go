package cli

import (
	"bytes"
	"context"
	"errors"
	"strings"
	"testing"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/defs"
	"polymetrics.ai/internal/connectors/engine"
)

func TestGitHubLabelDeleteValidatesRequiredInputBeforeCredentialResolution(t *testing.T) {
	root := t.TempDir()
	var initOut, initErr bytes.Buffer
	if code := Run([]string{"init", "--root", root, "--json"}, &initOut, &initErr); code != 0 {
		t.Fatalf("init code=%d stdout=%s stderr=%s", code, initOut.String(), initErr.String())
	}

	var helpOut, helpErr bytes.Buffer
	if code := Run([]string{"github", "label", "delete", "--help", "--root", root, "--json"}, &helpOut, &helpErr); code != 0 {
		t.Fatalf("label delete help code=%d stdout=%s stderr=%s", code, helpOut.String(), helpErr.String())
	}
	if !strings.Contains(helpOut.String(), "--name") {
		t.Fatalf("label delete help omitted required input: %s", helpOut.String())
	}

	var missingOut, missingErr bytes.Buffer
	code := Run([]string{"github", "label", "delete", "--root", root, "--json"}, &missingOut, &missingErr)
	if code == 0 || !strings.Contains(missingOut.String()+missingErr.String(), "missing required flag --name") {
		t.Fatalf("bare label delete code=%d stdout=%s stderr=%s", code, missingOut.String(), missingErr.String())
	}
	if strings.Contains(missingOut.String()+missingErr.String(), "missing --credential") {
		t.Fatalf("bare label delete reached credential resolution before input validation: stdout=%s stderr=%s", missingOut.String(), missingErr.String())
	}

	var validOut, validErr bytes.Buffer
	code = Run([]string{"github", "label", "delete", "--name", "bug", "--root", root, "--json"}, &validOut, &validErr)
	if code == 0 || !strings.Contains(validOut.String()+validErr.String(), "missing --credential") {
		t.Fatalf("complete label delete code=%d stdout=%s stderr=%s", code, validOut.String(), validErr.String())
	}
}

func TestConnectorCommandPlanValidatesRequestBeforePlanLookup(t *testing.T) {
	root := t.TempDir()
	var initOut, initErr bytes.Buffer
	if code := Run([]string{"init", "--root", root, "--json"}, &initOut, &initErr); code != 0 {
		t.Fatalf("init code=%d stdout=%s stderr=%s", code, initOut.String(), initErr.String())
	}

	tests := []struct {
		name string
		args []string
		want string
	}{
		{
			name: "unknown argv", want: "unknown flag --bogus",
			args: []string{"github", "label", "delete", "--plan", "rplan_missing", "--name", "bug", "--bogus", "value", "--root", root, "--json"},
		},
		{
			name: "configured enum", want: "configured value",
			args: []string{"freshchat", "agents", "list", "--plan", "rplan_missing", "--config", "agents_is_deactivated=not-a-deactivation-state", "--root", root, "--json"},
		},
	}
	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			var stdout, stderr bytes.Buffer
			code := Run(testCase.args, &stdout, &stderr)
			combined := stdout.String() + stderr.String()
			if code == 0 || !strings.Contains(combined, testCase.want) {
				t.Fatalf("plan input code=%d stdout=%s stderr=%s; want request validation %q before plan lookup", code, stdout.String(), stderr.String(), testCase.want)
			}
			if strings.Contains(combined, `reverse plan "rplan_missing" not found`) || strings.Contains(combined, "missing --credential") {
				t.Fatalf("plan input crossed App state or credential boundary: stdout=%s stderr=%s", stdout.String(), stderr.String())
			}
		})
	}
}

func TestFreshchatConfiguredEnumValidatesBeforeCredentialResolution(t *testing.T) {
	bundle, err := engine.Load(defs.FS, "freshchat")
	if err != nil {
		t.Fatalf("load shipped Freshchat bundle: %v", err)
	}
	registry := connectors.NewEmptyRegistry()
	registry.Register(engine.New(bundle, nil))

	const invalidValue = "not-a-deactivation-state"
	var invalidOut, invalidErr bytes.Buffer
	err = runMaybeConnectorCommandWithRegistry(context.Background(), t.TempDir(), "freshchat", []string{
		"agents", "list",
		"--config", "agents_is_deactivated=" + invalidValue,
	}, &invalidOut, &invalidErr, true, registry)
	combined := invalidOut.String() + invalidErr.String()
	if err == nil || !strings.Contains(err.Error(), "configured value") {
		t.Fatalf("invalid Freshchat config error=%v stdout=%s stderr=%s", err, invalidOut.String(), invalidErr.String())
	}
	if strings.Contains(combined+err.Error(), invalidValue) {
		t.Fatalf("invalid Freshchat config value leaked: error=%v stdout=%s stderr=%s", err, invalidOut.String(), invalidErr.String())
	}
	if strings.Contains(combined+err.Error(), "missing --credential") {
		t.Fatalf("invalid Freshchat config reached credential resolution: error=%v stdout=%s stderr=%s", err, invalidOut.String(), invalidErr.String())
	}
}

func TestConnectorCommandInputDefectsFailBeforeWithApp(t *testing.T) {
	minimum := connectors.ExactNumber("2")
	bundle := engine.Bundle{
		Name:    "input-ordering",
		Streams: []engine.StreamSpec{{Name: "widgets", Method: "GET", Path: "/widgets"}},
		CLISurface: &engine.CLISurface{
			Tagline: "Input ordering fixture", Usage: "pm input-ordering widgets list",
			Commands: []engine.CLICommand{{
				Path: "widgets list", Summary: "List widgets", Intent: "etl", Availability: "implemented", Stream: "widgets",
				APISurface: []engine.CLISurfaceEndpointRef{{Method: "GET", Path: "/widgets"}},
				Flags: []engine.CLIFlag{
					{Name: "state", Type: "enum", Values: []string{"open", "closed"}, MapsTo: "query.state", Required: true},
					{Name: "batch", Type: "integer", MapsTo: "query.batch", Minimum: &minimum},
					{Name: "secret-input", Type: "string", MapsTo: "query.secret_input", EnvOnly: true},
				},
			}},
		},
	}
	registry := connectors.NewEmptyRegistry()
	registry.Register(engine.New(bundle, nil))
	tests := []struct {
		name string
		args []string
		want string
	}{
		{name: "missing required", args: []string{"widgets", "list"}, want: "missing required flag --state"},
		{name: "unknown", args: []string{"widgets", "list", "--state", "open", "--other", "value"}, want: "unknown flag --other"},
		{name: "enum", args: []string{"widgets", "list", "--state", "pending"}, want: "want one of"},
		{name: "minimum", args: []string{"widgets", "list", "--state", "open", "--batch", "1"}, want: "at least 2"},
		{name: "env only direct carrier", args: []string{"widgets", "list", "--state", "open", "--secret-input", "must-not-enter-argv"}, want: "--secret-input must be supplied through --from-env"},
	}
	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			var stdout, stderr bytes.Buffer
			err := runMaybeConnectorCommandWithRegistry(context.Background(), t.TempDir(), "input-ordering", testCase.args, &stdout, &stderr, true, registry)
			if err == nil || !strings.Contains(err.Error(), testCase.want) {
				t.Fatalf("input defect error=%v stdout=%s stderr=%s; want %q before app open", err, stdout.String(), stderr.String(), testCase.want)
			}
		})
	}

	var unknownOut, unknownErr bytes.Buffer
	err := runMaybeConnectorCommandWithRegistry(context.Background(), t.TempDir(), "input-ordering", []string{"widgets", "bogus"}, &unknownOut, &unknownErr, true, registry)
	var classified *cliError
	if !errors.As(err, &classified) || classified.category != categoryUsage {
		t.Fatalf("unknown command error=%v, want usage classification before app open", err)
	}

	var helpOut, helpErr bytes.Buffer
	if err := runMaybeConnectorCommandWithRegistry(context.Background(), t.TempDir(), "input-ordering", []string{"widgets", "list", "--help"}, &helpOut, &helpErr, false, registry); err != nil {
		t.Fatalf("help should return before request validation: %v", err)
	}
}
