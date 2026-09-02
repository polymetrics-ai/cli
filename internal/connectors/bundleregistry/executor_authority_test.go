package bundleregistry

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync/atomic"
	"testing"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/defs"
	"polymetrics.ai/internal/connectors/engine"
)

func TestEveryImplementedCommandHasProductionRuntimeSurface(t *testing.T) {
	bundles, err := engine.LoadAll(defs.FS)
	if err != nil {
		t.Fatalf("LoadAll(defs): %v", err)
	}
	registry := New()

	for _, bundle := range bundles {
		if bundle.CLISurface == nil {
			continue
		}
		declared := make(map[string]string, len(bundle.CLISurface.Commands))
		for _, command := range bundle.CLISurface.Commands {
			if command.Availability == "implemented" {
				declared[command.Path] = command.Availability
			}
		}
		if len(declared) == 0 {
			continue
		}

		connector, ok := registry.Get(bundle.Name)
		if !ok {
			t.Errorf("implemented commands for %q have no production connector", bundle.Name)
			continue
		}
		surfaceProvider, ok := connector.(connectors.CommandSurfaceProvider)
		if !ok || surfaceProvider.CommandSurface() == nil {
			t.Errorf("implemented commands for %q have no production command surface: %T", bundle.Name, connector)
			continue
		}
		actual := make(map[string]string, len(surfaceProvider.CommandSurface().Commands))
		for _, command := range surfaceProvider.CommandSurface().Commands {
			actual[command.Path] = command.Availability
		}
		for path, availability := range declared {
			if got, ok := actual[path]; !ok || got != availability {
				t.Errorf("implemented command %q for %q has production availability %q, want %q", path, bundle.Name, got, availability)
			}
		}
	}
}

func TestRegistryRejectsDuplicateConnectorNames(t *testing.T) {
	registry := connectors.NewEmptyRegistry()
	if err := registry.Register(connectors.Sample{}); err != nil {
		t.Fatalf("first Register() error = %v", err)
	}
	if err := registry.Register(connectors.Sample{}); err == nil {
		t.Fatal("Register() silently accepted a duplicate connector name")
	}
}

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(request *http.Request) (*http.Response, error) {
	return f(request)
}

func TestProductionAlphaVantageRejectsFixtureAndCallerOriginBeforeIO(t *testing.T) {
	previousTransport := http.DefaultTransport
	t.Cleanup(func() { http.DefaultTransport = previousTransport })

	var requests atomic.Int64
	http.DefaultTransport = roundTripFunc(func(*http.Request) (*http.Response, error) {
		requests.Add(1)
		return nil, errors.New("unexpected provider request")
	})

	connector, ok := New().Get("alpha-vantage")
	if !ok {
		t.Fatal("production registry missing alpha-vantage")
	}
	for _, config := range []connectors.RuntimeConfig{
		{Config: map[string]string{"mode": "fixture", "base_url": "https://untrusted.invalid"}},
		{Config: map[string]string{"base_url": "https://untrusted.invalid"}, Secrets: map[string]string{"api_key": "test-only"}},
	} {
		if err := connector.Check(context.Background(), config); err == nil {
			t.Fatal("production alpha-vantage Check() accepted forbidden fixture or caller-origin configuration")
		}
	}
	if got := requests.Load(); got != 0 {
		t.Fatalf("production alpha-vantage sent %d requests for forbidden configuration, want none", got)
	}
}

func TestProductionAlphaVantageGenericCheckAndRead(t *testing.T) {
	previousTransport := http.DefaultTransport
	t.Cleanup(func() { http.DefaultTransport = previousTransport })

	var requests atomic.Int64
	http.DefaultTransport = roundTripFunc(func(request *http.Request) (*http.Response, error) {
		requests.Add(1)
		if request.URL.Host != "www.alphavantage.co" || request.URL.Path != "/query" {
			t.Fatal("Alpha Vantage request did not use the fixed provider route")
		}
		if request.URL.Query().Get("apikey") == "" {
			t.Fatal("Alpha Vantage request omitted declared API-key query authentication")
		}
		var body string
		switch request.URL.Query().Get("function") {
		case "GLOBAL_QUOTE":
			body = `{"Global Quote":{"01. symbol":"IBM"}}`
		case "TIME_SERIES_DAILY":
			body = `{"Time Series (Daily)":{"2026-01-02":{"1. open":"100.25","2. high":"101.25","3. low":"99.25","4. close":"100.75","5. volume":"42"}}}`
		default:
			t.Fatal("Alpha Vantage request used an undeclared function")
		}
		return &http.Response{
			StatusCode: http.StatusOK,
			Header:     make(http.Header),
			Body:       io.NopCloser(strings.NewReader(body)),
			Request:    request,
		}, nil
	})

	connector, ok := New().Get("alpha-vantage")
	if !ok {
		t.Fatal("production registry missing alpha-vantage")
	}
	config := connectors.RuntimeConfig{
		Config:  map[string]string{"symbol": "IBM"},
		Secrets: map[string]string{"api_key": "test-only"},
	}
	if err := connector.Check(context.Background(), config); err != nil {
		t.Fatalf("production alpha-vantage Check() error = %v", err)
	}

	var records []connectors.Record
	if err := connector.Read(context.Background(), connectors.ReadRequest{Stream: "time_series_daily", Config: config}, func(record connectors.Record) error {
		records = append(records, record)
		return nil
	}); err != nil {
		t.Fatalf("production alpha-vantage Read() error = %v", err)
	}
	if len(records) != 1 || records[0]["date"] != "2026-01-02" || records[0]["symbol"] != "IBM" || records[0]["open"] != "100.25" {
		t.Fatalf("production alpha-vantage records = %#v, want mapped daily response", records)
	}
	if got := requests.Load(); got != 2 {
		t.Fatalf("production alpha-vantage requests = %d, want check plus read", got)
	}
}

func TestProductionApifyDatasetGenericCheckAndRead(t *testing.T) {
	previousTransport := http.DefaultTransport
	t.Cleanup(func() { http.DefaultTransport = previousTransport })

	var requests atomic.Int64
	http.DefaultTransport = roundTripFunc(func(request *http.Request) (*http.Response, error) {
		requests.Add(1)
		if request.URL.Host != "api.apify.com" {
			t.Fatal("Apify request did not use the fixed provider origin")
		}
		if request.Header.Get("Authorization") == "" {
			t.Fatal("Apify request omitted declared bearer authentication")
		}
		var body string
		switch request.URL.Path {
		case "/v2/datasets":
			body = `{"data":{"items":[]}}`
		case "/v2/datasets/dataset-1/items":
			body = `[{"id":"item-1","nested":{"value":42}}]`
		default:
			t.Fatal("Apify request used an undeclared route")
		}
		return &http.Response{
			StatusCode: http.StatusOK,
			Header:     make(http.Header),
			Body:       io.NopCloser(strings.NewReader(body)),
			Request:    request,
		}, nil
	})

	connector, ok := New().Get("apify-dataset")
	if !ok {
		t.Fatal("production registry missing apify-dataset")
	}
	config := connectors.RuntimeConfig{
		Config:  map[string]string{"dataset_id": "dataset-1"},
		Secrets: map[string]string{"token": "test-only"},
	}
	if err := connector.Check(context.Background(), config); err != nil {
		t.Fatalf("production apify-dataset Check() error = %v", err)
	}

	var records []connectors.Record
	if err := connector.Read(context.Background(), connectors.ReadRequest{Stream: "item_collection", Config: config}, func(record connectors.Record) error {
		records = append(records, record)
		return nil
	}); err != nil {
		t.Fatalf("production apify-dataset Read() error = %v", err)
	}
	if len(records) != 1 {
		t.Fatalf("production apify-dataset records = %#v, want one wrapped item", records)
	}
	wrapped, ok := records[0]["data"].(map[string]any)
	if !ok || wrapped["id"] != "item-1" {
		t.Fatalf("production apify-dataset wrapped item = %#v, want provider item", records[0]["data"])
	}
	if got := requests.Load(); got != 2 {
		t.Fatalf("production apify-dataset requests = %d, want check plus read", got)
	}
}

func TestProductionCloudTrailGenericCheckAndRead(t *testing.T) {
	previousTransport := http.DefaultTransport
	t.Cleanup(func() { http.DefaultTransport = previousTransport })

	var requests atomic.Int64
	http.DefaultTransport = roundTripFunc(func(request *http.Request) (*http.Response, error) {
		requests.Add(1)
		if request.URL.Host != "cloudtrail.us-east-1.amazonaws.com" || request.URL.Path != "/" {
			t.Fatal("CloudTrail request did not use the fixed provider route")
		}
		if request.Header.Get("Authorization") == "" || request.Header.Get("X-Amz-Date") == "" || request.Header.Get("X-Amz-Target") == "" {
			t.Fatal("CloudTrail request omitted declaration-bound SigV4 headers")
		}
		var body map[string]any
		if err := json.NewDecoder(request.Body).Decode(&body); err != nil {
			t.Fatalf("decode CloudTrail body: %v", err)
		}
		var response string
		switch requests.Load() {
		case 1:
			if body["MaxResults"] != float64(1) {
				t.Fatalf("CloudTrail check body = %#v, want MaxResults=1", body)
			}
			response = `{"Events":[]}`
		case 2:
			if body["MaxResults"] != float64(50) {
				t.Fatalf("CloudTrail first read body = %#v, want MaxResults=50", body)
			}
			response = `{"Events":[{"EventId":"one","EventTime":1}],"NextToken":"cursor-1"}`
		case 3:
			if body["NextToken"] != "cursor-1" {
				t.Fatalf("CloudTrail continuation body = %#v, want cursor-1", body)
			}
			response = `{"Events":[{"EventId":"two","EventTime":2}]}`
		default:
			t.Fatal("unexpected CloudTrail request")
		}
		return &http.Response{
			StatusCode: http.StatusOK,
			Header:     make(http.Header),
			Body:       io.NopCloser(strings.NewReader(response)),
			Request:    request,
		}, nil
	})

	connector, ok := New().Get("aws-cloudtrail")
	if !ok {
		t.Fatal("production registry missing aws-cloudtrail")
	}
	config := connectors.RuntimeConfig{Secrets: map[string]string{
		"aws_key_id":     "test-access-key",
		"aws_secret_key": "test-secret-key",
	}}
	if err := connector.Check(context.Background(), config); err != nil {
		t.Fatalf("production aws-cloudtrail Check() error = %v", err)
	}
	var records []connectors.Record
	if err := connector.Read(context.Background(), connectors.ReadRequest{Stream: "management_events", Config: config}, func(record connectors.Record) error {
		records = append(records, record)
		return nil
	}); err != nil {
		t.Fatalf("production aws-cloudtrail Read() error = %v", err)
	}
	if len(records) != 2 || records[0]["EventId"] != "one" || records[1]["EventId"] != "two" {
		t.Fatalf("production aws-cloudtrail records = %#v, want both body-cursor pages", records)
	}
	if got := requests.Load(); got != 3 {
		t.Fatalf("production aws-cloudtrail requests = %d, want check plus two reads", got)
	}
}

