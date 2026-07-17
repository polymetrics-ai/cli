// Package styles defines semantic terminal palette and glyph foundations for
// future Polymetrics TTY renderers.
package styles

import "fmt"

// ColorProfile describes the highest color capability available to a renderer.
type ColorProfile string

const (
	ProfileNone      ColorProfile = "none"
	ProfileANSI16    ColorProfile = "ansi16"
	ProfileANSI256   ColorProfile = "ansi256"
	ProfileTrueColor ColorProfile = "truecolor"
)

// Token names a semantic color role. Components should use tokens rather than
// raw colors so degradation and accessible remaps stay centralized.
type Token string

const (
	TokenFlow Token = "flow"
	TokenOK   Token = "ok"
	TokenWarn Token = "warn"
	TokenFail Token = "fail"
	TokenDim  Token = "dim"
	TokenInk  Token = "ink"
)

// Options controls palette resolution.
type Options struct {
	Profile    ColorProfile
	Dark       bool
	Accessible bool
}

// Palette is a resolved semantic palette for one terminal capability profile.
type Palette struct {
	profile ColorProfile
	colors  map[Token]string
}

// ResolvePalette returns a semantic palette that degrades from truecolor to
// ANSI16 or no-color. Accessible mode uses only standard 16-color indexes.
func ResolvePalette(opts Options) Palette {
	profile := opts.Profile
	if profile == "" {
		profile = ProfileNone
	}
	colors := map[Token]string{}
	if profile == ProfileNone {
		return Palette{profile: profile, colors: colors}
	}
	if opts.Accessible || profile == ProfileANSI16 {
		colors[TokenFlow] = "6"
		colors[TokenOK] = "2"
		colors[TokenWarn] = "3"
		colors[TokenFail] = "1"
		colors[TokenDim] = "8"
		return Palette{profile: ProfileANSI16, colors: colors}
	}
	if profile == ProfileANSI256 {
		colors[TokenFlow] = "37"
		colors[TokenOK] = "77"
		colors[TokenWarn] = "214"
		colors[TokenFail] = "203"
		colors[TokenDim] = "244"
		return Palette{profile: profile, colors: colors}
	}
	colors[TokenFlow] = lightDark(opts.Dark, "#2DD4BF", "#0F766E")
	colors[TokenOK] = lightDark(opts.Dark, "#4ADE80", "#15803D")
	colors[TokenWarn] = lightDark(opts.Dark, "#FBBF24", "#B45309")
	colors[TokenFail] = lightDark(opts.Dark, "#F87171", "#B91C1C")
	colors[TokenDim] = lightDark(opts.Dark, "#6B7280", "#9CA3AF")
	return Palette{profile: ProfileTrueColor, colors: colors}
}

// Color returns the resolved color value for token. It returns an empty string
// for no-color profiles and for TokenInk, which uses the terminal default.
func (p Palette) Color(token Token) string {
	return p.colors[token]
}

// Style wraps text in an ANSI sequence for token when the profile supports it.
// No-color profiles return text unchanged.
func (p Palette) Style(token Token, text string) string {
	color := p.Color(token)
	if color == "" {
		return text
	}
	switch p.profile {
	case ProfileANSI16:
		return fmt.Sprintf("\x1b[3%sm%s\x1b[0m", color, text)
	case ProfileANSI256:
		return fmt.Sprintf("\x1b[38;5;%sm%s\x1b[0m", color, text)
	case ProfileTrueColor:
		r, g, b, ok := parseHex(color)
		if !ok {
			return text
		}
		return fmt.Sprintf("\x1b[38;2;%d;%d;%dm%s\x1b[0m", r, g, b, text)
	default:
		return text
	}
}

func lightDark(dark bool, darkValue, lightValue string) string {
	if dark {
		return darkValue
	}
	return lightValue
}

func parseHex(hex string) (int, int, int, bool) {
	if len(hex) != 7 || hex[0] != '#' {
		return 0, 0, 0, false
	}
	var r, g, b int
	if _, err := fmt.Sscanf(hex, "#%02x%02x%02x", &r, &g, &b); err != nil {
		return 0, 0, 0, false
	}
	return r, g, b, true
}

// Glyphs contains semantic status and rail glyphs.
type Glyphs struct {
	OK           string
	Failed       string
	Running      string
	Pending      string
	Partial      string
	Warning      string
	Skipped      string
	RailVertical string
	RailBranch   string
	RailEnd      string
}

// ResolveGlyphs returns Unicode glyphs for capable terminals and ASCII-safe
// fallbacks for pipes, dumb terminals, and PM_ASCII-style constraints.
func ResolveGlyphs(ascii bool) Glyphs {
	if ascii {
		return Glyphs{
			OK:           "[ok]",
			Failed:       "[x]",
			Running:      "[*]",
			Pending:      "[ ]",
			Partial:      "[~]",
			Warning:      "[!]",
			Skipped:      "[-]",
			RailVertical: "|",
			RailBranch:   "+",
			RailEnd:      "`",
		}
	}
	return Glyphs{
		OK:           "✓",
		Failed:       "✗",
		Running:      "●",
		Pending:      "○",
		Partial:      "◐",
		Warning:      "▲",
		Skipped:      "–",
		RailVertical: "│",
		RailBranch:   "├",
		RailEnd:      "└",
	}
}
