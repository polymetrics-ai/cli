package rssparity_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/defs"
	"polymetrics.ai/internal/connectors/engine"
	_ "polymetrics.ai/internal/connectors/hooks/rss" // registers the StreamHook/CheckHook via init()
	"polymetrics.ai/internal/connectors/rss"
)

// This file is the migration parity suite for the rss bundle: rss is a
// Tier-2 StreamHook/CheckHook migration (docs/migration/conventions.md §1/
// §6) because every response is RSS/XML, which the engine's declarative
// read/check path cannot decode (it only ever decodes JSON). Both the
// legacy hand-written rss.Connector (internal/connectors/rss, read-only
// reference) and the engine-backed connector
// (engine.New(bundle, engine.HooksFor("rss"))) are driven against the SAME
// httptest.Server serving the identical RSS/XML fixture; RAW
// reflect.DeepEqual record equality (after JSON-canonicalization) is the
// parity bar.

const fixtureFeed = `<?xml version="1.0"?><rss version="2.0"><channel><title>Fixture Feed</title><link>https://example.test</link><description>A fixture RSS channel</description><lastBuildDate>Mon, 02 Jan 2006 15:04:05 MST</lastBuildDate><item><guid>one</guid><title>First item</title><link>https://example.test/one</link><description>First item description</description><pubDate>Mon, 02 Jan 2006 15:04:05 MST</pubDate></item><item><guid>two</guid><title>Second item</title><link>https://example.test/two</link></item></channel></rss>`

func loadRSSBundle(t *testing.T) engine.Bundle {
	t.Helper()
	bundles, _ := engine.LoadAll(defs.FS)
	for _, b := range bundles {
		if b.Name == "rss" {
			return b
		}
	}
	names := make([]string, 0, len(bundles))
	for _, b := range bundles {
		names = append(names, b.Name)
	}
	t.Fatalf("bundle %q not found in defs.FS (bundles: %v)", "rss", names)
	return engine.Bundle{}
}

func withRSSBaseURL(b engine.Bundle, baseURL string) engine.Bundle {
	b.HTTP.URL = baseURL
	return b
}

func newRSSEngineConnector(b engine.Bundle) connectors.Connector {
	return engine.New(b, engine.HooksFor("rss"))
}

func rssRuntimeConfig(feedURL string) connectors.RuntimeConfig {
	return connectors.RuntimeConfig{Config: map[string]string{"feed_url": feedURL}}
}

func readAllRSSRecords(t *testing.T, c connectors.Connector, req connectors.ReadRequest) []connectors.Record {
	t.Helper()
	var out []connectors.Record
	if err := c.Read(context.Background(), req, func(r connectors.Record) error {
		out = append(out, r)
		return nil
	}); err != nil {
		t.Fatalf("Read(%s): %v", req.Stream, err)
	}
	return out
}

// normalizeRSSRecord re-encodes r through encoding/json so both connectors
// compare on canonical JSON shape rather than incidental Go type identity.
func normalizeRSSRecord(t *testing.T, r connectors.Record) map[string]any {
	t.Helper()
	raw, err := json.Marshal(map[string]any(r))
	if err != nil {
		t.Fatalf("marshal record: %v", err)
	}
	var out map[string]any
	if err := json.Unmarshal(raw, &out); err != nil {
		t.Fatalf("unmarshal record: %v", err)
	}
	return out
}

func normalizeRSSRecords(t *testing.T, recs []connectors.Record) []map[string]any {
	t.Helper()
	out := make([]map[string]any, len(recs))
	for i, r := range recs {
		out[i] = normalizeRSSRecord(t, r)
	}
	return out
}

func rssFixtureServer(t *testing.T, sawAuth *bool) *httptest.Server {
	t.Helper()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if sawAuth != nil && r.Header.Get("Authorization") != "" {
			*sawAuth = true
		}
		w.Header().Set("Content-Type", "application/rss+xml")
		_, _ = w.Write([]byte(fixtureFeed))
	}))
	t.Cleanup(srv.Close)
	return srv
}

