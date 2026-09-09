package cli_test

import (
	"bytes"
	"polymetrics.ai/internal/cli"
	"strings"
	"testing"
)

func TestHelpOracleCompleteness173(t *testing.T) {
	var out, errOut bytes.Buffer
	if code := cli.Run([]string{"help", "etl"}, &out, &errOut); code != 0 {
		t.Fatalf("healthy help failed: %s", errOut.String())
	}
	original := out.String()
	if !etlHelpContract171(original) {
		t.Fatal("healthy actual help rejected")
	}
	for _, mode := range []string{"full_refresh_overwrite", "incremental_append", "incremental_dedupe"} {
		t.Run("remove_"+mode, func(t *testing.T) {
			start := strings.Index(original, "\n  "+mode+"\n")
			if start < 0 {
				t.Fatal("mode heading not reached")
			}
			tail := original[start+1:]
			end := strings.Index(tail, "\n\n")
			if end < 0 {
				t.Fatal("mode block end not reached")
			}
			changed := original[:start] + tail[end:]
			if strings.Contains(changed, "\n  "+mode+"\n") {
				t.Fatal("mode removal setup failed")
			}
			if etlHelpContract171(changed) {
				t.Fatalf("oracle accepts removal of complete %s heading and description", mode)
			}
		})
	}
	t.Run("misleading_overwrite", func(t *testing.T) {
		old := "Replaces the write-ahead log with this run's records, then atomically\n    replaces the final Parquet table only after the run succeeds."
		if !strings.Contains(original, old) {
			t.Fatal("actual overwrite description not reached")
		}
		changed := strings.Replace(original, old, "Appends every record and keeps all existing rows on every run.", 1)
		if etlHelpContract171(changed) {
			t.Fatal("oracle accepts append semantics for full_refresh_overwrite")
		}
	})
}

func TestEveryModeHelpContractFalsified173(t *testing.T) {
	var out, errOut bytes.Buffer
	if code := cli.Run([]string{"help", "etl"}, &out, &errOut); code != 0 {
		t.Fatalf("actual help: %s", errOut.String())
	}
	original := out.String()
	if !etlHelpContract171(original) {
		t.Fatal("healthy help rejected")
	}
	cases := []struct{ mode, meaning, falseMeaning string }{
		{"full_refresh_append", "Duplicates across runs are expected.", "Every repeated row is discarded automatically."},
		{"full_refresh_overwrite", "only after the run succeeds.", "before the run starts."},
		{"full_refresh_overwrite_deduped", "refuses before source I/O until a matching executor is available.", "executes with no matching executor."},
		{"incremental_append", "Cursor state advances only after successful writes.", "Cursor state advances before writes."},
		{"incremental_append_deduped", "deduplicates by the declared primary key.", "keeps all duplicate primary keys."},
		{"incremental_dedupe", "Retains one current record per declared primary key.", "Retains every historical record for each key."},
		{"incremental_dedupe_history", "_valid_from, _valid_to, and\n    _is_current fields.", "no version interval or current-state fields."},
	}
	for _, tc := range cases {
		t.Run(tc.mode, func(t *testing.T) {
			start := strings.Index(original, "\n  "+tc.mode+"\n")
			if start < 0 {
				t.Fatal("exact heading absent in healthy output")
			}
			offset := strings.Index(original[start+1:], "\n\n")
			if offset < 0 {
				t.Fatal("section end not reached")
			}
			end := start + 1 + offset
			section := original[start:end]
			if !strings.Contains(section, tc.meaning) {
				t.Fatalf("independent semantic mutation not reached: %q", tc.meaning)
			}
			without := original[:start] + original[end:]
			if etlHelpContract171(without) {
				t.Fatal("missing exact mode accepted")
			}
			changed := original[:start] + strings.Replace(section, tc.meaning, tc.falseMeaning, 1) + original[end:]
			if etlHelpContract171(changed) {
				t.Fatal("false mode semantics accepted")
			}
		})
	}
	if !etlHelpContract171(strings.ToUpper(original)) {
		t.Fatal("harmless case variation refused")
	}
	if etlHelpContract171(original + "\n  incremental_append\n    ambiguous duplicate\n") {
		t.Fatal("duplicate heading accepted")
	}
}
