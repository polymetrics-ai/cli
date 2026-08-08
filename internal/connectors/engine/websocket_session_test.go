package engine

import (
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

func mapFile(data string) *fstest.MapFile {
	return &fstest.MapFile{Data: []byte(data)}
}
