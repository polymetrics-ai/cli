package tallyprime_test

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"polymetrics.ai/internal/connectors"
	native "polymetrics.ai/internal/connectors/native/tally-prime"
)

// --- grep-guard -------------------------------------------------------

// TestNoInitRegistration is the required grep-guard (mirrors
// native/postgres's and native/amazon-sqs's TestNoInitRegistration): the
// native package must NOT call RegisterFactory from
// anywhere in its own source, nor declare an init() function. The
// registration flip (wiring native/tally-prime into the production
// registry) is a later-wave change; this wave only builds and tests the
// package standalone.
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
			// The grep-guard covers the package's own production source, not
			// its tests (this very test file legitimately mentions the
			// forbidden identifiers in prose/identifiers above).
			continue
		}
		found = true
		raw, err := os.ReadFile(filepath.Join(dir, e.Name()))
		if err != nil {
			t.Fatalf("ReadFile(%s): %v", e.Name(), err)
		}
		src := string(raw)
		if strings.Contains(src, "RegisterFactory(") {
			t.Fatalf("%s calls RegisterFactory — native/tally-prime must NOT self-register (registration flip is a later wave)", e.Name())
		}
		if strings.Contains(src, "func init()") {
			t.Fatalf("%s declares an init() function — native/tally-prime must perform no registration side effects", e.Name())
		}
	}
	if !found {
		t.Fatal("no non-test .go source files found in native/tally-prime; grep-guard did not actually scan anything")
	}
}

// --- interface shape ----------------------------------------------------

// TestConnectorSatisfiesCoreInterfaces compile/runtime-asserts the shape
// required by API-CONTRACT.md / design §B.7 Tier-3: Connector,
// StatefulReader, DefinitionProvider. CDCReader is deliberately NOT
// asserted: TallyPrime's Gateway Server is polled per Read with no
// subscription/webhook mechanism, so there is no CDC path (unlike
// postgres's documented pglogrepl stub).
func TestConnectorSatisfiesCoreInterfaces(t *testing.T) {
	c := native.New()
	var _ connectors.Connector = c
	if _, ok := any(c).(connectors.StatefulReader); !ok {
		t.Fatal("native tally-prime connector must implement connectors.StatefulReader")
	}
	if _, ok := any(c).(connectors.DefinitionProvider); !ok {
		t.Fatal("native tally-prime connector must implement connectors.DefinitionProvider (engine.Base)")
	}
	if _, ok := any(c).(connectors.CDCReader); ok {
		t.Fatal("native tally-prime connector must NOT implement connectors.CDCReader (no subscription mechanism; documented in docs.md's Known limits)")
	}
}

func TestNameAndMetadata(t *testing.T) {
	c := native.New()
	if c.Name() != "tally-prime" {
		t.Fatalf("Name() = %q, want tally-prime", c.Name())
	}
	caps := c.Metadata().Capabilities
	if !caps.Check || !caps.Catalog || !caps.Read {
		t.Fatalf("capabilities = %+v, want Check && Catalog && Read", caps)
	}
	if caps.Write {
		t.Fatalf("tally-prime source connector must be read-only, got Write=true")
	}
}

func TestWriteUnsupported(t *testing.T) {
	c := native.New()
	_, err := c.Write(context.Background(), connectors.WriteRequest{Stream: "ledgers", Config: fixtureConfig()}, []connectors.Record{{"name": "x"}})
	if !errors.Is(err, connectors.ErrUnsupportedOperation) {
		t.Fatalf("Write = %v, want ErrUnsupportedOperation", err)
	}
}

// --- fixture-mode config ------------------------------------------------

func fixtureConfig() connectors.RuntimeConfig {
	return connectors.RuntimeConfig{
		Config: map[string]string{
			"mode":    "fixture",
			"company": "Acme Retail Pvt Ltd",
		},
	}
}

func liveConfig(gatewayURL string) connectors.RuntimeConfig {
	return connectors.RuntimeConfig{
		Config: map[string]string{
			"gateway_url": gatewayURL,
			"company":     "Acme Retail Pvt Ltd",
		},
	}
}

// --- fixture-mode behavior ------------------------------------------------

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

func TestCheckRequiresCompany(t *testing.T) {
	c := native.New()
	cfg := connectors.RuntimeConfig{Config: map[string]string{"mode": "fixture"}}
	if err := c.Check(context.Background(), cfg); err == nil {
		t.Fatal("Check(no company) = nil, want validation error")
	}
}

