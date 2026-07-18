package cli

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/spf13/cobra"

	"polymetrics.ai/internal/app"
	"polymetrics.ai/internal/config"
)

// This compile-time reference is intentional TDD evidence: the focused test
// checkpoint must fail until credentials has a native Cobra constructor.
var _ = newCredentialsCobraCommand

const (
	opaqueEnvFixture   = "opaque-redaction-fixture-env-429"
	opaqueStdinFixture = "opaque-redaction-fixture-stdin-429"
)

func TestCredentialsCommandIsNativeCobraSubtree(t *testing.T) {
	root := newRootCmd(context.Background(), testRouterConfig(".", false), io.Discard, io.Discard)
	credentials := findCobraCommand(root, "credentials")
	if credentials == nil {
		t.Fatal("missing credentials command")
	}
	if credentials.DisableFlagParsing {
		t.Fatal("credentials command must use native Cobra flag parsing")
	}
	if credentials.ValidArgsFunction == nil {
		t.Fatal("credentials command must suppress file completion until Phase 15")
	}
	values, directive := credentials.ValidArgsFunction(credentials, nil, "")
	if len(values) != 0 || directive != cobra.ShellCompDirectiveNoFileComp {
		t.Fatalf("credentials completion = (%v, %v), want no values and NoFileComp", values, directive)
	}

	for _, name := range []string{"add", "list", "inspect", "test", "remove"} {
		t.Run(name, func(t *testing.T) {
			action := findCobraCommand(credentials, name)
			if action == nil {
				t.Fatalf("missing credentials %s command", name)
			}
			if action.DisableFlagParsing {
				t.Fatalf("credentials %s must use native Cobra flag parsing", name)
			}
			if !action.FParseErrWhitelist.UnknownFlags {
				t.Fatalf("credentials %s must preserve unknown-flag tolerance", name)
			}
			if action.ValidArgsFunction == nil {
				t.Fatalf("credentials %s missing no-file completion seam", name)
			}
		})
	}

	add := findCobraCommand(credentials, "add")
	for _, name := range []string{"connector", "from-env", "value-stdin", "config"} {
		t.Run("add flag "+name, func(t *testing.T) {
			flag := add.Flags().Lookup(name)
			if flag == nil {
				t.Fatalf("credentials add missing --%s", name)
			}
			if got, want := flag.Value.Type(), "stringArray"; got != want {
				t.Fatalf("credentials add --%s type = %q, want %q", name, got, want)
			}
			if got, want := flag.NoOptDefVal, "true"; got != want {
				t.Fatalf("credentials add --%s NoOptDefVal = %q, want %q", name, got, want)
			}
		})
	}

	if add.Flags().Lookup("interactive") != nil {
		t.Fatal("credentials add must not expose interactive secret entry")
	}

	help := findCobraCommand(credentials, "help")
	if help == nil || !help.Hidden {
		t.Fatal("credentials must preserve hidden positional help until Phase 19")
	}
}

func TestCredentialsAddListRemovePreserveCurrentFlagForms(t *testing.T) {
	root := initCredentialsProject(t)

	stdout, err := executeCredentialsCommand(t, root, false, strings.NewReader(""),
		"credentials", "add", "config-only",
		"--connector", "not-selected",
		"--connector=sample",
		"--config", "mode=first",
		"--config=mode=fixture",
		"--unknown=ignored",
		"extra-positional",
	)
	if err != nil {
		t.Fatalf("credentials add error = %v", err)
	}
	if !strings.Contains(stdout, "Saved credential config-only for connector sample") {
		t.Fatalf("credentials add text output mismatch: %q", stdout)
	}

	stdout, err = executeCredentialsCommand(t, root, false, strings.NewReader(""),
		"credentials", "list", "--unknown=ignored", "extra-positional")
	if err != nil {
		t.Fatalf("credentials list error = %v", err)
	}
	if !strings.Contains(stdout, "config-only") || !strings.Contains(stdout, "sample") {
		t.Fatalf("credentials list text output missing metadata: %q", stdout)
	}

	stdout, err = executeCredentialsCommand(t, root, false, strings.NewReader(""),
		"credentials", "remove", "config-only", "--unknown=ignored", "extra-positional")
	if err != nil {
		t.Fatalf("credentials remove error = %v", err)
	}
	if stdout != "Removed credential config-only\n" {
		t.Fatalf("credentials remove output = %q", stdout)
	}

	stdout, err = executeCredentialsCommand(t, root, true, strings.NewReader(""), "credentials", "list")
	if err != nil {
		t.Fatalf("empty credentials list error = %v", err)
	}
	if !strings.Contains(stdout, `"kind": "CredentialList"`) || !strings.Contains(stdout, `"credentials": null`) {
		t.Fatalf("empty credentials list JSON mismatch: %q", stdout)
	}
}

