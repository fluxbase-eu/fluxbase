package ai

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestMetadataToJSON_EmptyMap(t *testing.T) {
	result, err := metadataToJSON(map[string]interface{}{})
	require.NoError(t, err)
	require.Equal(t, "{}", string(result))
}

func TestMetadataToJSON_NilMap(t *testing.T) {
	result, err := metadataToJSON(nil)
	require.NoError(t, err)
	require.Equal(t, "{}", string(result))
}

func TestMetadataToJSON_BasicTypes(t *testing.T) {
	m := map[string]interface{}{
		"string":  "value",
		"int":     42,
		"bool":    true,
		"strings": []string{"a", "b", "c"},
	}

	result, err := metadataToJSON(m)
	require.NoError(t, err)

	// Verify it's valid JSON by unmarshaling
	var parsed map[string]interface{}
	require.NoError(t, json.Unmarshal(result, &parsed))

	require.Equal(t, "value", parsed["string"])
	require.Equal(t, float64(42), parsed["int"]) // JSON numbers are float64
	require.Equal(t, true, parsed["bool"])
}

func TestMetadataToJSON_SpecialCharacters(t *testing.T) {
	tests := []struct {
		name  string
		input map[string]interface{}
	}{
		{
			name: "quotes",
			input: map[string]interface{}{
				"value": `string with "quotes" inside`,
			},
		},
		{
			name: "backslashes",
			input: map[string]interface{}{
				"path": `c:\path\to\file`,
			},
		},
		{
			name: "newlines",
			input: map[string]interface{}{
				"multiline": "line1\nline2\rline3",
			},
		},
		{
			name: "unicode",
			input: map[string]interface{}{
				"emoji": "emoji 🎉 test",
			},
		},
		{
			name: "mixed special chars",
			input: map[string]interface{}{
				"quote":     `he said "hello"`,
				"backslash": `c:\test\path`,
				"newline":   "line1\nline2",
				"tab":       "col1\tcol2",
				"unicode":   "日本語 🎉",
			},
		},
		{
			name: "table type values",
			input: map[string]interface{}{
				"table_type": "BASE TABLE",
				"schema":     "public",
				"table":      "place_visits",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := metadataToJSON(tt.input)
			require.NoError(t, err, "metadataToJSON should not error on %s", tt.name)

			// Most important check: PostgreSQL would accept this as valid JSON
			var parsed map[string]interface{}
			err = json.Unmarshal(result, &parsed)
			require.NoError(t, err, "Result should be valid JSON: %s", string(result))

			// Verify all keys are present
			for k := range tt.input {
				_, exists := parsed[k]
				require.True(t, exists, "Key %s should exist in parsed result", k)
			}
		})
	}
}

func TestMetadataToJSON_TableExporterMetadata(t *testing.T) {
	// Simulate the exact metadata structure used in ExportTable
	m := map[string]interface{}{
		"schema":           "public",
		"table":            "place_visits",
		"entity_type":      "table",
		"source":           "database_export",
		"table_type":       "BASE TABLE",
		"rls_enabled":      false,
		"exported_columns": 5,
		"total_columns":    10,
		"columns_filtered": true,
		"columns":          []string{"id", "name", "created_at"},
	}

	result, err := metadataToJSON(m)
	require.NoError(t, err)

	// Verify it's valid JSON
	var parsed map[string]interface{}
	require.NoError(t, json.Unmarshal(result, &parsed))

	// Verify key fields
	require.Equal(t, "public", parsed["schema"])
	require.Equal(t, "place_visits", parsed["table"])
	require.Equal(t, "BASE TABLE", parsed["table_type"])
}
