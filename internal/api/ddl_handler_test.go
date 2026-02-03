package api

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// =============================================================================
// DDLHandler Construction Tests
// =============================================================================

func TestNewDDLHandler(t *testing.T) {
	t.Run("creates handler with nil database", func(t *testing.T) {
		handler := NewDDLHandler(nil)
		assert.NotNil(t, handler)
		assert.Nil(t, handler.db)
	})
}

// =============================================================================
// Identifier Validation Tests
// =============================================================================

func TestIdentifierPattern(t *testing.T) {
	validIdentifiers := []string{
		"users",
		"my_table",
		"MyTable",
		"_private",
		"table123",
		"a",
		"A",
		"_",
		"user_profiles",
		"CamelCase",
		"snake_case",
		"MixedCase_123",
	}

	invalidIdentifiers := []string{
		"123table",   // starts with number
		"my-table",   // contains hyphen
		"my table",   // contains space
		"my.table",   // contains dot
		"",           // empty
		"table!",     // contains special char
		"table@name", // contains @
		"table#1",    // contains #
		"select*",    // contains *
		"table;drop", // contains semicolon
		"table'name", // contains quote
		`table"name`, // contains double quote
	}

	for _, id := range validIdentifiers {
		t.Run("valid: "+id, func(t *testing.T) {
			assert.True(t, identifierPattern.MatchString(id), "Expected %q to be valid", id)
		})
	}

	for _, id := range invalidIdentifiers {
		t.Run("invalid: "+id, func(t *testing.T) {
			assert.False(t, identifierPattern.MatchString(id), "Expected %q to be invalid", id)
		})
	}
}

func TestReservedKeywords(t *testing.T) {
	t.Run("common SQL keywords are reserved", func(t *testing.T) {
		keywords := []string{
			"user", "table", "column", "index",
			"select", "insert", "update", "delete",
			"from", "where", "group", "order",
			"limit", "offset", "join", "on",
		}

		for _, kw := range keywords {
			assert.True(t, reservedKeywords[kw], "Expected %q to be reserved", kw)
		}
	})

	t.Run("non-reserved words not in map", func(t *testing.T) {
		nonReserved := []string{
			"users", "posts", "comments", "profiles",
			"custom_table", "my_column",
		}

		for _, word := range nonReserved {
			assert.False(t, reservedKeywords[word], "Expected %q to not be reserved", word)
		}
	})
}

func TestValidDataTypes(t *testing.T) {
	t.Run("text types", func(t *testing.T) {
		textTypes := []string{"text", "varchar", "char"}
		for _, dt := range textTypes {
			assert.True(t, validDataTypes[dt], "Expected %q to be valid", dt)
		}
	})

	t.Run("numeric types", func(t *testing.T) {
		numericTypes := []string{
			"integer", "bigint", "smallint",
			"numeric", "decimal", "real", "double precision",
		}
		for _, dt := range numericTypes {
			assert.True(t, validDataTypes[dt], "Expected %q to be valid", dt)
		}
	})

	t.Run("boolean types", func(t *testing.T) {
		boolTypes := []string{"boolean", "bool"}
		for _, dt := range boolTypes {
			assert.True(t, validDataTypes[dt], "Expected %q to be valid", dt)
		}
	})

	t.Run("date/time types", func(t *testing.T) {
		dateTypes := []string{
			"date", "timestamp", "timestamptz", "time", "timetz",
		}
		for _, dt := range dateTypes {
			assert.True(t, validDataTypes[dt], "Expected %q to be valid", dt)
		}
	})

	t.Run("other types", func(t *testing.T) {
		otherTypes := []string{
			"uuid", "json", "jsonb",
			"bytea", "inet", "cidr", "macaddr",
		}
		for _, dt := range otherTypes {
			assert.True(t, validDataTypes[dt], "Expected %q to be valid", dt)
		}
	})

	t.Run("invalid types not in map", func(t *testing.T) {
		invalidTypes := []string{
			"string", "int", "float", "datetime",
			"blob", "longtext", "tinyint",
		}
		for _, dt := range invalidTypes {
			assert.False(t, validDataTypes[dt], "Expected %q to be invalid", dt)
		}
	})
}

// =============================================================================
// CreateSchemaRequest Tests
// =============================================================================

func TestCreateSchemaRequest_Struct(t *testing.T) {
	t.Run("all fields accessible", func(t *testing.T) {
		req := CreateSchemaRequest{
			Name: "my_schema",
		}

		assert.Equal(t, "my_schema", req.Name)
	})

	t.Run("JSON deserialization", func(t *testing.T) {
		jsonData := `{"name":"test_schema"}`

		var req CreateSchemaRequest
		err := json.Unmarshal([]byte(jsonData), &req)
		require.NoError(t, err)

		assert.Equal(t, "test_schema", req.Name)
	})

	t.Run("empty name", func(t *testing.T) {
		req := CreateSchemaRequest{
			Name: "",
		}

		assert.Empty(t, req.Name)
	})
}