func TestCredentialsAddBareFlagCompatibility(t *testing.T) {
	t.Run("connector", func(t *testing.T) {
		root := initCredentialsProject(t)
		_, err := executeCredentialsCommand(t, root, false, strings.NewReader(""),
			"credentials", "add", "bare-connector", "--connector")
		if err == nil || !strings.Contains(err.Error(), `connector "true" not found`) {
			t.Fatalf("bare --connector error = %v", err)
		}
	})

	t.Run("from env", func(t *testing.T) {
		root := initCredentialsProject(t)
		_, err := executeCredentialsCommand(t, root, false, strings.NewReader(""),
			"credentials", "add", "bare-env", "--connector=sample", "--from-env")
		if err == nil || !strings.Contains(err.Error(), "invalid --from-env") {
			t.Fatalf("bare --from-env error = %v", err)
		}
	})

	t.Run("value stdin", func(t *testing.T) {
		root := initCredentialsProject(t)
		stdout, err := executeCredentialsCommand(t, root, true, strings.NewReader(opaqueStdinFixture+"\r\n"),
			"credentials", "add", "bare-stdin", "--connector=sample", "--value-stdin")
		if err != nil {
			t.Fatalf("bare --value-stdin error = %v", err)
		}
		if !strings.Contains(stdout, `"secret_fields": [`) || !strings.Contains(stdout, `"true"`) {
			t.Fatalf("bare --value-stdin did not select legacy field name")
		}
		assertOpaqueFixtureAbsent(t, stdout, "")
		assertStateDoesNotContainFixtures(t, root)
	})

	t.Run("config", func(t *testing.T) {
		root := initCredentialsProject(t)
		_, err := executeCredentialsCommand(t, root, false, strings.NewReader(""),
			"credentials", "add", "bare-config", "--connector=sample", "--config")
		if err == nil || !strings.Contains(err.Error(), "invalid key-value") {
			t.Fatalf("bare --config error = %v", err)
		}
	})
}

func TestCredentialsSecretSourcesUseOnlyNamedEnvironmentAndControlledStdin(t *testing.T) {
	t.Setenv("PM_TEST_CREDENTIAL_A", opaqueEnvFixture)
	t.Setenv("PM_TEST_CREDENTIAL_B", opaqueEnvFixture)
	root := initCredentialsProject(t)

	stdout, err := executeCredentialsCommand(t, root, true, strings.NewReader(opaqueStdinFixture+"\n"),
		"credentials", "add", "source-selection",
		"--connector", "sample",
		"--from-env", "env_a=PM_TEST_CREDENTIAL_A",
		"--from-env=env_b=PM_TEST_CREDENTIAL_B",
		"--value-stdin", "ignored_stdin_field",
		"--value-stdin=stdin_field",
	)
	if err != nil {
		t.Fatalf("credentials source selection error = %v", err)
	}
	for _, field := range []string{`"env_a"`, `"env_b"`, `"stdin_field"`} {
		if !strings.Contains(stdout, field) {
			t.Fatalf("credential metadata missing expected secret field %s", field)
		}
	}
	if strings.Contains(stdout, "ignored_stdin_field") {
		t.Fatal("repeated --value-stdin did not preserve final-value selection")
	}
	assertOpaqueFixtureAbsent(t, stdout, "")
	assertStateDoesNotContainFixtures(t, root)
}

