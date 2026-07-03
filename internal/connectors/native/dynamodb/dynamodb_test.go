package dynamodb_test

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"

	"polymetrics.ai/internal/connectors"
	native "polymetrics.ai/internal/connectors/native/dynamodb"
)

func fixtureConfig() connectors.RuntimeConfig {
	return connectors.RuntimeConfig{Config: map[string]string{"mode": "fixture"}}
}

func TestNameAndMetadata(t *testing.T) {
	c := native.New()
	if c.Name() != "dynamodb" {
		t.Fatalf("Name() = %q, want dynamodb", c.Name())
	}
	caps := c.Metadata().Capabilities
	if !caps.Check || !caps.Catalog || !caps.Read {
		t.Fatalf("capabilities = %+v, want Check && Catalog && Read", caps)
	}
	if caps.Write {
		t.Fatal("dynamodb is read-only; Write capability must be false")
	}
}

// TestNoInitRegistration is the required grep-guard, mirroring
// native/postgres's and native/bing-ads's identical guard (T-17 precedent):
// the native package must NOT call RegisterFactory/RegisterNativeLive from
// anywhere in its own source. The registration flip is a later-wave change;
// this package only builds and tests standalone.
func TestNoInitRegistration(t *testing.T) {
	_, thisFile, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("runtime.Caller(0) failed; cannot locate package directory")
	}
	dir := filepath.Dir(thisFile)

	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatalf("ReadDir(%s): %v", dir, err)
	}

	found := false
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".go") {
			continue
		}
		if strings.HasSuffix(e.Name(), "_test.go") {
			continue
		}
		found = true
		raw, err := os.ReadFile(filepath.Join(dir, e.Name()))
		if err != nil {
			t.Fatalf("ReadFile(%s): %v", e.Name(), err)
		}
		src := string(raw)
		if strings.Contains(src, "RegisterFactory(") {
			t.Fatalf("%s calls RegisterFactory — native/dynamodb must NOT self-register (registration flip is a later wave)", e.Name())
		}
		if strings.Contains(src, "RegisterNativeLive(") {
			t.Fatalf("%s calls RegisterNativeLive — native/dynamodb must NOT self-register (registration flip is a later wave)", e.Name())
		}
		if strings.Contains(src, "func init()") {
			t.Fatalf("%s declares an init() function — native/dynamodb must perform no registration side effects", e.Name())
		}
	}
	if !found {
		t.Fatal("no non-test .go source files found in native/dynamodb; grep-guard did not actually scan anything")
	}
}

// TestConnectorSatisfiesCoreInterfaces mirrors native/postgres's and
// native/bing-ads's identical assertion. Writer/CDCReader interfaces are
// deliberately NOT asserted: Write is unsupported (read-only source) and
// there is no CDC concept at all for this connector (see connector.go's doc
// comment on why there is no cdc.go).
func TestConnectorSatisfiesCoreInterfaces(t *testing.T) {
	c := native.New()
	var _ connectors.Connector = c
	if _, ok := any(c).(connectors.StatefulReader); !ok {
		t.Fatal("native dynamodb connector must implement connectors.StatefulReader")
	}
	if _, ok := any(c).(connectors.DefinitionProvider); !ok {
		t.Fatal("native dynamodb connector must implement connectors.DefinitionProvider (engine.Base)")
	}
}

func TestCheckFixtureModeOK(t *testing.T) {
	c := native.New()
	if err := c.Check(context.Background(), fixtureConfig()); err != nil {
		t.Fatalf("Check(fixture) = %v, want nil", err)
	}
}

func TestCheckRejectsCtxCancelled(t *testing.T) {
	c := native.New()
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if err := c.Check(ctx, fixtureConfig()); err == nil {
		t.Fatal("Check(cancelled ctx) = nil, want error")
	}
}