func TestParityRSS_ItemsStreamRecordsAndNoAuth(t *testing.T) {
	bundle := loadRSSBundle(t)

	var legacyAuth bool
	legacySrv := rssFixtureServer(t, &legacyAuth)
	legacy := rss.New()
	legacyRecs := readAllRSSRecords(t, legacy, connectors.ReadRequest{
		Stream: "items",
		Config: rssRuntimeConfig(legacySrv.URL),
	})
	if len(legacyRecs) != 2 {
		t.Fatalf("legacy records = %d, want 2 (test fixture bug)", len(legacyRecs))
	}
	if legacyAuth {
		t.Fatal("legacy sent credentials (test fixture bug)")
	}

	var engAuth bool
	engSrv := rssFixtureServer(t, &engAuth)
	eng := newRSSEngineConnector(withRSSBaseURL(bundle, engSrv.URL))
	engRecs := readAllRSSRecords(t, eng, connectors.ReadRequest{
		Stream: "items",
		Config: rssRuntimeConfig(engSrv.URL),
	})

	if engAuth {
		t.Fatal("engine-backed rss connector sent credentials")
	}
	if len(engRecs) != len(legacyRecs) {
		t.Fatalf("record count = %d, want %d (legacy)\nengine: %+v\nlegacy: %+v", len(engRecs), len(legacyRecs), engRecs, legacyRecs)
	}

	gotNorm := normalizeRSSRecords(t, engRecs)
	wantNorm := normalizeRSSRecords(t, legacyRecs)
	for i := range wantNorm {
		if !reflect.DeepEqual(gotNorm[i], wantNorm[i]) {
			t.Fatalf("record %d mismatch:\nengine:  %+v\nlegacy:  %+v", i, gotNorm[i], wantNorm[i])
		}
	}
}

func TestParityRSS_ChannelStream(t *testing.T) {
	bundle := loadRSSBundle(t)

	legacySrv := rssFixtureServer(t, nil)
	legacy := rss.New()
	legacyRecs := readAllRSSRecords(t, legacy, connectors.ReadRequest{
		Stream: "channel",
		Config: rssRuntimeConfig(legacySrv.URL),
	})

	engSrv := rssFixtureServer(t, nil)
	eng := newRSSEngineConnector(withRSSBaseURL(bundle, engSrv.URL))
	engRecs := readAllRSSRecords(t, eng, connectors.ReadRequest{
		Stream: "channel",
		Config: rssRuntimeConfig(engSrv.URL),
	})

	if len(engRecs) != len(legacyRecs) {
		t.Fatalf("record count = %d, want %d (legacy)\nengine: %+v\nlegacy: %+v", len(engRecs), len(legacyRecs), engRecs, legacyRecs)
	}

	gotNorm := normalizeRSSRecords(t, engRecs)
	wantNorm := normalizeRSSRecords(t, legacyRecs)
	for i := range wantNorm {
		if !reflect.DeepEqual(gotNorm[i], wantNorm[i]) {
			t.Fatalf("record %d mismatch:\nengine:  %+v\nlegacy:  %+v", i, gotNorm[i], wantNorm[i])
		}
	}
}

func TestParityRSS_CheckSucceedsAgainstValidFeed(t *testing.T) {
	bundle := loadRSSBundle(t)

	legacySrv := rssFixtureServer(t, nil)
	legacy := rss.New()
	if err := legacy.Check(context.Background(), rssRuntimeConfig(legacySrv.URL)); err != nil {
		t.Fatalf("legacy Check: %v", err)
	}

	engSrv := rssFixtureServer(t, nil)
	eng := newRSSEngineConnector(withRSSBaseURL(bundle, engSrv.URL))
	if err := eng.Check(context.Background(), rssRuntimeConfig(engSrv.URL)); err != nil {
		t.Fatalf("engine Check: %v", err)
	}
}

func TestParityRSS_BundleLoadsAndValidates(t *testing.T) {
	bundle := loadRSSBundle(t)
	if bundle.Name != "rss" {
		t.Fatalf("bundle.Name = %q, want rss", bundle.Name)
	}
	if !bundle.Metadata.Capabilities.Read || bundle.Metadata.Capabilities.Write {
		t.Fatalf("unexpected capabilities: %+v", bundle.Metadata.Capabilities)
	}
	if bundle.Metadata.Conformance == nil || !bundle.Metadata.Conformance.SkipDynamic {
		t.Fatal("bundle metadata.json should carry a bundle-level conformance.skip_dynamic marker")
	}
}