func TestProductionBabelforceGenericCheckAndRead(t *testing.T) {
	previousTransport := http.DefaultTransport
	t.Cleanup(func() { http.DefaultTransport = previousTransport })

	var requests atomic.Int64
	http.DefaultTransport = roundTripFunc(func(request *http.Request) (*http.Response, error) {
		requests.Add(1)
		if request.URL.Host != "services.babelforce.com" || request.URL.Path != "/api/v2/calls/reporting/simple" {
			t.Fatal("Babelforce request did not use the declared fixed provider route")
		}
		if request.Header.Get("X-Access-Key-ID") == "" || request.Header.Get("X-Auth-Access-Token") == "" {
			t.Fatal("Babelforce request omitted declared dual-header authentication")
		}
		return &http.Response{
			StatusCode: http.StatusOK,
			Header:     make(http.Header),
			Body:       io.NopCloser(strings.NewReader(`{"items":[{"id":"call-1","dateCreated":"1"}],"pagination":{"current":1,"max":1}}`)),
			Request:    request,
		}, nil
	})

	connector, ok := New().Get("babelforce")
	if !ok {
		t.Fatal("production registry missing babelforce")
	}
	config := connectors.RuntimeConfig{Secrets: map[string]string{
		"access_key_id": "test-access-key",
		"access_token":  "test-access-token",
	}}
	if err := connector.Check(context.Background(), config); err != nil {
		t.Fatalf("production babelforce Check() error = %v", err)
	}
	var records []connectors.Record
	if err := connector.Read(context.Background(), connectors.ReadRequest{Stream: "calls", Config: config}, func(record connectors.Record) error {
		records = append(records, record)
		return nil
	}); err != nil {
		t.Fatalf("production babelforce Read() error = %v", err)
	}
	if len(records) != 1 || records[0]["id"] != "call-1" {
		t.Fatalf("production babelforce records = %#v, want one provider record", records)
	}
	if got := requests.Load(); got != 2 {
		t.Fatalf("production babelforce requests = %d, want check plus read", got)
	}
}

func TestProductionBasecampGenericCheckAndRead(t *testing.T) {
	previousTransport := http.DefaultTransport
	t.Cleanup(func() { http.DefaultTransport = previousTransport })

	var requests atomic.Int64
	http.DefaultTransport = roundTripFunc(func(request *http.Request) (*http.Response, error) {
		requests.Add(1)
		if request.URL.Host != "3.basecampapi.com" || request.URL.Path != "/account-1/projects.json" {
			t.Fatal("Basecamp request did not use the declared account-bound provider route")
		}
		if request.Header.Get("Authorization") == "" {
			t.Fatal("Basecamp request omitted declared bearer authentication")
		}
		body := `[]`
		if requests.Load() == 2 {
			body = `[{"id":1,"updated_at":"2026-01-02T00:00:00Z"}]`
		}
		return &http.Response{StatusCode: http.StatusOK, Header: make(http.Header), Body: io.NopCloser(strings.NewReader(body)), Request: request}, nil
	})

	connector, ok := New().Get("basecamp")
	if !ok {
		t.Fatal("production registry missing basecamp")
	}
	config := connectors.RuntimeConfig{Config: map[string]string{"account_id": "account-1"}, Secrets: map[string]string{"access_token": "test-only"}}
	if err := connector.Check(context.Background(), config); err != nil {
		t.Fatalf("production basecamp Check() error = %v", err)
	}
	var records []connectors.Record
	if err := connector.Read(context.Background(), connectors.ReadRequest{Stream: "projects", Config: config}, func(record connectors.Record) error {
		records = append(records, record)
		return nil
	}); err != nil {
		t.Fatalf("production basecamp Read() error = %v", err)
	}
	if len(records) != 1 || fmt.Sprint(records[0]["id"]) != "1" {
		t.Fatalf("production basecamp records = %#v, want one provider project", records)
	}
	if got := requests.Load(); got != 2 {
		t.Fatalf("production basecamp requests = %d, want check plus read", got)
	}
}

func TestProductionBunnyGenericCheckAndRead(t *testing.T) {
	previousTransport := http.DefaultTransport
	t.Cleanup(func() { http.DefaultTransport = previousTransport })

	var requests atomic.Int64
	http.DefaultTransport = roundTripFunc(func(request *http.Request) (*http.Response, error) {
		requests.Add(1)
		if request.URL.Host != "tenant.bunny.com" || request.URL.Path != "/graphql" {
			t.Fatal("Bunny request did not use the declared tenant provider route")
		}
		if request.Header.Get("Authorization") == "" {
			t.Fatal("Bunny request omitted declared bearer authentication")
		}
		var requestBody map[string]any
		if err := json.NewDecoder(request.Body).Decode(&requestBody); err != nil {
			t.Fatalf("decode Bunny GraphQL body: %v", err)
		}
		query, _ := requestBody["query"].(string)
		if !strings.Contains(query, "accounts") {
			t.Fatal("Bunny request omitted declared accounts GraphQL operation")
		}
		body := `{"data":{"accounts":{"nodes":[],"pageInfo":{"hasNextPage":false}}}}`
		if requests.Load() == 2 {
			body = `{"data":{"accounts":{"nodes":[{"id":"account-1","updatedAt":"2026-01-02T00:00:00Z"}],"pageInfo":{"hasNextPage":false}}}}`
		}
		return &http.Response{StatusCode: http.StatusOK, Header: make(http.Header), Body: io.NopCloser(strings.NewReader(body)), Request: request}, nil
	})

	connector, ok := New().Get("bunny-inc")
	if !ok {
		t.Fatal("production registry missing bunny-inc")
	}
	config := connectors.RuntimeConfig{Config: map[string]string{"subdomain": "tenant"}, Secrets: map[string]string{"apikey": "test-only"}}
	if err := connector.Check(context.Background(), config); err != nil {
		t.Fatalf("production bunny-inc Check() error = %v", err)
	}
	var records []connectors.Record
	if err := connector.Read(context.Background(), connectors.ReadRequest{Stream: "accounts", Config: config}, func(record connectors.Record) error {
		records = append(records, record)
		return nil
	}); err != nil {
		t.Fatalf("production bunny-inc Read() error = %v", err)
	}
	if len(records) != 1 || records[0]["id"] != "account-1" {
		t.Fatalf("production bunny-inc records = %#v, want one provider account", records)
	}
	if got := requests.Load(); got != 2 {
		t.Fatalf("production bunny-inc requests = %d, want check plus read", got)
	}
}

func TestProductionCannyGenericCheckAndRead(t *testing.T) {
	previousTransport := http.DefaultTransport
	t.Cleanup(func() { http.DefaultTransport = previousTransport })

	var requests atomic.Int64
	http.DefaultTransport = roundTripFunc(func(request *http.Request) (*http.Response, error) {
		requests.Add(1)
		if request.URL.Host != "canny.io" {
			t.Fatal("Canny request did not use the fixed provider origin")
		}
		if !strings.HasPrefix(request.Header.Get("Content-Type"), "application/x-www-form-urlencoded") {
			t.Fatal("Canny request did not use the declared form body")
		}
		if err := request.ParseForm(); err != nil {
			t.Fatalf("parse Canny form: %v", err)
		}
		if request.Form.Get("apiKey") == "" {
			t.Fatal("Canny request omitted declared API key form field")
		}
		body := `{"boards":[]}`
		if requests.Load() == 2 {
			if request.URL.Path != "/api/v1/boards/list" {
				t.Fatal("Canny read did not use boards list route")
			}
			body = `{"boards":[{"id":"board-1","created":"2026-01-02T00:00:00Z"}],"hasMore":false}`
		}
		return &http.Response{StatusCode: http.StatusOK, Header: make(http.Header), Body: io.NopCloser(strings.NewReader(body)), Request: request}, nil
	})

	connector, ok := New().Get("canny")
	if !ok {
		t.Fatal("production registry missing canny")
	}
	config := connectors.RuntimeConfig{Secrets: map[string]string{"api_key": "test-only"}}
	if err := connector.Check(context.Background(), config); err != nil {
		t.Fatalf("production canny Check() error = %v", err)
	}
	var records []connectors.Record
	if err := connector.Read(context.Background(), connectors.ReadRequest{Stream: "boards", Config: config}, func(record connectors.Record) error {
		records = append(records, record)
		return nil
	}); err != nil {
		t.Fatalf("production canny Read() error = %v", err)
	}
	if len(records) != 1 || records[0]["id"] != "board-1" {
		t.Fatalf("production canny records = %#v, want one provider board", records)
	}
	if got := requests.Load(); got != 2 {
		t.Fatalf("production canny requests = %d, want check plus read", got)
	}
}

func TestProductionCopperGenericCheckAndRead(t *testing.T) {
	previousTransport := http.DefaultTransport
	t.Cleanup(func() { http.DefaultTransport = previousTransport })

	var requests atomic.Int64
	http.DefaultTransport = roundTripFunc(func(request *http.Request) (*http.Response, error) {
		requests.Add(1)
		if request.URL.Host != "api.copper.com" || request.URL.Path != "/developer_api/v1/people/search" {
			t.Fatal("Copper request did not use the declared fixed provider route")
		}
		if request.Header.Get("X-PW-AccessToken") == "" || request.Header.Get("X-PW-UserEmail") == "" {
			t.Fatal("Copper request omitted declared header authentication")
		}
		var body map[string]any
		if err := json.NewDecoder(request.Body).Decode(&body); err != nil {
			t.Fatalf("decode Copper body: %v", err)
		}
		if requests.Load() == 1 && body["page_size"] != float64(1) {
			t.Fatalf("Copper check body = %#v, want page_size=1", body)
		}
		response := `[]`
		if requests.Load() == 2 {
			if body["page_number"] != "1" {
				t.Fatalf("Copper read body = %#v, want body-carried page_number", body)
			}
			response = `[{"id":1,"date_modified":2}]`
		}
		return &http.Response{StatusCode: http.StatusOK, Header: make(http.Header), Body: io.NopCloser(strings.NewReader(response)), Request: request}, nil
	})

	connector, ok := New().Get("copper")
	if !ok {
		t.Fatal("production registry missing copper")
	}
	config := connectors.RuntimeConfig{Config: map[string]string{"user_email": "user@example.test"}, Secrets: map[string]string{"api_key": "test-only"}}
	if err := connector.Check(context.Background(), config); err != nil {
		t.Fatalf("production copper Check() error = %v", err)
	}
	var records []connectors.Record
	if err := connector.Read(context.Background(), connectors.ReadRequest{Stream: "people", Config: config}, func(record connectors.Record) error {
		records = append(records, record)
		return nil
	}); err != nil {
		t.Fatalf("production copper Read() error = %v", err)
	}
	if len(records) != 1 || fmt.Sprint(records[0]["id"]) != "1" {
		t.Fatalf("production copper records = %#v, want one provider person", records)
	}
	if got := requests.Load(); got != 2 {
		t.Fatalf("production copper requests = %d, want check plus read", got)
	}
}

