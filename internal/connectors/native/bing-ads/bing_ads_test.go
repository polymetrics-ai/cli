package bingads_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"polymetrics.ai/internal/connectors"
	native "polymetrics.ai/internal/connectors/native/bing-ads"
)

func fixtureConfig() connectors.RuntimeConfig {
	return connectors.RuntimeConfig{Config: map[string]string{"mode": "fixture"}}
}

func TestNameAndMetadata(t *testing.T) {
	c := native.New()
	if c.Name() != "bing-ads" {
		t.Fatalf("Name() = %q, want bing-ads", c.Name())
	}
	caps := c.Metadata().Capabilities
	if !caps.Check || !caps.Catalog || !caps.Read {
		t.Fatalf("capabilities = %+v, want Check && Catalog && Read", caps)
	}
	if caps.Write {
		t.Fatal("bing-ads is read-only; Write capability must be false")
	}
}

// TestNoInitRegistration is the required grep-guard, mirroring
// native/postgres's identical guard (T-17 precedent): the native package
// must NOT call RegisterFactory/RegisterNativeLive from anywhere in its own
// source. The registration flip is a wave6 change; wave0 only builds and
// tests the package.
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
			t.Fatalf("%s calls RegisterFactory — native/bing-ads must NOT self-register in wave0 (registration flip is wave6)", e.Name())
		}
		if strings.Contains(src, "RegisterNativeLive(") {
			t.Fatalf("%s calls RegisterNativeLive — native/bing-ads must NOT self-register in wave0 (registration flip is wave6)", e.Name())
		}
		if strings.Contains(src, "func init()") {
			t.Fatalf("%s declares an init() function — native/bing-ads must perform no registration side effects in wave0", e.Name())
		}
	}
	if !found {
		t.Fatal("no non-test .go source files found in native/bing-ads; grep-guard did not actually scan anything")
	}
}

