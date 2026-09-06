package main

import (
	"encoding/json"
	"testing"
)

func requireSourceLane(t *testing.T, cells []sourceLaneCell, lane, applicability string) sourceLaneCell {
	t.Helper()
	if len(cells) != 7 {
		t.Fatalf("got %d lane cells, want exactly seven", len(cells))
	}
	for _, cell := range cells {
		if cell.Lane == lane {
			if cell.Applicability != applicability {
				t.Fatalf("%s applicability=%s want %s; %+v", lane, cell.Applicability, applicability, cell)
			}
			return cell
		}
	}
	t.Fatalf("missing lane %s", lane)
	return sourceLaneCell{}
}

func TestSourceLaneVercelPostRead(t *testing.T) {
	row, doc := retainedFactFixture(t, "vercel", "vercel.rest.readSessionFile")
	facts := normalizeSourceFacts(row, doc, nil)
	var summary string
	if err := json.Unmarshal(facts.Groups["summary"], &summary); err != nil {
		t.Fatal(err)
	}
	a := sourceSemanticAnnotation{Key: row.Key, Semantics: "read", Citation: facts.Refs["summary"], Clause: summary}
	cells := classifySourceLanes(row.Key, facts, &a)
	for _, lane := range []string{"direct_read", "binary_download"} {
		requireSourceLane(t, cells, lane, "applicable")
	}
	for _, lane := range []string{"direct_write", "binary_upload", "etl", "reverse_etl", "sync_transport"} {
		requireSourceLane(t, cells, lane, "not_applicable")
	}
}

func TestSourceLaneCollectionAndMutation(t *testing.T) {
	for _, tc := range []struct {
		name, method, summary, response string
		read, write, etl, reverse       string
	}{
		{"finite collection", "GET", "List widgets", `{"type":"array","items":{"type":"object","properties":{"id":{"type":"string"}}}}`, "applicable", "not_applicable", "applicable", "not_applicable"},
		{"mutation", "POST", "Create widget", `{"type":"object","properties":{"id":{"type":"string"}}}`, "not_applicable", "applicable", "not_applicable", "applicable"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			node := json.RawMessage(`{"id":"fixture","protocol":"rest","method":"` + tc.method + `","path":"/widgets","source_operation":{"summary":"` + tc.summary + `","responses":{"200":{"content":{"application/json":{"schema":` + tc.response + `}}}}}}`)
			row := retainedSourceOperation{Key: sourceOperationKey{Connector: "fixture", Inventory: "primary", ID: "fixture"}, Observed: true, Node: node, Pointer: "/rest/operations/0"}
			doc := retainedSourceDocument{ID: "fixture", Payload: json.RawMessage(`{"rest":{"operations":[` + string(node) + `]}}`)}
			facts := normalizeSourceFacts(row, doc, nil)
			cells := classifySourceLanes(row.Key, facts, nil)
			requireSourceLane(t, cells, "direct_read", tc.read)
			requireSourceLane(t, cells, "direct_write", tc.write)
			requireSourceLane(t, cells, "etl", tc.etl)
			requireSourceLane(t, cells, "reverse_etl", tc.reverse)
		})
	}
}

func TestSourceLaneRuleCounterexample(t *testing.T) {
	row, doc := retainedFactFixture(t, "vercel", "vercel.rest.readSessionFile")
	facts := normalizeSourceFacts(row, doc, nil)
	var clause string
	if err := json.Unmarshal(facts.Groups["summary"], &clause); err != nil {
		t.Fatal(err)
	}
	a := sourceSemanticAnnotation{Key: row.Key, Semantics: "read", Citation: facts.Refs["summary"], Clause: clause}
	cells := classifySourceLanes(row.Key, facts, &a)
	want := map[string]string{"direct_read": "applicable", "direct_write": "not_applicable", "binary_download": "applicable", "binary_upload": "not_applicable", "etl": "not_applicable", "reverse_etl": "not_applicable", "sync_transport": "not_applicable"}
	if !sourceLaneOracleMatches(cells, want) {
		t.Fatal("valid independently specified seven-lane control rejected")
	}
	wrong := append([]sourceLaneCell(nil), cells...)
	wrong[1].Applicability = "applicable"
	if sourceLaneOracleMatches(wrong, want) {
		t.Fatal("oracle accepts POST read misclassified as mutation")
	}
	wrong = append([]sourceLaneCell(nil), cells...)
	wrong[2].Applicability = "not_applicable"
	if sourceLaneOracleMatches(wrong, want) {
		t.Fatal("oracle accepts missing binary applicability")
	}
	if sourceLaneOracleMatches(cells[:6], want) {
		t.Fatal("oracle accepts omitted lane")
	}
	a.Citation.ValueSHA256 = "wrong-readable-claim"
	invalid := classifySourceLanes(row.Key, facts, &a)
	if len(invalid[0].Diagnostics) == 0 || invalid[0].Diagnostics[0].Code != "source_annotation_invalid" {
		t.Fatal("changed citation accepted")
	}
}