func TestCredentialsSecretSourceValidationFailsClosed(t *testing.T) {
	t.Setenv("PM_TEST_CREDENTIAL_EMPTY", "")
	tests := []struct {
		name string
		args []string
		want string
	}{
		{name: "malformed mapping", args: []string{"--from-env=field"}, want: "invalid --from-env"},
		{name: "unsafe field", args: []string{"--from-env=../field=PM_TEST_CREDENTIAL_EMPTY"}, want: "secret field"},
		{name: "unsafe environment name", args: []string{"--from-env=field=BAD/ENV"}, want: "environment variable"},
		{name: "empty environment", args: []string{"--from-env=field=PM_TEST_CREDENTIAL_EMPTY"}, want: "is empty"},
		{name: "unsafe stdin field", args: []string{"--value-stdin=../field"}, want: "secret field"},
		{name: "unsafe config key", args: []string{"--config=../path=value"}, want: "config key"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			root := initCredentialsProject(t)
			args := append([]string{"credentials", "add", "invalid-source", "--connector=sample"}, tt.args...)
			_, err := executeCredentialsCommand(t, root, false, strings.NewReader(opaqueStdinFixture), args...)
			if err == nil || !strings.Contains(err.Error(), tt.want) {
				t.Fatalf("error = %v, want text %q", err, tt.want)
			}
			assertOpaqueFixtureAbsent(t, "", err.Error())
			assertStateDoesNotContainFixtures(t, root)
		})
	}
}

func TestCredentialsStrictNamesAndPathContainment(t *testing.T) {
	nameTests := []struct {
		name string
		args []string
	}{
		{name: "add credential traversal", args: []string{"add", "../credential", "--connector=sample"}},
		{name: "add credential leading hyphen", args: []string{"add", "-credential", "--connector=sample"}},
		{name: "add credential leading underscore", args: []string{"add", "_credential", "--connector=sample"}},
		{name: "add connector traversal", args: []string{"add", "credential", "--connector=../sample"}},
		{name: "add connector leading hyphen", args: []string{"add", "credential", "--connector=-sample"}},
		{name: "inspect traversal", args: []string{"inspect", "../credential"}},
		{name: "test traversal", args: []string{"test", "../credential"}},
		{name: "remove traversal", args: []string{"remove", "../credential"}},
	}
	for _, tt := range nameTests {
		t.Run(tt.name, func(t *testing.T) {
			root := initCredentialsProject(t)
			_, err := executeCredentialsCommand(t, root, false, strings.NewReader(""), append([]string{"credentials"}, tt.args...)...)
			if err == nil {
				t.Fatal("unsafe identifier succeeded")
			}
			classified := classifyError(mapCobraErr(err))
			if classified.category != categoryValidation {
				t.Fatalf("unsafe identifier category = %s, want validation", classified.category)
			}
		})
	}

	t.Run("warehouse traversal denied", func(t *testing.T) {
		root := initCredentialsProject(t)
		_, err := executeCredentialsCommand(t, root, false, strings.NewReader(""),
			"credentials", "add", "warehouse-traversal", "--connector=warehouse", "--config=path=../outside")
		if err == nil || !strings.Contains(err.Error(), "must not escape the project root") {
			t.Fatalf("warehouse traversal error = %v", err)
		}
	})

	t.Run("warehouse external path requires explicit opt in", func(t *testing.T) {
		root := initCredentialsProject(t)
		external := filepath.Join(t.TempDir(), "warehouse")
		_, err := executeCredentialsCommand(t, root, false, strings.NewReader(""),
			"credentials", "add", "warehouse-denied", "--connector=warehouse", "--config=path="+external)
		if err == nil || !strings.Contains(err.Error(), "requires allow_external_path=true") {
			t.Fatalf("external warehouse path error = %v", err)
		}
		_, err = executeCredentialsCommand(t, root, false, strings.NewReader(""),
			"credentials", "add", "warehouse-allowed", "--connector=warehouse", "--config=path="+external, "--config=allow_external_path=true")
		if err != nil {
			t.Fatalf("explicit external warehouse path error = %v", err)
		}
	})
}

