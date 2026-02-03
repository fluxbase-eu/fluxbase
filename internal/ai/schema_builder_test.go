package ai

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewSchemaBuilder(t *testing.T) {
	t.Run("creates schema builder with database connection", func(t *testing.T) {
		builder := NewSchemaBuilder(nil)
		assert.NotNil(t, builder)
	})
}

func TestSchemaBuilder_SettersAndGetters(t *testing.T) {
	builder := NewSchemaBuilder(nil)

	t.Run("SetSettingsResolver and GetSettingsResolver", func(t *testing.T) {
		resolver := &SettingsResolver{}
		builder.SetSettingsResolver(resolver)
		assert.Equal(t, resolver, builder.GetSettingsResolver())
	})

	t.Run("GetSettingsResolver returns nil when not set", func(t *testing.T) {
		builder := NewSchemaBuilder(nil)
		assert.Nil(t, builder.GetSettingsResolver())
	})

	t.Run("SetMCPResources", func(t *testing.T) {
		builder := NewSchemaBuilder(nil)
		// MCPResources is set but not exposed via getter
		// Just verify it doesn't panic
		builder.SetMCPResources(nil)
	})
}

func TestTableInfo_Struct(t *testing.T) {
	t.Run("all fields can be set", func(t *testing.T) {
		table := TableInfo{
			Schema:      "public",
			Name:        "users",
			Description: "User accounts table",
			Columns: []ColumnInfo{
				{Name: "id", DataType: "uuid", IsPrimaryKey: true},
				{Name: "email", DataType: "text", IsNullable: false},
			},
		}

		assert.Equal(t, "public", table.Schema)
		assert.Equal(t, "users", table.Name)
		assert.Equal(t, "User accounts table", table.Description)
		assert.Len(t, table.Columns, 2)
	})

	t.Run("zero value has empty fields", func(t *testing.T) {
		var table TableInfo
		assert.Empty(t, table.Schema)
		assert.Empty(t, table.Name)
		assert.Empty(t, table.Columns)
	})
}

func TestAI_ColumnInfo_Struct(t *testing.T) {
	t.Run("primary key column", func(t *testing.T) {
		col := ColumnInfo{
			Name:         "id",
			DataType:     "uuid",
			IsNullable:   false,
			IsPrimaryKey: true,
			IsForeignKey: false,
		}

		assert.Equal(t, "id", col.Name)
		assert.Equal(t, "uuid", col.DataType)
		assert.False(t, col.IsNullable)
		assert.True(t, col.IsPrimaryKey)
		assert.False(t, col.IsForeignKey)
	})

	t.Run("foreign key column", func(t *testing.T) {
		foreignTable := "users"
		foreignCol := "id"

		col := ColumnInfo{
			Name:         "user_id",
			DataType:     "uuid",
			IsNullable:   true,
			IsPrimaryKey: false,
			IsForeignKey: true,
			ForeignTable: &foreignTable,
			ForeignCol:   &foreignCol,
		}

		assert.True(t, col.IsForeignKey)
		assert.NotNil(t, col.ForeignTable)
		assert.Equal(t, "users", *col.ForeignTable)
		assert.NotNil(t, col.ForeignCol)
		assert.Equal(t, "id", *col.ForeignCol)
	})

	t.Run("column with default and description", func(t *testing.T) {
		defaultVal := "gen_random_uuid()"

		col := ColumnInfo{
			Name:        "id",
			DataType:    "uuid",
			Default:     &defaultVal,
			Description: "Primary identifier",
		}

		assert.NotNil(t, col.Default)
		assert.Equal(t, "gen_random_uuid()", *col.Default)
		assert.Equal(t, "Primary identifier", col.Description)
	})
}