func sourceLaneOracleMatches(cells []sourceLaneCell, expected map[string]string) bool {
	if len(cells) != 7 || len(expected) != 7 {
		return false
	}
	seen := map[string]bool{}
	for _, cell := range cells {
		if seen[cell.Lane] || expected[cell.Lane] != cell.Applicability {
			return false
		}
		seen[cell.Lane] = true
	}
	return true
}

func TestSourceLanePostReadNotion(t *testing.T) {
	row, doc := retainedFactFixture(t, "notion", "notion.rest.post-database-query")
	facts := normalizeSourceFacts(row, doc, nil)
	var summary string
	if err := json.Unmarshal(facts.Groups["summary"], &summary); err != nil {
		t.Fatal(err)
	}
	a := sourceSemanticAnnotation{Key: row.Key, Semantics: "read", Citation: facts.Refs["summary"], Clause: summary}
	cells := classifySourceLanes(row.Key, facts, &a)
	requireSourceLane(t, cells, "direct_read", "applicable")
	requireSourceLane(t, cells, "direct_write", "not_applicable")
}

func TestSourceLaneUnknownMediaAndPagingNotSync(t *testing.T) {
	node := json.RawMessage(`{"id":"fixture","protocol":"rest","method":"GET","path":"/widgets","source_operation":{"summary":"List widgets","parameters":[{"in":"query","name":"cursor","schema":{"type":"string"}}],"responses":{"200":{"description":"unknown media"}}}}`)
	row := retainedSourceOperation{Key: sourceOperationKey{Connector: "fixture", Inventory: "primary", ID: "fixture"}, Observed: true, Node: node, Pointer: "/rest/operations/0"}
	doc := retainedSourceDocument{ID: "fixture", Payload: json.RawMessage(`{"rest":{"operations":[` + string(node) + `]}}`)}
	facts := normalizeSourceFacts(row, doc, nil)
	cells := classifySourceLanes(row.Key, facts, nil)
	requireSourceLane(t, cells, "sync_transport", "not_applicable")
	// A response with an entirely absent content declaration cannot prove that
	// a body-bearing 200 response has no binary/record representation.
	requireSourceLane(t, cells, "binary_download", "undetermined")
}

func TestSourceLaneBinaryLocalReference(t *testing.T) {
	facts := sourceFacts{Document: json.RawMessage(`{"components":{"schemas":{"File":{"type":"string","format":"binary"}}}}`), Groups: map[string]json.RawMessage{"request_body": json.RawMessage(`{"content":{"multipart/form-data":{"schema":{"type":"object","properties":{"file":{"$ref":"#/components/schemas/File"}}}}}}`)}}
	got := sourceRequestShape(facts)
	if !got.Binary || !got.Known {
		t.Fatalf("local binary part not recognized: %+v", got)
	}
	facts.Groups["request_body"] = json.RawMessage(`{"content":{"application/json":{"schema":{"$ref":"https://not-fetched.example/schema"}}}}`)
	got = sourceRequestShape(facts)
	if got.Known || got.Binary {
		t.Fatalf("external schema became an applicability proof: %+v", got)
	}
}

func TestSourceLaneAnnotationContradiction(t *testing.T) {
	for _, tc := range []struct {
		name, method, summary, interpretation string
		valid                                 bool
	}{
		{"post read claimed mutation", "POST", "Read a file", "mutation", false},
		{"mutation mentions read target", "POST", "Create a read token", "read", false},
		{"substring is not read semantics", "POST", "Update the playlist", "read", false},
		{"negated mutation", "POST", "Do not create a widget", "mutation", false},
		{"positive post read", "POST", "Read a file", "read", true},
		{"positive mutation", "POST", "Create a widget", "mutation", true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			node, err := json.Marshal(map[string]any{"id": "fixture", "protocol": "rest", "method": tc.method, "path": "/widgets", "source_operation": map[string]any{"summary": tc.summary, "responses": map[string]any{"204": map[string]any{"description": "No content"}}}})
			if err != nil {
				t.Fatal(err)
			}
			row := retainedSourceOperation{Key: sourceOperationKey{Connector: "fixture", Inventory: "primary", ID: "fixture"}, Observed: true, Node: node, Pointer: "/rest/operations/0"}
			doc := retainedSourceDocument{ID: "fixture", Payload: json.RawMessage(`{"rest":{"operations":[` + string(node) + `]}}`)}
			facts := normalizeSourceFacts(row, doc, nil)
			annotation := sourceSemanticAnnotation{Key: row.Key, Semantics: tc.interpretation, Citation: facts.Refs["summary"], Clause: tc.summary}
			cells := classifySourceLanes(row.Key, facts, &annotation)
			invalid := false
			for _, cell := range cells {
				for _, diagnostic := range cell.Diagnostics {
					if diagnostic.Code == "source_annotation_invalid" {
						invalid = true
					}
				}
			}
			if invalid == tc.valid {
				t.Fatalf("annotation %q for cited %q: invalid=%v, want %v; cells=%+v", tc.interpretation, tc.summary, invalid, !tc.valid, cells)
			}
			if tc.valid {
				lane := "direct_read"
				if tc.interpretation == "mutation" {
					lane = "direct_write"
				}
				requireSourceLane(t, cells, lane, "applicable")
			}
		})
	}
}
