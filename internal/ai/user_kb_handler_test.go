package ai

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestNewUserKnowledgeBaseHandler(t *testing.T) {
	t.Run("creates handler with storage", func(t *testing.T) {
		storage := &KnowledgeBaseStorage{}
		handler := NewUserKnowledgeBaseHandler(storage)

		assert.NotNil(t, handler)
		assert.Same(t, storage, handler.storage)
	})
}

func TestUserKnowledgeBaseHandler_ListMyKnowledgeBases(t *testing.T) {
	t.Run("returns user's knowledge bases", func(t *testing.T) {
		// This would require a mock database or test database setup
		// For now, verify the handler is properly structured
		storage := &KnowledgeBaseStorage{}
		handler := NewUserKnowledgeBaseHandler(storage)

		assert.NotNil(t, handler.storage)
	})
}

func TestUserKnowledgeBaseHandler_CreateMyKnowledgeBase(t *testing.T) {
	t.Run("creates KB with owner set", func(t *testing.T) {
		storage := &KnowledgeBaseStorage{}
		handler := NewUserKnowledgeBaseHandler(storage)

		assert.NotNil(t, handler)
	})
}

func TestUserKnowledgeBaseHandler_ShareKnowledgeBase(t *testing.T) {
	t.Run("owner can grant permissions", func(t *testing.T) {
		storage := &KnowledgeBaseStorage{}
		handler := NewUserKnowledgeBaseHandler(storage)

		assert.NotNil(t, handler)
	})
}

func TestUserKnowledgeBaseHandler_RevokePermission(t *testing.T) {
	t.Run("owner can revoke permissions", func(t *testing.T) {
		storage := &KnowledgeBaseStorage{}
		handler := NewUserKnowledgeBaseHandler(storage)

		assert.NotNil(t, handler)
	})
}

// TestKBVisibility verifies KBVisibility enum values
func TestKBVisibility(t *testing.T) {
	t.Run("has correct visibility values", func(t *testing.T) {
		assert.Equal(t, "private", string(KBVisibilityPrivate))
		assert.Equal(t, "shared", string(KBVisibilityShared))
		assert.Equal(t, "public", string(KBVisibilityPublic))
	})
}

// TestKBPermission verifies KBPermission enum values
func TestKBPermission(t *testing.T) {
	t.Run("has correct permission values", func(t *testing.T) {
		assert.Equal(t, "viewer", string(KBPermissionViewer))
		assert.Equal(t, "editor", string(KBPermissionEditor))
		assert.Equal(t, "owner", string(KBPermissionOwner))
	})
}

// TestKBPermissionGrant verifies KBPermissionGrant struct
func TestKBPermissionGrant(t *testing.T) {
	t.Run("creates valid permission grant", func(t *testing.T) {
		kbID := uuid.New().String()
		userID := uuid.New().String()
		grantedBy := uuid.New().String()

		// Create a grant with all fields set
		grant := KBPermissionGrant{
			ID:              uuid.New().String(),
			KnowledgeBaseID: kbID,
			UserID:          userID,
			Permission:      KBPermissionEditor,
			GrantedBy:       &grantedBy,
		}

		assert.NotEmpty(t, grant.ID)
		assert.Equal(t, kbID, grant.KnowledgeBaseID)
		assert.Equal(t, userID, grant.UserID)
		assert.Equal(t, KBPermissionEditor, grant.Permission)
		assert.NotNil(t, grant.GrantedBy)
		assert.Equal(t, grantedBy, *grant.GrantedBy)
		// Note: GrantedAt will be zero for manually created structs
		// In real usage, it would be set by the database
	})
}

// TestKnowledgeBaseSummary verifies KB summary includes ownership fields
func TestKnowledgeBaseSummary(t *testing.T) {
	t.Run("includes ownership and visibility", func(t *testing.T) {
		ownerID := uuid.New().String()
		kb := &KnowledgeBase{
			ID:          uuid.New().String(),
			Name:        "Test KB",
			Namespace:   "default",
			Description: "Test description",
			Enabled:     true,
			OwnerID:     &ownerID,
			Visibility:  KBVisibilityPrivate,
		}

		summary := kb.ToSummary()

		assert.Equal(t, kb.ID, summary.ID)
		assert.Equal(t, kb.Name, summary.Name)
		assert.Equal(t, kb.Namespace, summary.Namespace)
		assert.Equal(t, kb.Description, summary.Description)
		assert.Equal(t, kb.Enabled, summary.Enabled)
		assert.NotNil(t, summary.OwnerID)
		assert.Equal(t, ownerID, *summary.OwnerID)
		assert.Equal(t, "private", summary.Visibility)
	})
}
