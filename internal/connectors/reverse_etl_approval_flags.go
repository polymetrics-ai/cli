package connectors

import (
	_ "embed"
	"encoding/json"
	"fmt"
)

//go:embed reverse_etl_approval_flags.json
var reverseETLApprovalFlagsJSON []byte

var reverseETLApprovalFlags = mustParseReverseETLApprovalFlags()

func mustParseReverseETLApprovalFlags() []CommandSurfaceFlag {
	var flags []CommandSurfaceFlag
	if err := json.Unmarshal(reverseETLApprovalFlagsJSON, &flags); err != nil {
		panic(fmt.Sprintf("parse reverse_etl_approval_flags.json: %v", err))
	}
	if len(flags) == 0 {
		panic("reverse_etl_approval_flags.json declares no flags")
	}
	for _, flag := range flags {
		if flag.Name == "" {
			panic("reverse_etl_approval_flags.json declares a flag with no name")
		}
	}
	return flags
}

// ReverseETLApprovalFlags returns the shared approval-carrier flags for write commands.
func ReverseETLApprovalFlags() []CommandSurfaceFlag {
	return append([]CommandSurfaceFlag(nil), reverseETLApprovalFlags...)
}
