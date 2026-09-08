package cli

import (
	"bytes"
	"encoding/json"
	"os"
	"strings"
	"testing"

	"polymetrics.ai/internal/connectors/engine"
)

func TestProviderAliasHelp185(t *testing.T) {
	raw, err := os.ReadFile("../../cmd/connectorgen/testdata/flag-ownership-181.json")
	if err != nil {
		t.Fatal(err)
	}
	var cases []struct {
		Connector string            `json:"connector"`
		NewName   string            `json:"new_name"`
		Command   engine.CLICommand `json:"command"`
	}
	if err := json.Unmarshal(raw, &cases); err != nil {
		t.Fatal(err)
	}
	if len(cases) != 22 {
		t.Fatal("incomplete source cohort")
	}
	for _, tc := range cases {
		t.Run(tc.Connector+"/"+tc.Command.Path, func(t *testing.T) {
			args := append([]string{tc.Connector}, strings.Fields(tc.Command.Path)...)
			args = append(args, "--help")
			var out, diag bytes.Buffer
			if Run(args, &out, &diag) != 0 {
				t.Fatalf("help failed: %s", diag.String())
			}
			if !strings.Contains(out.String(), "--"+tc.NewName) {
				t.Fatalf("help omitted provider alias %s", tc.NewName)
			}
			if tc.Command.Write != "" && !strings.Contains(out.String(), "--approval-token-stdin") {
				t.Fatal("help omitted approval boundary")
			}
		})
	}
	for _, name := range []string{"github", "gitlab"} {
		for _, args := range [][]string{{name}, {"help", name}, {"connectors", "inspect", name, "--json"}} {
			var out, diag bytes.Buffer
			if Run(args, &out, &diag) != 0 || out.Len() == 0 {
				t.Fatalf("discovery %v failed: %s", args, diag.String())
			}
		}
	}
}
