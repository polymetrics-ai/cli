package main

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func source118DemandAnnotation(key sourceOperationKey, facts sourceFacts) sourceSemanticAnnotation {
	return sourceSemanticAnnotation{Key: key, Citation: facts.Refs["summary"], Clause: sourceFactText(facts, "summary"), FoundationGap: "cli-webhook-event-surface-foundation-r1", AtlasID: "transport.sync-contract.v1", DecisionRefs: []string{"cli-batch1-vercel-inbound-sync-decision-r1", "cli-plan-webhook-receiver-foundation-r1-decision-webhook-exposure-product-boundary"}}
}

func TestSourceLane118DemandCannotBorrowAuthority(t *testing.T) {
	for _, tc := range []struct{ name, method, summary string }{
		{"read mention", "GET", "Get webhooks"},
		{"list mention", "GET", "List webhooks"},
		{"negated", "POST", "Does not create a webhook"},
		{"incidental", "POST", "Create a widget without a webhook"},
		{"unrelated registration", "POST", "Create a webhook"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			node, err := json.Marshal(map[string]any{"id": "fixture", "protocol": "rest", "method": tc.method, "path": "/widgets", "source_operation": map[string]any{"summary": tc.summary, "responses": map[string]any{"204": map[string]any{"description": "No content"}}}})
			if err != nil {
				t.Fatal(err)
			}
			row := retainedSourceOperation{Key: sourceOperationKey{Connector: "fixture", Inventory: "primary", ID: "fixture"}, Observed: true, Node: node, Pointer: "/rest/operations/0"}
			doc := retainedSourceDocument{ID: "fixture:primary", Payload: json.RawMessage(`{"rest":{"operations":[` + string(node) + `]}}`)}
			facts := normalizeSourceFacts(row, doc, nil)
			if facts.Status == "unavailable" || sourceFactText(facts, "summary") != tc.summary {
				t.Fatal("fixture failed retained normalization")
			}
			a := source118DemandAnnotation(row.Key, facts)
			cells := classifySourceLanes(row.Key, facts, &a)
			if len(cells) != 7 {
				t.Fatal("authority refusal lost lanes")
			}
			for _, cell := range cells {
				if cell.State == "missing_foundation" || cell.State == "implemented" {
					t.Fatalf("copied genuine labels gained authority: %+v", cell)
				}
			}
			if !proofHasDiagnostic(cells[6].Diagnostics, "source_annotation_invalid", "error") {
				t.Fatalf("asserted authority lacked explicit refusal: %+v", cells[6])
			}
		})
	}
}

