package certify_test

import (
	"bytes"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"polymetrics.ai/internal/connectors/certify"
)

func TestObservedTransportCapturesCompleteHTTPSExchange(t *testing.T) {
	const canary = "cert-canary-authorization-query-json-response"
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got := r.Header.Get("Authorization"); got != "Bearer "+canary {
			t.Fatalf("authorization = %q, want canary-bearing value", got)
		}
		if got := r.URL.Query().Get("access_token"); got != canary {
			t.Fatalf("access_token query = %q, want canary", got)
		}
		body, err := io.ReadAll(r.Body)
		if err != nil {
			t.Fatalf("read request body: %v", err)
		}
		if !bytes.Contains(body, []byte(canary)) {
			t.Fatalf("request body %q does not carry canary", body)
		}
		w.Header().Set("X-Canary", canary)
		_, _ = w.Write([]byte(`{"credential":"` + canary + `"}`))
	}))
	defer server.Close()

	observer := certify.NewObservedTransport(server.Client().Transport, 4096)
	client := &http.Client{Transport: observer}
	req, err := http.NewRequest(http.MethodPost, server.URL+"/proof?access_token="+canary, bytes.NewBufferString(`{"token":"`+canary+`"}`))
	if err != nil {
		t.Fatalf("new request: %v", err)
	}
	req.Header.Set("Authorization", "Bearer "+canary)
	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("do observed request: %v", err)
	}
	if _, err := io.ReadAll(resp.Body); err != nil {
		t.Fatalf("read response: %v", err)
	}
	if err := resp.Body.Close(); err != nil {
		t.Fatalf("close response: %v", err)
	}

	exchanges := observer.Exchanges()
	if len(exchanges) != 1 {
		t.Fatalf("observed exchange count = %d, want 1", len(exchanges))
	}
	exchange := exchanges[0]
	if exchange.Request.Method != http.MethodPost || exchange.Request.Target != server.URL+"/proof?access_token="+canary {
		t.Fatalf("observed request = %#v, want method and exact target", exchange.Request)
	}
	if got := exchange.Request.Headers.Get("Authorization"); got != "Bearer "+canary {
		t.Fatalf("observed authorization = %q, want exact canary-bearing value", got)
	}
	if !bytes.Contains(exchange.Request.Body.Bytes, []byte(canary)) || !exchange.Request.Body.Complete || exchange.Request.Body.Truncated {
		t.Fatalf("observed request body = %#v, want complete untruncated canary-bearing body", exchange.Request.Body)
	}
	if exchange.Response.Status != http.StatusOK || exchange.Response.Headers.Get("X-Canary") != canary {
		t.Fatalf("observed response = %#v, want status/header", exchange.Response)
	}
	if !bytes.Contains(exchange.Response.Body.Bytes, []byte(canary)) || !exchange.Response.Body.Complete || exchange.Response.Body.Truncated {
		t.Fatalf("observed response body = %#v, want complete untruncated canary-bearing body", exchange.Response.Body)
	}
}

func TestObservedTransportBoundsResponseBodyWithoutChangingChildBytes(t *testing.T) {
	body := bytes.Repeat([]byte("x"), 65)
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusBadGateway)
		_, _ = w.Write(body)
	}))
	defer server.Close()

	observer := certify.NewObservedTransport(server.Client().Transport, 32)
	resp, err := (&http.Client{Transport: observer}).Get(server.URL + "/binary")
	if err != nil {
		t.Fatalf("get binary response: %v", err)
	}
	gotBody, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("read child-visible response: %v", err)
	}
	if err := resp.Body.Close(); err != nil {
		t.Fatalf("close response: %v", err)
	}
	if !bytes.Equal(gotBody, body) {
		t.Fatalf("child-visible body = %d bytes, want original %d bytes", len(gotBody), len(body))
	}

	exchanges := observer.Exchanges()
	if len(exchanges) != 1 {
		t.Fatalf("observed exchange count = %d, want 1", len(exchanges))
	}
	captured := exchanges[0].Response.Body
	if !captured.Complete || !captured.Truncated || captured.OriginalBytes != len(body) || len(captured.Bytes) != 32 {
		t.Fatalf("bounded captured body = %#v, want complete 65-byte source truncated to 32 bytes", captured)
	}
}
