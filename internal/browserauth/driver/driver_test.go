package driver_test

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"polymetrics.ai/internal/browserauth/driver"
)

func TestResolveUsesUserSpecifiedPath(t *testing.T) {
	bin := writeExecutableStub(t)
	res, err := driver.Resolve(driver.Config{
		BrowserPath:     bin,
		RequiredCookies: []string{"session"},
	})
	if err != nil {
		t.Fatalf("Resolve() error = %v", err)
	}
	if res.Source != "user_specified" {
		t.Fatalf("Source = %q, want user_specified", res.Source)
	}
	if res.Path != bin {
		t.Fatalf("Path = %q, want %q", res.Path, bin)
	}
}

func TestResolveRejectsMissingUserSpecifiedPath(t *testing.T) {
	_, err := driver.Resolve(driver.Config{
		BrowserPath:     filepath.Join(t.TempDir(), "does-not-exist"),
		RequiredCookies: []string{"session"},
	})
	if err == nil {
		t.Fatalf("Resolve() with missing browser_path: want error, got nil")
	}
}

func TestResolveNeverDownloadsWithoutExplicitOptIn(t *testing.T) {
	// Force the "nothing installed" branch regardless of what this test
	// machine actually has, and give no BrowserPath, without AllowDownload —
	// Resolve must refuse to download and say so, never fetch silently.
	restore := driver.SetLookInstalledBrowserForTest(func() (string, bool) { return "", false })
	defer restore()

	_, err := driver.Resolve(driver.Config{
		RequiredCookies: []string{"session"},
		AllowDownload:   false,
	})
	if err == nil {
		t.Fatalf("Resolve() with no browser and AllowDownload=false: want error, got nil")
	}
	if !strings.Contains(err.Error(), "AllowDownload") {
		t.Fatalf("Resolve() error = %v, want it to name AllowDownload", err)
	}
}

func TestResolveUsesInstalledBrowserWhenNoPathConfigured(t *testing.T) {
	stub := writeExecutableStub(t)
	restore := driver.SetLookInstalledBrowserForTest(func() (string, bool) { return stub, true })
	defer restore()

	res, err := driver.Resolve(driver.Config{RequiredCookies: []string{"session"}})
	if err != nil {
		t.Fatalf("Resolve() error = %v", err)
	}
	if res.Source != "installed" {
		t.Fatalf("Source = %q, want installed", res.Source)
	}
	if res.Path != stub {
		t.Fatalf("Path = %q, want %q", res.Path, stub)
	}
}

func TestConfigRejectsEmptyRequiredCookies(t *testing.T) {
	_, err := driver.New(t.Context(), driver.Config{LoginURL: "https://example.invalid/login"})
	if err == nil {
		t.Fatalf("New() with no required_cookies: want error, got nil")
	}
	if !strings.Contains(err.Error(), "required_cookies") {
		t.Fatalf("New() error = %v, want it to name required_cookies", err)
	}
}

// writeExecutableStub writes a harmless placeholder file (never executed by
// this test — Resolve() only stat()s BrowserPath, it does not launch
// anything) and marks it executable, standing in for a real browser binary
// path.
func writeExecutableStub(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	name := "chrome-stub"
	if runtime.GOOS == "windows" {
		name += ".exe"
	}
	path := filepath.Join(dir, name)
	if err := os.WriteFile(path, []byte("#!/bin/sh\nexit 0\n"), 0o755); err != nil {
		t.Fatalf("write stub: %v", err)
	}
	return path
}
