package safety

import "testing"

func TestCanonicalProviderCitationURL(t *testing.T) {
	tests := []struct {
		name string
		raw  string
		want string
	}{
		{
			name: "canonical URL unchanged",
			raw:  "https://provider.example.test/v1/reference?a=1&b=2",
			want: "https://provider.example.test/v1/reference?a=1&b=2",
		},
		{
			name: "host port path and query normalized",
			raw:  "HTTPS://PROVIDER.EXAMPLE.TEST:443/v1/%72eference?b=2&a=1",
			want: "https://provider.example.test/v1/reference?a=1&b=2",
		},
		{
			name: "reserved path escape retained with uppercase hex",
			raw:  "https://provider.example.test/docs%2freference",
			want: "https://provider.example.test/docs%2Freference",
		},
		{
			name: "public IPv4 literal",
			raw:  "https://8.8.8.8:443/reference",
			want: "https://8.8.8.8/reference",
		},
		{
			name: "empty path becomes root",
			raw:  "https://provider.example.test",
			want: "https://provider.example.test/",
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			got, err := CanonicalProviderCitationURL(testCase.raw)
			if err != nil {
				t.Fatalf("canonical provider citation: %v", err)
			}
			if got != testCase.want {
				t.Fatalf("canonical provider citation = %q, want %q", got, testCase.want)
			}
		})
	}
}

func TestCanonicalProviderCitationURLRejectsUnsafeOrAmbiguousAuthority(t *testing.T) {
	tests := []struct {
		name string
		raw  string
	}{
		{name: "HTTP", raw: "http://provider.example.test/reference"},
		{name: "userinfo", raw: "https://user@provider.example.test/reference"},
		{name: "fragment", raw: "https://provider.example.test/reference#operation"},
		{name: "private literal", raw: "https://127.0.0.1/reference"},
		{name: "documentation literal", raw: "https://192.0.2.1/reference"},
		{name: "localhost", raw: "https://localhost/reference"},
		{name: "ambiguous single-label host", raw: "https://provider/reference"},
		{name: "trailing dot", raw: "https://provider.example.test./reference"},
		{name: "empty DNS label", raw: "https://provider..example.test/reference"},
		{name: "Unicode DNS label", raw: "https://prøvider.example.test/reference"},
		{name: "integer-form address", raw: "https://2130706433/reference"},
		{name: "ambiguous numeric address", raw: "https://127.0.0.01/reference"},
		{name: "plain dot segment", raw: "https://provider.example.test/docs/../reference"},
		{name: "escaped dot segment", raw: "https://provider.example.test/docs/%2e%2e/reference"},
		{name: "credential query", raw: "https://provider.example.test/reference?api_key=value"},
		{name: "repeated query key", raw: "https://provider.example.test/reference?a=1&a=2"},
		{name: "empty query marker", raw: "https://provider.example.test/reference?"},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			if _, err := CanonicalProviderCitationURL(testCase.raw); err == nil {
				t.Fatalf("unsafe or ambiguous provider citation %q passed", testCase.raw)
			}
		})
	}
}