func TestCheckLiveModeValidatesConfig(t *testing.T) {
	c := native.New()
	cases := []struct {
		name string
		cfg  connectors.RuntimeConfig
	}{
		{
			name: "missing endpoint",
			cfg: connectors.RuntimeConfig{
				Config:  map[string]string{"region": "us-east-1"},
				Secrets: map[string]string{"access_key_id": "AKID", "secret_access_key": "SECRET"},
			},
		},
		{
			name: "missing region",
			cfg: connectors.RuntimeConfig{
				Config:  map[string]string{"endpoint": "https://dynamodb.us-east-1.amazonaws.com"},
				Secrets: map[string]string{"access_key_id": "AKID", "secret_access_key": "SECRET"},
			},
		},
		{
			name: "missing access_key_id",
			cfg: connectors.RuntimeConfig{
				Config:  map[string]string{"endpoint": "https://dynamodb.us-east-1.amazonaws.com", "region": "us-east-1"},
				Secrets: map[string]string{"secret_access_key": "SECRET"},
			},
		},
		{
			name: "missing secret_access_key",
			cfg: connectors.RuntimeConfig{
				Config:  map[string]string{"endpoint": "https://dynamodb.us-east-1.amazonaws.com", "region": "us-east-1"},
				Secrets: map[string]string{"access_key_id": "AKID"},
			},
		},
		{
			name: "endpoint missing scheme",
			cfg: connectors.RuntimeConfig{
				Config:  map[string]string{"endpoint": "dynamodb.us-east-1.amazonaws.com", "region": "us-east-1"},
				Secrets: map[string]string{"access_key_id": "AKID", "secret_access_key": "SECRET"},
			},
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if err := c.Check(context.Background(), tc.cfg); err == nil {
				t.Fatalf("Check(%s) = nil, want validation error", tc.name)
			}
		})
	}
}

func TestCheckNeverLogsSecretAccessKey(t *testing.T) {
	c := native.New()
	cfg := connectors.RuntimeConfig{
		Config:  map[string]string{"endpoint": "not-a-valid-url-\x00", "region": "us-east-1"},
		Secrets: map[string]string{"access_key_id": "AKID", "secret_access_key": "top-secret-value-should-never-appear"},
	}
	err := c.Check(context.Background(), cfg)
	if err == nil {
		t.Fatal("expected a validation error for a malformed endpoint")
	}
	if strings.Contains(err.Error(), "top-secret-value-should-never-appear") {
		t.Fatalf("Check error leaked the secret access key: %v", err)
	}
}

func TestCatalogFixtureMode(t *testing.T) {
	c := native.New()
	cat, err := c.Catalog(context.Background(), fixtureConfig())
	if err != nil {
		t.Fatalf("Catalog: %v", err)
	}
	if cat.Connector != "dynamodb" || len(cat.Streams) != 1 {
		t.Fatalf("Catalog = %+v", cat)
	}
	s := cat.Streams[0]
	if s.Name != "items" || len(s.PrimaryKey) == 0 {
		t.Fatalf("catalog stream = %+v, want items with a primary key", s)
	}
}

func TestReadFixtureEmitsRows(t *testing.T) {
	c := native.New()
	var got []connectors.Record
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "items", Config: fixtureConfig()}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	})
	if err != nil {
		t.Fatalf("Read(fixture): %v", err)
	}
	if len(got) != 2 {
		t.Fatalf("read emitted %d rows, want 2", len(got))
	}
	for _, rec := range got {
		if rec["pk"] == nil {
			t.Fatalf("record missing pk: %+v", rec)
		}
		if rec["fixture"] != true {
			t.Fatalf("record missing fixture marker: %+v", rec)
		}
	}
}

func TestReadUnknownStream(t *testing.T) {
	c := native.New()
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "does_not_exist", Config: fixtureConfig()}, func(connectors.Record) error {
		return nil
	})
	if err == nil {
		t.Fatal("Read(unknown stream) = nil, want error")
	}
}

func TestInitialStateStatefulReader(t *testing.T) {
	c := native.New()
	sr, ok := any(c).(connectors.StatefulReader)
	if !ok {
		t.Fatal("dynamodb connector must implement StatefulReader")
	}
	state, err := sr.InitialState(context.Background(), "items", fixtureConfig())
	if err != nil {
		t.Fatalf("InitialState: %v", err)
	}
	if state == nil {
		t.Fatal("InitialState returned nil state map")
	}
}

