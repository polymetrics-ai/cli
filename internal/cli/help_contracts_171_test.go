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

// Each key is an exact standalone heading, never a substring of another mode.
// Expectations specify the public semantics independently of docs.go.
var etlModeContracts173 = map[string][]string{
	"full_refresh_append":            {"reads every source record", "appends to the write-ahead log", "rebuilds the final parquet table", "duplicates across runs are expected"},
	"full_refresh_overwrite":         {"replaces the write-ahead log", "this run's records", "atomically replaces the final parquet table only after the run succeeds"},
	"full_refresh_overwrite_deduped": {"replaces the output", "deduplicates by the declared primary key", "refuses before source i/o until a matching executor is available"},
	"incremental_append":             {"at or after the saved cursor", "appends accepted records to the write-ahead log", "cursor state advances only after successful writes"},
	"incremental_append_deduped":     {"appends newer records", "deduplicates by the declared primary key", "refuses before source i/o until a matching executor is available"},
	"incremental_dedupe":             {"one current record per declared primary key", "refuses before source i/o for incompatible source and destination pairs"},
	"incremental_dedupe_history":     {"deduplicated source versions", "_valid_from, _valid_to, and _is_current", "requires primary-key and cursor fields", "refuses before source i/o for incompatible pairs"},
}

func etlHelpContract171(s string) bool {
	sections := make(map[string]string)
	active := ""
	for _, line := range strings.Split(s, "\n") {
		label := strings.ToLower(strings.TrimSpace(line))
		if _, mode := etlModeContracts173[label]; mode {
			if _, duplicate := sections[label]; duplicate {
				return false
			}
			active = label
			sections[active] = ""
			continue
		}
		if label == "" {
			active = ""
			continue
		}
		if active != "" {
			sections[active] += " " + label
		}
	}
	if len(sections) != len(etlModeContracts173) {
		return false
	}
	for mode, required := range etlModeContracts173 {
		block := normalizedHelp171(sections[mode])
		for _, meaning := range required {
			if !strings.Contains(block, meaning) {
				return false
			}
		}
	}
	normalized := normalizedHelp171(s)
	return strings.Contains(normalized, "incremental modes require --cursor") && strings.Contains(normalized, "deduped modes require --primary-key") && !strings.Contains(normalized, "every mode executes unconditionally")
}
func TestHelpContractFalsifiers171(t *testing.T) {
	for _, tc := range []struct {
		name      string
		args      []string
		oracle    func(string) bool
		mutations [][2]string
	}{
		{"polling", []string{"connectors"}, pollingHelpContract171, [][2]string{{"static binding alone", "binding alone"}, {"implemented binding per selected catalog object", "unimplemented binding per selected catalog object"}, {"only after the same runtime preflight succeeds", "without runtime preflight"}}},
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