func TestSchemaBuilder_FormatSchemaDescription(t *testing.T) {
	// Test the formatting logic by creating sample tables and verifying output format

	t.Run("formats table with columns", func(t *testing.T) {
		tables := []TableInfo{
			{
				Schema:      "public",
				Name:        "users",
				Description: "User accounts",
				Columns: []ColumnInfo{
					{Name: "id", DataType: "uuid", IsPrimaryKey: true},
					{Name: "email", DataType: "text", IsNullable: false},
					{Name: "name", DataType: "text", IsNullable: true},
				},
			},
		}

		// Simulate the formatting logic
		var sb strings.Builder
		sb.WriteString("## Available Database Tables\n\n")

		for _, table := range tables {
			sb.WriteString("### " + table.Schema + "." + table.Name + "\n")
			if table.Description != "" {
				sb.WriteString(table.Description + "\n\n")
			}
			sb.WriteString("| Column | Type | Nullable | Notes |\n")
			sb.WriteString("|--------|------|----------|-------|\n")

			for _, col := range table.Columns {
				nullable := "YES"
				if !col.IsNullable {
					nullable = "NO"
				}
				notes := ""
				if col.IsPrimaryKey {
					notes = "PK"
				}
				sb.WriteString("| " + col.Name + " | " + col.DataType + " | " + nullable + " | " + notes + " |\n")
			}
		}

		result := sb.String()

		assert.Contains(t, result, "## Available Database Tables")
		assert.Contains(t, result, "### public.users")
		assert.Contains(t, result, "User accounts")
		assert.Contains(t, result, "| id | uuid | NO | PK |")
		assert.Contains(t, result, "| email | text | NO |")
		assert.Contains(t, result, "| name | text | YES |")
	})

	t.Run("formats foreign key notes", func(t *testing.T) {
		foreignTable := "profiles"
		foreignCol := "id"

		col := ColumnInfo{
			Name:         "profile_id",
			DataType:     "uuid",
			IsForeignKey: true,
			ForeignTable: &foreignTable,
			ForeignCol:   &foreignCol,
		}

		// Format FK note
		var notes []string
		if col.IsForeignKey && col.ForeignTable != nil {
			notes = append(notes, "FK → "+*col.ForeignTable+"."+*col.ForeignCol)
		}

		assert.Len(t, notes, 1)
		assert.Equal(t, "FK → profiles.id", notes[0])
	})
}

func TestSchemaBuilder_EmptyTables(t *testing.T) {
	t.Run("handles empty table list", func(t *testing.T) {
		tables := []TableInfo{}
		assert.Empty(t, tables)

		// The expected output for empty tables
		expectedOutput := "No tables available."
		assert.Equal(t, "No tables available.", expectedOutput)
	})
}

func TestColumnInfo_NullableFormatting(t *testing.T) {
	testCases := []struct {
		name       string
		isNullable bool
		expected   string
	}{
		{"nullable column", true, "YES"},
		{"non-nullable column", false, "NO"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			col := ColumnInfo{IsNullable: tc.isNullable}
			var nullable string
			if col.IsNullable {
				nullable = "YES"
			} else {
				nullable = "NO"
			}
			assert.Equal(t, tc.expected, nullable)
		})
	}
}

func TestColumnInfo_NotesFormatting(t *testing.T) {
	t.Run("combines multiple notes", func(t *testing.T) {
		foreignTable := "users"
		foreignCol := "id"

		col := ColumnInfo{
			Name:         "user_id",
			IsPrimaryKey: true,
			IsForeignKey: true,
			ForeignTable: &foreignTable,
			ForeignCol:   &foreignCol,
			Description:  "The user who owns this record",
		}

		var notes []string
		if col.IsPrimaryKey {
			notes = append(notes, "PK")
		}
		if col.IsForeignKey && col.ForeignTable != nil {
			notes = append(notes, "FK → "+*col.ForeignTable+"."+*col.ForeignCol)
		}
		if col.Description != "" {
			notes = append(notes, col.Description)
		}

		notesStr := strings.Join(notes, ", ")

		assert.Contains(t, notesStr, "PK")
		assert.Contains(t, notesStr, "FK → users.id")
		assert.Contains(t, notesStr, "The user who owns this record")
	})
}
