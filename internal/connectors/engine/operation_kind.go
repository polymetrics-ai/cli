package engine

import "strings"

type operationKindContract struct {
	executionBlock   string
	importParameters bool
}

var operationKindContracts = map[string]operationKindContract{
	"stream_etl":       {executionBlock: "composite"},
	"rest_read":        {executionBlock: "rest", importParameters: true},
	"rest_status":      {executionBlock: "rest", importParameters: true},
	"rest_write":       {executionBlock: "rest", importParameters: true},
	"provider_search":  {executionBlock: "rest", importParameters: true},
	"graphql_query":    {executionBlock: "graphql"},
	"graphql_mutation": {executionBlock: "graphql"},
	"xml_export":       {executionBlock: "xml"},
	"xml_import":       {executionBlock: "xml"},
	"binary_download":  {executionBlock: "binary", importParameters: true},
	"text_export":      {executionBlock: "binary", importParameters: true},
	"file_upload":      {executionBlock: "file"},
	"local_git":        {executionBlock: "local_git"},
	"local_file":       {executionBlock: "local_file"},
	"browser_open":     {executionBlock: "browser"},
	"composite":        {executionBlock: "composite"},
}

func expectedOperationBlock(kind string) string {
	return operationKindContracts[kind].executionBlock
}

func OperationRequestParameterBlock(kind string) (string, bool) {
	contract, ok := operationKindContracts[kind]
	return contract.executionBlock, ok && contract.importParameters
}

func IsReadSurfaceIntent(intent string) bool {
	switch intent {
	case "direct_read", "binary_download", "text_export", "status_check":
		return true
	default:
		return false
	}
}

func IsReadSurfaceMethod(method string) bool {
	switch strings.ToUpper(strings.TrimSpace(method)) {
	case "GET", "HEAD", "POST":
		return true
	default:
		return false
	}
}