func TestCredentialsOutputsAndErrorsNeverExposeOpaqueSecretFixtures(t *testing.T) {
	t.Setenv("PM_TEST_CREDENTIAL_REDACTION", opaqueEnvFixture)
	root := initCredentialsProject(t)

	var stdout, stderr bytes.Buffer
	code := Run([]string{
		"credentials", "add", "redacted-env", "--connector=sample",
		"--from-env=token=PM_TEST_CREDENTIAL_REDACTION", "--root", root, "--json",
	}, &stdout, &stderr)
	if code != 0 {
		t.Fatal("credentials add with named environment failed")
	}
	assertOpaqueFixtureAbsent(t, stdout.String(), stderr.String())
	assertStateDoesNotContainFixtures(t, root)
	assertProjectFilesDoNotContainFixtures(t, root)

	stdout.Reset()
	stderr.Reset()
	code = Run([]string{"credentials", "inspect", "redacted-env", "--root", root, "--json"}, &stdout, &stderr)
	if code != 0 {
		t.Fatal("credentials inspect of redacted metadata failed")
	}
	assertOpaqueFixtureAbsent(t, stdout.String(), stderr.String())

	stdout.Reset()
	stderr.Reset()
	code = Run([]string{
		"credentials", "add", "redacted-error", "--connector=sample",
		"--from-env=token=PM_TEST_CREDENTIAL_REDACTION", "--config=malformed", "--root", root, "--json",
	}, &stdout, &stderr)
	if code == 0 {
		t.Fatal("malformed config unexpectedly succeeded")
	}
	assertOpaqueFixtureAbsent(t, stdout.String(), stderr.String())
	assertStateDoesNotContainFixtures(t, root)

	stdoutText, err := executeCredentialsCommand(t, root, true, strings.NewReader(opaqueStdinFixture),
		"credentials", "add", "redacted-stdin-error", "--connector=sample", "--value-stdin=token", "--config=malformed")
	if err == nil {
		t.Fatal("stdin redaction error case unexpectedly succeeded")
	}
	assertOpaqueFixtureAbsent(t, stdoutText, err.Error())
	assertStateDoesNotContainFixtures(t, root)
}

func TestCredentialsHelpRoutesPreserveCanonicalManual(t *testing.T) {
	tests := []struct {
		name string
		args []string
		json bool
	}{
		{name: "bare", args: []string{"credentials"}},
		{name: "help topic", args: []string{"help", "credentials"}},
		{name: "long help", args: []string{"credentials", "--help"}},
		{name: "short help", args: []string{"credentials", "-h"}},
		{name: "positional help", args: []string{"credentials", "help"}},
		{name: "json bare", args: []string{"credentials", "--json"}, json: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var stdout, stderr bytes.Buffer
			if code := Run(tt.args, &stdout, &stderr); code != 0 {
				t.Fatalf("help code = %d stderr=%q", code, stderr.String())
			}
			if tt.json {
				if !strings.Contains(stdout.String(), `"kind": "CommandManual"`) || !strings.Contains(stdout.String(), `"command": "credentials"`) {
					t.Fatalf("JSON manual mismatch: %q", stdout.String())
				}
				return
			}
			for _, want := range []string{"NAME", "SYNOPSIS", "SECURITY", "pm credentials - manage encrypted connector credentials"} {
				if !strings.Contains(stdout.String(), want) {
					t.Fatalf("manual missing %q", want)
				}
			}
		})
	}
}

func TestCredentialsActionTailHelpAndLiteralSeparatorRemainLegacyInputs(t *testing.T) {
	tests := []struct {
		name string
		tail []string
	}{
		{name: "long help", tail: []string{"--help"}},
		{name: "assigned help", tail: []string{"--help=false"}},
		{name: "short help", tail: []string{"-h"}},
		{name: "short cluster", tail: []string{"-hx"}},
		{name: "literal separator", tail: []string{"--"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			root := initCredentialsProject(t)
			args := []string{"credentials", "add", "tail-compatible"}
			args = append(args, tt.tail...)
			args = append(args, "--connector=sample")
			stdout, err := executeCredentialsCommand(t, root, false, strings.NewReader(""), args...)
			if err != nil {
				t.Fatalf("tail-compatible add error = %v", err)
			}
			if strings.Contains(stdout, "SYNOPSIS") || !strings.Contains(stdout, "Saved credential") {
				t.Fatalf("tail token triggered help or suppressed action: %q", stdout)
			}
		})
	}
}

