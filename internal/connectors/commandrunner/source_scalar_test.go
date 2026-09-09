package commandrunner

import (
	"encoding/json"
	"polymetrics.ai/internal/connectors"
	"strings"
	"testing"
)

func TestGeneratedSourceScalarCodec234(t *testing.T) {
	for _, tc := range []struct {
		name, kind, raw string
		want            any
		invalid         bool
	}{
		{"false", "boolean", "false", false, false},
		{"true", "boolean", "true", true, false},
		{"truthy_number", "boolean", "1", nil, true},
		{"truthy_upper", "boolean", "TRUE", nil, true},
		{"empty_boolean", "boolean", "", nil, true},
		{"empty_string", "string", "", "", false},
		{"string_false", "string", "false", "false", false},
		{"exact_integer", "integer", "9007199254740994", json.Number("9007199254740994"), false},
		{"decimal_integer", "integer", "9007199254740994.5", nil, true},
		{"exact_decimal", "number", "0.10000000000000002", json.Number("0.10000000000000002"), false},
		{"fraction", "number", "1/10", nil, true},
		{"spaced_number", "number", " 1", nil, true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			empty := true
			got, err := coerceFlagValue(connectors.CommandSurfaceFlag{Name: "value", Type: tc.kind, InputCodec: "source_scalar_v1", AllowEmpty: &empty}, []string{tc.raw})
			if tc.invalid {
				if err == nil {
					t.Fatalf("accepted invalid source scalar %q as %#v", tc.raw, got)
				}
				return
			}
			if err != nil || got != tc.want {
				t.Fatalf("got %#v, %v; want %#v", got, err, tc.want)
			}
		})
	}
	t.Run("unknown_codec", func(t *testing.T) {
		if _, err := coerceFlagValue(connectors.CommandSurfaceFlag{Name: "value", Type: "string", InputCodec: "unknown"}, []string{"value"}); err == nil {
			t.Fatal("accepted unknown codec")
		}
	})
}

func TestSourceScalarResourceAndLegacy234(t *testing.T) {
	for _, raw := range []string{"1e999999999999999999999999999999", "1e1048577", "1e-1048577", strings.Repeat("1", (1<<20)+1)} {
		if _, err := coerceFlagValue(connectors.CommandSurfaceFlag{Name: "value", Type: "number", InputCodec: "source_scalar_v1"}, []string{raw}); err == nil {
			t.Fatal("accepted scalar beyond finite work budget")
		}
	}
	for _, raw := range []string{"1", "TRUE"} {
		got, err := coerceFlagValue(connectors.CommandSurfaceFlag{Name: "legacy", Type: "boolean"}, []string{raw})
		if err != nil || got != true {
			t.Fatalf("legacy boolean changed: %#v %v", got, err)
		}
	}
	for _, tc := range []struct {
		kind, raw string
		budget    int
	}{
		{"string", string([]byte{255}), 8}, {"number", "1", 0}, {"integer", "1e0", 8}, {"number", "null", 8}, {"number", "NaN", 8}, {"object", "{}", 8},
	} {
		if _, err := connectors.DecodeSourceScalar(tc.kind, tc.raw, tc.budget); err == nil {
			t.Fatalf("accepted invalid %s scalar", tc.kind)
		}
	}
	got, err := connectors.DecodeSourceScalar("number", "1e3", 8)
	if err != nil || got != json.Number("1e3") {
		t.Fatalf("exact exponent: %#v %v", got, err)
	}
}
