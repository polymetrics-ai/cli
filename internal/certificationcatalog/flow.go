package certificationcatalog

import "slices"

type FlowKind struct {
	ID              string
	SourceRole      string
	DestinationRole string
}

var (
	generatedFlowKindData    []FlowKind
	generatedFlowKindDataSet bool
)

func registerGeneratedFlowKinds(kinds []FlowKind) {
	if generatedFlowKindDataSet {
		generatedFlowKindData = nil
		return
	}
	generatedFlowKindDataSet = true
	generatedFlowKindData = slices.Clone(kinds)
}

func FlowKinds() []FlowKind {
	if !generatedFlowKindDataSet || !validFlowKinds(generatedFlowKindData) {
		return nil
	}
	return slices.Clone(generatedFlowKindData)
}

func validFlowKinds(kinds []FlowKind) bool {
	if len(kinds) == 0 {
		return false
	}
	identifiers := make(map[string]bool, len(kinds))
	for _, kind := range kinds {
		if !validFlowKindIdentifier(kind.ID) || !validFlowKindIdentifier(kind.SourceRole) || !validFlowKindIdentifier(kind.DestinationRole) || identifiers[kind.ID] {
			return false
		}
		identifiers[kind.ID] = true
	}
	return true
}

func validFlowKindIdentifier(value string) bool {
	if len(value) == 0 || len(value) > 200 {
		return false
	}
	for index := range len(value) {
		character := value[index]
		alphanumeric := character >= 'A' && character <= 'Z' || character >= 'a' && character <= 'z' || character >= '0' && character <= '9'
		if !alphanumeric && (index == 0 || character != '.' && character != '_' && character != ':' && character != '-') {
			return false
		}
	}
	return true
}
