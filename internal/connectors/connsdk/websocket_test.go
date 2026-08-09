package connsdk

import (
	"context"
	"crypto/sha1"
	"encoding/base64"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestOpenWebSocketUsesDeclaredUpgradeHandshakeAndAuthenticatedConnection(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, request *http.Request) {
		if request.Method != http.MethodGet || request.URL.Path != "/live" {
			t.Errorf("upgrade request = %s %s, want GET /live", request.Method, request.URL.Path)
		}
		if !strings.EqualFold(request.Header.Get("Upgrade"), "websocket") {
			t.Errorf("Upgrade = %q, want websocket", request.Header.Get("Upgrade"))
		}
		if !headerToken(request.Header.Get("Connection"), "upgrade") {
			t.Errorf("Connection = %q, want upgrade token", request.Header.Get("Connection"))
		}
		if got := request.Header.Get("Sec-WebSocket-Protocol"); got != "fixture-live" {
			t.Errorf("Sec-WebSocket-Protocol = %q, want fixture-live", got)
		}
		if got := request.Header.Get("X-Connector-Auth"); got != "present" {
			t.Errorf("connector auth marker = %q, want present", got)
		}

		connection, buffered, err := http.NewResponseController(w).Hijack()
		if err != nil {
			t.Errorf("Hijack: %v", err)
			return
		}
		defer connection.Close()

		key := request.Header.Get("Sec-WebSocket-Key")
		if key == "" {
			t.Error("Sec-WebSocket-Key is empty")
			return
		}
		sum := sha1.Sum([]byte(key + websocketAcceptGUID))
		if _, err := fmt.Fprintf(buffered, "HTTP/1.1 101 Switching Protocols\r\nConnection: Upgrade\r\nUpgrade: websocket\r\nSec-WebSocket-Accept: %s\r\nSec-WebSocket-Protocol: fixture-live\r\n\r\n", base64.StdEncoding.EncodeToString(sum[:])); err != nil {
			t.Errorf("write handshake: %v", err)
			return
		}
		if err := buffered.Flush(); err != nil {
			t.Errorf("flush handshake: %v", err)
			return
		}

		incoming := make([]byte, len("client-ping"))
		if _, err := io.ReadFull(buffered, incoming); err != nil {
			t.Errorf("read upgraded bytes: %v", err)
			return
		}
		if got := string(incoming); got != "client-ping" {
			t.Errorf("upgraded bytes = %q, want client-ping", got)
		}
		if _, err := buffered.WriteString("server-pong"); err != nil {
			t.Errorf("write upgraded bytes: %v", err)
			return
		}
		if err := buffered.Flush(); err != nil {
			t.Errorf("flush upgraded bytes: %v", err)
		}
	}))
	defer srv.Close()

	requester := &Requester{
		BaseURL: srv.URL,
		Auth: AuthFunc(func(_ context.Context, request *http.Request) error {
			request.Header.Set("X-Connector-Auth", "present")
			return nil
		}),
	}
	upgrade, err := requester.OpenWebSocket(context.Background(), "/live", "fixture-live")
	if err != nil {
		t.Fatalf("OpenWebSocket: %v", err)
	}
	defer upgrade.Conn.Close()

	if _, err := io.WriteString(upgrade.Conn, "client-ping"); err != nil {
		t.Fatalf("write upgraded connection: %v", err)
	}
	got := make([]byte, len("server-pong"))
	if _, err := io.ReadFull(upgrade.Conn, got); err != nil {
		t.Fatalf("read upgraded connection: %v", err)
	}
	if string(got) != "server-pong" {
		t.Fatalf("upgraded response = %q, want server-pong", got)
	}
}

func TestOpenWebSocketRejectsNonUpgradeAndInvalidSubprotocol(t *testing.T) {
	tests := []struct {
		name    string
		handler http.HandlerFunc
		want    string
	}{
		{
			name: "ordinary response",
			handler: func(w http.ResponseWriter, _ *http.Request) {
				w.WriteHeader(http.StatusOK)
			},
			want: "expected HTTP 101",
		},
		{
			name: "redirect response",
			handler: func(w http.ResponseWriter, request *http.Request) {
				http.Redirect(w, request, "/other", http.StatusTemporaryRedirect)
			},
			want: "expected HTTP 101",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			srv := httptest.NewServer(tt.handler)
			defer srv.Close()

			_, err := (&Requester{BaseURL: srv.URL}).OpenWebSocket(context.Background(), "/live", "fixture-live")
			if err == nil || !strings.Contains(err.Error(), tt.want) {
				t.Fatalf("OpenWebSocket error = %v, want %q", err, tt.want)
			}
		})
	}
}

func headerToken(value, want string) bool {
	for _, candidate := range strings.Split(value, ",") {
		if strings.EqualFold(strings.TrimSpace(candidate), want) {
			return true
		}
	}
	return false
}