// TestConnectorSatisfiesCoreInterfaces mirrors native/postgres's identical
// assertion. Writer interfaces are deliberately NOT asserted since Write is
// unsupported (read-only source).
func TestConnectorSatisfiesCoreInterfaces(t *testing.T) {
	c := native.New()
	var _ connectors.Connector = c
	if _, ok := any(c).(connectors.StatefulReader); !ok {
		t.Fatal("native bing-ads connector must implement connectors.StatefulReader")
	}
	if _, ok := any(c).(connectors.DefinitionProvider); !ok {
		t.Fatal("native bing-ads connector must implement connectors.DefinitionProvider (engine.Base)")
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

func TestCatalogStreams(t *testing.T) {
	c := native.New()
	cat, err := c.Catalog(context.Background(), connectors.RuntimeConfig{})
	if err != nil {
		t.Fatalf("Catalog: %v", err)
	}
	if cat.Connector != "bing-ads" {
		t.Fatalf("catalog connector = %q, want bing-ads", cat.Connector)
	}
	want := map[string]bool{"accounts": false, "campaigns": false, "ad_groups": false, "ads": false, "users": false}
	for _, s := range cat.Streams {
		if _, ok := want[s.Name]; ok {
			want[s.Name] = true
			if len(s.PrimaryKey) == 0 {
				t.Fatalf("stream %s missing primary key", s.Name)
			}
		}
	}
	for name, seen := range want {
		if !seen {
			t.Fatalf("catalog missing core stream %q", name)
		}
	}
}

// TestFixtureModeNeedsNoNetwork confirms the credential-free fixture path
// emits deterministic records without any HTTP call (mirrors legacy
// TestFixtureModeNeedsNoNetwork).
func TestFixtureModeNeedsNoNetwork(t *testing.T) {
	c := native.New()

	for _, stream := range []string{"accounts", "campaigns", "ad_groups", "ads", "users"} {
		var got []connectors.Record
		err := c.Read(context.Background(), connectors.ReadRequest{Stream: stream, Config: fixtureConfig()}, func(rec connectors.Record) error {
			got = append(got, rec)
			return nil
		})
		if err != nil {
			t.Fatalf("fixture Read(%s): %v", stream, err)
		}
		if len(got) == 0 {
			t.Fatalf("fixture Read(%s) emitted no records", stream)
		}
		for _, rec := range got {
			if rec["Id"] == nil {
				t.Fatalf("fixture %s record missing Id: %+v", stream, rec)
			}
		}
	}
}

// TestReadAccountsAuthenticatesAndMaps is the live-server test (ported from
// legacy bing_ads_test.go): it stands up a fake Microsoft OAuth token
// endpoint plus a fake Bing Ads Customer Management REST endpoint and
// asserts that the connector exchanges the refresh_token for an access
// token, the API call carries Authorization: Bearer <token> and the
// DeveloperToken header, and the AccountsInfo array is mapped into records
// keyed by Id.
func TestReadAccountsAuthenticatesAndMaps(t *testing.T) {
	var (
		sawGrant    string
		sawRefresh  string
		sawScope    string
		sawAuth     string
		sawDevToken string
		tokenCalled int
	)

	mux := http.NewServeMux()
	mux.HandleFunc("/oauth/token", func(w http.ResponseWriter, r *http.Request) {
		tokenCalled++
		_ = r.ParseForm()
		sawGrant = r.Form.Get("grant_type")
		sawRefresh = r.Form.Get("refresh_token")
		sawScope = r.Form.Get("scope")
		_, _ = w.Write([]byte(`{"access_token":"ACCESS_XYZ","token_type":"Bearer","expires_in":3600}`))
	})
	mux.HandleFunc("/CustomerManagement/v13/AccountsInfo/Query", func(w http.ResponseWriter, r *http.Request) {
		sawAuth = r.Header.Get("Authorization")
		sawDevToken = r.Header.Get("DeveloperToken")
		_, _ = w.Write([]byte(`{"AccountsInfo":[{"Id":"111","Name":"Acme","Number":"X0001","AccountLifeCycleStatus":"Active"},{"Id":"222","Name":"Globex","Number":"X0002","AccountLifeCycleStatus":"Paused"}]}`))
	})
	srv := httptest.NewServer(mux)
	defer srv.Close()

	c := native.New()
	cfg := connectors.RuntimeConfig{
		Config: map[string]string{
			"base_url":  srv.URL + "/CustomerManagement/v13",
			"token_url": srv.URL + "/oauth/token",
		},
		Secrets: map[string]string{
			"client_id":       "cid",
			"client_secret":   "csecret",
			"developer_token": "DEVTOKEN123",
			"refresh_token":   "rtok",
			"tenant_id":       "common",
		},
	}

	var got []connectors.Record
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "accounts", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	})
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	if tokenCalled == 0 {
		t.Fatal("OAuth token endpoint was never called")
	}
	if sawGrant != "refresh_token" {
		t.Fatalf("grant_type = %q, want refresh_token", sawGrant)
	}
	if sawRefresh != "rtok" {
		t.Fatalf("refresh_token = %q, want rtok", sawRefresh)
	}
	if !strings.Contains(sawScope, "msads.manage") {
		t.Fatalf("scope = %q, want it to contain msads.manage", sawScope)
	}
	if sawAuth != "Bearer ACCESS_XYZ" {
		t.Fatalf("Authorization = %q, want Bearer ACCESS_XYZ", sawAuth)
	}
	if sawDevToken != "DEVTOKEN123" {
		t.Fatalf("DeveloperToken = %q, want DEVTOKEN123", sawDevToken)
	}
	if len(got) != 2 {
		t.Fatalf("records = %d, want 2", len(got))
	}
	if got[0]["Id"] != "111" || got[0]["Name"] != "Acme" {
		t.Fatalf("record[0] = %+v, want Id=111 Name=Acme", got[0])
	}
}

