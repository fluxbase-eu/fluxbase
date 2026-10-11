package api

// addStorageEndpoints adds storage API endpoints to the spec
func (h *OpenAPIHandler) addStorageEndpoints(spec *OpenAPISpec) {
	// Storage object schema — mirrors the object items returned by
	// ListFiles/GetFileInfo (storage_files.go).
	spec.Components.Schemas["StorageObject"] = map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"id":         map[string]string{"type": "string"},
			"bucket":     map[string]string{"type": "string"},
			"path":       map[string]string{"type": "string"},
			"mime_type":  map[string]string{"type": "string"},
			"size":       map[string]string{"type": "integer"},
			"metadata":   map[string]interface{}{"type": "object", "additionalProperties": map[string]interface{}{}},
			"owner_id":   map[string]string{"type": "string"},
			"created_at": map[string]string{"type": "string", "format": "date-time"},
			"updated_at": map[string]string{"type": "string", "format": "date-time"},
		},
		"required": []string{"id", "bucket", "path", "size", "created_at", "updated_at"},
	}

	// Upload response schema — mirrors the JSON returned by UploadFile
	// (storage_files.go): the server has no object id/etag.
	spec.Components.Schemas["UploadFileResponse"] = map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"key":           map[string]string{"type": "string"},
			"bucket":        map[string]string{"type": "string"},
			"size":          map[string]string{"type": "integer"},
			"content_type":  map[string]string{"type": "string"},
			"last_modified": map[string]string{"type": "string", "format": "date-time"},
			"owner_id":      map[string]string{"type": "string"},
		},
		"required": []string{"key", "bucket", "size", "content_type", "last_modified"},
	}

	// Bucket schema
	spec.Components.Schemas["Bucket"] = map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"name":         map[string]string{"type": "string"},
			"created_date": map[string]string{"type": "string", "format": "date-time"},
		},
	}

	// GET /api/v1/storage/buckets - List all buckets
	spec.Paths["/api/v1/storage/buckets"] = OpenAPIPath{
		"get": OpenAPIOperation{
			Summary:     "List all storage buckets",
			Description: "Retrieve a list of all available storage buckets",
			OperationID: "list_buckets",
			Tags:        []string{"Storage"},
			Responses: map[string]OpenAPIResponse{
				"200": {
					Description: "List of buckets",
					Content: map[string]OpenAPIMedia{
						"application/json": {
							Schema: map[string]interface{}{
								"type": "object",
								"properties": map[string]interface{}{
									"buckets": map[string]interface{}{
										"type":  "array",
										"items": map[string]string{"$ref": "#/components/schemas/Bucket"},
									},
								},
							},
						},
					},
				},
			},
		},
	}

	// POST /api/v1/storage/buckets/:bucket - Create bucket
	spec.Paths["/api/v1/storage/buckets/{bucket}"] = OpenAPIPath{
		"post": OpenAPIOperation{
			Summary:     "Create a new storage bucket",
			Description: "Create a new bucket for storing files",
			OperationID: "create_bucket",
			Tags:        []string{"Storage"},
			Parameters: []OpenAPIParameter{
				{
					Name:        "bucket",
					In:          "path",
					Description: "Bucket name",
					Required:    true,
					Schema:      map[string]string{"type": "string"},
				},
			},
			Responses: map[string]OpenAPIResponse{
				"201": {
					Description: "Bucket created",
					Content: map[string]OpenAPIMedia{
						"application/json": {
							Schema: map[string]interface{}{
								"type": "object",
								"properties": map[string]interface{}{
									"bucket":  map[string]string{"type": "string"},
									"message": map[string]string{"type": "string"},
								},
							},
						},
					},
				},
				"409": {
					Description: "Bucket already exists",
					Content: map[string]OpenAPIMedia{
						"application/json": {
							Schema: map[string]string{"$ref": "#/components/schemas/Error"},
						},
					},
				},
			},
		},
		"delete": OpenAPIOperation{
			Summary:     "Delete a storage bucket",
			Description: "Delete an empty bucket",
			OperationID: "delete_bucket",
			Tags:        []string{"Storage"},
			Parameters: []OpenAPIParameter{
				{
					Name:        "bucket",
					In:          "path",
					Description: "Bucket name",
					Required:    true,
					Schema:      map[string]string{"type": "string"},
				},
			},
			Responses: map[string]OpenAPIResponse{
				"204": {
					Description: "Bucket deleted",
				},
				"404": {
					Description: "Bucket not found",
					Content: map[string]OpenAPIMedia{
						"application/json": {
							Schema: map[string]string{"$ref": "#/components/schemas/Error"},
						},
					},
				},
				"409": {
					Description: "Bucket is not empty",
					Content: map[string]OpenAPIMedia{
						"application/json": {
							Schema: map[string]string{"$ref": "#/components/schemas/Error"},
						},
					},
				},
			},
		},
	}

	// GET /api/v1/storage/:bucket - List files in bucket
	spec.Paths["/api/v1/storage/{bucket}"] = OpenAPIPath{
		"get": OpenAPIOperation{
			Summary:     "List files in bucket",
			Description: "List all files in a specific bucket with optional filtering",
			OperationID: "list_files",
			Tags:        []string{"Storage"},
			Parameters: []OpenAPIParameter{
				{
					Name:        "bucket",
					In:          "path",
					Description: "Bucket name",
					Required:    true,
					Schema:      map[string]string{"type": "string"},
				},
				{
					Name:        "prefix",
					In:          "query",
					Description: "Filter files by prefix",
					Required:    false,
					Schema:      map[string]string{"type": "string"},
				},
				{
					Name:        "delimiter",
					In:          "query",
					Description: "Delimiter for grouping",
					Required:    false,
					Schema:      map[string]string{"type": "string"},
				},
				{
					Name:        "limit",
					In:          "query",
					Description: "Maximum number of files to return",
					Required:    false,
					Schema:      map[string]string{"type": "integer", "default": "1000"},
				},
				{
					Name:        "offset",
					In:          "query",
					Description: "Number of files to skip",
					Required:    false,
					Schema:      map[string]string{"type": "integer", "default": "0"},
				},
			},
			Responses: map[string]OpenAPIResponse{
				"200": {
					Description: "List of files",
					Content: map[string]OpenAPIMedia{
						"application/json": {
							Schema: map[string]interface{}{
								"type": "object",
								"properties": map[string]interface{}{
									"bucket": map[string]string{"type": "string"},
									"objects": map[string]interface{}{
										"type":  "array",
										"items": map[string]string{"$ref": "#/components/schemas/StorageObject"},
									},
									"count": map[string]string{"type": "integer"},
									"prefix": map[string]interface{}{
										"type":        "string",
										"description": "Present when the delimiter query parameter is used",
									},
									"prefixes": map[string]interface{}{
										"type":        "array",
										"items":       map[string]string{"type": "string"},
										"description": "Present when the delimiter query parameter is used",
									},
								},
								"required": []string{"bucket", "objects", "count"},
							},
						},
					},
				},
			},
		},
	}

	// File operations: Upload, Download, Delete, Get Info
	spec.Paths["/api/v1/storage/{bucket}/{key}"] = OpenAPIPath{
		"post": OpenAPIOperation{
			Summary:     "Upload a file",
			Description: "Upload a file to the specified bucket and key",
			OperationID: "upload_file",
			Tags:        []string{"Storage"},
			Parameters: []OpenAPIParameter{
				{
					Name:        "bucket",
					In:          "path",
					Description: "Bucket name",
					Required:    true,
					Schema:      map[string]string{"type": "string"},
				},
				{
					Name:        "key",
					In:          "path",
					Description: "File key (path)",
					Required:    true,
					Schema:      map[string]string{"type": "string"},
				},
			},
			RequestBody: &OpenAPIRequestBody{
				Required:    true,
				Description: "File to upload. Custom metadata is sent as `metadata_<key>` form fields (the only metadata form the server reads).",
				Content: map[string]OpenAPIMedia{
					"multipart/form-data": {
						Schema: map[string]interface{}{
							"type": "object",
							"properties": map[string]interface{}{
								"file": map[string]string{
									"type":   "string",
									"format": "binary",
								},
								"metadata_<key>": map[string]interface{}{
									"type":        "string",
									"description": "Custom metadata entry; every form field prefixed with metadata_ is stored as object metadata (key = field name without the prefix).",
								},
							},
							"required": []string{"file"},
						},
					},
				},
			},
			Responses: map[string]OpenAPIResponse{
				"201": {
					Description: "File uploaded",
					Content: map[string]OpenAPIMedia{
						"application/json": {
							Schema: map[string]string{"$ref": "#/components/schemas/UploadFileResponse"},
						},
					},
				},
			},
		},
		"get": OpenAPIOperation{
			Summary:     "Download a file",
			Description: "Download a file from storage",
			OperationID: "download_file",
			Tags:        []string{"Storage"},
			Parameters: []OpenAPIParameter{
				{
					Name:        "bucket",
					In:          "path",
					Description: "Bucket name",
					Required:    true,
					Schema:      map[string]string{"type": "string"},
				},
				{
					Name:        "key",
					In:          "path",
					Description: "File key (path)",
					Required:    true,
					Schema:      map[string]string{"type": "string"},
				},
				{
					Name:        "download",
					In:          "query",
					Description: "Force download (set Content-Disposition header)",
					Required:    false,
					Schema:      map[string]string{"type": "boolean"},
				},
			},
			Responses: map[string]OpenAPIResponse{
				"200": {
					Description: "File content",
					Content: map[string]OpenAPIMedia{
						"application/octet-stream": {
							Schema: map[string]string{
								"type":   "string",
								"format": "binary",
							},
						},
					},
				},
				"404": {
					Description: "File not found",
					Content: map[string]OpenAPIMedia{
						"application/json": {
							Schema: map[string]string{"$ref": "#/components/schemas/Error"},
						},
					},
				},
			},
		},
		"head": OpenAPIOperation{
			Summary:     "Get file metadata",
			Description: "Get metadata about a file without downloading it",
			OperationID: "get_file_info",
			Tags:        []string{"Storage"},
			Parameters: []OpenAPIParameter{
				{
					Name:        "bucket",
					In:          "path",
					Description: "Bucket name",
					Required:    true,
					Schema:      map[string]string{"type": "string"},
				},
				{
					Name:        "key",
					In:          "path",
					Description: "File key (path)",
					Required:    true,
					Schema:      map[string]string{"type": "string"},
				},
			},
			Responses: map[string]OpenAPIResponse{
				"200": {
					Description: "File metadata",
					Content: map[string]OpenAPIMedia{
						"application/json": {
							Schema: map[string]string{"$ref": "#/components/schemas/StorageObject"},
						},
					},
				},
				"404": {
					Description: "File not found",
					Content: map[string]OpenAPIMedia{
						"application/json": {
							Schema: map[string]string{"$ref": "#/components/schemas/Error"},
						},
					},
				},
			},
		},
		"delete": OpenAPIOperation{
			Summary:     "Delete a file",
			Description: "Delete a file from storage",
			OperationID: "delete_file",
			Tags:        []string{"Storage"},
			Parameters: []OpenAPIParameter{
				{
					Name:        "bucket",
					In:          "path",
					Description: "Bucket name",
					Required:    true,
					Schema:      map[string]string{"type": "string"},
				},
				{
					Name:        "key",
					In:          "path",
					Description: "File key (path)",
					Required:    true,
					Schema:      map[string]string{"type": "string"},
				},
			},
			Responses: map[string]OpenAPIResponse{
				"204": {
					Description: "File deleted",
				},
				"404": {
					Description: "File not found",
					Content: map[string]OpenAPIMedia{
						"application/json": {
							Schema: map[string]string{"$ref": "#/components/schemas/Error"},
						},
					},
				},
			},
		},
	}

	// POST /api/v1/storage/:bucket/sign/* - Generate signed URL
	spec.Paths["/api/v1/storage/{bucket}/sign/{key}"] = OpenAPIPath{
		"post": OpenAPIOperation{
			Summary:     "Generate signed URL",
			Description: "Generate a presigned URL for temporary file access. The caller must have RLS visibility of the object; lifetimes are capped at 24 hours.",
			OperationID: "generate_signed_url",
			Tags:        []string{"Storage"},
			Parameters: []OpenAPIParameter{
				{
					Name:        "bucket",
					In:          "path",
					Description: "Bucket name",
					Required:    true,
					Schema:      map[string]string{"type": "string"},
				},
				{
					Name:        "key",
					In:          "path",
					Description: "File key (path)",
					Required:    true,
					Schema:      map[string]string{"type": "string"},
				},
			},
			RequestBody: &OpenAPIRequestBody{
				Description: "Signed URL options",
				Content: map[string]OpenAPIMedia{
					"application/json": {
						Schema: map[string]interface{}{
							"type": "object",
							"properties": map[string]interface{}{
								"expires_in": map[string]string{
									"type":        "integer",
									"description": "URL expiration time in seconds (default: 900)",
								},
								"method": map[string]interface{}{
									"type":        "string",
									"enum":        []string{"GET", "PUT", "DELETE"},
									"description": "HTTP method for the signed URL (default: GET)",
								},
							},
						},
					},
				},
			},
			Responses: map[string]OpenAPIResponse{
				"200": {
					Description: "Signed URL generated",
					Content: map[string]OpenAPIMedia{
						"application/json": {
							Schema: map[string]interface{}{
								"type": "object",
								"properties": map[string]interface{}{
									"signed_url": map[string]string{"type": "string"},
									"expires_in": map[string]string{"type": "integer"},
									"method":     map[string]string{"type": "string"},
								},
							},
						},
					},
				},
				"403": {
					Description: "Caller lacks RLS visibility of the object",
					Content: map[string]OpenAPIMedia{
						"application/json": {
							Schema: map[string]string{"$ref": "#/components/schemas/Error"},
						},
					},
				},
				"404": {
					Description: "Object not found or not visible",
					Content: map[string]OpenAPIMedia{
						"application/json": {
							Schema: map[string]string{"$ref": "#/components/schemas/Error"},
						},
					},
				},
				"400": {
					Description: "expires_in exceeds the 24h maximum or unsupported method",
					Content: map[string]OpenAPIMedia{
						"application/json": {
							Schema: map[string]string{"$ref": "#/components/schemas/Error"},
						},
					},
				},
			},
		},
	}
}