func TestProductionDixaGenericCheckAndRead(t *testing.T) {
	previousTransport := http.DefaultTransport
	t.Cleanup(func() { http.DefaultTransport = previousTransport })

	var requests atomic.Int64
	http.DefaultTransport = roundTripFunc(func(request *http.Request) (*http.Response, error) {
		requests.Add(1)
		if request.URL.Host != "exports.dixa.io" || request.URL.Path != "/conversation_export" {
			t.Fatal("Dixa request did not use the fixed provider route")
		}
		if request.Header.Get("Authorization") == "" {
			t.Fatal("Dixa request omitted declared bearer authentication")
		}
		if request.URL.Query().Get("updated_after") == "" || request.URL.Query().Get("updated_before") == "" {
			t.Fatal("Dixa request omitted declared export bounds")
		}
		body := `[]`
		if requests.Load() == 2 {
			body = `[{"id":1,"updated_at":2,"queue":{"id":"queue-1"}}]`
		}
		return &http.Response{StatusCode: http.StatusOK, Header: make(http.Header), Body: io.NopCloser(strings.NewReader(body)), Request: request}, nil
	})

	connector, ok := New().Get("dixa")
	if !ok {
		t.Fatal("production registry missing dixa")
	}
	config := connectors.RuntimeConfig{Secrets: map[string]string{"api_token": "test-only"}}
	if err := connector.Check(context.Background(), config); err != nil {
		t.Fatalf("production dixa Check() error = %v", err)
	}
	var records []connectors.Record
	if err := connector.Read(context.Background(), connectors.ReadRequest{Stream: "conversations", Config: config}, func(record connectors.Record) error {
		records = append(records, record)
		return nil
	}); err != nil {
		t.Fatalf("production dixa Read() error = %v", err)
	}
	if len(records) != 1 || fmt.Sprint(records[0]["id"]) != "1" {
		t.Fatalf("production dixa records = %#v, want one provider conversation", records)
	}
	if got := requests.Load(); got != 2 {
		t.Fatalf("production dixa requests = %d, want check plus read", got)
	}
}

func TestProductionFastbillGenericCheckAndRead(t *testing.T) {
	previousTransport := http.DefaultTransport
	t.Cleanup(func() { http.DefaultTransport = previousTransport })

	var requests atomic.Int64
	http.DefaultTransport = roundTripFunc(func(request *http.Request) (*http.Response, error) {
		requests.Add(1)
		if request.URL.Host != "my.fastbill.com" {
			t.Fatalf("FastBill host = %q, want fixed provider host", request.URL.Host)
		}
		if request.URL.Path != "/api/1.0/api.php/" {
			t.Fatalf("FastBill path = %q, want fixed provider route", request.URL.Path)
		}
		if request.Header.Get("Authorization") == "" {
			t.Fatal("FastBill request omitted declared Basic authentication")
		}
		var body map[string]any
		if err := json.NewDecoder(request.Body).Decode(&body); err != nil {
			t.Fatalf("decode FastBill body: %v", err)
		}
		if requests.Load() == 1 && body["SERVICE"] != "customer.get" {
			t.Fatalf("FastBill check body = %#v, want customer.get", body)
		}
		response := `{"RESPONSE":{"CUSTOMERS":[]}}`
		if requests.Load() == 2 {
			if body["OFFSET"] != "0" {
				t.Fatalf("FastBill read body = %#v, want body-carried offset", body)
			}
			response = `{"RESPONSE":{"CUSTOMERS":[{"CUSTOMER_ID":"customer-1"}]}}`
		}
		return &http.Response{StatusCode: http.StatusOK, Header: make(http.Header), Body: io.NopCloser(strings.NewReader(response)), Request: request}, nil
	})

	connector, ok := New().Get("fastbill")
	if !ok {
		t.Fatal("production registry missing fastbill")
	}
	config := connectors.RuntimeConfig{Config: map[string]string{"username": "user"}, Secrets: map[string]string{"api_key": "test-only"}}
	if err := connector.Check(context.Background(), config); err != nil {
		t.Fatalf("production fastbill Check() error = %v", err)
	}
	var records []connectors.Record
	if err := connector.Read(context.Background(), connectors.ReadRequest{Stream: "customers", Config: config}, func(record connectors.Record) error {
		records = append(records, record)
		return nil
	}); err != nil {
		t.Fatalf("production fastbill Read() error = %v", err)
	}
	if len(records) != 1 || records[0]["CUSTOMER_ID"] != "customer-1" {
		t.Fatalf("production fastbill records = %#v, want one provider customer", records)
	}
	if got := requests.Load(); got != 2 {
		t.Fatalf("production fastbill requests = %d, want check plus read", got)
	}
}

func TestProductionFeishuGenericCheckAndRead(t *testing.T) {
	previousTransport := http.DefaultTransport
	t.Cleanup(func() { http.DefaultTransport = previousTransport })

	var requests atomic.Int64
	http.DefaultTransport = roundTripFunc(func(request *http.Request) (*http.Response, error) {
		requests.Add(1)
		if request.URL.Host != "open.feishu.cn" {
			t.Fatalf("Feishu host = %q, want declared provider host", request.URL.Host)
		}

		switch request.URL.Path {
		case "/open-apis/auth/v3/tenant_access_token/internal":
			if request.Method != http.MethodPost {
				t.Fatalf("Feishu token method = %q, want POST", request.Method)
			}
			if !strings.HasPrefix(request.Header.Get("Content-Type"), "application/json") {
				t.Fatal("Feishu token request omitted declared JSON content type")
			}
			var body map[string]string
			if err := json.NewDecoder(request.Body).Decode(&body); err != nil {
				t.Fatalf("decode Feishu token body: %v", err)
			}
			if body["app_id"] == "" || body["app_secret"] == "" {
				t.Fatal("Feishu token request omitted declared app credentials")
			}
			return &http.Response{StatusCode: http.StatusOK, Header: make(http.Header), Body: io.NopCloser(strings.NewReader(`{"code":0,"tenant_access_token":"test-token","expire":7200}`)), Request: request}, nil
		case "/open-apis/bitable/v1/apps/app-1/tables":
			if request.URL.Query().Get("page_size") != "1" {
				t.Fatal("Feishu check omitted declared bounded page size")
			}
			if request.Header.Get("Authorization") != "Bearer test-token" {
				t.Fatal("Feishu check omitted exchanged bearer token")
			}
			return &http.Response{StatusCode: http.StatusOK, Header: make(http.Header), Body: io.NopCloser(strings.NewReader(`{"code":0,"data":{"items":[],"has_more":false}}`)), Request: request}, nil
		case "/open-apis/bitable/v1/apps/app-1/tables/table-1/records":
			if request.Header.Get("Authorization") != "Bearer test-token" {
				t.Fatal("Feishu Bitable request omitted exchanged bearer token")
			}
			return &http.Response{StatusCode: http.StatusOK, Header: make(http.Header), Body: io.NopCloser(strings.NewReader(`{"code":0,"data":{"items":[{"record_id":"record-1","fields":{"Name":"Example"}}],"has_more":false}}`)), Request: request}, nil
		default:
			t.Fatalf("Feishu request used undeclared route %q", request.URL.Path)
			return nil, nil
		}
	})

	connector, ok := New().Get("feishu")
	if !ok {
		t.Fatal("production registry missing feishu")
	}
	config := connectors.RuntimeConfig{
		Config: map[string]string{"region": "feishu.cn", "table_id": "table-1"},
		Secrets: map[string]string{
			"app_id":     "test-app-id",
			"app_secret": "test-app-secret",
			"app_token":  "app-1",
		},
	}
	if err := connector.Check(context.Background(), config); err != nil {
		t.Fatalf("production feishu Check() error = %v", err)
	}
	var records []connectors.Record
	if err := connector.Read(context.Background(), connectors.ReadRequest{Stream: "records", Config: config}, func(record connectors.Record) error {
		records = append(records, record)
		return nil
	}); err != nil {
		t.Fatalf("production feishu Read() error = %v", err)
	}
	if len(records) != 1 || records[0]["record_id"] != "record-1" {
		t.Fatalf("production feishu records = %#v, want one declared Bitable record", records)
	}
	if got := requests.Load(); got != 4 {
		t.Fatalf("production feishu requests = %d, want token plus bounded check and token plus records", got)
	}
}

func TestProductionFeishuRejectsFixtureAndCallerOriginBeforeIO(t *testing.T) {
	previousTransport := http.DefaultTransport
	t.Cleanup(func() { http.DefaultTransport = previousTransport })

	var requests atomic.Int64
	http.DefaultTransport = roundTripFunc(func(*http.Request) (*http.Response, error) {
		requests.Add(1)
		return nil, errors.New("unexpected provider request")
	})

	connector, ok := New().Get("feishu")
	if !ok {
		t.Fatal("production registry missing feishu")
	}
	secrets := map[string]string{"app_id": "test-app-id", "app_secret": "test-app-secret", "app_token": "app-1"}
	for _, config := range []connectors.RuntimeConfig{
		{Config: map[string]string{"table_id": "table-1", "base_url": "https://untrusted.invalid"}, Secrets: secrets},
		{Config: map[string]string{"table_id": "table-1", "mode": "fixture"}, Secrets: secrets},
	} {
		if err := connector.Check(context.Background(), config); err == nil {
			t.Fatal("production feishu Check() accepted forbidden fixture or caller-origin configuration")
		}
	}
	if got := requests.Load(); got != 0 {
		t.Fatalf("production feishu sent %d requests for forbidden configuration, want none", got)
	}
}

func TestProductionFreeAgentGenericCheckAndRead(t *testing.T) {
	previousTransport := http.DefaultTransport
	t.Cleanup(func() { http.DefaultTransport = previousTransport })

	var requests atomic.Int64
	http.DefaultTransport = roundTripFunc(func(request *http.Request) (*http.Response, error) {
		requests.Add(1)
		if request.URL.Host != "api.freeagent.com" {
			t.Fatalf("FreeAgent host = %q, want declared provider host", request.URL.Host)
		}
		switch request.URL.Path {
		case "/v2/token_endpoint":
			if request.Method != http.MethodPost {
				t.Fatalf("FreeAgent token method = %q, want POST", request.Method)
			}
			clientID, clientSecret, ok := request.BasicAuth()
			if !ok || clientID == "" || clientSecret == "" {
				t.Fatal("FreeAgent token request omitted declared Basic client authentication")
			}
			if err := request.ParseForm(); err != nil {
				t.Fatalf("parse FreeAgent token form: %v", err)
			}
			if request.Form.Get("grant_type") != "refresh_token" || request.Form.Get("refresh_token") == "" {
				t.Fatal("FreeAgent token request omitted declared refresh-token grant")
			}
			if request.Form.Get("client_id") != "" || request.Form.Get("client_secret") != "" {
				t.Fatal("FreeAgent token request duplicated Basic credentials in the form")
			}
			return &http.Response{StatusCode: http.StatusOK, Header: make(http.Header), Body: io.NopCloser(strings.NewReader(`{"access_token":"test-token","expires_in":3600}`)), Request: request}, nil
		case "/v2/contacts":
			if request.Header.Get("Authorization") != "Bearer test-token" {
				t.Fatal("FreeAgent contacts request omitted exchanged bearer token")
			}
			if request.URL.Query().Get("per_page") == "" {
				t.Fatal("FreeAgent contacts request omitted declared page size")
			}
			response := `{"contacts":[]}`
			if request.URL.Query().Get("per_page") == "100" {
				response = `{"contacts":[{"url":"https://api.freeagent.com/v2/contacts/1","updated_at":"2026-01-02T00:00:00Z"}]}`
			}
			return &http.Response{StatusCode: http.StatusOK, Header: make(http.Header), Body: io.NopCloser(strings.NewReader(response)), Request: request}, nil
		default:
			t.Fatalf("FreeAgent request used undeclared route %q", request.URL.Path)
			return nil, nil
		}
	})

	connector, ok := New().Get("free-agent-connector")
	if !ok {
		t.Fatal("production registry missing free-agent-connector")
	}
	config := connectors.RuntimeConfig{Secrets: map[string]string{
		"client_id":              "test-client-id",
		"client_secret":          "test-client-secret",
		"client_refresh_token_2": "test-refresh-token",
	}}
	if err := connector.Check(context.Background(), config); err != nil {
		t.Fatalf("production FreeAgent Check() error = %v", err)
	}
	var records []connectors.Record
	if err := connector.Read(context.Background(), connectors.ReadRequest{Stream: "contacts", Config: config}, func(record connectors.Record) error {
		records = append(records, record)
		return nil
	}); err != nil {
		t.Fatalf("production FreeAgent Read() error = %v", err)
	}
	if len(records) != 1 || records[0]["url"] != "https://api.freeagent.com/v2/contacts/1" {
		t.Fatalf("production FreeAgent records = %#v, want one declared contact", records)
	}
	if got := requests.Load(); got != 4 {
		t.Fatalf("production FreeAgent requests = %d, want token plus check and token plus read", got)
	}
}

