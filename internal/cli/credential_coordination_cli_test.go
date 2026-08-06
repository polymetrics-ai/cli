package cli_test

import (
	"bytes"
	"context"
	"strings"
	"testing"

	"polymetrics.ai/internal/app"
	"polymetrics.ai/internal/cli"
)

func TestCredentialsCoordinationLinkCLIAndHelp(t *testing.T) {
	root := t.TempDir()
	var stdout, stderr bytes.Buffer
	if code := cli.Run([]string{"init", "--root", root}, &stdout, &stderr); code != 0 {
		t.Fatalf("init code = %d stderr = %s", code, stderr.String())
	}

	stdout.Reset()
	stderr.Reset()
	code := cli.Run([]string{
		"credentials", "add", "sample-shared",
		"--connector", "sample",
		"--provider-family", "provider-fixture",
		"--auth-profile", "service-profile",
		"--root", root,
		"--json",
	}, &stdout, &stderr)
	if code != 0 {
		t.Fatalf("credentials add code = %d stdout = %s stderr = %s", code, stdout.String(), stderr.String())
	}
	if !strings.Contains(stdout.String(), `"provider_family": "provider-fixture"`) || !strings.Contains(stdout.String(), `"auth_profile": "service-profile"`) {
		t.Fatalf("credential JSON did not expose safe declared coordination metadata: %s", stdout.String())
	}
	if strings.Contains(stdout.String(), "binding") || strings.Contains(stdout.String(), "auth_cohort") || strings.Contains(stdout.String(), "rate_scope") {
		t.Fatal("credential JSON exposed protected coordination identity material")
	}

	stdout.Reset()
	stderr.Reset()
	code = cli.Run([]string{
		"credentials", "add", "faker-shared",
		"--connector", "faker",
		"--provider-family", "provider-fixture",
		"--auth-profile", "service-profile",
		"--root", root,
	}, &stdout, &stderr)
	if code != 0 {
		t.Fatalf("second credentials add code = %d stderr = %s", code, stderr.String())
	}

	stdout.Reset()
	stderr.Reset()
	code = cli.Run([]string{"credentials", "link", "faker-shared", "--to", "sample-shared", "--root", root, "--json"}, &stdout, &stderr)
	if code != 0 {
		t.Fatalf("credentials link code = %d stdout = %s stderr = %s", code, stdout.String(), stderr.String())
	}
	if strings.Contains(stdout.String(), "binding") || strings.Contains(stdout.String(), "auth_cohort") || strings.Contains(stdout.String(), "rate_scope") {
		t.Fatal("credentials link JSON exposed protected coordination identity material")
	}

	for _, args := range [][]string{
		{"help", "credentials"},
		{"credentials", "--help"},
		{"credentials"},
	} {
		stdout.Reset()
		stderr.Reset()
		code = cli.Run(args, &stdout, &stderr)
		if code != 0 {
			t.Fatalf("Run(%v) code = %d stdout = %s stderr = %s", args, code, stdout.String(), stderr.String())
		}
		for _, want := range []string{"--provider-family", "--auth-profile", "credentials link"} {
			if !strings.Contains(stdout.String(), want) {
				t.Fatalf("Run(%v) help missing %q", args, want)
			}
		}
	}
}

