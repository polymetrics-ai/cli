package cli_test

import (
	"bytes"
	"strings"
	"testing"

	"polymetrics.ai/internal/cli"
	"polymetrics.ai/internal/synccontract"
)

func normalizedHelp171(s string) string { return strings.ToLower(strings.Join(strings.Fields(s), " ")) }
func pollingHelpContract171(s string) bool {
	s = normalizedHelp171(s)
	for _, required := range []string{
		"planned, unsupported, or absent static binding alone does not implement a polling mode",
		"constructs an implemented binding per selected catalog object can become eligible only after the same runtime preflight succeeds",
		"checking the rendered native source and apply executors",
	} {
		if !strings.Contains(s, required) {
			return false
		}
	}
	return true
}
func etlHelpContract171(s string) bool {
	s = normalizedHelp171(s)
	for _, required := range []string{
		"full_refresh_append", "full_refresh_overwrite", "full_refresh_overwrite_deduped", "incremental_append", "incremental_append_deduped", "incremental_dedupe", "incremental_dedupe_history",
		"refuses before source i/o until a matching executor is available",
		"retains deduplicated source versions with _valid_from, _valid_to, and _is_current fields",
		"refuses before source i/o for incompatible pairs",
		"incremental modes require --cursor", "deduped modes require --primary-key",
	} {
		if !strings.Contains(s, required) {
			return false
		}
	}
	return !strings.Contains(s, "every mode executes unconditionally")
}
func TestHelpContractFalsifiers171(t *testing.T) {
	for _, tc := range []struct {
		name      string
		args      []string
		oracle    func(string) bool
		mutations [][2]string
	}{
		{"polling", []string{"connectors"}, pollingHelpContract171, [][2]string{{"static binding alone", "binding alone"}, {"implemented binding per selected catalog object", "unimplemented binding per selected catalog object"}, {"only after the same runtime preflight succeeds", "without runtime preflight"}}},
		{"etl", []string{"help", "etl"}, etlHelpContract171, [][2]string{{"incremental_dedupe_history", "removed_mode"}, {"refuses before source i/o until a matching executor is available", "executes unconditionally"}, {"refuses before source i/o for incompatible pairs", "executes incompatible pairs"}}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var out, errOut bytes.Buffer
			if code := cli.Run(tc.args, &out, &errOut); code != 0 {
				t.Fatalf("help: %s", errOut.String())
			}
			text := normalizedHelp171(out.String())
			if !tc.oracle(text) {
				t.Fatal("valid actual help rejected")
			}
			for _, mutation := range tc.mutations {
				if !strings.Contains(text, mutation[0]) {
					t.Fatalf("falsifier target missing: %q", mutation[0])
				}
				changed := strings.ReplaceAll(text, mutation[0], mutation[1])
				if tc.oracle(changed) {
					t.Fatalf("oracle accepted falsified contract: %q", mutation[0])
				}
			}
		})
	}
	for spelling, want := range map[string]synccontract.Mode{"full_refresh_overwrite_deduped": synccontract.ModeFullOverwrite, "full_refresh_overwrite_dedup": synccontract.ModeFullOverwrite, "full_refresh_deduped": synccontract.ModeFullOverwrite, "incremental_append_deduped": synccontract.ModeIncrementalDedupe, "incremental_append_dedup": synccontract.ModeIncrementalDedupe} {
		got, ok := synccontract.LookupPublicMode(spelling)
		if !ok || got.ContractMode != want || !got.TypedOnly || !got.RequiresCursor || !got.RequiresPrimaryKey {
			t.Fatalf("alias %q lost typed admission contract", spelling)
		}
	}
}