func TestCheckRejectsBadGatewayURL(t *testing.T) {
	c := native.New()
	cfg := connectors.RuntimeConfig{Config: map[string]string{"mode": "fixture", "company": "Acme", "gateway_url": "ftp://bad"}}
	if err := c.Check(context.Background(), cfg); err == nil {
		t.Fatal("Check(bad gateway_url scheme) = nil, want validation error")
	}
}

func TestCheckRejectsBadEnvelopeFormat(t *testing.T) {
	c := native.New()
	cfg := connectors.RuntimeConfig{Config: map[string]string{"mode": "fixture", "company": "Acme", "envelope_format": "yaml"}}
	if err := c.Check(context.Background(), cfg); err == nil {
		t.Fatal("Check(bad envelope_format) = nil, want validation error")
	}
}

func TestCatalogFixtureMode(t *testing.T) {
	c := native.New()
	cat, err := c.Catalog(context.Background(), fixtureConfig())
	if err != nil {
		t.Fatalf("Catalog(fixture) = %v", err)
	}
	if cat.Connector != "tally-prime" {
		t.Fatalf("catalog connector = %q, want tally-prime", cat.Connector)
	}
	wantStreams := []string{"companies", "ledgers", "groups", "stock_items", "vouchers"}
	if len(cat.Streams) != len(wantStreams) {
		t.Fatalf("catalog returned %d streams, want %d", len(cat.Streams), len(wantStreams))
	}
	for i, name := range wantStreams {
		if cat.Streams[i].Name != name {
			t.Fatalf("catalog.Streams[%d].Name = %q, want %q", i, cat.Streams[i].Name, name)
		}
		if len(cat.Streams[i].PrimaryKey) == 0 {
			t.Fatalf("stream %q missing primary key", name)
		}
		if len(cat.Streams[i].Fields) == 0 {
			t.Fatalf("stream %q missing fields", name)
		}
	}
	// Only vouchers carries an incremental cursor field.
	for _, s := range cat.Streams {
		if s.Name == "vouchers" {
			if len(s.CursorFields) == 0 {
				t.Fatal("vouchers stream missing cursor fields")
			}
		} else if len(s.CursorFields) != 0 {
			t.Fatalf("master stream %q unexpectedly has cursor fields %v", s.Name, s.CursorFields)
		}
	}
}

func TestReadFixtureEmitsRowsPerStream(t *testing.T) {
	c := native.New()
	for _, stream := range []string{"companies", "ledgers", "groups", "stock_items", "vouchers"} {
		t.Run(stream, func(t *testing.T) {
			var got []connectors.Record
			err := c.Read(context.Background(), connectors.ReadRequest{Stream: stream, Config: fixtureConfig()}, func(rec connectors.Record) error {
				got = append(got, rec)
				return nil
			})
			if err != nil {
				t.Fatalf("Read(fixture, %s): %v", stream, err)
			}
			if len(got) == 0 {
				t.Fatalf("Read(fixture, %s) emitted zero rows", stream)
			}
		})
	}
}

func TestReadFixtureVouchersIncrementalCursor(t *testing.T) {
	c := native.New()
	countWithState := func(state map[string]string) int {
		var n int
		_ = c.Read(context.Background(), connectors.ReadRequest{Stream: "vouchers", Config: fixtureConfig(), State: state}, func(connectors.Record) error {
			n++
			return nil
		})
		return n
	}
	full := countWithState(nil)
	high := countWithState(map[string]string{"cursor": "20250610"})
	if high >= full {
		t.Fatalf("incremental read returned %d rows with high cursor, want fewer than full %d", high, full)
	}
}

func TestReadFixtureMastersIgnoreCursor(t *testing.T) {
	// Master streams (non-vouchers) have no incremental filter; a cursor
	// value in State must not reduce the row count.
	c := native.New()
	countWithState := func(state map[string]string) int {
		var n int
		_ = c.Read(context.Background(), connectors.ReadRequest{Stream: "ledgers", Config: fixtureConfig(), State: state}, func(connectors.Record) error {
			n++
			return nil
		})
		return n
	}
	full := countWithState(nil)
	withCursor := countWithState(map[string]string{"cursor": "99999999"})
	if withCursor != full {
		t.Fatalf("master stream read with cursor = %d rows, want unchanged %d", withCursor, full)
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

func TestReadRequiresStream(t *testing.T) {
	c := native.New()
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "", Config: fixtureConfig()}, func(connectors.Record) error {
		return nil
	})
	if err == nil {
		t.Fatal("Read(no stream) = nil, want error")
	}
}

