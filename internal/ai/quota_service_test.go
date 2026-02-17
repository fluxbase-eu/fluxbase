package ai

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestQuotaError_Error(t *testing.T) {
	t.Run("formats error message correctly", func(t *testing.T) {
		err := &QuotaError{
			ResourceType: "documents",
			Used:         95,
			Limit:        100,
			Requested:    10,
		}

		expected := "quota exceeded for documents: used=95, limit=100, requested=10"
		assert.Equal(t, expected, err.Error())
	})
}

func TestIsQuotaError(t *testing.T) {
	t.Run("identifies quota errors", func(t *testing.T) {
		quotaErr := &QuotaError{ResourceType: "documents"}
		otherErr := errors.New("some other error")

		assert.True(t, IsQuotaError(quotaErr))
		assert.False(t, IsQuotaError(otherErr))
	})
}

func TestDefaultSystemQuotaLimits(t *testing.T) {
	t.Run("returns sensible defaults", func(t *testing.T) {
		limits := DefaultSystemQuotaLimits()

		assert.Equal(t, 10000, limits.MaxDocuments)
		assert.Equal(t, 500000, limits.MaxChunks)
		assert.Equal(t, int64(10*1024*1024*1024), limits.MaxStorageBytes) // 10GB
	})
}

func TestDefaultKBQuotaLimits(t *testing.T) {
	t.Run("returns sensible defaults", func(t *testing.T) {
		limits := DefaultKBQuotaLimits()

		assert.Equal(t, 1000, limits.MaxDocuments)
		assert.Equal(t, 50000, limits.MaxChunks)
		assert.Equal(t, int64(1*1024*1024*1024), limits.MaxStorageBytes) // 1GB
	})
}

func TestQuotaService_CheckUserQuota_Basic(t *testing.T) {
	t.Run("quota error fields are correct", func(t *testing.T) {
		quotaErr := &QuotaError{
			ResourceType: "chunks",
			Used:         950,
			Limit:        1000,
			Requested:    100,
		}

		assert.Equal(t, "chunks", quotaErr.ResourceType)
		assert.Equal(t, int64(950), quotaErr.Used)
		assert.Equal(t, int64(1000), quotaErr.Limit)
		assert.Equal(t, int64(100), quotaErr.Requested)
		assert.Contains(t, quotaErr.Error(), "chunks")
	})
}

func TestQuotaUsage_Struct(t *testing.T) {
	t.Run("quota usage structure", func(t *testing.T) {
		usage := &QuotaUsage{
			UserID:         "user-123",
			DocumentsUsed:  50,
			DocumentsLimit: 100,
			ChunksUsed:     500,
			ChunksLimit:    1000,
			StorageUsed:    512 * 1024,
			StorageLimit:   1024 * 1024,
			CanAddDocument: true,
			CanAddChunks:   true,
		}

		assert.Equal(t, "user-123", usage.UserID)
		assert.Equal(t, 50, usage.DocumentsUsed)
		assert.Equal(t, 100, usage.DocumentsLimit)
		assert.True(t, usage.CanAddDocument)
		assert.True(t, usage.CanAddChunks)
	})
}

func TestSetUserQuotaRequest_Struct(t *testing.T) {
	t.Run("set quota request structure", func(t *testing.T) {
		req := SetUserQuotaRequest{
			MaxDocuments:    5000,
			MaxChunks:       100000,
			MaxStorageBytes: 5 * 1024 * 1024 * 1024, // 5GB
		}

		assert.Equal(t, 5000, req.MaxDocuments)
		assert.Equal(t, 100000, req.MaxChunks)
		assert.Equal(t, int64(5*1024*1024*1024), req.MaxStorageBytes)
	})
}
