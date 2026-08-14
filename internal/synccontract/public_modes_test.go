package synccontract

import (
	"reflect"
	"testing"
)

func TestPublicModesResolveCompatibilityNamesToClosedContracts(t *testing.T) {
	tests := []struct {
		name     string
		raw      string
		wantName string
		contract Mode
		typed    bool
	}{
		{name: "full refresh append", raw: "full_refresh_append", wantName: "full_refresh_append", contract: ModeFullAppend},
		{name: "full refresh overwrite", raw: "full_refresh_overwrite", wantName: "full_refresh_overwrite", contract: ModeFullOverwrite},
		{name: "full overwrite compatibility", raw: "full_refresh_overwrite_deduped", wantName: "full_refresh_overwrite_deduped", contract: ModeFullOverwrite, typed: true},
		{name: "full overwrite retained spelling", raw: "full_refresh_deduped", wantName: "full_refresh_overwrite_deduped", contract: ModeFullOverwrite, typed: true},
		{name: "incremental append", raw: "incremental_append", wantName: "incremental_append", contract: ModeIncrementalAppend},
		{name: "incremental dedupe compatibility", raw: "incremental_append_deduped", wantName: "incremental_append_deduped", contract: ModeIncrementalDedupe, typed: true},
		{name: "incremental dedupe retained spelling", raw: "incremental_append_dedup", wantName: "incremental_append_deduped", contract: ModeIncrementalDedupe, typed: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mode, ok := LookupPublicMode(tt.raw)
			if !ok {
				t.Fatalf("LookupPublicMode(%q) did not resolve", tt.raw)
			}
			if mode.Name != tt.wantName || mode.ContractMode != tt.contract || mode.TypedOnly != tt.typed {
				t.Fatalf("LookupPublicMode(%q) = %+v, want name=%q contract=%q typed=%t", tt.raw, mode, tt.wantName, tt.contract, tt.typed)
			}
		})
	}
}

func TestPublicModeCapabilitiesAndDefaultsUseMaterializingModes(t *testing.T) {
	tests := []struct {
		name         string
		capabilities PublicModeCapabilities
		wantModes    []string
		wantDefault  string
	}{
		{
			name:        "neither",
			wantModes:   []string{"full_refresh_append", "full_refresh_overwrite"},
			wantDefault: "full_refresh_append",
		},
		{
			name:         "primary key only",
			capabilities: PublicModeCapabilities{HasPrimaryKey: true},
			wantModes:    []string{"full_refresh_append", "full_refresh_overwrite"},
			wantDefault:  "full_refresh_overwrite",
		},
		{
			name:         "primary key and cursor without incremental executor",
			capabilities: PublicModeCapabilities{HasPrimaryKey: true, HasCursor: true},
			wantModes: []string{
				"full_refresh_append", "full_refresh_overwrite", "full_refresh_overwrite_deduped",
			},
			wantDefault: "full_refresh_overwrite",
		},
		{
			name:         "cursor and incremental executor without primary key",
			capabilities: PublicModeCapabilities{HasCursor: true, HasIncrementalExecutor: true},
			wantModes:    []string{"full_refresh_append", "full_refresh_overwrite", "incremental_append"},
			wantDefault:  "incremental_append",
		},
		{
			name:         "all capabilities",
			capabilities: PublicModeCapabilities{HasPrimaryKey: true, HasCursor: true, HasIncrementalExecutor: true},
			wantModes: []string{
				"full_refresh_append", "full_refresh_overwrite", "full_refresh_overwrite_deduped",
				"incremental_append", "incremental_append_deduped",
			},
			wantDefault: "incremental_append",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := SupportedPublicModeNames(tt.capabilities); !reflect.DeepEqual(got, tt.wantModes) {
				t.Fatalf("SupportedPublicModeNames(%+v) = %v, want %v", tt.capabilities, got, tt.wantModes)
			}
			if got := DefaultPublicModeName(tt.capabilities); got != tt.wantDefault {
				t.Fatalf("DefaultPublicModeName(%+v) = %q, want %q", tt.capabilities, got, tt.wantDefault)
			}
		})
	}

	wantMaterializing := []string{"full_refresh_append", "full_refresh_overwrite", "incremental_append"}
	if got := MaterializingPublicModeNames(); !reflect.DeepEqual(got, wantMaterializing) {
		t.Fatalf("MaterializingPublicModeNames() = %v, want %v", got, wantMaterializing)
	}
}
