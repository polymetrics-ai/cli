// Command windowsversion renders a deterministic Windows VERSIONINFO resource script for pm.exe.
//
// The generated .rc file is source text. CI compiles it to an architecture-specific .syso file with
// Windows SDK tools before building Windows snapshots. Generated .syso files are build artifacts and
// must not be committed.
package main

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"
)

const (
	companyName      = "Polymetrics AI"
	fileDescription  = "Polymetrics CLI"
	internalName     = "pm"
	originalFilename = "pm.exe"
	productName      = "Polymetrics CLI"
	copyright        = "Copyright (c) 2026 Polymetrics AI"
)

var semverPattern = regexp.MustCompile(`^v?([0-9]+)\.([0-9]+)\.([0-9]+)(?:[-+].*)?$`)

// Version is the four-part numeric Windows file/product version.
type Version struct {
	Major int
	Minor int
	Patch int
	Build int
}

// String returns the dotted four-part Windows version string.
func (v Version) String() string {
	return fmt.Sprintf("%d.%d.%d.%d", v.Major, v.Minor, v.Patch, v.Build)
}

// CommaString returns the comma-separated numeric form required by VERSIONINFO FILEVERSION fields.
func (v Version) CommaString() string {
	return fmt.Sprintf("%d,%d,%d,%d", v.Major, v.Minor, v.Patch, v.Build)
}

// NormalizeVersion converts a PM release version into a four-part Windows VERSIONINFO version.
func NormalizeVersion(input string) (Version, error) {
	trimmed := strings.TrimSpace(input)
	if trimmed == "" {
		return Version{}, errors.New("version is required")
	}

	matches := semverPattern.FindStringSubmatch(trimmed)
	if matches == nil {
		return Version{}, fmt.Errorf("version %q must look like vMAJOR.MINOR.PATCH", input)
	}

	parts := make([]int, 3)
	for i := range parts {
		value, err := strconv.Atoi(matches[i+1])
		if err != nil {
			return Version{}, fmt.Errorf("parse version component %q: %w", matches[i+1], err)
		}
		if value < 0 || value > 65535 {
			return Version{}, fmt.Errorf("version component %d out of Windows VERSIONINFO range 0..65535", value)
		}
		parts[i] = value
	}

	return Version{Major: parts[0], Minor: parts[1], Patch: parts[2], Build: 0}, nil
}

// RenderRC renders the Windows resource script consumed by rc.exe.
func RenderRC(version Version) string {
	commaVersion := version.CommaString()
	textVersion := version.String()

	return fmt.Sprintf(`1 VERSIONINFO
FILEVERSION %s
PRODUCTVERSION %s
FILEFLAGSMASK 0x3fL
FILEFLAGS 0x0L
FILEOS 0x40004L
FILETYPE 0x1L
FILESUBTYPE 0x0L
BEGIN
    BLOCK "StringFileInfo"
    BEGIN
        BLOCK "040904B0"
        BEGIN
            VALUE "CompanyName", "%s"
            VALUE "FileDescription", "%s"
            VALUE "FileVersion", "%s"
            VALUE "InternalName", "%s"
            VALUE "OriginalFilename", "%s"
            VALUE "ProductName", "%s"
            VALUE "ProductVersion", "%s"
            VALUE "LegalCopyright", "%s"
        END
    END
    BLOCK "VarFileInfo"
    BEGIN
        VALUE "Translation", 0x0409, 1200
    END
END
`, commaVersion, commaVersion, companyName, fileDescription, textVersion, internalName, originalFilename, productName, textVersion, copyright)
}

func main() {
	versionFlag := flag.String("version", "", "PM release version, for example v1.2.3")
	outFlag := flag.String("out", "", "output .rc path; writes to stdout when omitted")
	flag.Parse()

	version, err := NormalizeVersion(*versionFlag)
	if err != nil {
		fmt.Fprintf(os.Stderr, "windowsversion: %v\n", err)
		os.Exit(2)
	}

	content := []byte(RenderRC(version))
	if *outFlag == "" {
		if _, err := os.Stdout.Write(content); err != nil {
			fmt.Fprintf(os.Stderr, "windowsversion: write stdout: %v\n", err)
			os.Exit(1)
		}
		return
	}
	if err := os.WriteFile(*outFlag, content, 0o644); err != nil {
		fmt.Fprintf(os.Stderr, "windowsversion: write %s: %v\n", *outFlag, err)
		os.Exit(1)
	}
}