func TestSourceLane118DemandActualBuilder(t *testing.T) {
	row, doc := retainedFactFixture(t, "vercel", "vercel.rest.createWebhook")
	facts := normalizeSourceFacts(row, doc, nil)
	var archived struct {
		Rest struct {
			Operations []struct {
				ID string `json:"id"`
			} `json:"operations"`
		} `json:"rest"`
	}
	if err := json.Unmarshal(doc.Payload, &archived); err != nil {
		t.Fatal(err)
	}
	ids := []string{}
	for _, r := range archived.Rest.Operations {
		ids = append(ids, r.ID)
	}
	repo, err := repoRoot()
	if err != nil {
		t.Fatal(err)
	}
	atlasPath := "docs/connector-canon/foundations/catalog.json"
	atlas, err := os.ReadFile(filepath.Join(repo, atlasPath))
	if err != nil {
		t.Fatal(err)
	}
	for _, which := range []string{"valid", "absent annotation", "missing Atlas", "wrong Atlas", "unknown owner", "missing owner", "duplicate owner", "extra owner", "wrong gap", "wrong claimed Atlas", "stale source bytes", "wrong retained path", "changed events", "changed route", "changed provider operation"} {
		t.Run(which, func(t *testing.T) {
			root := t.TempDir()
			fixtureDoc := doc
			fixtureFacts := facts
			if which == "stale source bytes" {
				fixtureDoc.Payload = append(append(json.RawMessage(nil), doc.Payload...), '\n')
			}
			if which == "wrong retained path" {
				fixtureDoc.Path = "source-other.json"
			}
			if which == "changed events" || which == "changed route" || which == "changed provider operation" {
				var whole map[string]any
				if err := json.Unmarshal(doc.Payload, &whole); err != nil {
					t.Fatal(err)
				}
				node := whole["rest"].(map[string]any)["operations"].([]any)[381].(map[string]any)
				switch which {
				case "changed route":
					node["path"] = "/v2/webhooks"
				case "changed provider operation":
					node["operation_id"] = "other"
				case "changed events":
					node["source_operation"].(map[string]any)["requestBody"] = map[string]any{"content": map[string]any{"application/json": map[string]any{"schema": map[string]any{"type": "object"}}}}
				}
				var err error
				fixtureDoc.Payload, err = json.Marshal(whole)
				if err != nil {
					t.Fatal(err)
				}
				changedRow := row
				changedRow.Node, err = json.Marshal(node)
				if err != nil {
					t.Fatal(err)
				}
				fixtureFacts = normalizeSourceFacts(changedRow, fixtureDoc, nil)
			}
			fixtureDoc.RetainedFileSHA256 = sourceBytesHash(fixtureDoc.Payload)
			proofWrite(t, root, fixtureDoc.Path, fixtureDoc.Payload)
			if which != "missing Atlas" {
				body := atlas
				if which == "wrong Atlas" {
					body = []byte(`{"foundations":[]}`)
				}
				proofWrite(t, root, atlasPath, body)
			}
			a := source118DemandAnnotation(row.Key, fixtureFacts)
			switch which {
			case "unknown owner":
				a.DecisionRefs[0] = "unknown"
			case "missing owner":
				a.DecisionRefs = a.DecisionRefs[:1]
			case "duplicate owner":
				a.DecisionRefs = []string{a.DecisionRefs[0], a.DecisionRefs[0]}
			case "extra owner":
				a.DecisionRefs = append(a.DecisionRefs, "unknown")
			case "wrong gap":
				a.FoundationGap = "other"
			case "wrong claimed Atlas":
				a.AtlasID = "other"
			}
			cohort := sourceLaneCohort{SchemaVersion: 1, CohortID: "fixture", Inventories: []sourceInventoryAnchor{{Connector: "vercel", Inventory: "primary", Class: "primary", Path: fixtureDoc.Path, SHA256: fixtureDoc.RetainedFileSHA256, ExpectedIDs: ids, ExpectedCount: len(ids)}}}
			annotations := []sourceSemanticAnnotation{a}
			if which == "absent annotation" {
				annotations = nil
			}
			got, err := buildSourceLaneManifest(context.Background(), root, cohort, annotations)
			if err != nil {
				t.Fatal(err)
			}
			if len(got.SourceOperations) != 400 || got.SourceTotals.Cells != 2800 {
				t.Fatalf("authority changed retained membership: %+v", got.SourceTotals)
			}
			atlasPins := 0
			for _, pin := range got.Inputs {
				if pin.Path == atlasPath {
					atlasPins++
					if pin.SHA256 != sourceBytesHash(atlas) && which != "wrong Atlas" {
						t.Fatal("Atlas pin does not identify consumed bytes")
					}
				}
			}
			wantPins := 1
			if which == "missing Atlas" || which == "absent annotation" {
				wantPins = 0
			}
			if atlasPins != wantPins {
				t.Fatalf("Atlas consumed pin count=%d want %d", atlasPins, wantPins)
			}
			if which == "valid" && got.Validation.Status != "valid" {
				t.Fatalf("valid demand fixture rejected before authority assertion: %+v", got.Validation)
			}
			found := false
			for _, r := range got.SourceOperations {
				if r.Source.Key == row.Key {
					found = true
					for _, c := range r.Lanes {
						if c.Lane == "sync_transport" {
							if (c.State == "missing_foundation") != (which == "valid") {
								t.Fatalf("authority %s: %+v", which, c)
							}
							if c.State == "implemented" || len(c.ProofRefs) != 0 {
								t.Fatal("demand became executable proof")
							}
						}
					}
				}
			}
			if !found {
				t.Fatal("retained operation absent")
			}
		})
	}
}
