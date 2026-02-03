package adminui

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNew(t *testing.T) {
	t.Run("creates handler with public base URL", func(t *testing.T) {
		handler := New("http://localhost:8080")
		assert.NotNil(t, handler)
		assert.Equal(t, "http://localhost:8080", handler.config.PublicBaseURL)
		assert.NotEmpty(t, handler.configScript)
	})

	t.Run("creates handler with empty URL", func(t *testing.T) {
		handler := New("")
		assert.NotNil(t, handler)
		assert.Empty(t, handler.config.PublicBaseURL)
	})

	t.Run("config script contains public URL", func(t *testing.T) {
		handler := New("https://api.example.com")
		scriptStr := string(handler.configScript)
		assert.Contains(t, scriptStr, "__FLUXBASE_CONFIG__")
		assert.Contains(t, scriptStr, "https://api.example.com")
	})
}

func TestConfig_Struct(t *testing.T) {
	t.Run("JSON serialization", func(t *testing.T) {
		cfg := Config{
			PublicBaseURL: "http://localhost:3000",
		}

		data, err := json.Marshal(cfg)
		require.NoError(t, err)
		assert.Contains(t, string(data), `"publicBaseURL":"http://localhost:3000"`)
	})

	t.Run("JSON deserialization", func(t *testing.T) {
		jsonData := `{"publicBaseURL": "https://api.example.com"}`

		var cfg Config
		err := json.Unmarshal([]byte(jsonData), &cfg)
		require.NoError(t, err)
		assert.Equal(t, "https://api.example.com", cfg.PublicBaseURL)
	})
}

func TestHandler_InjectConfig(t *testing.T) {
	handler := New("http://localhost:8080")

	t.Run("injects config before </head>", func(t *testing.T) {
		html := []byte(`<!DOCTYPE html><html><head><title>Test</title></head><body></body></html>`)

		result := handler.injectConfig(html)

		resultStr := string(result)
		assert.Contains(t, resultStr, "__FLUXBASE_CONFIG__")
		assert.Contains(t, resultStr, "</head>")
		// Config should appear before </head>
		configIdx := findIndex(result, []byte("__FLUXBASE_CONFIG__"))
		headIdx := findIndex(result, []byte("</head>"))
		assert.Less(t, configIdx, headIdx)
	})

	t.Run("handles missing </head>", func(t *testing.T) {
		html := []byte(`<html><body>No head tag</body></html>`)

		result := handler.injectConfig(html)

		// Should return content unchanged
		assert.Equal(t, html, result)
	})

	t.Run("handles empty content", func(t *testing.T) {
		html := []byte(``)

		result := handler.injectConfig(html)

		assert.Empty(t, result)
	})

	t.Run("preserves content after </head>", func(t *testing.T) {
		html := []byte(`<head></head><body><p>Content</p></body>`)

		result := handler.injectConfig(html)

		resultStr := string(result)
		assert.Contains(t, resultStr, "<body><p>Content</p></body>")
	})
}

func TestGetContentType(t *testing.T) {
	tests := []struct {
		name     string
		path     string
		expected string
	}{
		{
			name:     "JavaScript file",
			path:     "/assets/app.js",
			expected: "application/javascript",
		},
		{
			name:     "CSS file",
			path:     "/assets/style.css",
			expected: "text/css",
		},
		{
			name:     "JSON file",
			path:     "/config.json",
			expected: "application/json",
		},
		{
			name:     "PNG image",
			path:     "/images/logo.png",
			expected: "image/png",
		},
		{
			name:     "JPG image",
			path:     "/images/photo.jpg",
			expected: "image/jpeg",
		},
		{
			name:     "JPEG image",
			path:     "/images/photo.jpeg",
			expected: "image/jpeg",
		},
		{
			name:     "SVG image",
			path:     "/icons/icon.svg",
			expected: "image/svg+xml",
		},
		{
			name:     "WOFF font",
			path:     "/fonts/font.woff",
			expected: "font/woff",
		},
		{
			name:     "WOFF2 font",
			path:     "/fonts/font.woff2",
			expected: "font/woff2",
		},
		{
			name:     "TTF font",
			path:     "/fonts/font.ttf",
			expected: "font/ttf",
		},
		{
			name:     "HTML file",
			path:     "/index.html",
			expected: "text/html",
		},
		{
			name:     "unknown extension",
			path:     "/file.xyz",
			expected: "application/octet-stream",
		},
		{
			name:     "no extension",
			path:     "/file",
			expected: "application/octet-stream",
		},
		{
			name:     "path with multiple dots",
			path:     "/app.bundle.min.js",
			expected: "application/javascript",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := getContentType(tt.path)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// Helper function to find index of a byte slice within another
func findIndex(data, pattern []byte) int {
	for i := 0; i <= len(data)-len(pattern); i++ {
		match := true
		for j := 0; j < len(pattern); j++ {
			if data[i+j] != pattern[j] {
				match = false
				break
			}
		}
		if match {
			return i
		}
	}
	return -1
}