func TestWriteUnsupported(t *testing.T) {
	c := native.New()
	_, err := c.Write(context.Background(), connectors.WriteRequest{}, nil)
	if !errors.Is(err, connectors.ErrUnsupportedOperation) {
		t.Fatalf("Write = %v, want ErrUnsupportedOperation", err)
	}
}

// TestReadItemsSignsPaginatesAndMapsRecords is ported rule-for-rule from
// legacy internal/connectors/dynamodb/dynamodb_test.go's
// TestReadItemsSignsPaginatesAndMapsRecords: it drives a real httptest.Server
// and asserts the SigV4 Authorization header shape, the X-Amz-Target header,
// pagination via ExclusiveStartKey/LastEvaluatedKey, and AttributeValue
// flattening (S/N/BOOL), all with an injected deterministic clock.
func TestReadItemsSignsPaginatesAndMapsRecords(t *testing.T) {
	var calls int
	var sawAuth, sawTarget string
	var bodies []string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		sawAuth = r.Header.Get("Authorization")
		sawTarget = r.Header.Get("X-Amz-Target")
		body, _ := io.ReadAll(r.Body)
		bodies = append(bodies, string(body))
		if r.Method != http.MethodPost || r.URL.Path != "/" {
			t.Fatalf("unexpected request %s %s", r.Method, r.URL.Path)
		}
		switch calls {
		case 1:
			_, _ = w.Write([]byte(`{"Items":[{"pk":{"S":"user#1"},"name":{"S":"Ada"}},{"pk":{"S":"user#2"},"score":{"N":"42"}}],"LastEvaluatedKey":{"pk":{"S":"user#2"}}}`))
		case 2:
			if !strings.Contains(string(body), "ExclusiveStartKey") {
				t.Fatalf("second request missing ExclusiveStartKey body=%s", body)
			}
			_, _ = w.Write([]byte(`{"Items":[{"pk":{"S":"user#3"},"active":{"BOOL":true}}]}`))
		default:
			t.Fatalf("unexpected call %d", calls)
		}
	}))
	defer srv.Close()

	c := native.Connector{Client: srv.Client(), Now: func() time.Time { return time.Date(2026, 6, 26, 12, 0, 0, 0, time.UTC) }}
	cfg := connectors.RuntimeConfig{
		Config:  map[string]string{"endpoint": srv.URL, "region": "us-east-1", "table_name": "users", "page_size": "2"},
		Secrets: map[string]string{"access_key_id": "AKID", "secret_access_key": "SECRET"},
	}
	var got []connectors.Record
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "items", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	})
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	if sawTarget != "DynamoDB_20120810.Scan" {
		t.Fatalf("X-Amz-Target = %q", sawTarget)
	}
	if !strings.Contains(sawAuth, "AWS4-HMAC-SHA256") || !strings.Contains(sawAuth, "Credential=AKID/20260626/us-east-1/dynamodb/aws4_request") || strings.Contains(sawAuth, "SECRET") {
		t.Fatalf("Authorization header not signed as expected: %q", sawAuth)
	}
	if !strings.Contains(bodies[0], `"TableName":"users"`) {
		t.Fatalf("first request body missing table name: %s", bodies[0])
	}
	if len(got) != 3 || got[0]["name"] != "Ada" || got[1]["score"] != "42" || got[2]["active"] != true {
		t.Fatalf("records mapped wrong: %+v", got)
	}
}

// TestScanRejectsNonSuccessStatus covers the http-error branch legacy's
// Connector.scan also guards (dynamodb.go:161-163): a non-2xx response is a
// read error, not a silently-empty page.
func TestScanRejectsNonSuccessStatus(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(`{"__type":"ValidationException","message":"boom"}`))
	}))
	defer srv.Close()

	c := native.Connector{Client: srv.Client(), Now: func() time.Time { return time.Date(2026, 6, 26, 12, 0, 0, 0, time.UTC) }}
	cfg := connectors.RuntimeConfig{
		Config:  map[string]string{"endpoint": srv.URL, "region": "us-east-1", "table_name": "users"},
		Secrets: map[string]string{"access_key_id": "AKID", "secret_access_key": "SECRET"},
	}
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "items", Config: cfg}, func(connectors.Record) error { return nil })
	if err == nil {
		t.Fatal("Read = nil, want error on non-2xx Scan response")
	}
}
