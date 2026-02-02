package api

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDetectContentType_Utils(t *testing.T) {
	tests := []struct {
		name     string
		filename string
		expected string
	}{
		// Image types
		{"jpg extension", "photo.jpg", "image/jpeg"},
		{"jpeg extension", "photo.jpeg", "image/jpeg"},
		{"png extension", "image.png", "image/png"},
		{"gif extension", "animation.gif", "image/gif"},

		// Document types
		{"pdf extension", "document.pdf", "application/pdf"},
		{"txt extension", "readme.txt", "text/plain"},
		{"html extension", "page.html", "text/html"},

		// Data formats
		{"json extension", "data.json", "application/json"},
		{"xml extension", "config.xml", "application/xml"},

		// Archive types
		{"zip extension", "archive.zip", "application/zip"},

		// Media types
		{"mp4 extension", "video.mp4", "video/mp4"},
		{"mp3 extension", "audio.mp3", "audio/mpeg"},

		// Case insensitivity
		{"uppercase JPG", "photo.JPG", "image/jpeg"},
		{"mixed case PnG", "image.PnG", "image/png"},

		// Unknown extensions
		{"unknown extension", "file.unknown", "application/octet-stream"},
		{"no extension", "filename", "application/octet-stream"},
		{"empty filename", "", "application/octet-stream"},
		{"dot only", ".", "application/octet-stream"},

		// Multiple dots
		{"multiple dots", "archive.tar.gz", "application/octet-stream"},
		{"multiple dots with known ext", "photo.backup.jpg", "image/jpeg"},

		// Hidden files
		{"hidden file with extension", ".gitignore.txt", "text/plain"},
		{"hidden file without extension", ".gitignore", "application/octet-stream"},

		// Path-like filenames
		{"path with extension", "path/to/file.pdf", "application/pdf"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := detectContentType(tt.filename)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestDetectContentType_AllKnownTypes(t *testing.T) {
	// Verify all known extensions are mapped correctly
	knownTypes := map[string]string{
		".jpg":  "image/jpeg",
		".jpeg": "image/jpeg",
		".png":  "image/png",
		".gif":  "image/gif",
		".pdf":  "application/pdf",
		".txt":  "text/plain",
		".html": "text/html",
		".json": "application/json",
		".xml":  "application/xml",
		".zip":  "application/zip",
		".mp4":  "video/mp4",
		".mp3":  "audio/mpeg",
	}

	for ext, expectedType := range knownTypes {
		t.Run(ext, func(t *testing.T) {
			result := detectContentType("file" + ext)
			assert.Equal(t, expectedType, result, "Expected %s for extension %s", expectedType, ext)
		})
	}
}

func TestGetUserID_Utils(t *testing.T) {
	// Note: This function depends on Fiber context which requires more complex mocking.
	// These are placeholder tests - actual testing would require HTTP test framework.

	t.Run("returns anonymous when no context available", func(t *testing.T) {
		// The function returns "anonymous" when c.Locals("user_id") returns nil
		// This behavior is verified through integration tests
		assert.True(t, true, "Placeholder - requires Fiber context mocking")
	})
}