// =============================================================================
// CreateTableRequest Tests
// =============================================================================

func TestCreateTableRequest_Struct(t *testing.T) {
	t.Run("all fields accessible", func(t *testing.T) {
		req := CreateTableRequest{
			Schema: "public",
			Name:   "users",
			Columns: []CreateColumnRequest{
				{Name: "id", Type: "uuid", PrimaryKey: true},
				{Name: "name", Type: "text", Nullable: false},
			},
		}

		assert.Equal(t, "public", req.Schema)
		assert.Equal(t, "users", req.Name)
		assert.Len(t, req.Columns, 2)
	})

	t.Run("JSON deserialization", func(t *testing.T) {
		jsonData := `{
			"schema": "public",
			"name": "posts",
			"columns": [
				{"name": "id", "type": "uuid", "primaryKey": true},
				{"name": "title", "type": "text", "nullable": false},
				{"name": "content", "type": "text", "nullable": true}
			]
		}`

		var req CreateTableRequest
		err := json.Unmarshal([]byte(jsonData), &req)
		require.NoError(t, err)

		assert.Equal(t, "public", req.Schema)
		assert.Equal(t, "posts", req.Name)
		assert.Len(t, req.Columns, 3)
		assert.True(t, req.Columns[0].PrimaryKey)
	})

	t.Run("empty columns", func(t *testing.T) {
		req := CreateTableRequest{
			Schema:  "public",
			Name:    "empty_table",
			Columns: []CreateColumnRequest{},
		}

		assert.Empty(t, req.Columns)
	})
}

// =============================================================================
// CreateColumnRequest Tests
// =============================================================================

func TestCreateColumnRequest_Struct(t *testing.T) {
	t.Run("basic column", func(t *testing.T) {
		col := CreateColumnRequest{
			Name:     "id",
			Type:     "uuid",
			Nullable: false,
		}

		assert.Equal(t, "id", col.Name)
		assert.Equal(t, "uuid", col.Type)
		assert.False(t, col.Nullable)
	})

	t.Run("primary key column", func(t *testing.T) {
		col := CreateColumnRequest{
			Name:       "id",
			Type:       "integer",
			Nullable:   false,
			PrimaryKey: true,
		}

		assert.True(t, col.PrimaryKey)
		assert.False(t, col.Nullable)
	})

	t.Run("column with default value", func(t *testing.T) {
		col := CreateColumnRequest{
			Name:         "created_at",
			Type:         "timestamptz",
			Nullable:     false,
			DefaultValue: "NOW()",
		}

		assert.Equal(t, "NOW()", col.DefaultValue)
	})

	t.Run("nullable column", func(t *testing.T) {
		col := CreateColumnRequest{
			Name:     "description",
			Type:     "text",
			Nullable: true,
		}

		assert.True(t, col.Nullable)
	})

	t.Run("column with all options", func(t *testing.T) {
		col := CreateColumnRequest{
			Name:         "status",
			Type:         "text",
			Nullable:     false,
			PrimaryKey:   false,
			DefaultValue: "'pending'",
		}

		assert.Equal(t, "status", col.Name)
		assert.Equal(t, "text", col.Type)
		assert.False(t, col.Nullable)
		assert.False(t, col.PrimaryKey)
		assert.Equal(t, "'pending'", col.DefaultValue)
	})
}

// =============================================================================
// CreateSchema Handler Tests
// =============================================================================

