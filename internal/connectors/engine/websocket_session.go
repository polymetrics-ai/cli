package engine

import (
	"bytes"
	"context"
	cryptorand "crypto/rand"
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math"
	"net/http"
	"strings"
	"time"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/connsdk"
	"polymetrics.ai/internal/safety"
)

const (
	websocketFrameOpcodeText   = 0x1
	websocketFrameOpcodeBinary = 0x2
	websocketFrameOpcodeClose  = 0x8
	websocketFrameOpcodePing   = 0x9
	websocketFrameOpcodePong   = 0xa

	// maxWebSocketSessionSeconds is an absolute ceiling for one closed
	// declaration-owned session. It prevents a malformed bundle from using a
	// nominally finite integer as an effectively unbounded terminal command.
	maxWebSocketSessionSeconds = 60 * 60
)

// WebSocketSessionRequest is the closed runtime input for one declared live
// PCM16 transcription session. PCM16 is intentionally bytes rather than an
// io.Reader: callers cannot use this API as an unbounded raw-frame tunnel, and
// the declared operation cap is checked before any connection is opened.
type WebSocketSessionRequest struct {
	Operation     string
	Config        connectors.RuntimeConfig
	SessionUpdate map[string]any
	PCM16         []byte
}

// WebSocketSessionResult carries only bounded, redacted JSON events. It never
// exposes a socket, raw provider headers, or transport controls to a command.
type WebSocketSessionResult struct {
	Connector     string `json:"connector"`
	Operation     string `json:"operation"`
	Method        string `json:"method"`
	Path          string `json:"path"`
	Status        int    `json:"status"`
	BytesSent     int    `json:"bytes_sent"`
	BytesReceived int    `json:"bytes_received"`
	Events        []any  `json:"events"`
}

// OperationWebSocketSession executes one fixed, declaration-owned WebSocket
// session. It sends exactly the schema-bound configuration text frame followed
// by finite masked PCM16 binary frames, then a close. It accepts only bounded
// JSON text events in response and redacts them before returning.
func OperationWebSocketSession(ctx context.Context, b Bundle, req WebSocketSessionRequest, h Hooks) (WebSocketSessionResult, error) {
	if err := ctx.Err(); err != nil {
		return WebSocketSessionResult{}, err
	}
	op, err := operationWebSocketSessionSpec(b, req.Operation)
	if err != nil {
		return WebSocketSessionResult{}, err
	}
	spec := op.WebSocket
	update, err := websocketSessionUpdatePayload(op, req.SessionUpdate)
	if err != nil {
		return WebSocketSessionResult{}, err
	}
	if err := validateWebSocketSessionPCM16(update, req.PCM16, spec); err != nil {
		return WebSocketSessionResult{}, err
	}
	sessionCtx, cancelSession := context.WithTimeout(ctx, time.Duration(spec.MaxSessionSeconds)*time.Second)
	defer cancelSession()

	cfg := materializeConfigDefaults(b, req.Config)
	baseURL, err := Interpolate(b.HTTP.URL, requestVars(cfg, nil, ""))
	if err != nil {
		return WebSocketSessionResult{}, fmt.Errorf("websocket session resolve base URL: %w", err)
	}
	if strings.TrimSpace(baseURL) == "" {
		return WebSocketSessionResult{}, fmt.Errorf("websocket session base URL is required")
	}
	requestPath := normalizeDirectReadPathForBaseURL(spec.Path, baseURL)
	rt, err := newRuntime(sessionCtx, b, cfg, h)
	if err != nil {
		return WebSocketSessionResult{}, err
	}
	requester, err := rt.RequesterFor(http.MethodGet, spec.Path)
	if err != nil {
		return WebSocketSessionResult{}, err
	}
	upgrade, err := requester.OpenWebSocket(sessionCtx, requestPath, spec.Subprotocol)
	if err != nil {
		return WebSocketSessionResult{}, websocketSessionTransportError(b, op, err)
	}
	conn := upgrade.Conn
	cancelDone := make(chan struct{})
	stopCancel := context.AfterFunc(sessionCtx, func() {
		_ = conn.Close()
		close(cancelDone)
	})
	defer func() {
		if !stopCancel() {
			<-cancelDone
		}
		_ = conn.Close()
	}()

	if err := writeWebSocketClientFrame(sessionCtx, conn, websocketFrameOpcodeText, update, spec.MaxFrameBytes); err != nil {
		return WebSocketSessionResult{}, websocketSessionFrameError(err)
	}
	for offset := 0; offset < len(req.PCM16); {
		end := offset + spec.MaxFrameBytes
		if end > len(req.PCM16) {
			end = len(req.PCM16)
		}
		if err := writeWebSocketClientFrame(sessionCtx, conn, websocketFrameOpcodeBinary, req.PCM16[offset:end], spec.MaxFrameBytes); err != nil {
			return WebSocketSessionResult{}, websocketSessionFrameError(err)
		}
		offset = end
	}
	if err := writeWebSocketClientFrame(sessionCtx, conn, websocketFrameOpcodeClose, []byte{0x03, 0xe8}, spec.MaxFrameBytes); err != nil {
		return WebSocketSessionResult{}, websocketSessionFrameError(err)
	}

	events, received, err := readWebSocketSessionEvents(sessionCtx, conn, spec.MaxFrameBytes, spec.MaxOutputBytes)
	if err != nil {
		return WebSocketSessionResult{}, websocketSessionFrameError(err)
	}
	return WebSocketSessionResult{
		Connector:     b.Name,
		Operation:     op.ID,
		Method:        http.MethodGet,
		Path:          spec.Path,
		Status:        http.StatusSwitchingProtocols,
		BytesSent:     len(req.PCM16),
		BytesReceived: received,
		Events:        events,
	}, nil
}

