package engine

import (
	"strings"
	"testing"
	"testing/fstest"
)

// TestBundleLoadAcceptsClosedWebSocketSessionContract is the foundation RED
// contract. A provider WebSocket must be a fixed declaration-owned operation:
// no caller-selectable URL, protocol, frame type, or unbounded payload is
// acceptable as a substitute.
func TestBundleLoadAcceptsClosedWebSocketSessionContract(t *testing.T) {
	fSys := fullValidBundleFS("acme")
	fSys["acme/api_surface.json"].Data = []byte(`{
		"api": "test API v1",
		"endpoints": [
			{"method":"GET","path":"/widgets","covered_by":{"stream":"widgets"}},
			{"method":"GET","path":"/live","covered_by":{"websocket_session":"acme.live"}}
		]
	}`)
	fSys["acme/operations.json"] = mapFile(`{
		"operations": [{
			"id": "acme.live",
			"kind": "websocket_session",
			"summary": "Stream one declared live transcription session",
			"risk": "medium",
			"approval": "none",
			"output_policy": "json_redacted",
			"websocket": {
				"method": "GET",
				"path": "/live",
				"subprotocol": "live-asr",
				"max_input_bytes": 65536,
				"max_output_bytes": 65536,
				"max_frame_bytes": 4096,
				"session_update_schema": {
					"type": "object",
					"additionalProperties": false,
					"required": ["type", "input_audio_format"],
					"properties": {
						"type": {"type": "string", "enum": ["session.update"]},
						"input_audio_format": {"type": "string", "enum": ["pcm16"]}
					}
				}
			}
		}]
	}`)

	if _, err := Load(fSys, "acme"); err != nil {
		t.Fatalf("Load closed WebSocket session operation: %v", err)
	}
}

func TestBundleRejectsUnsafeWebSocketSessionContracts(t *testing.T) {
	tests := []struct {
		name   string
		mutate func(string) string
		want   string
	}{
		{
			name: "non_get_upgrade",
			mutate: func(operation string) string {
				return strings.Replace(operation, `"method": "GET"`, `"method": "POST"`, 1)
			},
			want: "websocket_session method must be GET",
		},
		{
			name: "absolute_endpoint",
			mutate: func(operation string) string {
				return strings.Replace(operation, `"path": "/live"`, `"path": "https://untrusted.example/live"`, 1)
			},
			want: "websocket_session path must be connector-relative",
		},
		{
			name: "empty_subprotocol",
			mutate: func(operation string) string {
				return strings.Replace(operation, `"subprotocol": "live-asr"`, `"subprotocol": ""`, 1)
			},
			want: "websocket_session requires subprotocol",
		},
		{
			name: "unbounded_frame",
			mutate: func(operation string) string {
				return strings.Replace(operation, `"max_frame_bytes": 4096`, `"max_frame_bytes": 0`, 1)
			},
			want: "websocket_session max_frame_bytes must be positive",
		},
		{
			name: "open_session_update",
			mutate: func(operation string) string {
				return strings.Replace(operation, `"additionalProperties": false`, `"additionalProperties": true`, 1)
			},
			want: "websocket_session session_update_schema must declare additionalProperties false",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fSys := websocketSessionBundleFS(tt.mutate)
			_, err := Load(fSys, "acme")
			if err == nil || !strings.Contains(err.Error(), tt.want) {
				t.Fatalf("Load unsafe websocket session contract = %v, want error containing %q", err, tt.want)
			}
		})
	}
}

func websocketSessionBundleFS(mutate func(string) string) fstest.MapFS {
	fSys := fullValidBundleFS("acme")
	fSys["acme/api_surface.json"].Data = []byte(`{
		"api": "test API v1",
		"endpoints": [
			{"method":"GET","path":"/widgets","covered_by":{"stream":"widgets"}},
			{"method":"GET","path":"/live","covered_by":{"websocket_session":"acme.live"}}
		]
	}`)
	operation := `{
		"operations": [{
			"id": "acme.live",
			"kind": "websocket_session",
			"summary": "Stream one declared live transcription session",
			"risk": "medium",
			"approval": "none",
			"output_policy": "json_redacted",
			"websocket": {
				"method": "GET",
				"path": "/live",
				"subprotocol": "live-asr",
				"max_input_bytes": 65536,
				"max_output_bytes": 65536,
				"max_frame_bytes": 4096,
				"session_update_schema": {
					"type": "object",
					"additionalProperties": false,
					"required": ["type", "input_audio_format"],
					"properties": {
						"type": {"type": "string", "enum": ["session.update"]},
						"input_audio_format": {"type": "string", "enum": ["pcm16"]}
					}
				}
			}
		}]
	}`
	fSys["acme/operations.json"] = mapFile(mutate(operation))
	return fSys
}

func mapFile(data string) *fstest.MapFile {
	return &fstest.MapFile{Data: []byte(data)}
}
