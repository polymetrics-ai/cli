package alphavantage

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"

	"polymetrics.ai/internal/connectors"
)

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}

func TestFixtureModeNoLongerBypassesCredentialBoundary(t *testing.T) {
	var sent atomic.Int32
	connector := Connector{Client: &http.Client{Transport: roundTripFunc(func(*http.Request) (*http.Response, error) {
		sent.Add(1)
		return nil, errors.New("unexpected request")
	})}}

	err := connector.Check(context.Background(), connectors.RuntimeConfig{
		Config: map[string]string{"mode": "fixture"},
	})
	if err == nil || !strings.Contains(err.Error(), "requires secret api_key") {
		t.Fatalf("Check(mode=fixture) error = %v, want ordinary missing api_key error", err)
	}
	if got := sent.Load(); got != 0 {
		t.Fatalf("Check(mode=fixture) sent %d requests, want none before credential validation", got)
	}
}

func TestAlphaVantageRejectsUntrustedBaseURLBeforeSecretSend(t *testing.T) {
	var requests atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
		requests.Add(1)
	}))
	defer server.Close()

	err := Connector{}.Check(context.Background(), connectors.RuntimeConfig{
		Config:  map[string]string{"base_url": server.URL},
		Secrets: map[string]string{"api_key": "test-only"},
	})
	if got := requests.Load(); got != 0 {
		t.Fatalf("Check() sent %d requests to an untrusted origin, want none", got)
	}
	if err == nil || !strings.Contains(err.Error(), "base_url") {
		t.Fatalf("Check() error = %v, want untrusted base_url rejection", err)
	}
}
