package cli_test

import (
	"bytes"
	"strings"
	"testing"

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
	if code == 0 {
		t.Fatal("credentials link accepted incompatible provider families")
	}
	combined := stdout.String() + stderr.String()
	if !strings.Contains(combined, "provider family") {
		t.Fatal("incompatible link error did not name the failed field")
	}
	if strings.Contains(combined, "provider-fixture") || strings.Contains(combined, "other-provider") {
		t.Fatal("incompatible link error echoed a declared metadata value")
	}
}
