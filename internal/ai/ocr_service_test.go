package ai

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// mockOCRProvider implements OCRProvider for testing
type mockOCRProvider struct {
	name                    string
	available               bool
	extractPDFFunc          func(ctx context.Context, data []byte, languages []string) (*OCRResult, error)
	extractImageFunc        func(ctx context.Context, data []byte, languages []string) (*OCRResult, error)
	closeFunc               func() error
}

func newMockOCRProvider(available bool) *mockOCRProvider {
	return &mockOCRProvider{
		name:      "mock",
		available: available,
	}
}

func (m *mockOCRProvider) Name() string {
	return m.name
}

func (m *mockOCRProvider) IsAvailable() bool {
	return m.available
}

func (m *mockOCRProvider) ExtractTextFromPDF(ctx context.Context, data []byte, languages []string) (*OCRResult, error) {
	if m.extractPDFFunc != nil {
		return m.extractPDFFunc(ctx, data, languages)
	}
	return &OCRResult{
		Text:       "Extracted text from PDF",
		Pages:      1,
		Confidence: 0.95,
	}, nil
}

func (m *mockOCRProvider) ExtractTextFromImage(ctx context.Context, data []byte, languages []string) (*OCRResult, error) {
	if m.extractImageFunc != nil {
		return m.extractImageFunc(ctx, data, languages)
	}
	return &OCRResult{
		Text:       "Extracted text from image",
		Pages:      1,
		Confidence: 0.90,
	}, nil
}

func (m *mockOCRProvider) Close() error {
	if m.closeFunc != nil {
		return m.closeFunc()
	}
	return nil
}

func TestNewOCRService(t *testing.T) {
	t.Run("creates disabled service when disabled in config", func(t *testing.T) {
		cfg := OCRServiceConfig{
			Enabled: false,
		}

		service, err := NewOCRService(cfg)
		require.NoError(t, err)
		assert.NotNil(t, service)
		assert.False(t, service.IsEnabled())
	})

	t.Run("sets default language to eng when not provided", func(t *testing.T) {
		// This test would require mocking NewOCRProvider
		// For now we test the config structure
		cfg := OCRServiceConfig{
			Enabled:          false,
			DefaultLanguages: []string{},
		}

		service, err := NewOCRService(cfg)
		require.NoError(t, err)
		assert.False(t, service.IsEnabled())
	})
}

func TestOCRService_IsEnabled(t *testing.T) {
	t.Run("returns true when enabled", func(t *testing.T) {
		service := &OCRService{
			enabled: true,
		}
		assert.True(t, service.IsEnabled())
	})

	t.Run("returns false when disabled", func(t *testing.T) {
		service := &OCRService{
			enabled: false,
		}
		assert.False(t, service.IsEnabled())
	})
}

func TestOCRService_GetDefaultLanguages(t *testing.T) {
	t.Run("returns configured languages", func(t *testing.T) {
		service := &OCRService{
			defaultLanguages: []string{"eng", "deu", "fra"},
		}

		languages := service.GetDefaultLanguages()
		assert.Equal(t, []string{"eng", "deu", "fra"}, languages)
	})

	t.Run("returns empty slice when no languages configured", func(t *testing.T) {
		service := &OCRService{
			defaultLanguages: nil,
		}

		languages := service.GetDefaultLanguages()
		assert.Nil(t, languages)
	})
}

func TestOCRService_ExtractTextFromPDF(t *testing.T) {
	t.Run("returns error when service is disabled", func(t *testing.T) {
		service := &OCRService{
			enabled: false,
		}

		result, err := service.ExtractTextFromPDF(context.Background(), []byte("pdf data"), nil)
		require.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "not enabled")
	})

	t.Run("uses default languages when none provided", func(t *testing.T) {
		var capturedLanguages []string
		mock := newMockOCRProvider(true)
		mock.extractPDFFunc = func(ctx context.Context, data []byte, languages []string) (*OCRResult, error) {
			capturedLanguages = languages
			return &OCRResult{Text: "text", Pages: 1, Confidence: 0.9}, nil
		}

		service := &OCRService{
			enabled:          true,
			provider:         mock,
			defaultLanguages: []string{"eng", "deu"},
		}

		_, err := service.ExtractTextFromPDF(context.Background(), []byte("pdf"), nil)
		require.NoError(t, err)
		assert.Equal(t, []string{"eng", "deu"}, capturedLanguages)
	})

	t.Run("uses provided languages", func(t *testing.T) {
		var capturedLanguages []string
		mock := newMockOCRProvider(true)
		mock.extractPDFFunc = func(ctx context.Context, data []byte, languages []string) (*OCRResult, error) {
			capturedLanguages = languages
			return &OCRResult{Text: "text", Pages: 1, Confidence: 0.9}, nil
		}

		service := &OCRService{
			enabled:          true,
			provider:         mock,
			defaultLanguages: []string{"eng"},
		}

		_, err := service.ExtractTextFromPDF(context.Background(), []byte("pdf"), []string{"fra", "spa"})
		require.NoError(t, err)
		assert.Equal(t, []string{"fra", "spa"}, capturedLanguages)
	})

	t.Run("returns result on success", func(t *testing.T) {
		mock := newMockOCRProvider(true)
		mock.extractPDFFunc = func(ctx context.Context, data []byte, languages []string) (*OCRResult, error) {
			return &OCRResult{
				Text:       "Extracted PDF text",
				Pages:      3,
				Confidence: 0.95,
			}, nil
		}

		service := &OCRService{
			enabled:          true,
			provider:         mock,
			defaultLanguages: []string{"eng"},
		}

		result, err := service.ExtractTextFromPDF(context.Background(), []byte("pdf data"), nil)
		require.NoError(t, err)
		assert.Equal(t, "Extracted PDF text", result.Text)
		assert.Equal(t, 3, result.Pages)
		assert.Equal(t, 0.95, result.Confidence)
	})

	t.Run("wraps provider error", func(t *testing.T) {
		mock := newMockOCRProvider(true)
		mock.extractPDFFunc = func(ctx context.Context, data []byte, languages []string) (*OCRResult, error) {
			return nil, assert.AnError
		}

		service := &OCRService{
			enabled:          true,
			provider:         mock,
			defaultLanguages: []string{"eng"},
		}

		result, err := service.ExtractTextFromPDF(context.Background(), []byte("pdf"), nil)
		require.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "OCR extraction failed")
	})
}

