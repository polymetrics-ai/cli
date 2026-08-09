package engine

import (
	"bufio"
	"bytes"
	"context"
	"crypto/sha1"
	"encoding/base64"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestOperationWebSocketSessionStreamsBoundedMaskedPCM16AndRedactsEvents(t *testing.T) {
	audio := bytes.Repeat([]byte{0x01, 0x80}, 256)
	srv := websocketSessionFixtureServer(t, func(t *testing.T, connection io.ReadWriteCloser, reader *bufio.ReadWriter) {
		opcode, masked, payload := websocketFixtureReadFrame(t, reader)
		if opcode != websocketOpcodeText || !masked {
			t.Errorf("session.update frame opcode/masked = %d/%t, want text/masked", opcode, masked)
		}
		var update map[string]any
		if err := json.Unmarshal(payload, &update); err != nil {
			t.Errorf("decode session.update: %v", err)
		} else if update["type"] != "session.update" || update["input_audio_format"] != "pcm16" {
			t.Errorf("session.update = %#v, want declared PCM16 update", update)
		}

		var gotAudio []byte
		for range 2 {
			opcode, masked, payload = websocketFixtureReadFrame(t, reader)
			if opcode != websocketOpcodeBinary || !masked {
				t.Errorf("audio frame opcode/masked = %d/%t, want binary/masked", opcode, masked)
			}
			if len(payload) > 256 {
				t.Errorf("audio frame length = %d, exceeds declared 256", len(payload))
			}
			gotAudio = append(gotAudio, payload...)
		}
		if !bytes.Equal(gotAudio, audio) {
			t.Error("PCM16 frames did not preserve the bounded input")
		}
		opcode, masked, _ = websocketFixtureReadFrame(t, reader)
		if opcode != websocketOpcodeClose || !masked {
			t.Errorf("terminal client frame opcode/masked = %d/%t, want close/masked", opcode, masked)
		}

		websocketFixtureWriteFrame(t, reader, websocketOpcodePing, []byte("health"))
		opcode, masked, payload = websocketFixtureReadFrame(t, reader)
		if opcode != websocketOpcodePong || !masked || string(payload) != "health" {
			t.Errorf("pong opcode/masked/payload = %d/%t/%q", opcode, masked, payload)
		}
		websocketFixtureWriteFrame(t, reader, websocketOpcodeText, []byte(`{"event":"transcript","access_token":"not-exported"}`))
		websocketFixtureWriteFrame(t, reader, websocketOpcodeClose, []byte{0x03, 0xe8})
		_ = connection.Close()
	})
	defer srv.Close()

	result, err := OperationWebSocketSession(context.Background(), websocketSessionFixtureBundle(srv.URL, 256), WebSocketSessionRequest{
		Operation:     "acme.live",
		SessionUpdate: map[string]any{"type": "session.update", "input_audio_format": "pcm16"},
		PCM16:         audio,
	}, nil)
	if err != nil {
		t.Fatalf("OperationWebSocketSession: %v", err)
	}
	if result.Status != http.StatusSwitchingProtocols || result.BytesSent != len(audio) || result.BytesReceived == 0 {
		t.Fatalf("session result = %#v, want upgraded bounded accounting", result)
	}
	encoded, err := json.Marshal(result.Events)
	if err != nil {
		t.Fatalf("marshal redacted events: %v", err)
	}
	if bytes.Contains(encoded, []byte("not-exported")) {
		t.Fatal("redacted websocket event leaked a protected field")
	}
	if !bytes.Contains(encoded, []byte("access_token_redacted")) {
		t.Fatalf("redacted websocket event lacks marker: %s", encoded)
	}
}

func TestOperationWebSocketSessionRejectsOversizeServerFrame(t *testing.T) {
	srv := websocketSessionFixtureServer(t, func(t *testing.T, _ io.ReadWriteCloser, reader *bufio.ReadWriter) {
		for range 3 {
			websocketFixtureReadFrame(t, reader)
		}
		websocketFixtureWriteFrame(t, reader, websocketOpcodeText, bytes.Repeat([]byte{'x'}, 257))
	})
	defer srv.Close()

	_, err := OperationWebSocketSession(context.Background(), websocketSessionFixtureBundle(srv.URL, 256), WebSocketSessionRequest{
		Operation:     "acme.live",
		SessionUpdate: map[string]any{"type": "session.update", "input_audio_format": "pcm16"},
		PCM16:         []byte{0x01, 0x80},
	}, nil)
	if err == nil || !strings.Contains(err.Error(), "max_frame_bytes") {
		t.Fatalf("oversize server frame error = %v, want max_frame_bytes refusal", err)
	}
}

func TestOperationWebSocketSessionExpiresDeclarationOwnedLifetime(t *testing.T) {
	serverReceivedClientClose := make(chan struct{})
	srv := websocketSessionFixtureServer(t, func(t *testing.T, connection io.ReadWriteCloser, reader *bufio.ReadWriter) {
		for range 3 {
			websocketFixtureReadFrame(t, reader)
		}
		close(serverReceivedClientClose)
		// The fixture deliberately withholds its terminal close. The engine must
		// enforce the declaration's wall-clock bound without a caller deadline.
		_, _ = io.Copy(io.Discard, reader)
		_ = connection.Close()
	})
	defer srv.Close()

	bundle := websocketSessionFixtureBundle(srv.URL, 256)
	bundle.Operations[0].WebSocket.MaxSessionSeconds = 1
	started := time.Now()
	_, err := OperationWebSocketSession(context.Background(), bundle, WebSocketSessionRequest{
		Operation:     "acme.live",
		SessionUpdate: map[string]any{"type": "session.update", "input_audio_format": "pcm16"},
		PCM16:         []byte{0x01, 0x80},
	}, nil)
	if err == nil || !strings.Contains(err.Error(), context.DeadlineExceeded.Error()) {
		t.Fatalf("lifetime-bounded session error = %v, want context deadline exceeded", err)
	}
	if elapsed := time.Since(started); elapsed > 5*time.Second {
		t.Fatalf("lifetime-bounded session took %s, want prompt deadline", elapsed)
	}
	select {
	case <-serverReceivedClientClose:
	case <-time.After(2 * time.Second):
		t.Fatal("fixture never received the closed client session")
	}
}

func websocketSessionFixtureBundle(baseURL string, maxFrameBytes int) Bundle {
	return Bundle{
		Name: "acme",
		HTTP: HTTPBase{URL: baseURL},
		Operations: []OperationSpec{{
			ID:           "acme.live",
			Kind:         "websocket_session",
			Summary:      "Stream one fixture session",
			Risk:         "medium",
			Approval:     "none",
			OutputPolicy: "json_redacted",
			WebSocket: &WebSocketSessionSpec{
				Method:            http.MethodGet,
				Path:              "/live",
				Subprotocol:       "fixture-live",
				MaxInputBytes:     1024,
				MaxOutputBytes:    1024,
				MaxFrameBytes:     maxFrameBytes,
				MaxSessionSeconds: 60,
				SessionUpdateSchema: json.RawMessage(`{
					"type":"object",
					"additionalProperties":false,
					"required":["type","input_audio_format"],
					"properties":{
						"type":{"type":"string","enum":["session.update"]},
						"input_audio_format":{"type":"string","enum":["pcm16"]}
					}
				}`),
			},
		}},
		Surface: &APISurface{Endpoints: []SurfaceEndpoint{{
			Method: http.MethodGet,
			Path:   "/live",
			CoveredBy: &SurfaceCoverage{
				WebSocketSession: "acme.live",
			},
		}}},
	}
}

func websocketSessionFixtureServer(t *testing.T, run func(*testing.T, io.ReadWriteCloser, *bufio.ReadWriter)) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, request *http.Request) {
		if request.Method != http.MethodGet || request.URL.Path != "/live" {
			t.Errorf("websocket request = %s %s, want GET /live", request.Method, request.URL.Path)
		}
		connection, buffered, err := http.NewResponseController(w).Hijack()
		if err != nil {
			t.Errorf("Hijack: %v", err)
			return
		}
		defer func() {
			_ = connection.Close()
		}()

		key := request.Header.Get("Sec-WebSocket-Key")
		sum := sha1.Sum([]byte(key + "258EAFA5-E914-47DA-95CA-C5AB0DC85B11"))
		if _, err := fmt.Fprintf(buffered, "HTTP/1.1 101 Switching Protocols\r\nConnection: Upgrade\r\nUpgrade: websocket\r\nSec-WebSocket-Accept: %s\r\nSec-WebSocket-Protocol: fixture-live\r\n\r\n", base64.StdEncoding.EncodeToString(sum[:])); err != nil {
			t.Errorf("write handshake: %v", err)
			return
		}
		if err := buffered.Flush(); err != nil {
			t.Errorf("flush handshake: %v", err)
			return
		}
		run(t, connection, buffered)
	}))
}

