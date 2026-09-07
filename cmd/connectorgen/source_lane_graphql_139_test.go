package main

import (
	"encoding/json"
	"os"
	"testing"
)

func TestSourceLane139GraphQLSchemaReader(t *testing.T) {
	raw, err := os.ReadFile("testdata/source-lane-graphql-139.json")
	if err != nil {
		t.Fatal(err)
	}
	var cases []struct {
		Name       string          `json:"name"`
		Annotation json.RawMessage `json:"annotation"`
		Valid      bool            `json:"valid"`
	}
	if err := decodeStrictJSON(raw, &cases); err != nil {
		t.Fatal(err)
	}
	if len(cases) != 11 {
		t.Fatalf("expected11 literal cases, got%d", len(cases))
	}
	for _, tc := range cases {
		t.Run(tc.Name, func(t *testing.T) {
			var annotation sourceSemanticAnnotation
			err := decodeStrictJSON(tc.Annotation, &annotation)
			valid := err == nil && sourceLaneGraphQLRefsShape(annotation.GraphQL)
			if valid != tc.Valid {
				t.Fatalf("actual annotation reader valid=%v want%v: %v", valid, tc.Valid, err)
			}
		})
	}
}
