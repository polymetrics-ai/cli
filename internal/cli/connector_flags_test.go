package cli

import (
	"strings"
	"testing"

	"polymetrics.ai/internal/connectors"
)

func TestValidateConnectorCommandFlagValuesRejectsBareNonBooleanFlags(t *testing.T) {
	command := connectors.CommandSurfaceCommand{Flags: []connectors.CommandSurfaceFlag{
		{Name: "coins", Type: "string_array"},
		{Name: "include-inactive", Type: "boolean"},
	}}
	tests := []struct {
		name      string
		args      []string
		wantErr   string
		wantCoins string
	}{
		{name: "bare non boolean", args: []string{"lookup", "coins", "--coins"}, wantErr: "--coins requires a value"},
		{name: "explicit true literal", args: []string{"lookup", "coins", "--coins=true"}, wantCoins: "true"},
		{name: "bare boolean", args: []string{"lookup", "coins", "--include-inactive"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			flags := parseFlags(tt.args)
			err := validateConnectorCommandFlagValues(flags, command)
			if tt.wantErr != "" {
				if err == nil || !strings.Contains(err.Error(), tt.wantErr) {
					t.Fatalf("validateConnectorCommandFlagValues() error = %v, want %q", err, tt.wantErr)
				}
				return
			}
			if err != nil {
				t.Fatalf("validateConnectorCommandFlagValues() error = %v", err)
			}
			if tt.wantCoins != "" && flags.first("coins") != tt.wantCoins {
				t.Fatalf("coins = %q, want %q", flags.first("coins"), tt.wantCoins)
			}
		})
	}
}