func TestProductionFreeAgentRejectsFixtureAndCallerOriginBeforeIO(t *testing.T) {
	previousTransport := http.DefaultTransport
	t.Cleanup(func() { http.DefaultTransport = previousTransport })

	var requests atomic.Int64
	http.DefaultTransport = roundTripFunc(func(*http.Request) (*http.Response, error) {
		requests.Add(1)
		return nil, errors.New("unexpected provider request")
	})

	connector, ok := New().Get("free-agent-connector")
	if !ok {
		t.Fatal("production registry missing free-agent-connector")
	}
	secrets := map[string]string{
		"client_id":              "test-client-id",
		"client_secret":          "test-client-secret",
		"client_refresh_token_2": "test-refresh-token",
	}
	for _, config := range []connectors.RuntimeConfig{
		{Config: map[string]string{"base_url": "https://untrusted.invalid"}, Secrets: secrets},
		{Config: map[string]string{"mode": "fixture"}, Secrets: secrets},
	} {
		if err := connector.Check(context.Background(), config); err == nil {
			t.Fatal("production FreeAgent Check() accepted forbidden fixture or caller-origin configuration")
		}
	}
	if got := requests.Load(); got != 0 {
		t.Fatalf("production FreeAgent sent %d requests for forbidden configuration, want none", got)
	}
}

func TestProductionFreightviewGenericCheckAndRead(t *testing.T) {
	previousTransport := http.DefaultTransport
	t.Cleanup(func() { http.DefaultTransport = previousTransport })

	var requests atomic.Int64
	http.DefaultTransport = roundTripFunc(func(request *http.Request) (*http.Response, error) {
		requests.Add(1)
		if request.URL.Host != "api.freightview.com" {
			t.Fatalf("Freightview host = %q, want declared provider host", request.URL.Host)
		}
		switch request.URL.Path {
		case "/v2.0/auth/token":
			if request.Method != http.MethodPost {
				t.Fatalf("Freightview token method = %q, want POST", request.Method)
			}
			if err := request.ParseForm(); err != nil {
				t.Fatalf("parse Freightview token form: %v", err)
			}
			if request.Form.Get("grant_type") != "client_credentials" || request.Form.Get("client_id") == "" || request.Form.Get("client_secret") == "" {
				t.Fatal("Freightview token request omitted declared client-credentials grant")
			}
			return &http.Response{StatusCode: http.StatusOK, Header: make(http.Header), Body: io.NopCloser(strings.NewReader(`{"access_token":"test-token","expires_in":3600}`)), Request: request}, nil
		case "/v2.0/shipments":
			if request.Header.Get("Authorization") != "Bearer test-token" {
				t.Fatal("Freightview shipments request omitted exchanged bearer token")
			}
			response := `{"shipments":[]}`
			if request.URL.Query().Get("limit") == "" {
				response = `{"shipments":[{"shipmentId":"shipment-1"}],"continuationToken":null}`
			}
			return &http.Response{StatusCode: http.StatusOK, Header: make(http.Header), Body: io.NopCloser(strings.NewReader(response)), Request: request}, nil
		case "/v2.0/shipments/shipment-1/quotes":
			if request.Header.Get("Authorization") != "Bearer test-token" {
				t.Fatal("Freightview quotes request omitted exchanged bearer token")
			}
			return &http.Response{StatusCode: http.StatusOK, Header: make(http.Header), Body: io.NopCloser(strings.NewReader(`{"quotes":[{"quoteId":"quote-1"}]}`)), Request: request}, nil
		default:
			t.Fatalf("Freightview request used undeclared route %q", request.URL.Path)
			return nil, nil
		}
	})

	connector, ok := New().Get("freightview")
	if !ok {
		t.Fatal("production registry missing freightview")
	}
	config := connectors.RuntimeConfig{Secrets: map[string]string{"client_id": "test-client-id", "client_secret": "test-client-secret"}}
	if err := connector.Check(context.Background(), config); err != nil {
		t.Fatalf("production Freightview Check() error = %v", err)
	}
	var records []connectors.Record
	if err := connector.Read(context.Background(), connectors.ReadRequest{Stream: "shipments", Config: config}, func(record connectors.Record) error {
		records = append(records, record)
		return nil
	}); err != nil {
		t.Fatalf("production Freightview Read() error = %v", err)
	}
	if len(records) != 1 || records[0]["shipmentId"] != "shipment-1" {
		t.Fatalf("production Freightview records = %#v, want one declared shipment", records)
	}
	var quotes []connectors.Record
	if err := connector.Read(context.Background(), connectors.ReadRequest{Stream: "quotes", Config: config}, func(record connectors.Record) error {
		quotes = append(quotes, record)
		return nil
	}); err != nil {
		t.Fatalf("production Freightview quotes Read() error = %v", err)
	}
	if len(quotes) != 1 || quotes[0]["quoteId"] != "quote-1" || quotes[0]["shipment_id"] != "shipment-1" {
		t.Fatalf("production Freightview quotes = %#v, want one stamped declared quote", quotes)
	}
	if got := requests.Load(); got != 7 {
		t.Fatalf("production Freightview requests = %d, want check, root read, and declared fan-out", got)
	}
}

func TestProductionFreightviewRejectsFixtureAndCallerOriginBeforeIO(t *testing.T) {
	previousTransport := http.DefaultTransport
	t.Cleanup(func() { http.DefaultTransport = previousTransport })

	var requests atomic.Int64
	http.DefaultTransport = roundTripFunc(func(*http.Request) (*http.Response, error) {
		requests.Add(1)
		return nil, errors.New("unexpected provider request")
	})

	connector, ok := New().Get("freightview")
	if !ok {
		t.Fatal("production registry missing freightview")
	}
	secrets := map[string]string{"client_id": "test-client-id", "client_secret": "test-client-secret"}
	for _, config := range []connectors.RuntimeConfig{
		{Config: map[string]string{"base_url": "https://untrusted.invalid"}, Secrets: secrets},
		{Config: map[string]string{"mode": "fixture"}, Secrets: secrets},
	} {
		if err := connector.Check(context.Background(), config); err == nil {
			t.Fatal("production Freightview Check() accepted forbidden fixture or caller-origin configuration")
		}
	}
	if got := requests.Load(); got != 0 {
		t.Fatalf("production Freightview sent %d requests for forbidden configuration, want none", got)
	}
}

func TestProductionGoogleAnalyticsDataGenericCheckAndRead(t *testing.T) {
	previousTransport := http.DefaultTransport
	t.Cleanup(func() { http.DefaultTransport = previousTransport })

	var requests atomic.Int64
	http.DefaultTransport = roundTripFunc(func(request *http.Request) (*http.Response, error) {
		requests.Add(1)
		if request.URL.Host != "analyticsdata.googleapis.com" || request.URL.Path != "/v1beta/properties/123:runReport" {
			t.Fatalf("GA4 route = %s, want declared runReport provider route", request.URL)
		}
		if request.Method != http.MethodPost || request.Header.Get("Authorization") != "Bearer test-token" {
			t.Fatal("GA4 request omitted declared POST bearer authentication")
		}
		var body map[string]any
		if err := json.NewDecoder(request.Body).Decode(&body); err != nil {
			t.Fatalf("decode GA4 body: %v", err)
		}
		if body["limit"] == nil || body["dimensions"] == nil || body["metrics"] == nil {
			t.Fatalf("GA4 body = %#v, want declared report dimensions, metrics, and limit", body)
		}
		return &http.Response{StatusCode: http.StatusOK, Header: make(http.Header), Body: io.NopCloser(strings.NewReader(`{
			"dimensionHeaders":[{"name":"date"}],
			"metricHeaders":[{"name":"activeUsers"},{"name":"newUsers"},{"name":"sessions"}],
			"rows":[{"dimensionValues":[{"value":"20260102"}],"metricValues":[{"value":"42"},{"value":"3"},{"value":"7"}]}],
			"rowCount":"1"
		}`)), Request: request}, nil
	})

	connector, ok := New().Get("google-analytics-data-api")
	if !ok {
		t.Fatal("production registry missing google-analytics-data-api")
	}
	config := connectors.RuntimeConfig{Config: map[string]string{"property_ids": "123"}, Secrets: map[string]string{"access_token": "test-token"}}
	if err := connector.Check(context.Background(), config); err != nil {
		t.Fatalf("production GA4 Check() error = %v", err)
	}
	var records []connectors.Record
	if err := connector.Read(context.Background(), connectors.ReadRequest{Stream: "daily_active_users", Config: config}, func(record connectors.Record) error {
		records = append(records, record)
		return nil
	}); err != nil {
		t.Fatalf("production GA4 Read() error = %v", err)
	}
	if len(records) != 1 || records[0]["property_id"] != "123" || records[0]["date"] != "20260102" || records[0]["activeUsers"] != "42" {
		t.Fatalf("production GA4 records = %#v, want source-declared flattened report row", records)
	}
	if got := requests.Load(); got != 2 {
		t.Fatalf("production GA4 requests = %d, want bounded check plus read", got)
	}
}

func TestProductionGoogleAnalyticsDataRejectsFixtureAndCallerOriginBeforeIO(t *testing.T) {
	previousTransport := http.DefaultTransport
	t.Cleanup(func() { http.DefaultTransport = previousTransport })

	var requests atomic.Int64
	http.DefaultTransport = roundTripFunc(func(*http.Request) (*http.Response, error) {
		requests.Add(1)
		return nil, errors.New("unexpected provider request")
	})

	connector, ok := New().Get("google-analytics-data-api")
	if !ok {
		t.Fatal("production registry missing google-analytics-data-api")
	}
	secrets := map[string]string{"access_token": "test-token"}
	for _, config := range []connectors.RuntimeConfig{
		{Config: map[string]string{"property_ids": "123", "base_url": "https://untrusted.invalid"}, Secrets: secrets},
		{Config: map[string]string{"property_ids": "123", "mode": "fixture"}, Secrets: secrets},
	} {
		if err := connector.Check(context.Background(), config); err == nil {
			t.Fatal("production GA4 Check() accepted forbidden fixture or caller-origin configuration")
		}
	}
	if got := requests.Load(); got != 0 {
		t.Fatalf("production GA4 sent %d requests for forbidden configuration, want none", got)
	}
}