func operationWebSocketSessionSpec(b Bundle, id string) (OperationSpec, error) {
	op, err := findOperation(b, id)
	if err != nil {
		return OperationSpec{}, err
	}
	if op.Kind != "websocket_session" || op.WebSocket == nil {
		return OperationSpec{}, fmt.Errorf("websocket session requires websocket_session operation, got %q", op.Kind)
	}
	if err := validateWebSocketSessionSemantics(0, op); err != nil {
		return OperationSpec{}, err
	}
	if err := requireOperationWebSocketSessionEndpoint(b, op); err != nil {
		return OperationSpec{}, err
	}
	return op, nil
}

// OperationWebSocketSessionMetadata returns only the file-bound needed by the
// command boundary. The raw WebSocket declaration remains engine-private so it
// cannot become a generic transport configuration surface.
func OperationWebSocketSessionMetadata(b Bundle, id string) (connectors.OperationWebSocketSessionMetadata, error) {
	op, err := operationWebSocketSessionSpec(b, id)
	if err != nil {
		return connectors.OperationWebSocketSessionMetadata{}, err
	}
	return connectors.OperationWebSocketSessionMetadata{
		Operation:     op.ID,
		OutputPolicy:  op.OutputPolicy,
		MaxInputBytes: op.WebSocket.MaxInputBytes,
	}, nil
}

// PreflightOperationWebSocketSession proves the command's fixed binding
// agrees with the declaration before credentials, local files, or a network
// connection are touched.
func PreflightOperationWebSocketSession(b Bundle, id, method, path, outputPolicy string) error {
	op, err := operationWebSocketSessionSpec(b, id)
	if err != nil {
		return err
	}
	if got := strings.ToUpper(strings.TrimSpace(method)); got != http.MethodGet {
		return fmt.Errorf("websocket session %q command method must be GET, got %s", op.ID, got)
	}
	if path != op.WebSocket.Path {
		return fmt.Errorf("websocket session %q command path %q does not match declared path %q", op.ID, path, op.WebSocket.Path)
	}
	if outputPolicy != op.OutputPolicy {
		return fmt.Errorf("websocket session %q command output_policy %q does not match declared policy %q", op.ID, outputPolicy, op.OutputPolicy)
	}
	return nil
}

func requireOperationWebSocketSessionEndpoint(b Bundle, op OperationSpec) error {
	if b.Surface == nil {
		return nil
	}
	for _, endpoint := range b.Surface.Endpoints {
		if !strings.EqualFold(endpoint.Method, http.MethodGet) || endpoint.Path != op.WebSocket.Path {
			continue
		}
		if endpoint.CoveredBy != nil && endpoint.CoveredBy.WebSocketSession == op.ID {
			return nil
		}
		// A source-ledger operation row is the only valid bootstrap state for
		// surface-reconcile: runtime preflight proves this exact declaration
		// before the generator replaces that row with covered_by. Requiring the
		// generated coverage here would make that transition impossible.
		if endpoint.Operation != nil && endpoint.Operation.Model == "websocket_session" {
			return nil
		}
	}
	return fmt.Errorf("api_surface endpoint GET %s is not declared for websocket session %q", op.WebSocket.Path, op.ID)
}

