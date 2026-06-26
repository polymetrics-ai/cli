package youneedabudgetynab_test

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"polymetrics.ai/internal/connectors"
	ynab "polymetrics.ai/internal/connectors/you-need-a-budget-ynab"
)

func TestReadBudgetsAuthenticatesAndMaps(t *testing.T) {
	var sawAuth string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sawAuth = r.Header.Get("Authorization")
		if r.URL.Path != "/v1/budgets" {
			http.NotFound(w, r)
			return
		}
		_, _ = w.Write([]byte(`{"data":{"budgets":[{"id":"budget_1","name":"Household","last_modified_on":"2026-01-01T00:00:00Z"}]}}`))
	}))
	defer srv.Close()

	c := ynab.New()
	cfg := connectors.RuntimeConfig{Config: map[string]string{"base_url": srv.URL + "/v1"}, Secrets: map[string]string{"api_key": "ynab_key"}}
	var got []connectors.Record
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "budgets", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	})
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	if sawAuth != "Bearer ynab_key" {
		t.Fatalf("Authorization = %q, want bearer token", sawAuth)
	}
	if len(got) != 1 || got[0]["id"] != "budget_1" || got[0]["name"] != "Household" {
		t.Fatalf("records = %+v, want mapped budget", got)
	}
}

func TestFixtureCatalogRegistrationAndWrite(t *testing.T) {
	c := ynab.New()
	fixture := connectors.RuntimeConfig{Config: map[string]string{"mode": "fixture"}}
	if err := c.Check(context.Background(), fixture); err != nil {
		t.Fatalf("Check fixture: %v", err)
	}
	var rows []connectors.Record
	if err := c.Read(context.Background(), connectors.ReadRequest{Stream: "accounts", Config: fixture}, func(rec connectors.Record) error {
		rows = append(rows, rec)
		return nil
	}); err != nil {
		t.Fatalf("Read fixture: %v", err)
	}
	if len(rows) == 0 || rows[0]["id"] == nil {
		t.Fatalf("fixture rows = %+v, want records with id", rows)
	}
	cat, err := c.Catalog(context.Background(), connectors.RuntimeConfig{})
	if err != nil {
		t.Fatalf("Catalog: %v", err)
	}
	if cat.Connector != "you-need-a-budget-ynab" || len(cat.Streams) < 3 {
		t.Fatalf("catalog = %+v, want ynab streams", cat)
	}
	if _, ok := connectors.NewRegistry().Get("you-need-a-budget-ynab"); !ok {
		t.Fatal("registry did not resolve you-need-a-budget-ynab")
	}
	if _, err := c.Write(context.Background(), connectors.WriteRequest{}, nil); !errors.Is(err, connectors.ErrUnsupportedOperation) {
		t.Fatalf("Write error = %v, want ErrUnsupportedOperation", err)
	}
}