func TestCreateSchema_Validation(t *testing.T) {
	t.Run("invalid request body", func(t *testing.T) {
		app := fiber.New()
		handler := NewDDLHandler(nil)

		app.Post("/schemas", handler.CreateSchema)

		req := httptest.NewRequest(http.MethodPost, "/schemas", bytes.NewReader([]byte("invalid json")))
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.Test(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)

		respBody, err := io.ReadAll(resp.Body)
		require.NoError(t, err)

		var result map[string]interface{}
		err = json.Unmarshal(respBody, &result)
		require.NoError(t, err)

		assert.Contains(t, result["error"], "Invalid request body")
	})

	t.Run("empty schema name", func(t *testing.T) {
		app := fiber.New()
		handler := NewDDLHandler(nil)

		app.Post("/schemas", handler.CreateSchema)

		body := `{"name":""}`
		req := httptest.NewRequest(http.MethodPost, "/schemas", bytes.NewReader([]byte(body)))
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.Test(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
	})
}

// =============================================================================
// CreateTable Handler Tests
// =============================================================================

func TestCreateTable_Validation(t *testing.T) {
	t.Run("invalid request body", func(t *testing.T) {
		app := fiber.New()
		handler := NewDDLHandler(nil)

		app.Post("/tables", handler.CreateTable)

		req := httptest.NewRequest(http.MethodPost, "/tables", bytes.NewReader([]byte("not json")))
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.Test(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
	})

	t.Run("missing columns", func(t *testing.T) {
		app := fiber.New()
		handler := NewDDLHandler(nil)

		app.Post("/tables", handler.CreateTable)

		body := `{"schema":"public","name":"test","columns":[]}`
		req := httptest.NewRequest(http.MethodPost, "/tables", bytes.NewReader([]byte(body)))
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.Test(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)

		respBody, err := io.ReadAll(resp.Body)
		require.NoError(t, err)

		var result map[string]interface{}
		err = json.Unmarshal(respBody, &result)
		require.NoError(t, err)

		assert.Contains(t, result["error"], "At least one column is required")
	})

	t.Run("invalid schema name", func(t *testing.T) {
		app := fiber.New()
		handler := NewDDLHandler(nil)

		app.Post("/tables", handler.CreateTable)

		body := `{"schema":"123invalid","name":"test","columns":[{"name":"id","type":"uuid"}]}`
		req := httptest.NewRequest(http.MethodPost, "/tables", bytes.NewReader([]byte(body)))
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.Test(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
	})

	t.Run("invalid table name", func(t *testing.T) {
		app := fiber.New()
		handler := NewDDLHandler(nil)

		app.Post("/tables", handler.CreateTable)

		body := `{"schema":"public","name":"my-table","columns":[{"name":"id","type":"uuid"}]}`
		req := httptest.NewRequest(http.MethodPost, "/tables", bytes.NewReader([]byte(body)))
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.Test(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
	})
}

// =============================================================================
// JSON Serialization Tests
// =============================================================================

func TestDDLRequests_JSONSerialization(t *testing.T) {
	t.Run("CreateSchemaRequest serializes correctly", func(t *testing.T) {
		req := CreateSchemaRequest{Name: "my_schema"}

		data, err := json.Marshal(req)
		require.NoError(t, err)

		assert.Contains(t, string(data), `"name":"my_schema"`)
	})

	t.Run("CreateTableRequest serializes correctly", func(t *testing.T) {
		req := CreateTableRequest{
			Schema: "public",
			Name:   "users",
			Columns: []CreateColumnRequest{
				{Name: "id", Type: "uuid", PrimaryKey: true},
			},
		}

		data, err := json.Marshal(req)
		require.NoError(t, err)

		assert.Contains(t, string(data), `"schema":"public"`)
		assert.Contains(t, string(data), `"name":"users"`)
		assert.Contains(t, string(data), `"columns"`)
	})

	t.Run("CreateColumnRequest serializes correctly", func(t *testing.T) {
		col := CreateColumnRequest{
			Name:         "created_at",
			Type:         "timestamptz",
			Nullable:     false,
			DefaultValue: "NOW()",
		}

		data, err := json.Marshal(col)
		require.NoError(t, err)

		assert.Contains(t, string(data), `"name":"created_at"`)
		assert.Contains(t, string(data), `"type":"timestamptz"`)
		assert.Contains(t, string(data), `"defaultValue":"NOW()"`)
	})
}

// =============================================================================
// Common Column Definitions Tests
// =============================================================================

func TestCommonColumnDefinitions(t *testing.T) {
	t.Run("UUID primary key column", func(t *testing.T) {
		col := CreateColumnRequest{
			Name:         "id",
			Type:         "uuid",
			Nullable:     false,
			PrimaryKey:   true,
			DefaultValue: "gen_random_uuid()",
		}

		assert.Equal(t, "uuid", col.Type)
		assert.True(t, col.PrimaryKey)
		assert.Equal(t, "gen_random_uuid()", col.DefaultValue)
	})

	t.Run("timestamp columns", func(t *testing.T) {
		createdAt := CreateColumnRequest{
			Name:         "created_at",
			Type:         "timestamptz",
			Nullable:     false,
			DefaultValue: "NOW()",
		}

		updatedAt := CreateColumnRequest{
			Name:         "updated_at",
			Type:         "timestamptz",
			Nullable:     false,
			DefaultValue: "NOW()",
		}

		assert.Equal(t, "timestamptz", createdAt.Type)
		assert.Equal(t, "timestamptz", updatedAt.Type)
	})

	t.Run("foreign key column", func(t *testing.T) {
		col := CreateColumnRequest{
			Name:     "user_id",
			Type:     "uuid",
			Nullable: false,
		}

		assert.Equal(t, "uuid", col.Type)
		assert.False(t, col.Nullable)
		assert.False(t, col.PrimaryKey)
	})

	t.Run("JSON column", func(t *testing.T) {
		col := CreateColumnRequest{
			Name:         "metadata",
			Type:         "jsonb",
			Nullable:     true,
			DefaultValue: "'{}'::jsonb",
		}

		assert.Equal(t, "jsonb", col.Type)
		assert.True(t, col.Nullable)
	})
}
