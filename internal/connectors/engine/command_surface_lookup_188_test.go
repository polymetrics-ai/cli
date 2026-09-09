package engine

import (
	"os"
	"sync"
	"testing"
)

func TestCommandSurfaceLookupContinuity188(t *testing.T) {
	op := func(id string, cap int) OperationSpec {
		return OperationSpec{ID: id, Kind: "rest_read", REST: &RESTOperationSpec{Path: "/same", Parameters: []OperationParameter{{Name: "id", In: "path", MaxBytes: cap}}}}
	}
	for _, tc := range []struct {
		name       string
		operations []OperationSpec
		id         string
		cliCap     int
		want       int
	}{
		{"exact_id_same_route", []OperationSpec{op("a", 11), op("b", 23)}, "b", 0, 23},
		{"reordered", []OperationSpec{op("b", 23), op("a", 11)}, "b", 0, 23},
		{"first_duplicate", []OperationSpec{op("b", 23), op("b", 41)}, "b", 0, 23},
		{"missing", []OperationSpec{op("a", 11)}, "missing", 0, defaultOperationParameterMaxBytes},
		{"non_rest", []OperationSpec{{ID: "b"}}, "b", 0, defaultOperationParameterMaxBytes},
		{"empty_id", []OperationSpec{op("", 11)}, "", 0, defaultOperationParameterMaxBytes},
		{"exact_not_trimmed", []OperationSpec{op("b", 23)}, " b ", 0, defaultOperationParameterMaxBytes},
		{"stricter_cli", []OperationSpec{op("b", 23)}, "b", 7, 7},
		{"stricter_provider", []OperationSpec{op("b", 23)}, "b", 99, 23},
		{"provider_default", []OperationSpec{op("b", 0)}, "b", 0, defaultOperationParameterMaxBytes},
	} {
		t.Run(tc.name, func(t *testing.T) {
			bundle := Bundle{Operations: tc.operations, CLISurface: &CLISurface{Commands: []CLICommand{{Path: "get", Operation: tc.id, Flags: []CLIFlag{{Name: "id", MapsTo: "path.id", MaxBytes: tc.cliCap}}}}}}
			got := synthesizeCommandSurface(bundle).Commands[0].Flags[0]
			if got.MaxBytes != tc.want || got.MapsTo != "path.id" || got.MaxBytesOrigin != "pm_policy" || got.MaxBytesPolicyVersion != OperationParameterExecutionPolicyVersion {
				t.Fatalf("projection = %+v; expected cap %d and original binding/policy", got, tc.want)
			}
		})
	}
}

func TestCommandSurfaceLookupIsolation188(t *testing.T) {
	allowed := true
	bundle := Bundle{Name: "same", Operations: []OperationSpec{{ID: "get", REST: &RESTOperationSpec{Parameters: []OperationParameter{{Name: "id", In: "query", MaxBytes: 23}}}}}, CLISurface: &CLISurface{Commands: []CLICommand{{Path: "get", Operation: "get", Flags: []CLIFlag{{Name: "id", MapsTo: "query.id", Values: []string{"original"}, AllowEmpty: &allowed}}}}}}
	first := synthesizeCommandSurface(bundle)
	first.Commands[0].Flags[0].Values[0] = "changed"
	*first.Commands[0].Flags[0].AllowEmpty = false
	first.Commands[0].Flags[0].MaxBytes = 1
	second := synthesizeCommandSurface(bundle).Commands[0].Flags[0]
	if second.Values[0] != "original" || !*second.AllowEmpty || second.MaxBytes != 23 || !allowed || bundle.CLISurface.Commands[0].Flags[0].Values[0] != "original" {
		t.Fatal("returned projection mutated another projection or source")
	}
	// A new source with the same connector and operation names must be observed
	// on the next projection; no name-keyed persistent cache is permissible.
	bundle.Operations[0].REST.Parameters[0].MaxBytes = 47
	if got := synthesizeCommandSurface(bundle).Commands[0].Flags[0].MaxBytes; got != 47 {
		t.Fatalf("new source cap = %d, want 47", got)
	}
	var wg sync.WaitGroup
	for range 8 {
		wg.Go(func() {
			got := synthesizeCommandSurface(bundle).Commands[0].Flags[0]
			if got.MaxBytes != 47 || got.Values[0] != "original" || !*got.AllowEmpty {
				t.Error("concurrent projection did not preserve source")
			}
			got.Values[0] = "private"
			*got.AllowEmpty = false
		})
	}
	wg.Wait()
}

func BenchmarkCommandSurfaceHubSpot188(b *testing.B) {
	bundle, err := Load(os.DirFS("../defs"), "hubspot")
	if err != nil {
		b.Fatal(err)
	}
	b.ReportAllocs()
	b.ResetTimer()
	for range b.N {
		out := synthesizeCommandSurface(bundle)
		if len(out.Commands) != 3141 {
			b.Fatalf("commands = %d", len(out.Commands))
		}
	}
}