func TestCredentialsCoordinationLinkRejectsIncompatibleProfileWithoutEchoingValues(t *testing.T) {
	root := t.TempDir()
	var stdout, stderr bytes.Buffer
	if code := cli.Run([]string{"init", "--root", root}, &stdout, &stderr); code != 0 {
		t.Fatalf("init code = %d stderr = %s", code, stderr.String())
	}
	for _, args := range [][]string{
		{"credentials", "add", "sample-shared", "--connector", "sample", "--provider-family", "provider-fixture", "--auth-profile", "service-profile", "--root", root},
		{"credentials", "add", "faker-incompatible", "--connector", "faker", "--provider-family", "other-provider", "--auth-profile", "service-profile", "--root", root},
	} {
		stdout.Reset()
		stderr.Reset()
		if code := cli.Run(args, &stdout, &stderr); code != 0 {
			t.Fatalf("Run(%v) code = %d stderr = %s", args, code, stderr.String())
		}
	}

	stdout.Reset()
	stderr.Reset()
	code := cli.Run([]string{"credentials", "link", "faker-incompatible", "--to", "sample-shared", "--root", root, "--json"}, &stdout, &stderr)
	if code != 3 {
		t.Fatalf("credentials link code = %d, want validation exit 3; stdout = %s stderr = %s", code, stdout.String(), stderr.String())
	}
	combined := stdout.String() + stderr.String()
	if !strings.Contains(combined, "provider family") {
		t.Fatal("incompatible link error did not name the failed field")
	}
	if strings.Contains(combined, "provider-fixture") || strings.Contains(combined, "other-provider") {
		t.Fatal("incompatible link error echoed a declared metadata value")
	}
	if !strings.Contains(stdout.String(), `"category": "validation"`) {
		t.Fatalf("incompatible link did not return validation category: %s", stdout.String())
	}

	stdout.Reset()
	stderr.Reset()
	code = cli.Run([]string{"credentials", "add", "faker-incompatible-on-create", "--connector", "faker", "--provider-family", "other-provider", "--auth-profile", "service-profile", "--link-credential", "sample-shared", "--root", root, "--json"}, &stdout, &stderr)
	if code != 3 {
		t.Fatalf("credentials add with incompatible link code = %d, want validation exit 3; stdout = %s stderr = %s", code, stdout.String(), stderr.String())
	}
	combined = stdout.String() + stderr.String()
	if !strings.Contains(combined, "provider family") {
		t.Fatal("incompatible add-time link error did not name the failed field")
	}
	if strings.Contains(combined, "provider-fixture") || strings.Contains(combined, "other-provider") {
		t.Fatal("incompatible add-time link error echoed a declared metadata value")
	}
	if !strings.Contains(stdout.String(), `"category": "validation"`) {
		t.Fatalf("incompatible add-time link did not return validation category: %s", stdout.String())
	}
}

func TestCredentialsCoordinationInputErrorsUseDocumentedCategories(t *testing.T) {
	root := t.TempDir()
	var stdout, stderr bytes.Buffer
	if code := cli.Run([]string{"init", "--root", root}, &stdout, &stderr); code != 0 {
		t.Fatalf("init code = %d stderr = %s", code, stderr.String())
	}

	for _, test := range []struct {
		name     string
		args     []string
		wantCode int
		category string
	}{
		{
			name:     "invalid provider family",
			args:     []string{"credentials", "add", "invalid-provider-family", "--connector", "sample", "--provider-family", "invalid family", "--root", root, "--json"},
			wantCode: 3,
			category: "validation",
		},
		{
			name:     "invalid auth profile",
			args:     []string{"credentials", "add", "invalid-auth-profile", "--connector", "sample", "--auth-profile", "invalid profile", "--root", root, "--json"},
			wantCode: 3,
			category: "validation",
		},
		{
			name:     "missing link target",
			args:     []string{"credentials", "link", "sample-local", "--root", root, "--json"},
			wantCode: 2,
			category: "usage",
		},
		{
			name:     "bare provider family",
			args:     []string{"credentials", "add", "bare-provider-family", "--connector", "sample", "--provider-family", "--root", root, "--json"},
			wantCode: 2,
			category: "usage",
		},
		{
			name:     "bare auth profile",
			args:     []string{"credentials", "add", "bare-auth-profile", "--connector", "sample", "--auth-profile", "--root", root, "--json"},
			wantCode: 2,
			category: "usage",
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			stdout.Reset()
			stderr.Reset()
			code := cli.Run(test.args, &stdout, &stderr)
			if code != test.wantCode {
				t.Fatalf("Run(%v) code = %d, want %d; stdout = %s stderr = %s", test.args, code, test.wantCode, stdout.String(), stderr.String())
			}
			if !strings.Contains(stdout.String(), `"category": "`+test.category+`"`) {
				t.Fatalf("Run(%v) did not return %s error category: %s", test.args, test.category, stdout.String())
			}
		})
	}
	stdout.Reset()
	stderr.Reset()
	if code := cli.Run([]string{"credentials", "list", "--root", root, "--json"}, &stdout, &stderr); code != 0 {
		t.Fatalf("credentials list code = %d stdout = %s stderr = %s", code, stdout.String(), stderr.String())
	}
	for _, name := range []string{"bare-provider-family", "bare-auth-profile"} {
		if strings.Contains(stdout.String(), name) {
			t.Fatalf("bare declaration flag persisted credential %q: %s", name, stdout.String())
		}
	}
}