func websocketSessionUpdatePayload(op OperationSpec, update map[string]any) ([]byte, error) {
	if update == nil {
		return nil, fmt.Errorf("websocket session %q requires session.update input", op.ID)
	}
	schema, err := CompileSchema(op.WebSocket.SessionUpdateSchema)
	if err != nil {
		return nil, fmt.Errorf("websocket session %q compile session_update_schema: %w", op.ID, err)
	}
	if err := schema.Validate(update); err != nil {
		return nil, fmt.Errorf("websocket session %q session_update_schema: %w", op.ID, err)
	}
	payload, err := json.Marshal(update)
	if err != nil {
		return nil, fmt.Errorf("websocket session %q encode session.update: %w", op.ID, err)
	}
	return payload, nil
}

func validateWebSocketSessionPCM16(update, pcm16 []byte, spec *WebSocketSessionSpec) error {
	if len(pcm16) == 0 {
		return fmt.Errorf("websocket session requires non-empty PCM16 input")
	}
	if len(pcm16)%2 != 0 {
		return fmt.Errorf("websocket session PCM16 input must contain complete 16-bit samples")
	}
	if len(update) > spec.MaxFrameBytes {
		return fmt.Errorf("websocket session session.update exceeds declared max_frame_bytes")
	}
	if len(update) > spec.MaxInputBytes || len(pcm16) > spec.MaxInputBytes-len(update) {
		return fmt.Errorf("websocket session input exceeds declared max_input_bytes")
	}
	return nil
}

func writeWebSocketClientFrame(ctx context.Context, writer io.Writer, opcode byte, payload []byte, maxFrameBytes int) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if len(payload) > maxFrameBytes {
		return fmt.Errorf("websocket client frame exceeds declared max_frame_bytes")
	}
	if opcode == websocketFrameOpcodeClose && len(payload) > 125 {
		return fmt.Errorf("websocket close frame exceeds RFC control-frame bound")
	}

	var mask [4]byte
	if _, err := cryptorand.Read(mask[:]); err != nil {
		return fmt.Errorf("generate websocket client frame mask: %w", err)
	}
	header := websocketClientFrameHeader(opcode, len(payload), mask)
	masked := make([]byte, len(payload))
	for i := range payload {
		masked[i] = payload[i] ^ mask[i%len(mask)]
	}
	if err := writeAll(ctx, writer, header); err != nil {
		return err
	}
	return writeAll(ctx, writer, masked)
}

func websocketClientFrameHeader(opcode byte, payloadLen int, mask [4]byte) []byte {
	header := make([]byte, 0, 14)
	header = append(header, 0x80|opcode)
	switch {
	case payloadLen < 126:
		header = append(header, 0x80|byte(payloadLen))
	case payloadLen <= math.MaxUint16:
		header = append(header, 0x80|126, byte(payloadLen>>8), byte(payloadLen))
	default:
		header = append(header, 0x80|127)
		var size [8]byte
		binary.BigEndian.PutUint64(size[:], uint64(payloadLen))
		header = append(header, size[:]...)
	}
	return append(header, mask[:]...)
}

func writeAll(ctx context.Context, writer io.Writer, payload []byte) error {
	for len(payload) > 0 {
		if err := ctx.Err(); err != nil {
			return err
		}
		written, err := writer.Write(payload)
		if err != nil {
			if ctx.Err() != nil {
				return ctx.Err()
			}
			return err
		}
		if written <= 0 {
			return io.ErrShortWrite
		}
		payload = payload[written:]
	}
	return nil
}