func TestProductionGoogleClassroomGenericCheckAndRead(t *testing.T) {
	previousTransport := http.DefaultTransport
	t.Cleanup(func() { http.DefaultTransport = previousTransport })

	var requests atomic.Int64
	http.DefaultTransport = roundTripFunc(func(request *http.Request) (*http.Response, error) {
		requests.Add(1)
		switch request.URL.Host {
		case "oauth2.googleapis.com":
			if request.URL.Path != "/token" || request.Method != http.MethodPost {
				t.Fatalf("Classroom token route = %s %s, want declared token POST", request.Method, request.URL)
			}
			if err := request.ParseForm(); err != nil {
				t.Fatalf("parse Classroom token form: %v", err)
			}
			if request.Form.Get("grant_type") != "refresh_token" || request.Form.Get("refresh_token") == "" {
				t.Fatal("Classroom token request omitted declared refresh-token grant")
			}
			return &http.Response{StatusCode: http.StatusOK, Header: make(http.Header), Body: io.NopCloser(strings.NewReader(`{"access_token":"test-token","expires_in":3600}`)), Request: request}, nil
		case "classroom.googleapis.com":
			if request.Header.Get("Authorization") != "Bearer test-token" {
				t.Fatal("Classroom request omitted exchanged bearer token")
			}
			switch request.URL.Path {
			case "/v1/courses":
				return &http.Response{StatusCode: http.StatusOK, Header: make(http.Header), Body: io.NopCloser(strings.NewReader(`{"courses":[{"id":"course-1","name":"Example"}]}`)), Request: request}, nil
			case "/v1/courses/course-1/teachers":
				return &http.Response{StatusCode: http.StatusOK, Header: make(http.Header), Body: io.NopCloser(strings.NewReader(`{"teachers":[{"userId":"teacher-1","profile":{"name":{"fullName":"Teacher Example"},"emailAddress":"teacher@example.test"}}]}`)), Request: request}, nil
			}
		}
		t.Fatalf("Classroom request used undeclared route %s", request.URL)
		return nil, nil
	})

	connector, ok := New().Get("google-classroom")
	if !ok {
		t.Fatal("production registry missing google-classroom")
	}
	config := connectors.RuntimeConfig{Secrets: map[string]string{"client_id": "test-client", "client_secret": "test-secret", "client_refresh_token": "test-refresh"}}
	if err := connector.Check(context.Background(), config); err != nil {
		t.Fatalf("production Classroom Check() error = %v", err)
	}
	var courses []connectors.Record
	if err := connector.Read(context.Background(), connectors.ReadRequest{Stream: "courses", Config: config}, func(record connectors.Record) error {
		courses = append(courses, record)
		return nil
	}); err != nil {
		t.Fatalf("production Classroom courses Read() error = %v", err)
	}
	var teachers []connectors.Record
	if err := connector.Read(context.Background(), connectors.ReadRequest{Stream: "teachers", Config: config}, func(record connectors.Record) error {
		teachers = append(teachers, record)
		return nil
	}); err != nil {
		t.Fatalf("production Classroom teachers Read() error = %v", err)
	}
	if len(courses) != 1 || courses[0]["id"] != "course-1" || len(teachers) != 1 || teachers[0]["courseId"] != "course-1" || teachers[0]["fullName"] != "Teacher Example" {
		t.Fatalf("production Classroom records = courses:%#v teachers:%#v", courses, teachers)
	}
	if got := requests.Load(); got != 7 {
		t.Fatalf("production Classroom requests = %d, want token/check plus root and fan-out reads", got)
	}
}

func TestProductionGooglePageSpeedGenericCheckAndRead(t *testing.T) {
	previousTransport := http.DefaultTransport
	t.Cleanup(func() { http.DefaultTransport = previousTransport })

	var requests atomic.Int64
	http.DefaultTransport = roundTripFunc(func(request *http.Request) (*http.Response, error) {
		requests.Add(1)
		if request.URL.Host != "www.googleapis.com" || request.URL.Path != "/pagespeedonline/v5/runPagespeed" {
			t.Fatalf("PageSpeed route = %s, want fixed provider runPagespeed route", request.URL)
		}
		if request.Method != http.MethodGet {
			t.Fatalf("PageSpeed method = %s, want GET", request.Method)
		}
		query := request.URL.Query()
		if query.Get("url") == "https://example.com" {
			if query.Get("strategy") != "desktop" || query.Get("category") != "performance" {
				t.Fatalf("PageSpeed check query = %s, want fixed bounded probe", request.URL.RawQuery)
			}
		} else {
			if got := strings.Join(query["category"], ","); got != "accessibility,best-practices,performance,pwa,seo" {
				t.Fatalf("PageSpeed categories = %q, want fixed declared categories", got)
			}
		}
		return &http.Response{
			StatusCode: http.StatusOK,
			Header:     make(http.Header),
			Body: io.NopCloser(strings.NewReader(`{
				"id":"report",
				"kind":"pagespeedonline#result",
				"analysisUTCTimestamp":"2026-01-02T00:00:00Z",
				"loadingExperience":{"overall_category":"FAST"},
				"lighthouseResult":{
					"requestedUrl":"https://one.example",
					"finalUrl":"https://one.example",
					"lighthouseVersion":"12.0.0",
					"fetchTime":"2026-01-02T00:00:00Z",
					"categories":{
						"performance":{"score":0.91},
						"accessibility":{"score":0.88},
						"best-practices":{"score":0.92},
						"seo":{"score":1},
						"pwa":{"score":0.73}
					}
				}
			}`)),
			Request: request,
		}, nil
	})

	connector, ok := New().Get("google-pagespeed-insights")
	if !ok {
		t.Fatal("production registry missing google-pagespeed-insights")
	}
	config := connectors.RuntimeConfig{Config: map[string]string{
		"urls":       "https://one.example,https://two.example",
		"strategies": "mobile,desktop",
	}}
	if err := connector.Check(context.Background(), config); err != nil {
		t.Fatalf("production PageSpeed Check() error = %v", err)
	}
	var records []connectors.Record
	if err := connector.Read(context.Background(), connectors.ReadRequest{Stream: "pagespeed_reports", Config: config}, func(record connectors.Record) error {
		records = append(records, record)
		return nil
	}); err != nil {
		t.Fatalf("production PageSpeed Read() error = %v", err)
	}
	if len(records) != 4 {
		t.Fatalf("production PageSpeed records = %#v, want one record per URL/strategy pair", records)
	}
	for _, record := range records {
		if record["url"] == nil || record["strategy"] == nil || fmt.Sprint(record["performance_score"]) != "0.91" || record["overall_loading_experience"] != "FAST" {
			t.Fatalf("production PageSpeed record = %#v, want declared fan-out stamps and projected report fields", record)
		}
	}
	if got := requests.Load(); got != 5 {
		t.Fatalf("production PageSpeed requests = %d, want one bounded check plus four URL/strategy reads", got)
	}
}

func TestProductionLessAnnoyingCRMGenericCheckAndRead(t *testing.T) {
	previousTransport := http.DefaultTransport
	t.Cleanup(func() { http.DefaultTransport = previousTransport })

	var requests atomic.Int64
	http.DefaultTransport = roundTripFunc(func(request *http.Request) (*http.Response, error) {
		requests.Add(1)
		if request.URL.Host != "api.lessannoyingcrm.com" || request.URL.Path != "/v2/" {
			t.Fatalf("Less Annoying CRM route = %s, want fixed v2 provider route", request.URL)
		}
		if request.Method != http.MethodPost || request.Header.Get("Authorization") != "test-key" {
			t.Fatal("Less Annoying CRM request omitted declared POST API-key header authentication")
		}
		var body map[string]any
		if err := json.NewDecoder(request.Body).Decode(&body); err != nil {
			t.Fatalf("decode Less Annoying CRM body: %v", err)
		}
		switch body["Function"] {
		case "GetUsers":
			if request.URL.RawQuery != "" {
				t.Fatalf("Less Annoying CRM check query = %s, want no bulk pagination", request.URL.RawQuery)
			}
			return &http.Response{StatusCode: http.StatusOK, Header: make(http.Header), Body: io.NopCloser(strings.NewReader(`[{"UserId":"user-1"}]`)), Request: request}, nil
		case "GetContacts":
			if request.URL.Query().Get("Page") != "1" || request.URL.Query().Get("MaxNumberOfResults") != "500" {
				t.Fatalf("Less Annoying CRM query = %s, want declared first page", request.URL.RawQuery)
			}
			return &http.Response{StatusCode: http.StatusOK, Header: make(http.Header), Body: io.NopCloser(strings.NewReader(`{"Results":[{"ContactId":"contact-1","Name":"Example","IsCompany":false}]}`)), Request: request}, nil
		default:
			t.Fatalf("Less Annoying CRM Function = %#v, want declared check or contacts read", body["Function"])
			return nil, nil
		}
	})

	connector, ok := New().Get("less-annoying-crm")
	if !ok {
		t.Fatal("production registry missing less-annoying-crm")
	}
	config := connectors.RuntimeConfig{Secrets: map[string]string{"api_key": "test-key"}}
	if err := connector.Check(context.Background(), config); err != nil {
		t.Fatalf("production Less Annoying CRM Check() error = %v", err)
	}
	var records []connectors.Record
	if err := connector.Read(context.Background(), connectors.ReadRequest{Stream: "contacts", Config: config}, func(record connectors.Record) error {
		records = append(records, record)
		return nil
	}); err != nil {
		t.Fatalf("production Less Annoying CRM Read() error = %v", err)
	}
	if len(records) != 1 || records[0]["ContactId"] != "contact-1" || records[0]["Name"] != "Example" {
		t.Fatalf("production Less Annoying CRM records = %#v, want declared contact projection", records)
	}
	if got := requests.Load(); got != 2 {
		t.Fatalf("production Less Annoying CRM requests = %d, want bounded check plus contacts read", got)
	}
}

func TestProductionLokaliseGenericCheckAndRead(t *testing.T) {
	previousTransport := http.DefaultTransport
	t.Cleanup(func() { http.DefaultTransport = previousTransport })

	var requests atomic.Int64
	http.DefaultTransport = roundTripFunc(func(request *http.Request) (*http.Response, error) {
		requests.Add(1)
		if request.URL.Host != "api.lokalise.com" || request.URL.Path != "/api2/projects/project-1/languages" {
			t.Fatalf("Lokalise route = %s, want fixed project languages route", request.URL)
		}
		if request.Method != http.MethodGet || request.Header.Get("X-Api-Token") != "test-key" {
			t.Fatal("Lokalise request omitted declared GET API-key header authentication")
		}
		if request.URL.Query().Get("limit") == "" {
			t.Fatalf("Lokalise query = %s, want declared limit", request.URL.RawQuery)
		}
		return &http.Response{StatusCode: http.StatusOK, Header: make(http.Header), Body: io.NopCloser(strings.NewReader(`{"languages":[{"lang_id":1,"lang_iso":"en","lang_name":"English","is_rtl":false,"plural_forms":["one","other"]}]}`)), Request: request}, nil
	})

	connector, ok := New().Get("lokalise")
	if !ok {
		t.Fatal("production registry missing lokalise")
	}
	config := connectors.RuntimeConfig{Config: map[string]string{"project_id": "project-1"}, Secrets: map[string]string{"api_key": "test-key"}}
	if err := connector.Check(context.Background(), config); err != nil {
		t.Fatalf("production Lokalise Check() error = %v", err)
	}
	var records []connectors.Record
	if err := connector.Read(context.Background(), connectors.ReadRequest{Stream: "languages", Config: config}, func(record connectors.Record) error {
		records = append(records, record)
		return nil
	}); err != nil {
		t.Fatalf("production Lokalise Read() error = %v", err)
	}
	if len(records) != 1 || fmt.Sprint(records[0]["lang_id"]) != "1" || records[0]["lang_name"] != "English" {
		t.Fatalf("production Lokalise records = %#v, want declared language projection", records)
	}
	if got := requests.Load(); got != 2 {
		t.Fatalf("production Lokalise requests = %d, want bounded check plus languages read", got)
	}
}

