package manifestidentity

import (
	"testing"
	"testing/fstest"
)

func TestForFSIncludesRateLimitsInDigestAndCharge(t *testing.T) {
	withoutRate := fstest.MapFS{
		"alpha/metadata.json": &fstest.MapFile{Data: []byte(`{"name":"alpha"}`)},
	}
	first, err := ForFS(withoutRate, "alpha", EmbeddedGeneration)
	if err != nil {
		t.Fatal(err)
	}
	withRate := fstest.MapFS{
		"alpha/metadata.json":    &fstest.MapFile{Data: []byte(`{"name":"alpha"}`)},
		"alpha/rate_limits.json": &fstest.MapFile{Data: []byte(`{"state":"not_applicable"}`)},
	}
	second, err := ForFS(withRate, "alpha", EmbeddedGeneration)
	if err != nil {
		t.Fatal(err)
	}
	if first.Digest == second.Digest || first.Bytes >= second.Bytes {
		t.Fatalf("rate_limits identity = %#v, want digest and byte charge distinct from %#v", second, first)
	}
}

func TestForFSExcludesAuthoringArtifacts(t *testing.T) {
	base := fstest.MapFS{
		"alpha/metadata.json": &fstest.MapFile{Data: []byte(`{"name":"alpha"}`)},
	}
	withLock := fstest.MapFS{
		"alpha/metadata.json":    &fstest.MapFile{Data: []byte(`{"name":"alpha"}`)},
		"alpha/source.lock.json": &fstest.MapFile{Data: []byte(`{"provider_evidence":"authoring-only"}`)},
	}
	first, err := ForFS(base, "alpha", EmbeddedGeneration)
	if err != nil {
		t.Fatal(err)
	}
	second, err := ForFS(withLock, "alpha", EmbeddedGeneration)
	if err != nil {
		t.Fatal(err)
	}
	if first != second {
		t.Fatalf("authoring-only lock changed execution identity: first=%#v second=%#v", first, second)
	}
}
