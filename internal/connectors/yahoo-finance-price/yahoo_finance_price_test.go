package yahoofinanceprice_test

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"polymetrics.ai/internal/connectors"
	yahoofinanceprice "polymetrics.ai/internal/connectors/yahoo-finance-price"
)

func TestFixtureCatalogRegistrationAndWrite(t *testing.T) {
	c := yahoofinanceprice.New()
	fixture := connectors.RuntimeConfig{Config: map[string]string{"mode": "fixture"}}
	if err := c.Check(context.Background(), fixture); err != nil {
		t.Fatalf("Check fixture: %v", err)
	}
	cat, err := c.Catalog(context.Background(), connectors.RuntimeConfig{})
	if err != nil {
		t.Fatalf("Catalog: %v", err)
	}
	if cat.Connector != "yahoo-finance-price" || len(cat.Streams) != 1 {
		t.Fatalf("catalog = %+v, want yahoo-finance-price stream", cat)
	}
	var rows []connectors.Record
	err = c.Read(context.Background(), connectors.ReadRequest{Stream: "prices", Config: fixture}, func(rec connectors.Record) error {
		rows = append(rows, rec)
		return nil
	})
	if err != nil {
		t.Fatalf("Read fixture: %v", err)
	}
	if len(rows) == 0 || rows[0]["fixture"] != true || rows[0]["symbol"] == nil {
		t.Fatalf("fixture rows = %+v, want fixture price records", rows)
	}
	if _, ok := connectors.NewRegistry().Get("yahoo-finance-price"); !ok {
		t.Fatal("registry did not resolve yahoo-finance-price")
	}
	if _, err := c.Write(context.Background(), connectors.WriteRequest{}, nil); !errors.Is(err, connectors.ErrUnsupportedOperation) {
		t.Fatalf("Write error = %v, want ErrUnsupportedOperation", err)
	}
}

func TestReadLiveFlattensChartPrices(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v8/finance/chart/MSFT" {
			http.NotFound(w, r)
			return
		}
		_, _ = w.Write([]byte(`{"chart":{"result":[{"meta":{"symbol":"MSFT","currency":"USD"},"timestamp":[1767225600,1767312000],"indicators":{"quote":[{"open":[1.1,2.2],"high":[1.2,2.3],"low":[1.0,2.1],"close":[1.15,2.25],"volume":[10,20]}],"adjclose":[{"adjclose":[1.14,2.24]}]}}],"error":null}}`))
	}))
	defer srv.Close()

	c := yahoofinanceprice.New()
	cfg := connectors.RuntimeConfig{Config: map[string]string{"base_url": srv.URL, "symbol": "MSFT", "range": "2d", "interval": "1d"}}
	var got []connectors.Record
	if err := c.Read(context.Background(), connectors.ReadRequest{Stream: "prices", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	}); err != nil {
		t.Fatalf("Read: %v", err)
	}
	if len(got) != 2 || got[0]["symbol"] != "MSFT" || got[1]["close"] != 2.25 {
		t.Fatalf("records = %+v, want flattened MSFT prices", got)
	}
}