func TestProductionMendeleyGenericCheckAndRead(t *testing.T) {
	previousTransport := http.DefaultTransport
	t.Cleanup(func() { http.DefaultTransport = previousTransport })

	var requests atomic.Int64
	http.DefaultTransport = roundTripFunc(func(request *http.Request) (*http.Response, error) {
		requests.Add(1)
		switch request.URL.Host {
		case "api.mendeley.com":
			if request.URL.Path == "/oauth/token" {
				if request.Method != http.MethodPost {
					t.Fatalf("Mendeley token method = %s, want POST", request.Method)
				}
				return &http.Response{StatusCode: http.StatusOK, Header: make(http.Header), Body: io.NopCloser(strings.NewReader(`{"access_token":"test-token","expires_in":3600}`)), Request: request}, nil
			}
			if request.URL.Path != "/documents" || request.Method != http.MethodGet || request.Header.Get("Authorization") != "Bearer test-token" {
				t.Fatalf("Mendeley request = %s %s, want declared bearer documents route", request.Method, request.URL)
			}
			return &http.Response{StatusCode: http.StatusOK, Header: make(http.Header), Body: io.NopCloser(strings.NewReader(`[{"id":"document-1","title":"Example","last_modified":"2026-01-02T00:00:00Z"}]`)), Request: request}, nil
		default:
			t.Fatalf("Mendeley host = %s, want fixed provider host", request.URL.Host)
			return nil, nil
		}
	})

	connector, ok := New().Get("mendeley")
	if !ok {
		t.Fatal("production registry missing mendeley")
	}
	config := connectors.RuntimeConfig{Secrets: map[string]string{"client_id": "test-client", "client_secret": "test-secret", "client_refresh_token": "test-refresh"}}
	if err := connector.Check(context.Background(), config); err != nil {
		t.Fatalf("production Mendeley Check() error = %v", err)
	}
	var records []connectors.Record
	if err := connector.Read(context.Background(), connectors.ReadRequest{Stream: "documents", Config: config}, func(record connectors.Record) error {
		records = append(records, record)
		return nil
	}); err != nil {
		t.Fatalf("production Mendeley Read() error = %v", err)
	}
	if len(records) != 1 || records[0]["id"] != "document-1" || records[0]["title"] != "Example" {
		t.Fatalf("production Mendeley records = %#v, want declared document projection", records)
	}
	if got := requests.Load(); got != 4 {
		t.Fatalf("production Mendeley requests = %d, want one token exchange for each bounded check and document read", got)
	}
}

func TestProductionRootlyGenericCheckAndRead(t *testing.T) {
	previousTransport := http.DefaultTransport
	t.Cleanup(func() { http.DefaultTransport = previousTransport })

	var requests atomic.Int64
	http.DefaultTransport = roundTripFunc(func(request *http.Request) (*http.Response, error) {
		requests.Add(1)
		if request.URL.Host != "api.rootly.com" || request.URL.Path != "/v1/incidents" || request.Method != http.MethodGet {
			t.Fatalf("Rootly request = %s %s, want fixed incidents GET route", request.Method, request.URL)
		}
		if request.Header.Get("Authorization") != "Bearer test-token" {
			t.Fatal("Rootly request omitted declared bearer authentication")
		}
		if request.URL.Query().Get("page[size]") == "" {
			t.Fatalf("Rootly query = %s, want declared page size", request.URL.RawQuery)
		}
		return &http.Response{StatusCode: http.StatusOK, Header: make(http.Header), Body: io.NopCloser(strings.NewReader(`{"data":[{"id":"incident-1","attributes":{"title":"Example","status":"open"}}],"links":{"next":""}}`)), Request: request}, nil
	})

	connector, ok := New().Get("rootly")
	if !ok {
		t.Fatal("production registry missing rootly")
	}
	config := connectors.RuntimeConfig{Secrets: map[string]string{"api_key": "test-token"}}
	if err := connector.Check(context.Background(), config); err != nil {
		t.Fatalf("production Rootly Check() error = %v", err)
	}
	var records []connectors.Record
	if err := connector.Read(context.Background(), connectors.ReadRequest{Stream: "incidents", Config: config}, func(record connectors.Record) error {
		records = append(records, record)
		return nil
	}); err != nil {
		t.Fatalf("production Rootly Read() error = %v", err)
	}
	if len(records) != 1 || records[0]["id"] != "incident-1" || records[0]["title"] != "Example" || records[0]["status"] != "open" {
		t.Fatalf("production Rootly records = %#v, want flattened JSON:API incident", records)
	}
	if got := requests.Load(); got != 2 {
		t.Fatalf("production Rootly requests = %d, want bounded check plus incidents read", got)
	}
}

func TestProductionMyHoursGenericCheckAndRead(t *testing.T) {
	previousTransport := http.DefaultTransport
	t.Cleanup(func() { http.DefaultTransport = previousTransport })

	var requests atomic.Int64
	http.DefaultTransport = roundTripFunc(func(request *http.Request) (*http.Response, error) {
		requests.Add(1)
		if request.URL.Host != "api2.myhours.com" {
			t.Fatalf("My Hours host = %s, want fixed provider host", request.URL.Host)
		}
		switch request.URL.Path {
		case "/api/tokens/login":
			if request.Method != http.MethodPost {
				t.Fatalf("My Hours login method = %s, want POST", request.Method)
			}
			var body map[string]string
			if err := json.NewDecoder(request.Body).Decode(&body); err != nil {
				t.Fatalf("decode My Hours login body: %v", err)
			}
			if body["email"] != "person@example.test" || body["password"] != "test-password" {
				t.Fatal("My Hours login body omitted the declared email/password binding")
			}
			return &http.Response{StatusCode: http.StatusOK, Header: make(http.Header), Body: io.NopCloser(strings.NewReader(`{"accessToken":"test-token"}`)), Request: request}, nil
		case "/api/Clients":
			if request.Method != http.MethodGet || request.Header.Get("Authorization") != "Bearer test-token" || request.Header.Get("api-version") != "1.0" {
				t.Fatal("My Hours data request omitted the declared bearer or static API version header")
			}
			return &http.Response{StatusCode: http.StatusOK, Header: make(http.Header), Body: io.NopCloser(strings.NewReader(`[{"id":1,"name":"Example","active":true}]`)), Request: request}, nil
		case "/api/Reports/activity":
			if request.URL.Query().Get("DateFrom") != "2026-01-01" || request.URL.Query().Get("DateTo") != "2026-01-02" {
				t.Fatalf("My Hours date window = %s, want declared first UTC window", request.URL.RawQuery)
			}
			return &http.Response{StatusCode: http.StatusOK, Header: make(http.Header), Body: io.NopCloser(strings.NewReader(`[{"logId":1,"date":"2026-01-01","project_name":"Example"}]`)), Request: request}, nil
		default:
			t.Fatalf("My Hours route = %s, want declared login or data route", request.URL)
			return nil, nil
		}
	})

	connector, ok := New().Get("my-hours")
	if !ok {
		t.Fatal("production registry missing my-hours")
	}
	config := connectors.RuntimeConfig{Config: map[string]string{"email": "person@example.test", "start_date": "2026-01-01", "end_date": "2026-01-02", "logs_batch_size": "2"}, Secrets: map[string]string{"password": "test-password"}}
	if err := connector.Check(context.Background(), config); err != nil {
		t.Fatalf("production My Hours Check() error = %v", err)
	}
	var records []connectors.Record
	if err := connector.Read(context.Background(), connectors.ReadRequest{Stream: "clients", Config: config}, func(record connectors.Record) error {
		records = append(records, record)
		return nil
	}); err != nil {
		t.Fatalf("production My Hours Read() error = %v", err)
	}
	if len(records) != 1 || fmt.Sprint(records[0]["id"]) != "1" || records[0]["name"] != "Example" {
		t.Fatalf("production My Hours records = %#v, want declared clients projection", records)
	}
	var timeLogs []connectors.Record
	if err := connector.Read(context.Background(), connectors.ReadRequest{Stream: "time_logs", Config: config}, func(record connectors.Record) error {
		timeLogs = append(timeLogs, record)
		return nil
	}); err != nil {
		t.Fatalf("production My Hours time_logs Read() error = %v", err)
	}
	if len(timeLogs) != 1 || fmt.Sprint(timeLogs[0]["logId"]) != "1" {
		t.Fatalf("production My Hours time_logs = %#v, want declared window record", timeLogs)
	}
	if got := requests.Load(); got != 6 {
		t.Fatalf("production My Hours requests = %d, want token plus data request for check, clients, and time_logs", got)
	}
}

func TestProductionSafetyCultureGenericCheckAndRead(t *testing.T) {
	previousTransport := http.DefaultTransport
	t.Cleanup(func() { http.DefaultTransport = previousTransport })

	var requests atomic.Int64
	http.DefaultTransport = roundTripFunc(func(request *http.Request) (*http.Response, error) {
		requests.Add(1)
		if request.URL.Host != "api.safetyculture.io" || request.URL.Path != "/audits" || request.Method != http.MethodGet {
			t.Fatalf("SafetyCulture request = %s %s, want fixed audits GET route", request.Method, request.URL)
		}
		if request.Header.Get("Authorization") != "Bearer test-token" || request.URL.Query().Get("page_size") == "" {
			t.Fatal("SafetyCulture request omitted declared bearer auth or page size")
		}
		return &http.Response{StatusCode: http.StatusOK, Header: make(http.Header), Body: io.NopCloser(strings.NewReader(`{"audits":[{"id":"audit-1","name":"Example","modified_at":"2026-01-01T00:00:00Z"}],"links":{"next":""}}`)), Request: request}, nil
	})

	connector, ok := New().Get("safetyculture")
	if !ok {
		t.Fatal("production registry missing safetyculture")
	}
	config := connectors.RuntimeConfig{Secrets: map[string]string{"access_token": "test-token"}}
	if err := connector.Check(context.Background(), config); err != nil {
		t.Fatalf("production SafetyCulture Check() error = %v", err)
	}
	var records []connectors.Record
	if err := connector.Read(context.Background(), connectors.ReadRequest{Stream: "audits", Config: config}, func(record connectors.Record) error {
		records = append(records, record)
		return nil
	}); err != nil {
		t.Fatalf("production SafetyCulture Read() error = %v", err)
	}
	if len(records) != 1 || records[0]["id"] != "audit-1" || records[0]["name"] != "Example" {
		t.Fatalf("production SafetyCulture records = %#v, want declared audit projection", records)
	}
	if got := requests.Load(); got != 2 {
		t.Fatalf("production SafetyCulture requests = %d, want bounded check plus audits read", got)
	}
}

