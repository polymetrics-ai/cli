package engine

import (
	"strings"
	"testing"
)

func baseVars() Vars {
	return Vars{
		Config: map[string]string{
			"base_url":   "https://api.example.com",
			"repository": "a/../b",
			"query_char": "a?x=1&y=2",
			"space":      "a b",
			"unicode":    "héllo",
			"double_enc": "%2e%2e",
			"auth_type":  "token",
			"empty":      "",
		},
		Secrets: map[string]string{
			"token": "sekret-token",
		},
		Record: map[string]any{
			"user": map[string]any{
				"login": "octocat",
			},
			"created_at": "2024-01-02T03:04:05Z",
		},
		Cursor: "cursor-value",
	}
}

func TestInterpolateResolution(t *testing.T) {
	vars := baseVars()

	tests := []struct {
		name     string
		template string
		want     string
	}{
		{"config", "{{ config.base_url }}", "https://api.example.com"},
		{"secrets", "{{ secrets.token }}", "sekret-token"},
		{"record dotted path", "{{ record.user.login }}", "octocat"},
		{"cursor", "{{ cursor }}", "cursor-value"},
		{"literal text passthrough", "prefix-{{ cursor }}-suffix", "prefix-cursor-value-suffix"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Interpolate(tt.template, vars)
			if err != nil {
				t.Fatalf("Interpolate error: %v", err)
			}
			if got != tt.want {
				t.Fatalf("Interpolate(%q) = %q, want %q", tt.template, got, tt.want)
			}
		})
	}
}

func TestInterpolateMissingKey(t *testing.T) {
	vars := baseVars()
	_, err := Interpolate("{{ config.nope }}", vars)
	if err == nil {
		t.Fatalf("expected error for missing key")
	}
	if !strings.Contains(err.Error(), "nope") || !strings.Contains(err.Error(), "config") {
		t.Fatalf("error %q does not name key+namespace", err.Error())
	}

	_, err = Interpolate("{{ secrets.nope }}", vars)
	if err == nil {
		t.Fatalf("expected error for missing secret key")
	}
	if !strings.Contains(err.Error(), "nope") || !strings.Contains(err.Error(), "secrets") {
		t.Fatalf("error %q does not name key+namespace", err.Error())
	}
}

func TestInterpolatePathDefaultURLEncode(t *testing.T) {
	vars := baseVars()

	tests := []struct {
		name     string
		template string
		want     string
	}{
		{
			name:     "path traversal encoded",
			template: "/repos/{{ config.repository }}",
			want:     "/repos/a%2F..%2Fb",
		},
		{
			name:     "query metachars encoded",
			template: "/x/{{ config.query_char }}",
			want:     "/x/a%3Fx%3D1%26y%3D2",
		},
		{
			name:     "space and unicode encoded",
			template: "/{{ config.space }}/{{ config.unicode }}",
			want:     "/a%20b/h%C3%A9llo",
		},
		{
			name:     "double encode guard: percent literal is re-encoded",
			template: "/{{ config.double_enc }}",
			want:     "/%252e%252e",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := InterpolatePath(tt.template, vars)
			if err != nil {
				t.Fatalf("InterpolatePath error: %v", err)
			}
			if got != tt.want {
				t.Fatalf("InterpolatePath(%q) = %q, want %q", tt.template, got, tt.want)
			}
		})
	}
}

func TestInterpolateFilters(t *testing.T) {
	vars := baseVars()

	t.Run("unix_seconds on rfc3339", func(t *testing.T) {
		got, err := Interpolate("{{ record.created_at | unix_seconds }}", vars)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got != "1704164645" {
			t.Fatalf("got %q, want unix seconds for 2024-01-02T03:04:05Z", got)
		}
	})

	t.Run("unix_seconds on bad input errors", func(t *testing.T) {
		v := baseVars()
		v.Config["bad_date"] = "not-a-date"
		_, err := Interpolate("{{ config.bad_date | unix_seconds }}", v)
		if err == nil {
			t.Fatalf("expected error for bad date input")
		}
	})

	t.Run("base64", func(t *testing.T) {
		got, err := Interpolate("{{ secrets.token | base64 }}", vars)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got != "c2VrcmV0LXRva2Vu" {
			t.Fatalf("got %q", got)
		}
	})

	t.Run("explicit urlencode filter in non-path context", func(t *testing.T) {
		got, err := Interpolate("{{ config.space | urlencode }}", vars)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got != "a%20b" {
			t.Fatalf("got %q", got)
		}
	})
}

func TestInterpolateHeaderCRLFInjectionRejected(t *testing.T) {
	vars := baseVars()
	vars.Config["evil"] = "value\r\nX-Injected: true"

	_, err := InterpolateHeader("{{ config.evil }}", vars)
	if err == nil {
		t.Fatalf("expected error for CRLF injection in header value")
	}

	vars.Config["evil_path"] = "a\r\nb"
	_, err = InterpolatePath("/{{ config.evil_path }}", vars)
	if err == nil {
		t.Fatalf("expected error for CRLF injection in path value")
	}
}

func TestEvalWhen(t *testing.T) {
	vars := baseVars()

	tests := []struct {
		name string
		cond string
		want bool
	}{
		{"equality true", "{{ config.auth_type == 'token' }}", true},
		{"equality false", "{{ config.auth_type == 'public' }}", false},
		{"in list true", "{{ config.auth_type in ['auto', 'token'] }}", true},
		{"in list false", "{{ config.auth_type in ['auto', 'public'] }}", false},
		{"truthiness non-empty", "{{ config.base_url }}", true},
		{"truthiness empty", "{{ config.empty }}", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := EvalWhen(tt.cond, vars)
			if err != nil {
				t.Fatalf("EvalWhen error: %v", err)
			}
			if got != tt.want {
				t.Fatalf("EvalWhen(%q) = %v, want %v", tt.cond, got, tt.want)
			}
		})
	}
}

func TestEvalWhenUnknownOperatorIsCompileError(t *testing.T) {
	vars := baseVars()
	_, err := EvalWhen("{{ config.auth_type >= 'token' }}", vars)
	if err == nil {
		t.Fatalf("expected compile error for unknown operator")
	}
}

func TestResolveCheck(t *testing.T) {
	specKeys := map[string]bool{"repository": true, "base_url": true}

	if err := ResolveCheck("/repos/{{ config.repository }}", specKeys); err != nil {
		t.Fatalf("unexpected error for known key: %v", err)
	}

	err := ResolveCheck("/repos/{{ config.unknown_key }}", specKeys)
	if err == nil {
		t.Fatalf("expected validation finding for unknown spec key")
	}
	if !strings.Contains(err.Error(), "unknown_key") {
		t.Fatalf("error %q does not name the offending key", err.Error())
	}

	// record/cursor/secrets references are not checked against specKeys.
	if err := ResolveCheck("{{ record.user.login }}", specKeys); err != nil {
		t.Fatalf("unexpected error for record reference: %v", err)
	}
	if err := ResolveCheck("{{ cursor }}", specKeys); err != nil {
		t.Fatalf("unexpected error for cursor reference: %v", err)
	}
}