func TestCredentialsLeadingInvalidTokensCannotDiscoverActions(t *testing.T) {
	leading := [][]string{
		{"--unknown=x"},
		{"--unknown"},
		{"-x"},
		{"--help=false"},
		{"--"},
	}
	for _, head := range leading {
		name := strings.NewReplacer("-", "dash", "=", "eq").Replace(strings.Join(head, "_"))
		t.Run("add_"+name, func(t *testing.T) {
			root := initCredentialsProject(t)
			args := append([]string{"credentials"}, head...)
			args = append(args, "add", "must-not-exist", "--connector=sample")
			var stdout, stderr bytes.Buffer
			code := Run(append(args, "--root", root, "--json"), &stdout, &stderr)
			if code != 2 {
				t.Fatalf("leading invalid add code = %d, want usage 2", code)
			}
			if got := credentialCount(t, root); got != 0 {
				t.Fatalf("leading invalid token executed add; credential count = %d", got)
			}
		})

		t.Run("remove_"+name, func(t *testing.T) {
			root := initCredentialsProject(t)
			if _, err := executeCredentialsCommand(t, root, false, strings.NewReader(""),
				"credentials", "add", "must-remain", "--connector=sample"); err != nil {
				t.Fatalf("seed credential: %v", err)
			}
			args := append([]string{"credentials"}, head...)
			args = append(args, "remove", "must-remain")
			var stdout, stderr bytes.Buffer
			code := Run(append(args, "--root", root, "--json"), &stdout, &stderr)
			if code != 2 {
				t.Fatalf("leading invalid remove code = %d, want usage 2", code)
			}
			if got := credentialCount(t, root); got != 1 {
				t.Fatalf("leading invalid token executed remove; credential count = %d", got)
			}
		})
	}
}

func TestCredentialsLeadingInvalidNameTokensCannotDiscoverLaterNames(t *testing.T) {
	leading := [][]string{
		{"--unknown=x"},
		{"--unknown"},
		{"-x"},
		{"--help=false"},
		{"--"},
	}
	for _, head := range leading {
		name := strings.NewReplacer("-", "dash", "=", "eq").Replace(strings.Join(head, "_"))
		t.Run("add_"+name, func(t *testing.T) {
			root := initCredentialsProject(t)
			args := []string{"credentials", "add"}
			args = append(args, head...)
			args = append(args, "must-not-exist", "--connector=sample")
			var stdout, stderr bytes.Buffer
			code := Run(append(args, "--root", root, "--json"), &stdout, &stderr)
			if code == 0 {
				t.Fatal("invalid leading credential-name token discovered and executed later add name")
			}
			if got := credentialCount(t, root); got != 0 {
				t.Fatalf("invalid leading credential-name token executed add; credential count = %d", got)
			}
		})

		t.Run("remove_"+name, func(t *testing.T) {
			root := initCredentialsProject(t)
			if _, err := executeCredentialsCommand(t, root, false, strings.NewReader(""),
				"credentials", "add", "must-remain", "--connector=sample"); err != nil {
				t.Fatalf("seed credential: %v", err)
			}
			args := []string{"credentials", "remove"}
			args = append(args, head...)
			args = append(args, "must-remain")
			var stdout, stderr bytes.Buffer
			code := Run(append(args, "--root", root, "--json"), &stdout, &stderr)
			if code == 0 {
				t.Fatal("invalid leading credential-name token discovered and executed later remove name")
			}
			if got := credentialCount(t, root); got != 1 {
				t.Fatalf("invalid leading credential-name token executed remove; credential count = %d", got)
			}
		})
	}
}