func TestProductionPocketGenericCheckAndRead(t *testing.T) {
	previousTransport := http.DefaultTransport
	t.Cleanup(func() { http.DefaultTransport = previousTransport })

	var requests atomic.Int64
	http.DefaultTransport = roundTripFunc(func(request *http.Request) (*http.Response, error) {
		requests.Add(1)
		if request.URL.Host != "getpocket.com" || request.URL.Path != "/v3/get" || request.Method != http.MethodPost {
			t.Fatalf("Pocket request = %s %s, want fixed retrieve route", request.Method, request.URL)
		}
		var body map[string]any
		if err := json.NewDecoder(request.Body).Decode(&body); err != nil {
			t.Fatalf("decode Pocket body: %v", err)
		}
		if body["consumer_key"] != "test-consumer" || body["access_token"] != "test-access" || body["count"] == nil || body["offset"] == nil {
			t.Fatal("Pocket request omitted declared credential and pagination bindings")
		}
		if body["count"] == float64(1) {
			return &http.Response{StatusCode: http.StatusOK, Header: make(http.Header), Body: io.NopCloser(strings.NewReader(`{"list":{}}`)), Request: request}, nil
		}
		return &http.Response{StatusCode: http.StatusOK, Header: make(http.Header), Body: io.NopCloser(strings.NewReader(`{"list":{"123":{"resolved_title":"Example","resolved_url":"https://example.test","excerpt":"Excerpt","time_updated":"2026-01-01"}}}`)), Request: request}, nil
	})

	connector, ok := New().Get("pocket")
	if !ok {
		t.Fatal("production registry missing pocket")
	}
	config := connectors.RuntimeConfig{Secrets: map[string]string{"consumer_key": "test-consumer", "access_token": "test-access"}}
	if err := connector.Check(context.Background(), config); err != nil {
		t.Fatalf("production Pocket Check() error = %v", err)
	}
	var records []connectors.Record
	if err := connector.Read(context.Background(), connectors.ReadRequest{Stream: "items", Config: config}, func(record connectors.Record) error {
		records = append(records, record)
		return nil
	}); err != nil {
		t.Fatalf("production Pocket Read() error = %v", err)
	}
	if len(records) != 1 || records[0]["item_id"] != "123" || records[0]["title"] != "Example" || records[0]["url"] != "https://example.test" {
		t.Fatalf("production Pocket records = %#v, want declared keyed-object projection", records)
	}
	if got := requests.Load(); got != 2 {
		t.Fatalf("production Pocket requests = %d, want bounded check plus items read", got)
	}
}

func TestProductionModeGenericCheckAndRead(t *testing.T) {
	previousTransport := http.DefaultTransport
	t.Cleanup(func() { http.DefaultTransport = previousTransport })

	var requests atomic.Int64
	http.DefaultTransport = roundTripFunc(func(request *http.Request) (*http.Response, error) {
		requests.Add(1)
		if request.URL.Host != "app.mode.com" || request.URL.Path != "/api/workspace-1/spaces" || request.Method != http.MethodGet {
			t.Fatalf("Mode request = %s %s, want fixed workspace spaces route", request.Method, request.URL)
		}
		if request.Header.Get("Accept") != "application/hal+json" || request.Header.Get("Authorization") == "" {
			t.Fatal("Mode request omitted declared HAL Accept or Basic authentication")
		}
		return &http.Response{StatusCode: http.StatusOK, Header: make(http.Header), Body: io.NopCloser(strings.NewReader(`{"_embedded":{"spaces":[{"token":"space-1","name":"Example","updated_at":"2026-01-01T00:00:00Z"}]},"_links":{"next":{"href":""}}}`)), Request: request}, nil
	})

	connector, ok := New().Get("mode")
	if !ok {
		t.Fatal("production registry missing mode")
	}
	config := connectors.RuntimeConfig{Config: map[string]string{"workspace": "workspace-1"}, Secrets: map[string]string{"api_token": "test-token", "api_secret": "test-secret"}}
	if err := connector.Check(context.Background(), config); err != nil {
		t.Fatalf("production Mode Check() error = %v", err)
	}
	var records []connectors.Record
	if err := connector.Read(context.Background(), connectors.ReadRequest{Stream: "spaces", Config: config}, func(record connectors.Record) error {
		records = append(records, record)
		return nil
	}); err != nil {
		t.Fatalf("production Mode Read() error = %v", err)
	}
	if len(records) != 1 || records[0]["token"] != "space-1" {
		t.Fatalf("production Mode records = %#v, want declared HAL projection", records)
	}
	if got := requests.Load(); got != 2 {
		t.Fatalf("production Mode requests = %d, want bounded check plus spaces read", got)
	}
}

func TestProductionMercadoAdsGenericCheckAndRead(t *testing.T) {
	previousTransport := http.DefaultTransport
	t.Cleanup(func() { http.DefaultTransport = previousTransport })

	var requests atomic.Int64
	http.DefaultTransport = roundTripFunc(func(request *http.Request) (*http.Response, error) {
		requests.Add(1)
		switch request.URL.Path {
		case "/oauth/token":
			if request.Method != http.MethodPost {
				t.Fatalf("Mercado token method = %s, want POST", request.Method)
			}
			return &http.Response{StatusCode: http.StatusOK, Header: make(http.Header), Body: io.NopCloser(strings.NewReader(`{"access_token":"test-token","expires_in":3600}`)), Request: request}, nil
		case "/advertising/advertisers":
			if request.URL.Host != "api.mercadolibre.com" || request.Method != http.MethodGet || request.Header.Get("Authorization") != "Bearer test-token" || request.Header.Get("Api-Version") != "1" || request.URL.Query().Get("product_id") != "BADS" {
				t.Fatalf("Mercado advertiser request = %s %s, want declared OAuth advertiser route", request.Method, request.URL)
			}
			return &http.Response{StatusCode: http.StatusOK, Header: make(http.Header), Body: io.NopCloser(strings.NewReader(`{"advertisers":[{"advertiser_id":"advertiser-1","advertiser_name":"Example","site_id":"MLB"}]}`)), Request: request}, nil
		default:
			t.Fatalf("Mercado request used undeclared route %s", request.URL)
			return nil, nil
		}
	})

	connector, ok := New().Get("mercado-ads")
	if !ok {
		t.Fatal("production registry missing mercado-ads")
	}
	config := connectors.RuntimeConfig{Config: map[string]string{"lookback_days": "30"}, Secrets: map[string]string{"client_id": "test-client", "client_secret": "test-secret", "client_refresh_token": "test-refresh"}}
	if err := connector.Check(context.Background(), config); err != nil {
		t.Fatalf("production Mercado Check() error = %v", err)
	}
	var records []connectors.Record
	if err := connector.Read(context.Background(), connectors.ReadRequest{Stream: "brand_advertisers", Config: config}, func(record connectors.Record) error {
		records = append(records, record)
		return nil
	}); err != nil {
		t.Fatalf("production Mercado Read() error = %v", err)
	}
	if len(records) != 1 || records[0]["advertiser_id"] != "advertiser-1" {
		t.Fatalf("production Mercado records = %#v, want declared advertiser projection", records)
	}
	if got := requests.Load(); got != 4 {
		t.Fatalf("production Mercado requests = %d, want token plus advertiser request for check and read", got)
	}
}

func TestProductionAshbyGenericCheckAndReadPreservesCommandSurface(t *testing.T) {
	previousTransport := http.DefaultTransport
	t.Cleanup(func() { http.DefaultTransport = previousTransport })

	var requests atomic.Int64
	http.DefaultTransport = roundTripFunc(func(request *http.Request) (*http.Response, error) {
		requests.Add(1)
		if request.URL.Host != "api.ashbyhq.com" || request.URL.Path != "/candidate.list" || request.Method != http.MethodPost {
			t.Fatalf("Ashby request = %s %s, want fixed candidate.list POST", request.Method, request.URL)
		}
		if request.Header.Get("Authorization") == "" || request.Header.Get("Accept") != "application/json; version=1" {
			t.Fatal("Ashby request omitted declared Basic authentication or versioned Accept header")
		}
		return &http.Response{StatusCode: http.StatusOK, Header: make(http.Header), Body: io.NopCloser(strings.NewReader(`{"results":[{"id":"candidate-1","name":"Example"}],"moreDataAvailable":false}`)), Request: request}, nil
	})

	bundle, err := engine.Load(defs.FS, "ashby")
	if err != nil {
		t.Fatalf("Load(ashby): %v", err)
	}
	if bundle.CLISurface == nil || len(bundle.CLISurface.Commands) != 178 {
		t.Fatalf("Ashby command surface = %#v, want all 178 declared commands", bundle.CLISurface)
	}
	connector, ok := New().Get("ashby")
	if !ok {
		t.Fatal("production registry missing ashby")
	}
	config := connectors.RuntimeConfig{Secrets: map[string]string{"api_key": "test-key"}}
	if err := connector.Check(context.Background(), config); err != nil {
		t.Fatalf("production Ashby Check() error = %v", err)
	}
	var records []connectors.Record
	if err := connector.Read(context.Background(), connectors.ReadRequest{Stream: "candidates", Config: config}, func(record connectors.Record) error {
		records = append(records, record)
		return nil
	}); err != nil {
		t.Fatalf("production Ashby Read() error = %v", err)
	}
	if len(records) != 1 || records[0]["id"] != "candidate-1" {
		t.Fatalf("production Ashby records = %#v, want declared candidate result", records)
	}
	if got := requests.Load(); got != 2 {
		t.Fatalf("production Ashby requests = %d, want bounded check plus candidates read", got)
	}
}

func TestProductionAshbyRejectsCallerOriginAndFixtureBeforeIO(t *testing.T) {
	previousTransport := http.DefaultTransport
	t.Cleanup(func() { http.DefaultTransport = previousTransport })

	var requests atomic.Int64
	http.DefaultTransport = roundTripFunc(func(*http.Request) (*http.Response, error) {
		requests.Add(1)
		return &http.Response{StatusCode: http.StatusOK, Header: make(http.Header), Body: io.NopCloser(strings.NewReader(`{"results":[]}`))}, nil
	})

	connector, ok := New().Get("ashby")
	if !ok {
		t.Fatal("production registry missing ashby")
	}
	for _, config := range []connectors.RuntimeConfig{
		{Config: map[string]string{"base_url": "https://hostile.example"}, Secrets: map[string]string{"api_key": "test-key"}},
		{Config: map[string]string{"mode": "fixture"}, Secrets: map[string]string{"api_key": "test-key"}},
	} {
		if err := connector.Check(context.Background(), config); err == nil {
			t.Fatal("Ashby accepted caller-controlled legacy configuration")
		}
	}
	if got := requests.Load(); got != 0 {
		t.Fatalf("Ashby hostile configuration requests = %d, want zero before credential or transport", got)
	}
}

func TestProductionPrestaShopTenantOriginRejectsLegacyConfigBeforeIO(t *testing.T) {
	previousTransport := http.DefaultTransport
	t.Cleanup(func() { http.DefaultTransport = previousTransport })
	var requests atomic.Int64
	http.DefaultTransport = roundTripFunc(func(request *http.Request) (*http.Response, error) {
		requests.Add(1)
		if request.URL.Host != "127.0.0.1:8443" || request.URL.Path != "/api/customers" {
			t.Fatalf("PrestaShop route = %s, want declared local tenant API route", request.URL)
		}
		return &http.Response{StatusCode: http.StatusOK, Header: make(http.Header), Body: io.NopCloser(strings.NewReader(`{"customers":[]}`)), Request: request}, nil
	})
	connector, ok := New().Get("prestashop")
	if !ok {
		t.Fatal("production registry missing prestashop")
	}
	valid := connectors.RuntimeConfig{Config: map[string]string{"url": "http://127.0.0.1:8443"}, Secrets: map[string]string{"access_key": "test-key"}}
	if err := connector.Check(context.Background(), valid); err != nil {
		t.Fatalf("production PrestaShop Check() error = %v", err)
	}
	for _, config := range []connectors.RuntimeConfig{
		{Config: map[string]string{"base_url": "https://example.invalid"}, Secrets: valid.Secrets},
		{Config: map[string]string{"url": "https://example.invalid", "mode": "fixture"}, Secrets: valid.Secrets},
	} {
		if err := connector.Check(context.Background(), config); err == nil {
			t.Fatal("production PrestaShop accepted legacy placeholder configuration")
		}
	}
	if got := requests.Load(); got != 1 {
		t.Fatalf("PrestaShop requests = %d, want one local tenant check only", got)
	}
}

