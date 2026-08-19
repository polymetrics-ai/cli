package synccontract

import (
	"crypto/sha256"
	_ "embed"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"sort"
	"sync"
)

//go:embed testdata/conformance/v1.json
var embeddedFixtureCorpus []byte

// ConformanceFixture is an immutable, versioned scenario that every engine
// lane consumes unchanged when proving a native sync executor.
type ConformanceFixture struct {
	ID              string `json:"id"`
	Guarantee       string `json:"guarantee"`
	ExpectedOutcome string `json:"expected_outcome"`
}

type fixtureCorpus struct {
	Version  uint                 `json:"version"`
	Fixtures []ConformanceFixture `json:"fixtures"`
}

var (
	loadFixturesOnce sync.Once
	loadedFixtures   fixtureCorpus
)

func mustFixtureCorpus() fixtureCorpus {
	loadFixturesOnce.Do(func() {
		if err := json.Unmarshal(embeddedFixtureCorpus, &loadedFixtures); err != nil {
			panic(fmt.Sprintf("parse embedded sync conformance fixtures: %v", err))
		}
		if loadedFixtures.Version == 0 || len(loadedFixtures.Fixtures) == 0 {
			panic("embedded sync conformance fixtures are incomplete")
		}
		seen := make(map[string]struct{}, len(loadedFixtures.Fixtures))
		for _, fixture := range loadedFixtures.Fixtures {
			if fixture.ID == "" || fixture.Guarantee == "" || fixture.ExpectedOutcome == "" {
				panic("embedded sync conformance fixture has an empty required field")
			}
			if _, exists := seen[fixture.ID]; exists {
				panic("embedded sync conformance fixture IDs must be unique")
			}
			seen[fixture.ID] = struct{}{}
		}
	})
	return loadedFixtures
}

// ConformanceFixtures returns a defensive copy of the corpus. Engine lanes
// may select cases but cannot alter the product-wide definitions.
func ConformanceFixtures() []ConformanceFixture {
	corpus := mustFixtureCorpus()
	return append([]ConformanceFixture(nil), corpus.Fixtures...)
}

// ConformanceEvidence identifies the exact fixture corpus an executor has
// proved. It contains case identities, not user-controlled protocol payloads.
type ConformanceEvidence struct {
	FixtureVersion uint     `json:"fixture_version"`
	FixtureDigest  string   `json:"fixture_digest"`
	FixtureIDs     []string `json:"fixture_ids"`
}

// RequiredConformanceEvidence returns the evidence a native contract must
// carry before it can become executable.
func RequiredConformanceEvidence() ConformanceEvidence {
	corpus := mustFixtureCorpus()
	digest := sha256.Sum256(embeddedFixtureCorpus)
	ids := make([]string, 0, len(corpus.Fixtures))
	for _, fixture := range corpus.Fixtures {
		ids = append(ids, fixture.ID)
	}
	sort.Strings(ids)
	return ConformanceEvidence{
		FixtureVersion: corpus.Version,
		FixtureDigest:  hex.EncodeToString(digest[:]),
		FixtureIDs:     ids,
	}
}

func (e ConformanceEvidence) matchesRequired() bool {
	required := RequiredConformanceEvidence()
	if e.FixtureVersion != required.FixtureVersion || e.FixtureDigest != required.FixtureDigest || len(e.FixtureIDs) != len(required.FixtureIDs) {
		return false
	}
	got := append([]string(nil), e.FixtureIDs...)
	want := append([]string(nil), required.FixtureIDs...)
	sort.Strings(got)
	sort.Strings(want)
	for i := range got {
		if got[i] != want[i] {
			return false
		}
	}
	return true
}

func (e ConformanceEvidence) equal(other ConformanceEvidence) bool {
	if e.FixtureVersion != other.FixtureVersion || e.FixtureDigest != other.FixtureDigest || len(e.FixtureIDs) != len(other.FixtureIDs) {
		return false
	}
	left := append([]string(nil), e.FixtureIDs...)
	right := append([]string(nil), other.FixtureIDs...)
	sort.Strings(left)
	sort.Strings(right)
	for i := range left {
		if left[i] != right[i] {
			return false
		}
	}
	return true
}
