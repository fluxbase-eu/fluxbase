package ai

import (
	"context"
	"fmt"
)

// QuotaService handles quota checking and enforcement for knowledge bases
type QuotaService struct {
	storage *KnowledgeBaseStorage
}

// NewQuotaService creates a new quota service
func NewQuotaService(storage *KnowledgeBaseStorage) *QuotaService {
	return &QuotaService{
		storage: storage,
	}
}

// SystemQuotaLimits defines system-wide quota defaults
type SystemQuotaLimits struct {
	MaxDocuments    int
	MaxChunks       int
	MaxStorageBytes int64
}

// DefaultSystemQuotaLimits returns default system quota limits
func DefaultSystemQuotaLimits() SystemQuotaLimits {
	return SystemQuotaLimits{
		MaxDocuments:    10000,
		MaxChunks:       500000,
		MaxStorageBytes: 10 * 1024 * 1024 * 1024, // 10GB
	}
}

// DefaultKBQuotaLimits returns default KB-level quota limits
func DefaultKBQuotaLimits() SystemQuotaLimits {
	return SystemQuotaLimits{
		MaxDocuments:    1000,
		MaxChunks:       50000,
		MaxStorageBytes: 1 * 1024 * 1024 * 1024, // 1GB
	}
}

// CheckUserQuota checks if adding resources would exceed user's quota
func (s *QuotaService) CheckUserQuota(ctx context.Context, userID string, additionalDocs int, additionalChunks int, additionalBytes int64) error {
	quota, err := s.storage.GetUserQuota(ctx, userID)
	if err != nil {
		// If quota doesn't exist, use system defaults
		quota = &UserQuota{
			UserID:          userID,
			MaxDocuments:    DefaultSystemQuotaLimits().MaxDocuments,
			MaxChunks:       DefaultSystemQuotaLimits().MaxChunks,
			MaxStorageBytes: DefaultSystemQuotaLimits().MaxStorageBytes,
			UsedDocuments:   0,
			UsedChunks:      0,
			UsedStorageBytes: 0,
		}
	}

	// Check document quota
	if quota.UsedDocuments+additionalDocs > quota.MaxDocuments {
		return &QuotaError{
			ResourceType: "documents",
			Used:         int64(quota.UsedDocuments),
			Limit:        int64(quota.MaxDocuments),
			Requested:    int64(additionalDocs),
		}
	}

	// Check chunk quota
	if quota.UsedChunks+additionalChunks > quota.MaxChunks {
		return &QuotaError{
			ResourceType: "chunks",
			Used:         int64(quota.UsedChunks),
			Limit:        int64(quota.MaxChunks),
			Requested:    int64(additionalChunks),
		}
	}

	// Check storage quota
	if quota.UsedStorageBytes+additionalBytes > quota.MaxStorageBytes {
		return &QuotaError{
			ResourceType: "storage",
			Used:         quota.UsedStorageBytes,
			Limit:        quota.MaxStorageBytes,
			Requested:    additionalBytes,
		}
	}

	return nil
}

// CheckKBQuota checks if adding resources would exceed KB's quota
func (s *QuotaService) CheckKBQuota(ctx context.Context, kbID string, additionalDocs int, additionalChunks int, additionalBytes int64) error {
	kb, err := s.storage.GetKnowledgeBase(ctx, kbID)
	if err != nil {
		return fmt.Errorf("failed to get knowledge base: %w", err)
	}

	// Check document quota
	if kb.DocumentCount+additionalDocs > kb.QuotaMaxDocuments {
		return &QuotaError{
			ResourceType: "documents",
			Used:         int64(kb.DocumentCount),
			Limit:        int64(kb.QuotaMaxDocuments),
			Requested:    int64(additionalDocs),
		}
	}

	// Check chunk quota
	if kb.TotalChunks+additionalChunks > kb.QuotaMaxChunks {
		return &QuotaError{
			ResourceType: "chunks",
			Used:         int64(kb.TotalChunks),
			Limit:        int64(kb.QuotaMaxChunks),
			Requested:    int64(additionalChunks),
		}
	}

	// Storage quota check requires calculating current storage
	// For now, we'll track this in the storage layer
	return nil
}

// GetUserQuotaUsage returns current quota usage for a user
func (s *QuotaService) GetUserQuotaUsage(ctx context.Context, userID string) (*QuotaUsage, error) {
	quota, err := s.storage.GetUserQuota(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user quota: %w", err)
	}

	return &QuotaUsage{
		UserID:         userID,
		DocumentsUsed:  quota.UsedDocuments,
		DocumentsLimit: quota.MaxDocuments,
		ChunksUsed:     quota.UsedChunks,
		ChunksLimit:    quota.MaxChunks,
		StorageUsed:    quota.UsedStorageBytes,
		StorageLimit:   quota.MaxStorageBytes,
		CanAddDocument: quota.UsedDocuments < quota.MaxDocuments,
		CanAddChunks:   quota.UsedChunks < quota.MaxChunks,
	}, nil
}

// SetUserQuota sets quota limits for a user
func (s *QuotaService) SetUserQuota(ctx context.Context, userID string, req SetUserQuotaRequest) error {
	quota := &UserQuota{
		UserID: userID,
	}

	// Only update fields that are provided
	if req.MaxDocuments > 0 {
		quota.MaxDocuments = req.MaxDocuments
	}
	if req.MaxChunks > 0 {
		quota.MaxChunks = req.MaxChunks
	}
	if req.MaxStorageBytes > 0 {
		quota.MaxStorageBytes = req.MaxStorageBytes
	}

	return s.storage.SetUserQuota(ctx, quota)
}