func TestCredentialsInvalidActionsAndGlobalBooleans(t *testing.T) {
	root := initCredentialsProject(t)
	for _, args := range [][]string{
		{"credentials", "bogus", "--root", root},
		{"credentials", "add", "--root", root},
		{"credentials", "remove", "--root", root},
	} {
		var stdout, stderr bytes.Buffer
		if code := Run(args, &stdout, &stderr); code != 2 {
			t.Fatalf("Run(%v) code = %d, want usage 2", args, code)
		}
	}

	tests := []struct {
		name     string
		args     []string
		wantCode int
		wantJSON bool
	}{
		{name: "json true", args: []string{"credentials", "list", "--json=true", "--plain=false", "--no-input=true", "--root=" + root}, wantCode: 0, wantJSON: true},
		{name: "json false", args: []string{"--json=false", "credentials", "list", "--root", root}, wantCode: 0},
		{name: "invalid json", args: []string{"credentials", "list", "--json=maybe", "--root", root}, wantCode: 3},
		{name: "invalid plain", args: []string{"credentials", "list", "--plain=maybe", "--root", root}, wantCode: 3},
		{name: "invalid no input", args: []string{"credentials", "list", "--no-input=maybe", "--root", root}, wantCode: 3},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var stdout, stderr bytes.Buffer
			if code := Run(tt.args, &stdout, &stderr); code != tt.wantCode {
				t.Fatalf("code = %d, want %d", code, tt.wantCode)
			}
			if tt.wantJSON && !strings.Contains(stdout.String(), `"kind": "CredentialList"`) {
				t.Fatal("assigned global booleans did not preserve JSON output")
			}
		})
	}
}

func executeCredentialsCommand(t *testing.T, root string, jsonOut bool, stdin io.Reader, args ...string) (string, error) {
	t.Helper()
	var stdout, stderr bytes.Buffer
	cmd := newRootCmd(context.Background(), config.Config{Root: root, JSON: jsonOut}, &stdout, &stderr)
	cmd.SetIn(stdin)
	err := executeRootCmd(cmd, args)
	return stdout.String(), err
}

func initCredentialsProject(t *testing.T) string {
	t.Helper()
	root := t.TempDir()
	if err := app.InitProject(root); err != nil {
		t.Fatalf("InitProject() error = %v", err)
	}
	return root
}

func credentialCount(t *testing.T, root string) int {
	t.Helper()
	a, err := app.Open(root)
	if err != nil {
		t.Fatalf("app.Open() error = %v", err)
	}
	return len(a.ListCredentials())
}

func assertOpaqueFixtureAbsent(t *testing.T, stdout, stderr string) {
	t.Helper()
	if strings.Contains(stdout, opaqueEnvFixture) || strings.Contains(stdout, opaqueStdinFixture) ||
		strings.Contains(stderr, opaqueEnvFixture) || strings.Contains(stderr, opaqueStdinFixture) {
		t.Fatal("opaque credential fixture appeared in command output")
	}
}

func assertProjectFilesDoNotContainFixtures(t *testing.T, root string) {
	t.Helper()
	err := filepath.Walk(filepath.Join(root, ".polymetrics"), func(path string, info os.FileInfo, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if info.IsDir() {
			return nil
		}
		content, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		if bytes.Contains(content, []byte(opaqueEnvFixture)) || bytes.Contains(content, []byte(opaqueStdinFixture)) {
			return errors.New("opaque credential fixture found in project file")
		}
		return nil
	})
	if err != nil {
		t.Fatalf("scan temporary project for plaintext fixture: %v", err)
	}
}

func assertStateDoesNotContainFixtures(t *testing.T, root string) {
	t.Helper()
	state, err := os.ReadFile(filepath.Join(root, ".polymetrics", "state.json"))
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("read state metadata: %v", err)
	}
	if bytes.Contains(state, []byte(opaqueEnvFixture)) || bytes.Contains(state, []byte(opaqueStdinFixture)) {
		t.Fatal("opaque credential fixture appeared in state metadata")
	}
	if len(state) > 0 {
		var decoded any
		if err := json.Unmarshal(state, &decoded); err != nil {
			t.Fatalf("decode state metadata: %v", err)
		}
	}
}
