package cli

import (
	"bytes"
	"strings"
	"testing"
)

func TestETLManualAndSkillDescribeDeliveredReconciliationTerminalRun(t *testing.T) {
	var stdout, stderr bytes.Buffer
	if code := Run([]string{"etl"}, &stdout, &stderr); code != 0 {
		t.Fatalf("Run(etl) code = %d stderr = %s", code, stderr.String())
	}
	manual := stdout.String()
	for _, want := range []string{
		"delivered_reconciliation_required",
		"exact terminal ETLRun",
		"before endpoint resolution",
		"never replays source or destination I/O",
	} {
		if !strings.Contains(manual, want) {
			t.Fatalf("ETL manual omitted durable reconciliation contract %q:\n%s", want, manual)
		}
	}

	stdout.Reset()
	stderr.Reset()
	if code := Run([]string{"etl", "transport", "declarative-typed-destination"}, &stdout, &stderr); code != 0 {
		t.Fatalf("Run(etl transport declarative-typed-destination) code = %d stderr = %s", code, stderr.String())
	}
	transportHelp := stdout.String()
	for _, want := range []string{
		"delivered_reconciliation_required",
		"exact terminal ETLRun",
		"before endpoint resolution",
		"never replays source or destination I/O",
	} {
		if !strings.Contains(transportHelp, want) {
			t.Fatalf("declarative transport help omitted durable reconciliation contract %q:\n%s", want, transportHelp)
		}
	}

	var etlSkill string
	for _, skill := range mustBaseSkillDocs(t) {
		if skill.Name == "pm-etl" {
			etlSkill = skill.Body
			break
		}
	}
	if etlSkill == "" {
		t.Fatal("pm-etl skill is missing")
	}
	for _, want := range []string{
		"delivered_reconciliation_required",
		"exact `ETLRun`",
		"before endpoint resolution",
		"Never replace it with a normal retry or a raw provider action.",
	} {
		if !strings.Contains(etlSkill, want) {
			t.Fatalf("pm-etl skill omitted durable reconciliation contract %q:\n%s", want, etlSkill)
		}
	}
}

func mustBaseSkillDocs(t *testing.T) []skillDoc {
	t.Helper()
	skills, err := baseSkillDocs(nil)
	if err != nil {
		t.Fatalf("baseSkillDocs() = %v", err)
	}
	return skills
}
