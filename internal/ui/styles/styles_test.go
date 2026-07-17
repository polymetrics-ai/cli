package styles

import (
	"strings"
	"testing"
)

func TestGlyphsDegradeToASCII(t *testing.T) {
	unicode := ResolveGlyphs(false)
	if unicode.OK != "✓" || unicode.Failed != "✗" || unicode.RailVertical != "│" || unicode.RailBranch != "├" || unicode.RailEnd != "└" {
		t.Fatalf("unicode glyphs = %+v", unicode)
	}

	ascii := ResolveGlyphs(true)
	if ascii.OK != "[ok]" || ascii.Failed != "[x]" || ascii.Running != "[*]" || ascii.Pending != "[ ]" || ascii.Warning != "[!]" || ascii.RailVertical != "|" || ascii.RailBranch != "+" || ascii.RailEnd != "`" {
		t.Fatalf("ascii glyphs = %+v", ascii)
	}
}

func TestPaletteANSI16DimUsesBrightBlackSGR(t *testing.T) {
	palette := ResolvePalette(Options{Profile: ProfileANSI16})

	if got := palette.Style(TokenOK, "ok"); got != "\x1b[32mok\x1b[0m" {
		t.Fatalf("ANSI16 ok style = %q, want green SGR 32", got)
	}
	if got := palette.Style(TokenDim, "dim"); got != "\x1b[90mdim\x1b[0m" {
		t.Fatalf("ANSI16 dim style = %q, want bright-black SGR 90", got)
	}
}

func TestPaletteDegradesColorProfiles(t *testing.T) {
	plain := ResolvePalette(Options{Profile: ProfileNone, Dark: true})
	if got := plain.Style(TokenOK, "ok"); got != "ok" {
		t.Fatalf("plain style = %q, want unstyled text", got)
	}
	if got := plain.Color(TokenOK); got != "" {
		t.Fatalf("plain color = %q, want empty", got)
	}

	trueColor := ResolvePalette(Options{Profile: ProfileTrueColor, Dark: true})
	if got := trueColor.Color(TokenOK); got != "#4ADE80" {
		t.Fatalf("dark ok color = %q, want #4ADE80", got)
	}
	if got := trueColor.Style(TokenOK, "ok"); !strings.Contains(got, "\x1b[38;2;") || !strings.HasSuffix(got, "ok\x1b[0m") {
		t.Fatalf("truecolor style did not use ANSI truecolor escape: %q", got)
	}

	light := ResolvePalette(Options{Profile: ProfileTrueColor, Dark: false})
	if got := light.Color(TokenFlow); got != "#0F766E" {
		t.Fatalf("light flow color = %q, want #0F766E", got)
	}

	accessible := ResolvePalette(Options{Profile: ProfileANSI16, Accessible: true})
	if got := accessible.Color(TokenWarn); got != "3" {
		t.Fatalf("accessible warn color = %q, want ANSI 3", got)
	}
	if got := accessible.Style(TokenWarn, "warning"); !strings.Contains(got, "\x1b[33m") || !strings.HasSuffix(got, "warning\x1b[0m") {
		t.Fatalf("accessible ANSI16 style = %q", got)
	}
}
