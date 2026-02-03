package ai

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestOCRProviderType_Constants(t *testing.T) {
	t.Run("tesseract type is defined", func(t *testing.T) {
		assert.Equal(t, OCRProviderType("tesseract"), OCRProviderTypeTesseract)
	})
}

func TestOCRResult_Struct(t *testing.T) {
	t.Run("all fields can be set", func(t *testing.T) {
		result := OCRResult{
			Text:       "Extracted text content",
			Confidence: 0.95,
			Pages:      3,
			Language:   "eng",
		}

		assert.Equal(t, "Extracted text content", result.Text)
		assert.Equal(t, 0.95, result.Confidence)
		assert.Equal(t, 3, result.Pages)
		assert.Equal(t, "eng", result.Language)
	})

	t.Run("zero value has expected defaults", func(t *testing.T) {
		var result OCRResult
		assert.Empty(t, result.Text)
		assert.Zero(t, result.Confidence)
		assert.Zero(t, result.Pages)
		assert.Empty(t, result.Language)
	})
}

func TestOCRProviderConfig_Struct(t *testing.T) {
	t.Run("all fields can be set", func(t *testing.T) {
		cfg := OCRProviderConfig{
			Type:      OCRProviderTypeTesseract,
			Languages: []string{"eng", "deu", "fra"},
		}

		assert.Equal(t, OCRProviderTypeTesseract, cfg.Type)
		assert.Equal(t, []string{"eng", "deu", "fra"}, cfg.Languages)
	})
}

func TestNewOCRProvider(t *testing.T) {
	// Note: These tests may fail if Tesseract is not installed
	// They test the factory function behavior

	t.Run("creates provider for tesseract type", func(t *testing.T) {
		cfg := OCRProviderConfig{
			Type:      OCRProviderTypeTesseract,
			Languages: []string{"eng"},
		}

		provider, err := NewOCRProvider(cfg)
		// Provider should be created even if Tesseract is not installed
		// It will just report IsAvailable() = false
		assert.NoError(t, err)
		assert.NotNil(t, provider)
	})

	t.Run("defaults to tesseract for unknown type", func(t *testing.T) {
		cfg := OCRProviderConfig{
			Type:      OCRProviderType("unknown"),
			Languages: []string{"eng"},
		}

		provider, err := NewOCRProvider(cfg)
		assert.NoError(t, err)
		assert.NotNil(t, provider)
	})

	t.Run("defaults to tesseract for empty type", func(t *testing.T) {
		cfg := OCRProviderConfig{
			Type:      "",
			Languages: []string{"eng"},
		}

		provider, err := NewOCRProvider(cfg)
		assert.NoError(t, err)
		assert.NotNil(t, provider)
	})
}
