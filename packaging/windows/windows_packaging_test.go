package windows_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestWixSourceEncodesApprovedInstallerDefaults(t *testing.T) {
	t.Parallel()

	data, err := os.ReadFile("pm.wxs")
	if err != nil {
		t.Fatalf("read pm.wxs: %v", err)
	}
	wxs := string(data)

	for _, want := range []string{
		`Name="Polymetrics CLI"`,
		`Manufacturer="Polymetrics AI"`,
		`Scope="perMachine"`,
		`Id="ProgramFiles64Folder"`,
		`Name="Polymetrics"`,
		`Name="CLI"`,
		`Name="pm.exe"`,
		`Id="AddPmToMachinePath"`,
		`Name="PATH"`,
		`Value="[INSTALLFOLDER]"`,
		`System="yes"`,
		`Permanent="no"`,
		`<MajorUpgrade DowngradeErrorMessage=`,
	} {
		if !strings.Contains(wxs, want) {
			t.Fatalf("pm.wxs missing %q", want)
		}
	}
}

func TestWindowsPackagingReadmeDocumentsStableUpgradeCodes(t *testing.T) {
	t.Parallel()

	data, err := os.ReadFile("README.md")
	if err != nil {
		t.Fatalf("read README.md: %v", err)
	}
	readme := string(data)

	for _, want := range []string{
		`PolymetricsAI.PolymetricsCLI`,
		`{34C3F556-5634-5381-AE18-E1668FDECFA7}`,
		`{EFEAFAA1-4276-509D-945A-D4F9BF7DBA30}`,
		`%ProgramFiles%\Polymetrics\CLI\pm.exe`,
		`arm64`,
		`x64`,
	} {
		if !strings.Contains(readme, want) {
			t.Fatalf("README.md missing %q", want)
		}
	}
}

func TestVerifierNormalizesWindowsInstallerScalarPadding(t *testing.T) {
	t.Parallel()

	data, err := os.ReadFile(filepath.Join("..", "..", "scripts", "verify-windows-package.ps1"))
	if err != nil {
		t.Fatalf("read verify-windows-package.ps1: %v", err)
	}
	script := string(data)

	for _, want := range []string{
		"function Normalize-MsiScalar",
		"$Value.Trim()",
		"return Normalize-MsiScalar",
	} {
		if !strings.Contains(script, want) {
			t.Fatalf("verify-windows-package.ps1 missing %q", want)
		}
	}
}
