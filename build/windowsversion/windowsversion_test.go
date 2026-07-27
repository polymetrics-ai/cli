package main

import (
	"bytes"
	"debug/pe"
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
		`1 VERSIONINFO`,
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
	if strings.Contains(rc, "#include") {
		t.Fatalf("RenderRC() emitted SDK header include:\n%s", rc)
	}
}

func TestRenderSysoGeneratesLinkableResourceObject(t *testing.T) {
	t.Parallel()

	version := Version{Major: 1, Minor: 2, Patch: 3, Build: 0}
	tests := []struct {
		name      string
		goarch    string
		machine   uint16
		relocType uint16
	}{
		{
			name:      "amd64",
			goarch:    "amd64",
			machine:   coffMachineAMD64,
			relocType: coffRelAMD64Addr32NB,
		},
		{
			name:      "arm64",
			goarch:    "arm64",
			machine:   coffMachineARM64,
			relocType: coffRelARM64Addr32NB,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			first, err := RenderSyso(version, tt.goarch)
			if err != nil {
				t.Fatalf("RenderSyso(%q) returned error: %v", tt.goarch, err)
			}
			second, err := RenderSyso(version, tt.goarch)
			if err != nil {
				t.Fatalf("RenderSyso(%q) second run returned error: %v", tt.goarch, err)
			}
			if !bytes.Equal(first, second) {
				t.Fatal("RenderSyso() is not deterministic")
			}

			file, err := pe.NewFile(bytes.NewReader(first))
			if err != nil {
				t.Fatalf("generated syso is not a PE/COFF object: %v", err)
			}
			defer file.Close()

			if file.Machine != tt.machine {
				t.Fatalf("generated machine = %#x, want %#x", file.Machine, tt.machine)
			}
			if len(file.Sections) != 1 || file.Sections[0].Name != ".rsrc" {
				t.Fatalf("generated sections = %#v, want single .rsrc section", file.Sections)
			}
			resource, err := file.Sections[0].Data()
			if err != nil {
				t.Fatalf("read .rsrc data: %v", err)
			}
			if !bytes.Contains(resource, utf16Bytes("Polymetrics AI")) {
				t.Fatalf(".rsrc data does not contain CompanyName")
			}
			if !bytes.Contains(resource, utf16Bytes("1.2.3.0")) {
				t.Fatalf(".rsrc data does not contain normalized version")
			}
			if len(file.Sections[0].Relocs) != 1 {
				t.Fatalf(".rsrc relocations = %#v, want one relocation", file.Sections[0].Relocs)
			}
			if file.Sections[0].Relocs[0].Type != tt.relocType {
				t.Fatalf(".rsrc relocation type = %#x, want %#x", file.Sections[0].Relocs[0].Type, tt.relocType)
			}
		})
	}
}

func TestRenderSysoRejectsUnsupportedArch(t *testing.T) {
	t.Parallel()

	if _, err := RenderSyso(Version{Major: 1, Minor: 2, Patch: 3, Build: 0}, "386"); err == nil {
		t.Fatal("RenderSyso() succeeded for unsupported arch, want error")
	}
}
