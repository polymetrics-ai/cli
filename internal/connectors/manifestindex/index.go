// Package manifestindex stores compact immutable execution-manifest metadata.
package manifestindex

import (
	"fmt"
	"sort"
)

type Entry struct{ Connector, Generation, Digest, Executor string }
type Index struct{ entries []Entry }

func New(entries []Entry, maxEntries int) (Index, error) {
	if len(entries) > maxEntries {
		return Index{}, fmt.Errorf("manifest index exceeds entry limit")
	}
	out := append([]Entry(nil), entries...)
	sort.Slice(out, func(i, j int) bool { return out[i].Connector < out[j].Connector })
	for i, entry := range out {
		if entry.Connector == "" || entry.Generation == "" || entry.Digest == "" || entry.Executor == "" {
			return Index{}, fmt.Errorf("manifest entry %d is incomplete", i)
		}
		if i > 0 && out[i-1].Connector == entry.Connector {
			return Index{}, fmt.Errorf("duplicate manifest connector %q", entry.Connector)
		}
	}
	return Index{entries: out}, nil
}
func (i Index) List() []Entry { return append([]Entry(nil), i.entries...) }
func (i Index) Lookup(connector string) (Entry, bool) {
	n := sort.Search(len(i.entries), func(n int) bool { return i.entries[n].Connector >= connector })
	if n < len(i.entries) && i.entries[n].Connector == connector {
		return i.entries[n], true
	}
	return Entry{}, false
}