func TestInitialStateStatefulReader(t *testing.T) {
	c := native.New()
	sr, ok := any(c).(connectors.StatefulReader)
	if !ok {
		t.Fatal("tally-prime connector must implement StatefulReader")
	}
	state, err := sr.InitialState(context.Background(), "vouchers", fixtureConfig())
	if err != nil {
		t.Fatalf("InitialState: %v", err)
	}
	if state == nil {
		t.Fatal("InitialState returned nil state map")
	}
}

// --- live mode: httptest envelope construction/decode --------------------

// TestCheckLiveJSONPostsExportCollectionEnvelope verifies Check builds a
// well-formed TALLYREQUEST=Export/TYPE=Collection JSON envelope for the
// companies collection and honors a 2xx response.
func TestCheckLiveJSONPostsExportCollectionEnvelope(t *testing.T) {
	var captured map[string]any
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("method = %s, want POST", r.Method)
		}
		if ct := r.Header.Get("Content-Type"); ct != "application/json" {
			t.Errorf("Content-Type = %q, want application/json", ct)
		}
		if err := json.NewDecoder(r.Body).Decode(&captured); err != nil {
			t.Fatalf("decode request body: %v", err)
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"ENVELOPE":{"BODY":{"DATA":{"COLLECTION":{}}}}}`))
	}))
	defer srv.Close()

	c := native.New()
	if err := c.Check(context.Background(), liveConfig(srv.URL)); err != nil {
		t.Fatalf("Check(live json) = %v, want nil", err)
	}

	header, ok := captured["HEADER"].(map[string]any)
	if !ok {
		t.Fatalf("captured envelope missing HEADER: %+v", captured)
	}
	if header["TALLYREQUEST"] != "Export" {
		t.Fatalf("HEADER.TALLYREQUEST = %v, want Export", header["TALLYREQUEST"])
	}
	if header["TYPE"] != "Collection" {
		t.Fatalf("HEADER.TYPE = %v, want Collection", header["TYPE"])
	}
	if header["ID"] != "ID_Companies" {
		t.Fatalf("HEADER.ID = %v, want ID_Companies", header["ID"])
	}

	body, ok := captured["BODY"].(map[string]any)
	if !ok {
		t.Fatalf("captured envelope missing BODY: %+v", captured)
	}
	desc, ok := body["DESC"].(map[string]any)
	if !ok {
		t.Fatalf("captured envelope missing BODY.DESC: %+v", body)
	}
	sv, ok := desc["STATICVARIABLES"].(map[string]any)
	if !ok {
		t.Fatalf("captured envelope missing DESC.STATICVARIABLES: %+v", desc)
	}
	if sv["SVEXPORTFORMAT"] != "$$SysName:UTF8JSON" {
		t.Fatalf("SVEXPORTFORMAT = %v, want $$SysName:UTF8JSON (json mode default)", sv["SVEXPORTFORMAT"])
	}
	if sv["SVCURRENTCOMPANY"] != "Acme Retail Pvt Ltd" {
		t.Fatalf("SVCURRENTCOMPANY = %v, want Acme Retail Pvt Ltd", sv["SVCURRENTCOMPANY"])
	}
}

// TestCheckLiveNon2xxFails verifies a non-2xx Gateway Server response is
// surfaced as a Check error.
func TestCheckLiveNon2xxFails(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer srv.Close()

	c := native.New()
	if err := c.Check(context.Background(), liveConfig(srv.URL)); err == nil {
		t.Fatal("Check(live, 500) = nil, want error")
	}
}

// TestReadLiveJSONDecodesLedgers verifies Read builds the ledgers
// Export/Collection envelope and decodes a native-JSON response envelope
// into records with the connector's snake_case output field names.
func TestReadLiveJSONDecodesLedgers(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req map[string]any
		_ = json.NewDecoder(r.Body).Decode(&req)
		header := req["HEADER"].(map[string]any)
		if header["ID"] != "ID_Ledgers" {
			t.Fatalf("HEADER.ID = %v, want ID_Ledgers", header["ID"])
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{
			"ENVELOPE": {
				"BODY": {
					"DATA": {
						"COLLECTION": {
							"ID_Ledgers": [
								{"NAME": "Sales Account", "PARENT": "Sales Accounts", "OPENINGBALANCE": "0", "CLOSINGBALANCE": "125000.50", "ISBILLWISEON": "No"},
								{"NAME": "ABC Distributors", "PARENT": "Sundry Debtors", "OPENINGBALANCE": "5000", "CLOSINGBALANCE": "18250.75", "ISBILLWISEON": "Yes"}
							]
						}
					}
				}
			}
		}`))
	}))
	defer srv.Close()

	c := native.New()
	var got []connectors.Record
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "ledgers", Config: liveConfig(srv.URL)}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	})
	if err != nil {
		t.Fatalf("Read(live json ledgers) = %v", err)
	}
	if len(got) != 2 {
		t.Fatalf("got %d records, want 2", len(got))
	}
	if got[0]["name"] != "Sales Account" {
		t.Fatalf("record[0].name = %v, want Sales Account", got[0]["name"])
	}
	if got[0]["closing_balance"] != 125000.50 {
		t.Fatalf("record[0].closing_balance = %v, want 125000.50", got[0]["closing_balance"])
	}
	if got[0]["is_bill_wise_on"] != false {
		t.Fatalf("record[0].is_bill_wise_on = %v, want false", got[0]["is_bill_wise_on"])
	}
	if got[1]["is_bill_wise_on"] != true {
		t.Fatalf("record[1].is_bill_wise_on = %v, want true", got[1]["is_bill_wise_on"])
	}
}

