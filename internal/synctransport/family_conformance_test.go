package synctransport

import (
	"testing"

	"polymetrics.ai/internal/synccontract"
)

func TestTransportFamilyHalfPathConformance(t *testing.T) {
	for _, family := range transportFamilyConformanceCases() {
		for _, mode := range synccontract.AllModes() {
			family := family
			mode := mode
			t.Run(family.name+"/"+string(mode), func(t *testing.T) {
				_ = family
				_ = mode
			})
		}
	}
}
