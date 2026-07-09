package conformance

import (
	"fmt"
	"os"
	"sort"
	"testing"

	"polymetrics.ai/internal/connectors/engine"
)

func TestFreshchatImplementedETLCommandsHaveReplayFixtures(t *testing.T) {
	b, err := engine.Load(os.DirFS("../defs"), "freshchat")
	if err != nil {
		t.Fatalf("load freshchat bundle: %v", err)
	}
	if b.CLISurface == nil {
		t.Fatal("freshchat CLISurface is nil")
	}

	commandsByStream := map[string]string{}
	for _, cmd := range b.CLISurface.Commands {
		if cmd.Intent != "etl" || cmd.Availability != "implemented" || cmd.Stream == "" {
			continue
		}
		if prev, exists := commandsByStream[cmd.Stream]; exists {
			t.Fatalf("stream %q is mapped by multiple ETL commands: %q and %q", cmd.Stream, prev, cmd.Path)
		}
		commandsByStream[cmd.Stream] = cmd.Path
	}
	if len(commandsByStream) != 18 {
		t.Fatalf("implemented Freshchat ETL command streams = %d, want 18", len(commandsByStream))
	}

	streamNames := make([]string, 0, len(commandsByStream))
	for stream := range commandsByStream {
		streamNames = append(streamNames, stream)
	}
	sort.Strings(streamNames)

	replay := newReusableStreamReplayServer()
	defer replay.Close()
	for _, stream := range streamNames {
		stream := stream
		command := commandsByStream[stream]
		t.Run(fmt.Sprintf("%s/%s", stream, command), func(t *testing.T) {
			pages, err := loadFixturePages(b.Fixtures, stream)
			if err != nil {
				t.Fatalf("load fixture pages: %v", err)
			}
			if len(pages) == 0 {
				t.Fatalf("stream %q mapped by command %q has no replay fixture pages", stream, command)
			}

			count := 0
			if err := readRawRecordsWithReplay(b, stream, nil, replay, func(map[string]any) error {
				count++
				return nil
			}); err != nil {
				t.Fatalf("replay read failed: %v", err)
			}
			if count == 0 {
				t.Fatalf("stream %q mapped by command %q emitted zero fixture records", stream, command)
			}
		})
	}
}