// TestReadCampaignsFansOutByAccount asserts that campaign-scoped streams
// send the CustomerId/CustomerAccountId headers and the AccountId in the
// POST body, iterate the configured account ids, and that the Campaigns
// array maps correctly.
func TestReadCampaignsFansOutByAccount(t *testing.T) {
	var (
		sawCustomerID string
		seenAccounts  []string
	)

	mux := http.NewServeMux()
	mux.HandleFunc("/oauth/token", func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"access_token":"ACCESS_XYZ","token_type":"Bearer","expires_in":3600}`))
	})
	mux.HandleFunc("/CampaignManagement/v13/Campaigns/QueryByAccountId", func(w http.ResponseWriter, r *http.Request) {
		sawCustomerID = r.Header.Get("CustomerId")
		var body map[string]any
		_ = json.NewDecoder(r.Body).Decode(&body)
		acct, _ := body["AccountId"].(string)
		seenAccounts = append(seenAccounts, acct)
		switch acct {
		case "111":
			_, _ = w.Write([]byte(`{"Campaigns":[{"Id":"900","Name":"Brand","Status":"Active","CampaignType":"Search"}]}`))
		case "222":
			_, _ = w.Write([]byte(`{"Campaigns":[{"Id":"901","Name":"Shopping","Status":"Paused","CampaignType":"Shopping"}]}`))
		default:
			_, _ = w.Write([]byte(`{"Campaigns":[]}`))
		}
	})
	srv := httptest.NewServer(mux)
	defer srv.Close()

	c := native.New()
	cfg := connectors.RuntimeConfig{
		Config: map[string]string{
			"campaign_base_url": srv.URL + "/CampaignManagement/v13",
			"token_url":         srv.URL + "/oauth/token",
			"customer_id":       "C42",
			"account_ids":       "111,222",
		},
		Secrets: map[string]string{
			"client_id":       "cid",
			"developer_token": "DEVTOKEN123",
			"refresh_token":   "rtok",
		},
	}

	var got []connectors.Record
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "campaigns", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	})
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	if sawCustomerID != "C42" {
		t.Fatalf("CustomerId header = %q, want C42", sawCustomerID)
	}
	if len(seenAccounts) != 2 || seenAccounts[0] != "111" || seenAccounts[1] != "222" {
		t.Fatalf("fanned-out accounts = %v, want [111 222]", seenAccounts)
	}
	if len(got) != 2 {
		t.Fatalf("campaign records = %d, want 2 (one per account)", len(got))
	}
	if got[0]["Id"] != "900" || got[1]["Id"] != "901" {
		t.Fatalf("campaign ids = %v / %v, want 900 / 901", got[0]["Id"], got[1]["Id"])
	}
	// No account-origin marker is stamped onto the record (matches legacy's
	// campaignRecord exactly — see docs.md's parity-deviation ledger entry).
	if _, ok := got[0]["account_id"]; ok {
		t.Fatalf("campaign record unexpectedly carries an account_id field: %+v", got[0])
	}
}

func TestBaseURLValidationRejectsBadScheme(t *testing.T) {
	c := native.New()
	cfg := connectors.RuntimeConfig{
		Config:  map[string]string{"base_url": "ftp://evil.example.com"},
		Secrets: map[string]string{"developer_token": "d", "refresh_token": "r", "client_id": "c"},
	}
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "accounts", Config: cfg}, func(connectors.Record) error { return nil })
	if err == nil || !strings.Contains(err.Error(), "base_url") {
		t.Fatalf("expected base_url validation error, got %v", err)
	}
}

func TestMissingSecretsRejected(t *testing.T) {
	c := native.New()
	cfg := connectors.RuntimeConfig{Secrets: map[string]string{"client_id": "c"}}
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "accounts", Config: cfg}, func(connectors.Record) error { return nil })
	if err == nil {
		t.Fatal("expected missing-secret error, got nil")
	}
}

func TestUnknownStreamRejected(t *testing.T) {
	c := native.New()
	cfg := fixtureConfig()
	delete(cfg.Config, "mode")
	cfg.Secrets = map[string]string{"client_id": "c", "developer_token": "d", "refresh_token": "r"}
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "does-not-exist", Config: cfg}, func(connectors.Record) error { return nil })
	if err == nil || !strings.Contains(err.Error(), "not found") {
		t.Fatalf("expected unknown-stream error, got %v", err)
	}
}

func TestInitialStateEmptyCursor(t *testing.T) {
	c := native.New()
	state, err := c.InitialState(context.Background(), "accounts", connectors.RuntimeConfig{})
	if err != nil {
		t.Fatalf("InitialState: %v", err)
	}
	if state["cursor"] != "" {
		t.Fatalf("InitialState cursor = %q, want empty (full_refresh only)", state["cursor"])
	}
}

func TestWriteUnsupported(t *testing.T) {
	c := native.New()
	_, err := c.Write(context.Background(), connectors.WriteRequest{}, nil)
	if err != connectors.ErrUnsupportedOperation {
		t.Fatalf("Write error = %v, want ErrUnsupportedOperation", err)
	}
}
