package connectors

import "testing"

func TestBareName(t *testing.T) {
	cases := map[string]string{
		"source-github":        "github",
		"destination-bigquery": "bigquery",
		"github":               "github",
		"Source-GitHub":        "github",
		"source-google-sheets": "google-sheets",
		"destination-pinecone": "pinecone",
		"":                     "",
	}
	for in, want := range cases {
		if got := BareName(in); got != want {
			t.Errorf("BareName(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestDirectionForSlug(t *testing.T) {
	cases := map[string]Direction{
		"source-github":        DirectionSource,
		"destination-bigquery": DirectionDestination,
		"github":               "",
	}
	for in, want := range cases {
		if got := DirectionForSlug(in); got != want {
			t.Errorf("DirectionForSlug(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestConnectorDefinitionBySlugResolvesBareName(t *testing.T) {
	// Legacy slug still resolves exactly.
	if _, ok := ConnectorDefinitionBySlug("source-github"); !ok {
		t.Fatal("source-github should resolve (exact)")
	}
	// Bare name resolves to the same system.
	def, ok := ConnectorDefinitionBySlug("github")
	if !ok {
		t.Fatal("bare name github should resolve")
	}
	if BareName(def.Slug) != "github" {
		t.Fatalf("resolved slug = %q, want bare github", def.Slug)
	}
	// Unknown name does not resolve.
	if _, ok := ConnectorDefinitionBySlug("definitely-not-a-connector-xyz"); ok {
		t.Fatal("unknown name should not resolve")
	}
}

func TestUnifyPairDetection(t *testing.T) {
	// postgres exists as both a source and a destination -> bidirectional.
	defs := ConnectorDefinitionsByBareName("postgres")
	if len(defs) < 2 {
		t.Fatalf("postgres entries = %d, want >= 2 (source+destination)", len(defs))
	}
	if dir := CanonicalDirection("postgres"); dir != DirectionBidirectional {
		t.Fatalf("CanonicalDirection(postgres) = %q, want bidirectional", dir)
	}
	// A pure source resolves as source-only.
	if dir := CanonicalDirection("stripe"); dir != DirectionSource {
		t.Fatalf("CanonicalDirection(stripe) = %q, want source", dir)
	}
}

func TestUnifyPairsMatchKnownSet(t *testing.T) {
	// The 21 systems that appear as both source and destination, per the plan.
	want := []string{
		"azure-blob-storage", "bigquery", "clickhouse", "convex", "customer-io",
		"dynamodb", "elasticsearch", "firebolt", "gcs", "google-sheets", "hubspot",
		"kafka", "mssql", "mysql", "oracle", "postgres", "redshift", "s3",
		"salesforce", "snowflake", "teradata",
	}
	for _, name := range want {
		if dir := CanonicalDirection(name); dir != DirectionBidirectional {
			t.Errorf("CanonicalDirection(%q) = %q, want bidirectional", name, dir)
		}
	}
}
