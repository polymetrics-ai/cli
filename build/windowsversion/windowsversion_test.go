package main

import (
	"strings"
	"testing"
)

func TestNormalizeVersion(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		input    string
		want     Version
		wantText string
	}{
		{
			name:     "plain semver",
			input:    "0.1.0",
			want:     Version{Major: 0, Minor: 1, Patch: 0, Build: 0},
			wantText: "0.1.0.0",
		},
		{
			name:     "leading v",
			input:    "v1.2.3",
			want:     Version{Major: 1, Minor: 2, Patch: 3, Build: 0},
			wantText: "1.2.3.0",
		},
		{
			name:     "prerelease and build metadata stripped",
			input:    "v2.3.4-next.5+abcdef",
			want:     Version{Major: 2, Minor: 3, Patch: 4, Build: 0},
			wantText: "2.3.4.0",
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, err := NormalizeVersion(tt.input)
			if err != nil {
				t.Fatalf("NormalizeVersion(%q) returned error: %v", tt.input, err)
			}
			if got != tt.want {
				t.Fatalf("NormalizeVersion(%q) = %+v, want %+v", tt.input, got, tt.want)
			}
			if got.String() != tt.wantText {
				t.Fatalf("NormalizeVersion(%q).String() = %q, want %q", tt.input, got.String(), tt.wantText)
			}
		})
	}
}

func TestNormalizeVersionRejectsInvalidInput(t *testing.T) {
	t.Parallel()

	for _, input := range []string{"", "v1", "1.2", "1.2.x", "1.2.3.4", "70000.1.2"} {
		input := input
		t.Run(input, func(t *testing.T) {
			t.Parallel()

			if _, err := NormalizeVersion(input); err == nil {
				t.Fatalf("NormalizeVersion(%q) succeeded, want error", input)
			}
		})
	}
}

func TestRenderRCIncludesApprovedMetadata(t *testing.T) {
	t.Parallel()

	version := Version{Major: 1, Minor: 2, Patch: 3, Build: 0}
	rc := RenderRC(version)

	for _, want := range []string{
		`FILEVERSION 1,2,3,0`,
		`PRODUCTVERSION 1,2,3,0`,
		`VALUE "CompanyName", "Polymetrics AI"`,
		`VALUE "FileDescription", "Polymetrics CLI"`,
		`VALUE "FileVersion", "1.2.3.0"`,
		`VALUE "InternalName", "pm"`,
		`VALUE "OriginalFilename", "pm.exe"`,
		`VALUE "ProductName", "Polymetrics CLI"`,
		`VALUE "ProductVersion", "1.2.3.0"`,
		`VALUE "Translation", 0x0409, 1200`,
	} {
		if !strings.Contains(rc, want) {
			t.Fatalf("RenderRC() missing %q in:\n%s", want, rc)
		}
	}
}