func readWebSocketSessionEvents(ctx context.Context, reader io.ReadWriter, maxFrameBytes, maxOutputBytes int) ([]any, int, error) {
	events := make([]any, 0)
	received := 0
	for {
		if err := ctx.Err(); err != nil {
			return nil, 0, err
		}
		frame, err := readWebSocketServerFrame(reader, maxFrameBytes)
		if err != nil {
			if ctx.Err() != nil {
				return nil, 0, ctx.Err()
			}
			return nil, 0, err
		}
		switch frame.opcode {
		case websocketFrameOpcodeClose:
			return events, received, nil
		case websocketFrameOpcodePing:
			if err := writeWebSocketClientFrame(ctx, reader, websocketFrameOpcodePong, frame.payload, maxFrameBytes); err != nil {
				return nil, 0, err
			}
		case websocketFrameOpcodePong:
			continue
		case websocketFrameOpcodeText:
			if len(frame.payload) > maxOutputBytes-received {
				return nil, 0, fmt.Errorf("websocket session output exceeds declared max_output_bytes")
			}
			event, err := decodeWebSocketJSONEvent(frame.payload)
			if err != nil {
				return nil, 0, err
			}
			received += len(frame.payload)
			events = append(events, redactJSONValue(event))
		default:
			return nil, 0, fmt.Errorf("websocket session received unsupported frame opcode %d", frame.opcode)
		}
	}
}

type websocketServerFrame struct {
	opcode  byte
	payload []byte
}

func readWebSocketServerFrame(reader io.Reader, maxFrameBytes int) (websocketServerFrame, error) {
	var head [2]byte
	if _, err := io.ReadFull(reader, head[:]); err != nil {
		return websocketServerFrame{}, err
	}
	if head[0]&0x70 != 0 || head[0]&0x80 == 0 {
		return websocketServerFrame{}, fmt.Errorf("websocket server frame must be final and cannot set reserved bits")
	}
	if head[1]&0x80 != 0 {
		return websocketServerFrame{}, fmt.Errorf("websocket server frame must not be masked")
	}
	opcode := head[0] & 0x0f
	length, err := websocketServerFrameLength(reader, head[1]&0x7f)
	if err != nil {
		return websocketServerFrame{}, err
	}
	if length > uint64(maxFrameBytes) || length > uint64(math.MaxInt) {
		return websocketServerFrame{}, fmt.Errorf("websocket server frame exceeds declared max_frame_bytes")
	}
	if (opcode == websocketFrameOpcodeClose || opcode == websocketFrameOpcodePing || opcode == websocketFrameOpcodePong) && length > 125 {
		return websocketServerFrame{}, fmt.Errorf("websocket control frame exceeds RFC control-frame bound")
	}
	payload := make([]byte, int(length))
	if _, err := io.ReadFull(reader, payload); err != nil {
		return websocketServerFrame{}, err
	}
	return websocketServerFrame{opcode: opcode, payload: payload}, nil
}

func websocketServerFrameLength(reader io.Reader, encoded byte) (uint64, error) {
	switch encoded {
	case 126:
		var size [2]byte
		if _, err := io.ReadFull(reader, size[:]); err != nil {
			return 0, err
		}
		return uint64(binary.BigEndian.Uint16(size[:])), nil
	case 127:
		var size [8]byte
		if _, err := io.ReadFull(reader, size[:]); err != nil {
			return 0, err
		}
		value := binary.BigEndian.Uint64(size[:])
		if value&1<<63 != 0 {
			return 0, fmt.Errorf("websocket server frame length is invalid")
		}
		return value, nil
	default:
		return uint64(encoded), nil
	}
}

func decodeWebSocketJSONEvent(payload []byte) (any, error) {
	decoder := json.NewDecoder(bytes.NewReader(payload))
	decoder.UseNumber()
	var event any
	if err := decoder.Decode(&event); err != nil {
		return nil, fmt.Errorf("websocket session event is not JSON: %w", err)
	}
	var trailing any
	if err := decoder.Decode(&trailing); !errors.Is(err, io.EOF) {
		return nil, fmt.Errorf("websocket session event must contain one JSON value")
	}
	return event, nil
}

func websocketSessionTransportError(b Bundle, op OperationSpec, err error) error {
	var httpErr *connsdk.HTTPError
	if errors.As(err, &httpErr) {
		class, hint := applyErrorMap(b.HTTP.ErrorMap, err)
		message := fmt.Sprintf("websocket session %q received HTTP status %d", op.ID, httpErr.Status)
		if class != "" {
			message = class + ": " + message
		}
		if hint != "" {
			message += ": " + hint
		}
		return errors.New(message)
	}
	return fmt.Errorf("websocket session %q: %s", op.ID, safety.RedactErrorText(err.Error()))
}

func websocketSessionFrameError(err error) error {
	if err == nil {
		return nil
	}
	return fmt.Errorf("websocket session: %s", safety.RedactErrorText(err.Error()))
}
