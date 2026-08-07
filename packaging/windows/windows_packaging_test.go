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
		// The arm64 UpgradeCode stays documented even though the target is not
		// published, so the same MSI upgrade identity is reused if it returns.
		`{EFEAFAA1-4276-509D-945A-D4F9BF7DBA30}`,
		`%ProgramFiles%\Polymetrics\CLI\pm.exe`,
		`arm64`,
		`x64`,
		// Why arm64 is unpublished has to be written down where a packager
		// looks, or its absence reads as an oversight.
		`not published`,
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
		"return ([string]$Value).Trim()",
		"[void]$view.Execute()",
		"[void]$view.Close()",
		"$value = Invoke-MsiScalar",
		"return (Normalize-MsiScalar -Value $value)",
	} {
		if !strings.Contains(script, want) {
			t.Fatalf("verify-windows-package.ps1 missing %q", want)
		}
	}
	if strings.Contains(script, "$Value -is [string]") {
		t.Fatal("verify-windows-package.ps1 must cast MSI scalar values before trimming; COM output can arrive as non-string objects")
	}
}
