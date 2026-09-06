// Package manifestindex stores compact immutable execution-manifest metadata.
package manifestindex

import (
	"fmt"
	"sort"
	"strings"

	"polymetrics.ai/internal/connectors"
)

type unknownExecutorError struct {
	executor string
}

func (e unknownExecutorError) Error() string {
	return fmt.Sprintf("unknown executor %q", e.executor)
}

func (e unknownExecutorError) UnknownExecutor() string {
	return e.executor
}

func knownExecutor(executor string) bool {
	if executor == "api_engine.v1" {
		return true
	}
	for _, prefix := range []string{"native_database/", "closed_typed/"} {
		if !strings.HasPrefix(executor, prefix) {
			continue
		}
		id := strings.TrimPrefix(executor, prefix)
		return id != "" && strings.HasSuffix(id, ".v1")
	}
	return false
}

type Entry struct {
	Connector, Generation, Digest, Executor, Extension string
	CommandUsage, CommandTagline                       string
	Metadata                                           connectors.Metadata
	Bytes                                              int
}
type Index struct{ entries []Entry }

func New(entries []Entry, maxEntries int) (Index, error) {
	if len(entries) > maxEntries {
		return Index{}, fmt.Errorf("manifest index exceeds entry limit")
	}
	out := append([]Entry(nil), entries...)
	sort.Slice(out, func(i, j int) bool { return out[i].Connector < out[j].Connector })
	for i, entry := range out {
		if entry.Connector == "" || entry.Generation == "" || entry.Digest == "" || entry.Executor == "" || entry.Bytes <= 0 {
			return Index{}, fmt.Errorf("manifest entry %d is incomplete", i)
		}
		if entry.Metadata.Name != "" && entry.Metadata.Name != entry.Connector {
			return Index{}, fmt.Errorf("manifest entry %q metadata name %q does not match connector", entry.Connector, entry.Metadata.Name)
		}
		if entry.Extension != "" && entry.Extension != strings.TrimSpace(entry.Extension) {
			return Index{}, fmt.Errorf("manifest entry %q extension is invalid", entry.Connector)
		}
		if entry.CommandTagline != "" && strings.TrimSpace(entry.CommandUsage) == "" {
			return Index{}, fmt.Errorf("manifest entry %q command tagline has no usage", entry.Connector)
		}
		if !knownExecutor(entry.Executor) {
			return Index{}, unknownExecutorError{executor: entry.Executor}
		}
		if i > 0 && out[i-1].Connector == entry.Connector {
			return Index{}, fmt.Errorf("duplicate manifest connector %q", entry.Connector)
		}
	}
	return Index{entries: out}, nil
}
func (i Index) List() []Entry {
	out := make([]Entry, len(i.entries))
	for n, entry := range i.entries {
		out[n] = cloneEntry(entry)
	}
	return out
}
func (i Index) Lookup(connector string) (Entry, bool) {
	n := sort.Search(len(i.entries), func(n int) bool { return i.entries[n].Connector >= connector })
	if n < len(i.entries) && i.entries[n].Connector == connector {
		return cloneEntry(i.entries[n]), true
	}
	return Entry{}, false
}

func cloneEntry(entry Entry) Entry {
	if entry.Metadata.Icon != nil {
		icon := *entry.Metadata.Icon
		entry.Metadata.Icon = &icon
	}
	return entry
}
