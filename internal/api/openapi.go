package api

import (
	_ "embed"
	"encoding/json"
	"net/http"
	"sync"
)

//go:embed openapi.json
var openapiBase []byte

//go:embed openapi.fork.json
var openapiFork []byte

var (
	openapiMerged     []byte
	openapiMergedOnce sync.Once
)

// handleOpenAPI serves the OpenAPI 3.1 spec describing Librarr's HTTP API.
// AI agents and tooling can introspect this to discover endpoints, request
// shapes, and response shapes without prior knowledge of the codebase.
//
// The served document is openapi.json (the upstream base) with the
// fork-specific fragment from openapi.fork.json merged in, so each side edits
// its own file and the two never conflict.
func (s *Server) handleOpenAPI(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.Header().Set("Cache-Control", "public, max-age=600")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(mergedOpenAPISpec())
}

// mergedOpenAPISpec returns the base spec with fork paths and schemas merged
// in. The result is built once and cached; on any malformed input it falls
// back to the base document so the endpoint never breaks.
func mergedOpenAPISpec() []byte {
	openapiMergedOnce.Do(func() {
		openapiMerged = buildMergedOpenAPISpec(openapiBase, openapiFork)
	})
	return openapiMerged
}

// buildMergedOpenAPISpec overlays forkDoc's "paths" and
// "components.schemas" entries onto baseDoc and re-marshals the result.
func buildMergedOpenAPISpec(baseDoc, forkDoc []byte) []byte {
	var base map[string]any
	if err := json.Unmarshal(baseDoc, &base); err != nil {
		return baseDoc
	}
	var fork map[string]any
	if err := json.Unmarshal(forkDoc, &fork); err != nil {
		return baseDoc
	}
	if paths, ok := fork["paths"].(map[string]any); ok {
		basePaths, ok := base["paths"].(map[string]any)
		if !ok {
			basePaths = map[string]any{}
			base["paths"] = basePaths
		}
		for k, v := range paths {
			basePaths[k] = v
		}
	}
	if forkComponents, ok := fork["components"].(map[string]any); ok {
		baseComponents, ok := base["components"].(map[string]any)
		if !ok {
			baseComponents = map[string]any{}
			base["components"] = baseComponents
		}
		if forkSchemas, ok := forkComponents["schemas"].(map[string]any); ok {
			baseSchemas, ok := baseComponents["schemas"].(map[string]any)
			if !ok {
				baseSchemas = map[string]any{}
				baseComponents["schemas"] = baseSchemas
			}
			for k, v := range forkSchemas {
				baseSchemas[k] = v
			}
		}
	}
	merged, err := json.Marshal(base)
	if err != nil {
		return baseDoc
	}
	return merged
}