const (
	websocketOpcodeText   = 0x1
	websocketOpcodeBinary = 0x2
	websocketOpcodeClose  = 0x8
	websocketOpcodePing   = 0x9
	websocketOpcodePong   = 0xa
)

func websocketFixtureReadFrame(t *testing.T, reader *bufio.ReadWriter) (opcode byte, masked bool, payload []byte) {
	t.Helper()
	first, err := reader.ReadByte()
	if err != nil {
		t.Fatalf("read frame first byte: %v", err)
	}
	second, err := reader.ReadByte()
	if err != nil {
		t.Fatalf("read frame second byte: %v", err)
	}
	opcode = first & 0x0f
	masked = second&0x80 != 0
	length := int(second & 0x7f)
	if length == 126 {
		var encoded uint16
		if err := binary.Read(reader, binary.BigEndian, &encoded); err != nil {
			t.Fatalf("read extended frame length: %v", err)
		}
		length = int(encoded)
	}
	if length > 4096 {
		t.Fatalf("fixture received oversized frame length %d", length)
	}
	var mask [4]byte
	if masked {
		if _, err := io.ReadFull(reader, mask[:]); err != nil {
			t.Fatalf("read frame mask: %v", err)
		}
	}
	payload = make([]byte, length)
	if _, err := io.ReadFull(reader, payload); err != nil {
		t.Fatalf("read frame payload: %v", err)
	}
	if masked {
		for i := range payload {
			payload[i] ^= mask[i%len(mask)]
		}
	}
	return opcode, masked, payload
}

func websocketFixtureWriteFrame(t *testing.T, writer *bufio.ReadWriter, opcode byte, payload []byte) {
	t.Helper()
	if len(payload) > 65535 {
		t.Fatalf("fixture payload too large: %d", len(payload))
	}
	if err := writer.WriteByte(0x80 | opcode); err != nil {
		t.Fatalf("write frame first byte: %v", err)
	}
	if len(payload) < 126 {
		if err := writer.WriteByte(byte(len(payload))); err != nil {
			t.Fatalf("write frame length: %v", err)
		}
	} else {
		if err := writer.WriteByte(126); err != nil {
			t.Fatalf("write frame extended marker: %v", err)
		}
		if err := binary.Write(writer, binary.BigEndian, uint16(len(payload))); err != nil {
			t.Fatalf("write frame extended length: %v", err)
		}
	}
	if _, err := writer.Write(payload); err != nil {
		t.Fatalf("write frame payload: %v", err)
	}
	if err := writer.Flush(); err != nil {
		t.Fatalf("flush frame: %v", err)
	}
}
