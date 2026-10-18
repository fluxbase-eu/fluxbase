package api

// addVectorEndpoints adds vector/embedding endpoints to the spec
func (h *OpenAPIHandler) addVectorEndpoints(spec *OpenAPISpec) {
	// Vector schemas
	spec.Components.Schemas["EmbedRequest"] = map[string]interface{}{
		"type":     "object",
		"required": []string{"input"},
		"properties": map[string]interface{}{
			"input": map[string]interface{}{
				"oneOf": []map[string]interface{}{
					{"type": "string"},
					{"type": "array", "items": map[string]string{"type": "string"}},
				},
			},
			"model": map[string]string{"type": "string"},
		},
	}

	spec.Components.Schemas["EmbedResponse"] = map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"embeddings": map[string]interface{}{
				"type": "array",
				"items": map[string]interface{}{
					"type":  "array",
					"items": map[string]string{"type": "number"},
				},
			},
			"model":      map[string]string{"type": "string"},
			"usage":      map[string]interface{}{"type": "object"},
			"dimensions": map[string]string{"type": "integer"},
		},
	}

	// Mirrors vector.SearchRequest (internal/api/vector/handler.go): either query or vector
	// must be provided to build the query embedding.
	spec.Components.Schemas["VectorSearchRequest"] = map[string]interface{}{
		"type":     "object",
		"required": []string{"table", "column"},
		"properties": map[string]interface{}{
			"table":           map[string]string{"type": "string", "description": "Table to search"},
			"column":          map[string]string{"type": "string", "description": "Vector column to compare against"},
			"query":           map[string]string{"type": "string", "description": "Natural-language query; embedded server-side (one of query/vector is required)"},
			"vector":          map[string]interface{}{"type": "array", "items": map[string]string{"type": "number"}, "description": "Pre-computed query embedding (one of query/vector is required)"},
			"metric":          map[string]interface{}{"type": "string", "enum": []string{"cosine", "l2", "inner_product"}, "default": "cosine"},
			"match_threshold": map[string]string{"type": "number", "description": "Minimum similarity threshold"},
			"match_count":     map[string]string{"type": "integer", "description": "Maximum number of matches to return"},
			"select":          map[string]string{"type": "string", "description": "Comma-separated columns to return (default: *)"},
			"filters": map[string]interface{}{
				"type":        "array",
				"description": "Row filters applied before the similarity search",
				"items": map[string]interface{}{
					"type":     "object",
					"required": []string{"column", "operator", "value"},
					"properties": map[string]interface{}{
						"column":   map[string]string{"type": "string"},
						"operator": map[string]string{"type": "string"},
						"value":    map[string]interface{}{},
					},
				},
			},
		},
	}

	// Mirrors vector.SearchResponse (internal/api/vector/handler.go)
	spec.Components.Schemas["VectorSearchResult"] = map[string]interface{}{
		"type":     "object",
		"required": []string{"data"},
		"properties": map[string]interface{}{
			"data": map[string]interface{}{
				"type":        "array",
				"items":       map[string]string{"type": "object"},
				"description": "Matched rows",
			},
			"distances": map[string]interface{}{
				"type":  "array",
				"items": map[string]string{"type": "number"},
			},
			"model": map[string]string{"type": "string", "description": "Embedding model used when the query was embedded server-side"},
		},
	}

	// GET /api/v1/capabilities/vector
	spec.Paths["/api/v1/capabilities/vector"] = OpenAPIPath{
		"get": OpenAPIOperation{
			Summary:     "Get vector capabilities",
			Description: "Get available vector/embedding capabilities",
			OperationID: "vector_capabilities",
			Tags:        []string{"Vector"},
			Responses: map[string]OpenAPIResponse{
				"200": {
					Description: "Vector capabilities",
					Content: map[string]OpenAPIMedia{
						"application/json": {
							Schema: map[string]interface{}{
								"type": "object",
								"properties": map[string]interface{}{
									"enabled":    map[string]string{"type": "boolean"},
									"models":     map[string]interface{}{"type": "array", "items": map[string]string{"type": "string"}},
									"dimensions": map[string]string{"type": "integer"},
								},
							},
						},
					},
				},
			},
		},
	}

	// POST /api/v1/vector/embed
	spec.Paths["/api/v1/vector/embed"] = OpenAPIPath{
		"post": OpenAPIOperation{
			Summary:     "Generate embeddings",
			Description: "Generate vector embeddings for text input",
			OperationID: "vector_embed",
			Tags:        []string{"Vector"},
			Security: []map[string][]string{
				{"bearerAuth": {}},
			},
			RequestBody: &OpenAPIRequestBody{
				Required: true,
				Content: map[string]OpenAPIMedia{
					"application/json": {
						Schema: map[string]string{"$ref": "#/components/schemas/EmbedRequest"},
					},
				},
			},
			Responses: map[string]OpenAPIResponse{
				"200": {
					Description: "Embeddings generated",
					Content: map[string]OpenAPIMedia{
						"application/json": {
							Schema: map[string]string{"$ref": "#/components/schemas/EmbedResponse"},
						},
					},
				},
			},
		},
	}

	// POST /api/v1/vector/search
	spec.Paths["/api/v1/vector/search"] = OpenAPIPath{
		"post": OpenAPIOperation{
			Summary:     "Vector similarity search",
			Description: "Search for similar vectors in a table",
			OperationID: "vector_search",
			Tags:        []string{"Vector"},
			Security: []map[string][]string{
				{"bearerAuth": {}},
			},
			RequestBody: &OpenAPIRequestBody{
				Required: true,
				Content: map[string]OpenAPIMedia{
					"application/json": {
						Schema: map[string]string{"$ref": "#/components/schemas/VectorSearchRequest"},
					},
				},
			},
			Responses: map[string]OpenAPIResponse{
				"200": {
					Description: "Search results",
					Content: map[string]OpenAPIMedia{
						"application/json": {
							Schema: map[string]string{"$ref": "#/components/schemas/VectorSearchResult"},
						},
					},
				},
			},
		},
	}
}