func TestProductionPrestaShopGenericCheckAndRead(t *testing.T) {
	previousTransport := http.DefaultTransport
	t.Cleanup(func() { http.DefaultTransport = previousTransport })

	var requests atomic.Int64
	http.DefaultTransport = roundTripFunc(func(request *http.Request) (*http.Response, error) {
		requests.Add(1)
		if request.URL.Host != "127.0.0.1:8443" || request.URL.Path != "/api/customers" || request.Method != http.MethodGet {
			t.Fatalf("PrestaShop request = %s %s, want declared customers route", request.Method, request.URL)
		}
		username, password, ok := request.BasicAuth()
		if !ok || username != "test-key" || password != "" {
			t.Fatal("PrestaShop request omitted declared Basic access-key authentication")
		}
		if request.URL.Query().Get("output_format") != "JSON" || request.URL.Query().Get("display") != "full" {
			t.Fatalf("PrestaShop query = %s, want declared JSON full-resource query", request.URL.RawQuery)
		}
		switch request.URL.Query().Get("limit") {
		case "1":
			return &http.Response{StatusCode: http.StatusOK, Header: make(http.Header), Body: io.NopCloser(strings.NewReader(`{"customers":{"customer":[]}}`)), Request: request}, nil
		case "100":
			if request.URL.Query().Get("offset") != "0" {
				t.Fatalf("PrestaShop offset = %q, want first declared window", request.URL.Query().Get("offset"))
			}
			return &http.Response{StatusCode: http.StatusOK, Header: make(http.Header), Body: io.NopCloser(strings.NewReader(`{"customers":{"customer":[{"id":1,"firstname":"Ada","date_upd":"2026-01-02T00:00:00Z"}]}}`)), Request: request}, nil
		default:
			t.Fatalf("PrestaShop limit = %q, want declared check or first read window", request.URL.Query().Get("limit"))
			return nil, nil
		}
	})

	connector, ok := New().Get("prestashop")
	if !ok {
		t.Fatal("production registry missing prestashop")
	}
	config := connectors.RuntimeConfig{Config: map[string]string{"url": "http://127.0.0.1:8443", "start_date": "2026-01-01T00:00:00Z"}, Secrets: map[string]string{"access_key": "test-key"}}
	if err := connector.Check(context.Background(), config); err != nil {
		t.Fatalf("production PrestaShop Check() error = %v", err)
	}
	var records []connectors.Record
	if err := connector.Read(context.Background(), connectors.ReadRequest{Stream: "customers", Config: config}, func(record connectors.Record) error {
		records = append(records, record)
		return nil
	}); err != nil {
		t.Fatalf("production PrestaShop Read() error = %v", err)
	}
	if len(records) != 1 || fmt.Sprint(records[0]["id"]) != "1" || records[0]["firstname"] != "Ada" {
		t.Fatalf("production PrestaShop records = %#v, want declared customer record", records)
	}
	if got := requests.Load(); got != 2 {
		t.Fatalf("PrestaShop requests = %d, want one declared check and read", got)
	}
}

func TestProductionMetabaseDeclaredSessionCheckAndRead(t *testing.T) {
	previousTransport := http.DefaultTransport
	t.Cleanup(func() { http.DefaultTransport = previousTransport })

	var requests atomic.Int64
	http.DefaultTransport = roundTripFunc(func(request *http.Request) (*http.Response, error) {
		requests.Add(1)
		if request.URL.Host != "127.0.0.1:8080" {
			t.Fatalf("Metabase request = %s %s, want declared loopback tenant origin", request.Method, request.URL)
		}
		switch request.URL.Path {
		case "/api/session":
			if request.Method != http.MethodPost {
				t.Fatalf("Metabase session method = %s, want POST", request.Method)
			}
			var body map[string]string
			if err := json.NewDecoder(request.Body).Decode(&body); err != nil {
				t.Fatalf("decode Metabase session body: %v", err)
			}
			if body["username"] != "person@example.test" || body["password"] != "test-password" {
				t.Fatal("Metabase session body omitted the declared username/password binding")
			}
			return &http.Response{StatusCode: http.StatusOK, Header: make(http.Header), Body: io.NopCloser(strings.NewReader(`{"id":"issued-session"}`)), Request: request}, nil
		case "/api/card":
			session := request.Header.Get("X-Metabase-Session")
			if request.Method != http.MethodGet || (session != "issued-session" && session != "existing-session") {
				t.Fatal("Metabase card request omitted the exchanged or existing declared session header")
			}
			if request.Header.Get("password") != "" {
				t.Fatal("Metabase password leaked beyond the fixed session request")
			}
			return &http.Response{StatusCode: http.StatusOK, Header: make(http.Header), Body: io.NopCloser(strings.NewReader(`[{"id":1,"name":"Example"}]`)), Request: request}, nil
		default:
			t.Fatalf("Metabase route = %s, want declared session or card route", request.URL)
			return nil, nil
		}
	})

	config := connectors.RuntimeConfig{
		Config:  map[string]string{"instance_api_url": "http://127.0.0.1:8080", "username": "person@example.test"},
		Secrets: map[string]string{"password": "test-password"},
	}
	connector, ok := New().Get("metabase")
	if !ok {
		t.Fatal("production registry missing metabase")
	}
	if _, ok := connector.(*engine.Connector); !ok {
		t.Fatalf("production Metabase connector = %T, want rendered engine connector", connector)
	}
	if err := connector.Check(context.Background(), config); err != nil {
		t.Fatalf("production Metabase Check() error = %v", err)
	}
	var records []connectors.Record
	if err := connector.Read(context.Background(), connectors.ReadRequest{Stream: "cards", Config: config}, func(record connectors.Record) error {
		records = append(records, record)
		return nil
	}); err != nil {
		t.Fatalf("production Metabase Read() error = %v", err)
	}
	if len(records) != 1 || fmt.Sprint(records[0]["id"]) != "1" || records[0]["name"] != "Example" {
		t.Fatalf("production Metabase cards = %#v, want declared card record", records)
	}
	existingSessionConfig := config
	existingSessionConfig.Secrets = map[string]string{"session_token": "existing-session"}
	if err := connector.Check(context.Background(), existingSessionConfig); err != nil {
		t.Fatalf("production Metabase existing-session Check() error = %v", err)
	}
	if got := requests.Load(); got != 5 {
		t.Fatalf("production Metabase requests = %d, want session plus data request for check and cards read, then one existing-session check", got)
	}
}

func TestProductionYahooFinancePriceArrayZipCheckAndRead(t *testing.T) {
	previousTransport := http.DefaultTransport
	t.Cleanup(func() { http.DefaultTransport = previousTransport })

	var requests atomic.Int64
	http.DefaultTransport = roundTripFunc(func(request *http.Request) (*http.Response, error) {
		requests.Add(1)
		if request.URL.Host != "query1.finance.yahoo.com" || request.URL.Path != "/v8/finance/chart/AAPL" || request.Method != http.MethodGet {
			t.Fatalf("Yahoo Finance request = %s %s, want declared chart route", request.Method, request.URL)
		}
		if request.URL.Query().Get("interval") != "1d" || request.URL.Query().Get("range") != "5d" {
			t.Fatalf("Yahoo Finance query = %s, want declared interval and range", request.URL.RawQuery)
		}
		return &http.Response{StatusCode: http.StatusOK, Header: make(http.Header), Body: io.NopCloser(strings.NewReader(`{"chart":{"result":[{"meta":{"symbol":"AAPL","currency":"USD"},"timestamp":[1,2],"indicators":{"quote":[{"open":[10,11],"high":[12,13],"low":[9,10],"close":[11,12],"volume":[100,200]}],"adjclose":[{"adjclose":[10.5,11.5]}]}}],"error":null}}`)), Request: request}, nil
	})

	connector, ok := New().Get("yahoo-finance-price")
	if !ok {
		t.Fatal("production registry missing yahoo-finance-price")
	}
	config := connectors.RuntimeConfig{Config: map[string]string{"symbol": "AAPL", "interval": "1d", "range": "5d"}}
	if err := connector.Check(context.Background(), config); err != nil {
		t.Fatalf("production Yahoo Finance Check() error = %v", err)
	}
	var records []connectors.Record
	if err := connector.Read(context.Background(), connectors.ReadRequest{Stream: "prices", Config: config}, func(record connectors.Record) error {
		records = append(records, record)
		return nil
	}); err != nil {
		t.Fatalf("production Yahoo Finance Read() error = %v", err)
	}
	if len(records) != 2 {
		t.Fatalf("Yahoo Finance records = %#v, want two zipped OHLCV rows", records)
	}
	if records[0]["symbol"] != "AAPL" || records[0]["currency"] != "USD" || fmt.Sprint(records[0]["timestamp"]) != "1" || fmt.Sprint(records[0]["open"]) != "10" || fmt.Sprint(records[0]["adjclose"]) != "10.5" {
		t.Fatalf("first Yahoo Finance record = %#v, want first declared OHLCV row", records[0])
	}
	if fmt.Sprint(records[1]["timestamp"]) != "2" || fmt.Sprint(records[1]["close"]) != "12" || fmt.Sprint(records[1]["volume"]) != "200" {
		t.Fatalf("second Yahoo Finance record = %#v, want second declared OHLCV row", records[1])
	}
	if got := requests.Load(); got != 2 {
		t.Fatalf("Yahoo Finance requests = %d, want one bounded check plus one read", got)
	}
}

func TestProductionYahooFinancePriceRejectsChartErrorBeforeEmit(t *testing.T) {
	previousTransport := http.DefaultTransport
	t.Cleanup(func() { http.DefaultTransport = previousTransport })

	http.DefaultTransport = roundTripFunc(func(request *http.Request) (*http.Response, error) {
		if request.URL.Host != "query1.finance.yahoo.com" || request.URL.Path != "/v8/finance/chart/AAPL" {
			t.Fatalf("Yahoo Finance error request = %s, want declared chart route", request.URL)
		}
		return &http.Response{StatusCode: http.StatusOK, Header: make(http.Header), Body: io.NopCloser(strings.NewReader(`{"chart":{"result":null,"error":{"description":"symbol unavailable"}}}`)), Request: request}, nil
	})

	connector, ok := New().Get("yahoo-finance-price")
	if !ok {
		t.Fatal("production registry missing yahoo-finance-price")
	}
	emitted := 0
	err := connector.Read(context.Background(), connectors.ReadRequest{Stream: "prices", Config: connectors.RuntimeConfig{Config: map[string]string{"symbol": "AAPL"}}}, func(connectors.Record) error {
		emitted++
		return nil
	})
	var responseErr *engine.DeclaredResponseError
	if !errors.As(err, &responseErr) {
		t.Fatalf("Yahoo Finance chart error = %T %v, want DeclaredResponseError", err, err)
	}
	if responseErr.Path != "chart.error" || responseErr.Message != "symbol unavailable" {
		t.Fatalf("Yahoo Finance response error = %#v, want declared chart error", responseErr)
	}
	if emitted != 0 {
		t.Fatalf("Yahoo Finance emitted %d records after chart error, want zero", emitted)
	}
}
