package rlm

import (
	"strings"
	"testing"
)

func TestParseSpecValid(t *testing.T) {
	data := []byte(`{
		"name": "likely-customers",
		"features": [
			{"name": "email", "weight": 0.5, "score_if_set": 1.0},
			{"name": "company", "weight": 0.5, "score_if_set": 1.0}
		]
	}`)
	spec, err := ParseSpec(data)
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if spec == nil {
		t.Fatal("expected non-nil spec")
	}
	if spec.Name != "likely-customers" {
		t.Errorf("name = %q, want %q", spec.Name, "likely-customers")
	}
	if len(spec.Features) != 2 {
		t.Errorf("features count = %d, want 2", len(spec.Features))
	}
}

func TestParseSpecMissingName(t *testing.T) {
	data := []byte(`{
		"features": [
			{"name": "email", "weight": 0.5, "score_if_set": 1.0}
		]
	}`)
	_, err := ParseSpec(data)
	if err == nil {
		t.Fatal("expected error for missing name, got nil")
	}
	if !strings.Contains(err.Error(), "name") {
		t.Errorf("error %q should mention 'name'", err.Error())
	}
}

func TestParseSpecEmptyFeatures(t *testing.T) {
	data := []byte(`{"name": "test", "features": []}`)
	_, err := ParseSpec(data)
	if err == nil {
		t.Fatal("expected validation error for empty features, got nil")
	}
}

func TestParseSpecNegativeWeight(t *testing.T) {
	data := []byte(`{
		"name": "test",
		"features": [
			{"name": "email", "weight": -1.0, "score_if_set": 1.0}
		]
	}`)
	_, err := ParseSpec(data)
	if err == nil {
		t.Fatal("expected error for negative weight, got nil")
	}
	if !strings.Contains(err.Error(), "weight") {
		t.Errorf("error %q should mention 'weight'", err.Error())
	}
}

func TestParseSpecZeroWeight(t *testing.T) {
	// Zero weight is allowed (contributes no score but is not invalid).
	data := []byte(`{
		"name": "test",
		"features": [
			{"name": "email", "weight": 0.0, "score_if_set": 1.0}
		]
	}`)
	_, err := ParseSpec(data)
	if err != nil {
		t.Fatalf("zero weight should be allowed, got error: %v", err)
	}
}

func TestParseSpecScoreIfGTMissingThreshold(t *testing.T) {
	scoreVal := 1.0
	_ = scoreVal // used in JSON below
	data := []byte(`{
		"name": "test",
		"features": [
			{"name": "amount", "weight": 1.0, "score_if_gt": 1.0}
		]
	}`)
	// score_if_gt set but threshold absent → validation error
	_, err := ParseSpec(data)
	if err == nil {
		t.Fatal("expected validation error when score_if_gt set without threshold, got nil")
	}
}