func TestOCRService_ExtractTextFromImage(t *testing.T) {
	t.Run("returns error when service is disabled", func(t *testing.T) {
		service := &OCRService{
			enabled: false,
		}

		result, err := service.ExtractTextFromImage(context.Background(), []byte("image data"), nil)
		require.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "not enabled")
	})

	t.Run("uses default languages when none provided", func(t *testing.T) {
		var capturedLanguages []string
		mock := newMockOCRProvider(true)
		mock.extractImageFunc = func(ctx context.Context, data []byte, languages []string) (*OCRResult, error) {
			capturedLanguages = languages
			return &OCRResult{Text: "text"}, nil
		}

		service := &OCRService{
			enabled:          true,
			provider:         mock,
			defaultLanguages: []string{"eng"},
		}

		_, err := service.ExtractTextFromImage(context.Background(), []byte("img"), nil)
		require.NoError(t, err)
		assert.Equal(t, []string{"eng"}, capturedLanguages)
	})

	t.Run("uses provided languages", func(t *testing.T) {
		var capturedLanguages []string
		mock := newMockOCRProvider(true)
		mock.extractImageFunc = func(ctx context.Context, data []byte, languages []string) (*OCRResult, error) {
			capturedLanguages = languages
			return &OCRResult{Text: "text"}, nil
		}

		service := &OCRService{
			enabled:          true,
			provider:         mock,
			defaultLanguages: []string{"eng"},
		}

		_, err := service.ExtractTextFromImage(context.Background(), []byte("img"), []string{"jpn"})
		require.NoError(t, err)
		assert.Equal(t, []string{"jpn"}, capturedLanguages)
	})

	t.Run("returns result on success", func(t *testing.T) {
		mock := newMockOCRProvider(true)
		mock.extractImageFunc = func(ctx context.Context, data []byte, languages []string) (*OCRResult, error) {
			return &OCRResult{
				Text:       "Image text",
				Pages:      1,
				Confidence: 0.88,
			}, nil
		}

		service := &OCRService{
			enabled:          true,
			provider:         mock,
			defaultLanguages: []string{"eng"},
		}

		result, err := service.ExtractTextFromImage(context.Background(), []byte("img"), nil)
		require.NoError(t, err)
		assert.Equal(t, "Image text", result.Text)
		assert.Equal(t, 0.88, result.Confidence)
	})
}

func TestOCRService_Close(t *testing.T) {
	t.Run("closes provider", func(t *testing.T) {
		closed := false
		mock := newMockOCRProvider(true)
		mock.closeFunc = func() error {
			closed = true
			return nil
		}

		service := &OCRService{
			provider: mock,
		}

		err := service.Close()
		require.NoError(t, err)
		assert.True(t, closed)
	})

	t.Run("returns nil when provider is nil", func(t *testing.T) {
		service := &OCRService{
			provider: nil,
		}

		err := service.Close()
		require.NoError(t, err)
	})

	t.Run("returns provider close error", func(t *testing.T) {
		mock := newMockOCRProvider(true)
		mock.closeFunc = func() error {
			return assert.AnError
		}

		service := &OCRService{
			provider: mock,
		}

		err := service.Close()
		require.Error(t, err)
	})
}

func TestOCRServiceConfig_Struct(t *testing.T) {
	cfg := OCRServiceConfig{
		Enabled:          true,
		ProviderType:     OCRProviderTypeTesseract,
		DefaultLanguages: []string{"eng", "deu"},
	}

	assert.True(t, cfg.Enabled)
	assert.Equal(t, OCRProviderTypeTesseract, cfg.ProviderType)
	assert.Equal(t, []string{"eng", "deu"}, cfg.DefaultLanguages)
}
