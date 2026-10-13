package api

// addAIAdminEndpoints adds the user knowledge-base extensions and the
// admin AI surface (chatbots, providers, integrations, conversations,
// audit, table export, sync configs, document permissions) to the spec so
// the contract-drift check guards their request/response shapes. These
// routes were previously invisible to the drift check, which is exactly
// the bug class that let the SDK's toggle verb and KB update/delete
// routes drift.
func (h *OpenAPIHandler) addAIAdminEndpoints(spec *OpenAPISpec) {
	// ---- Schemas ----

	spec.Components.Schemas["KnowledgeBase"] = map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"id":                         map[string]string{"type": "string", "format": "uuid"},
			"name":                       map[string]string{"type": "string"},
			"namespace":                  map[string]string{"type": "string"},
			"description":                map[string]string{"type": "string"},
			"embedding_model":            map[string]string{"type": "string"},
			"embedding_dimensions":       map[string]string{"type": "integer"},
			"chunk_size":                 map[string]string{"type": "integer"},
			"chunk_overlap":              map[string]string{"type": "integer"},
			"chunk_strategy":             map[string]string{"type": "string"},
			"enabled":                    map[string]string{"type": "boolean"},
			"document_count":             map[string]string{"type": "integer"},
			"total_chunks":               map[string]string{"type": "integer"},
			"visibility":                 map[string]string{"type": "string", "enum": "private,shared,public"},
			"entity_extraction_enabled":  map[string]string{"type": "boolean"},
			"entity_extraction_strategy": map[string]string{"type": "string", "enum": ",rules,llm,hybrid"},
			"rerank_enabled":             map[string]string{"type": "boolean"},
			"owner_id":                   map[string]string{"type": "string", "format": "uuid"},
			"created_at":                 map[string]string{"type": "string", "format": "date-time"},
			"updated_at":                 map[string]string{"type": "string", "format": "date-time"},
		},
	}

	spec.Components.Schemas["KnowledgeBaseDocument"] = map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"id":                map[string]string{"type": "string", "format": "uuid"},
			"knowledge_base_id": map[string]string{"type": "string", "format": "uuid"},
			"title":             map[string]string{"type": "string"},
			"source_type":       map[string]string{"type": "string"},
			"mime_type":         map[string]string{"type": "string"},
			"content":           map[string]string{"type": "string"},
			"status":            map[string]string{"type": "string", "enum": "pending,processing,indexed,failed"},
			"chunks_count":      map[string]string{"type": "integer"},
			"tags":              map[string]interface{}{"type": "array", "items": map[string]string{"type": "string"}},
		},
	}

	spec.Components.Schemas["KBSearchRequest"] = map[string]interface{}{
		"type":     "object",
		"required": []string{"query"},
		"properties": map[string]interface{}{
			"query":           map[string]string{"type": "string"},
			"limit":           map[string]string{"type": "integer"},
			"mode":            map[string]string{"type": "string", "enum": "semantic,keyword,hybrid"},
			"semantic_weight": map[string]string{"type": "number"},
			"threshold":       map[string]string{"type": "number"},
			"rerank":          map[string]string{"type": "boolean"},
		},
	}

	spec.Components.Schemas["KBSearchResult"] = map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"chunk_id":            map[string]string{"type": "string"},
			"document_id":         map[string]string{"type": "string"},
			"document_title":      map[string]string{"type": "string"},
			"knowledge_base_name": map[string]string{"type": "string"},
			"content":             map[string]string{"type": "string"},
			"similarity":          map[string]string{"type": "number"},
		},
	}

	spec.Components.Schemas["TableExportSyncConfig"] = map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"id":                   map[string]string{"type": "string", "format": "uuid"},
			"knowledge_base_id":    map[string]string{"type": "string", "format": "uuid"},
			"schema_name":          map[string]string{"type": "string"},
			"table_name":           map[string]string{"type": "string"},
			"sync_mode":            map[string]string{"type": "string", "enum": "manual,automatic"},
			"sync_on_insert":       map[string]string{"type": "boolean"},
			"sync_on_update":       map[string]string{"type": "boolean"},
			"sync_on_delete":       map[string]string{"type": "boolean"},
			"debounce_seconds":     map[string]string{"type": "integer"},
			"include_foreign_keys": map[string]string{"type": "boolean"},
			"include_indexes":      map[string]string{"type": "boolean"},
		},
	}

	spec.Components.Schemas["AIMetrics"] = map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"total_requests":          map[string]string{"type": "integer"},
			"total_tokens":            map[string]string{"type": "integer"},
			"total_prompt_tokens":     map[string]string{"type": "integer"},
			"total_completion_tokens": map[string]string{"type": "integer"},
			"active_conversations":    map[string]string{"type": "integer"},
			"total_conversations":     map[string]string{"type": "integer"},
			"error_rate":              map[string]string{"type": "number"},
			"avg_response_time_ms":    map[string]string{"type": "number"},
		},
	}

	kbIDParam := OpenAPIParameter{Name: "id", In: "path", Required: true, Schema: map[string]string{"type": "string", "format": "uuid"}}
	jsonContent := func(schema interface{}) map[string]OpenAPIMedia {
		return map[string]OpenAPIMedia{"application/json": {Schema: schema}}
	}

	// ---- User knowledge-base extensions ----

	spec.Paths["/api/v1/ai/knowledge-bases/{id}"] = OpenAPIPath{
		"get": OpenAPIOperation{
			Summary:     "Get knowledge base",
			OperationID: "ai_kb_user_get",
			Tags:        []string{"Knowledge Bases"},
			Security:    []map[string][]string{{"bearerAuth": {}}},
			Parameters:  []OpenAPIParameter{kbIDParam},
			Responses: map[string]OpenAPIResponse{
				"200": {Description: "Knowledge base", Content: jsonContent(map[string]string{"$ref": "#/components/schemas/KnowledgeBase"})},
			},
		},
		"patch": OpenAPIOperation{
			Summary:     "Update knowledge base",
			Description: "Update a knowledge base the caller owns or has editor permission on.",
			OperationID: "ai_kb_user_update",
			Tags:        []string{"Knowledge Bases"},
			Security:    []map[string][]string{{"bearerAuth": {}}},
			Parameters:  []OpenAPIParameter{kbIDParam},
			RequestBody: &OpenAPIRequestBody{
				Content: jsonContent(map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"name":                       map[string]string{"type": "string"},
						"description":                map[string]string{"type": "string"},
						"visibility":                 map[string]string{"type": "string", "enum": "private,shared,public"},
						"chunk_size":                 map[string]string{"type": "integer"},
						"chunk_overlap":              map[string]string{"type": "integer"},
						"chunk_strategy":             map[string]string{"type": "string"},
						"enabled":                    map[string]string{"type": "boolean"},
						"entity_extraction_enabled":  map[string]string{"type": "boolean"},
						"entity_extraction_strategy": map[string]string{"type": "string", "enum": ",rules,llm,hybrid"},
						"rerank_enabled":             map[string]string{"type": "boolean"},
					},
				}),
			},
			Responses: map[string]OpenAPIResponse{
				"200": {Description: "Updated knowledge base", Content: jsonContent(map[string]string{"$ref": "#/components/schemas/KnowledgeBase"})},
				"403": {Description: "Editor permission required"},
				"404": {Description: "Knowledge base not found"},
			},
		},
		"delete": OpenAPIOperation{
			Summary:     "Delete knowledge base",
			Description: "Delete a knowledge base the caller owns. Removes all documents and chunks.",
			OperationID: "ai_kb_user_delete",
			Tags:        []string{"Knowledge Bases"},
			Security:    []map[string][]string{{"bearerAuth": {}}},
			Parameters:  []OpenAPIParameter{kbIDParam},
			Responses: map[string]OpenAPIResponse{
				"204": {Description: "Knowledge base deleted"},
				"403": {Description: "Only owner can delete"},
				"404": {Description: "Knowledge base not found"},
			},
		},
	}

	spec.Paths["/api/v1/ai/knowledge-bases/{id}/search"] = OpenAPIPath{
		"post": OpenAPIOperation{
			Summary:     "Search knowledge base",
			Description: "Hybrid (reciprocal rank fusion) search with optional reranking. Per-user isolation applies.",
			OperationID: "ai_kb_user_search",
			Tags:        []string{"Knowledge Bases"},
			Security:    []map[string][]string{{"bearerAuth": {}}},
			Parameters:  []OpenAPIParameter{kbIDParam},
			RequestBody: &OpenAPIRequestBody{
				Required: true,
				Content:  jsonContent(map[string]string{"$ref": "#/components/schemas/KBSearchRequest"}),
			},
			Responses: map[string]OpenAPIResponse{
				"200": {Description: "Search results", Content: jsonContent(map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"results": map[string]interface{}{"type": "array", "items": map[string]string{"$ref": "#/components/schemas/KBSearchResult"}},
						"count":   map[string]string{"type": "integer"},
					},
				})},
			},
		},
	}

	// ---- Admin: metrics, conversations, audit ----

	spec.Paths["/api/v1/admin/ai/metrics"] = OpenAPIPath{
		"get": OpenAPIOperation{
			Summary:     "Get AI metrics",
			OperationID: "ai_admin_metrics",
			Tags:        []string{"AI Admin"},
			Responses: map[string]OpenAPIResponse{
				"200": {Description: "Aggregated AI metrics", Content: jsonContent(map[string]string{"$ref": "#/components/schemas/AIMetrics"})},
			},
		},
	}

	spec.Paths["/api/v1/admin/ai/conversations"] = OpenAPIPath{
		"get": OpenAPIOperation{
			Summary:     "List AI conversations",
			OperationID: "ai_admin_conversations_list",
			Tags:        []string{"AI Admin"},
			Parameters: []OpenAPIParameter{
				{Name: "chatbot_id", In: "query", Schema: map[string]string{"type": "string", "format": "uuid"}},
				{Name: "user_id", In: "query", Schema: map[string]string{"type": "string", "format": "uuid"}},
				{Name: "status", In: "query", Schema: map[string]string{"type": "string"}},
				{Name: "limit", In: "query", Schema: map[string]string{"type": "integer"}},
				{Name: "offset", In: "query", Schema: map[string]string{"type": "integer"}},
			},
			Responses: map[string]OpenAPIResponse{
				"200": {Description: "Conversations with total count"},
			},
		},
	}

	spec.Paths["/api/v1/admin/ai/conversations/{id}/messages"] = OpenAPIPath{
		"get": OpenAPIOperation{
			Summary:     "Get AI conversation messages",
			OperationID: "ai_admin_conversation_messages",
			Tags:        []string{"AI Admin"},
			Parameters: []OpenAPIParameter{
				{Name: "id", In: "path", Required: true, Schema: map[string]string{"type": "string", "format": "uuid"}},
				{Name: "limit", In: "query", Schema: map[string]string{"type": "integer"}},
				{Name: "offset", In: "query", Schema: map[string]string{"type": "integer"}},
			},
			Responses: map[string]OpenAPIResponse{
				"200": {Description: "Messages"},
			},
		},
	}

	spec.Paths["/api/v1/admin/ai/audit"] = OpenAPIPath{
		"get": OpenAPIOperation{
			Summary:     "Get AI audit log",
			Description: "Generated-SQL audit trail for AI chatbot queries.",
			OperationID: "ai_admin_audit",
			Tags:        []string{"AI Admin"},
			Parameters: []OpenAPIParameter{
				{Name: "chatbot_id", In: "query", Schema: map[string]string{"type": "string", "format": "uuid"}},
				{Name: "user_id", In: "query", Schema: map[string]string{"type": "string", "format": "uuid"}},
				{Name: "success", In: "query", Schema: map[string]string{"type": "boolean"}},
				{Name: "limit", In: "query", Schema: map[string]string{"type": "integer"}},
				{Name: "offset", In: "query", Schema: map[string]string{"type": "integer"}},
			},
			Responses: map[string]OpenAPIResponse{
				"200": {Description: "Audit log entries"},
			},
		},
	}

	// ---- Admin: table export + sync configs ----

	spec.Paths["/api/v1/admin/ai/tables"] = OpenAPIPath{
		"get": OpenAPIOperation{
			Summary:     "List exportable tables",
			OperationID: "ai_admin_tables_list",
			Tags:        []string{"AI Admin"},
			Responses: map[string]OpenAPIResponse{
				"200": {Description: "Tables available for knowledge base export"},
			},
		},
	}

	spec.Paths["/api/v1/admin/ai/tables/{schema}/{table}/export"] = OpenAPIPath{
		"post": OpenAPIOperation{
			Summary:     "Export table to knowledge base",
			OperationID: "ai_admin_table_export",
			Tags:        []string{"AI Admin"},
			Parameters: []OpenAPIParameter{
				{Name: "schema", In: "path", Required: true, Schema: map[string]string{"type": "string"}},
				{Name: "table", In: "path", Required: true, Schema: map[string]string{"type": "string"}},
			},
			RequestBody: &OpenAPIRequestBody{
				Required: true,
				Content: jsonContent(map[string]interface{}{
					"type":     "object",
					"required": []string{"knowledge_base_id"},
					"properties": map[string]interface{}{
						"knowledge_base_id":    map[string]string{"type": "string", "format": "uuid"},
						"columns":              map[string]interface{}{"type": "array", "items": map[string]string{"type": "string"}},
						"include_sample_rows":  map[string]string{"type": "boolean"},
						"sample_row_count":     map[string]string{"type": "integer"},
						"include_foreign_keys": map[string]string{"type": "boolean"},
						"include_indexes":      map[string]string{"type": "boolean"},
						"user_id_column":       map[string]string{"type": "string"},
					},
				}),
			},
			Responses: map[string]OpenAPIResponse{
				"200": {Description: "Export result with document id"},
			},
		},
	}

	syncConfigBody := &OpenAPIRequestBody{
		Content: jsonContent(map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"schema_name":          map[string]string{"type": "string"},
				"table_name":           map[string]string{"type": "string"},
				"columns":              map[string]interface{}{"type": "array", "items": map[string]string{"type": "string"}},
				"sync_mode":            map[string]string{"type": "string", "enum": "manual,automatic"},
				"sync_on_insert":       map[string]string{"type": "boolean"},
				"sync_on_update":       map[string]string{"type": "boolean"},
				"sync_on_delete":       map[string]string{"type": "boolean"},
				"debounce_seconds":     map[string]string{"type": "integer"},
				"include_foreign_keys": map[string]string{"type": "boolean"},
				"include_indexes":      map[string]string{"type": "boolean"},
			},
		}),
	}

	spec.Paths["/api/v1/admin/ai/knowledge-bases/{id}/sync-configs"] = OpenAPIPath{
		"post": OpenAPIOperation{
			Summary:     "Create sync config",
			OperationID: "ai_admin_sync_config_create",
			Tags:        []string{"AI Admin"},
			Parameters:  []OpenAPIParameter{kbIDParam},
			RequestBody: syncConfigBody,
			Responses: map[string]OpenAPIResponse{
				"201": {Description: "Created sync config", Content: jsonContent(map[string]string{"$ref": "#/components/schemas/TableExportSyncConfig"})},
			},
		},
		"get": OpenAPIOperation{
			Summary:     "List sync configs",
			OperationID: "ai_admin_sync_config_list",
			Tags:        []string{"AI Admin"},
			Parameters:  []OpenAPIParameter{kbIDParam},
			Responses: map[string]OpenAPIResponse{
				"200": {Description: "Sync configs", Content: jsonContent(map[string]interface{}{
					"type": "array", "items": map[string]string{"$ref": "#/components/schemas/TableExportSyncConfig"},
				})},
			},
		},
	}

	spec.Paths["/api/v1/admin/ai/knowledge-bases/{id}/sync-configs/{syncId}"] = OpenAPIPath{
		"patch": OpenAPIOperation{
			Summary:     "Update sync config",
			OperationID: "ai_admin_sync_config_update",
			Tags:        []string{"AI Admin"},
			Parameters: []OpenAPIParameter{
				kbIDParam,
				{Name: "syncId", In: "path", Required: true, Schema: map[string]string{"type": "string", "format": "uuid"}},
			},
			RequestBody: syncConfigBody,
			Responses: map[string]OpenAPIResponse{
				"200": {Description: "Updated sync config", Content: jsonContent(map[string]string{"$ref": "#/components/schemas/TableExportSyncConfig"})},
			},
		},
		"delete": OpenAPIOperation{
			Summary:     "Delete sync config",
			OperationID: "ai_admin_sync_config_delete",
			Tags:        []string{"AI Admin"},
			Parameters: []OpenAPIParameter{
				kbIDParam,
				{Name: "syncId", In: "path", Required: true, Schema: map[string]string{"type": "string", "format": "uuid"}},
			},
			Responses: map[string]OpenAPIResponse{
				"204": {Description: "Deleted"},
			},
		},
	}

	spec.Paths["/api/v1/admin/ai/knowledge-bases/{id}/sync-configs/{syncId}/trigger"] = OpenAPIPath{
		"post": OpenAPIOperation{
			Summary:     "Trigger sync",
			OperationID: "ai_admin_sync_config_trigger",
			Tags:        []string{"AI Admin"},
			Parameters: []OpenAPIParameter{
				kbIDParam,
				{Name: "syncId", In: "path", Required: true, Schema: map[string]string{"type": "string", "format": "uuid"}},
			},
			Responses: map[string]OpenAPIResponse{
				"200": {Description: "Sync result"},
			},
		},
	}

	// ---- Admin: document permissions ----

	spec.Paths["/api/v1/admin/ai/knowledge-bases/{kb_id}/documents/{doc_id}/permissions"] = OpenAPIPath{
		"post": OpenAPIOperation{
			Summary:     "Grant document permission",
			OperationID: "ai_admin_doc_perm_grant",
			Tags:        []string{"AI Admin"},
			Parameters: []OpenAPIParameter{
				{Name: "kb_id", In: "path", Required: true, Schema: map[string]string{"type": "string", "format": "uuid"}},
				{Name: "doc_id", In: "path", Required: true, Schema: map[string]string{"type": "string", "format": "uuid"}},
			},
			RequestBody: &OpenAPIRequestBody{
				Required: true,
				Content: jsonContent(map[string]interface{}{
					"type":     "object",
					"required": []string{"user_id", "permission"},
					"properties": map[string]interface{}{
						"user_id":    map[string]string{"type": "string", "format": "uuid"},
						"permission": map[string]string{"type": "string", "enum": "viewer,editor,owner"},
					},
				}),
			},
			Responses: map[string]OpenAPIResponse{
				"201": {Description: "Permission granted"},
			},
		},
		"get": OpenAPIOperation{
			Summary:     "List document permissions",
			OperationID: "ai_admin_doc_perm_list",
			Tags:        []string{"AI Admin"},
			Parameters: []OpenAPIParameter{
				{Name: "kb_id", In: "path", Required: true, Schema: map[string]string{"type": "string", "format": "uuid"}},
				{Name: "doc_id", In: "path", Required: true, Schema: map[string]string{"type": "string", "format": "uuid"}},
			},
			Responses: map[string]OpenAPIResponse{
				"200": {Description: "Permission grants"},
			},
		},
	}

	spec.Paths["/api/v1/admin/ai/knowledge-bases/{kb_id}/documents/{doc_id}/permissions/{user_id}"] = OpenAPIPath{
		"delete": OpenAPIOperation{
			Summary:     "Revoke document permission",
			OperationID: "ai_admin_doc_perm_revoke",
			Tags:        []string{"AI Admin"},
			Parameters: []OpenAPIParameter{
				{Name: "kb_id", In: "path", Required: true, Schema: map[string]string{"type": "string", "format": "uuid"}},
				{Name: "doc_id", In: "path", Required: true, Schema: map[string]string{"type": "string", "format": "uuid"}},
				{Name: "user_id", In: "path", Required: true, Schema: map[string]string{"type": "string", "format": "uuid"}},
			},
			Responses: map[string]OpenAPIResponse{
				"204": {Description: "Permission revoked"},
			},
		},
	}
}