// TestReadLiveXMLDecodesVouchers verifies Read builds an XML-fallback
// vouchers envelope and decodes an XML response envelope into records,
// including the incremental cursor filter.
func TestReadLiveXMLDecodesVouchers(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if ct := r.Header.Get("Content-Type"); ct != "text/xml" {
			t.Errorf("Content-Type = %q, want text/xml", ct)
		}
		w.WriteHeader(http.StatusOK)
		w.Header().Set("Content-Type", "text/xml")
		_, _ = w.Write([]byte(`<ENVELOPE>
			<COLLECTION>
				<GUID>guid-1</GUID>
				<VOUCHERNUMBER>1001</VOUCHERNUMBER>
				<VOUCHERTYPENAME>Sales</VOUCHERTYPENAME>
				<DATE>20250601</DATE>
				<PARTYNAME>ABC Distributors</PARTYNAME>
				<AMOUNT>12500.00</AMOUNT>
				<NARRATION>Sales voucher 1</NARRATION>
			</COLLECTION>
			<COLLECTION>
				<GUID>guid-2</GUID>
				<VOUCHERNUMBER>1002</VOUCHERNUMBER>
				<VOUCHERTYPENAME>Sales</VOUCHERTYPENAME>
				<DATE>20250701</DATE>
				<PARTYNAME>ABC Distributors</PARTYNAME>
				<AMOUNT>5750.75</AMOUNT>
				<NARRATION>Sales voucher 2</NARRATION>
			</COLLECTION>
		</ENVELOPE>`))
	}))
	defer srv.Close()

	cfg := liveConfig(srv.URL)
	cfg.Config["envelope_format"] = "xml"

	c := native.New()
	var got []connectors.Record
	err := c.Read(context.Background(), connectors.ReadRequest{
		Stream: "vouchers",
		Config: cfg,
		State:  map[string]string{"cursor": "20250601"},
	}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	})
	if err != nil {
		t.Fatalf("Read(live xml vouchers) = %v", err)
	}
	// The cursor lower bound (20250601) should exclude the first voucher
	// (date == lower bound) and keep the second (date > lower bound).
	if len(got) != 1 {
		t.Fatalf("got %d records, want 1 (cursor filter should exclude the first)", len(got))
	}
	if got[0]["guid"] != "guid-2" {
		t.Fatalf("record[0].guid = %v, want guid-2", got[0]["guid"])
	}
	if got[0]["amount"] != 5750.75 {
		t.Fatalf("record[0].amount = %v, want 5750.75", got[0]["amount"])
	}
}

// TestReadLiveRequiresCompany verifies the live path validates config the
// same way Check does (company required).
func TestReadLiveRequiresCompany(t *testing.T) {
	c := native.New()
	cfg := connectors.RuntimeConfig{Config: map[string]string{"gateway_url": "http://127.0.0.1:1"}}
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "ledgers", Config: cfg}, func(connectors.Record) error {
		return nil
	})
	if err == nil {
		t.Fatal("Read(live, no company) = nil, want validation error")
	}
}

// --- Definition() smoke ---------------------------------------------------

func TestDefinitionSmoke(t *testing.T) {
	c := native.New()
	def := c.Definition()
	if def.Name != "tally-prime" {
		t.Fatalf("Definition().Name = %q, want tally-prime", def.Name)
	}
	if !def.Capabilities.Read || def.Capabilities.Write {
		t.Fatalf("Definition().Capabilities = %+v, want Read=true Write=false", def.Capabilities)
	}
	if len(def.Spec) == 0 {
		t.Fatal("Definition().Spec is empty")
	}
}