func TestCredentialsCoordinationBareLinkTargetsRequireCredentialIdentifiers(t *testing.T) {
	root := t.TempDir()
	var stdout, stderr bytes.Buffer
	if code := cli.Run([]string{"init", "--root", root}, &stdout, &stderr); code != 0 {
		t.Fatalf("init code = %d stderr = %s", code, stderr.String())
	}
	for _, args := range [][]string{
		{"credentials", "add", "true", "--connector", "sample", "--root", root},
		{"credentials", "add", "source", "--connector", "sample", "--root", root},
	} {
		stdout.Reset()
		stderr.Reset()
		if code := cli.Run(args, &stdout, &stderr); code != 0 {
			t.Fatalf("Run(%v) code = %d stdout = %s stderr = %s", args, code, stdout.String(), stderr.String())
		}
	}

	for _, test := range []struct {
		name string
		args []string
		flag string
	}{
		{
			name: "add time target",
			args: []string{"credentials", "add", "bare-link-on-create", "--connector", "sample", "--link-credential", "--root", root, "--json"},
			flag: "--link-credential",
		},
		{
			name: "empty equals target",
			args: []string{"credentials", "add", "empty-link-on-create", "--connector", "sample", "--link-credential=", "--root", root, "--json"},
			flag: "--link-credential",
		},
		{
			name: "whitespace equals target",
			args: []string{"credentials", "add", "whitespace-link-on-create", "--connector", "sample", "--link-credential=   ", "--root", root, "--json"},
			flag: "--link-credential",
		},
		{
			name: "existing target",
			args: []string{"credentials", "link", "source", "--to", "--root", root, "--json"},
			flag: "--to",
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			stdout.Reset()
			stderr.Reset()
			if code := cli.Run(test.args, &stdout, &stderr); code != 2 {
				t.Fatalf("Run(%v) code = %d, want usage exit 2; stdout = %s stderr = %s", test.args, code, stdout.String(), stderr.String())
			}
			combined := stdout.String() + stderr.String()
			if !strings.Contains(combined, test.flag+" requires a credential identifier") {
				t.Fatalf("Run(%v) did not identify the missing credential target: %s", test.args, combined)
			}
		})
	}

	stdout.Reset()
	stderr.Reset()
	if code := cli.Run([]string{"credentials", "list", "--root", root, "--json"}, &stdout, &stderr); code != 0 {
		t.Fatalf("credentials list code = %d stdout = %s stderr = %s", code, stdout.String(), stderr.String())
	}
	for _, name := range []string{"bare-link-on-create", "empty-link-on-create", "whitespace-link-on-create"} {
		if strings.Contains(stdout.String(), name) {
			t.Fatalf("invalid --link-credential persisted credential %q: %s", name, stdout.String())
		}
	}

	instance, err := app.Open(root)
	if err != nil {
		t.Fatalf("Open() error = %v", err)
	}
	_, sourceRuntime, err := instance.ResolveConnectorCredential(context.Background(), "sample", "source", nil)
	if err != nil {
		t.Fatalf("ResolveConnectorCredential(source) error = %v", err)
	}
	_, trueRuntime, err := instance.ResolveConnectorCredential(context.Background(), "sample", "true", nil)
	if err != nil {
		t.Fatalf("ResolveConnectorCredential(true) error = %v", err)
	}
	if sourceRuntime.CoordinationIdentity.AuthCohortKey() == trueRuntime.CoordinationIdentity.AuthCohortKey() {
		t.Fatal("bare --to linked the source to the parser-generated true credential")
	}
}
